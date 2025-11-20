package auth

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

// ---------- PROFIL ----------

type ProfileData struct {
	Username    string
	Email       string
	Avatar      string
	ELO         int
	GamesPlayed int
	Wins        int
	Losses      int
	Draws       int
}

// ProfileHandler : GET = affiche profil / POST = met à jour Email + Avatar (si MySQL)
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := currentUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ctx := context.Background()

	data := ProfileData{
		Username: username,
		Avatar:   "/static/avatars/avatar1.png",
		ELO:      1200,
	}

	// Infos issues du repository
	if repo != nil {
		if u, _ := repo.GetByUsername(ctx, username); u != nil {
			data.Email = u.Email
			if u.AvatarURL != "" {
				data.Avatar = u.AvatarURL
			} else {
				data.Avatar = pickAvatar(u.ID)
			}
		}

		// Stats via MySQL si dispo
		if mr, ok := repo.(*mysqlRepo); ok {
			uid := dataFromUserID(repo, username)

			var rating sql.NullInt64
			_ = mr.db.QueryRowContext(ctx,
				"SELECT rating FROM user_ratings WHERE user_id=?",
				uid,
			).Scan(&rating)
			if rating.Valid {
				data.ELO = int(rating.Int64)
			}

			var gp, wins sql.NullInt64
			_ = mr.db.QueryRowContext(ctx, `
                SELECT COUNT(*) as gp,
                       SUM(status='finished' AND winner_id=?) as wins
                FROM games g
                WHERE g.player1_id=? OR g.player2_id=?`,
				uid, uid, uid,
			).Scan(&gp, &wins)

			if gp.Valid {
				data.GamesPlayed = int(gp.Int64)
			}
			if wins.Valid {
				data.Wins = int(wins.Int64)
			}
		}
	}

	// Mise à jour via formulaire
	if r.Method == http.MethodPost {
		if mr, ok := repo.(*mysqlRepo); ok {
			_ = r.ParseForm()
			newEmail := r.FormValue("email")
			newAvatar := r.FormValue("avatar")

			uid := dataFromUserID(repo, username)
			_, err := mr.db.ExecContext(ctx,
				"UPDATE users SET email=?, avatar_url=? WHERE id=?",
				newEmail, newAvatar, uid,
			)
			if err != nil {
				log.Printf("profile update error for '%s': %v", username, err)
				http.Error(w, "Erreur mise à jour profil", http.StatusInternalServerError)
				return
			}

			data.Email = newEmail
			if newAvatar != "" {
				data.Avatar = newAvatar
			}
		}
	}

	if err := tpl.ExecuteTemplate(w, "profile.gohtml", data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
}

// ---------- LEADERBOARD ----------

type PlayerRow struct {
	Username    string
	DisplayName string
	Avatar      string
	Rating      int
	GamesPlayed int
	Wins        int
	Losses      int
	Draws       int
}

type LeaderboardData struct {
	Players []PlayerRow
}

// LeaderboardHandler : affiche le classement (MySQL si dispo, sinon placeholder)
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	data := LeaderboardData{}

	// Classement réel si MySQL
	if mr, ok := repo.(*mysqlRepo); ok {
		rows, err := mr.db.QueryContext(ctx, `
            SELECT u.username,
                   COALESCE(u.avatar_url, ''),
                   COALESCE(r.rating, 1200) AS rating
            FROM users u
            LEFT JOIN user_ratings r ON r.user_id = u.id
            ORDER BY rating DESC
            LIMIT 20
        `)
		if err != nil {
			log.Printf("leaderboard query error: %v", err)
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var username, avatarURL string
			var rating int
			if err := rows.Scan(&username, &avatarURL, &rating); err != nil {
				log.Printf("leaderboard scan error: %v", err)
				continue
			}
			row := PlayerRow{
				Username:    username,
				DisplayName: username,
				Avatar:      avatarURL,
				Rating:      rating,
			}
			if row.Avatar == "" {
				row.Avatar = pickAvatarForName(username)
			}
			data.Players = append(data.Players, row)
		}
	} else {
		// Fallback : faux joueurs pour garder le visuel
		for i := 1; i <= 8; i++ {
			name := "Joueur" + strconv.Itoa(i)
			data.Players = append(data.Players, PlayerRow{
				Username:    name,
				DisplayName: name,
				Avatar:      pickAvatarForName(name),
				Rating:      1200 - i*5,
				GamesPlayed: 10 + i,
				Wins:        5 + i%3,
				Losses:      4 + (i % 2),
				Draws:       1,
			})
		}
	}

	log.Printf("leaderboard: rendering %d players", len(data.Players))

	if err := tpl.ExecuteTemplate(w, "leaderboard.gohtml", data); err != nil {
		log.Printf("leaderboard render error: %v", err)
		http.Error(w, "render error: please check server logs", http.StatusInternalServerError)
		return
	}
}

// ---------- Helpers communs ----------

// dataFromUserID : récupère l'ID utilisateur depuis le repo
func dataFromUserID(r Repository, username string) int {
	if r == nil {
		return 0
	}
	u, _ := r.GetByUsername(context.Background(), username)
	if u == nil {
		return 0
	}
	return u.ID
}

// pickAvatar : avatar basé sur l'ID
func pickAvatar(id int) string {
	n := (id % 12)
	if n <= 0 {
		n = 1
	}
	return "/static/avatars/avatar" + strconv.Itoa(n) + ".png"
}

// pickAvatarForName : avatar déterministe à partir du pseudo
func pickAvatarForName(name string) string {
	if name == "" {
		return "/static/avatars/avatar1.png"
	}
	h := 0
	for i := 0; i < len(name); i++ {
		h = (h*31 + int(name[i]))
	}
	n := (h % 12)
	if n < 0 {
		n = -n
	}
	if n == 0 {
		n = 1
	}
	return "/static/avatars/avatar" + strconv.Itoa(n) + ".png"
}
