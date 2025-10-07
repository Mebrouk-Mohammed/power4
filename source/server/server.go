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

// Server : état partagé (jeu) + templates + config d'affichage
type Server struct {
	mu        sync.Mutex
	g         *game.Game
	tpls      *template.Template
	boardTmpl string // "board_small" | "board_medium" | "board_large"
	autoPlayActive bool // true si le jeton auto est en cours
}

// NewDefault initialise le serveur avec le plateau medium
func NewDefault() *Server {
	rand.Seed(time.Now().UnixNano())

	funcs := template.FuncMap{
		"rangeN": func(n int) []int {
			out := make([]int, n)
			for i := range out {
				out[i] = i
			}
			return out
		},
	}

	tpls := template.Must(template.New("").Funcs(funcs).ParseFiles(
		"templates/layout.gohtml",
		"templates/board_small.gohtml",
		"templates/board_medium.gohtml",
		"templates/board_large.gohtml",
		"templates/token_p1.gohtml",
		"templates/token_p2.gohtml",
	))

	s := &Server{
		g:         game.New(6, 9),
		tpls:      tpls,
		boardTmpl: "board_medium",
		autoPlayActive: false,
	}
	s.g.ApplyBlocked(genBlocked(6, 9, 5))
	return s
}

// Listen démarre le serveur HTTP
func (s *Server) Listen(addr string) error {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", s.safe(s.handleIndex))
	http.HandleFunc("/play", s.safe(s.handlePlay))
	http.HandleFunc("/reset", s.safe(s.handleReset))
	http.HandleFunc("/new", s.safe(s.handleNew))

	log.Printf("✅ Serveur Power4 lancé sur http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// safe protège les handlers contre les panic
func (s *Server) safe(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Println("🔥 PANIC:", rec)
				http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

// viewData : données passées au template
type viewData struct {
	   Board         [][]int
	   Rows, Cols    int
	   CurrentPlayer int
	   Winner        int
	   BoardTemplate string
	   Debug         bool
	   AutoPlayActive bool
}

// handleIndex : page principale du jeu
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	   v := viewData{
		   Board:         s.g.Board,
		   Rows:          s.g.Rows,
		   Cols:          s.g.Cols,
		   CurrentPlayer: s.g.CurrentPlayer,
		   Winner:        s.g.Winner,
		   BoardTemplate: s.boardTmpl,
		   Debug:         r.URL.Query().Get("debug") == "1",
		   AutoPlayActive: s.autoPlayActive,
	   }
	s.mu.Unlock()

	if err := s.tpls.ExecuteTemplate(w, "layout", v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handlePlay : joue un coup
func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
   if r.Method != http.MethodPost {
	   http.Redirect(w, r, "/", http.StatusSeeOther)
	   return
   }
   if err := r.ParseForm(); err != nil {
	   http.Error(w, "Formulaire invalide", http.StatusBadRequest)
	   return
   }

   colStr := r.Form.Get("col")
   var col int
   var err error
   if colStr != "" {
	   col, err = strconv.Atoi(colStr)
	   if err != nil {
		   http.Error(w, "Colonne invalide", http.StatusBadRequest)
		   return
	   }
   } else {
	   // Si aucune colonne n'est fournie, attendre 10 secondes puis jouer automatiquement
	   timer := time.NewTimer(10 * time.Second)
	   played := make(chan bool, 1)

	   go func() {
		   <-timer.C
		   s.mu.Lock()
		   // Cherche les colonnes valides
		   validCols := []int{}
		   for c := 0; c < s.g.Cols; c++ {
			   for r := s.g.Rows - 1; r >= 0; r-- {
				   if s.g.Board[r][c] == 0 {
					   validCols = append(validCols, c)
					   break
				   }
			   }
		   }
		   if len(validCols) > 0 {
			   randCol := validCols[rand.Intn(len(validCols))]
			   _ = s.g.Drop(randCol)
		   }
		   s.mu.Unlock()
		   played <- true
	   }()

	   <-played
	   http.Redirect(w, r, "/", http.StatusSeeOther)
	   return
   }

   s.mu.Lock()
   _ = s.g.Drop(col)
   s.mu.Unlock()

   http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleReset : réinitialise la partie
func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleNew : change la taille du plateau
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")
	log.Println("🔁 Changement de difficulté →", size)

	var rows, cols, nbBlocked int
	var tmpl string

	switch size {
	case "small":
		rows, cols = 6, 7
		nbBlocked = 3
		tmpl = "board_small"
	case "large":
		rows, cols = 7, 8
		nbBlocked = 7
		tmpl = "board_large"
	default:
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

// genBlocked : génère n cases bloquées aléatoires
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
