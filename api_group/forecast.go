package apigroup

import (
	"Zephyr/openmeteo"
	"Zephyr/qweather"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Forecast(c *gin.Context) {
	latitude := c.Query("latitude")
	longitude := c.Query("longitude")
	unit := c.Query("unit")
	language := c.Query("accept-language")
	source := c.Query("source")

	switch source {
	case "om":
		weatherResult := openmeteo.GetAllForecastDetails(latitude, longitude, language, unit)
		c.JSON(http.StatusOK, weatherResult)
		return
	case "qweather":
		weatherResult := qweather.GetAllForecastDetails(latitude, longitude, language, unit)
		c.JSON(http.StatusOK, weatherResult)
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported source"})
		return
	}
}
