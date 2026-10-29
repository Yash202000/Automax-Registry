package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/automax/registry/internal/config"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	log.Println("registry redis connected")
	return client, nil
}

func CloseRedis(client *redis.Client) {
	if client == nil {
		return
	}
	_ = client.Close()
}
