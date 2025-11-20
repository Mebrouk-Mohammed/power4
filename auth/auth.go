package auth // Déclare le package auth

import (
	"database/sql"  // Permet de gérer la base de données
	"html/template" // Gère les templates HTML
	"net/http"      // Gère le serveur HTTP
	"path/filepath" // Permet de gérer les chemins de fichiers
	"time"          // Permet de gérer les dates et durées

	_ "modernc.org/sqlite" // Importe le driver SQLite
)

var (
	db  *sql.DB            // Variable globale pour accéder à la base de données
	tpl *template.Template // Variable globale pour stocker les templates HTML
)

// Init ouvre la base users.db, crée la table si besoin et charge les templates
func Init() error {
	var err error // Variable pour stocker les erreurs

	db, err = sql.Open("sqlite", filepath.Join("auth", "users.db")) // Ouvre la base SQLite
	if err != nil {
		return err // Renvoie l’erreur si échec ouverture
	}

	// Crée la table users si elle n'existe pas
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE
	)`); err != nil {
		return err // Renvoie l’erreur si échec création
	}

	// Charge les templates login, register et home
	tpl, err = template.ParseFiles(
		filepath.Join("templates", "login.gohtml"),
		filepath.Join("templates", "register.gohtml"),
		filepath.Join("templates", "home.gohtml"),
	)

	return err // Renvoie l’erreur éventuelle
}

// RegisterRoutes associe les URL aux handlers correspondants
func RegisterRoutes() {
	http.HandleFunc("/login", LoginHandler)       // Route /login
	http.HandleFunc("/register", RegisterHandler) // Route /register
	http.HandleFunc("/home", HomeHandler)         // Route /home
	http.HandleFunc("/logout", LogoutHandler)     // Route /logout
}

// RegisterHandler gère l'inscription utilisateur
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method { // Vérifie la méthode HTTP
	case http.MethodPost: // Si formulaire envoyé (POST)
		username := r.FormValue("username") // Récupère le nom entré

		if username == "" { // Vérifie que le nom n’est pas vide
			_ = tpl.ExecuteTemplate(w, "register.gohtml", "Le nom d'utilisateur est requis.")
			return
		}

		// Tente d'ajouter l’utilisateur dans la base
		if _, err := db.Exec("INSERT INTO users(username) VALUES(?)", username); err != nil {
			_ = tpl.ExecuteTemplate(w, "register.gohtml", "Nom d'utilisateur déjà pris.")
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther) // Redirection vers /login

	default: // Si GET, affiche juste le formulaire
		_ = tpl.ExecuteTemplate(w, "register.gohtml", nil)
	}
}

// LoginHandler gère la connexion utilisateur
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method { // Vérifie la méthode
	case http.MethodPost: // Traitement du formulaire
		username := r.FormValue("username") // Récupère le nom

		if username == "" { // Vérifie que le nom n’est pas vide
			_ = tpl.ExecuteTemplate(w, "login.gohtml", "Le nom d'utilisateur est requis.")
			return
		}

		// Vérifie si l’utilisateur existe dans la base
		row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)

		var id int                            // Variable pour stocker l'id
		if err := row.Scan(&id); err != nil { // Si aucun résultat, erreur
			_ = tpl.ExecuteTemplate(w, "login.gohtml", "Nom d'utilisateur inconnu.")
			return
		}

		// Crée un cookie simple pour la session
		http.SetCookie(w, &http.Cookie{
			Name:     "user",                              // Nom du cookie
			Value:    username,                            // Valeur = nom utilisateur
			Path:     "/",                                 // Accessible sur tout le site
			HttpOnly: true,                                // Non accessible depuis JS
			Expires:  time.Now().Add(30 * 24 * time.Hour), // Durée 30 jours
		})

		http.Redirect(w, r, "/home", http.StatusSeeOther) // Redirection vers /home

	default:
		_ = tpl.ExecuteTemplate(w, "login.gohtml", nil) // Affiche le formulaire
	}
}

// HomeHandler affiche la page d’accueil si l’utilisateur est connecté
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r) // Récupère le nom depuis le cookie

	if user == "" { // Si pas connecté
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_ = tpl.ExecuteTemplate(w, "home.gohtml", user) // Affiche la page avec le nom
}

// LogoutHandler supprime le cookie et renvoie vers /login
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user",          // Nom du cookie
		Value:    "",              // Valeur vide
		Path:     "/",             // Chemin global
		HttpOnly: true,            // Protégé
		Expires:  time.Unix(0, 0), // Expire immédiatement
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther) // Redirection
}

// currentUser récupère le nom d’utilisateur via le cookie
func currentUser(r *http.Request) string {
	c, err := r.Cookie("user") // Essaie de lire le cookie

	if err != nil || c == nil { // Si erreur ou absent
		return ""
	}

	return c.Value // Renvoie le nom contenu dans le cookie
}
