package game // Package du jeu

import "sync" // Permet d'utiliser des verrous pour la concurrence

const (
	Empty = 0 // case vide
	P1    = 1 // joueur 1
	P2    = 2 // joueur 2
)

type Position struct{ R, C int } // Structure représentant une position (ligne, colonne)

type Game struct {
	Rows, Cols      int        // Dimensions du plateau
	Board           [][]int    // Contenu du plateau
	CurrentPlayer   int        // Joueur courant (P1 ou P2)
	Winner          int        // Vainqueur (0 = aucun, -1 = égalité)
	MoveCount       int        // Nombre total de coups joués
	InvertedGravity bool       // Mode : gravité inversée ou normale
	Mu              sync.Mutex // Verrou pour empêcher les accès concurrents
}

func New(rows, cols int) *Game {
	g := &Game{
		Rows:          rows,                // Nombre de lignes
		Cols:          cols,                // Nombre de colonnes
		CurrentPlayer: P1,                  // Le joueur 1 commence
		Winner:        0,                   // Aucun gagnant au début
		MoveCount:     0,                   // Aucun coup joué
		Board:         make([][]int, rows), // Création structure du plateau
	}
	for r := range g.Board {
		g.Board[r] = make([]int, cols) // Chaque ligne contient "cols" cases
	}
	return g // Retourne la nouvelle instance du jeu
}

func (g *Game) Reset(rows, cols int) {
	g.Mu.Lock()         // Verrouille pour éviter les accès concurrentiels
	defer g.Mu.Unlock() // Déverrouille à la fin

	g.Rows = rows                 // Redéfinit le nombre de lignes
	g.Cols = cols                 // Redéfinit le nombre de colonnes
	g.Board = make([][]int, rows) // Recrée le plateau vide
	for r := range g.Board {
		g.Board[r] = make([]int, cols)
	}
	g.CurrentPlayer = P1 // Le joueur 1 recommence
	g.Winner = 0         // Pas de gagnant au reset
	g.MoveCount = 0      // Nombre de coups remis à zéro
}

func (g *Game) Drop(col int) bool {
	g.Mu.Lock()         // Verrou pour la sécurité en multi-accès
	defer g.Mu.Unlock() // Déverrouille ensuite

	if col < 0 || col >= g.Cols || g.Winner != 0 { // Vérifie si colonne valide ou si la partie est déjà finie
		return false
	}

	if g.InvertedGravity { // Mode gravité inversée : les jetons montent
		for r := 0; r < g.Rows; r++ { // Parcours du haut vers le bas
			if g.Board[r][col] == Empty { // Trouve la première case vide
				g.Board[r][col] = g.CurrentPlayer // Place le jeton
				g.MoveCount++                     // Incrémente le nombre de coups
				g.checkEnd(r, col)                // Vérifie si fin de partie
				if g.Winner == 0 {
					g.CurrentPlayer = 3 - g.CurrentPlayer // Change de joueur
				}
				return true
			}
		}
	} else { // Mode gravité normale : les jetons tombent vers le bas
		for r := g.Rows - 1; r >= 0; r-- { // Parcours du bas vers le haut
			if g.Board[r][col] == Empty { // Première case vide en bas
				g.Board[r][col] = g.CurrentPlayer // Place le jeton
				g.MoveCount++                     // Incrémente le compteur
				g.checkEnd(r, col)                // Vérifie fin de la partie
				if g.Winner == 0 {
					g.CurrentPlayer = 3 - g.CurrentPlayer // Alterne entre 1 et 2
				}
				return true
			}
		}
	}
	return false // Si aucune case vide -> coup impossible
}

func (g *Game) checkEnd(r, c int) {
	p := g.Board[r][c] // Récupère le joueur ayant joué
	if p != P1 && p != P2 {
		return // Si la case n'appartient à personne, on ne vérifie rien
	}

	// Vérifie les 4 directions : horizontal, vertical, diagonales
	if g.four(r, c, 1, 0, p) || // ligne horizontale
		g.four(r, c, 0, 1, p) || // colonne verticale
		g.four(r, c, 1, 1, p) || // diagonale ↘
		g.four(r, c, 1, -1, p) { // diagonale ↗
		g.Winner = p // Déclare le gagnant
		return
	}

	// Si toutes les cases sont jouées -> égalité
	if g.MoveCount == g.Rows*g.Cols {
		g.Winner = -1
	}
}

func (g *Game) four(r, c, dr, dc, p int) bool {
	// Vérifie si la somme des jetons alignés dans les deux directions >= 4
	return 1+g.countDir(r, c, dr, dc, p)+g.countDir(r, c, -dr, -dc, p) >= 4
}

func (g *Game) countDir(r, c, dr, dc, p int) int {
	n := 0 // Nombre de jetons alignés
	for i, j := r+dr, c+dc; i >= 0 && i < g.Rows && j >= 0 && j < g.Cols; i, j = i+dr, j+dc {
		if g.Board[i][j] != p { // Si la case n'appartient pas au même joueur
			break // Stoppe le comptage
		}
		n++ // Compte un jeton de plus
	}
	return n // Retourne le nombre trouvé
}
