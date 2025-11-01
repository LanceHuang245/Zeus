package main

import (
	"Zephyr/internal/api"
	"Zephyr/internal/config"
	"Zephyr/internal/providers/qweather"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from .env file
	config.LoadConfig()

	// Initialize Redis
	config.InitRedis()

	r := gin.Default()

	// API routes
	r.GET("/api/v1/city/search", api.SearchCities)
	r.GET("/api/v1/weather/alert", qweather.WeatherWarningFromQweather)
	r.GET("/api/v1/weather/forecast", api.Forecast)
	r.GET("/api/v1/healthcheck", api.HealthCheck)

	// Start server with configuration
	if config.EnableTLS {
		log.Printf("Starting HTTPS server on %s", config.ServerPort)
		r.RunTLS(config.ServerPort, config.CertFile, config.KeyFile)
	} else {
		log.Printf("Starting HTTP server on %s", config.ServerPort)
		r.Run(config.ServerPort)
	}
}
