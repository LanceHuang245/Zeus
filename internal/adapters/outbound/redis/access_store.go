package redis

import (
	"context"
	"strconv"
	"time"

	redisclient "github.com/go-redis/redis/v8"
)

// AccessStore adapts Redis sorted sets to the health access store contract
type AccessStore struct {
	client *redisclient.Client
}

// NewAccessStore creates a Redis access store adapter
func NewAccessStore(client *redisclient.Client) *AccessStore {
	return &AccessStore{client: client}
}

// RecordAndCount records an access event and returns the active window count
func (s *AccessStore) RecordAndCount(ctx context.Context, key string, now time.Time, window time.Duration, member interface{}) (int64, error) {
	pipe := s.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now.Add(-window).Unix(), 10))
	pipe.ZAdd(ctx, key, &redisclient.Z{Score: float64(now.Unix()), Member: member})
	pipe.Expire(ctx, key, window)
	countCmd := pipe.ZCard(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return countCmd.Val(), nil
}
