// Package coordinator provides the main coordinator agent implementation
package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// LLMClient handles communication with Claude 3.5 Sonnet via OpenRouter
type LLMClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewLLMClient creates a new LLM client for OpenRouter
func NewLLMClient(apiKey string) *LLMClient {
	return &LLMClient{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// OpenRouterRequest represents the request format for OpenRouter API
type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenRouterResponse represents the response format from OpenRouter API
type OpenRouterResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// GenerateOrchestrationPlan uses Claude 3.5 Sonnet to analyze a query and generate an orchestration plan
func (c *LLMClient) GenerateOrchestrationPlan(ctx context.Context, query models.Query) (*models.OrchestrationPlan, error) {
	utils.Debug("Generating orchestration plan for query: %s", query.Text)

	// Create the prompt for Claude
	prompt := c.buildOrchestrationPrompt(query)

	// Call Claude 3.5 Sonnet
	response, err := c.callClaude(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude: %w", err)
	}

	// Parse the response into an orchestration plan
	plan, err := c.parseOrchestrationResponse(response, query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse orchestration response: %w", err)
	}

	utils.Info("Generated orchestration plan with %d tasks, strategy: %s",
		len(plan.Tasks), plan.Strategy)

	return plan, nil
}

// buildOrchestrationPrompt creates the prompt for Claude to analyze the query
func (c *LLMClient) buildOrchestrationPrompt(query models.Query) string {
	return fmt.Sprintf(`You are an intelligent agent coordinator. Analyze the following user query and determine which sub-agents should be invoked and how they should execute.

Available sub-agents:
1. Temperature Agent: Retrieves current temperature and weather information for US cities
2. DateTime Agent: Retrieves current date and time information for US cities with timezone handling
3. Echo Agent: Simple text echo functionality (only use when explicitly requested to echo text)

Query: "%s"
City: "%s"

IMPORTANT RULES:
- Only invoke Echo Agent when the user explicitly asks to echo text (e.g., "echo hello world", "repeat this text")
- For Echo Agent requests, extract the text to echo from the query (e.g., for "echo hello world", use "hello world" as the text parameter)
- For weather/temperature and datetime queries, do NOT invoke the Echo Agent
- Use parallel execution when multiple data types are requested (e.g., both weather and time)
- Use sequential execution when one result depends on another or for single requests
- Provide clear reasoning for your decisions

Respond with a JSON object in this exact format:
{
  "strategy": "parallel" | "sequential",
  "tasks": [
    {
      "agent_type": "temperature" | "datetime" | "echo",
      "priority": 1,
      "dependencies": [],
      "parameters": {
        "city": "city_name",
        "text": "text_to_echo"
      }
    }
  ],
  "reasoning": "Explanation of why this orchestration plan was chosen"
}`, query.Text, query.City)
}

// callClaude makes an API call to Claude 3.5 Sonnet via OpenRouter
func (c *LLMClient) callClaude(ctx context.Context, prompt string) (string, error) {
	request := OpenRouterRequest{
		Model: "anthropic/claude-3.5-sonnet",
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/steve/llm-agents")
	httpReq.Header.Set("X-Title", "LLM Multi-Agent System")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	utils.Debug("OpenRouter response status: %d", resp.StatusCode)
	utils.Debug("OpenRouter response body: %s", string(responseBody))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(responseBody))
	}

	var response OpenRouterResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// parseOrchestrationResponse parses Claude's response into an OrchestrationPlan
func (c *LLMClient) parseOrchestrationResponse(response string, query models.Query) (*models.OrchestrationPlan, error) {
	// Try to extract JSON from the response (Claude might include explanation text)
	// Find the JSON object by matching braces
	jsonStart := -1
	braceCount := 0
	jsonEnd := -1

	for i, r := range response {
		if r == '{' {
			if jsonStart == -1 {
				jsonStart = i
			}
			braceCount++
		} else if r == '}' {
			braceCount--
			if braceCount == 0 && jsonStart != -1 {
				jsonEnd = i + 1
				break
			}
		}
	}

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON found in response: %s", response)
	}

	jsonStr := response[jsonStart:jsonEnd]
	utils.Info("Full LLM Response: %s", response)
	utils.Info("Extracted JSON: %s", jsonStr)

	// Parse the JSON response
	var planData struct {
		Strategy string `json:"strategy"`
		Tasks    []struct {
			AgentType    string            `json:"agent_type"`
			Priority     int               `json:"priority"`
			Dependencies []string          `json:"dependencies"`
			Parameters   map[string]string `json:"parameters"`
		} `json:"tasks"`
		Reasoning string `json:"reasoning"`
	}

	utils.Info("Parsing JSON into planData structure...")
	if err := json.Unmarshal([]byte(jsonStr), &planData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal orchestration plan: %w", err)
	}

	// Convert to our OrchestrationPlan format
	plan := &models.OrchestrationPlan{
		QueryID:   query.ID,
		Reasoning: planData.Reasoning,
	}

	// Set execution strategy
	switch planData.Strategy {
	case "parallel":
		plan.Strategy = models.ExecutionParallel
	case "sequential":
		plan.Strategy = models.ExecutionSequential
	default:
		return nil, fmt.Errorf("invalid execution strategy: %s", planData.Strategy)
	}

	// Convert tasks
	for _, taskData := range planData.Tasks {
		task := models.AgentTask{
			TaskID:    fmt.Sprintf("task-%d", len(plan.Tasks)+1),
			DependsOn: taskData.Dependencies,
		}

		// Set agent type
		switch taskData.AgentType {
		case "temperature":
			task.AgentType = models.AgentTypeTemperature
			task.City = taskData.Parameters["city"]
		case "datetime":
			task.AgentType = models.AgentTypeDateTime
			task.City = taskData.Parameters["city"]
		case "echo":
			task.AgentType = models.AgentTypeEcho
			task.EchoText = taskData.Parameters["text"]
			utils.Info("Echo task created with text: '%s'", task.EchoText)
			// Validate that echo text is provided
			if task.EchoText == "" {
				return nil, fmt.Errorf("echo agent requires non-empty 'text' parameter")
			}
		default:
			return nil, fmt.Errorf("invalid agent type: %s", taskData.AgentType)
		}

		plan.Tasks = append(plan.Tasks, task)
	}

	if len(plan.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks generated in orchestration plan")
	}

	return plan, nil
}
