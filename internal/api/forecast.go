package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Forecast handles forecast requests while preserving the legacy response contract
func (h *Handler) Forecast(c *gin.Context) {
	latitude := c.Query("latitude")
	longitude := c.Query("longitude")
	unit := c.Query("unit")
	language := c.Query("accept-language")
	source := c.Query("source")

	weatherResult, err := h.forecastService.Forecast(c.Request.Context(), source, latitude, longitude, language, unit)
	if err != nil {
		// Preserve the legacy HTTP 200 response with an empty result on provider failure
		// This keeps the frontend contract compatible
		if err.Error() != "unsupported source" {
			c.JSON(http.StatusOK, weatherResult)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported source"})
		return
	}
	c.JSON(http.StatusOK, weatherResult)
}
