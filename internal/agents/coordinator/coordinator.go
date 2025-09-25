// Package coordinator provides the main coordinator agent implementation
package coordinator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/steve/llm-agents/internal/agents/datetime"
	"github.com/steve/llm-agents/internal/agents/echo"
	"github.com/steve/llm-agents/internal/agents/temperature"
	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// Coordinator implements the main coordinator agent
type Coordinator struct {
	llmClient        *LLMClient
	temperatureAgent *temperature.Agent
	datetimeAgent    *datetime.Agent
	echoAgent        *echo.Agent
	requestCounter   int64
	mu               sync.RWMutex
}

// NewCoordinator creates a new coordinator agent
func NewCoordinator(openRouterAPIKey, weatherServerURL, datetimeServerURL, echoServerURL string, timeout time.Duration) *Coordinator {
	return &Coordinator{
		llmClient:        NewLLMClient(openRouterAPIKey),
		temperatureAgent: temperature.NewAgent(weatherServerURL, timeout),
		datetimeAgent:    datetime.NewAgent(datetimeServerURL, timeout),
		echoAgent:        echo.NewAgent(echoServerURL, timeout),
		requestCounter:   0,
	}
}

// ProcessQuery processes a user query and coordinates sub-agents
func (c *Coordinator) ProcessQuery(ctx context.Context, query models.Query) (*models.QueryResponse, error) {
	utils.Info("Coordinator processing query: %s (city: %s)", query.Text, query.City)

	// Generate orchestration plan using LLM
	plan, err := c.llmClient.GenerateOrchestrationPlan(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate orchestration plan: %w", err)
	}

	utils.Info("Orchestration plan: %s strategy with %d tasks", plan.Strategy, len(plan.Tasks))
	utils.Debug("Plan reasoning: %s", plan.Reasoning)

	// Execute the orchestration plan
	responses, err := c.executePlan(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute orchestration plan: %w", err)
	}

	// Build the final response
	queryResponse := c.buildQueryResponse(query, plan, responses)

	utils.Info("Query processed successfully, invoked agents: %v", queryResponse.InvokedAgents)
	return queryResponse, nil
}

// executePlan executes the orchestration plan
func (c *Coordinator) executePlan(ctx context.Context, plan *models.OrchestrationPlan) ([]*models.AgentResponse, error) {
	switch plan.Strategy {
	case models.ExecutionParallel:
		return c.executeParallel(ctx, plan.Tasks)
	case models.ExecutionSequential:
		return c.executeSequential(ctx, plan.Tasks)
	default:
		return nil, fmt.Errorf("unsupported execution strategy: %s", plan.Strategy)
	}
}

// executeParallel executes tasks in parallel
func (c *Coordinator) executeParallel(ctx context.Context, tasks []models.AgentTask) ([]*models.AgentResponse, error) {
	utils.Debug("Executing %d tasks in parallel", len(tasks))

	type result struct {
		response *models.AgentResponse
		err      error
		index    int
	}

	resultChan := make(chan result, len(tasks))
	var wg sync.WaitGroup

	// Start all tasks concurrently
	for i, task := range tasks {
		wg.Add(1)
		go func(taskIndex int, t models.AgentTask) {
			defer wg.Done()

			response, err := c.executeTask(ctx, t)
			resultChan <- result{
				response: response,
				err:      err,
				index:    taskIndex,
			}
		}(i, task)
	}

	// Wait for all tasks to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	responses := make([]*models.AgentResponse, len(tasks))
	var firstError error

	for res := range resultChan {
		if res.err != nil && firstError == nil {
			firstError = res.err
			utils.Error("Task %d failed: %v", res.index, res.err)
		}
		responses[res.index] = res.response
	}

	if firstError != nil {
		return responses, firstError
	}

	return responses, nil
}

// executeSequential executes tasks sequentially
func (c *Coordinator) executeSequential(ctx context.Context, tasks []models.AgentTask) ([]*models.AgentResponse, error) {
	utils.Debug("Executing %d tasks sequentially", len(tasks))

	responses := make([]*models.AgentResponse, len(tasks))

	for i, task := range tasks {
		response, err := c.executeTask(ctx, task)
		if err != nil {
			return responses, fmt.Errorf("task %d failed: %w", i, err)
		}
		responses[i] = response
	}

	return responses, nil
}

// executeTask executes a single agent task
func (c *Coordinator) executeTask(ctx context.Context, task models.AgentTask) (*models.AgentResponse, error) {
	// Generate unique IDs for this task
	c.mu.Lock()
	c.requestCounter++
	requestID := fmt.Sprintf("req-%d", c.requestCounter)
	taskID := fmt.Sprintf("task-%d", c.requestCounter)
	c.mu.Unlock()

	// Create agent request
	request := models.AgentRequest{
		RequestID: requestID,
		TaskID:    taskID,
		AgentType: task.AgentType,
		City:      task.City,
		EchoText:  task.EchoText,
		Timeout:   15 * time.Second, // Default timeout for agent operations
	}

	utils.Debug("Executing task: %s agent for %s", task.AgentType, getTaskDescription(task))

	// Route to appropriate agent
	switch task.AgentType {
	case models.AgentTypeTemperature:
		return c.temperatureAgent.ProcessRequest(ctx, request)
	case models.AgentTypeDateTime:
		return c.datetimeAgent.ProcessRequest(ctx, request)
	case models.AgentTypeEcho:
		return c.echoAgent.ProcessRequest(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", task.AgentType)
	}
}

// buildQueryResponse builds the final query response
func (c *Coordinator) buildQueryResponse(query models.Query, plan *models.OrchestrationPlan, responses []*models.AgentResponse) *models.QueryResponse {
	startTime := time.Now()
	response := &models.QueryResponse{
		QueryID:          query.ID,
		InvokedAgents:    make([]models.AgentType, 0),
		OrchestrationLog: make([]string, 0),
		Errors:           make([]string, 0),
	}

	// Extract invoked agents and collect data
	agentsSeen := make(map[models.AgentType]bool)
	hasError := false

	for i, agentResponse := range responses {
		if agentResponse == nil {
			continue
		}

		// Get agent type from the plan task
		var agentType models.AgentType
		if i < len(plan.Tasks) {
			agentType = plan.Tasks[i].AgentType
		}

		// Track invoked agents
		if !agentsSeen[agentType] {
			response.InvokedAgents = append(response.InvokedAgents, agentType)
			agentsSeen[agentType] = true
		}

		// Add orchestration log entry
		logEntry := fmt.Sprintf("%s agent: ", agentType)
		if agentResponse.Success {
			logEntry += "success"
		} else {
			logEntry += "failed"
			hasError = true
			if agentResponse.Error != "" {
				response.Errors = append(response.Errors, agentResponse.Error)
			}
		}
		response.OrchestrationLog = append(response.OrchestrationLog, logEntry)

		// Collect successful responses data
		if agentResponse.Success && agentResponse.Data != nil {
			switch agentType {
			case models.AgentTypeTemperature:
				if tempData, ok := agentResponse.Data.(*models.TemperatureData); ok {
					response.Temperature = tempData
				}
			case models.AgentTypeDateTime:
				if dateTimeData, ok := agentResponse.Data.(*models.DateTimeData); ok {
					response.DateTime = dateTimeData
				}
			case models.AgentTypeEcho:
				if echoData, ok := agentResponse.Data.(*models.EchoData); ok {
					response.Echo = echoData
				}
			}
		}
	}

	// Set message based on results
	if hasError {
		response.Message = "Query completed with errors"
	} else {
		response.Message = "Query completed successfully"
	}

	response.Duration = time.Since(startTime)
	return response
}

// getTaskDescription returns a human-readable description of the task
func getTaskDescription(task models.AgentTask) string {
	switch task.AgentType {
	case models.AgentTypeTemperature:
		return fmt.Sprintf("city: %s", task.City)
	case models.AgentTypeDateTime:
		return fmt.Sprintf("city: %s", task.City)
	case models.AgentTypeEcho:
		return fmt.Sprintf("text: %s", task.EchoText)
	default:
		return "unknown"
	}
}

// Validate validates the coordinator configuration
func (c *Coordinator) Validate() error {
	if c.llmClient == nil {
		return fmt.Errorf("LLM client is not initialized")
	}
	if c.temperatureAgent == nil {
		return fmt.Errorf("temperature agent is not initialized")
	}
	if c.datetimeAgent == nil {
		return fmt.Errorf("datetime agent is not initialized")
	}
	if c.echoAgent == nil {
		return fmt.Errorf("echo agent is not initialized")
	}

	// Validate sub-agents
	if err := c.temperatureAgent.Validate(); err != nil {
		return fmt.Errorf("temperature agent validation failed: %w", err)
	}
	if err := c.datetimeAgent.Validate(); err != nil {
		return fmt.Errorf("datetime agent validation failed: %w", err)
	}
	if err := c.echoAgent.Validate(); err != nil {
		return fmt.Errorf("echo agent validation failed: %w", err)
	}

	return nil
}

// Close closes the coordinator and all sub-agents
func (c *Coordinator) Close() {
	if c.temperatureAgent != nil {
		c.temperatureAgent.Close()
	}
	if c.datetimeAgent != nil {
		c.datetimeAgent.Close()
	}
	if c.echoAgent != nil {
		c.echoAgent.Close()
	}
}
