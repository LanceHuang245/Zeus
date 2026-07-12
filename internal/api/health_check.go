package api

import (
	"strings"
	"time"

	"Zephyr/internal/application"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	RateLimitPerMinute       = int(application.RateLimitPerMinute)
	RateLimitWindowSeconds   = int(application.RateLimitWindow / time.Second)
	AnomalyThreshold         = int(application.AnomalyThreshold)
	AnomalyWindowMinutes     = int(application.AnomalyWindow / time.Minute)
	AllowedApplicationHeader = "Zephyr"
)

// AccessLog contains structured health check access data
type AccessLog struct {
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Method    string    `json:"method"`
}

// HealthHandler exposes the health check endpoint
type HealthHandler struct {
	service *application.HealthService
}

// NewHealthHandler creates a health check handler
func NewHealthHandler(service *application.HealthService) *HealthHandler {
	return &HealthHandler{service: service}
}

// getClientIP resolves the client IP from proxy headers or the request
func getClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	return c.ClientIP()
}

// validateHeaders checks the required application and user agent headers
func validateHeaders(c *gin.Context) bool {
	if c.GetHeader("Application") != AllowedApplicationHeader {
		return false
	}
	return c.GetHeader("User-Agent") != ""
}

// logAccess records a successful health check access event
func logAccess(c *gin.Context, clientIP string) {
	accessLog := AccessLog{
		IP:        clientIP,
		UserAgent: c.GetHeader("User-Agent"),
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}
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

// logAnomaly records an anomalous health check access event
func logAnomaly(c *gin.Context, clientIP string, count int64, timestamp time.Time) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Error("Anomalous access detected",
		zap.String("ip", clientIP),
		zap.Int("request_count", int(count)),
		zap.Duration("window", application.AnomalyWindow),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.Time("timestamp", timestamp),
	)
}

// HealthCheck validates and processes a health check request
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	clientIP := getClientIP(c)
	if !validateHeaders(c) {
		c.JSON(400, gin.H{"error": "Invalid request headers", "code": "INVALID_HEADERS"})
		return
	}

	result := h.service.Check(c.Request.Context(), clientIP)
	if result.RateLimitExceeded {
		c.JSON(429, gin.H{"error": "Too many requests, please try again later", "code": "RATE_LIMIT_EXCEEDED"})
		return
	}
	if result.AnomalyDetected {
		logAnomaly(c, clientIP, result.AnomalyCount, result.AnomalyTime)
		c.JSON(403, gin.H{"error": "Anomalous access detected, request rejected", "code": "ANOMALOUS_ACCESS_DETECTED"})
		return
	}

	logAccess(c, clientIP)
	c.JSON(200, gin.H{
		"message":   "pong",
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}
