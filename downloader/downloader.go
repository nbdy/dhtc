package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TransmissionClient struct {
	URL       string
	User      string
	Pass      string
	SessionID string
}

func (c *TransmissionClient) AddMagnet(magnet string) error {
	payload := map[string]interface{}{
		"method": "torrent-add",
		"arguments": map[string]interface{}{
			"filename": magnet,
		},
	}

	for i := 0; i < 2; i++ {
		body, _ := json.Marshal(payload)
		req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(body))
		if err != nil {
			return err
		}

		if c.User != "" {
			req.SetBasicAuth(c.User, c.Pass)
		}
		if c.SessionID != "" {
			req.Header.Set("X-Transmission-Session-Id", c.SessionID)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusConflict {
			c.SessionID = resp.Header.Get("X-Transmission-Session-Id")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("transmission returned status %d", resp.StatusCode)
		}

		return nil
	}

	return fmt.Errorf("failed to get transmission session id")
}

type Aria2Client struct {
	URL   string
	Token string
}

func (c *Aria2Client) AddMagnet(magnet string) error {
	params := []interface{}{}
	if c.Token != "" {
		params = append(params, "token:"+c.Token)
	}
	params = append(params, []string{magnet})

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "dhtc",
		"method":  "aria2.addUri",
		"params":  params,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(c.URL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("aria2 returned status %d: %s", resp.StatusCode, string(b))
	}

	return nil
}

type DelugeClient struct {
	URL  string
	Pass string
}

func (c *DelugeClient) AddMagnet(magnet string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 1. Login
	loginPayload := map[string]interface{}{
		"id":     1,
		"method": "auth.login",
		"params": []string{c.Pass},
	}
	loginBody, _ := json.Marshal(loginPayload)
	resp, err := client.Post(c.URL, "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deluge login failed with status %d", resp.StatusCode)
	}

	var cookies []*http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "_session_id" {
			cookies = append(cookies, cookie)
		}
	}

	// 2. Add magnet
	addPayload := map[string]interface{}{
		"id":     2,
		"method": "core.add_torrent_magnet",
		"params": []interface{}{magnet, map[string]interface{}{}},
	}
	addBody, _ := json.Marshal(addPayload)
	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(addBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deluge add magnet failed with status %d", resp.StatusCode)
	}

	return nil
}

type QBittorrentClient struct {
	URL  string
	User string
	Pass string
}

func (c *QBittorrentClient) AddMagnet(magnet string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 1. Login
	loginURL := fmt.Sprintf("%s/api/v2/auth/login", c.URL)
	loginData := fmt.Sprintf("username=%s&password=%s", c.User, c.Pass)
	resp, err := client.Post(loginURL, "application/x-www-form-urlencoded", bytes.NewBufferString(loginData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qbittorrent login failed with status %d", resp.StatusCode)
	}

	var cookies []*http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			cookies = append(cookies, cookie)
		}
	}

	// 2. Add magnet
	addURL := fmt.Sprintf("%s/api/v2/torrents/add", c.URL)
	addData := fmt.Sprintf("urls=%s", magnet)
	req, err := http.NewRequest("POST", addURL, bytes.NewBufferString(addData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qbittorrent add magnet failed with status %d", resp.StatusCode)
	}

	return nil
}
