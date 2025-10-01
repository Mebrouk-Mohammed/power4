package auth

import (
	"database/sql"
	"html/template"
	"net/http"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var tpl *template.Template

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite", "c:/Users/blank/Desktop/power4/auth/users.db")
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE
    )`)
	if err != nil {
		return err
	}
	tpl, err = template.ParseGlob(filepath.Join(".", "auth", "templates", "*.gohtml"))

	return err
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		_, err := db.Exec("INSERT INTO users(username) VALUES(?)", username)
		if err != nil {
			tpl.ExecuteTemplate(w, "register.gohtml", "Nom d'utilisateur déjà pris.")
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "register.gohtml", nil)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		row := db.QueryRow("SELECT id FROM users WHERE username=?", username)
		var id int
		err := row.Scan(&id)
		if err != nil {
			tpl.ExecuteTemplate(w, "login.gohtml", "Nom d'utilisateur inconnu.")
			return
		}
		http.Redirect(w, r, "/?user="+username, http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "login.gohtml", nil)
}
func RenderHome(w http.ResponseWriter, username string) error {
	return tpl.ExecuteTemplate(w, "home.gohtml", username)
}
func RenderGame(w http.ResponseWriter) error {
	return tpl.ExecuteTemplate(w, "game.gohtml", nil)
}
