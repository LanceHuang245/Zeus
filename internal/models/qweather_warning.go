package models

// QWeatherWarningResponse contains the QWeather warning response
type QWeatherWarningResponse struct {
	Code       string            `json:"code"`
	UpdateTime string            `json:"updateTime"`
	FxLink     string            `json:"fxLink"`
	Warning    []QWeatherWarning `json:"warning"`
	Refer      QWeatherRefer     `json:"refer"`
}

// QWeatherWarning contains one weather warning item
type QWeatherWarning struct {
	ID            string `json:"id"`
	Sender        string `json:"sender"`
	PubTime       string `json:"pubTime"`
	Title         string `json:"title"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Status        string `json:"status"`
	Level         string `json:"level"`
	Severity      string `json:"severity"`
	SeverityColor string `json:"severityColor"`
	Type          string `json:"type"`
	TypeName      string `json:"typeName"`
	Urgency       string `json:"urgency"`
	Certainty     string `json:"certainty"`
	Text          string `json:"text"`
	Related       string `json:"related"`
}

// QWeatherRefer contains QWeather source and license references
type QWeatherRefer struct {
	Sources []string `json:"sources"`
	License []string `json:"license"`
}
