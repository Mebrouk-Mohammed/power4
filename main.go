package main

import (
	"log"
	"net/http"
	"power4/auth"
	"power4/source/server"
)

func main() {
	if err := auth.Init(); err != nil {
		log.Fatal(err)
	}
	auth.RegisterRoutes()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/register", http.StatusSeeOther) })
	s := server.NewDefault()
	if err := s.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
