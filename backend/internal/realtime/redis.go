package realtime

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gabrielevieira/palpitai/backend/internal/cache"
	"github.com/gabrielevieira/palpitai/backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

const RedisChannel = "palpitai:realtime:events"

type RedisPublisher struct {
	client *cache.RedisClient
	logger *slog.Logger
}

func NewRedisPublisher(client *cache.RedisClient, logger *slog.Logger) RedisPublisher {
	if logger == nil {
		logger = slog.Default()
	}

	return RedisPublisher{
		client: client,
		logger: logger,
	}
}

func (publisher RedisPublisher) Publish(ctx context.Context, event domain.Event) {
	if publisher.client == nil {
		publisher.logger.Warn("redis realtime publish skipped", "reason", "redis client not configured", "event", event.Name)
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		publisher.logger.Warn("redis realtime marshal failed", "event", event.Name, "error", err)
		return
	}

	if err := publisher.client.Publish(ctx, RedisChannel, payload); err != nil {
		publisher.logger.Warn("redis realtime publish failed", "event", event.Name, "room", event.Room, "error", err)
	}
}

func SubscribeRedis(ctx context.Context, client *cache.RedisClient, hub Publisher, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	if client == nil || hub == nil {
		logger.Warn("redis realtime subscriber disabled", "reason", "redis client or hub not configured")
		return
	}

	subscription := client.Subscribe(ctx, RedisChannel)
	defer subscription.Close()

	if _, err := subscription.Receive(ctx); err != nil {
		logger.Error("redis realtime subscribe failed", "channel", RedisChannel, "error", err)
		return
	}

	logger.Info("redis realtime subscriber started", "channel", RedisChannel)
	messages := subscription.Channel(redis.WithChannelHealthCheckInterval(0))

	for {
		select {
		case <-ctx.Done():
			logger.Info("redis realtime subscriber stopped")
			return
		case message, ok := <-messages:
			if !ok {
				logger.Warn("redis realtime subscriber channel closed")
				return
			}

			var event domain.Event
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				logger.Warn("redis realtime event ignored", "error", err)
				continue
			}

			logger.Info(
				"redis realtime event received",
				"name", event.Name,
				"room", event.Room,
				"match_id", event.Payload["match_id"],
				"group_id", event.Payload["group_id"],
			)
			hub.Publish(ctx, event)
		}
	}
}
