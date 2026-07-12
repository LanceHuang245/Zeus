package application

import (
	"context"
	"testing"
	"time"
)

// fakeAccessStore provides deterministic access counts for health tests
type fakeAccessStore struct {
	counts []int64
	keys   []string
}

// RecordAndCount returns the next configured access count
func (s *fakeAccessStore) RecordAndCount(_ context.Context, key string, _ time.Time, _ time.Duration, _ interface{}) (int64, error) {
	s.keys = append(s.keys, key)
	count := s.counts[0]
	s.counts = s.counts[1:]
	return count, nil
}

// TestHealthServicePreservesCheckOrder verifies the legacy policy order
func TestHealthServicePreservesCheckOrder(t *testing.T) {
	store := &fakeAccessStore{counts: []int64{11, 99}}
	result := NewHealthService(store).Check(context.Background(), "127.0.0.1")
	if !result.RateLimitExceeded || result.AnomalyDetected {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(store.keys) != 1 {
		t.Fatalf("anomaly check should not run after rate limit: %v", store.keys)
	}
}
