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
	for r := 0; r < Rows; r++ {
		for c := 0; c < Columns; c++ {
			b.Cells[r][c] = Empty
		}
	}
	b.CurrentPlayer = Player1
	b.Winner = 0
	b.GameOver = false
	b.Message = "Joueur 1 (orange) commence"
}

func (b *Board) Drop(col int) (int, bool) {
	if b.GameOver {
		return -1, false
	}

	for r := Rows - 1; r >= 0; r-- {
		if b.Cells[r][col] == Empty {
			b.Cells[r][col] = b.CurrentPlayer

			// Vérifier victoire
			if b.checkWin(r, col) {
				b.Winner = b.CurrentPlayer
				b.GameOver = true
				if b.CurrentPlayer == Player1 {
					b.Message = "Joueur 1 (Orange) gagne !"
				} else {
					b.Message = "Joueur 2 (Mauve) gagne !"
				}
			} else if b.isBoardFull() {
				b.GameOver = true
				b.Message = "Match nul !"
			} else {
				// Changer de joueur
				if b.CurrentPlayer == Player1 {
					b.CurrentPlayer = Player2
					b.Message = "Tour: Joueur 2 (Mauve)"
				} else {
					b.CurrentPlayer = Player1
					b.Message = "Tour: Joueur 1 (Orange)"
				}
			}
			return r, true
		}
	}
	return -1, false
}

func (b *Board) checkWin(row, col int) bool {
	player := b.Cells[row][col]

	// Vérifications: horizontal, vertical, diagonales
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, dir := range directions {
		count := 1

		// Vérifier dans une direction
		for i := 1; i < 4; i++ {
			r, c := row+dir[0]*i, col+dir[1]*i
			if r >= 0 && r < Rows && c >= 0 && c < Columns && b.Cells[r][c] == player {
				count++
			} else {
				break
			}
		}

		// Vérifier dans l'autre direction
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

func stateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameBoard)
}

func dropHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Column int `json:"column"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Column < 0 || req.Column >= Columns {
		http.Error(w, "Invalid column", http.StatusBadRequest)
		return
	}

	gameBoard.Drop(req.Column)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameBoard)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	gameBoard.Reset()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameBoard)
}

// 🎮 PAGE DE JEU AVEC PLATEAU ET JETONS
func gameHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8">
  <title>Puissance 4 - Jeu</title>
  <style>
    body {
      font-family: 'Segoe UI', sans-serif;
      background: linear-gradient(to right, #ffecd2, #fcb69f);
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 20px;
      margin: 0;
      min-height: 100vh;
    }
    .game-container {
      background: white;
      padding: 20px;
      border-radius: 12px;
      box-shadow: 0 0 15px rgba(0,0,0,0.2);
      text-align: center;
    }
    .board {
      display: inline-block;
      background-image: url('/assets/picture/plateau-facile.png');
      background-size: cover;
      background-repeat: no-repeat;
      width: 490px;
      height: 420px;
      position: relative;
      margin: 20px;
    }
    .cell {
      position: absolute;
      width: 60px;
      height: 60px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .token {
      width: 50px;
      height: 50px;
      border-radius: 50%;
    }
    .token-orange {
      background-image: url('/assets/picture/jeton-orange.png');
      background-size: cover;
    }
    .token-mauve {
      background-image: url('/assets/picture/jeton-mauve.png');
      background-size: cover;
    }
    .message {
      font-size: 18px;
      font-weight: bold;
      margin: 20px 0;
      color: #333;
    }
    .controls {
      margin: 20px 0;
    }
    button {
      padding: 10px 20px;
      background-color: #fcb69f;
      color: white;
      border: none;
      border-radius: 6px;
      font-weight: bold;
      cursor: pointer;
      margin: 0 10px;
    }
    button:hover {
      background-color: #ff9a8b;
    }
    a {
      display: inline-block;
      padding: 10px 20px;
      background-color: #ccc;
      color: white;
      text-decoration: none;
      border-radius: 6px;
      margin: 0 10px;
    }
  </style>
</head>
<body>
  <div class="game-container">
    <h1>🎮 Puissance 4</h1>
    <div class="message" id="message">Chargement...</div>
    <div class="board" id="board"></div>
    <div class="controls">
      <button onclick="resetGame()">Nouvelle Partie</button>
      <a href="/">Retour à l'accueil</a>
    </div>
  </div>

  <script>
    let gameState = null;

    function createBoard() {
      const board = document.getElementById('board');
      board.innerHTML = '';
      
      for (let row = 0; row < 6; row++) {
        for (let col = 0; col < 7; col++) {
          const cell = document.createElement('div');
          cell.className = 'cell';
          cell.style.left = (25 + col * 70) + 'px';
          cell.style.top = (25 + row * 70) + 'px';
          cell.onclick = () => dropToken(col);
          
          if (gameState && gameState.cells[row][col] !== 0) {
            const token = document.createElement('div');
            token.className = 'token ' + (gameState.cells[row][col] === 1 ? 'token-orange' : 'token-mauve');
            cell.appendChild(token);
          }
          
          board.appendChild(cell);
        }
      }
    }

    function updateBoard() {
      fetch('/api/state')
        .then(response => response.json())
        .then(data => {
          gameState = data;
          createBoard();
          document.getElementById('message').textContent = data.message;
        });
    }

    function dropToken(column) {
      if (gameState && gameState.gameOver) return;
      
      fetch('/api/drop', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({column: column})
      })
      .then(response => response.json())
      .then(data => {
        gameState = data;
        createBoard();
        document.getElementById('message').textContent = data.message;
      });
    }

    function resetGame() {
      fetch('/api/reset', {method: 'POST'})
        .then(response => response.json())
        .then(data => {
          gameState = data;
          createBoard();
          document.getElementById('message').textContent = data.message;
        });
    }

    // Initialiser le jeu
    updateBoard();
    setInterval(updateBoard, 1000);
  </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tmpl)
}

// 🖼️ FONCTION POUR AFFICHER LA PAGE WEB AVEC VOS IMAGES
func webGameHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8">
  <title>Accueil Puissance 4</title>
  <style>
    body {
      font-family: 'Segoe UI', sans-serif;
      background: linear-gradient(to right, #ffecd2, #fcb69f);
      display: flex;
      justify-content: center;
      align-items: center;
      height: 100vh;
      margin: 0;
    }
    .container {
      background: white;
      padding: 40px;
      border-radius: 12px;
      box-shadow: 0 0 15px rgba(0,0,0,0.2);
      text-align: center;
      width: 400px;
    }
    h1 {
      margin-bottom: 20px;
      color: #333;
    }
    .plateau-preview {
      width: 300px;
      height: 250px;
      background-image: url('/assets/picture/plateau-facile.png');
      background-size: contain;
      background-repeat: no-repeat;
      background-position: center;
      margin: 20px auto;
      border-radius: 8px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }
    a {
      display: inline-block;
      margin: 10px;
      padding: 10px 20px;
      background-color: #fcb69f;
      color: white;
      text-decoration: none;
      border-radius: 6px;
      font-weight: bold;
      transition: background-color 0.3s ease;
    }
    a:hover {
      background-color: #ff9a8b;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Bienvenue sur notre Puissance 4 en ligne !</h1>
    <div class="plateau-preview"></div>
    <p>Alignez 4 jetons de votre couleur pour gagner !</p>
    <a href="/game">Lancer une partie</a>
  </div>
</html>`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tmpl)
}

func main() {
	// Initialiser le jeu
	gameBoard.Reset()

	// 🌐 PAGE PRINCIPALE AVEC VOS IMAGES
	http.HandleFunc("/", webGameHandler)

	// 🎮 PAGE DE JEU AVEC PLATEAU ET JETONS
	http.HandleFunc("/game", gameHandler)

	// 🖼️ SERVIR VOS IMAGES DEPUIS LE DOSSIER assets/
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../../assets/"))))

	// APIs du jeu
	http.HandleFunc("/api/state", stateHandler)
	http.HandleFunc("/api/drop", dropHandler)
	http.HandleFunc("/api/reset", resetHandler)

	fmt.Println("🎮 Serveur Puissance 4 démarré sur http://localhost:8080")
	fmt.Println("📱 Ouvrez votre navigateur à cette adresse pour jouer !")
	fmt.Println("🖼️ Images servies depuis : ../../assets/picture/")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
