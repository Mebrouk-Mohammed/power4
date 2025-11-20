package main

import (
	"log"

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
	if err := s.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
