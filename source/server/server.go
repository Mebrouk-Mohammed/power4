package server

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"power4/game"
)

type viewData struct {
	Board           [][]int
	Rows, Cols      int
	CurrentPlayer   int
	Winner          int
	BoardTemplate   string
	Debug           bool // ← pour le mode debug d'alignement
	InvertedGravity bool // ← pour le mode gravité inversée
}

// Server : état partagé (jeu) + templates + config d'affichage
type Server struct {
	mu        sync.Mutex
	g         *game.Game
	tpls      *template.Template
	boardTmpl string // "board_small" | "board_medium" | "board_large"
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
	))

	g := game.New(6, 9) // medium par défaut
	g.ApplyBlocked(genBlocked(6, 9, 5))

	return &Server{
		g:         g,
		tpls:      tpls,
		boardTmpl: "board_medium",
	}
}

func (s *Server) Listen(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", safe(s.handleIndex))
	mux.HandleFunc("/play", safe(s.handlePlay))
	mux.HandleFunc("/reset", safe(s.handleReset))
	mux.HandleFunc("/new", safe(s.handleNew))
	mux.HandleFunc("/gravity", safe(s.handleGravity))

	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func safe(h http.HandlerFunc) http.HandlerFunc {
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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	v := viewData{
		Board:           s.g.Board,
		Rows:            s.g.Rows,
		Cols:            s.g.Cols,
		CurrentPlayer:   s.g.CurrentPlayer,
		Winner:          s.g.Winner,
		BoardTemplate:   s.boardTmpl,
		InvertedGravity: s.g.InvertedGravity,
	}
	s.mu.Unlock()

	// active le mode debug si /?debug=1
	v.Debug = (r.URL.Query().Get("debug") == "1")

	if err := s.tpls.ExecuteTemplate(w, "layout", v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
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
	_ = s.g.Drop(col) // ignore si colonne pleine/terminée
	s.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	rows, cols := s.g.Rows, s.g.Cols
	s.g.Reset(rows, cols)
	// remettre des blocs selon le plateau courant
	switch s.boardTmpl {
	case "board_small":
		s.g.ApplyBlocked(genBlocked(rows, cols, 3))
	case "board_large":
		s.g.ApplyBlocked(genBlocked(rows, cols, 7))
	default:
		s.g.ApplyBlocked(genBlocked(rows, cols, 5))
	}
	s.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// /new?size=small|medium|large
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")
	log.Println("Switch difficulty →", size)

	var rows, cols, nbBlocked int
	var tmpl string

	switch size {
	case "small": // Easy : 6x7, 3 cases bloquées
		rows, cols = 6, 7
		nbBlocked = 3
		tmpl = "board_small"
	case "large": // Hard : 7x8, 7 cases bloquées
		rows, cols = 7, 8
		nbBlocked = 7
		tmpl = "board_large"
	default: // Medium/Normal : 6x9, 5 cases bloquées
		rows, cols = 6, 9
		nbBlocked = 5
		tmpl = "board_medium"
	}

	s.mu.Lock()
	s.g.Reset(rows, cols)
	s.g.ApplyBlocked(genBlocked(rows, cols, nbBlocked))
	s.boardTmpl = tmpl
	s.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// genBlocked choisit 'n' cases à bloquer aléatoirement
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
		k := rand.Intn(max) // 0..rows*cols-1
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

func (s *Server) handleGravity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	inverted := r.URL.Query().Get("inverted") == "true"
	
	s.mu.Lock()
	s.g.InvertedGravity = inverted
	s.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}