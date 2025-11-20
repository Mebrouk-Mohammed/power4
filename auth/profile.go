package auth

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
)

// --------- PROFIL ---------

// ProfileData est passé au template profile.gohtml
type ProfileData struct {
	Username     string
	Email        string
	Avatar       string
	ELO          int
	Rank         string
	RankMin      int // ELO minimum du rang actuel
	RankMax      int // ELO nécessaire pour passer au rang suivant
	RankProgress int // Progression (0–100) vers le rang suivant
	GamesPlayed  int
	Wins         int
	Losses       int
	Draws        int
}

// ProfileHandler : affiche le profil du joueur connecté
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := currentUser(r)
	if username == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := ProfileData{
		Username: username,
		Avatar:   "/static/avatars/avatar1.png",
		ELO:      1200,
	}

	ctx := context.Background()

	if repo != nil {
		// Récup info user (email, avatar, id…)
		if u, err := repo.GetByUsername(ctx, username); err == nil && u != nil {
			data.Username = u.Username
			data.Email = u.Email
			if u.AvatarURL != "" {
				data.Avatar = u.AvatarURL
			} else {
				data.Avatar = pickAvatar(u.ID)
			}
		}

		// Si on a un mysqlRepo, on essaie de récupérer le rating + quelques stats
		if mr, ok := repo.(*mysqlRepo); ok {
			userID := dataFromUserID(repo, username)

			// ELO (table user_ratings)
			if userID != 0 {
				var rating sql.NullInt64
				err := mr.db.QueryRowContext(ctx,
					"SELECT rating FROM user_ratings WHERE user_id = ?",
					userID,
				).Scan(&rating)
				if err != nil && err != sql.ErrNoRows {
					log.Printf("profile: rating query error for user %s: %v", username, err)
				}
				if rating.Valid {
					data.ELO = int(rating.Int64)
				}
			}

			// (Optionnel) Stats de parties si tu as une table games
			var gp, wins sql.NullInt64
			err := mr.db.QueryRowContext(ctx, `
				SELECT 
					COUNT(*) AS gp,
					SUM(CASE WHEN status='finished' AND winner_id = ? THEN 1 ELSE 0 END) AS wins
				FROM games
				WHERE player1_id = ? OR player2_id = ?`,
				userID, userID, userID,
			).Scan(&gp, &wins)
			if err != nil && err != sql.ErrNoRows {
				log.Printf("profile: stats query error for user %s: %v", username, err)
			}
			if gp.Valid {
				data.GamesPlayed = int(gp.Int64)
			}
			if wins.Valid {
				data.Wins = int(wins.Int64)
			}
		}
	}

	// Calcul du rang + progression
	data.Rank = RankFromELO(data.ELO)
	data.RankMin, data.RankMax = RankBounds(data.ELO)
	data.RankProgress = RankProgress(data.ELO)

	if err := tpl.ExecuteTemplate(w, "profile.gohtml", data); err != nil {
		log.Printf("profile: template error for '%s': %v", username, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}

// --------- LEADERBOARD ---------

type PlayerRow struct {
	Username    string
	DisplayName string
	Avatar      string
	Rating      int
	Rank        string
	GamesPlayed int
	Wins        int
	Losses      int
	Draws       int
}

type LeaderboardData struct {
	Players []PlayerRow
}

// LeaderboardHandler : classement trié par ELO
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	data := LeaderboardData{}
	ctx := context.Background()

	// Si on a du MySQL, on essaie de récupérer les vrais ELO
	if mr, ok := repo.(*mysqlRepo); ok {
		rows, err := mr.db.QueryContext(ctx, `
			SELECT u.id, u.username, u.avatar_url,
			       COALESCE(r.rating, 1200) AS rating
			FROM users u
			LEFT JOIN user_ratings r ON r.user_id = u.id
			ORDER BY rating DESC, u.username ASC
			LIMIT 50`)
		if err != nil {
			log.Printf("leaderboard: query error: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var (
					id        int
					username  string
					avatarURL sql.NullString
					rating    int
				)
				if err := rows.Scan(&id, &username, &avatarURL, &rating); err != nil {
					log.Printf("leaderboard: scan error: %v", err)
					continue
				}
				avatar := pickAvatar(id)
				if avatarURL.Valid && avatarURL.String != "" {
					avatar = avatarURL.String
				}
				row := PlayerRow{
					Username:    username,
					DisplayName: username,
					Avatar:      avatar,
					Rating:      rating,
					Rank:        RankFromELO(rating),
				}
				data.Players = append(data.Players, row)
			}
		}
	}

	// Pas de DB ou erreur → faux leaderboard
	if len(data.Players) == 0 {
		for i := 1; i <= 8; i++ {
			name := "Joueur" + strconv.Itoa(i)
			rating := 1200 - i*5
			data.Players = append(data.Players, PlayerRow{
				Username:    name,
				DisplayName: name,
				Avatar:      pickAvatarForName(name),
				Rating:      rating,
				Rank:        RankFromELO(rating),
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

// --------- HELPERS ---------

// dataFromUserID récupère l'ID num d'un user à partir de son username.
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

// pickAvatar : avatar statique basé sur l'ID
func pickAvatar(id int) string {
	n := (id % 12)
	if n <= 0 {
		n = 1
	}
	return "/static/avatars/avatar" + strconv.Itoa(n) + ".png"
}

// pickAvatarForName : avatar déterministe basé sur le pseudo
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

// RankFromELO : transforme un ELO en rang texte
func RankFromELO(elo int) string {
	switch {
	case elo < 1000:
		return "Bronze"
	case elo < 1300:
		return "Silver"
	case elo < 1600:
		return "Gold"
	case elo < 1900:
		return "Platine"
	case elo < 2200:
		return "Diamant"
	default:
		return "Master"
	}
}

// RankBounds : retourne la plage ELO du rang actuel (min, max pour passer rang suivant)
func RankBounds(elo int) (min int, max int) {
	switch {
	case elo < 1000: // Bronze
		return 0, 1000
	case elo < 1300: // Silver
		return 1000, 1300
	case elo < 1600: // Gold
		return 1300, 1600
	case elo < 1900: // Platine
		return 1600, 1900
	case elo < 2200: // Diamant
		return 1900, 2200
	default: // Master : on fixe une plage arbitraire 2200–2600
		return 2200, 2600
	}
}

// RankProgress : progression en % dans le rang actuel vers le prochain rang
func RankProgress(elo int) int {
	min, max := RankBounds(elo)
	if max <= min {
		return 100
	}
	if elo < min {
		elo = min
	}
	if elo > max {
		elo = max
	}
	return (elo - min) * 100 / (max - min)
}
