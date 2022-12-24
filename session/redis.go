package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v9"
)

// RedisClient Provides functions to work with redis
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

// Set Sets a new value in redis
func (client *RedisClient) Set(ctx context.Context, key string, session *Session) error {
	body, err := json.Marshal(session)
	if err != nil {
		return err
	}

	t1 := time.Now()
	t2 := session.ExpiresAt
	til := t2.Sub(t1)

	return client.rdb.Set(ctx, key, body, til).Err()
}

// Get Gets value from redis by key
func (client *RedisClient) Get(ctx context.Context, key string) (*Session, error) {
	cmd := client.rdb.Get(ctx, key)
	_, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("couldn't get value from redis: %w", err)
	}

	res, err := cmd.Bytes()
	if err != nil {
		return nil, fmt.Errorf("couldn't get res as bytes: %w", err)
	}

	var session Session
	err = json.Unmarshal(res, &session)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal: %w", err)
	}

	return &session, nil
}

// Del Deletes a value from redis
func (client *RedisClient) Del(ctx context.Context, key string) error {
	return client.rdb.Del(ctx, key).Err()
}
