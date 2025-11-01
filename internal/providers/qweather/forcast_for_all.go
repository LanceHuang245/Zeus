package qweather

import (
	"Zephyr/internal/config"
	"Zephyr/internal/models"
	"Zephyr/internal/providers/qweather/auth"
	"Zephyr/pkg/utils"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type StringFloat64 float64

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

type StringInt int

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

func fetchAPI(apiURL string, target interface{}) error {
	token, err := auth.GenerateJWT()
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	log.Printf("QWeather API Response for %s: %s", apiURL, string(body))

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func fetchNowWeatherData(latitude, longitude, language, unit string) (models.CurrentWeatherResult, error) {
	type qWeatherNowResponse struct {
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

	var response qWeatherNowResponse
	apiURL := fmt.Sprintf("%s/v7/weather/now?location=%s,%s&lang=%s&unit=%s", config.QweatherUrl, longitude, latitude, language, unit)
	if err := fetchAPI(apiURL, &response); err != nil {
		return models.CurrentWeatherResult{}, err
	}

	qNow := response.Now
	currentWeather := models.CurrentWeatherResult{
		Temperature:         float64(qNow.Temp),
		ApparentTemperature: float64(qNow.FeelsLike),
		WeatherCode:         utils.ToWmoCode("qweather", int(qNow.Icon)),
		WindSpeed:           float64(qNow.WindSpeed),
		WindDirection:       float64(qNow.Wind360),
		Humidity:            float64(qNow.Humidity),
		SurfacePressure:     float64(qNow.Pressure),
		Visibility:          float64(qNow.Visibility),
	}

	return currentWeather, nil
}

func fetchNowAirQualityData(latitude, longitude, language, unit string) (models.CurrentWeatherResult, error) {
	type qAirQualityResponse struct {
		Now struct {
			Aqi      string `json:"aqi"`
			Category string `json:"category"`
			Primary  string `json:"primary"`
			Pm10     string `json:"pm10"`
			Pm2p5    string `json:"pm2p5"`
			No2      string `json:"no2"`
			So2      string `json:"so2"`
			Co       string `json:"co"`
			O3       string `json:"o3"`
		} `json:"now"`
	}

	var response qAirQualityResponse
	apiURL := fmt.Sprintf("%s/v7/air/now?location=%s,%s&lang=%s", config.QweatherUrl, longitude, latitude, language)
	if err := fetchAPI(apiURL, &response); err != nil {
		return models.CurrentWeatherResult{}, err
	}

	airQuality := models.CurrentWeatherResult{}
	if aqi, err := strconv.ParseFloat(response.Now.Aqi, 64); err == nil {
		airQuality.AQI = aqi
	}
	if pm25, err := strconv.ParseFloat(response.Now.Pm2p5, 64); err == nil {
		airQuality.Pm25 = pm25
	}
	if pm10, err := strconv.ParseFloat(response.Now.Pm10, 64); err == nil {
		airQuality.Pm10 = pm10
	}
	if o3, err := strconv.ParseFloat(response.Now.O3, 64); err == nil {
		airQuality.Ozone = o3
	}
	if no2, err := strconv.ParseFloat(response.Now.No2, 64); err == nil {
		airQuality.NitrogenDioxide = no2
	}
	if so2, err := strconv.ParseFloat(response.Now.So2, 64); err == nil {
		airQuality.SulfurDioxide = so2
	}

	return airQuality, nil
}

func fetchDailyWeatherData(latitude, longitude, language, unit string) ([]models.DailyWeatherResult, error) {
	type qDailyResponse struct {
		Daily []struct {
			FxDate  string        `json:"fxDate"`
			TempMax StringFloat64 `json:"tempMax"`
			TempMin StringFloat64 `json:"tempMin"`
			IconDay StringInt     `json:"iconDay"`
			UvIndex StringFloat64 `json:"uvIndex"`
		} `json:"daily"`
	}

	var response qDailyResponse
	apiURL := fmt.Sprintf("%s/v7/weather/7d?location=%s,%s&lang=%s&unit=%s", config.QweatherUrl, longitude, latitude, language, unit)
	if err := fetchAPI(apiURL, &response); err != nil {
		return nil, err
	}

	dailyWeathers := make([]models.DailyWeatherResult, 0, len(response.Daily))
	for _, day := range response.Daily {
		dailyWeathers = append(dailyWeathers, models.DailyWeatherResult{
			Date:        day.FxDate,
			TempMax:     float64(day.TempMax),
			TempMin:     float64(day.TempMin),
			UvIndexMax:  float64(day.UvIndex),
			WeatherCode: utils.ToWmoCode("qweather", int(day.IconDay)),
		})
	}
	return dailyWeathers, nil
}

func fetchHourlyWeatherData(latitude, longitude, language, unit string) ([]models.HourlyWeatherResult, error) {
	type qHourlyResponse struct {
		Hourly []struct {
			FxTime    string        `json:"fxTime"`
			Temp      StringFloat64 `json:"temp"`
			Icon      StringInt     `json:"icon"`
			Precip    StringFloat64 `json:"precip"`
			WindSpeed StringFloat64 `json:"windSpeed"`
			Pressure  StringFloat64 `json:"pressure"`
		} `json:"hourly"`
	}

	var response qHourlyResponse
	apiURL := fmt.Sprintf("%s/v7/weather/24h?location=%s,%s&lang=%s&unit=%s", config.QweatherUrl, longitude, latitude, language, unit)
	if err := fetchAPI(apiURL, &response); err != nil {
		return nil, err
	}

	hourlyWeathers := make([]models.HourlyWeatherResult, 0, len(response.Hourly))
	for _, hour := range response.Hourly {
		hourlyWeathers = append(hourlyWeathers, models.HourlyWeatherResult{
			Time:            hour.FxTime,
			Temperature:     float64(hour.Temp),
			WeatherCode:     utils.ToWmoCode("qweather", int(hour.Icon)),
			Precipitation:   float64(hour.Precip),
			WindSpeed:       float64(hour.WindSpeed),
			SurfacePressure: float64(hour.Pressure),
		})
	}
	return hourlyWeathers, nil
}

func GetAllForecastDetails(latitude, longitude, language, unit string) models.WeatherResult {
	// Convert to float64 type
	latFloat, _ := strconv.ParseFloat(latitude, 64)
	lonFloat, _ := strconv.ParseFloat(longitude, 64)

	// Cache geolocation within approximately 1.11 kilometer range
	cacheLatitude := fmt.Sprintf("%.2f", latFloat)
	cacheLongitude := fmt.Sprintf("%.2f", lonFloat)
	cacheKey := fmt.Sprintf("weather:qweather:%s:%s:%s:%s", cacheLatitude, cacheLongitude, language, unit)
	if cachedData, err := config.RedisClient.Get(config.Ctx, cacheKey).Result(); err == nil {
		var weatherResult models.WeatherResult
		if err := json.Unmarshal([]byte(cachedData), &weatherResult); err == nil {
			log.Printf("Retrieved weather data from cache: %s\n", cacheKey)
			return weatherResult
		}
	}

	var wg sync.WaitGroup
	var currentWeatherData models.CurrentWeatherResult
	var airQualityData models.CurrentWeatherResult
	var dailyWeatherData []models.DailyWeatherResult
	var hourlyWeatherData []models.HourlyWeatherResult

	errChan := make(chan error, 4)

	wg.Add(4)
	go func() {
		defer wg.Done()
		var err error
		currentWeatherData, err = fetchNowWeatherData(latitude, longitude, language, unit)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		var err error
		airQualityData, err = fetchNowAirQualityData(latitude, longitude, language, unit)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		var err error
		dailyWeatherData, err = fetchDailyWeatherData(latitude, longitude, language, unit)
		errChan <- err
	}()
	go func() {
		defer wg.Done()
		var err error
		hourlyWeatherData, err = fetchHourlyWeatherData(latitude, longitude, language, unit)
		errChan <- err
	}()
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			log.Printf("Error fetching QWeather data: %v", err)
			return models.WeatherResult{}
		}
	}

	finalCurrent := currentWeatherData
	finalCurrent.AQI = airQualityData.AQI
	finalCurrent.Pm25 = airQualityData.Pm25
	finalCurrent.Pm10 = airQualityData.Pm10
	finalCurrent.Ozone = airQualityData.Ozone
	finalCurrent.NitrogenDioxide = airQualityData.NitrogenDioxide
	finalCurrent.SulfurDioxide = airQualityData.SulfurDioxide

	weatherResult := models.WeatherResult{
		CWR: finalCurrent,
		DWR: dailyWeatherData,
		HWR: hourlyWeatherData,
	}

	if cachedData, err := json.Marshal(weatherResult); err == nil {
		log.Printf("Cached weather data: %s\n", cacheKey)
		config.RedisClient.Set(config.Ctx, cacheKey, cachedData, config.CacheTTL)
	}

	return weatherResult
}
