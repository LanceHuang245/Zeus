package qweather

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"Zephyr/internal/models"
)

// formatLocation normalizes coordinate precision for cache keys and requests
func formatLocation(location string) string {
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return location
	}
	lon, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err1 != nil || err2 != nil {
		return location
	}
	return fmt.Sprintf("%.2f,%.2f", lon, lat)
}

// Warning retrieves the current QWeather warning response
func (c *Client) Warning(ctx context.Context, location, language string) (models.QWeatherWarningResponse, int, error) {
	location = formatLocation(location)
	cacheKey := fmt.Sprintf("qweather:warning:%s:%s", location, language)
	if c.cache != nil {
		if value, err := c.cache.Get(ctx, cacheKey); err == nil {
			var response models.QWeatherWarningResponse
			if err := json.Unmarshal([]byte(value), &response); err == nil {
				return response, 200, nil
			}
		}
	}

	apiURL := fmt.Sprintf("%s%s?location=%s&lang=%s", c.baseURL, "/v7/warning/now", location, language)
	body, status, err := c.do(ctx, apiURL)
	if err != nil {
		return models.QWeatherWarningResponse{}, 0, err
	}
	var response models.QWeatherWarningResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return models.QWeatherWarningResponse{}, status, &RequestError{Stage: "parse", Body: body, Err: err}
	}
	if c.cache != nil {
		if cacheBytes, err := json.Marshal(response); err == nil {
			_ = c.cache.Set(ctx, cacheKey, cacheBytes, c.cacheTTL)
		}
	}
	return response, status, nil
}
