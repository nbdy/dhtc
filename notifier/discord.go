package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type DiscordNotifier struct {
	WebhookURL string
}

func (n *DiscordNotifier) Notify(message string) error {
	if n.WebhookURL == "" {
		return nil
	}
	payload := map[string]string{
		"content": message,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord returned status %d", resp.StatusCode)
	}
	return nil
}
