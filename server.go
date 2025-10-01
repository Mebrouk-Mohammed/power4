package main

import (
	"fmt"
	"net/http"
	"power4/auth"
)

func handler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	err := auth.RenderHome(w, user)
	if err != nil {
		http.Error(w, "Erreur de rendu de la page d'accueil", http.StatusInternalServerError)
	}
}
func gameHandler(w http.ResponseWriter, r *http.Request) {
	err := auth.RenderGame(w)
	if err != nil {
		http.Error(w, "Erreur de rendu du jeu", http.StatusInternalServerError)
	}
}

func main() {
	err := auth.InitDB()
	if err != nil {
		fmt.Println("Erreur DB:", err)
		return
	}
	http.HandleFunc("/", handler)
	http.HandleFunc("/register", auth.RegisterHandler)
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/game", gameHandler)

	fmt.Println("Serveur lancé sur http://localhost:8080/home")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Erreur lors du lancement du serveur :", err)
	}
}
