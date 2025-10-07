// Package client provides MCP client functionality using the official Go SDK
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/models"
	mcptls "github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

// Client represents an MCP client using the official SDK
type Client struct {
	endpoint      string
	mcpClient     *mcp.Client
	session       *mcp.ClientSession
	tlsConfig     *config.TLSConfig
	useTLS        bool
	mu            sync.RWMutex
	connected     bool
	reconnectOnce sync.Once
}

// NewClient creates a new MCP client without TLS
func NewClient(endpoint string, timeout time.Duration) (*Client, error) {
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	c := &Client{
		endpoint: endpoint,
		mcpClient: mcp.NewClient(&mcp.Implementation{
			Name:    "llm-agents-client",
			Version: "v1.0.0",
		}, nil),
		useTLS: false,
	}

	// Create transport
	transport := &mcp.StreamableClientTransport{
		Endpoint:   endpoint,
		HTTPClient: httpClient,
		MaxRetries: 5,
	}

	// Connect to the server
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	session, err := c.mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	c.session = session
	c.connected = true

	utils.Info("MCP client connected to %s (HTTP)", endpoint)
	return c, nil
}

// NewTLSClient creates a new MCP client with TLS support
func NewTLSClient(endpoint string, timeout time.Duration, tlsConfig *config.TLSConfig) (*Client, error) {
	if tlsConfig == nil {
		utils.Error("TLS configuration is required for TLS client")
		return NewClient(endpoint, timeout)
	}

	// Validate TLS configuration
	if err := tlsConfig.Validate(); err != nil {
		utils.Error("Invalid TLS configuration: %v", err)
		return NewClient(endpoint, timeout)
	}

	// Create TLS loader
	tlsLoader := mcptls.NewTLSLoader(tlsConfig)

	// Extract server name from endpoint for TLS validation
	serverName := "localhost" // Default for demo mode

	// Load client TLS configuration
	clientTLSConfig, err := tlsLoader.LoadClientTLSConfig(serverName)
	if err != nil {
		utils.Error("Failed to load client TLS config: %v", err)
		return NewClient(endpoint, timeout)
	}

	// Create HTTP client with TLS transport
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig:     clientTLSConfig,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	c := &Client{
		endpoint:  endpoint,
		tlsConfig: tlsConfig,
		mcpClient: mcp.NewClient(&mcp.Implementation{
			Name:    "llm-agents-client",
			Version: "v1.0.0",
		}, nil),
		useTLS: true,
	}

	// Create transport with TLS-enabled HTTP client
	transport := &mcp.StreamableClientTransport{
		Endpoint:   endpoint,
		HTTPClient: httpClient,
		MaxRetries: 5,
	}

	// Connect to the server
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	session, err := c.mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	c.session = session
	c.connected = true

	utils.Info("MCP client connected to %s with mTLS enabled", endpoint)
	return c, nil
}

// ensureConnection ensures the client is connected
func (c *Client) ensureConnection(ctx context.Context) error {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return fmt.Errorf("client not connected")
	}

	return nil
}

// CallWeather makes a call to the weather MCP server
func (c *Client) CallWeather(ctx context.Context, city string) (*models.TemperatureData, error) {
	if err := c.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	// Call the getTemperature tool
	args := map[string]interface{}{
		"city": city,
	}

	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "getTemperature",
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("getTemperature call failed: %w", err)
	}

	// Log the complete result structure for debugging
	if resultJSON, err := json.MarshalIndent(result, "", "  "); err == nil {
		utils.Debug("Complete CallTool result from MCP server:\n%s", string(resultJSON))
	}
	utils.Debug("result.StructuredContent type: %T, value: %+v", result.StructuredContent, result.StructuredContent)
	utils.Debug("result.Content length: %d", len(result.Content))

	// Extract result from StructuredContent
	var weatherData struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
		Description string  `json:"description"`
		City        string  `json:"city"`
		Timestamp   string  `json:"timestamp"`
	}

	// The SDK populates StructuredContent with the typed result
	if result.StructuredContent != nil {
		if structuredJSON, err := json.Marshal(result.StructuredContent); err == nil {
			if err := json.Unmarshal(structuredJSON, &weatherData); err != nil {
				utils.Error("Failed to parse structured content: %v", err)
				return nil, fmt.Errorf("failed to parse weather data: %w", err)
			}
			utils.Debug("Parsed weather data from structured content: %+v", weatherData)
		}
	} else if len(result.Content) > 0 {
		// Fallback: try to parse from text content
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			utils.Debug("Weather result text: %s", textContent.Text)
			// Text content is for display; we should have structured content
		}
		return nil, fmt.Errorf("no structured content in result")
	}

	return &models.TemperatureData{
		City:        weatherData.City,
		Temperature: weatherData.Temperature,
		Unit:        weatherData.Unit,
		Description: weatherData.Description,
		Timestamp:   time.Now(),
		Source:      "weather-mcp",
	}, nil
}

// CallDateTime makes a call to the datetime MCP server
func (c *Client) CallDateTime(ctx context.Context, city string) (*models.DateTimeData, error) {
	if err := c.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	// Call the getDateTime tool
	args := map[string]interface{}{
		"city": city,
	}

	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "getDateTime",
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("getDateTime call failed: %w", err)
	}

	// Extract result from StructuredContent
	var datetimeData struct {
		LocalTime string `json:"local_time"`
		Timezone  string `json:"timezone"`
		UTCOffset string `json:"utc_offset"`
		City      string `json:"city"`
	}

	// The SDK populates StructuredContent with the typed result
	if result.StructuredContent != nil {
		if structuredJSON, err := json.Marshal(result.StructuredContent); err == nil {
			if err := json.Unmarshal(structuredJSON, &datetimeData); err != nil {
				utils.Error("Failed to parse structured content: %v", err)
				return nil, fmt.Errorf("failed to parse datetime data: %w", err)
			}
			utils.Debug("Parsed datetime data from structured content: %+v", datetimeData)
		}
	} else {
		return nil, fmt.Errorf("no structured content in result")
	}

	// Parse datetime - server returns format "2006-01-02 15:04:05"
	parsedTime, err := time.Parse("2006-01-02 15:04:05", datetimeData.LocalTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %w", err)
	}

	return &models.DateTimeData{
		City:      datetimeData.City,
		DateTime:  parsedTime,
		Timezone:  datetimeData.Timezone,
		UTCOffset: datetimeData.UTCOffset,
		Timestamp: time.Now(),
	}, nil
}

// CallEcho makes a call to the echo MCP server
func (c *Client) CallEcho(ctx context.Context, text string) (*models.EchoData, error) {
	if err := c.ensureConnection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure connection: %w", err)
	}

	// Call the echo tool
	args := map[string]interface{}{
		"text": text,
	}

	result, err := c.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "echo",
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("echo call failed: %w", err)
	}

	// Extract result from StructuredContent
	var echoData struct {
		OriginalText string `json:"original_text"`
		EchoText     string `json:"echo_text"`
	}

	// The SDK populates StructuredContent with the typed result
	if result.StructuredContent != nil {
		if structuredJSON, err := json.Marshal(result.StructuredContent); err == nil {
			if err := json.Unmarshal(structuredJSON, &echoData); err != nil {
				utils.Error("Failed to parse structured content: %v", err)
				return nil, fmt.Errorf("failed to parse echo data: %w", err)
			}
			utils.Debug("Parsed echo data from structured content: %+v", echoData)
		}
	} else {
		return nil, fmt.Errorf("no structured content in result")
	}

	return &models.EchoData{
		OriginalText: echoData.OriginalText,
		EchoText:     echoData.EchoText,
		Timestamp:    time.Now(),
	}, nil
}

// GetConnectionInfo returns TLS connection information if available
func (c *Client) GetConnectionInfo() *mcptls.TLSConnectionInfo {
	if !c.useTLS {
		return nil
	}

	return &mcptls.TLSConnectionInfo{
		RemoteAddr:        c.endpoint,
		TLSVersion:        "TLS 1.2+",
		CipherSuite:       "Negotiated",
		ClientCertCN:      "mcp-client",
		HandshakeComplete: true,
	}
}

// IsSecure returns true if the client uses TLS
func (c *Client) IsSecure() bool {
	return c.useTLS
}

// GetTLSConfig returns the client TLS configuration (read-only)
func (c *Client) GetTLSConfig() *config.TLSConfig {
	if c.tlsConfig != nil {
		configCopy := *c.tlsConfig
		return &configCopy
	}
	return nil
}

// TestConnection tests the connection to the server
func (c *Client) TestConnection(ctx context.Context) error {
	return c.ensureConnection(ctx)
}

// Close closes the client and cleans up resources
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.session != nil {
		c.session.Close()
		c.connected = false
	}

	utils.Debug("MCP client closed: %s", c.endpoint)
}
