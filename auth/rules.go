package auth

import (
	"log"
	"net/http"
)

// RulesHandler affiche la page des règles du Puissance 4.
func RulesHandler(w http.ResponseWriter, r *http.Request) {

	// On affiche simplement le template rules.gohtml
	if err := tpl.ExecuteTemplate(w, "rules.gohtml", nil); err != nil {
		log.Printf("rules: template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
