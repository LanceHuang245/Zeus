package openmeteo

import (
	"Zephyr/config"
	"Zephyr/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

func fetchWeatherData(latitude, longitude, language, unit string) ([]byte, error) {
	urlStr := config.OmForcastUrl + "?latitude=" + latitude + "&longitude=" + longitude +
		"&current=apparent_temperature,temperature_2m,weather_code,relative_humidity_2m,wind_speed_10m,winddirection_10m,surface_pressure" +
		"&hourly=weather_code,temperature_2m,precipitation,visibility,wind_speed_10m,wind_speed_80m,wind_speed_120m,pressure_msl,surface_pressure" +
		"&daily=temperature_2m_max,temperature_2m_min,weather_code,uv_index_max" +
		"&timezone=auto" + "&lang=" + language + "&temperature_unit=" + unit
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

func fetchAirQualityData(latitude, longitude string) ([]byte, error) {
	urlStr := config.OmAirQualityUrl + "?latitude=" + latitude + "&longitude=" + longitude +
		"&current=pm2_5,pm10,ozone,nitrogen_dioxide,sulphur_dioxide,european_aqi" +
		"&timezone=auto"
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

func GetAllForecastDetails(latitude, longitude, language, unit string) models.WeatherResult {
	// convert to float64
	latFloat, _ := strconv.ParseFloat(latitude, 64)
	lonFloat, _ := strconv.ParseFloat(longitude, 64)

	// Geolocation cached within an approximate range of 1.11 kilometers
	cacheLatitude := fmt.Sprintf("%.2f", latFloat)
	cacheLongitude := fmt.Sprintf("%.2f", lonFloat)

	cacheKey := fmt.Sprintf("weather:openmeteo:%s:%s:%s:%s", cacheLatitude, cacheLongitude, language, unit)
	if cachedData, err := config.RedisClient.Get(config.Ctx, cacheKey).Result(); err == nil {
		var weatherResult models.WeatherResult
		err := json.Unmarshal([]byte(cachedData), &weatherResult)
		if err == nil {
			log.Printf("Retrieved weather data from cache: %s\n", cacheKey)
			return weatherResult
		}
	}
	weatherData, err := fetchWeatherData(latitude, longitude, language, unit)
	if err != nil {
		return models.WeatherResult{}
	}

	airQualityData, err := fetchAirQualityData(latitude, longitude)
	if err != nil {
		return models.WeatherResult{}
	}

	var weatherResult models.WeatherResult

	var weatherMap map[string]interface{}
	err = json.Unmarshal(weatherData, &weatherMap)
	if err != nil {
		return models.WeatherResult{}
	}

	var airQualityMap map[string]interface{}
	err = json.Unmarshal(airQualityData, &airQualityMap)
	if err != nil {
		return models.WeatherResult{}
	}

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

		// Add air quality data
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
			hourlyWeather := models.HourlyWeatherResult{
				Time:            getValueByIndex(times, i),
				Temperature:     getFloatValueByIndex(temperatures, i),
				WeatherCode:     getIntValueByIndex(weatherCodes, i),
				Precipitation:   getFloatValueByIndex(precipitations, i),
				Visibility:      getFloatValueByIndex(visibilities, i),
				WindSpeed:       getFloatValueByIndex(windSpeeds, i),
				PressureMsl:     getFloatValueByIndex(pressuresMsl, i),
				SurfacePressure: getFloatValueByIndex(surfacePressures, i),
			}
			weatherResult.HWR = append(weatherResult.HWR, hourlyWeather)
		}
	}

	if daily, ok := weatherMap["daily"].(map[string]interface{}); ok {
		dates := getStringArray(daily, "time")
		tempMaxs := getFloatArray(daily, "temperature_2m_max")
		tempMins := getFloatArray(daily, "temperature_2m_min")
		weatherCodes := getIntArray(daily, "weather_code")
		uvIndexMaxs := getFloatArray(daily, "uv_index_max")

		for i := 0; i < len(dates); i++ {
			dailyWeather := models.DailyWeatherResult{
				Date:        getValueByIndex(dates, i),
				TempMax:     getFloatValueByIndex(tempMaxs, i),
				TempMin:     getFloatValueByIndex(tempMins, i),
				WeatherCode: getIntValueByIndex(weatherCodes, i),
				UvIndexMax:  getFloatValueByIndex(uvIndexMaxs, i),
			}
			weatherResult.DWR = append(weatherResult.DWR, dailyWeather)
		}
	}
	if cachedData, err := json.Marshal(weatherResult); err == nil {
		log.Printf("Cached weather data: %s\n", cacheKey)
		config.RedisClient.Set(config.Ctx, cacheKey, cachedData, config.CacheTTL)
	}

	return weatherResult
}

// Retrieve float64 values from the map
func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return 0
}

// Retrieve integer value from the map
func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// Retrieve a string array from the map
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

// Retrieve a float64 array from the map
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

// Retrieve an int array from the map
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

// Retrieve values from a string array based on their index
func getValueByIndex(arr []string, index int) string {
	if index < len(arr) {
		return arr[index]
	}
	return ""
}

// Retrieve values from a float64 array based on an index
func getFloatValueByIndex(arr []float64, index int) float64 {
	if index < len(arr) {
		return arr[index]
	}
	return 0
}

// Retrieve values from an int array based on their index
func getIntValueByIndex(arr []int, index int) int {
	if index < len(arr) {
		return arr[index]
	}
	return 0
}
