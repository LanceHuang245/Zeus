package qweather

import (
	"Zephyr/internal/config"
	"Zephyr/internal/models"
	"Zephyr/internal/providers/qweather/auth"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func SearchCitiesFromQw(location, lang string) ([]byte, error) {
	token, err := auth.GenerateJWT()
	if err != nil {
		return nil, err
	}

	apiURL := config.QweatherUrl + "/geo/v2/city/lookup?location=" + location + "&lang=" + lang
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	log.Println(token)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyReader := resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		bodyReader = gz
	}
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	log.Println("QWeather API Response:", string(body))

	var qweatherResponse struct {
		Location []struct {
			Name    string `json:"name"`
			Lat     string `json:"lat"`
			Lon     string `json:"lon"`
			Adm1    string `json:"adm1"`
			Country string `json:"country"`
		} `json:"location"`
	}
	if err := json.Unmarshal(body, &qweatherResponse); err != nil {
		return nil, err
	}

	var filteredResults []models.FilteredSearchResult
	for _, loc := range qweatherResponse.Location {
		filteredResult := models.FilteredSearchResult{
			Name: loc.Name,
			Lat:  loc.Lat,
			Lon:  loc.Lon,
			Address: struct {
				State   string `json:"state"`
				Country string `json:"country"`
			}{
				State:   loc.Adm1,
				Country: loc.Country,
			},
		}
		filteredResults = append(filteredResults, filteredResult)
	}

	return json.Marshal(filteredResults)
}
