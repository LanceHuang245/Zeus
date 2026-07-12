package models

// CurrentWeatherResult contains normalized current weather data
type CurrentWeatherResult struct {
	Temperature         float64 `json:"temperature"`
	WeatherCode         int     `json:"weather_code"`
	WindSpeed           float64 `json:"wind_speed"`
	WindDirection       float64 `json:"wind_direction"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	Humidity            float64 `json:"humidity"`
	SurfacePressure     float64 `json:"surface_pressure"`
	Pm25                float64 `json:"pm2_5"`
	Pm10                float64 `json:"pm10"`
	Ozone               float64 `json:"ozone"`
	NitrogenDioxide     float64 `json:"nitrogen_dioxide"`
	SulfurDioxide       float64 `json:"sulfur_dioxide"`
	AQI                 float64 `json:"aqi"`
	Visibility          float64 `json:"visibility"`
}

// HourlyWeatherResult contains normalized hourly weather data
type HourlyWeatherResult struct {
	Time            string  `json:"time"`
	Temperature     float64 `json:"temperature"`
	WeatherCode     int     `json:"weather_code"`
	Precipitation   float64 `json:"precipitation"`
	Visibility      float64 `json:"visibility"`
	WindSpeed       float64 `json:"wind_speed"`
	PressureMsl     float64 `json:"pressure_msl"`
	SurfacePressure float64 `json:"surface_pressure"`
}

// DailyWeatherResult contains normalized daily weather data
type DailyWeatherResult struct {
	Date        string  `json:"date"`
	TempMax     float64 `json:"temp_max"`
	TempMin     float64 `json:"temp_min"`
	WeatherCode int     `json:"weather_code"`
	UvIndexMax  float64 `json:"uv_index_max"`
}

// WeatherResult contains current hourly and daily weather data
type WeatherResult struct {
	CWR CurrentWeatherResult  `json:"current"`
	HWR []HourlyWeatherResult `json:"hourly"`
	DWR []DailyWeatherResult  `json:"daily"`
}
