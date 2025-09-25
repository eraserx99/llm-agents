// Package echo provides the echo sub-agent implementation
package echo

import (
	"context"
	"fmt"
	"time"

	"github.com/steve/llm-agents/internal/mcp/client"
	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// Agent implements the echo sub-agent
type Agent struct {
	mcpClient *client.Client
}

// NewAgent creates a new echo agent
func NewAgent(echoServerURL string, timeout time.Duration) *Agent {
	return &Agent{
		mcpClient: client.NewClient(echoServerURL, timeout),
	}
}

// ProcessRequest processes an echo request
func (a *Agent) ProcessRequest(ctx context.Context, request models.AgentRequest) (*models.AgentResponse, error) {
	utils.Debug("Echo agent processing request: %+v", request)

	// Validate request
	if request.AgentType != models.AgentTypeEcho {
		return nil, fmt.Errorf("invalid agent type: expected %s, got %s",
			models.AgentTypeEcho, request.AgentType)
	}

	if request.EchoText == "" {
		return nil, fmt.Errorf("echo text parameter is required for echo requests")
	}

	// Create context with timeout
	reqCtx := ctx
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}

	// Call echo MCP server
	echoData, err := a.mcpClient.CallEcho(reqCtx, request.EchoText)
	if err != nil {
		utils.Error("Echo agent failed to process text: %v", err)
		return &models.AgentResponse{
			RequestID: request.RequestID,
			TaskID:    request.TaskID,
			Success:   false,
			Error:     fmt.Sprintf("Failed to process echo request: %v", err),
		}, nil
	}

	utils.Info("Echo agent processed text: %s -> %s",
		echoData.OriginalText, echoData.EchoText)

	// Create successful response
	response := &models.AgentResponse{
		RequestID: request.RequestID,
		TaskID:    request.TaskID,
		Success:   true,
		Data:      echoData,
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
