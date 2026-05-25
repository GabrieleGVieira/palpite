package realtime

import (
	"context"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, event domain.Event)
}

func (hub *Hub) Publish(_ context.Context, event domain.Event) {
	outbound := outboundEvent{
		Name:    event.Name,
		Payload: event.Payload,
		Room:    event.Room,
	}

	staleClients := []*client{}
	deliveredClients := 0

	hub.mu.RLock()
	for client := range hub.clients {
		if !client.subscribesTo(event.Room) {
			continue
		}

		select {
		case client.send <- outbound:
			deliveredClients++
		default:
			staleClients = append(staleClients, client)
		}
	}
	hub.mu.RUnlock()

	hub.logger.Info(
		"realtime event delivered to websocket clients",
		"name", event.Name,
		"room", event.Room,
		"match_id", event.Payload["match_id"],
		"group_id", event.Payload["group_id"],
		"clients", deliveredClients,
		"stale_clients", len(staleClients),
	)

	for _, client := range staleClients {
		hub.closeClient(client)
	}
}
