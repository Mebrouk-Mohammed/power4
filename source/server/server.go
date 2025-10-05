package server

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"power4/auth"
	"power4/game"
)

type Server struct {
	mu        sync.Mutex
	g         *game.Game
	tpls      *template.Template
	boardTmpl string
}

func NewDefault() *Server {
	rand.Seed(time.Now().UnixNano())

	fm := template.FuncMap{
		"rangeN": func(n int) []int {
			out := make([]int, n)
			for i := 0; i < n; i++ {
				out[i] = i
			}
			return out
		},
	}

	tpls := template.Must(template.New("").Funcs(fm).ParseFiles(
		"templates/layout.gohtml",
		"templates/board_small.gohtml",
		"templates/board_medium.gohtml",
		"templates/board_large.gohtml",
		"templates/token_p1.gohtml",
		"templates/token_p2.gohtml",
		"templates/auth_menu.gohtml", // menu Auth
	))

	// Démarrage : Medium/Normal (6x9 + 5 blocs)
	s := &Server{
		g:         game.New(6, 9),
		tpls:      tpls,
		boardTmpl: "board_medium",
	}
	s.g.ApplyBlocked(genBlocked(6, 9, 5))
	return s
}

func (s *Server) Listen(addr string) error {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth / Accueil
	http.HandleFunc("/", s.safe(s.handleAuthMenu)) // ⬅️ page d'accueil = menu compte
	http.HandleFunc("/auth", s.safe(s.handleAuthMenu))
	http.HandleFunc("/login", auth.LoginHandler)
	http.HandleFunc("/register", auth.RegisterHandler)

	// Jeu
	http.HandleFunc("/game", s.safe(s.handleIndex)) // ⬅️ le jeu est ici maintenant
	http.HandleFunc("/play", s.safe(s.handlePlay))
	http.HandleFunc("/reset", s.safe(s.handleReset))
	http.HandleFunc("/new", s.safe(s.handleNew))

	log.Printf("Power4 Web → http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// empêche un panic d'arrêter le serveur
func (s *Server) safe(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Println("PANIC:", rec)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

type viewData struct {
	Board         [][]int
	Rows, Cols    int
	CurrentPlayer int
	Winner        int
	BoardTemplate string
	Debug         bool // utilisé par layout.gohtml
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	dbg := r.URL.Query().Get("debug") == "1"

	s.mu.Lock()
	v := viewData{
		Board:         s.g.Board,
		Rows:          s.g.Rows,
		Cols:          s.g.Cols,
		CurrentPlayer: s.g.CurrentPlayer,
		Winner:        s.g.Winner,
		BoardTemplate: s.boardTmpl,
		Debug:         dbg,
	}
	s.mu.Unlock()

	if err := s.tpls.ExecuteTemplate(w, "layout", v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	col, err := strconv.Atoi(r.Form.Get("col"))
	if err != nil {
		http.Error(w, "invalid col", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	_ = s.g.Drop(col)
	s.mu.Unlock()

	if r.URL.Query().Get("debug") == "1" {
		http.Redirect(w, r, "/game?debug=1", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	dbg := r.URL.Query().Get("debug") == "1"

	s.mu.Lock()
	rows, cols := s.g.Rows, s.g.Cols
	s.g.Reset(rows, cols)
	switch s.boardTmpl {
	case "board_small":
		s.g.ApplyBlocked(genBlocked(rows, cols, 3))
	case "board_large":
		s.g.ApplyBlocked(genBlocked(rows, cols, 7))
	default:
		s.g.ApplyBlocked(genBlocked(rows, cols, 5))
	}
	s.mu.Unlock()

	if dbg {
		http.Redirect(w, r, "/game?debug=1", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

// /new?size=small|medium|large[&debug=1]
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")
	dbg := r.URL.Query().Get("debug") == "1"

	var rows, cols, nbBlocked int
	var tmpl string
	switch size {
	case "small": // Easy
		rows, cols, nbBlocked, tmpl = 6, 7, 3, "board_small"
	case "large": // Hard
		rows, cols, nbBlocked, tmpl = 7, 8, 7, "board_large"
	default: // Medium/Normal
		rows, cols, nbBlocked, tmpl = 6, 9, 5, "board_medium"
	}

	s.mu.Lock()
	s.g.Reset(rows, cols)
	s.g.ApplyBlocked(genBlocked(rows, cols, nbBlocked))
	s.boardTmpl = tmpl
	s.mu.Unlock()

	if dbg {
		http.Redirect(w, r, "/game?debug=1", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

// Menu Auth
func (s *Server) handleAuthMenu(w http.ResponseWriter, r *http.Request) {
	if err := s.tpls.ExecuteTemplate(w, "auth_menu", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// utils
func genBlocked(rows, cols, n int) []game.Position {
	if n <= 0 {
		return nil
	}
	seen := make(map[int]struct{}, n)
	out := make([]game.Position, 0, n)
	max := rows * cols
	if n > max {
		n = max
	}
	for len(out) < n {
		k := rand.Intn(max)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		r := k / cols
		c := k % cols
		out = append(out, game.Position{R: r, C: c})
	}
	return out
}
