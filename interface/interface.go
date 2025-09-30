package main

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	Rows    = 6
	Columns = 7
	Empty   = 0
	Player1 = 1
	Player2 = 2
)

type Board struct {
	cells [Rows][Columns]int
}

func (b *Board) Reset() {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Columns; c++ {
			b.cells[r][c] = Empty
		}
	}
}
func (b *Board) Drop(col, player int) int {
	for r := Rows - 1; r >= 0; r-- {
		if b.cells[r][col] == Empty {
			b.cells[r][col] = player
			return r
		}
	}
	return -1
}
func main() {
	rand.Seed(time.Now().UnixNano())
	myApp := app.New()
	w := myApp.NewWindow("Puissance 4")

	var board Board
	board.Reset()

	// Charger les images
	emptyImg := "assets/plateau.png"
	redImg := "assets/jeton-orange.png"
	yellowImg := "jeton-mauve.png"

	var imgs [Rows][Columns]*canvas.Image

	grid := container.NewGridWithColumns(Columns)
	for r := 0; r < Rows; r++ {
		for c := 0; c < Columns; c++ {
			img := canvas.NewImageFromFile(emptyImg)
			img.FillMode = canvas.ImageFillContain
			imgs[r][c] = img
			grid.Add(img)
		}
	}

	// Helper to refresh the board images based on the board state
	refresh := func() {
		for r := 0; r < Rows; r++ {
			for c := 0; c < Columns; c++ {
				switch board.cells[r][c] {
				case Player1:
					imgs[r][c].File = redImg
				case Player2:
					imgs[r][c].File = yellowImg
				default:
					imgs[r][c].File = emptyImg
				}
				imgs[r][c].Refresh()
			}
		}
	}

	status := widget.NewLabel("Joueur 1 (orange) commence")
	current := Player1
	topRow := container.NewHBox()
	for c := 0; c < Columns; c++ {
		col := c
		btn := widget.NewButton(fmt.Sprintf("↓ %d", c+1), func() {
			row := board.Drop(col, current)
			if row == -1 {
				status.SetText("Colonne pleine !")
				return
			}
			refresh()
			if current == Player1 {
				current = Player2
				status.SetText("Tour: Joueur 2 (mauve)")
			} else {
				current = Player1
				status.SetText("Tour: Joueur 1 (orange)")
			}
		})
		topRow.Add(btn)
	}

	// Set the window content and show the window
	content := container.NewVBox(
		topRow,
		grid,
		status,
	)
	w.SetContent(content)
	w.ShowAndRun()
}
