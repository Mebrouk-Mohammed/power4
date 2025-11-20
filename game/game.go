package game

import "sync"

const (
	Empty = 0 // case vide
	P1    = 1 // joueur 1
	P2    = 2 // joueur 2
)

type Position struct{ R, C int }

type Game struct {
	Rows, Cols      int
	Board           [][]int
	CurrentPlayer   int
	Winner          int
	MoveCount       int
	InvertedGravity bool
	Mu              sync.Mutex
}

func New(rows, cols int) *Game {
	g := &Game{
		Rows:          rows,
		Cols:          cols,
		CurrentPlayer: P1,
		Winner:        0,
		MoveCount:     0,
		Board:         make([][]int, rows),
	}
	for r := range g.Board {
		g.Board[r] = make([]int, cols)
	}
	return g
}

func (g *Game) Reset(rows, cols int) {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	g.Rows = rows
	g.Cols = cols
	g.Board = make([][]int, rows)
	for r := range g.Board {
		g.Board[r] = make([]int, cols)
	}
	g.CurrentPlayer = P1
	g.Winner = 0
	g.MoveCount = 0
}

func (g *Game) Drop(col int) bool {
	g.Mu.Lock()
	defer g.Mu.Unlock()

	if col < 0 || col >= g.Cols || g.Winner != 0 {
		return false
	}

	if g.InvertedGravity {
		// Gravité inversée : les jetons tombent vers le haut
		for r := 0; r < g.Rows; r++ {
			if g.Board[r][col] == Empty {
				g.Board[r][col] = g.CurrentPlayer
				g.MoveCount++
				g.checkEnd(r, col)
				if g.Winner == 0 {
					g.CurrentPlayer = 3 - g.CurrentPlayer
				}
				return true
			}
		}
	} else {
		// Gravité normale : les jetons tombent vers le bas
		for r := g.Rows - 1; r >= 0; r-- {
			if g.Board[r][col] == Empty {
				g.Board[r][col] = g.CurrentPlayer
				g.MoveCount++
				g.checkEnd(r, col)
				if g.Winner == 0 {
					g.CurrentPlayer = 3 - g.CurrentPlayer
				}
				return true
			}
		}
	}
	return false
}

func (g *Game) checkEnd(r, c int) {
	p := g.Board[r][c]
	if p != P1 && p != P2 {
		return
	}
	if g.four(r, c, 1, 0, p) || // horizontal
		g.four(r, c, 0, 1, p) || // vertical
		g.four(r, c, 1, 1, p) || // diag ↘
		g.four(r, c, 1, -1, p) { // diag ↗
		g.Winner = p
		return
	}
	if g.MoveCount == g.Rows*g.Cols {
		g.Winner = -1
	}
}

func (g *Game) four(r, c, dr, dc, p int) bool {
	return 1+g.countDir(r, c, dr, dc, p)+g.countDir(r, c, -dr, -dc, p) >= 4
}

func (g *Game) countDir(r, c, dr, dc, p int) int {
	n := 0
	for i, j := r+dr, c+dc; i >= 0 && i < g.Rows && j >= 0 && j < g.Cols; i, j = i+dr, j+dc {
		if g.Board[i][j] != p {
			break
		}
		n++
	}
	return n
}
