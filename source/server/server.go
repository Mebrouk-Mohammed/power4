package server

import (
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"

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
	// websocket clients (simple hub)
	wsUpgrader websocket.Upgrader
	wsClients  map[*websocket.Conn]bool
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

	return &Server{
		g:         g,
		tpls:      tpls,
		boardTmpl: "board_medium",
		wsUpgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		wsClients:  make(map[*websocket.Conn]bool),
	}
}

func (s *Server) Listen(addr string) error {
	mux := http.NewServeMux()
	// Serve static files (images, css, js) from the "static" directory
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// websocket endpoint
	mux.HandleFunc("/ws", safe(s.handleWS))
	mux.HandleFunc("/", safe(s.handleIndex))
	mux.HandleFunc("/play", safe(s.handlePlay))
	mux.HandleFunc("/random_move", safe(s.handleRandomMove))
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

	// broadcast to websocket clients that state changed
	go s.broadcastState()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleRandomMove choisit une colonne au hasard (parmi les colonnes non pleines)
// et y pose le jeton courant. Utilisé quand le timer côté client expire.
func (s *Server) handleRandomMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.g.Winner != 0 {
		http.Error(w, "game over", http.StatusConflict)
		return
	}

	// construire la liste des colonnes disponibles
	avail := make([]int, 0)
	for c := 0; c < s.g.Cols; c++ {
		if s.g.Board[0][c] == game.Empty {
			avail = append(avail, c)
		}
	}
	if len(avail) == 0 {
		http.Error(w, "board full", http.StatusConflict)
		return
	}

	col := avail[rand.Intn(len(avail))]
	_ = s.g.Drop(col)

	// notifier clients
	go s.broadcastState()

	// Répondre avec l'état minimal pour le client (ok:true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	rows, cols := s.g.Rows, s.g.Cols
	s.g.Reset(rows, cols)
	s.mu.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// /new?size=small|medium|large
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")
	log.Println("Switch difficulty →", size)

	var rows, cols int
	var tmpl string

	switch size {
	case "small": // Easy : 6x7
		rows, cols = 6, 7
		tmpl = "board_small"
	case "large": // Hard : 7x8
		rows, cols = 7, 8
		tmpl = "board_large"
	default: // Medium/Normal : 6x9
		rows, cols = 6, 9
		tmpl = "board_medium"
	}

	s.mu.Lock()
	s.g.Reset(rows, cols)
	s.boardTmpl = tmpl
	s.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
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

// WebSocket handler: upgrade and register client, send initial state and keep connection alive
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws upgrade:", err)
		return
	}

	s.mu.Lock()
	s.wsClients[conn] = true
	s.mu.Unlock()

	// send initial state
	s.sendStateToConn(conn)

	// read loop — we don't expect messages, but detect close
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}

	// cleanup
	s.mu.Lock()
	delete(s.wsClients, conn)
	s.mu.Unlock()
	conn.Close()
}

// sendStateToConn sends current game state as JSON to a single connection
func (s *Server) sendStateToConn(conn *websocket.Conn) {
	s.mu.Lock()
	payload := map[string]interface{}{
		"board":         s.g.Board,
		"rows":          s.g.Rows,
		"cols":          s.g.Cols,
		"currentPlayer": s.g.CurrentPlayer,
		"winner":        s.g.Winner,
		"boardTemplate": s.boardTmpl,
	}
	s.mu.Unlock()

	b, _ := json.Marshal(payload)
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
		log.Println("ws write error:", err)
	}
}

// broadcastState sends the current game state JSON to all connected websocket clients
func (s *Server) broadcastState() {
	s.mu.Lock()
	payload := map[string]interface{}{
		"board":         s.g.Board,
		"rows":          s.g.Rows,
		"cols":          s.g.Cols,
		"currentPlayer": s.g.CurrentPlayer,
		"winner":        s.g.Winner,
		"boardTemplate": s.boardTmpl,
	}
	clients := make([]*websocket.Conn, 0, len(s.wsClients))
	for c := range s.wsClients {
		clients = append(clients, c)
	}
	s.mu.Unlock()

	b, _ := json.Marshal(payload)
	for _, c := range clients {
		c.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if err := c.WriteMessage(websocket.TextMessage, b); err != nil {
			// remove broken client
			c.Close()
			s.mu.Lock()
			delete(s.wsClients, c)
			s.mu.Unlock()
		}
	}
}
