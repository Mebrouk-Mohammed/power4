// Package game : moteur Puissance 4 (taille variable, cases bloquées possibles).
package game

import "sync"

// --- Constantes / types ---

const (
	Empty   = 0  // case vide
	P1      = 1  // joueur 1
	P2      = 2  // joueur 2
	Blocked = -1 // case bloquée (indisponible)
)

// Position désigne une case (ligne, colonne).
type Position struct{ R, C int }

// Game contient l'état complet d'une partie.
type Game struct {
	Rows, Cols    int
	Board         [][]int // 0=vide, 1=P1, 2=P2, -1=bloquée
	CurrentPlayer int     // 1 ou 2
	Winner        int     // 0=en cours, 1/2=gagnant, -1=égalité
	MoveCount     int

	mu sync.Mutex
}

// New crée une partie vide (sans cases bloquées).
func New(rows, cols int) *Game {
	g := &Game{
		Rows:          rows,
		Cols:          cols,
		CurrentPlayer: P1,
		Winner:        0,
		MoveCount:     0,
		Board:         make([][]int, rows),
	}
	for r := 0; r < rows; r++ {
		g.Board[r] = make([]int, cols)
	}
	return g
}

// Reset remet à zéro la partie avec (éventuellement) une nouvelle taille.
// ⚠️ Ne PAS copier la struct (pour ne pas copier l'état du mutex).
func (g *Game) Reset(rows, cols int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.Rows = rows
	g.Cols = cols

	g.Board = make([][]int, rows)
	for r := 0; r < rows; r++ {
		g.Board[r] = make([]int, cols)
	}

	g.CurrentPlayer = P1
	g.Winner = 0
	g.MoveCount = 0
}

// ApplyBlocked marque des cases comme indisponibles (si elles sont vides).
func (g *Game) ApplyBlocked(blocked []Position) {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, p := range blocked {
		if p.R >= 0 && p.R < g.Rows && p.C >= 0 && p.C < g.Cols {
			if g.Board[p.R][p.C] == Empty {
				g.Board[p.R][p.C] = Blocked
			}
		}
	}
}

// Drop joue un pion dans la colonne col (0-index). Retourne true si OK.
// Ignore si colonne invalide, pleine ou si la partie est terminée.
func (g *Game) Drop(col int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if col < 0 || col >= g.Cols || g.Winner != 0 {
		return false
	}
	for r := g.Rows - 1; r >= 0; r-- {
		if g.Board[r][col] == Empty {
			g.Board[r][col] = g.CurrentPlayer
			g.MoveCount++
			g.checkEndLocked(r, col)
			if g.Winner == 0 {
				g.CurrentPlayer = 3 - g.CurrentPlayer // alterne 1<->2
			}
			return true
		}
		// si c'est Blocked on remonte simplement
	}
	return false // colonne pleine (ou bloquée jusqu'en haut)
}

// --- Détection de fin (attend le verrou détenu) ---

func (g *Game) checkEndLocked(r, c int) {
	p := g.Board[r][c]
	if p != P1 && p != P2 {
		return
	}
	if g.fourLocked(r, c, 1, 0, p) || // horizontal
		g.fourLocked(r, c, 0, 1, p) || // vertical
		g.fourLocked(r, c, 1, 1, p) || // diag ↘
		g.fourLocked(r, c, 1, -1, p) { // diag ↗
		g.Winner = p
		return
	}
	if g.MoveCount == g.Rows*g.Cols-g.countBlockedLocked() {
		g.Winner = -1 // égalité en tenant compte des cases bloquées
	}
}

func (g *Game) fourLocked(r, c, dr, dc, p int) bool {
	cnt := 1 + g.countDirLocked(r, c, dr, dc, p) + g.countDirLocked(r, c, -dr, -dc, p)
	return cnt >= 4
}

func (g *Game) countDirLocked(r, c, dr, dc, p int) int {
	n, i, j := 0, r+dr, c+dc
	for i >= 0 && i < g.Rows && j >= 0 && j < g.Cols {
		if g.Board[i][j] != p {
			break
		}
		n++
		i += dr
		j += dc
	}
	return n
}

func (g *Game) countBlockedLocked() int {
	nb := 0
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			if g.Board[r][c] == Blocked {
				nb++
			}
		}
	}
	return nb
}
