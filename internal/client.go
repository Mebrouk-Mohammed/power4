package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	// Ex: "http://localhost/powerfour/power4/php/php_api"
	BaseURL string
	// Doit matcher API_KEY dans php/api_boot.php
	APIKey string
	HTTP   *http.Client
}

func (c *Client) get(path string, out any) error {
	req, _ := http.NewRequest("GET", c.BaseURL+path, nil)
	req.Header.Set("X-Api-Key", c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var e map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("HTTP %d: %v", resp.StatusCode, e["error"])
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) post(path string, body any, out any) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var e map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("HTTP %d: %v", resp.StatusCode, e["error"])
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

type Game struct {
	ID           int    `json:"id"`
	Status       string `json:"status"`
	Player1ID    int    `json:"player1_id"`
	Player2ID    *int   `json:"player2_id"`
	PlayerToMove *int   `json:"player_to_move"`
	RowsCount    int    `json:"rows_count"`
	ColsCount    int    `json:"cols_count"`
	ConnectN     int    `json:"connect_n"`
}

type Move struct {
	MoveNo      int    `json:"move_no"`
	PlayerID    int    `json:"player_id"`
	ColumnIndex int    `json:"column_index"`
	RowIndex    int    `json:"row_index"`
	DiscColor   string `json:"disc_color"`
	PlayedAt    string `json:"played_at"`
}

type GameGetResp struct {
	Game  Game   `json:"game"`
	Moves []Move `json:"moves"`
}

/* ---------- Création / lecture ---------- */

func (c *Client) CreateGame(p1 int, p2 *int) (int, error) {
	var out struct {
		OK     bool   `json:"ok"`
		GameID int    `json:"game_id"`
		Error  string `json:"error"`
	}
	body := map[string]any{"player1_id": p1}
	if p2 != nil {
		body["player2_id"] = *p2
	}
	if err := c.post("/game_create.php", body, &out); err != nil {
		return 0, err
	}
	if !out.OK && out.Error != "" {
		return 0, fmt.Errorf(out.Error)
	}
	return out.GameID, nil
}

func (c *Client) GetGame(gameID int) (*GameGetResp, error) {
	var out GameGetResp
	err := c.get(fmt.Sprintf("/game_get.php?id=%d", gameID), &out)
	return &out, err
}

/* ---------- Commit de coup (gameplay côté Go) ---------- */

// FinishInfo permet de clore la partie avec ce coup.
type FinishInfo struct {
	Status   string `json:"status"`              // "finished"
	WinnerID *int   `json:"winner_id,omitempty"` // null pour un nul
}

// CommitMoveReq est envoyé à move_commit.php
type CommitMoveReq struct {
	GameID       int         `json:"game_id"`
	PlayerID     int         `json:"player_id"`
	ColumnIndex  int         `json:"column_index"`
	RowIndex     int         `json:"row_index"`                // calculé par TON moteur Go
	NextPlayerID *int        `json:"next_player_id,omitempty"` // requis si la partie CONTINUE
	Finish       *FinishInfo `json:"finish,omitempty"`         // présent si la partie se TERMINE
}

// EloDelta (renvoyé quand la partie se termine)
type EloDelta struct {
	P1 struct {
		ID  int `json:"id"`
		Old int `json:"old"`
		New int `json:"new"`
	} `json:"p1"`
	P2 struct {
		ID  int `json:"id"`
		Old int `json:"old"`
		New int `json:"new"`
	} `json:"p2"`
}

// CommitMoveResp réponse de move_commit.php
type CommitMoveResp struct {
	OK       bool      `json:"ok"`
	MoveNo   int       `json:"move_no"`
	RowIndex int       `json:"row_index"`
	Disc     string    `json:"disc"`
	Finished bool      `json:"finished"`
	Error    string    `json:"error"`
	Elo      *EloDelta `json:"elo,omitempty"`
}

// CommitMove : TU valides le coup côté Go, PHP ne fait que commit + vérifier la cohérence.
func (c *Client) CommitMove(req CommitMoveReq) (*CommitMoveResp, error) {
	var out CommitMoveResp
	if err := c.post("/move_commit.php", req, &out); err != nil {
		return nil, err
	}
	if !out.OK && out.Error != "" {
		return nil, fmt.Errorf(out.Error)
	}
	return &out, nil
}

/* ---------- Compat (si tu veux les garder) ---------- */

// PlayMove (legacy) : appelle move_play.php — usage déconseillé si le gameplay est 100% Go.
func (c *Client) PlayMove(gameID, playerID, column int) (moveNo, rowIndex int, disc string, err error) {
	var out struct {
		OK       bool   `json:"ok"`
		MoveNo   int    `json:"move_no"`
		RowIndex int    `json:"row_index"`
		Disc     string `json:"disc"`
		Error    string `json:"error"`
	}
	err = c.post("/move_play.php", map[string]int{
		"game_id": gameID, "player_id": playerID, "column_index": column,
	}, &out)
	if err != nil {
		return 0, 0, "", err
	}
	if !out.OK && out.Error != "" {
		return 0, 0, "", fmt.Errorf(out.Error)
	}
	return out.MoveNo, out.RowIndex, out.Disc, nil
}

// FinishGame (legacy) : fin séparée — préfère CommitMove avec FinishInfo.
func (c *Client) FinishGame(gameID int, status string, winnerID *int) error {
	body := map[string]any{"game_id": gameID, "status": status}
	if winnerID != nil {
		body["winner_id"] = *winnerID
	}
	var out map[string]any
	return c.post("/game_finish.php", body, &out)
}
