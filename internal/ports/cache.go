package ports

import (
	"context"
	"time"
)

// Cache is the application-facing cache contract
// Infrastructure details such as Redis remain outside the application layer
// Cache defines the application cache operations
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}
