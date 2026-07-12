package redis

import (
	"context"
	"time"

	redisclient "github.com/go-redis/redis/v8"
)

// Cache adapts Redis to the application cache contract
type Cache struct {
	client *redisclient.Client
}

// NewCache creates a Redis cache adapter
func NewCache(client *redisclient.Client) *Cache {
	return &Cache{client: client}
}

// Get retrieves a cached value by key
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set stores a cached value with a time to live
func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}
