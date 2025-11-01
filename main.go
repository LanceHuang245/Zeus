package main

import (
	apigroup "Zephyr/api_group"
	"Zephyr/config"
	"Zephyr/qweather"
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
	r.GET("/api/v1/city/search", apigroup.SearchCities)
	r.GET("/api/v1/weather/alert", qweather.WeatherWarningFromQweather)
	r.GET("/api/v1/weather/forecast", apigroup.Forecast)
	r.GET("/api/v1/healthcheck", apigroup.HealthCheck)

	// Start server with configuration
	if config.EnableTLS {
		log.Printf("Starting HTTPS server on %s", config.ServerPort)
		r.RunTLS(config.ServerPort, config.CertFile, config.KeyFile)
	} else {
		log.Printf("Starting HTTP server on %s", config.ServerPort)
		r.Run(config.ServerPort)
	}
}
