package auth

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // driver SQLite sans CGO
)

var (
	db  *sql.DB
	tpl *template.Template
)

// InitDB initialise la DB (auth/users.db) et charge auth/templates/*.gohtml
func InitDB() error {
	// s'assure que le dossier et le sous-dossier templates existent
	if err := os.MkdirAll(filepath.Join("auth", "templates"), 0o755); err != nil {
		return err
	}

	// ouverture/creation de la DB locale
	dbPath := filepath.Join("auth", "users.db")
	d, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	if err := d.Ping(); err != nil {
		return err
	}
	db = d

	// table users
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE
	)`); err != nil {
		return err
	}

	// chargement des templates auth (si fichiers présents)
	var t *template.Template
	t, err = template.ParseGlob(filepath.Join("auth", "templates", "*.gohtml"))
	if err != nil || t == nil {
		// fallback minimal si les fichiers n'existent pas
		t = template.Must(template.New("login.gohtml").Parse(defaultLoginTpl))
		t = template.Must(t.New("register.gohtml").Parse(defaultRegisterTpl))
	}
	tpl = t
	return nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		_ = tpl.ExecuteTemplate(w, "register.gohtml", nil)
	case http.MethodPost:
		username := strings.TrimSpace(r.FormValue("username"))
		if username == "" {
			_ = tpl.ExecuteTemplate(w, "register.gohtml", "Nom d'utilisateur requis.")
			return
		}
		if _, err := db.Exec("INSERT INTO users(username) VALUES(?)", username); err != nil {
			_ = tpl.ExecuteTemplate(w, "register.gohtml", "Nom d'utilisateur déjà pris.")
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		_ = tpl.ExecuteTemplate(w, "login.gohtml", nil)
	case http.MethodPost:
		username := strings.TrimSpace(r.FormValue("username"))
		if username == "" {
			_ = tpl.ExecuteTemplate(w, "login.gohtml", "Nom d'utilisateur requis.")
			return
		}
		row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)
		var id int
		if err := row.Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = tpl.ExecuteTemplate(w, "login.gohtml", "Nom d'utilisateur inconnu.")
				return
			}
			http.Error(w, "erreur serveur", http.StatusInternalServerError)
			return
		}
		// ✅ après login on va sur la page du jeu
		http.Redirect(w, r, "/game?user="+username, http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// --- Templates de secours embarqués (si auth/templates/*.gohtml manquent) ---

const defaultLoginTpl = `
{{define "login.gohtml"}}
<!doctype html><html lang="fr"><head><meta charset="utf-8">
<title>Se connecter</title>
<style>
body{font-family:system-ui;background:#0b1020;color:#f1faee;margin:0}
.wrap{max-width:560px;margin:80px auto;padding:24px}
.card{background:#10223f;padding:28px;border-radius:14px;box-shadow:0 8px 24px rgba(0,0,0,.35)}
label{display:block;margin:10px 0 6px}
input[type=text]{width:100%;padding:10px;border-radius:8px;border:1px solid #2d3f70;background:#0f1b35;color:#fff}
button{margin-top:14px;padding:10px 16px;border-radius:10px;border:1px solid #2d3f70;background:#1b335e;color:#fff;font-weight:700;cursor:pointer}
a{color:#a8dadc}
.msg{color:#ffb4a2;margin:8px 0}
.toplinks{margin-bottom:10px}
</style></head><body>
<div class="wrap"><div class="card">
  <div class="toplinks"><a href="/auth">← Menu compte</a></div>
  <h1>Se connecter</h1>
  {{if .}}<div class="msg">{{.}}</div>{{end}}
  <form method="post" action="/login">
    <label>Nom d'utilisateur</label>
    <input type="text" name="username" placeholder="Votre pseudo" />
    <button type="submit">Connexion</button>
  </form>
  <p><a href="/register">Créer un compte</a> · <a href="/game">Retour au jeu</a></p>
</div></div>
</body></html>
{{end}}
`

const defaultRegisterTpl = `
{{define "register.gohtml"}}
<!doctype html><html lang="fr"><head><meta charset="utf-8">
<title>Créer un compte</title>
<style>
body{font-family:system-ui;background:#0b1020;color:#f1faee;margin:0}
.wrap{max-width:560px;margin:80px auto;padding:24px}
.card{background:#10223f;padding:28px;border-radius:14px;box-shadow:0 8px 24px rgba(0,0,0,.35)}
label{display:block;margin:10px 0 6px}
input[type=text]{width:100%;padding:10px;border-radius:8px;border:1px solid #2d3f70;background:#0f1b35;color:#fff}
button{margin-top:14px;padding:10px 16px;border-radius:10px;border:1px solid #2d3f70;background:#1b335e;color:#fff;font-weight:700;cursor:pointer}
a{color:#a8dadc}
.msg{color:#ffb4a2;margin:8px 0}
.toplinks{margin-bottom:10px}
</style></head><body>
<div class="wrap"><div class="card">
  <div class="toplinks"><a href="/auth">← Menu compte</a></div>
  <h1>Créer un compte</h1>
  {{if .}}<div class="msg">{{.}}</div>{{end}}
  <form method="post" action="/register">
    <label>Nom d'utilisateur</label>
    <input type="text" name="username" placeholder="Choisissez un pseudo" />
    <button type="submit">Inscription</button>
  </form>
  <p><a href="/login">Se connecter</a> · <a href="/game">Retour au jeu</a></p>
</div></div>
</body></html>
{{end}}
`
