// Package datetime provides datetime handling for the MCP server
package datetime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/steve/llm-agents/internal/utils"
)

// Handler implements datetime MCP method handling
type Handler struct {
	cityTimezones map[string]string
}

// NewHandler creates a new datetime handler
func NewHandler() *Handler {
	return &Handler{
		cityTimezones: getCityTimezones(),
	}
}

// Handle handles the getDateTime method
func (h *Handler) Handle(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var request struct {
		City string `json:"city"`
	}

	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if request.City == "" {
		return nil, fmt.Errorf("city parameter is required")
	}

	// Normalize city name
	normalizedCity := h.normalizeCity(request.City)

	// Get timezone for the city
	timezone, ok := h.cityTimezones[strings.ToLower(normalizedCity)]
	if !ok {
		return nil, fmt.Errorf("city not found: %s", request.City)
	}

	// Load the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone %s: %w", timezone, err)
	}

	// Get current time in the city's timezone
	now := time.Now().In(loc)

	// Calculate UTC offset
	_, offset := now.Zone()
	offsetHours := offset / 3600
	offsetMins := (offset % 3600) / 60
	utcOffset := fmt.Sprintf("%+03d:%02d", offsetHours, offsetMins)

	result := struct {
		DateTime  string `json:"datetime"`
		Timezone  string `json:"timezone"`
		UTCOffset string `json:"utc_offset"`
	}{
		DateTime:  now.Format(time.RFC3339),
		Timezone:  timezone,
		UTCOffset: utcOffset,
	}

	utils.Debug("DateTime data retrieved for %s: %s (%s)", request.City, result.DateTime, timezone)
	return result, nil
}

// getCityTimezones returns a mapping of US cities to their IANA timezones
func getCityTimezones() map[string]string {
	return map[string]string{
		// Major US cities with their IANA timezone identifiers
		"new york":      "America/New_York",
		"new york city": "America/New_York",
		"nyc":           "America/New_York",
		"manhattan":     "America/New_York",
		"brooklyn":      "America/New_York",
		"queens":        "America/New_York",
		"bronx":         "America/New_York",
		"staten island": "America/New_York",

		"los angeles":   "America/Los_Angeles",
		"la":            "America/Los_Angeles",
		"hollywood":     "America/Los_Angeles",
		"beverly hills": "America/Los_Angeles",
		"santa monica":  "America/Los_Angeles",

		"chicago":     "America/Chicago",
		"houston":     "America/Chicago",
		"dallas":      "America/Chicago",
		"san antonio": "America/Chicago",
		"austin":      "America/Chicago",
		"fort worth":  "America/Chicago",

		"phoenix":    "America/Phoenix",
		"scottsdale": "America/Phoenix",
		"tempe":      "America/Phoenix",
		"mesa":       "America/Phoenix",

		"philadelphia":  "America/New_York",
		"boston":        "America/New_York",
		"washington":    "America/New_York",
		"washington dc": "America/New_York",
		"baltimore":     "America/New_York",
		"atlanta":       "America/New_York",
		"miami":         "America/New_York",
		"orlando":       "America/New_York",
		"tampa":         "America/New_York",
		"jacksonville":  "America/New_York",

		"detroit":      "America/Detroit",
		"columbus":     "America/New_York",
		"indianapolis": "America/Indiana/Indianapolis",
		"milwaukee":    "America/Chicago",
		"nashville":    "America/Chicago",
		"memphis":      "America/Chicago",
		"louisville":   "America/Kentucky/Louisville",

		"san francisco": "America/Los_Angeles",
		"san jose":      "America/Los_Angeles",
		"oakland":       "America/Los_Angeles",
		"sacramento":    "America/Los_Angeles",
		"fresno":        "America/Los_Angeles",
		"san diego":     "America/Los_Angeles",

		"seattle":  "America/Los_Angeles",
		"portland": "America/Los_Angeles",
		"spokane":  "America/Los_Angeles",

		"denver":           "America/Denver",
		"colorado springs": "America/Denver",
		"aurora":           "America/Denver",

		"las vegas": "America/Los_Angeles",
		"reno":      "America/Los_Angeles",

		"salt lake city": "America/Denver",
		"provo":          "America/Denver",

		"albuquerque": "America/Denver",
		"santa fe":    "America/Denver",

		"kansas city":   "America/Chicago",
		"st. louis":     "America/Chicago",
		"oklahoma city": "America/Chicago",
		"tulsa":         "America/Chicago",

		"minneapolis": "America/Chicago",
		"saint paul":  "America/Chicago",
		"duluth":      "America/Chicago",

		"new orleans": "America/Chicago",
		"baton rouge": "America/Chicago",

		"birmingham": "America/Chicago",
		"mobile":     "America/Chicago",

		"charlotte":  "America/New_York",
		"raleigh":    "America/New_York",
		"greensboro": "America/New_York",

		"virginia beach": "America/New_York",
		"norfolk":        "America/New_York",
		"richmond":       "America/New_York",

		"charleston": "America/New_York",
		"columbia":   "America/New_York",

		"tallahassee":    "America/New_York",
		"st. petersburg": "America/New_York",

		"savannah": "America/New_York",
		"augusta":  "America/New_York",

		"buffalo":   "America/New_York",
		"rochester": "America/New_York",
		"syracuse":  "America/New_York",
		"albany":    "America/New_York",

		"pittsburgh": "America/New_York",
		"allentown":  "America/New_York",
		"erie":       "America/New_York",

		"cleveland":  "America/New_York",
		"cincinnati": "America/New_York",
		"toledo":     "America/New_York",
		"akron":      "America/New_York",
		"dayton":     "America/New_York",

		"grand rapids": "America/Detroit",
		"lansing":      "America/Detroit",
		"flint":        "America/Detroit",

		"green bay": "America/Chicago",
		"madison":   "America/Chicago",

		"des moines":   "America/Chicago",
		"cedar rapids": "America/Chicago",

		"omaha":   "America/Chicago",
		"lincoln": "America/Chicago",

		"wichita": "America/Chicago",
		"topeka":  "America/Chicago",

		"little rock":  "America/Chicago",
		"fayetteville": "America/Chicago",

		"jackson":  "America/Chicago",
		"gulfport": "America/Chicago",

		"knoxville":   "America/New_York",
		"chattanooga": "America/New_York",

		"lexington":     "America/New_York",
		"bowling green": "America/Chicago",

		"charleston wv": "America/New_York", // West Virginia
		"huntington":    "America/New_York",

		"richmond va": "America/New_York", // Virginia
		"chesapeake":  "America/New_York",

		"wilmington": "America/New_York",
		"dover":      "America/New_York",

		"bridgeport": "America/New_York",
		"hartford":   "America/New_York",
		"new haven":  "America/New_York",

		"providence": "America/New_York",
		"newport":    "America/New_York",

		"manchester nh": "America/New_York", // New Hampshire
		"nashua":        "America/New_York",

		"burlington": "America/New_York",
		"montpelier": "America/New_York",

		"portland me": "America/New_York", // Maine
		"bangor":      "America/New_York",

		"anchorage": "America/Anchorage",
		"fairbanks": "America/Anchorage",
		"juneau":    "America/Juneau",

		"honolulu": "Pacific/Honolulu",
		"hilo":     "Pacific/Honolulu",
		"kailua":   "Pacific/Honolulu",
	}
}

// normalizeCity normalizes city names for timezone lookup
func (h *Handler) normalizeCity(city string) string {
	// Remove country suffix if present
	city = strings.TrimSuffix(city, ", US")
	city = strings.TrimSuffix(city, ", USA")
	city = strings.TrimSuffix(city, ", United States")

	// Handle common variations
	lowerCity := strings.ToLower(city)

	// Map common variations to names we use in our timezone map
	cityMap := map[string]string{
		"new york city": "new york",
		"nyc":          "new york",
		"la":           "los angeles",
		"sf":           "san francisco",
		"dc":           "washington dc",
		"philly":       "philadelphia",
	}

	if normalized, ok := cityMap[lowerCity]; ok {
		return normalized
	}

	// For other cities, just clean up and return lowercase
	city = strings.TrimSpace(city)

	// If it ends with "City", handle special cases
	if strings.HasSuffix(lowerCity, " city") {
		// Keep "City" for certain cities in our map
		if lowerCity == "new york city" || lowerCity == "kansas city" ||
		   lowerCity == "oklahoma city" || lowerCity == "salt lake city" {
			return lowerCity
		}
		// Remove "City" for others
		return strings.TrimSuffix(lowerCity, " city")
	}

	return strings.ToLower(city)
}
