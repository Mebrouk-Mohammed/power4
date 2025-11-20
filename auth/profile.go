package auth

import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "path/filepath"
    "strconv"
    "text/template"
)

// ProfileData is passed to profile template
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

// ProfileHandler renders the logged-in user's profile.
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
    username := currentUser(r)
    if username == "" {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    tmplPath := filepath.Join("templates", "profile.gohtml")
    t, err := template.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "template error", http.StatusInternalServerError)
        return
    }

    data := ProfileData{Username: username, Avatar: "/static/avatars/avatar1.png", ELO: 1200}

    // try to enrich from repo if possible
    if repo != nil {
        if u, _ := repo.GetByUsername(context.Background(), username); u != nil {
            data.Email = u.Email
            if u.AvatarURL != "" {
                // if AvatarURL looks like a local filename, serve from /static/avatars
                data.Avatar = u.AvatarURL
            } else {
                // choose deterministic avatar based on user id
                data.Avatar = pickAvatar(u.ID)
            }
        }
        // if repo is mysqlRepo we can query ratings and stats directly
        if mr, ok := repo.(*mysqlRepo); ok {
            // rating
            var rating sql.NullInt64
            _ = mr.db.QueryRowContext(context.Background(), "SELECT rating FROM user_ratings WHERE user_id=?", dataFromUserID(repo, username)).Scan(&rating)
            if rating.Valid { data.ELO = int(rating.Int64) }
            // games stats (best-effort)
            var gp, wins sql.NullInt64
            _ = mr.db.QueryRowContext(context.Background(), `
                SELECT COUNT(*) as gp,
                    SUM(status='finished' AND winner_id=?) as wins
                FROM games g
                WHERE g.player1_id=? OR g.player2_id=?
            `, dataFromUserID(repo, username), dataFromUserID(repo, username), dataFromUserID(repo, username)).Scan(&gp, &wins)
            if gp.Valid { data.GamesPlayed = int(gp.Int64) }
            if wins.Valid { data.Wins = int(wins.Int64) }
        }
    }

    if err := t.Execute(w, data); err != nil {
        http.Error(w, "render error", http.StatusInternalServerError)
        return
    }
}

// Leaderboard related types
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

// LeaderboardHandler renders the leaderboard.
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
    tmplPath := filepath.Join("templates", "leaderboard.gohtml")
    // add simple func map for add
    t := template.New("leaderboard").Funcs(template.FuncMap{"add": func(i, j int) int { return i + j }})
    t, err := t.ParseFiles(tmplPath)
    if err != nil {
        http.Error(w, "template error", http.StatusInternalServerError)
        return
    }
    // For now always render a static leaderboard visual with placeholder players
    data := LeaderboardData{}
    for i := 1; i <= 8; i++ {
        name := "Joueur" + strconv.Itoa(i)
        data.Players = append(data.Players, PlayerRow{
            Username:    name,
            DisplayName: name,
            Avatar:      pickAvatarForName(name),
            Rating:      1200 - i*5,
            GamesPlayed: 10 + i,
            Wins:        5 + i%3,
            Losses:      4 + (i%2),
            Draws:       1,
        })
    }

    log.Printf("leaderboard: rendering %d players", len(data.Players))
    if err := t.Execute(w, data); err != nil {
        log.Printf("leaderboard render error: %v", err)
        http.Error(w, "render error: please check server logs", http.StatusInternalServerError)
        return
    }
}

// dataFromUserID is a helper that attempts to get numeric user ID from repo by username.
func dataFromUserID(r Repository, username string) int {
    if r == nil { return 0 }
    u, _ := r.GetByUsername(context.Background(), username)
    if u == nil { return 0 }
    return u.ID
}

// pickAvatar returns a static avatar path based on user ID.
func pickAvatar(id int) string {
    // we have avatar1..avatar12 in static/avatars
    n := (id % 12)
    if n <= 0 { n = 1 }
    return "/static/avatars/avatar" + strconv.Itoa(n) + ".png"
}

// pickAvatarForName picks an avatar deterministically from username
func pickAvatarForName(name string) string {
    if name == "" { return "/static/avatars/avatar1.png" }
    h := 0
    for i := 0; i < len(name); i++ { h = (h*31 + int(name[i])) }
    n := (h % 12)
    if n < 0 { n = -n }
    if n == 0 { n = 1 }
    return "/static/avatars/avatar" + strconv.Itoa(n) + ".png"
}
