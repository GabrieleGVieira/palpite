package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, redisURL string) (*RedisClient, error) {
	if strings.TrimSpace(redisURL) == "" {
		return nil, errors.New("REDIS_URL is required")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	opts.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolSize = 5
	opts.MinIdleConns = 1

	client := redis.NewClient(opts)
	cache := &RedisClient{client: client}

	if err := cache.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}

	return cache, nil
}

func (cache *RedisClient) Ping(ctx context.Context) error {
	return cache.client.Ping(ctx).Err()
}

func (cache *RedisClient) Close() error {
	return cache.client.Close()
}
