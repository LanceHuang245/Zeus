package application

import (
	"context"
	"fmt"

	"Zephyr/internal/models"
	"Zephyr/internal/ports"
)

// ForecastService coordinates forecast providers without knowing their transport or vendor-specific details
type ForecastService struct {
	providers map[string]ports.ForecastProvider
}

// NewForecastService creates a forecast application service
func NewForecastService(providers map[string]ports.ForecastProvider) *ForecastService {
	return &ForecastService{providers: providers}
}

// Forecast selects a provider and retrieves forecast data
func (s *ForecastService) Forecast(ctx context.Context, source, latitude, longitude, language, unit string) (models.WeatherResult, error) {
	provider, ok := s.providers[source]
	if !ok {
		return models.WeatherResult{}, fmt.Errorf("unsupported source")
	}
	return provider.Forecast(ctx, latitude, longitude, language, unit)
}

// CitySearchService coordinates location search providers
type CitySearchService struct {
	providers map[string]ports.CitySearchProvider
}

// NewCitySearchService creates a city search application service
func NewCitySearchService(providers map[string]ports.CitySearchProvider) *CitySearchService {
	return &CitySearchService{providers: providers}
}

// SearchCities selects a provider and retrieves city results
func (s *CitySearchService) SearchCities(ctx context.Context, source, query, acceptLanguage string) ([]models.FilteredSearchResult, error) {
	provider, ok := s.providers[source]
	if !ok {
		return nil, fmt.Errorf("unsupported source")
	}
	return provider.SearchCities(ctx, query, acceptLanguage)
}

// WarningService keeps warning orchestration independent from Gin and HTTP
type WarningService struct {
	provider ports.WarningProvider
}

// NewWarningService creates a warning application service
func NewWarningService(provider ports.WarningProvider) *WarningService {
	return &WarningService{provider: provider}
}

// Warning retrieves a warning response from the configured provider
func (s *WarningService) Warning(ctx context.Context, location, language string) (models.QWeatherWarningResponse, int, error) {
	return s.provider.Warning(ctx, location, language)
}
