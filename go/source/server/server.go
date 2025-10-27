package server

import (
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"power4/go/game"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	mu        sync.Mutex
	g         *game.Game
	tpls      *template.Template
	boardTmpl string
	mux       *http.ServeMux
	phpBase   string
}

/* ------------------------ Construction ------------------------ */

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

	root := os.Getenv("APP_ROOT")
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	tplDir := filepath.Join(root, "templates")

	if _, err := os.Stat(tplDir); os.IsNotExist(err) {
		alt := filepath.Join(root, "go", "templates")
		if _, err2 := os.Stat(alt); err2 == nil {
			tplDir = alt
		}
	}
	log.Printf("Templates dir: %s", tplDir)

	files, _ := filepath.Glob(filepath.Join(tplDir, "*.gohtml"))
	if len(files) == 0 {
		log.Printf("⚠️ Aucun template trouvé – fallback")
		const fallback = `{{define "layout"}}<h1>Templates manquants</h1>{{end}}`
		tpls := template.Must(template.New("").Funcs(fm).Parse(fallback))
		return &Server{g: game.New(6, 9), tpls: tpls, boardTmpl: "board_medium", mux: http.NewServeMux()}
	}

	tpls, err := template.New("").Funcs(fm).ParseFiles(files...)
	if err != nil {
		log.Printf("⚠️ Erreur templates : %v", err)
		const fallback = `{{define "layout"}}<h1>Erreur templates</h1>{{end}}`
		tpls = template.Must(template.New("").Funcs(fm).Parse(fallback))
	}

	s := &Server{
		g:         game.New(6, 9),
		tpls:      tpls,
		boardTmpl: "board_medium",
		mux:       http.NewServeMux(),
	}
	s.g.ApplyBlocked(genBlocked(6, 9, 5))
	return s
}

/* ------------------------ Serveur + Routes ------------------------ */

func (s *Server) ListenWithPHP(addr, phpBase string) error {
	if phpBase == "" {
		phpBase = "http://localhost/power4"
	}
	s.phpBase = phpBase

	staticDir := "static"
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		if _, err2 := os.Stat(filepath.Join("go", "static")); err2 == nil {
			staticDir = filepath.Join("go", "static")
		}
	}
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	s.mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("pong"))
	})

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.phpBase+"/register.php", http.StatusSeeOther)
	})

	s.mux.HandleFunc("/game", s.safe(s.handleIndex))
	s.mux.HandleFunc("/play", s.safe(s.handlePlay))
	s.mux.HandleFunc("/reset", s.safe(s.handleReset))
	s.mux.HandleFunc("/new", s.safe(s.handleNew))

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("❌ Port %s indisponible : %v", addr, err)
		return err
	}
	log.Printf("✅ Serveur Go → http://%s (PHP=%s)", addr, s.phpBase)
	return http.Serve(ln, s.mux)
}

func (s *Server) safe(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				http.Error(w, "Erreur serveur", 500)
			}
		}()
		h(w, r)
	}
}

/* ------------------------ Handlers ------------------------ */

type viewData struct {
	Board         [][]int
	Rows, Cols    int
	CurrentPlayer int
	Winner        int
	BoardTemplate string
	Username      string
	PHPBase       string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	var user string
	if ck, err := r.Cookie("user"); err == nil {
		user = ck.Value
	}

	s.mu.Lock()
	data := viewData{
		Board:         s.g.Board,
		Rows:          s.g.Rows,
		Cols:          s.g.Cols,
		CurrentPlayer: s.g.CurrentPlayer,
		Winner:        s.g.Winner,
		BoardTemplate: s.boardTmpl,
		Username:      user,
		PHPBase:       s.phpBase,
	}
	s.mu.Unlock()

	if err := s.tpls.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/game", 303)
		return
	}
	_ = r.ParseForm()
	col, _ := strconv.Atoi(r.Form.Get("col"))
	s.mu.Lock()
	s.g.Drop(col)
	s.mu.Unlock()
	http.Redirect(w, r, "/game", 303)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.g.Reset(s.g.Rows, s.g.Cols)
	s.mu.Unlock()
	http.Redirect(w, r, "/game", 303)
}

func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")

	s.mu.Lock()
	switch size {
	case "small":
		s.boardTmpl = "board_small"
		s.g.Reset(6, 7)
	case "large":
		s.boardTmpl = "board_large"
		s.g.Reset(7, 8)
	default:
		s.boardTmpl = "board_medium"
		s.g.Reset(6, 9)
	}
	s.mu.Unlock()
	http.Redirect(w, r, "/game", 303)
}

/* ------------------------ Utils ------------------------ */

func genBlocked(rows, cols, n int) []game.Position {
	blocks := []game.Position{}
	return blocks
}
