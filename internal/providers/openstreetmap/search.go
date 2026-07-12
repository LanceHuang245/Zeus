package osm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"Zephyr/internal/models"
)

// Client accesses the OpenStreetMap city search API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an OpenStreetMap client
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

// SearchCities retrieves and filters city search results
func (c *Client) SearchCities(ctx context.Context, query, acceptLanguage string) ([]models.FilteredSearchResult, error) {
	// The query is already escaped by the HTTP adapter to preserve legacy request semantics
	urlStr := c.baseURL + "?format=json&q=" + query + "&accept-language=" + acceptLanguage + "&limit=30&addressdetails=1&featureType=city"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Zephyr/2.2.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Nominatim API returned status %d: %s", resp.StatusCode, string(body))
	}

	var places []models.FilteredSearchResult
	if err := json.Unmarshal(body, &places); err != nil {
		return nil, err
	}
	return places, nil
}
