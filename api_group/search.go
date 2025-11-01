package apigroup

import (
	"Zephyr/models"
	osm "Zephyr/openstreetmap"
	"Zephyr/qweather"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func SearchCities(c *gin.Context) {
	query := c.Query("query")
	encodedQuery := url.QueryEscape(query)
	acceptLanguage := c.Query("accept-language")
	source := c.Query("source")

	switch source {
	case "om":
		resp, err := osm.SearchCitiesFromOsm(encodedQuery, acceptLanguage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var places []models.FilteredSearchResult
		if err := json.Unmarshal(resp, &places); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, places)
		return
	case "qweather":
		resp, err := qweather.SearchCitiesFromQw(encodedQuery, acceptLanguage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var places []models.FilteredSearchResult
		if err := json.Unmarshal(resp, &places); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, places)
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported source"})
		return
	}
}
