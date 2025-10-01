package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bienvenue sur notre puissance 4 en ligne !")
}

func main() {
	http.HandleFunc("/", handler) // Route racine
	fmt.Println("Serveur lancé sur http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}
