package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	Rows    = 6
	Columns = 7
	Empty   = 0
	Player1 = 1
	Player2 = 2
)

type Board struct {
	Cells         [Rows][Columns]int `json:"cells"`
	CurrentPlayer int                `json:"currentPlayer"`
	Winner        int                `json:"winner"`
	GameOver      bool               `json:"gameOver"`
	Message       string             `json:"message"`
}

var gameBoard Board

func (b *Board) Reset() {
	for r := range b.Cells {
		for c := range b.Cells[r] {
			b.Cells[r][c] = Empty
		}
	}
	b.CurrentPlayer = Player1
	b.Winner = 0
	b.GameOver = false
	b.Message = "Joueur 1 (orange) commence"
}

func (b *Board) Drop(col int) (int, bool) {
	if b.GameOver || col < 0 || col >= Columns {
		return -1, false
	}

	for r := Rows - 1; r >= 0; r-- {
		if b.Cells[r][col] == Empty {
			b.Cells[r][col] = b.CurrentPlayer
			if b.checkWin(r, col) {
				b.Winner = b.CurrentPlayer
				b.GameOver = true
				b.Message = fmt.Sprintf("Joueur %d (%s) gagne !", b.CurrentPlayer, b.color())
			} else if b.isBoardFull() {
				b.GameOver = true
				b.Message = "Match nul !"
			} else {
				b.switchPlayer()
			}
			return r, true
		}
	}
	return -1, false
}

func (b *Board) switchPlayer() {
	if b.CurrentPlayer == Player1 {
		b.CurrentPlayer = Player2
		b.Message = "Tour: Joueur 2 (Mauve)"
	} else {
		b.CurrentPlayer = Player1
		b.Message = "Tour: Joueur 1 (Orange)"
	}
}

func (b *Board) color() string {
	if b.CurrentPlayer == Player1 {
		return "Orange"
	}
	return "Mauve"
}

func (b *Board) checkWin(row, col int) bool {
	player := b.Cells[row][col]
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, dir := range directions {
		count := 1
		for i := 1; i < 4; i++ {
			r, c := row+dir[0]*i, col+dir[1]*i
			if r >= 0 && r < Rows && c >= 0 && c < Columns && b.Cells[r][c] == player {
				count++
			} else {
				break
			}
		}
		for i := 1; i < 4; i++ {
			r, c := row-dir[0]*i, col-dir[1]*i
			if r >= 0 && r < Rows && c >= 0 && c < Columns && b.Cells[r][c] == player {
				count++
			} else {
				break
			}
		}
		if count >= 4 {
			return true
		}
	}
	return false
}

func (b *Board) isBoardFull() bool {
	for c := 0; c < Columns; c++ {
		if b.Cells[0][c] == Empty {
			return false
		}
	}
	return true
}

// --- API Handlers ---

func stateHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, gameBoard)
}

func dropHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Column int `json:"column"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	gameBoard.Drop(req.Column)
	jsonResponse(w, gameBoard)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	gameBoard.Reset()
	jsonResponse(w, gameBoard)
}

func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// --- HTML Pages ---

func gamePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/game.html")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/home.html")
}

// --- Main ---

func main() {
	gameBoard.Reset()

	http.HandleFunc("/", homePage)
	http.HandleFunc("/game", gamePage)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/api/state", stateHandler)
	http.HandleFunc("/api/drop", dropHandler)
	http.HandleFunc("/api/reset", resetHandler)

	fmt.Println("🎮 Serveur Puissance 4 lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
