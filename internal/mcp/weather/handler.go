// Package weather provides weather data handling for the MCP server
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/steve/llm-agents/internal/utils"
)

// Handler implements weather MCP method handling
type Handler struct {
	client *http.Client
}

// NewHandler creates a new weather handler
func NewHandler() *Handler {
	return &Handler{
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
	}
}

// Handle handles the getTemperature method
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

	// Get weather data from wttr.in
	temp, description, err := h.getWeatherData(ctx, request.City)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather data: %w", err)
	}

	result := struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
		Description string  `json:"description"`
	}{
		Temperature: temp,
		Unit:        "F",
		Description: description,
	}

	utils.Debug("Weather data retrieved for %s: %.1f°F, %s", request.City, temp, description)
	return result, nil
}

// getWeatherData fetches weather data from NWS API
func (h *Handler) getWeatherData(ctx context.Context, city string) (float64, string, error) {
	// Normalize city name
	normalizedCity := h.normalizeCity(city)

	// Get coordinates for the city
	lat, lon, err := h.getCityCoordinates(normalizedCity)
	if err != nil {
		return 0, "", fmt.Errorf("city not found: %s", city)
	}

	// Step 1: Get the grid point for the coordinates
	gridURL := fmt.Sprintf("https://api.weather.gov/points/%.4f,%.4f", lat, lon)

	req, err := http.NewRequestWithContext(ctx, "GET", gridURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create grid request: %w", err)
	}
	req.Header.Set("User-Agent", "llm-agents/1.0")

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch grid data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, "", fmt.Errorf("grid service returned status %d", resp.StatusCode)
	}

	var gridData struct {
		Properties struct {
			ForecastURL string `json:"forecast"`
		} `json:"properties"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&gridData); err != nil {
		return 0, "", fmt.Errorf("failed to parse grid response: %w", err)
	}

	// Step 2: Get the forecast data
	forecastReq, err := http.NewRequestWithContext(ctx, "GET", gridData.Properties.ForecastURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create forecast request: %w", err)
	}
	forecastReq.Header.Set("User-Agent", "llm-agents/1.0")

	forecastResp, err := h.client.Do(forecastReq)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch forecast: %w", err)
	}
	defer forecastResp.Body.Close()

	if forecastResp.StatusCode != 200 {
		return 0, "", fmt.Errorf("forecast service returned status %d", forecastResp.StatusCode)
	}

	var forecastData struct {
		Properties struct {
			Periods []struct {
				Temperature int    `json:"temperature"`
				ShortForecast string `json:"shortForecast"`
			} `json:"periods"`
		} `json:"properties"`
	}

	if err := json.NewDecoder(forecastResp.Body).Decode(&forecastData); err != nil {
		return 0, "", fmt.Errorf("failed to parse forecast: %w", err)
	}

	if len(forecastData.Properties.Periods) == 0 {
		return 0, "", fmt.Errorf("no forecast data available")
	}

	// Get current period (first one)
	current := forecastData.Properties.Periods[0]

	return float64(current.Temperature), current.ShortForecast, nil
}

// parseWeatherResponse parses the weather response from wttr.in
func (h *Handler) parseWeatherResponse(response string) (float64, string, error) {
	// Response format: "+72°F:Partly cloudy"
	response = strings.TrimSpace(response)

	// Check if response indicates an error (common error patterns)
	if strings.Contains(response, "Unknown location") ||
		strings.Contains(response, "Sorry") ||
		len(response) < 3 {
		return 0, "", fmt.Errorf("city not found or invalid location")
	}

	parts := strings.Split(response, ":")
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("unexpected response format: %s", response)
	}

	tempStr := parts[0]
	description := strings.TrimSpace(parts[1])

	// Extract temperature value (remove + sign and °F)
	tempRegex := regexp.MustCompile(`([+-]?\d+(?:\.\d+)?)°?[CF]?`)
	matches := tempRegex.FindStringSubmatch(tempStr)
	if len(matches) < 2 {
		return 0, "", fmt.Errorf("could not parse temperature from: %s", tempStr)
	}

	temp, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid temperature value: %s", matches[1])
	}

	// Convert to Fahrenheit if needed (wttr.in usually returns Fahrenheit by default for US queries)
	// If the response was in Celsius, convert it
	if strings.Contains(tempStr, "°C") {
		temp = temp*9/5 + 32
	}

	return temp, description, nil
}

// normalizeCity normalizes city names for wttr.in API
func (h *Handler) normalizeCity(city string) string {
	// Remove country suffix if present
	city = strings.TrimSuffix(city, ", US")
	city = strings.TrimSuffix(city, ", USA")
	city = strings.TrimSuffix(city, ", United States")

	// Handle common variations
	lowerCity := strings.ToLower(city)

	// Map common variations to standard names
	cityMap := map[string]string{
		"new york city": "New York",
		"nyc":          "New York",
		"la":           "Los Angeles",
		"sf":           "San Francisco",
		"dc":           "Washington DC",
		"philly":       "Philadelphia",
	}

	if normalized, ok := cityMap[lowerCity]; ok {
		return normalized
	}

	// For other cities, just clean up the input
	city = strings.TrimSpace(city)

	// If it ends with "City", try removing it for most cities
	if strings.HasSuffix(city, " City") {
		cityWithoutSuffix := strings.TrimSuffix(city, " City")
		// But keep it for specific cities that need it
		if lowerCity != "kansas city" && lowerCity != "oklahoma city" && lowerCity != "salt lake city" {
			city = cityWithoutSuffix
		}
	}

	return city
}

// getCityCoordinates returns latitude and longitude for major US cities
func (h *Handler) getCityCoordinates(city string) (float64, float64, error) {
	city = strings.ToLower(strings.TrimSpace(city))

	// Coordinates for major US cities
	coordinates := map[string]struct{ lat, lon float64 }{
		"new york":      {40.7128, -74.0060},
		"los angeles":   {34.0522, -118.2437},
		"chicago":       {41.8781, -87.6298},
		"houston":       {29.7604, -95.3698},
		"phoenix":       {33.4484, -112.0740},
		"philadelphia":  {39.9526, -75.1652},
		"san antonio":   {29.4241, -98.4936},
		"san diego":     {32.7157, -117.1611},
		"dallas":        {32.7767, -96.7970},
		"san jose":      {37.3382, -121.8863},
		"austin":        {30.2672, -97.7431},
		"jacksonville":  {30.3322, -81.6557},
		"fort worth":    {32.7555, -97.3308},
		"columbus":      {39.9612, -82.9988},
		"charlotte":     {35.2271, -80.8431},
		"san francisco": {37.7749, -122.4194},
		"indianapolis":  {39.7684, -86.1581},
		"seattle":       {47.6062, -122.3321},
		"denver":        {39.7392, -104.9903},
		"washington":    {38.9072, -77.0369},
		"washington dc": {38.9072, -77.0369},
		"boston":        {42.3601, -71.0589},
		"el paso":       {31.7619, -106.4850},
		"detroit":       {42.3314, -83.0458},
		"nashville":     {36.1627, -86.7816},
		"portland":      {45.5152, -122.6784},
		"memphis":       {35.1495, -90.0490},
		"oklahoma city": {35.4676, -97.5164},
		"las vegas":     {36.1699, -115.1398},
		"louisville":    {38.2527, -85.7585},
		"baltimore":     {39.2904, -76.6122},
		"milwaukee":     {43.0389, -87.9065},
		"albuquerque":   {35.0853, -106.6056},
		"tucson":        {32.2226, -110.9747},
		"fresno":        {36.7378, -119.7871},
		"mesa":          {33.4152, -111.8315},
		"sacramento":    {38.5816, -121.4944},
		"atlanta":       {33.7490, -84.3880},
		"kansas city":   {39.0997, -94.5786},
		"colorado springs": {38.8339, -104.8214},
		"miami":         {25.7617, -80.1918},
		"raleigh":       {35.7796, -78.6382},
		"omaha":         {41.2565, -95.9345},
		"long beach":    {33.7701, -118.1937},
		"virginia beach": {36.8529, -75.9780},
		"oakland":       {37.8044, -122.2712},
		"minneapolis":   {44.9778, -93.2650},
		"tulsa":         {36.1540, -95.9928},
		"arlington":     {32.7357, -97.1081},
		"new orleans":   {29.9511, -90.0715},
		"wichita":       {37.6872, -97.3301},
	}

	if coords, ok := coordinates[city]; ok {
		return coords.lat, coords.lon, nil
	}

	return 0, 0, fmt.Errorf("coordinates not found for city: %s", city)
}
