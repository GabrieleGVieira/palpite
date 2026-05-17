package httpapi

import (
	"context"
	"net/http"

	"github.com/gabrielevieira/palpitai/backend/internal/config"
	"github.com/gabrielevieira/palpitai/backend/internal/matchsync"
)

type websocketHub interface {
	ServeWS(w http.ResponseWriter, r *http.Request, userID string, rooms []string)
}

type realtimePublisher interface {
	Publish(ctx context.Context, event matchsync.Event)
}

type realtimeService interface {
	realtimePublisher
	websocketHub
}

func realtimeHandler(cfg config.Config, db datastore, hub websocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hub == nil {
			writeError(w, http.StatusServiceUnavailable, "Realtime indisponivel.")
			return
		}

		userID, err := userIDFromToken(r, cfg, r.URL.Query().Get("token"))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Informe um token de autenticacao valido.")
			return
		}

		rooms := []string{}
		groupID := r.URL.Query().Get("group_id")
		if groupID != "" {
			if err := ensureActiveGroupMember(r.Context(), db, userID, groupID); err != nil {
				writeError(w, http.StatusForbidden, "Você precisa participar deste grupo.")
				return
			}

			rooms = append(rooms, "group:"+groupID)
		}

		hub.ServeWS(w, r, userID, rooms)
	}
}
