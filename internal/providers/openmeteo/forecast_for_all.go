package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"Zephyr/internal/models"
	"Zephyr/internal/ports"
)

// Client accesses the OpenMeteo forecast and air quality APIs
type Client struct {
	forecastURL   string
	airQualityURL string
	httpClient    *http.Client
	cache         ports.Cache
	cacheTTL      time.Duration
}

// NewClient creates an OpenMeteo client
func NewClient(forecastURL, airQualityURL string, httpClient *http.Client, cache ports.Cache, cacheTTL time.Duration) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		forecastURL:   forecastURL,
		airQualityURL: airQualityURL,
		httpClient:    httpClient,
		cache:         cache,
		cacheTTL:      cacheTTL,
	}
}

// Forecast retrieves and normalizes forecast data
func (c *Client) Forecast(ctx context.Context, latitude, longitude, language, unit string) (models.WeatherResult, error) {
	latFloat, _ := strconv.ParseFloat(latitude, 64)
	lonFloat, _ := strconv.ParseFloat(longitude, 64)
	cacheLatitude := fmt.Sprintf("%.2f", latFloat)
	cacheLongitude := fmt.Sprintf("%.2f", lonFloat)
	cacheKey := fmt.Sprintf("weather:openmeteo:%s:%s:%s:%s", cacheLatitude, cacheLongitude, language, unit)

	if c.cache != nil {
		if cachedData, err := c.cache.Get(ctx, cacheKey); err == nil {
			var weatherResult models.WeatherResult
			if err := json.Unmarshal([]byte(cachedData), &weatherResult); err == nil {
				return weatherResult, nil
			}
		}
	}

	weatherData, err := c.fetchWeatherData(ctx, latitude, longitude, language, unit)
	if err != nil {
		return models.WeatherResult{}, err
	}
	airQualityData, err := c.fetchAirQualityData(ctx, latitude, longitude)
	if err != nil {
		return models.WeatherResult{}, err
	}

	var weatherMap map[string]interface{}
	if err := json.Unmarshal(weatherData, &weatherMap); err != nil {
		return models.WeatherResult{}, err
	}
	var airQualityMap map[string]interface{}
	if err := json.Unmarshal(airQualityData, &airQualityMap); err != nil {
		return models.WeatherResult{}, err
	}

	var weatherResult models.WeatherResult
	if current, ok := weatherMap["current"].(map[string]interface{}); ok {
		currentWeather := models.CurrentWeatherResult{
			Temperature:         getFloatValue(current, "temperature_2m"),
			WeatherCode:         getIntValue(current, "weather_code"),
			WindSpeed:           getFloatValue(current, "wind_speed_10m"),
			WindDirection:       getFloatValue(current, "winddirection_10m"),
			ApparentTemperature: getFloatValue(current, "apparent_temperature"),
			Humidity:            getFloatValue(current, "relative_humidity_2m"),
			SurfacePressure:     getFloatValue(current, "surface_pressure"),
		}
		if airQualityCurrent, ok := airQualityMap["current"].(map[string]interface{}); ok {
			currentWeather.Pm25 = getFloatValue(airQualityCurrent, "pm2_5")
			currentWeather.Pm10 = getFloatValue(airQualityCurrent, "pm10")
			currentWeather.Ozone = getFloatValue(airQualityCurrent, "ozone")
			currentWeather.NitrogenDioxide = getFloatValue(airQualityCurrent, "nitrogen_dioxide")
			currentWeather.SulfurDioxide = getFloatValue(airQualityCurrent, "sulphur_dioxide")
			currentWeather.AQI = getFloatValue(airQualityCurrent, "european_aqi")
		}
		weatherResult.CWR = currentWeather
	}

	if hourly, ok := weatherMap["hourly"].(map[string]interface{}); ok {
		times := getStringArray(hourly, "time")
		temperatures := getFloatArray(hourly, "temperature_2m")
		weatherCodes := getIntArray(hourly, "weather_code")
		precipitations := getFloatArray(hourly, "precipitation")
		visibilities := getFloatArray(hourly, "visibility")
		windSpeeds := getFloatArray(hourly, "wind_speed_10m")
		pressuresMsl := getFloatArray(hourly, "pressure_msl")
		surfacePressures := getFloatArray(hourly, "surface_pressure")
		for i := 0; i < len(times); i++ {
			weatherResult.HWR = append(weatherResult.HWR, models.HourlyWeatherResult{
				Time:            getValueByIndex(times, i),
				Temperature:     getFloatValueByIndex(temperatures, i),
				WeatherCode:     getIntValueByIndex(weatherCodes, i),
				Precipitation:   getFloatValueByIndex(precipitations, i),
				Visibility:      getFloatValueByIndex(visibilities, i),
				WindSpeed:       getFloatValueByIndex(windSpeeds, i),
				PressureMsl:     getFloatValueByIndex(pressuresMsl, i),
				SurfacePressure: getFloatValueByIndex(surfacePressures, i),
			})
		}
	}

	if daily, ok := weatherMap["daily"].(map[string]interface{}); ok {
		dates := getStringArray(daily, "time")
		tempMaxs := getFloatArray(daily, "temperature_2m_max")
		tempMins := getFloatArray(daily, "temperature_2m_min")
		weatherCodes := getIntArray(daily, "weather_code")
		uvIndexMaxs := getFloatArray(daily, "uv_index_max")
		for i := 0; i < len(dates); i++ {
			weatherResult.DWR = append(weatherResult.DWR, models.DailyWeatherResult{
				Date:        getValueByIndex(dates, i),
				TempMax:     getFloatValueByIndex(tempMaxs, i),
				TempMin:     getFloatValueByIndex(tempMins, i),
				WeatherCode: getIntValueByIndex(weatherCodes, i),
				UvIndexMax:  getFloatValueByIndex(uvIndexMaxs, i),
			})
		}
	}

	if c.cache != nil {
		if cachedData, err := json.Marshal(weatherResult); err == nil {
			_ = c.cache.Set(ctx, cacheKey, cachedData, c.cacheTTL)
		}
	}
	return weatherResult, nil
}

// fetchWeatherData retrieves the weather payload
func (c *Client) fetchWeatherData(ctx context.Context, latitude, longitude, language, unit string) ([]byte, error) {
	urlStr := c.forecastURL + "?latitude=" + latitude + "&longitude=" + longitude +
		"&current=apparent_temperature,temperature_2m,weather_code,relative_humidity_2m,wind_speed_10m,winddirection_10m,surface_pressure" +
		"&hourly=weather_code,temperature_2m,precipitation,visibility,wind_speed_10m,wind_speed_80m,wind_speed_120m,pressure_msl,surface_pressure" +
		"&daily=temperature_2m_max,temperature_2m_min,weather_code,uv_index_max" +
		"&timezone=auto&lang=" + language + "&temperature_unit=" + unit
	return c.get(ctx, urlStr)
}

// fetchAirQualityData retrieves the air quality payload
func (c *Client) fetchAirQualityData(ctx context.Context, latitude, longitude string) ([]byte, error) {
	urlStr := c.airQualityURL + "?latitude=" + latitude + "&longitude=" + longitude +
		"&current=pm2_5,pm10,ozone,nitrogen_dioxide,sulphur_dioxide,european_aqi&timezone=auto"
	return c.get(ctx, urlStr)
}

// get sends a request and reads the response body
func (c *Client) get(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// getFloatValue reads a floating point value from a decoded payload
func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

// getIntValue reads an integer value from a decoded payload
func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// getStringArray reads a string array from a decoded payload
func getStringArray(m map[string]interface{}, key string) []string {
	var result []string
	if arr, ok := m[key].([]interface{}); ok {
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}

// getFloatArray reads a floating point array from a decoded payload
func getFloatArray(m map[string]interface{}, key string) []float64 {
	var result []float64
	if arr, ok := m[key].([]interface{}); ok {
		for _, item := range arr {
			if f, ok := item.(float64); ok {
				result = append(result, f)
			}
		}
	}
	return result
}

// getIntArray reads an integer array from a decoded payload
func getIntArray(m map[string]interface{}, key string) []int {
	var result []int
	if arr, ok := m[key].([]interface{}); ok {
		for _, item := range arr {
			if f, ok := item.(float64); ok {
				result = append(result, int(f))
			}
		}
	}
	return result
}

// getValueByIndex safely reads a string array item
func getValueByIndex(arr []string, index int) string {
	if index < len(arr) {
		return arr[index]
	}
	return ""
}

// getFloatValueByIndex safely reads a floating point array item
func getFloatValueByIndex(arr []float64, index int) float64 {
	if index < len(arr) {
		return arr[index]
	}
	return 0
}

// getIntValueByIndex safely reads an integer array item
func getIntValueByIndex(arr []int, index int) int {
	if index < len(arr) {
		return arr[index]
	}
	return 0
}
