package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type APIClient struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewAPI() *APIClient {
	url := os.Getenv("API_URL")
	if url == "" {
		url = "http://localhost/powerfour/power4/php/php_api"
	}
	key := os.Getenv("API_KEY")
	if key == "" {
		key = "cle-api-fuefijefe524895"
	}
	return &APIClient{BaseURL: url, APIKey: key, HTTP: http.DefaultClient}
}

func (c *APIClient) CreateGame(p1, p2 int) (int, error) {
	body := map[string]any{"player1_id": p1}
	if p2 > 0 {
		body["player2_id"] = p2
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", c.BaseURL+"/game_create.php", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var out struct {
		OK     bool   `json:"ok"`
		GameID int    `json:"game_id"`
		Error  string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	if !out.OK {
		return 0, fmt.Errorf(out.Error)
	}
	return out.GameID, nil
}

// 🧩 ➜ COLLE TA NOUVELLE FONCTION ICI :
func (c *APIClient) GetCurrentUser() (int, string, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/whoami.php", nil)
	req.Header.Set("X-Api-Key", c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	var out struct {
		OK       bool   `json:"ok"`
		UserID   int    `json:"user_id"`
		Username string `json:"username"`
		Error    string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, "", err
	}
	if !out.OK {
		return 0, "", fmt.Errorf(out.Error)
	}
	return out.UserID, out.Username, nil
}
