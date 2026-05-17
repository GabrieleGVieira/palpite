package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gabrielevieira/palpitai/backend/internal/matchsync"
	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	writeWait  = 10 * time.Second
	pingPeriod = 45 * time.Second
)

type Hub struct {
	clients  map[*client]struct{}
	logger   *slog.Logger
	mu       sync.RWMutex
	upgrader websocket.Upgrader
}

type client struct {
	conn   *websocket.Conn
	rooms  map[string]struct{}
	send   chan outboundEvent
	userID string
}

type outboundEvent struct {
	Name    string         `json:"name"`
	Payload map[string]any `json:"payload"`
	Room    string         `json:"room,omitempty"`
}

func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}

	return &Hub{
		clients: make(map[*client]struct{}),
		logger:  logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
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

	nextClient := &client{
		conn:   conn,
		rooms:  make(map[string]struct{}),
		send:   make(chan outboundEvent, 16),
		userID: userID,
	}

	nextClient.rooms["matches"] = struct{}{}
	nextClient.rooms["rankings"] = struct{}{}
	nextClient.rooms["user:"+userID] = struct{}{}
	for _, room := range rooms {
		if room != "" {
			nextClient.rooms[room] = struct{}{}
		}
	}

	hub.mu.Lock()
	hub.clients[nextClient] = struct{}{}
	hub.mu.Unlock()

	go hub.writePump(nextClient)
	hub.readPump(nextClient)
}

func (hub *Hub) Publish(_ context.Context, event matchsync.Event) {
	outbound := outboundEvent{
		Name:    event.Name,
		Payload: event.Payload,
		Room:    event.Room,
	}

	staleClients := []*client{}

	hub.mu.RLock()
	for client := range hub.clients {
		if event.Room != "" {
			if _, ok := client.rooms[event.Room]; !ok {
				continue
			}
		}

		select {
		case client.send <- outbound:
		default:
			staleClients = append(staleClients, client)
		}
	}
	hub.mu.RUnlock()

	for _, client := range staleClients {
		hub.closeClient(client)
	}
}

func (hub *Hub) readPump(client *client) {
	defer hub.closeClient(client)

	client.conn.SetReadLimit(512)
	_ = client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := client.conn.NextReader(); err != nil {
			return
		}
	}
}

func (hub *Hub) writePump(client *client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		hub.closeClient(client)
	}()

	for {
		select {
		case event, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			payload, err := json.Marshal(event)
			if err != nil {
				hub.logger.Warn("websocket event marshal failed", "error", err)
				continue
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (hub *Hub) closeClient(client *client) {
	hub.mu.Lock()
	if _, ok := hub.clients[client]; ok {
		delete(hub.clients, client)
		close(client.send)
	}
	hub.mu.Unlock()

	_ = client.conn.Close()
}
