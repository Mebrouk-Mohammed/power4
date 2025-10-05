package main

import (
	"log"

	"power4/auth"
	"power4/source/server"
)

func main() {
	// Initialise la base pour /login et /register (ne bloque pas le jeu si ça échoue)
	if err := auth.InitDB(); err != nil {
		log.Printf("auth.InitDB: %v (login/register désactivés)", err)
	}

	s := server.NewDefault()
	if err := s.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
