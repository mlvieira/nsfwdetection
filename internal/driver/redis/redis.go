package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisService provides Redis operations
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient initializes a new Redis service
func NewRedisClient(addr, password string, db int) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{client: client}
}

// SetValue sets a key-value pair in Redis with an expiration time
func (rs *RedisClient) SetValue(ctx context.Context, key string, value string, expiration time.Duration) error {
	return rs.client.Set(ctx, key, value, expiration).Err()
}

// GetValue retrieves a value by key from Redis
func (rs *RedisClient) GetValue(ctx context.Context, key string) (string, error) {
	return rs.client.Get(ctx, key).Result()
}

// DeleteKey deletes a key from Redis
func (rs *RedisClient) DeleteKey(ctx context.Context, key string) error {
	return rs.client.Del(ctx, key).Err()
}
