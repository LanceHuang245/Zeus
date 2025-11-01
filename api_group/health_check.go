package apigroup

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"Zephyr/config"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Security configuration constants
const (
	// Rate limit: maximum 10 requests per minute
	RateLimitPerMinute = 10
	// Rate limit window time (seconds)
	RateLimitWindowSeconds = 60
	// Anomaly access threshold: more than 50 requests within 5 minutes
	AnomalyThreshold = 50
	// Anomaly detection window time (minutes)
	AnomalyWindowMinutes = 5
	// Allowed Application Header value
	AllowedApplicationHeader = "Zephyr"
)

// AccessLog access log structure
type AccessLog struct {
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
}

// getClientIP get the real client IP address
func getClientIP(c *gin.Context) string {
	// Preferentially get X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Get X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr as the last fallback
	return c.ClientIP()
}

// checkRateLimit check rate limiting
func checkRateLimit(clientIP string) bool {
	// Use Redis to implement sliding window rate limiting
	key := fmt.Sprintf("rate_limit:health_check:%s", clientIP)

	// Get current timestamp
	currentTime := time.Now().Unix()

	// Use pipeline for performance improvement
	pipe := config.RedisClient.Pipeline()

	// Remove expired request records
	pipe.ZRemRangeByScore(config.Ctx, key, "0", strconv.FormatInt(currentTime-RateLimitWindowSeconds, 10))

	// Add current request
	pipe.ZAdd(config.Ctx, key, &redis.Z{
		Score:  float64(currentTime),
		Member: currentTime,
	})

	// Set expiration time
	pipe.Expire(config.Ctx, key, time.Duration(RateLimitWindowSeconds)*time.Second)

	// Get the number of requests in the current window
	countCmd := pipe.ZCard(config.Ctx, key)

	// Execute pipeline
	_, err := pipe.Exec(config.Ctx)
	if err != nil {
		// When Redis fails, allow requests to pass for safety
		return true
	}

	currentCount := countCmd.Val()
	return currentCount <= RateLimitPerMinute
}

// detectAnomaly detect abnormal access
func detectAnomaly(c *gin.Context, clientIP string) bool {
	// Check the number of requests within 5 minutes
	key := fmt.Sprintf("anomaly:health_check:%s", clientIP)

	currentTime := time.Now()
	windowStart := currentTime.Add(-time.Duration(AnomalyWindowMinutes) * time.Minute)

	// Use ZSet to record request time
	pipe := config.RedisClient.Pipeline()

	// Clean up expired data
	pipe.ZRemRangeByScore(config.Ctx, key, "0", strconv.FormatInt(windowStart.Unix(), 10))

	// Add current request
	pipe.ZAdd(config.Ctx, key, &redis.Z{
		Score:  float64(currentTime.Unix()),
		Member: currentTime.UnixNano(),
	})

	// Set expiration time
	pipe.Expire(config.Ctx, key, time.Duration(AnomalyWindowMinutes)*time.Minute)

	// Get the number of requests in the current window
	countCmd := pipe.ZCard(config.Ctx, key)

	// Execute pipeline
	_, err := pipe.Exec(config.Ctx)
	if err != nil {
		return false
	}

	currentCount := countCmd.Val()

	// If threshold exceeded, log as anomaly
	if currentCount > AnomalyThreshold {
		// Log anomaly
		logger, _ := zap.NewProduction()
		defer logger.Sync()

		logger.Error("Anomalous access detected",
			zap.String("ip", clientIP),
			zap.Int("request_count", int(currentCount)),
			zap.Duration("window", time.Duration(AnomalyWindowMinutes)*time.Minute),
			zap.String("user_agent", c.GetHeader("User-Agent")),
			zap.Time("timestamp", currentTime),
		)

		return true
	}

	return false
}

// validateHeaders validate request headers
func validateHeaders(c *gin.Context) bool {
	// Validate Application Header
	appHeader := c.GetHeader("Application")
	if appHeader != AllowedApplicationHeader {
		return false
	}

	if userAgent := c.GetHeader("User-Agent"); userAgent == "" {
		return false
	}

	return true
}

// logAccess log access records
func logAccess(c *gin.Context, clientIP string) {
	accessLog := AccessLog{
		IP:        clientIP,
		UserAgent: c.GetHeader("User-Agent"),
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}

	// Use structured logging
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Health check access",
		zap.String("ip", accessLog.IP),
		zap.String("user_agent", accessLog.UserAgent),
		zap.Time("timestamp", accessLog.Timestamp),
		zap.String("method", accessLog.Method),
		zap.String("path", accessLog.Path),
	)
}

// HealthCheck security-enhanced health check endpoint
func HealthCheck(c *gin.Context) {
	clientIP := getClientIP(c)

	// 1. Validate request headers
	if !validateHeaders(c) {
		c.JSON(400, gin.H{
			"error": "Invalid request headers",
			"code":  "INVALID_HEADERS",
		})
		return
	}

	// 2. Rate limit check
	if !checkRateLimit(clientIP) {
		c.JSON(429, gin.H{
			"error": "Too many requests, please try again later",
			"code":  "RATE_LIMIT_EXCEEDED",
		})
		return
	}

	// 3. Anomaly detection
	if detectAnomaly(c, clientIP) {
		c.JSON(403, gin.H{
			"error": "Anomalous access detected, request rejected",
			"code":  "ANOMALOUS_ACCESS_DETECTED",
		})
		return
	}

	// 4. Log access record
	logAccess(c, clientIP)

	// 5. Return health check response (with some randomness)
	c.JSON(200, gin.H{
		"message":   "pong",
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}
