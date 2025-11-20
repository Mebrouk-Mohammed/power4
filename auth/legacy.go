package auth

import (
    "net/http"
    "os"
    "strings"
    "text/template"
    "path/filepath"
    "context"
    "log"
)

// LegacyIndexData is passed to the index template.
type LegacyIndexData struct {
    Username string
    ELO      int
    GOBase   string
}

// LegacyIndexHandler renders the converted index menu.
func LegacyIndexHandler(w http.ResponseWriter, r *http.Request) {
    // ensure templates are parsed via tpl; load the index template separately
    tpath := filepath.Join("templates", "index.gohtml")
    t, err := template.ParseFiles(tpath)
    if err != nil {
        http.Error(w, "template parse error", http.StatusInternalServerError)
        return
    }

    username := currentUser(r)
    if username == "" {
        // redirect to login if not authenticated
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // try to fetch user info from repo; ignore errors and fall back to cookie username
    var elo int = 1200
    if repo != nil {
        if u, _ := repo.GetByUsername(context.Background(), username); u != nil {
            username = u.Username
            // TODO: query rating if available; keep default 1200 for now
        }
    }

    data := LegacyIndexData{
        Username: username,
        ELO:      elo,
        GOBase:   os.Getenv("GO_BASE") ,
    }
    // default GO_BASE
    if data.GOBase == "" {
        data.GOBase = "http://localhost:8080"
    }
    // normalize GOBase: remove trailing slash to avoid //game in template
    data.GOBase = strings.TrimRight(data.GOBase, "/")
    log.Printf("legacy: using GOBase=%s", data.GOBase)

    if err := t.Execute(w, data); err != nil {
        http.Error(w, "template execute error", http.StatusInternalServerError)
        return
    }
}
