package auth

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	tpl *template.Template
	repo Repository
)

// Init ouvre la DB (auth/users.db), crée la table si besoin, et charge les templates d'auth.
func Init() error {
	// Choisit le repository : MySQL si configuré via env, sinon mémoire
	var err error
	// Try environment first. If not present, try sensible defaults (local MySQL like in db.php).
	var dbErr error
	if os.Getenv("DB_USER") != "" && os.Getenv("DB_NAME") != "" {
		repo, dbErr = NewMySQLFromEnv()
		if dbErr != nil {
			log.Printf("mysql connect via env failed: %v — trying defaults", dbErr)
			repo, dbErr = NewMySQLFromDefaults()
			if dbErr != nil {
				log.Printf("mysql connect via defaults failed: %v — using memory repo", dbErr)
				repo = NewMemoryRepo()
			}
		}
	} else {
		repo, dbErr = NewMySQLFromDefaults()
		if dbErr != nil {
			log.Printf("mysql connect via defaults failed: %v — using memory repo", dbErr)
			repo = NewMemoryRepo()
		}
	}

	// Log which repository we're using (helpful for debugging)
	switch repo.(type) {
	case *mysqlRepo:
		log.Println("auth: using MySQL repository")
	default:
		log.Println("auth: using memory repository")
	}

	// On ne charge ici que les templates d'auth (les autres sont gérés par ton serveur de jeu)
	tpl, err = template.ParseFiles(
		filepath.Join("templates", "login.gohtml"),
		filepath.Join("templates", "register.gohtml"),
		filepath.Join("templates", "home.gohtml"),
	)
	return err
}

// RegisterRoutes enregistre les routes /login /register /home /logout
func RegisterRoutes() {
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/home", HomeHandler)
	http.HandleFunc("/logout", LogoutHandler)
	http.HandleFunc("/legacy", LegacyIndexHandler)
	http.HandleFunc("/profile", ProfileHandler)
	http.HandleFunc("/leaderboard", LeaderboardHandler)
	http.HandleFunc("/public_profile", PublicProfileHandler)
	http.HandleFunc("/choose_avatar", ChooseAvatarHandler)
	http.HandleFunc("/delete_account", DeleteAccountHandler)
	// Debug auth endpoint (dev only) - requires ?key=... or DEBUG_KEY env
	http.HandleFunc("/debug/auth", DebugAuthHandler)
}

// DebugAuthHandler tests authentication using repo.Authenticate and returns JSON.
// Protect with key: pass ?key=... matching DEBUG_KEY env var or default 'dev_debug_key'.
func DebugAuthHandler(w http.ResponseWriter, r *http.Request) {
	// determine server key
	serverKey := os.Getenv("DEBUG_KEY")
	if serverKey == "" {
		serverKey = "dev_debug_key"
	}

	// accept key via query or form
	reqKey := r.URL.Query().Get("key")
	if reqKey == "" {
		_ = r.ParseForm()
		reqKey = r.FormValue("key")
	}
	if reqKey != serverKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// get credentials
	user := r.URL.Query().Get("user")
	pass := r.URL.Query().Get("pass")
	if user == "" || pass == "" {
		_ = r.ParseForm()
		if user == "" { user = r.FormValue("username") }
		if pass == "" { pass = r.FormValue("password") }
	}

	out := map[string]interface{}{"ok": false}
	w.Header().Set("Content-Type", "application/json")
	if repo == nil {
		out["reason"] = "no repo configured"
		_ = json.NewEncoder(w).Encode(out)
		return
	}

	u, err := repo.Authenticate(r.Context(), user, pass)
	if err != nil {
		out["reason"] = err.Error()
		_ = json.NewEncoder(w).Encode(out)
		return
	}
	if u == nil {
		out["reason"] = "not found or invalid credentials"
		_ = json.NewEncoder(w).Encode(out)
		return
	}
	out["ok"] = true
	out["username"] = u.Username
	out["id"] = u.ID
	_ = json.NewEncoder(w).Encode(out)
}

// DBCheckHandler performs simple checks against the repository/database.
func DBCheckHandler(w http.ResponseWriter, r *http.Request) {
	if repo == nil {
		http.Error(w, "no repo configured", http.StatusInternalServerError)
		return
	}
	// Try a simple users count
	if u, err := repo.GetByUsername(r.Context(), "Test"); err == nil && u != nil {
		_, _ = w.Write([]byte("found user Test in repo\n"))
	} else if err != nil {
		_, _ = w.Write([]byte("GetByUsername error: " + err.Error() + "\n"))
	} else {
		_, _ = w.Write([]byte("user Test not found\n"))
	}
	// Indicate which repo type
	switch repo.(type) {
	case *memoryRepo:
		_, _ = w.Write([]byte("repo type: memoryRepo\n"))
	default:
		// assume mysqlRepo or other DB-backed repo
		_, _ = w.Write([]byte("repo type: mysqlRepo or DB-backed repo\n"))
	}
}
 

// Handlers

// GET: formulaire / POST: création utilisateur puis redirection /login
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		email := r.FormValue("email")
		if username == "" || password == "" {
			_ = tpl.ExecuteTemplate(w, "register.gohtml", "Pseudo et mot de passe requis.")
			return
		}
		if _, err := repo.CreateUser(r.Context(), username, email, password); err != nil {
			// Log detailed error server-side, and show message to user (include error for easier debugging)
			log.Printf("register error for user '%s': %v", username, err)
			msg := "Nom d'utilisateur déjà pris ou erreur."
			// expose the underlying error message in dev mode — append it to msg
			msg = msg + " (" + err.Error() + ")"
			_ = tpl.ExecuteTemplate(w, "register.gohtml", msg)
			return
		}
		// Auto-login: set cookie and redirect to legacy menu
		http.SetCookie(w, &http.Cookie{
			Name:     "user",
			Value:    username,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})
		http.Redirect(w, r, "/legacy", http.StatusSeeOther)
	default:
		_ = tpl.ExecuteTemplate(w, "register.gohtml", nil)
	}
}

// GET: formulaire / POST: vérifie l'utilisateur, pose un cookie et redirige /home
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		if username == "" || password == "" {
			_ = tpl.ExecuteTemplate(w, "login.gohtml", "Pseudo et mot de passe requis.")
			return
		}
		log.Printf("login attempt for username='%s'", username)
		u, err := repo.Authenticate(r.Context(), username, password)
		if err != nil {
			log.Printf("authenticate error for '%s': %v", username, err)
		}
		if u == nil {
			log.Printf("authenticate: user not found for '%s'", username)
			_ = tpl.ExecuteTemplate(w, "login.gohtml", "Nom d'utilisateur inconnu ou mot de passe incorrect.")
			return
		}
		// Cookie de 'session' ultra simple
		http.SetCookie(w, &http.Cookie{
			Name:     "user",
			Value:    username,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	default:
		_ = tpl.ExecuteTemplate(w, "login.gohtml", nil)
	}
}

// Pas de stockage fichier — tout reste en mémoire.

// Page d'accueil authentifiée
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_ = tpl.ExecuteTemplate(w, "home.gohtml", user)
}

// Déconnexion : on efface le cookie et on renvoie au /login
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Helpers

func currentUser(r *http.Request) string {
	c, err := r.Cookie("user")
	if err != nil || c == nil {
		return ""
	}
	return c.Value
}