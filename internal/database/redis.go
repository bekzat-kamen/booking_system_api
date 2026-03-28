package database

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedisConnection(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return client, nil
}

func CloseRedis(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}
