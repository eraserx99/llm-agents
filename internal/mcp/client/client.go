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

	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// Client represents an MCP client
type Client struct {
	baseURL    string
	httpClient *http.Client
	requestID  int64
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
	}
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

// Close closes the client and cleans up resources
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.httpClient != nil {
		// Close idle connections
		c.httpClient.CloseIdleConnections()
	}
}
