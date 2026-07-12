package qweather

import (
	"context"
	"encoding/json"

	"Zephyr/internal/models"
)

// SearchCities retrieves and filters QWeather city results
func (c *Client) SearchCities(ctx context.Context, location, lang string) ([]models.FilteredSearchResult, error) {
	apiURL := c.baseURL + "/geo/v2/city/lookup?location=" + location + "&lang=" + lang
	body, _, err := c.do(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	var response struct {
		Location []struct {
			Name    string `json:"name"`
			Lat     string `json:"lat"`
			Lon     string `json:"lon"`
			Adm1    string `json:"adm1"`
			Country string `json:"country"`
		} `json:"location"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	var results []models.FilteredSearchResult
	for _, location := range response.Location {
		results = append(results, models.FilteredSearchResult{
			Name: location.Name,
			Lat:  location.Lat,
			Lon:  location.Lon,
			Address: struct {
				State   string `json:"state"`
				Country string `json:"country"`
			}{State: location.Adm1, Country: location.Country},
		})
	}
	return results, nil
}
