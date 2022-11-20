package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
)

// Provides functions to work with redis
type RedisClient struct {
	rdb *redis.Client
}

// NewRedisClient creates and returns a new RedisClient
func NewRedisClient(address, password string) (*RedisClient, error) {
	client := &RedisClient{
		rdb: redis.NewClient(&redis.Options{
			Addr:        address,
			Password:    password,
			DB:          0, // default db
			DialTimeout: 100 * time.Millisecond,
			ReadTimeout: 100 * time.Millisecond,
		}),
	}

	if _, err := client.rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// Sets a new value in redis
func (client *RedisClient) Set(ctx context.Context, key string, body []byte, til time.Duration) error {
	return client.rdb.Set(ctx, key, body, til).Err()
}

// Gets value from redis by key
func (client *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	cmd := client.rdb.Get(ctx, key)
	_, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	return cmd.Bytes()
}

// Deletes a value from redis
func (client *RedisClient) Del(ctx context.Context, key string) error {
	return client.rdb.Del(ctx, key).Err()
}
