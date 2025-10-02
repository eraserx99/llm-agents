// Package datetime provides the datetime sub-agent implementation
package datetime

import (
	"context"
	"fmt"
	"time"

	"github.com/steve/llm-agents/internal/agents/client"
	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// Agent implements the datetime sub-agent
type Agent struct {
	mcpClient *client.MCPClient
}

// NewAgent creates a new datetime agent
func NewAgent(datetimeServerURL string, timeout time.Duration) *Agent {
	mcpClient, err := client.NewMCPClient(datetimeServerURL, timeout)
	if err != nil {
		utils.Error("Failed to create MCP client: %v", err)
		return nil
	}
	return &Agent{
		mcpClient: mcpClient,
	}
}

// NewTLSAgent creates a new datetime agent with TLS support
func NewTLSAgent(datetimeServerURL string, timeout time.Duration, tlsConfig *config.TLSConfig) *Agent {
	mcpClient, err := client.NewTLSMCPClient(datetimeServerURL, timeout, tlsConfig)
	if err != nil {
		utils.Error("Failed to create TLS MCP client: %v", err)
		// Fall back to regular client
		return NewAgent(datetimeServerURL, timeout)
	}
	return &Agent{
		mcpClient: mcpClient,
	}
}

// ProcessRequest processes a datetime request
func (a *Agent) ProcessRequest(ctx context.Context, request models.AgentRequest) (*models.AgentResponse, error) {
	utils.Debug("DateTime agent processing request: %+v", request)

	// Validate request
	if request.AgentType != models.AgentTypeDateTime {
		return nil, fmt.Errorf("invalid agent type: expected %s, got %s",
			models.AgentTypeDateTime, request.AgentType)
	}

	if request.City == "" {
		return nil, fmt.Errorf("city parameter is required for datetime requests")
	}

	// Create context with timeout
	reqCtx := ctx
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}

	// Call datetime MCP server
	dateTimeData, err := a.mcpClient.CallDateTime(reqCtx, request.City)
	if err != nil {
		utils.Error("DateTime agent failed to get datetime data for %s: %v", request.City, err)
		return &models.AgentResponse{
			RequestID: request.RequestID,
			TaskID:    request.TaskID,
			Success:   false,
			Error:     fmt.Sprintf("Failed to retrieve datetime data: %v", err),
		}, nil
	}

	utils.Info("DateTime agent retrieved data for %s: %s (%s)",
		dateTimeData.City, dateTimeData.DateTime.Format(time.RFC3339), dateTimeData.Timezone)

	// Create successful response
	response := &models.AgentResponse{
		RequestID: request.RequestID,
		TaskID:    request.TaskID,
		Success:   true,
		Data:      dateTimeData,
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
