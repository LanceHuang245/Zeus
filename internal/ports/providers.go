package ports

import (
	"context"
	"time"

	"Zephyr/internal/models"
)

// ForecastProvider defines the forecast capability required by the application
type ForecastProvider interface {
	Forecast(ctx context.Context, latitude, longitude, language, unit string) (models.WeatherResult, error)
}

// CitySearchProvider defines the city search capability required by the application
type CitySearchProvider interface {
	SearchCities(ctx context.Context, query, acceptLanguage string) ([]models.FilteredSearchResult, error)
}

// WarningProvider defines the warning capability required by the application
type WarningProvider interface {
	Warning(ctx context.Context, location, language string) (models.QWeatherWarningResponse, int, error)
}

// AccessStore defines the health access persistence operations
type AccessStore interface {
	RecordAndCount(ctx context.Context, key string, now time.Time, window time.Duration, member interface{}) (int64, error)
}
