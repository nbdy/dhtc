package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GotifyNotifier struct {
	URL   string
	Token string
}

func (n *GotifyNotifier) Notify(message string) error {
	if n.URL == "" || n.Token == "" {
		return nil
	}
	url := fmt.Sprintf("%s/message?token=%s", n.URL, n.Token)
	payload := map[string]interface{}{
		"message":  message,
		"priority": 5,
		"title":    "dhtc",
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("gotify returned status %d", resp.StatusCode)
	}
	return nil
}
