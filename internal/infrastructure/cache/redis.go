// Package cache provides caching functionality using Redis
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache provides Redis-based caching
type RedisCache struct {
	client  *redis.Client
	enabled bool
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(url string, password string, db int, enabled bool) (*RedisCache, error) {
	if !enabled {
		return &RedisCache{enabled: false}, nil
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis URL: %w", err)
	}

	if password != "" {
		opt.Password = password
	}
	opt.DB = db

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client:  client,
		enabled: true,
	}, nil
}

// Get retrieves a value from cache and unmarshals it into the target
func (c *RedisCache) Get(ctx context.Context, key string, target interface{}) error {
	if !c.enabled {
		return fmt.Errorf("cache is disabled")
	}

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get from cache: %w", err)
	}

	if err := json.Unmarshal([]byte(val), target); err != nil {
		return fmt.Errorf("failed to unmarshal cached value: %w", err)
	}

	return nil
}

// Set stores a value in cache with the given TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.enabled {
		return nil // Silently skip if disabled
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if !c.enabled {
		return nil
	}

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	if c.enabled && c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Ping checks if Redis is responsive
func (c *RedisCache) Ping(ctx context.Context) error {
	if !c.enabled {
		return fmt.Errorf("cache is disabled")
	}
	return c.client.Ping(ctx).Err()
}

// IsEnabled returns whether caching is enabled
func (c *RedisCache) IsEnabled() bool {
	return c.enabled
}
