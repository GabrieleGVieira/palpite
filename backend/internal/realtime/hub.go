package realtime

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}

	return &Hub{
		clients: make(map[*client]struct{}),
		logger:  logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool {
				return true
			},
		},
	}
}

func (hub *Hub) ServeWS(w http.ResponseWriter, r *http.Request, userID string, rooms []string) {
	conn, err := hub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.logger.Warn("websocket upgrade failed", "error", err)
		return
	}

	nextClient := newClient(userID, rooms, conn)
	hub.registerClient(nextClient)
	hub.logger.Info("websocket client connected", "user_id", userID, "rooms", rooms)

	go hub.writePump(nextClient)
	hub.readPump(nextClient)
}

func (hub *Hub) registerClient(client *client) {
	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()
}

func (hub *Hub) closeClient(client *client) {
	hub.mu.Lock()
	if _, ok := hub.clients[client]; ok {
		delete(hub.clients, client)
		close(client.send)
	}
	hub.mu.Unlock()

	if client.conn != nil {
		_ = client.conn.Close()
	}
	hub.logger.Info("websocket client disconnected", "user_id", client.userID)
}
