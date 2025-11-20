package auth

import (
	"context"
	"log"
	"net/http"
	"time"
)

// PublicProfileHandler shows a public profile for any username (query param `username`).
func PublicProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	data := ProfileData{
		Username: username,
		Avatar:   pickAvatarForName(username),
		ELO:      1200,
	}

	// Enrichir avec les infos de la BDD si possible
	if repo != nil {
		if u, err := repo.GetByUsername(context.Background(), username); err == nil && u != nil {
			data.Email = u.Email
			if u.AvatarURL != "" {
				data.Avatar = u.AvatarURL
			}
		}
	}

	// On utilise le tpl global (ParseGlob dans Init)
	if err := tpl.ExecuteTemplate(w, "public_profile.gohtml", data); err != nil {
		log.Printf("public profile template error for '%s': %v", username, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

// ChooseAvatarHandler shows a simple avatar chooser (GET) and sets avatar (POST).
func ChooseAvatarHandler(w http.ResponseWriter, r *http.Request) {
	username := currentUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if repo == nil {
		http.Error(w, "repository not configured", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	u, err := repo.GetByUsername(ctx, username)
	if err != nil {
		log.Printf("choose_avatar: GetByUsername error for '%s': %v", username, err)
		http.Error(w, "user lookup error", http.StatusInternalServerError)
		return
	}
	if u == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// POST : on met à jour l'avatar
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form data", http.StatusBadRequest)
			return
		}
		avatar := r.FormValue("avatar")
		if avatar == "" {
			http.Error(w, "avatar required", http.StatusBadRequest)
			return
		}

		if err := repo.UpdateAvatar(ctx, u.ID, avatar); err != nil {
			log.Printf("choose_avatar: UpdateAvatar error for '%s': %v", username, err)
			http.Error(w, "update avatar error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// GET : afficher la page de choix d'avatar
	if err := tpl.ExecuteTemplate(w, "choose_avatar.gohtml", u); err != nil {
		log.Printf("choose_avatar: template error for '%s': %v", username, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

// DeleteAccountHandler deletes the currently logged-in account.
// ⚠️ Version "rapide" : accepte GET et POST (un simple clic sur le lien supprime le compte).
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	username := currentUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if repo == nil {
		http.Error(w, "repository not configured", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	u, err := repo.GetByUsername(ctx, username)
	if err != nil {
		log.Printf("delete_account: GetByUsername error for '%s': %v", username, err)
		http.Error(w, "Erreur lors de la suppression du compte", http.StatusInternalServerError)
		return
	}
	if u == nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := repo.DeleteUser(ctx, u.ID); err != nil {
		log.Printf("delete_account: DeleteUser error for '%s': %v", username, err)
		http.Error(w, "Erreur lors de la suppression du compte", http.StatusInternalServerError)
		return
	}

	// clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	})

	// retour vers la page de connexion
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
