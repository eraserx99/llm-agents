// Package client provides MCP client functionality using official MCP Go SDK
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/transport"
	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// MCPClient represents an MCP client using official SDK with custom transport
type MCPClient struct {
	client    *mcp.Client
	session   *mcp.ClientSession
	transport *transport.HTTPSSETransport
	serverURL string
	useTLS    bool
	tlsConfig *config.TLSConfig
}

// NewMCPClient creates a new MCP client using official SDK
func NewMCPClient(serverURL string, timeout time.Duration) (*MCPClient, error) {
	// Create MCP client using official SDK
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "llm-agents-client",
		Version: "v1.0.0",
	}, nil)

	// Create custom HTTP/SSE transport
	mcpTransport := transport.NewClientTransport(serverURL, nil)

	mcpClient := &MCPClient{
		client:    client,
		transport: mcpTransport,
		serverURL: serverURL,
		useTLS:    false,
	}

	utils.Info("MCP client created for %s", serverURL)
	return mcpClient, nil
}

// NewTLSMCPClient creates a new MCP client with TLS support using official SDK
func NewTLSMCPClient(serverURL string, timeout time.Duration, tlsConfig *config.TLSConfig) (*MCPClient, error) {
	if tlsConfig == nil {
		return nil, fmt.Errorf("TLS configuration is required for TLS client")
	}

	// Validate TLS configuration
	if err := tlsConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid TLS configuration: %w", err)
	}

	// Create MCP client using official SDK
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "llm-agents-client-tls",
		Version: "v1.0.0",
	}, nil)

	// Create custom HTTP/SSE transport with TLS
	mcpTransport := transport.NewClientTransport(serverURL, tlsConfig)

	mcpClient := &MCPClient{
		client:    client,
		transport: mcpTransport,
		serverURL: serverURL,
		useTLS:    true,
		tlsConfig: tlsConfig,
	}

	utils.Info("TLS MCP client created for %s with mTLS enabled", serverURL)
	return mcpClient, nil
}

// Initialize initializes the MCP client connection
func (c *MCPClient) Initialize(ctx context.Context) error {
	utils.Info("Initializing MCP client connection to %s", c.serverURL)

	// Connect to server using custom transport
	session, err := c.client.Connect(ctx, c.transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	c.session = session
	utils.Info("MCP client connected successfully")
	return nil
}

// ensureConnected ensures the client is connected
func (c *MCPClient) ensureConnected(ctx context.Context) error {
	if c.session == nil {
		return c.Initialize(ctx)
	}
	return nil
}

// CallWeather makes a call to the weather MCP server using official SDK
func (c *MCPClient) CallWeather(ctx context.Context, city string) (*models.TemperatureData, error) {
	if err := c.ensureConnected(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	utils.Info("Calling weather MCP server: getTemperature for city %s", city)

	// Call tool using official SDK with correct parameter structure
	toolParams := &mcp.CallToolParams{
		Name: "getTemperature",
		Arguments: map[string]any{
			"city": city,
		},
	}

	toolResult, err := c.session.CallTool(ctx, toolParams)
	if err != nil {
		return nil, fmt.Errorf("weather call failed: %w", err)
	}

	utils.Debug("Weather MCP response: %+v", toolResult)

	// Extract temperature data from result
	if len(toolResult.Content) == 0 {
		return nil, fmt.Errorf("no content in weather response")
	}

	// Parse the text content to extract temperature info
	textContent := ""
	for _, content := range toolResult.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			textContent = tc.Text
			break
		}
	}

	if textContent == "" {
		return nil, fmt.Errorf("no text content in weather response")
	}

	// Parse the response text: "Weather in {city}: {temp}°C, {description}"
	// Example: "Weather in Boston: 23.5°C, Sunny"
	temperature, description, err := parseWeatherResponse(textContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse weather response: %w", err)
	}

	return &models.TemperatureData{
		City:        city,
		Temperature: temperature,
		Unit:        "°C",
		Description: description,
		Timestamp:   time.Now(),
		Source:      "weather-mcp-streaming",
	}, nil
}

// CallDateTime makes a call to the datetime MCP server using official SDK
func (c *MCPClient) CallDateTime(ctx context.Context, city string) (*models.DateTimeData, error) {
	if err := c.ensureConnected(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	utils.Info("Calling datetime MCP server: getDateTime for city %s", city)

	// Call tool using official SDK
	toolParams := &mcp.CallToolParams{
		Name: "getDateTime",
		Arguments: map[string]any{
			"city": city,
		},
	}

	toolResult, err := c.session.CallTool(ctx, toolParams)
	if err != nil {
		return nil, fmt.Errorf("datetime call failed: %w", err)
	}

	utils.Debug("DateTime MCP response: %+v", toolResult)

	// Extract datetime data from result
	if len(toolResult.Content) == 0 {
		return nil, fmt.Errorf("no content in datetime response")
	}

	// Parse the text content to extract datetime info
	textContent := ""
	for _, content := range toolResult.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			textContent = tc.Text
			break
		}
	}

	if textContent == "" {
		return nil, fmt.Errorf("no text content in datetime response")
	}

	// Parse the response text
	localTimeStr, timezone, utcOffset, err := parseDateTimeResponse(textContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime response: %w", err)
	}

	// Parse the local time string
	localTime, err := time.Parse("2006-01-02 15:04:05", localTimeStr)
	if err != nil {
		// If parsing fails, use current time
		localTime = time.Now()
		utils.Warn("Failed to parse datetime '%s', using current time: %v", localTimeStr, err)
	}

	return &models.DateTimeData{
		City:      city,
		DateTime:  localTime,
		Timezone:  timezone,
		UTCOffset: utcOffset,
		Timestamp: time.Now(),
	}, nil
}

// CallEcho makes a call to the echo MCP server using official SDK
func (c *MCPClient) CallEcho(ctx context.Context, text string) (*models.EchoData, error) {
	if err := c.ensureConnected(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	utils.Info("Calling echo MCP server: echo for text %s", text)

	// Call tool using official SDK
	toolParams := &mcp.CallToolParams{
		Name: "echo",
		Arguments: map[string]any{
			"text": text,
		},
	}

	toolResult, err := c.session.CallTool(ctx, toolParams)
	if err != nil {
		return nil, fmt.Errorf("echo call failed: %w", err)
	}

	utils.Debug("Echo MCP response: %+v", toolResult)

	// Extract echo data from result
	if len(toolResult.Content) == 0 {
		return nil, fmt.Errorf("no content in echo response")
	}

	// Parse the text content
	textContent := ""
	for _, content := range toolResult.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			textContent = tc.Text
			break
		}
	}

	if textContent == "" {
		return nil, fmt.Errorf("no text content in echo response")
	}

	return &models.EchoData{
		OriginalText: text,
		EchoText:     textContent,
		Timestamp:    time.Now(),
	}, nil
}

// Close closes the MCP client connection
func (c *MCPClient) Close() error {
	if c.session != nil {
		utils.Debug("Closing MCP client session")
		c.session.Close()
		c.session = nil
	}
	utils.Debug("MCP client closed")
	return nil
}

// IsSecure returns true if the client uses TLS
func (c *MCPClient) IsSecure() bool {
	return c.useTLS
}

// GetServerURL returns the server URL
func (c *MCPClient) GetServerURL() string {
	return c.serverURL
}

// TestConnection tests the connection to the MCP server
func (c *MCPClient) TestConnection(ctx context.Context) error {
	if err := c.ensureConnected(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	// Test connection by listing tools
	toolsResult, err := c.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	utils.Debug("Connection test successful, found %d tools", len(toolsResult.Tools))
	return nil
}

// ListTools lists available tools from the MCP server
func (c *MCPClient) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	if err := c.ensureConnected(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	toolsResult, err := c.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return toolsResult.Tools, nil
}

// parseWeatherResponse parses weather text response
// Expected format: "Weather in {city}: {temp}°C, {description}"
func parseWeatherResponse(text string) (float64, string, error) {
	// Find the temperature value
	// Look for pattern: {number}°C
	tempStart := -1
	tempEnd := -1

	for i := 0; i < len(text)-2; i++ {
		if text[i:i+2] == "°C" {
			tempEnd = i
			// Find start of number (walk backwards)
			j := i - 1
			for j >= 0 && (text[j] >= '0' && text[j] <= '9' || text[j] == '.') {
				j--
			}
			tempStart = j + 1
			break
		}
	}

	if tempStart == -1 || tempEnd == -1 {
		return 0, "", fmt.Errorf("temperature not found in response: %s", text)
	}

	tempStr := text[tempStart:tempEnd]
	temperature := 0.0
	if _, err := fmt.Sscanf(tempStr, "%f", &temperature); err != nil {
		return 0, "", fmt.Errorf("failed to parse temperature '%s': %w", tempStr, err)
	}

	// Extract description (everything after "°C, ")
	description := ""
	descStart := tempEnd + 4 // Skip "°C, "
	if descStart < len(text) {
		description = text[descStart:]
	}

	return temperature, description, nil
}

// parseDateTimeResponse parses datetime text response
// Expected format: "Time in {city}: {time} ({timezone}, UTC{offset})"
func parseDateTimeResponse(text string) (string, string, string, error) {
	// Simple parsing for now - extract components from known format
	// Example: "Time in New York: 2025-10-02 14:30:00 (America/New_York, UTC-05:00)"

	// Find the colon after city
	colonIdx := -1
	for i := 0; i < len(text); i++ {
		if text[i] == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx == -1 {
		return "", "", "", fmt.Errorf("invalid datetime response format: %s", text)
	}

	// Extract everything after the colon
	remainder := text[colonIdx+2:] // Skip ": "

	// Find the opening parenthesis
	parenIdx := -1
	for i := 0; i < len(remainder); i++ {
		if remainder[i] == '(' {
			parenIdx = i
			break
		}
	}

	if parenIdx == -1 {
		// No timezone info, just return the time
		return remainder, "Unknown", "+00:00", nil
	}

	localTime := remainder[:parenIdx-1] // Remove space before paren

	// Extract timezone and offset from parentheses
	tzInfo := remainder[parenIdx+1 : len(remainder)-1] // Remove ( and )

	// Split by comma
	parts := []string{}
	current := ""
	for _, ch := range tzInfo {
		if ch == ',' {
			parts = append(parts, current)
			current = ""
		} else if ch != ' ' || len(current) > 0 {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	timezone := "Unknown"
	utcOffset := "+00:00"

	if len(parts) >= 1 {
		timezone = parts[0]
	}
	if len(parts) >= 2 {
		utcOffset = parts[1]
		// Remove "UTC" prefix if present
		if len(utcOffset) > 3 && utcOffset[:3] == "UTC" {
			utcOffset = utcOffset[3:]
		}
	}

	return localTime, timezone, utcOffset, nil
}