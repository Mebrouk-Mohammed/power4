package main

import (
	"log"
	"os"

	"power4/auth"          // ⬅️ ajoute l'auth
	"power4/source/server" // serveur du jeu
)

func main() {
	// 1) Initialiser l'auth (DB + templates)
	if err := auth.Init(); err != nil {
		log.Fatal(err)
	}
	// 2) Enregistrer les routes d'auth dans le même mux
	auth.RegisterRoutes()

	// 3) Démarrer le serveur du jeu (qui écoute déjà /, /play, /reset, /new)
	s := server.NewDefault()

	// Écouter sur le port fourni par l'environnement (ex. PORT=8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	if err := s.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
