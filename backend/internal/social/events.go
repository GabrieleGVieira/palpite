package social

import (
	"context"
	"log/slog"

	"github.com/gabrielevieira/palpitai/backend/internal/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domain.SocialEvent) error
}

type LogEventPublisher struct{}

func (LogEventPublisher) Publish(_ context.Context, event domain.SocialEvent) error {
	slog.Info("social event emitted", "type", event.Type, "actor_user_id", event.ActorUserID, "target_type", event.TargetType, "target_id", event.TargetID)
	return nil
}
