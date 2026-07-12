package qweather

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"Zephyr/internal/models"
	"Zephyr/internal/ports"
	"Zephyr/internal/providers/qweather/auth"
	"Zephyr/pkg/utils"
)

// Client accesses the QWeather forecast warning and search APIs
type Client struct {
	baseURL     string
	credentials models.QweatherConfig
	httpClient  *http.Client
	cache       ports.Cache
	cacheTTL    time.Duration
}

// RequestError describes a failed QWeather request stage
type RequestError struct {
	Stage string
	Body  []byte
	Err   error
}

// Error returns the underlying request error message
func (e *RequestError) Error() string {
	return e.Err.Error()
}

// Unwrap exposes the underlying request error
func (e *RequestError) Unwrap() error {
	return e.Err
}

// NewClient creates a QWeather client
func NewClient(baseURL string, credentials models.QweatherConfig, httpClient *http.Client, cache ports.Cache, cacheTTL time.Duration) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:     baseURL,
		credentials: credentials,
		httpClient:  httpClient,
		cache:       cache,
		cacheTTL:    cacheTTL,
	}
}

// StringFloat64 accepts numeric values encoded as strings or numbers
type StringFloat64 float64

// UnmarshalJSON decodes a flexible floating point value
func (sf *StringFloat64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		var f float64
		if err := json.Unmarshal(b, &f); err != nil {
			return fmt.Errorf("could not unmarshal as string or float: %w", err)
		}
		*sf = StringFloat64(f)
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("could not parse string to float: %w", err)
	}
	*sf = StringFloat64(f)
	return nil
}

// StringInt accepts integer values encoded as strings or numbers
type StringInt int

// UnmarshalJSON decodes a flexible integer value
func (si *StringInt) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		var i int
		if err := json.Unmarshal(b, &i); err != nil {
			return fmt.Errorf("could not unmarshal as string or int: %w", err)
		}
		*si = StringInt(i)
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("could not parse string to int: %w", err)
	}
	*si = StringInt(i)
	return nil
}

// fetchAPI requests and decodes a QWeather payload
func (c *Client) fetchAPI(ctx context.Context, apiURL string, target interface{}) error {
	body, _, err := c.do(ctx, apiURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

// do sends an authenticated QWeather request and reads its body
func (c *Client) do(ctx context.Context, apiURL string) ([]byte, int, error) {
	token, err := auth.GenerateJWT(c.credentials)
	if err != nil {
		return nil, 0, &RequestError{Stage: "jwt", Err: fmt.Errorf("failed to generate JWT: %w", err)}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, 0, &RequestError{Stage: "request", Err: fmt.Errorf("failed to create HTTP request: %w", err)}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, &RequestError{Stage: "http", Err: fmt.Errorf("failed to execute HTTP request: %w", err)}
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, &RequestError{Stage: "gzip", Err: fmt.Errorf("failed to create gzip reader: %w", err)}
		}
		defer gz.Close()
		reader = gz
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, resp.StatusCode, &RequestError{Stage: "read", Err: fmt.Errorf("failed to read response body: %w", err)}
	}
	return body, resp.StatusCode, nil
}

// Forecast retrieves and normalizes QWeather forecast data
func (c *Client) Forecast(ctx context.Context, latitude, longitude, language, unit string) (models.WeatherResult, error) {
	latFloat, _ := strconv.ParseFloat(latitude, 64)
	lonFloat, _ := strconv.ParseFloat(longitude, 64)
	cacheLatitude := fmt.Sprintf("%.2f", latFloat)
	cacheLongitude := fmt.Sprintf("%.2f", lonFloat)
	cacheKey := fmt.Sprintf("weather:qweather:%s:%s:%s:%s", cacheLatitude, cacheLongitude, language, unit)

	if c.cache != nil {
		if cachedData, err := c.cache.Get(ctx, cacheKey); err == nil {
			var weatherResult models.WeatherResult
			if err := json.Unmarshal([]byte(cachedData), &weatherResult); err == nil {
				return weatherResult, nil
			}
		}
	}

	var currentWeatherData models.CurrentWeatherResult
	var airQualityData models.CurrentWeatherResult
	var dailyWeatherData []models.DailyWeatherResult
	var hourlyWeatherData []models.HourlyWeatherResult
	var wg sync.WaitGroup
	errChan := make(chan error, 4)
	wg.Add(4)
	go func() {
		defer wg.Done()
		var err error
		currentWeatherData, err = c.fetchNowWeatherData(ctx, latitude, longitude, language, unit)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		var err error
		airQualityData, err = c.fetchNowAirQualityData(ctx, latitude, longitude, language)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		var err error
		dailyWeatherData, err = c.fetchDailyWeatherData(ctx, latitude, longitude, language, unit)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		var err error
		hourlyWeatherData, err = c.fetchHourlyWeatherData(ctx, latitude, longitude, language, unit)
		errChan <- err
	}()
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return models.WeatherResult{}, err
		}
	}

	currentWeatherData.AQI = airQualityData.AQI
	currentWeatherData.Pm25 = airQualityData.Pm25
	currentWeatherData.Pm10 = airQualityData.Pm10
	currentWeatherData.Ozone = airQualityData.Ozone
	currentWeatherData.NitrogenDioxide = airQualityData.NitrogenDioxide
	currentWeatherData.SulfurDioxide = airQualityData.SulfurDioxide
	weatherResult := models.WeatherResult{CWR: currentWeatherData, DWR: dailyWeatherData, HWR: hourlyWeatherData}

	if c.cache != nil {
		if cachedData, err := json.Marshal(weatherResult); err == nil {
			_ = c.cache.Set(ctx, cacheKey, cachedData, c.cacheTTL)
		}
	}
	return weatherResult, nil
}

// fetchNowWeatherData retrieves current weather data
func (c *Client) fetchNowWeatherData(ctx context.Context, latitude, longitude, language, unit string) (models.CurrentWeatherResult, error) {
	type response struct {
		Now struct {
			Temp       StringFloat64 `json:"temp"`
			FeelsLike  StringFloat64 `json:"feelsLike"`
			Icon       StringInt     `json:"icon"`
			Wind360    StringFloat64 `json:"wind360"`
			WindSpeed  StringFloat64 `json:"windSpeed"`
			Humidity   StringFloat64 `json:"humidity"`
			Pressure   StringFloat64 `json:"pressure"`
			Visibility StringFloat64 `json:"vis"`
		} `json:"now"`
	}
	var result response
	apiURL := fmt.Sprintf("%s/v7/weather/now?location=%s,%s&lang=%s&unit=%s", c.baseURL, longitude, latitude, language, unit)
	if err := c.fetchAPI(ctx, apiURL, &result); err != nil {
		return models.CurrentWeatherResult{}, err
	}
	return models.CurrentWeatherResult{
		Temperature:         float64(result.Now.Temp),
		ApparentTemperature: float64(result.Now.FeelsLike),
		WeatherCode:         utils.ToWmoCode("qweather", int(result.Now.Icon)),
		WindSpeed:           float64(result.Now.WindSpeed),
		WindDirection:       float64(result.Now.Wind360),
		Humidity:            float64(result.Now.Humidity),
		SurfacePressure:     float64(result.Now.Pressure),
		Visibility:          float64(result.Now.Visibility),
	}, nil
}

// fetchNowAirQualityData retrieves current air quality data
func (c *Client) fetchNowAirQualityData(ctx context.Context, latitude, longitude, language string) (models.CurrentWeatherResult, error) {
	type response struct {
		Now struct {
			Aqi  string `json:"aqi"`
			Pm10 string `json:"pm10"`
			Pm25 string `json:"pm2p5"`
			No2  string `json:"no2"`
			So2  string `json:"so2"`
			O3   string `json:"o3"`
		} `json:"now"`
	}
	var result response
	apiURL := fmt.Sprintf("%s/v7/air/now?location=%s,%s&lang=%s", c.baseURL, longitude, latitude, language)
	if err := c.fetchAPI(ctx, apiURL, &result); err != nil {
		return models.CurrentWeatherResult{}, err
	}
	toFloat := func(value string) float64 {
		parsed, _ := strconv.ParseFloat(value, 64)
		return parsed
	}
	return models.CurrentWeatherResult{
		AQI:             toFloat(result.Now.Aqi),
		Pm25:            toFloat(result.Now.Pm25),
		Pm10:            toFloat(result.Now.Pm10),
		Ozone:           toFloat(result.Now.O3),
		NitrogenDioxide: toFloat(result.Now.No2),
		SulfurDioxide:   toFloat(result.Now.So2),
	}, nil
}

// fetchDailyWeatherData retrieves daily forecast data
func (c *Client) fetchDailyWeatherData(ctx context.Context, latitude, longitude, language, unit string) ([]models.DailyWeatherResult, error) {
	type response struct {
		Daily []struct {
			FxDate  string        `json:"fxDate"`
			TempMax StringFloat64 `json:"tempMax"`
			TempMin StringFloat64 `json:"tempMin"`
			IconDay StringInt     `json:"iconDay"`
			UvIndex StringFloat64 `json:"uvIndex"`
		} `json:"daily"`
	}
	var result response
	apiURL := fmt.Sprintf("%s/v7/weather/7d?location=%s,%s&lang=%s&unit=%s", c.baseURL, longitude, latitude, language, unit)
	if err := c.fetchAPI(ctx, apiURL, &result); err != nil {
		return nil, err
	}
	daily := make([]models.DailyWeatherResult, 0, len(result.Daily))
	for _, day := range result.Daily {
		daily = append(daily, models.DailyWeatherResult{
			Date:        day.FxDate,
			TempMax:     float64(day.TempMax),
			TempMin:     float64(day.TempMin),
			UvIndexMax:  float64(day.UvIndex),
			WeatherCode: utils.ToWmoCode("qweather", int(day.IconDay)),
		})
	}
	return daily, nil
}

// fetchHourlyWeatherData retrieves hourly forecast data
func (c *Client) fetchHourlyWeatherData(ctx context.Context, latitude, longitude, language, unit string) ([]models.HourlyWeatherResult, error) {
	type response struct {
		Hourly []struct {
			FxTime    string        `json:"fxTime"`
			Temp      StringFloat64 `json:"temp"`
			Icon      StringInt     `json:"icon"`
			Precip    StringFloat64 `json:"precip"`
			WindSpeed StringFloat64 `json:"windSpeed"`
			Pressure  StringFloat64 `json:"pressure"`
		} `json:"hourly"`
	}
	var result response
	apiURL := fmt.Sprintf("%s/v7/weather/24h?location=%s,%s&lang=%s&unit=%s", c.baseURL, longitude, latitude, language, unit)
	if err := c.fetchAPI(ctx, apiURL, &result); err != nil {
		return nil, err
	}
	hourly := make([]models.HourlyWeatherResult, 0, len(result.Hourly))
	for _, hour := range result.Hourly {
		hourly = append(hourly, models.HourlyWeatherResult{
			Time:            hour.FxTime,
			Temperature:     float64(hour.Temp),
			WeatherCode:     utils.ToWmoCode("qweather", int(hour.Icon)),
			Precipitation:   float64(hour.Precip),
			WindSpeed:       float64(hour.WindSpeed),
			SurfacePressure: float64(hour.Pressure),
		})
	}
	return hourly, nil
}
