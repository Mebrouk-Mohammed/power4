package main // Déclare le package principal (point d'entrée du programme)

import (
	"encoding/json" // Pour encoder/décoder du JSON
	"fmt"           // Pour afficher du texte dans la console et écrire dans la réponse
	"log"           // Pour gérer les logs et erreurs serveur
	"net/http"      // Pour créer le serveur HTTP
)

const (
	Rows    = 6 // Nombre de lignes du plateau
	Columns = 7 // Nombre de colonnes du plateau
	Empty   = 0 // Case vide
	Player1 = 1 // Joueur 1
	Player2 = 2 // Joueur 2
)

// Board représente l'état du jeu côté serveur
type Board struct {
	Cells         [Rows][Columns]int `json:"cells"`         // Grille de jeu (6x7)
	CurrentPlayer int                `json:"currentPlayer"` // Joueur courant (1 ou 2)
	Winner        int                `json:"winner"`        // Gagnant (0 = aucun, 1 ou 2, ou autre si match nul)
	GameOver      bool               `json:"gameOver"`      // Indique si la partie est terminée
	Message       string             `json:"message"`       // Message d'information pour le joueur
}

var gameBoard Board // Variable globale qui contient l'état du plateau

// Reset réinitialise la partie
func (b *Board) Reset() {
	for r := 0; r < Rows; r++ { // Parcourt chaque ligne
		for c := 0; c < Columns; c++ { // Parcourt chaque colonne
			b.Cells[r][c] = Empty // Vide chaque case
		}
	}
	b.CurrentPlayer = Player1                // Le joueur 1 commence
	b.Winner = 0                             // Pas de gagnant
	b.GameOver = false                       // La partie n'est pas terminée
	b.Message = "Joueur 1 (orange) commence" // Message initial affiché au joueur
}

// Drop fait tomber un jeton dans la colonne donnée
func (b *Board) Drop(col int) (int, bool) {
	if b.GameOver { // Si la partie est déjà terminée
		return -1, false // On refuse le coup
	}

	for r := Rows - 1; r >= 0; r-- { // On part du bas et on remonte
		if b.Cells[r][col] == Empty { // Si la case est vide
			b.Cells[r][col] = b.CurrentPlayer // Place le jeton du joueur courant

			// Vérifier si ce coup provoque une victoire
			if b.checkWin(r, col) {
				b.Winner = b.CurrentPlayer // On enregistre le gagnant
				b.GameOver = true          // On marque la partie comme finie
				if b.CurrentPlayer == Player1 {
					b.Message = "Joueur 1 (Orange) gagne !" // Message de victoire joueur 1
				} else {
					b.Message = "Joueur 2 (Mauve) gagne !" // Message de victoire joueur 2
				}
			} else if b.isBoardFull() { // Si le plateau est plein
				b.GameOver = true         // La partie est finie
				b.Message = "Match nul !" // Message de match nul
			} else {
				// Sinon on change de joueur
				if b.CurrentPlayer == Player1 {
					b.CurrentPlayer = Player2            // Passage au joueur 2
					b.Message = "Tour: Joueur 2 (Mauve)" // Message de tour
				} else {
					b.CurrentPlayer = Player1             // Retour au joueur 1
					b.Message = "Tour: Joueur 1 (Orange)" // Message de tour
				}
			}
			return r, true // On renvoie la ligne où le jeton est tombé et true
		}
	}
	return -1, false // Si aucune case libre dans la colonne, coup impossible
}

// checkWin vérifie si le dernier coup joué gagne la partie
func (b *Board) checkWin(row, col int) bool {
	player := b.Cells[row][col] // Récupère le joueur ayant joué ce coup

	// Liste des directions : horizontale, verticale, diagonale ↘, diagonale ↗
	directions := [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}}

	for _, dir := range directions { // Pour chaque direction
		count := 1 // On compte déjà le jeton posé

		// Vérifier dans une direction (avant)
		for i := 1; i < 4; i++ { // On regarde au maximum 3 cases plus loin
			r, c := row+dir[0]*i, col+dir[1]*i // Nouvelle position à vérifier
			if r >= 0 && r < Rows && c >= 0 && c < Columns && b.Cells[r][c] == player {
				count++ // Même joueur trouvé, on augmente le compteur
			} else {
				break // Si ce n'est plus le même joueur ou hors plateau, on arrête
			}
		}

		// Vérifier dans l'autre direction (arrière)
		for i := 1; i < 4; i++ { // Même logique dans le sens opposé
			r, c := row-dir[0]*i, col-dir[1]*i
			if r >= 0 && r < Rows && c >= 0 && c < Columns && b.Cells[r][c] == player {
				count++ // Même joueur dans l'autre sens
			} else {
				break // On arrête si plus aligné
			}
		}

		if count >= 4 { // Si on a au moins 4 jetons alignés
			return true // Il y a une victoire
		}
	}
	return false // Aucun alignement gagnant trouvé
}

// isBoardFull vérifie si le plateau est plein
func (b *Board) isBoardFull() bool {
	for c := 0; c < Columns; c++ { // On parcourt chaque colonne
		if b.Cells[0][c] == Empty { // Si la première ligne contient une case vide
			return false // Le plateau n'est pas plein
		}
	}
	return true // Si aucune case vide en haut, le plateau est plein
}

// stateHandler renvoie l'état actuel du jeu en JSON
func stateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Indique qu'on renvoie du JSON
	json.NewEncoder(w).Encode(gameBoard)               // Encode l'état du jeu en JSON et l'envoie
}

// dropHandler reçoit une colonne en JSON, joue le coup, et renvoie le nouvel état
func dropHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Column int `json:"column"` // Structure pour récupérer la colonne envoyée par le client
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // Décode la requête JSON
		http.Error(w, "Invalid JSON", http.StatusBadRequest) // Renvoie une erreur si JSON invalide
		return
	}

	if req.Column < 0 || req.Column >= Columns { // Vérifie que la colonne demandée est valide
		http.Error(w, "Invalid column", http.StatusBadRequest) // Erreur si colonne incorrecte
		return
	}

	gameBoard.Drop(req.Column) // Joue le coup sur le plateau

	w.Header().Set("Content-Type", "application/json") // Réponse en JSON
	json.NewEncoder(w).Encode(gameBoard)               // Renvoie le nouvel état de jeu
}

// resetHandler réinitialise la partie et renvoie le nouvel état
func resetHandler(w http.ResponseWriter, r *http.Request) {
	gameBoard.Reset()                                  // Réinitialise le plateau
	w.Header().Set("Content-Type", "application/json") // Réponse au format JSON
	json.NewEncoder(w).Encode(gameBoard)               // Envoie l'état réinitialisé
}

// gameHandler renvoie la page HTML principale du jeu (plateau + jetons)
func gameHandler(w http.ResponseWriter, r *http.Request) {
	// tmpl contient tout le HTML, CSS et JavaScript de la page de jeu
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
      width: 45px;
      height: 45px;
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
          cell.style.left = (55 + col * 62) + 'px';
          cell.style.top = (65 + row * 58) + 'px';
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
	w.Header().Set("Content-Type", "text/html") // Indique qu'on renvoie de l'HTML
	fmt.Fprint(w, tmpl)                         // Écrit le template dans la réponse
}

// webGameHandler renvoie la page d'accueil HTML
func webGameHandler(w http.ResponseWriter, r *http.Request) {
	// tmpl contient le HTML/CSS de la page d'accueil
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
	w.Header().Set("Content-Type", "text/html") // Réponse en HTML
	fmt.Fprint(w, tmpl)                         // Envoie le HTML au navigateur
}

// main est le point d'entrée du programme
func main() {
	// Initialiser l'état du jeu au démarrage
	gameBoard.Reset()

	// Route pour la page d'accueil
	http.HandleFunc("/", webGameHandler)

	// Route pour la page de jeu (plateau + interaction)
	http.HandleFunc("/game", gameHandler)

	// Route pour servir les fichiers du dossier assets/ (images)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("../../assets/"))))

	// Routes de l'API pour l'état, les coups et le reset
	http.HandleFunc("/api/state", stateHandler) // Récupérer l'état du jeu
	http.HandleFunc("/api/drop", dropHandler)   // Jouer un coup
	http.HandleFunc("/api/reset", resetHandler) // Réinitialiser la partie

	fmt.Println("Serveur Puissance 4 démarré sur http://localhost:8080") // Message console
	fmt.Println("Ouvrez votre navigateur à cette adresse pour jouer !")  // Indication à l'utilisateur
	fmt.Println("Images servies depuis : ./assets/picture/")             // Info sur le dossier d'images

	log.Fatal(http.ListenAndServe(":8080", nil)) // Lance le serveur HTTP sur le port 8080 et log en cas d'erreur
}
