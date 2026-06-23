package osm

import (
	"Zephyr/internal/config"
	"fmt"
	"io"
	"net/http"
)

func SearchCitiesFromOsm(query, acceptLanguage string) ([]byte, error) {
	urlStr := config.OsmUrl + "?format=json&q=" + query + "&accept-language=" + acceptLanguage + "&limit=30&addressdetails=1&featureType=city"

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Zephyr/2.2.0")

	resp, err := http.DefaultClient.Do(req)
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

	return body, nil
}
