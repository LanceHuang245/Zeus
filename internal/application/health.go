package application

import (
	"context"
	"time"

	"Zephyr/internal/ports"
)

const (
	RateLimitPerMinute int64 = 10
	RateLimitWindow          = time.Minute
	AnomalyThreshold   int64 = 50
	AnomalyWindow            = 5 * time.Minute
)

// HealthResult contains the result of health access checks
type HealthResult struct {
	RateLimitExceeded bool
	AnomalyDetected   bool
	AnomalyCount      int64
	AnomalyTime       time.Time
}

// HealthService applies rate limiting and anomaly detection policies
type HealthService struct {
	store ports.AccessStore
}

// NewHealthService creates a health application service
func NewHealthService(store ports.AccessStore) *HealthService {
	return &HealthService{store: store}
}

// Check applies the health access policies in their legacy order
func (s *HealthService) Check(ctx context.Context, clientIP string) HealthResult {
	result := HealthResult{}
	now := time.Now()

	rateCount, err := s.store.RecordAndCount(
		ctx,
		"rate_limit:health_check:"+clientIP,
		now,
		RateLimitWindow,
		now.Unix(),
	)
	if err == nil {
		result.RateLimitExceeded = rateCount > RateLimitPerMinute
	}
	if result.RateLimitExceeded {
		return result
	}

	anomalyNow := time.Now()
	anomalyCount, err := s.store.RecordAndCount(
		ctx,
		"anomaly:health_check:"+clientIP,
		anomalyNow,
		AnomalyWindow,
		anomalyNow.UnixNano(),
	)
	if err == nil {
		result.AnomalyCount = anomalyCount
		result.AnomalyTime = anomalyNow
		result.AnomalyDetected = anomalyCount > AnomalyThreshold
	}
	return result
}
