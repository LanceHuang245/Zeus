package utils

// ToWmoCode converts a weather code from a specific source (e.g., "qweather")
// into the corresponding WMO weather interpretation code, which is used by OpenMeteo.
// If the source is already "openmeteo", it returns the code as is.
func ToWmoCode(source string, code int) int {
	if source == "openmeteo" {
		return code // Already in WMO format
	}

	if source == "qweather" {
		switch code {
		// Clear
		case 100, 150:
			return 0
		// Cloudy, Partly Cloudy
		case 101, 102, 103, 151, 152, 153:
			return 1 // Represents "Mainly clear, partly cloudy, and overcast"
		// Overcast
		case 104:
			return 3
		// Fog, Haze, Sand, Dust
		case 500, 501, 502, 503, 504, 507, 508, 509, 510, 511, 512, 513, 514, 515:
			return 45
		// Drizzle
		case 309:
			return 51 // Drizzle: Light intensity
		// Rain
		case 305, 306, 307, 308, 310, 311, 312, 314, 315, 316, 317, 318, 399:
			return 63 // Rain: Moderate intensity
		// Freezing Rain
		case 313:
			return 67 // Freezing Rain: Heavy intensity
		// Rain Shower
		case 300, 301, 350, 351:
			return 80 // Rain showers: Slight
		// Snow & Sleet
		case 400, 401, 402, 403, 407, 408, 409, 410, 499:
			return 73 // Snow fall: Moderate intensity
		case 404, 405, 406, 456, 457: // Rain and snow, Sleet
			return 85 // Snow showers: Slight
		// Thunderstorm
		case 302, 303, 304:
			return 95 // Thunderstorm: Slight or moderate
		// Unknown or other codes (900 Hot, 901 Cold, 999 Unknown)
		default:
			return 1 // Default to Cloudy, matching frontend logic
		}
	}

	// Default for any other unknown source
	return 1
}
