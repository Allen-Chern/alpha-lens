package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Event is the JSON payload sent to all connected clients.
type Event struct {
	Type         string `json:"type"`
	EpisodeID    int    `json:"episode_id,omitempty"`
	PodcastID    int    `json:"podcast_id,omitempty"`
	Status       string `json:"status,omitempty"`
	Src          string `json:"src,omitempty"`
	Chars        int    `json:"chars,omitempty"`
	MentionCount int    `json:"mention_count,omitempty"`
	Message      string `json:"message,omitempty"`
}

// Hub maintains the set of connected WebSocket clients and broadcasts events to them.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) Register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.Close()
}

// Broadcast sends e to all connected clients; per-client write errors are silently dropped.
func (h *Hub) Broadcast(e Event) {
	b, _ := json.Marshal(e)
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		c.WriteMessage(websocket.TextMessage, b) //nolint:errcheck
	}
}
