package ui

import (
	"dhtc/db"
	dhtcclient "dhtc/dhtc-client"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Error().Err(err).Msg("could not write message to websocket")
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

type TrawlMessage struct {
	Name         string
	InfoHashHex  string
	TotalSize    uint64
	DiscoveredOn int64
	Files        []dhtcclient.File
	Categories   []string
}

func (h *Hub) BroadcastMetadata(md dhtcclient.Metadata) {
	msg := TrawlMessage{
		Name:         md.Name,
		InfoHashHex:  fmt.Sprintf("%x", md.InfoHash),
		TotalSize:    md.TotalSize,
		DiscoveredOn: md.DiscoveredOn,
		Files:        md.Files,
		Categories:   db.Categorize(md),
	}
	h.Broadcast(msg)
}

func (h *Hub) Broadcast(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("could not marshal broadcast message")
		return
	}
	h.broadcast <- data
}

func (c *Controller) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("could not upgrade connection")
		return
	}
	c.Hub.register <- conn
	defer func() {
		c.Hub.unregister <- conn
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
