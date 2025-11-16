package main

import "fmt"

func main() {
	api := NewAPI()

	uid, username, err := api.GetCurrentUser()
	if err != nil {
		panic("❌ Impossible de récupérer l'utilisateur connecté : " + err.Error())
	}

	fmt.Println("👤 Connecté en tant que :", username, "(ID", uid, ")")

	id, err := api.CreateGame(uid, 0)
	if err != nil {
		panic("❌ Erreur création de partie : " + err.Error())
	}

	fmt.Println("✅ Partie créée avec l'ID :", id)
}
