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
	tpl  *template.Template
	repo Repository
)

// Init : choix du repository (MySQL ou mémoire) + chargement de TOUS les templates.
func Init() error {
	var err error

	// Choix du repo : MySQL via env, sinon MySQL default, sinon mémoire
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

	switch repo.(type) {
	case *mysqlRepo:
		log.Println("auth: using MySQL repository")
	default:
		log.Println("auth: using memory repository")
	}

	// Chargement global de tous les templates *.gohtml
	tpl, err = template.
		New("base").
		Funcs(template.FuncMap{
			"add": func(i, j int) int { return i + j }, // pour le leaderboard
			"rangeN": func(n int) []int { // pour board_*.gohtml
				arr := make([]int, n)
				for i := 0; i < n; i++ {
					arr[i] = i
				}
				return arr
			},
		}).
		ParseGlob(filepath.Join("templates", "*.gohtml"))

	return err
}

// RegisterRoutes enregistre les routes d'auth + profil + leaderboard
func RegisterRoutes() {
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/home", HomeHandler) // tu peux le garder ou plus l'utiliser
	http.HandleFunc("/logout", LogoutHandler)

	http.HandleFunc("/legacy", LegacyIndexHandler)      // défini dans legacy.go
	http.HandleFunc("/profile", ProfileHandler)         // défini dans profile.go
	http.HandleFunc("/leaderboard", LeaderboardHandler) // défini dans profile.go
	http.HandleFunc("/public_profile", PublicProfileHandler)
	http.HandleFunc("/choose_avatar", ChooseAvatarHandler)
	http.HandleFunc("/delete_account", DeleteAccountHandler)

	// 🔥 nouvelle route pour les règles du Puissance 4 (RulesHandler dans rules.go)
	http.HandleFunc("/rules", RulesHandler)

	// Endpoints de debug
	http.HandleFunc("/debug/auth", DebugAuthHandler)
	http.HandleFunc("/debug/dbcheck", DBCheckHandler)
}

// DebugAuthHandler : test d'authentification, renvoie du JSON.
func DebugAuthHandler(w http.ResponseWriter, r *http.Request) {
	serverKey := os.Getenv("DEBUG_KEY")
	if serverKey == "" {
		serverKey = "dev_debug_key"
	}

	reqKey := r.URL.Query().Get("key")
	if reqKey == "" {
		_ = r.ParseForm()
		reqKey = r.FormValue("key")
	}
	if reqKey != serverKey {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user := r.URL.Query().Get("user")
	pass := r.URL.Query().Get("pass")
	if user == "" || pass == "" {
		_ = r.ParseForm()
		if user == "" {
			user = r.FormValue("username")
		}
		if pass == "" {
			pass = r.FormValue("password")
		}
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

// DBCheckHandler : petit check sur la DB
func DBCheckHandler(w http.ResponseWriter, r *http.Request) {
	if repo == nil {
		http.Error(w, "no repo configured", http.StatusInternalServerError)
		return
	}

	if u, err := repo.GetByUsername(r.Context(), "Test"); err == nil && u != nil {
		_, _ = w.Write([]byte("found user Test in repo\n"))
	} else if err != nil {
		_, _ = w.Write([]byte("GetByUsername error: " + err.Error() + "\n"))
	} else {
		_, _ = w.Write([]byte("user Test not found\n"))
	}

	switch repo.(type) {
	case *memoryRepo:
		_, _ = w.Write([]byte("repo type: memoryRepo\n"))
	default:
		_, _ = w.Write([]byte("repo type: mysqlRepo or DB-backed repo\n"))
	}
}

// RegisterHandler : GET = formulaire / POST = création + auto-login
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		_ = r.ParseForm()
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		email := r.FormValue("email")

		if username == "" || password == "" {
			_ = tpl.ExecuteTemplate(w, "register.gohtml", "Pseudo et mot de passe requis.")
			return
		}

		if _, err := repo.CreateUser(r.Context(), username, email, password); err != nil {
			log.Printf("register error for user '%s': %v", username, err)
			msg := "Nom d'utilisateur déjà pris ou erreur. (" + err.Error() + ")"
			_ = tpl.ExecuteTemplate(w, "register.gohtml", msg)
			return
		}

		// cookie user + redirection vers /legacy (ton vrai menu)
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

// LoginHandler : GET = formulaire / POST = vérif + cookie + redirect /legacy
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		_ = r.ParseForm()
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

		http.SetCookie(w, &http.Cookie{
			Name:     "user",
			Value:    username,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})
		// Redirection vers ton vrai menu
		http.Redirect(w, r, "/legacy", http.StatusSeeOther)

	default:
		_ = tpl.ExecuteTemplate(w, "login.gohtml", nil)
	}
}

// HomeHandler : encore là si tu veux tester /home, mais plus utilisé pour le flux normal
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	if user == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	_ = tpl.ExecuteTemplate(w, "home.gohtml", user)
}

// LogoutHandler : efface le cookie et renvoie sur /login
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

// currentUser : récupère le nom d'utilisateur depuis le cookie
func currentUser(r *http.Request) string {
	c, err := r.Cookie("user")
	if err != nil || c == nil {
		return ""
	}
	return c.Value
}
