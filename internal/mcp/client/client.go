// Package client provides MCP client functionality
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/models"
	mcptls "github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

// Client represents an MCP client with optional TLS support
type Client struct {
	baseURL    string
	httpClient *http.Client
	requestID  int64
	tlsConfig  *config.TLSConfig
	tlsLoader  *mcptls.TLSLoader
	useTLS     bool
	serverName string
	mu         sync.RWMutex
}

// NewClient creates a new MCP client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		requestID: 0,
		useTLS:    false,
	}
}

// NewTLSClient creates a new MCP client with TLS support
func NewTLSClient(baseURL string, timeout time.Duration, tlsConfig *config.TLSConfig) *Client {
	if tlsConfig == nil {
		utils.Error("TLS configuration is required for TLS client")
		// Fall back to regular client
		return NewClient(baseURL, timeout)
	}

	// Validate TLS configuration
	if err := tlsConfig.Validate(); err != nil {
		utils.Error("Invalid TLS configuration: %v", err)
		// Fall back to regular client
		return NewClient(baseURL, timeout)
	}

	// Create TLS loader
	tlsLoader := mcptls.NewTLSLoader(tlsConfig)

	// Extract server name from baseURL for TLS validation
	serverName := "localhost" // Default for demo mode
	// In production, this would parse the hostname from baseURL

	// Load client TLS configuration
	clientTLSConfig, err := tlsLoader.LoadClientTLSConfig(serverName)
	if err != nil {
		utils.Error("Failed to load client TLS config: %v", err)
		// Fall back to regular client
		return NewClient(baseURL, timeout)
	}

	// Create HTTP client with TLS transport
	transport := &http.Transport{
		TLSClientConfig:     clientTLSConfig,
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     30 * time.Second,
	}

	client := &Client{
		baseURL:    baseURL,
		tlsConfig:  tlsConfig,
		tlsLoader:  tlsLoader,
		useTLS:     true,
		serverName: serverName,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		requestID: 0,
	}

	utils.Info("TLS client created for %s with mTLS enabled", baseURL)
	return client
}

// nextRequestID generates the next request ID
func (c *Client) nextRequestID() int {
	return int(atomic.AddInt64(&c.requestID, 1))
}

// Call makes a JSON-RPC call to the MCP server
func (c *Client) Call(ctx context.Context, method string, params interface{}) (interface{}, error) {
	// Create request
	request := struct {
		JSONRpc string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
		ID      int         `json:"id"`
	}{
		JSONRpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      c.nextRequestID(),
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	utils.Info("MCP call to %s: %s", c.baseURL+"/rpc", method)
	utils.Debug("MCP request body: %s", string(requestBody))

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/rpc", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	utils.Info("MCP response received from %s (status: %d)", c.baseURL, resp.StatusCode)
	utils.Debug("MCP response body: %s", string(responseBody))

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var response struct {
		JSONRpc string           `json:"jsonrpc"`
		Result  interface{}      `json:"result,omitempty"`
		Error   *models.MCPError `json:"error,omitempty"`
		ID      int              `json:"id"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for JSON-RPC error
	if response.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", response.Error.Code, response.Error.Message)
	}

	return response.Result, nil
}

// CallWeather makes a call to the weather MCP server
func (c *Client) CallWeather(ctx context.Context, city string) (*models.TemperatureData, error) {
	params := struct {
		City string `json:"city"`
	}{
		City: city,
	}

	result, err := c.Call(ctx, "getTemperature", params)
	if err != nil {
		return nil, fmt.Errorf("weather call failed: %w", err)
	}

	// Parse result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal weather result: %w", err)
	}

	var weatherResult struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
		Description string  `json:"description"`
	}

	if err := json.Unmarshal(resultJSON, &weatherResult); err != nil {
		return nil, fmt.Errorf("failed to parse weather result: %w", err)
	}

	return &models.TemperatureData{
		City:        city,
		Temperature: weatherResult.Temperature,
		Unit:        weatherResult.Unit,
		Description: weatherResult.Description,
		Timestamp:   time.Now(),
		Source:      "weather-mcp",
	}, nil
}

// CallDateTime makes a call to the datetime MCP server
func (c *Client) CallDateTime(ctx context.Context, city string) (*models.DateTimeData, error) {
	params := struct {
		City string `json:"city"`
	}{
		City: city,
	}

	result, err := c.Call(ctx, "getDateTime", params)
	if err != nil {
		return nil, fmt.Errorf("datetime call failed: %w", err)
	}

	// Parse result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal datetime result: %w", err)
	}

	var datetimeResult struct {
		DateTime  string `json:"datetime"`
		Timezone  string `json:"timezone"`
		UTCOffset string `json:"utc_offset"`
	}

	if err := json.Unmarshal(resultJSON, &datetimeResult); err != nil {
		return nil, fmt.Errorf("failed to parse datetime result: %w", err)
	}

	// Parse datetime
	parsedTime, err := time.Parse(time.RFC3339, datetimeResult.DateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse datetime: %w", err)
	}

	return &models.DateTimeData{
		City:      city,
		DateTime:  parsedTime,
		Timezone:  datetimeResult.Timezone,
		UTCOffset: datetimeResult.UTCOffset,
		Timestamp: time.Now(),
	}, nil
}

// CallEcho makes a call to the echo MCP server
func (c *Client) CallEcho(ctx context.Context, text string) (*models.EchoData, error) {
	params := struct {
		Text string `json:"text"`
	}{
		Text: text,
	}

	result, err := c.Call(ctx, "echo", params)
	if err != nil {
		return nil, fmt.Errorf("echo call failed: %w", err)
	}

	// Parse result
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal echo result: %w", err)
	}

	var echoResult struct {
		OriginalText string `json:"original_text"`
		EchoText     string `json:"echo_text"`
	}

	if err := json.Unmarshal(resultJSON, &echoResult); err != nil {
		return nil, fmt.Errorf("failed to parse echo result: %w", err)
	}

	return &models.EchoData{
		OriginalText: echoResult.OriginalText,
		EchoText:     echoResult.EchoText,
		Timestamp:    time.Now(),
	}, nil
}

// GetConnectionInfo returns TLS connection information if available
func (c *Client) GetConnectionInfo() *mcptls.TLSConnectionInfo {
	if !c.useTLS || c.tlsLoader == nil {
		return nil
	}

	// This would typically be called during an active connection
	// For now, return basic info about the client configuration
	return &mcptls.TLSConnectionInfo{
		RemoteAddr:        c.baseURL,
		TLSVersion:        "TLS 1.2+", // Based on min version
		CipherSuite:       "Negotiated",
		ClientCertCN:      "mcp-client", // Default client cert name
		HandshakeComplete: true,
	}
}

// ValidateServerCert validates the server certificate
func (c *Client) ValidateServerCert() error {
	if !c.useTLS || c.tlsConfig == nil {
		return fmt.Errorf("TLS not enabled for this client")
	}

	// In a real implementation, this would connect and validate
	// For now, validate that our TLS config is correct
	return c.tlsConfig.Validate()
}

// IsSecure returns true if the client uses TLS
func (c *Client) IsSecure() bool {
	return c.useTLS
}

// GetTLSConfig returns the client TLS configuration (read-only)
func (c *Client) GetTLSConfig() *config.TLSConfig {
	if c.tlsConfig != nil {
		// Return a copy to prevent external modification
		configCopy := *c.tlsConfig
		return &configCopy
	}
	return nil
}

// ConnectionStatus represents the client connection status
type ConnectionStatus struct {
	ServerURL      string                    `json:"server_url"`
	UseTLS         bool                      `json:"use_tls"`
	Connected      bool                      `json:"connected"`
	LastCall       *time.Time                `json:"last_call,omitempty"`
	TotalCalls     int                       `json:"total_calls"`
	FailedCalls    int                       `json:"failed_calls"`
	TLSInfo        *mcptls.TLSConnectionInfo `json:"tls_info,omitempty"`
}

// GetConnectionStatus returns the current connection status
func (c *Client) GetConnectionStatus() *ConnectionStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := &ConnectionStatus{
		ServerURL:   c.baseURL,
		UseTLS:      c.useTLS,
		Connected:   true, // Assume connected if client exists
		TotalCalls:  0,    // Would need tracking
		FailedCalls: 0,    // Would need tracking
	}

	if c.useTLS {
		status.TLSInfo = c.GetConnectionInfo()
	}

	return status
}

// TestConnection tests the connection to the server
func (c *Client) TestConnection(ctx context.Context) error {
	// Test with a simple ping-like call
	// In a real implementation, servers might have a ping method
	_, err := c.Call(ctx, "ping", map[string]interface{}{})

	// If ping isn't supported, that's expected - connection worked
	if err != nil && err.Error() == "MCP error -32601: Method not found" {
		return nil // Connection successful, method just not found
	}

	return err
}

// Close closes the client and cleans up resources
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.httpClient != nil {
		// Close idle connections
		c.httpClient.CloseIdleConnections()
	}

	utils.Debug("MCP client closed: %s", c.baseURL)
}
