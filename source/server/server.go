package server // Déclare le package server

import (
	"encoding/json" // Pour encoder/décoder des données en JSON
	"html/template" // Pour gérer les templates HTML
	"log"           // Pour les logs serveur
	"math/rand"     // Pour les tirages aléatoires (coup random)
	"net/http"      // Pour le serveur HTTP
	"strconv"       // Pour convertir des chaînes en int
	"sync"          // Pour utiliser des verrous (mutex)
	"time"          // Pour initialiser le random avec l'heure

	"power4/game" // Importe le package game qui gère la logique du Puissance 4
)

// viewData contient toutes les données envoyées au template HTML
type viewData struct {
	Board           [][]int // Plateau de jeu (matrice de cases)
	Rows, Cols      int     // Nombre de lignes et colonnes
	CurrentPlayer   int     // Joueur courant (1 ou 2)
	Winner          int     // Vainqueur (0 = en cours, -1 = nul, 1 ou 2 si gagnant)
	BoardTemplate   string  // Nom du template de plateau à utiliser (small/medium/large)
	Debug           bool    // Mode debug activé ou non (affichage alignements)
	InvertedGravity bool    // Mode gravité inversée activé ou non
}

// Server regroupe l'état du jeu, les templates et la configuration
type Server struct {
	mu        sync.Mutex         // Mutex pour protéger l'accès concurrent à g et aux données
	g         *game.Game         // Pointeur vers l'objet Game (logique du jeu)
	tpls      *template.Template // Ensemble de templates HTML pré-compilés
	boardTmpl string             // Nom du template de plateau utilisé ("board_small" | "board_medium" | "board_large")
}

// NewDefault crée un serveur avec une configuration par défaut
func NewDefault() *Server {
	rand.Seed(time.Now().UnixNano()) // Initialise le générateur aléatoire avec l'heure actuelle

	// fm contient des fonctions utilitaires utilisables dans les templates HTML
	fm := template.FuncMap{
		"rangeN": func(n int) []int { // Fonction rangeN pour générer un slice [0..n-1]
			out := make([]int, n) // Crée un slice de n éléments
			for i := 0; i < n; i++ {
				out[i] = i // Remplit avec 0,1,2,...,n-1
			}
			return out
		},
	}

	// Charge les fichiers de templates et y associe les fonctions définies dans fm
	tpls := template.Must(template.New("").Funcs(fm).ParseFiles(
		"templates/layout.gohtml",       // Template principal (layout)
		"templates/board_small.gohtml",  // Template plateau version "small"
		"templates/board_medium.gohtml", // Template plateau version "medium"
		"templates/board_large.gohtml",  // Template plateau version "large"
		"templates/token_p1.gohtml",     // Template pour le jeton du joueur 1
		"templates/token_p2.gohtml",     // Template pour le jeton du joueur 2
	))

	// Crée un nouveau jeu avec une grille 6x9 (difficulté medium par défaut)
	g := game.New(6, 9)

	// Retourne une nouvelle instance de Server avec ces réglages par défaut
	return &Server{
		g:         g,              // Logique de jeu
		tpls:      tpls,           // Templates compilés
		boardTmpl: "board_medium", // Plateau par défaut
	}
}

// Listen démarre le serveur HTTP sur l'adresse donnée
func (s *Server) Listen(addr string) error {
	mux := http.NewServeMux() // Crée un nouveau multiplexer de routes

	// Sert les fichiers statiques (images, CSS, JS) depuis le dossier "static"
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Route pour la page d'accueil principale
	mux.HandleFunc("/", safe(s.handleIndex))

	// Route pour jouer un coup (via formulaire POST)
	mux.HandleFunc("/play", safe(s.handlePlay))

	// Route pour jouer un coup aléatoire (timer ou bot simple)
	mux.HandleFunc("/random_move", safe(s.handleRandomMove))

	// Route pour réinitialiser la partie
	mux.HandleFunc("/reset", safe(s.handleReset))

	// Route pour créer une nouvelle partie avec une autre taille de plateau
	mux.HandleFunc("/new", safe(s.handleNew))

	// Route pour activer/désactiver la gravité inversée
	mux.HandleFunc("/gravity", safe(s.handleGravity))

	log.Printf("Server listening on %s", addr) // Log l'adresse d'écoute
	return http.ListenAndServe(addr, mux)      // Lance le serveur HTTP et retourne une erreur si échec
}

// safe enveloppe un handler pour capturer les panics et éviter un crash du serveur
func safe(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Fonction de récupération d'un panic éventuel
		defer func() {
			if rec := recover(); rec != nil { // Si un panic est survenu
				log.Println("PANIC:", rec)                                             // Log l'erreur
				http.Error(w, "Internal server error", http.StatusInternalServerError) // Réponse HTTP 500
			}
		}()
		h(w, r) // Exécute le handler réel
	}
}

// handleIndex affiche la page principale avec le plateau de jeu
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock() // Verrouille l'accès aux données du serveur
	v := viewData{
		Board:           s.g.Board,           // Plateau actuel
		Rows:            s.g.Rows,            // Nombre de lignes
		Cols:            s.g.Cols,            // Nombre de colonnes
		CurrentPlayer:   s.g.CurrentPlayer,   // Joueur courant
		Winner:          s.g.Winner,          // Vainqueur éventuel
		BoardTemplate:   s.boardTmpl,         // Nom du template de plateau à utiliser
		InvertedGravity: s.g.InvertedGravity, // État du mode gravité inversée
	}
	s.mu.Unlock() // Déverrouille

	// Active le mode debug si l'URL contient ?debug=1
	v.Debug = (r.URL.Query().Get("debug") == "1")

	// Exécute le template "layout" avec les données v
	if err := s.tpls.ExecuteTemplate(w, "layout", v); err != nil {
		// En cas d'erreur de template, renvoie une erreur HTTP 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handlePlay gère un coup joué via un formulaire POST
func (s *Server) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // N'accepte que la méthode POST
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirige vers la page principale si mauvaise méthode
		return
	}
	if err := r.ParseForm(); err != nil { // Parse les champs du formulaire
		http.Error(w, "invalid form", http.StatusBadRequest) // Formulaire invalide
		return
	}
	// Récupère la colonne envoyée dans le formulaire
	col, err := strconv.Atoi(r.Form.Get("col"))
	if err != nil { // Erreur de conversion
		http.Error(w, "invalid col", http.StatusBadRequest)
		return
	}

	s.mu.Lock()       // Verrouille l'état du jeu
	_ = s.g.Drop(col) // Joue le coup dans la colonne (ignorer le retour si colonne pleine ou partie terminée)
	s.mu.Unlock()     // Déverrouille

	// Redirige vers la page principale pour afficher le nouveau plateau
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleRandomMove choisit une colonne au hasard (parmi les colonnes non pleines)
// et y joue le jeton du joueur courant. Utilisé en général par un timer côté client.
func (s *Server) handleRandomMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // N'accepte que la méthode POST
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()         // Verrouille l'accès au jeu
	defer s.mu.Unlock() // Déverrouille automatiquement à la fin de la fonction

	if s.g.Winner != 0 { // Si la partie est déjà terminée
		http.Error(w, "game over", http.StatusConflict) // Conflit : plus de coups possibles
		return
	}

	// Construit la liste des colonnes encore disponibles (non pleines)
	avail := make([]int, 0)
	for c := 0; c < s.g.Cols; c++ {
		// Si la première ligne de la colonne est vide, la colonne n'est pas pleine
		if s.g.Board[0][c] == game.Empty {
			avail = append(avail, c)
		}
	}
	if len(avail) == 0 { // Si aucune colonne disponible
		http.Error(w, "board full", http.StatusConflict) // Le plateau est plein
		return
	}

	// Choisit au hasard une colonne parmi celles disponibles
	col := avail[rand.Intn(len(avail))]
	_ = s.g.Drop(col) // Joue le coup dans cette colonne

	// Répond au client avec un JSON minimal {"ok": true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// handleReset réinitialise la partie avec la même taille de plateau
func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()                                   // Verrouille le jeu
	rows, cols := s.g.Rows, s.g.Cols              // Sauvegarde les dimensions actuelles du plateau
	s.g.Reset(rows, cols)                         // Réinitialise le jeu avec ces mêmes dimensions
	s.mu.Unlock()                                 // Déverrouille
	http.Redirect(w, r, "/", http.StatusSeeOther) // Retour à la page principale
}

// handleNew configure une nouvelle partie avec une taille de plateau différente
// URL attendue : /new?size=small|medium|large
func (s *Server) handleNew(w http.ResponseWriter, r *http.Request) {
	size := r.URL.Query().Get("size")        // Récupère le paramètre "size"
	log.Println("Switch difficulty →", size) // Log l'action de changement de difficulté

	var rows, cols int // Dimensions du plateau
	var tmpl string    // Nom du template de plateau à utiliser

	// Choisit les dimensions et le template en fonction du niveau
	switch size {
	case "small": // Easy : plateau 6x7
		rows, cols = 6, 7
		tmpl = "board_small"
	case "large": // Hard : plateau 7x8
		rows, cols = 7, 8
		tmpl = "board_large"
	default: // Medium/Normal : plateau 6x9
		rows, cols = 6, 9
		tmpl = "board_medium"
	}

	s.mu.Lock()           // Verrouille l'accès au jeu
	s.g.Reset(rows, cols) // Réinitialise le plateau avec les nouvelles dimensions
	s.boardTmpl = tmpl    // Met à jour le template de plateau à utiliser
	s.mu.Unlock()         // Déverrouille

	// Redirige vers la page principale pour afficher le nouveau plateau
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleGravity active ou désactive la gravité inversée
func (s *Server) handleGravity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { // On n'accepte que GET
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirige si mauvaise méthode
		return
	}

	// Récupère la valeur du paramètre ?inverted=true|false
	inverted := r.URL.Query().Get("inverted") == "true"

	s.mu.Lock()                    // Verrouille le jeu
	s.g.InvertedGravity = inverted // Met à jour l'état de la gravité
	s.mu.Unlock()                  // Déverrouille

	// Retour à la page principale pour refléter le changement
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
