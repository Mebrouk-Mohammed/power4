package auth

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
)

// LegacyIndexData is passed to the index template.
type LegacyIndexData struct {
	Username string
	ELO      int
	GOBase   string
}

// LegacyIndexHandler renders the converted index menu.
func LegacyIndexHandler(w http.ResponseWriter, r *http.Request) {
	username := currentUser(r)
	if username == "" {
		// redirect to login if not authenticated
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// try to fetch user info from repo; ignore errors and fall back to cookie username
	elo := 1200
	if repo != nil {
		if u, _ := repo.GetByUsername(context.Background(), username); u != nil {
			username = u.Username
			// TODO: récupérer l'ELO réel si tu ajoutes la table de rating
		}
	}

	data := LegacyIndexData{
		Username: username,
		ELO:      elo,
		GOBase:   os.Getenv("GO_BASE"),
	}

	// default GO_BASE
	if data.GOBase == "" {
		data.GOBase = "http://localhost:8080"
	}
	// normalize GOBase: remove trailing slash to avoid // in template
	data.GOBase = strings.TrimRight(data.GOBase, "/")
	log.Printf("legacy: using GOBase=%s", data.GOBase)

	// On utilise le tpl global déjà chargé dans Init()
	if err := tpl.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		log.Printf("legacy: template execute error: %v", err)
		http.Error(w, "template execute error", http.StatusInternalServerError)
		return
	}
}
