package api

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Allow all origins in dev; tighten with an allowlist in production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.Hub.Register(conn)
	defer h.Hub.Unregister(conn)
	// Read loop: drains client frames (pings/close) so the connection stays healthy.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
