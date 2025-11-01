package qweather

import (
	"Zephyr/internal/config"
	"Zephyr/internal/models"
	"Zephyr/internal/providers/qweather/auth"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func formatLocation(location string) string {
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return location
	}
	lon, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return location // Parsing failed, return directly
	}
	return fmt.Sprintf("%.2f,%.2f", lon, lat)
}

func WeatherWarningFromQweather(c *gin.Context) {
	location := c.Query("location")
	location = formatLocation(location)
	lang := c.DefaultQuery("lang", "zh")

	if location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "location parameter is required"})
		return
	}

	cacheKey := fmt.Sprintf("qweather:warning:%s:%s", location, lang)
	// 1. Check cache first
	if val, err := config.RedisClient.Get(config.Ctx, cacheKey).Result(); err == nil {
		var resp models.QWeatherWarningResponse
		if err := json.Unmarshal([]byte(val), &resp); err == nil {
			log.Printf("Retrieved warning data from cache: %s\n", cacheKey)
			c.JSON(http.StatusOK, resp)
			return
		}
	}

	// 2. Generate JWT
	token, err := auth.GenerateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT generation failed"})
		return
	}

	// 3. Request QWeather
	apiURL := fmt.Sprintf("%s%s?location=%s&lang=%s", config.QweatherUrl, "/v7/warning/now", location, lang)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "QWeather request failed"})
		return
	}
	defer resp.Body.Close()

	// Check if gzip compressed
	bodyReader := resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gzip decompression failed"})
			return
		}
		defer gz.Close()
		bodyReader = gz
	}
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read QWeather response"})
		return
	}

	// 4. Parse into struct
	var warningResp models.QWeatherWarningResponse
	if err := json.Unmarshal(body, &warningResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse QWeather response", "body": string(body)})
		return
	}

	// 5. Cache the serialized JSON of the struct
	cacheBytes, _ := json.Marshal(warningResp)
	config.RedisClient.Set(config.Ctx, cacheKey, cacheBytes, config.CacheTTL)

	// 6. Return struct
	c.JSON(resp.StatusCode, warningResp)
}
