// Package temperature provides the temperature sub-agent implementation
package temperature

import (
	"context"
	"fmt"
	"time"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/client"
	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// Agent implements the temperature sub-agent
type Agent struct {
	mcpClient *client.Client
}

// NewAgent creates a new temperature agent
func NewAgent(weatherServerURL string, timeout time.Duration) *Agent {
	return &Agent{
		mcpClient: client.NewClient(weatherServerURL, timeout),
	}
}

// NewTLSAgent creates a new temperature agent with TLS support
func NewTLSAgent(weatherServerURL string, timeout time.Duration, tlsConfig *config.TLSConfig) *Agent {
	return &Agent{
		mcpClient: client.NewTLSClient(weatherServerURL, timeout, tlsConfig),
	}
}

// ProcessRequest processes a temperature request
func (a *Agent) ProcessRequest(ctx context.Context, request models.AgentRequest) (*models.AgentResponse, error) {
	utils.Debug("Temperature agent processing request: %+v", request)

	// Validate request
	if request.AgentType != models.AgentTypeTemperature {
		return nil, fmt.Errorf("invalid agent type: expected %s, got %s",
			models.AgentTypeTemperature, request.AgentType)
	}

	if request.City == "" {
		return nil, fmt.Errorf("city parameter is required for temperature requests")
	}

	// Create context with timeout
	reqCtx := ctx
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}

	// Call weather MCP server
	tempData, err := a.mcpClient.CallWeather(reqCtx, request.City)
	if err != nil {
		utils.Error("Temperature agent failed to get weather data for %s: %v", request.City, err)
		return &models.AgentResponse{
			RequestID: request.RequestID,
			TaskID:    request.TaskID,
			Success:   false,
			Error:     fmt.Sprintf("Failed to retrieve temperature data: %v", err),
		}, nil
	}

	utils.Info("Temperature agent retrieved data for %s: %.1f°%s",
		tempData.City, tempData.Temperature, tempData.Unit)

	// Create successful response
	response := &models.AgentResponse{
		RequestID: request.RequestID,
		TaskID:    request.TaskID,
		Success:   true,
		Data:      tempData,
	}

	return response, nil
}

// Close closes the agent and cleans up resources
func (a *Agent) Close() {
	if a.mcpClient != nil {
		a.mcpClient.Close()
	}
}

// Validate validates the agent configuration
func (a *Agent) Validate() error {
	if a.mcpClient == nil {
		return fmt.Errorf("MCP client is not initialized")
	}
	return nil
}
