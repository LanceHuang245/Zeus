package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Zephyr/internal/application"

	"github.com/gin-gonic/gin"
)

// testAccessStore reports a single access so no policy is triggered
type testAccessStore struct{}

// RecordAndCount returns a count below every policy threshold
func (testAccessStore) RecordAndCount(context.Context, string, time.Time, time.Duration, interface{}) (int64, error) {
	return 1, nil
}

// TestHealthCheckAcceptsGetAndHeadWithoutHeaders verifies uptime checkers can probe the endpoint
func TestHealthCheckAcceptsGetAndHeadWithoutHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHealthHandler(application.NewHealthService(testAccessStore{}))
	router.Match([]string{http.MethodGet, http.MethodHead}, "/healthcheck", handler.HealthCheck)

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		req := httptest.NewRequest(method, "/healthcheck", nil)
		req.Header.Del("User-Agent")
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body = %s", method, recorder.Code, recorder.Body.String())
		}
	}
}
