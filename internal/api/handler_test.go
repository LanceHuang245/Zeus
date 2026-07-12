package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"Zephyr/internal/application"
	"Zephyr/internal/models"
	"Zephyr/internal/ports"

	"github.com/gin-gonic/gin"
)

// newTestHandler creates handlers backed by deterministic test providers
func newTestHandler() *Handler {
	return NewHandler(
		application.NewForecastService(map[string]ports.ForecastProvider{"om": testForecastProvider{}}),
		application.NewCitySearchService(map[string]ports.CitySearchProvider{"om": testSearchProvider{}}),
		application.NewWarningService(testWarningProvider{}),
	)
}

// testForecastProvider returns deterministic forecast data
type testForecastProvider struct{}

// Forecast returns a fixed forecast result
func (testForecastProvider) Forecast(context.Context, string, string, string, string) (models.WeatherResult, error) {
	return models.WeatherResult{CWR: models.CurrentWeatherResult{Temperature: 22}}, nil
}

// testSearchProvider returns deterministic city data
type testSearchProvider struct{}

// SearchCities returns a fixed city result
func (testSearchProvider) SearchCities(context.Context, string, string) ([]models.FilteredSearchResult, error) {
	return []models.FilteredSearchResult{{Name: "Shanghai", Lat: "31.23", Lon: "121.47"}}, nil
}

// testWarningProvider returns deterministic warning data
type testWarningProvider struct{}

// Warning returns a fixed warning result
func (testWarningProvider) Warning(context.Context, string, string) (models.QWeatherWarningResponse, int, error) {
	return models.QWeatherWarningResponse{Code: "200"}, http.StatusOK, nil
}

// TestHandlersKeepPublicResponseShape verifies public response compatibility
func TestHandlersKeepPublicResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := newTestHandler()
	router.GET("/forecast", handler.Forecast)
	router.GET("/search", handler.SearchCities)
	router.GET("/warning", handler.WeatherWarning)

	req := httptest.NewRequest(http.MethodGet, "/forecast?source=om&latitude=31.23&longitude=121.47", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("forecast status = %d", recorder.Code)
	}
	var forecast models.WeatherResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &forecast); err != nil || forecast.CWR.Temperature != 22 {
		t.Fatalf("unexpected forecast response: %s, %v", recorder.Body.String(), err)
	}

	query := url.Values{"source": {"om"}, "query": {"New York"}}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/search?"+query.Encode(), nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() == "null" {
		t.Fatalf("unexpected search response: %d %s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/warning?location=121.47,31.23", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"code":"200","updateTime":"","fxLink":"","warning":null,"refer":{"sources":null,"license":null}}` {
		t.Fatalf("unexpected warning response: %d %s", recorder.Code, recorder.Body.String())
	}
}

// TestHandlersRejectUnsupportedSource verifies unsupported source responses
func TestHandlersRejectUnsupportedSource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/forecast", newTestHandler().Forecast)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/forecast?source=unknown", nil))
	if recorder.Code != http.StatusBadRequest || recorder.Body.String() != `{"error":"unsupported source"}` {
		t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
	}
}
