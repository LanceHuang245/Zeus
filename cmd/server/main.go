package main

import (
	"Zephyr/internal/adapters/outbound/redis"
	"Zephyr/internal/api"
	"Zephyr/internal/application"
	"Zephyr/internal/config"
	"Zephyr/internal/ports"
	"Zephyr/internal/providers/openmeteo"
	osm "Zephyr/internal/providers/openstreetmap"
	"Zephyr/internal/providers/qweather"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// main builds the application and starts the HTTP server
func main() {
	// Load configuration from the environment file
	config.LoadConfig()

	cfg := config.Snapshot()
	redisClient := config.NewRedisClient(cfg)
	cache := redis.NewCache(redisClient)
	httpClient := &http.Client{Timeout: 15 * time.Second}

	openMeteoClient := openmeteo.NewClient(
		cfg.OpenMeteoURL,
		cfg.AirQualityURL,
		httpClient,
		cache,
		cfg.CacheTTL,
	)
	openStreetMapClient := osm.NewClient(cfg.OpenStreetMapURL, httpClient)
	qweatherClient := qweather.NewClient(
		cfg.QweatherURL,
		cfg.QweatherConfig,
		httpClient,
		cache,
		cfg.CacheTTL,
	)

	forecastService := application.NewForecastService(map[string]ports.ForecastProvider{
		"om":       openMeteoClient,
		"qweather": qweatherClient,
	})
	citySearchService := application.NewCitySearchService(map[string]ports.CitySearchProvider{
		"om":       openStreetMapClient,
		"qweather": qweatherClient,
	})
	warningService := application.NewWarningService(qweatherClient)
	handlers := api.NewHandler(forecastService, citySearchService, warningService)
	healthHandler := api.NewHealthHandler(application.NewHealthService(redis.NewAccessStore(redisClient)))

	r := gin.Default()

	// Register API routes
	r.GET("/api/v1/city/search", handlers.SearchCities)
	r.GET("/api/v1/weather/alert", handlers.WeatherWarning)
	r.GET("/api/v1/weather/forecast", handlers.Forecast)
	r.GET("/api/v1/healthcheck", healthHandler.HealthCheck)

	// Start the server with the configured transport
	if cfg.EnableTLS {
		log.Printf("Starting HTTPS server on %s", cfg.ServerPort)
		r.RunTLS(cfg.ServerPort, cfg.CertFile, cfg.KeyFile)
	} else {
		log.Printf("Starting HTTP server on %s", cfg.ServerPort)
		r.Run(cfg.ServerPort)
	}
}
