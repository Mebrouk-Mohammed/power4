package gameclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func New() *Client {
	base := getenv("API_URL", "http://localhost/powerfour/power4/php/php_api")
	key := getenv("API_KEY", "cle-api-fuefijefe524895")
	return &Client{BaseURL: base, APIKey: key, HTTP: http.DefaultClient}
}

func (c *Client) CreateGame(p1, p2 int) (int, error) {
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

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
