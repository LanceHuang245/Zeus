package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// SearchCities handles city search requests from the frontend
func (h *Handler) SearchCities(c *gin.Context) {
	query := c.Query("query")
	encodedQuery := url.QueryEscape(query)
	acceptLanguage := c.Query("accept-language")
	source := c.Query("source")

	places, err := h.citySearchService.SearchCities(c.Request.Context(), source, encodedQuery, acceptLanguage)
	if err != nil {
		if err.Error() == "unsupported source" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported source"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, places)
}
