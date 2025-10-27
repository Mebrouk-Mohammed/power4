package app

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func initDB() {
	// Connexion MySQL -> utilise maintenant la base power4_db
	dsn := "root:@tcp(127.0.0.1:3306)/power4_db?parseTime=true&charset=utf8mb4"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Erreur ouverture MySQL:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("Impossible de joindre MySQL:", err)
	}

	log.Println("✅ Connecté à MySQL (base power4_db)")
}
