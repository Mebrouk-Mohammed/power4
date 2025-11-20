package auth

import (
    "context"
    "net/http"
    "path/filepath"
    "text/template"
    "time"
)

// PublicProfileHandler shows a public profile for any username (query param `username`).
func PublicProfileHandler(w http.ResponseWriter, r *http.Request) {
    username := r.URL.Query().Get("username")
    if username == "" {
        http.Error(w, "username required", http.StatusBadRequest)
        return
    }

    tmplPath := filepath.Join("templates", "public_profile.gohtml")
    t, err := template.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "template error", http.StatusInternalServerError)
        return
    }

    data := ProfileData{Username: username, Avatar: pickAvatarForName(username), ELO: 1200}
    if repo != nil {
        if u, _ := repo.GetByUsername(context.Background(), username); u != nil {
            data.Email = u.Email
            if u.AvatarURL != "" { data.Avatar = u.AvatarURL }
        }
    }

    _ = t.ExecuteTemplate(w, "public_profile", data)
}

// ChooseAvatarHandler shows a simple avatar chooser (GET) and sets avatar (POST).
func ChooseAvatarHandler(w http.ResponseWriter, r *http.Request) {
    username := currentUser(r)
    if username == "" {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    u, _ := repo.GetByUsername(context.Background(), username)
    if u == nil {
        http.Error(w, "user not found", http.StatusNotFound)
        return
    }

    if r.Method == http.MethodPost {
        avatar := r.FormValue("avatar")
        if avatar == "" {
            http.Error(w, "avatar required", http.StatusBadRequest)
            return
        }
        // store avatar (could be a path like /static/avatars/avatar3.png)
        _ = repo.UpdateAvatar(context.Background(), u.ID, avatar)
        http.Redirect(w, r, "/profile", http.StatusSeeOther)
        return
    }

    tmplPath := filepath.Join("templates", "choose_avatar.gohtml")
    t, err := template.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "template error", http.StatusInternalServerError)
        return
    }

    // the template file defines "choose_avatar", render that named template
    _ = t.ExecuteTemplate(w, "choose_avatar", u)
}

// DeleteAccountHandler deletes the currently logged-in account (POST only).
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Redirect(w, r, "/profile", http.StatusSeeOther)
        return
    }
    username := currentUser(r)
    if username == "" {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
    u, _ := repo.GetByUsername(context.Background(), username)
    if u == nil {
        http.Error(w, "user not found", http.StatusNotFound)
        return
    }
    _ = repo.DeleteUser(context.Background(), u.ID)
    // clear cookie
    http.SetCookie(w, &http.Cookie{Name: "user", Value: "", Path: "/", Expires: time.Unix(0, 0)})
    http.Redirect(w, r, "/register", http.StatusSeeOther)
}

