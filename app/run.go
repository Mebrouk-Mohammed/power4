package app

import (
	"log"
	"net/http"
	"strconv"
)

func Main() {
	initDB()

	// 🟢 Route d'accueil : redirige vers ton site PHP (page register)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://localhost/power4/register.php", http.StatusFound)
	})

	// 🧩 Route de test DB (à garder si tu veux tester)
	http.HandleFunc("/ping-db", func(w http.ResponseWriter, r *http.Request) {
		row := DB.QueryRow("SELECT COUNT(*) FROM users")
		var count int
		if err := row.Scan(&count); err != nil {
			http.Error(w, "Erreur DB: "+err.Error(), 500)
			return
		}

		w.Write([]byte("OK DB. Utilisateurs en base: " + strconv.Itoa(count)))
	})

	log.Println("🚀 Serveur Go sur :8080")
	http.ListenAndServe(":8080", nil)
}
