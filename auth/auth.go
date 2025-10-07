package auth

import (
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var (
	db  *sql.DB
	tpl *template.Template
)

func Init() error {
	var err error
	db, err = sql.Open("sqlite", filepath.Join("auth", "users.db"))
	if err != nil {
		return err
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE
	)`); err != nil {
		return err
	}
	tpl, err = template.ParseFiles(
		filepath.Join("templates", "login.gohtml"),
		filepath.Join("templates", "register.gohtml"),
		filepath.Join("templates", "home.gohtml"),
	)
	return err
}

func RegisterRoutes() {
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/home", HomeHandler)
	http.HandleFunc("/logout", LogoutHandler)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		username := r.FormValue("username")
		if username == "" {
			tpl.ExecuteTemplate(w, "register.gohtml", "Le nom d'utilisateur est requis.")
			return
		}
		if _, err := db.Exec("INSERT INTO users(username) VALUES(?)", username); err != nil {
			tpl.ExecuteTemplate(w, "register.gohtml", "Nom d'utilisateur déjà pris.")
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	default:
		tpl.ExecuteTemplate(w, "register.gohtml", nil)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		username := r.FormValue("username")
		if username == "" {
			tpl.ExecuteTemplate(w, "login.gohtml", "Le nom d'utilisateur est requis.")
			return
		}
		row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)
		var id int
		if err := row.Scan(&id); err != nil {
			tpl.ExecuteTemplate(w, "login.gohtml", "Nom d'utilisateur inconnu.")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "user",
			Value:    username,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	default:
		tpl.ExecuteTemplate(w, "login.gohtml", nil)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r)
	if user == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "home.gohtml", user)
}

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

func CurrentUser(r *http.Request) string {
	c, err := r.Cookie("user")
	if err != nil || c == nil {
		return ""
	}
	return c.Value
}
