// Package models defines the core data structures for the multi-agent system
package models

import "time"

// Query represents user's natural language input
type Query struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	City      string    `json:"city"`
	QueryType QueryType `json:"query_type"`
	Timestamp time.Time `json:"timestamp"`
}

// QueryType defines the type of query being made
type QueryType int

const (
	QueryTypeTemperature QueryType = iota
	QueryTypeDateTime
	QueryTypeBoth
	QueryTypeEcho
)

// String returns the string representation of QueryType
func (qt QueryType) String() string {
	switch qt {
	case QueryTypeTemperature:
		return "temperature"
	case QueryTypeDateTime:
		return "datetime"
	case QueryTypeBoth:
		return "both"
	case QueryTypeEcho:
		return "echo"
	default:
		return "unknown"
	}
}

// City represents a US city with required metadata
type City struct {
	Name      string  `json:"name"`
	State     string  `json:"state"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

// TemperatureData contains temperature information from MCP weather server
type TemperatureData struct {
	City        string    `json:"city"`
	Temperature float64   `json:"temperature"`
	Unit        string    `json:"unit"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
}

// DateTimeData contains datetime information from MCP datetime server
type DateTimeData struct {
	City      string    `json:"city"`
	DateTime  time.Time `json:"datetime"`
	Timezone  string    `json:"timezone"`
	UTCOffset string    `json:"utc_offset"`
	Timestamp time.Time `json:"timestamp"`
}

// EchoData contains echo response from MCP echo server
type EchoData struct {
	OriginalText string    `json:"original_text"`
	EchoText     string    `json:"echo_text"`
	Timestamp    time.Time `json:"timestamp"`
}

// OrchestrationPlan represents LLM-generated execution strategy for the query
type OrchestrationPlan struct {
	QueryID   string            `json:"query_id"`
	Strategy  ExecutionStrategy `json:"strategy"`
	Tasks     []AgentTask       `json:"tasks"`
	Reasoning string            `json:"reasoning"`
}

// ExecutionStrategy defines how agents should be executed
type ExecutionStrategy string

const (
	ExecutionParallel   ExecutionStrategy = "parallel"
	ExecutionSequential ExecutionStrategy = "sequential"
)

// AgentTask represents a task to be executed by a specific agent
type AgentTask struct {
	TaskID    string    `json:"task_id"`
	AgentType AgentType `json:"agent_type"`
	City      string    `json:"city,omitempty"`
	EchoText  string    `json:"echo_text,omitempty"`
	DependsOn []string  `json:"depends_on,omitempty"`
}

// AgentType defines the type of agent
type AgentType string

const (
	AgentTypeTemperature AgentType = "temperature"
	AgentTypeDateTime    AgentType = "datetime"
	AgentTypeEcho        AgentType = "echo"
)

// AgentRequest represents a request from coordinator to sub-agent
type AgentRequest struct {
	RequestID string        `json:"request_id"`
	TaskID    string        `json:"task_id"`
	AgentType AgentType     `json:"agent_type"`
	City      string        `json:"city,omitempty"`
	EchoText  string        `json:"echo_text,omitempty"`
	Timeout   time.Duration `json:"timeout"`
}

// AgentResponse represents a response from sub-agent to coordinator
type AgentResponse struct {
	RequestID string      `json:"request_id"`
	TaskID    string      `json:"task_id"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// QueryResponse represents the final response to user
type QueryResponse struct {
	QueryID          string           `json:"query_id"`
	Message          string           `json:"message"`
	Temperature      *TemperatureData `json:"temperature,omitempty"`
	DateTime         *DateTimeData    `json:"datetime,omitempty"`
	Echo             *EchoData        `json:"echo,omitempty"`
	InvokedAgents    []AgentType      `json:"invoked_agents"`
	OrchestrationLog []string         `json:"orchestration_log"`
	Errors           []string         `json:"errors,omitempty"`
	Duration         time.Duration    `json:"duration"`
}

// MCP Protocol Messages

// WeatherRequest represents a request to the weather MCP server
type WeatherRequest struct {
	JSONRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		City string `json:"city"`
	} `json:"params"`
	ID int `json:"id"`
}

// WeatherResponse represents a response from the weather MCP server
type WeatherResponse struct {
	JSONRpc string `json:"jsonrpc"`
	Result  *struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
		Description string  `json:"description"`
	} `json:"result,omitempty"`
	Error *MCPError `json:"error,omitempty"`
	ID    int       `json:"id"`
}

// DateTimeRequest represents a request to the datetime MCP server
type DateTimeRequest struct {
	JSONRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		City string `json:"city"`
	} `json:"params"`
	ID int `json:"id"`
}

// DateTimeResponse represents a response from the datetime MCP server
type DateTimeResponse struct {
	JSONRpc string `json:"jsonrpc"`
	Result  *struct {
		DateTime  string `json:"datetime"`
		Timezone  string `json:"timezone"`
		UTCOffset string `json:"utc_offset"`
	} `json:"result,omitempty"`
	Error *MCPError `json:"error,omitempty"`
	ID    int       `json:"id"`
}

// EchoRequest represents a request to the echo MCP server
type EchoRequest struct {
	JSONRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Text string `json:"text"`
	} `json:"params"`
	ID int `json:"id"`
}

// EchoResponse represents a response from the echo MCP server
type EchoResponse struct {
	JSONRpc string `json:"jsonrpc"`
	Result  *struct {
		OriginalText string `json:"original_text"`
		EchoText     string `json:"echo_text"`
	} `json:"result,omitempty"`
	Error *MCPError `json:"error,omitempty"`
	ID    int       `json:"id"`
}

// MCPError represents an error in MCP protocol
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
