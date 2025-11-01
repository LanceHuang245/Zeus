package osm

import (
	"Zephyr/internal/config"
	"io"
	"net/http"
)

func SearchCitiesFromOsm(query, acceptLanguage string) ([]byte, error) {
	urlStr := config.OsmUrl + "?format=json&q=" + query + "&accept-language=" + acceptLanguage + "&limit=30&addressdetails=1&featureType=city"

	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
