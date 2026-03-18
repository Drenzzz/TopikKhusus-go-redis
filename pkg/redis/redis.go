package redis

import (
	"context"
	"fmt"
	"time"

	rds "github.com/redis/go-redis/v9"

	"topikkhusus-methodtracker/internal/config"
)

func NewClient(cfg config.Config) (*rds.Client, error) {
	client := rds.NewClient(&rds.Options{
		Addr:         cfg.RedisAddress(),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		DialTimeout:  cfg.RedisTimeout,
		ReadTimeout:  cfg.RedisTimeout,
		WriteTimeout: cfg.RedisTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RedisTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}

func HealthCheck(ctx context.Context, client *rds.Client, timeout time.Duration) error {
	healthCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := client.Ping(healthCtx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}
