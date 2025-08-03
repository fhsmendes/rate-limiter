package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/go-redis/redis/v8"
)

type RedisLimiter struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		ctx:    context.Background(),
	}
}

// IsAllowed checks if the request is allowed based on the rate limit and block duration.
func (rl *RedisLimiter) IsAllowed(key string, limit int, expTime time.Duration, blockTime time.Duration) bool {
	blockedKey := "blocked:" + key

	if rl.client.Exists(rl.ctx, blockedKey).Val() > 0 {
		return false //blocked
	}

	count, err := rl.client.Incr(rl.ctx, key).Result()
	if err != nil {
		log.Printf("error on incrementing Redis key, error: %s", err.Error())
		return false
	}

	if int(count) > limit {
		rl.client.Set(rl.ctx, blockedKey, "1", blockTime)
		rl.client.Expire(rl.ctx, key, expTime)
		return false
	}

	if rl.client.TTL(rl.ctx, key).Val() == -1 {
		rl.client.Expire(rl.ctx, key, expTime)
	}

	return true
}

// NewRedisClient creates a new Redis client and connects to the Redis server.
func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     configs.Envs.Redis_Host,
		Password: configs.Envs.Redis_Password,
		DB:       configs.Envs.Redis_DB,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("error on connection to Redis: %w", err)
	}

	return client, nil
}
