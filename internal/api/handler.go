package api

import (
	"errors"
	"net/http"

	"Zephyr/internal/application"
	"Zephyr/internal/providers/qweather"

	"github.com/gin-gonic/gin"
)

// Handler contains the inbound HTTP handlers for the application use cases
type Handler struct {
	forecastService   *application.ForecastService
	citySearchService *application.CitySearchService
	warningService    *application.WarningService
}

// NewHandler creates an HTTP handler with application services
func NewHandler(
	forecastService *application.ForecastService,
	citySearchService *application.CitySearchService,
	warningService *application.WarningService,
) *Handler {
	return &Handler{
		forecastService:   forecastService,
		citySearchService: citySearchService,
		warningService:    warningService,
	}
}

// WeatherWarning handles weather warning requests
func (h *Handler) WeatherWarning(c *gin.Context) {
	location := c.Query("location")
	lang := c.DefaultQuery("lang", "zh")
	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location parameter is required"})
		return
	}

	response, status, err := h.warningService.Warning(c.Request.Context(), location, lang)
	if err != nil {
		var requestErr *qweather.RequestError
		if errors.As(err, &requestErr) {
			switch requestErr.Stage {
			case "jwt":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT generation failed"})
			case "http":
				c.JSON(http.StatusBadGateway, gin.H{"error": "QWeather request failed"})
			case "gzip":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gzip decompression failed"})
			case "read":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read QWeather response"})
			case "parse":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse QWeather response", "body": string(requestErr.Body)})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": requestErr.Error()})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(status, response)
}
