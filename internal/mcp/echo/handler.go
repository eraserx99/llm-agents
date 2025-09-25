// Package echo provides echo functionality for the MCP server
package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/steve/llm-agents/internal/utils"
)

// Handler implements echo MCP method handling
type Handler struct{}

// NewHandler creates a new echo handler
func NewHandler() *Handler {
	return &Handler{}
}

// Handle handles the echo method
func (h *Handler) Handle(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var request struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	if request.Text == "" {
		return nil, fmt.Errorf("text parameter is required and cannot be empty")
	}

	// Validate text length (max 1000 characters as per contract)
	if len(request.Text) > 1000 {
		return nil, fmt.Errorf("text too long: maximum 1000 characters allowed")
	}

	// Simple echo - return the text exactly as received
	echoText := request.Text

	// Trim only leading/trailing whitespace if any (preserve internal formatting)
	originalText := strings.TrimSpace(request.Text)
	echoText = strings.TrimSpace(echoText)

	result := struct {
		OriginalText string `json:"original_text"`
		EchoText     string `json:"echo_text"`
	}{
		OriginalText: originalText,
		EchoText:     echoText,
	}

	utils.Debug("Echo processed: %q -> %q", originalText, echoText)
	return result, nil
}
