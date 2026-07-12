package application

import (
	"context"
	"testing"

	"Zephyr/internal/models"
	"Zephyr/internal/ports"
)

// fakeForecastProvider returns deterministic forecast data for service tests
type fakeForecastProvider struct{}

// Forecast returns a fixed forecast result
func (fakeForecastProvider) Forecast(context.Context, string, string, string, string) (models.WeatherResult, error) {
	return models.WeatherResult{CWR: models.CurrentWeatherResult{Temperature: 21.5}}, nil
}

// fakeSearchProvider returns deterministic city data for service tests
type fakeSearchProvider struct{}

// SearchCities returns a fixed city result
func (fakeSearchProvider) SearchCities(context.Context, string, string) ([]models.FilteredSearchResult, error) {
	return []models.FilteredSearchResult{{Name: "Shanghai", Lat: "31.23", Lon: "121.47"}}, nil
}

// fakeWarningProvider returns deterministic warning data for service tests
type fakeWarningProvider struct{}

// Warning returns a fixed warning result
func (fakeWarningProvider) Warning(context.Context, string, string) (models.QWeatherWarningResponse, int, error) {
	return models.QWeatherWarningResponse{Code: "200"}, 200, nil
}

// TestServicesRouteBySource verifies source based provider routing
func TestServicesRouteBySource(t *testing.T) {
	ctx := context.Background()
	forecast := NewForecastService(map[string]ports.ForecastProvider{"om": fakeForecastProvider{}})
	search := NewCitySearchService(map[string]ports.CitySearchProvider{"om": fakeSearchProvider{}})
	warning := NewWarningService(fakeWarningProvider{})

	forecastResult, err := forecast.Forecast(ctx, "om", "31.23", "121.47", "zh", "celsius")
	if err != nil || forecastResult.CWR.Temperature != 21.5 {
		t.Fatalf("unexpected forecast result: %+v, %v", forecastResult, err)
	}
	places, err := search.SearchCities(ctx, "om", "Shanghai", "zh")
	if err != nil || len(places) != 1 || places[0].Name != "Shanghai" {
		t.Fatalf("unexpected search result: %+v, %v", places, err)
	}
	warningResult, status, err := warning.Warning(ctx, "121.47,31.23", "zh")
	if err != nil || status != 200 || warningResult.Code != "200" {
		t.Fatalf("unexpected warning result: %+v, %d, %v", warningResult, status, err)
	}
}

// TestServicesRejectUnsupportedSource verifies unsupported source handling
func TestServicesRejectUnsupportedSource(t *testing.T) {
	_, err := NewForecastService(nil).Forecast(context.Background(), "unknown", "", "", "", "")
	if err == nil || err.Error() != "unsupported source" {
		t.Fatalf("expected unsupported source error, got %v", err)
	}
}
