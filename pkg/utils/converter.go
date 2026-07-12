package utils

// ToWmoCode converts a source weather code into the corresponding WMO code
// OpenMeteo already uses WMO codes so those values are returned unchanged
func ToWmoCode(source string, code int) int {
	if source == "openmeteo" {
		return code // Already uses WMO format
	}

	if source == "qweather" {
		switch code {
		// Clear
		case 100, 150:
			return 0
		// Cloudy and partly cloudy
		case 101, 102, 103, 151, 152, 153:
			return 1 // Represents mainly clear partly cloudy and overcast conditions
		// Overcast
		case 104:
			return 3
		// Fog haze sand and dust
		case 500, 501, 502, 503, 504, 507, 508, 509, 510, 511, 512, 513, 514, 515:
			return 45
		// Drizzle
		case 309:
			return 51 // Light drizzle
		// Rain
		case 305, 306, 307, 308, 310, 311, 312, 314, 315, 316, 317, 318, 399:
			return 63 // Moderate rain
		// Freezing rain
		case 313:
			return 67 // Heavy freezing rain
		// Rain shower
		case 300, 301, 350, 351:
			return 80 // Slight rain showers
		// Snow and sleet
		case 400, 401, 402, 403, 407, 408, 409, 410, 499:
			return 73 // Moderate snowfall
		case 404, 405, 406, 456, 457: // Rain snow and sleet
			return 85 // Slight snow showers
		// Thunderstorm
		case 302, 303, 304:
			return 95 // Slight or moderate thunderstorm
		// Unknown or other codes including hot cold and unknown values
		default:
			return 1 // Default to cloudy to match frontend logic
		}
	}

	// Default for any other unknown source
	return 1
}
