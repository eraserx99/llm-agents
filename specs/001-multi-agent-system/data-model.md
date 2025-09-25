# Data Model Specification

**Feature**: Multi-Agent System for Temperature and DateTime Queries
**Date**: 2025-09-23

## Core Entities

### 1. Query
Represents user's natural language input
```go
type Query struct {
    ID        string    // Unique query identifier
    Text      string    // Original user input
    City      string    // Extracted city name
    QueryType QueryType // Temperature, DateTime, or Both
    Timestamp time.Time // When query was received
}

type QueryType int
const (
    QueryTypeTemperature QueryType = iota
    QueryTypeDateTime
    QueryTypeBoth
)
```

### 2. City
Represents a US city with required metadata
```go
type City struct {
    Name      string  // City name (e.g., "New York City")
    State     string  // State code (e.g., "NY")
    Latitude  float64 // For weather API
    Longitude float64 // For weather API
    Timezone  string  // IANA timezone (e.g., "America/New_York")
}
```

### 3. Temperature Data
Temperature information from MCP weather server
```go
type TemperatureData struct {
    City        string    // City name
    Temperature float64   // Temperature value
    Unit        string    // "F" for Fahrenheit
    Description string    // Weather description (e.g., "Cloudy")
    Timestamp   time.Time // When data was fetched
    Source      string    // Data source identifier
}
```

### 4. DateTime Data
DateTime information from MCP datetime server
```go
type DateTimeData struct {
    City      string    // City name
    DateTime  time.Time // Local time in city
    Timezone  string    // IANA timezone string
    UTCOffset string    // UTC offset (e.g., "-05:00")
    Timestamp time.Time // When data was fetched
}
```

### 5. Echo Data
Echo response from MCP echo server
```go
type EchoData struct {
    OriginalText string    // Original text to echo
    EchoText     string    // Echoed text (should match original)
    Timestamp    time.Time // When echo was processed
}
```

### 6. Orchestration Plan
LLM-generated execution strategy for the query
```go
type OrchestrationPlan struct {
    QueryID     string            // Original query ID
    Strategy    ExecutionStrategy // Parallel or Sequential
    Tasks       []AgentTask       // Ordered list of agent tasks
    Reasoning   string            // LLM's reasoning for chosen strategy
}

type ExecutionStrategy string
const (
    ExecutionParallel   ExecutionStrategy = "parallel"
    ExecutionSequential ExecutionStrategy = "sequential"
)

type AgentTask struct {
    TaskID    string    // Unique task identifier
    AgentType AgentType // Temperature, DateTime, or Echo
    City      string    // City to query (empty for echo)
    EchoText  string    // Text to echo (only for echo agent)
    DependsOn []string  // Task IDs this depends on (empty for parallel)
}

type AgentType string
const (
    AgentTypeTemperature AgentType = "temperature"
    AgentTypeDateTime    AgentType = "datetime"
    AgentTypeEcho        AgentType = "echo"
)
```

### 7. Agent Request
Request from coordinator to sub-agent
```go
type AgentRequest struct {
    RequestID string        // Unique request identifier
    TaskID    string        // Task ID from orchestration plan
    AgentType AgentType     // Type of agent being requested
    City      string        // City to query (empty for echo)
    EchoText  string        // Text to echo (only for echo agent)
    Timeout   time.Duration // Request timeout
}
```

### 8. Agent Response
Response from sub-agent to coordinator
```go
type AgentResponse struct {
    RequestID string      // Matching request ID
    Success   bool        // Whether request succeeded
    Data      interface{} // TemperatureData or DateTimeData
    Error     string      // Error message if failed
}
```

### 9. Query Response
Final response to user
```go
type QueryResponse struct {
    QueryID          string            // Original query ID
    Message          string            // Natural language response
    Temperature      *TemperatureData  // Optional temperature data
    DateTime         *DateTimeData     // Optional datetime data
    Echo             *EchoData         // Optional echo data
    InvokedAgents    []AgentType       // Which agents were actually invoked
    OrchestrationLog []string          // Step-by-step orchestration decisions
    Errors           []string          // Any errors encountered
    Duration         time.Duration     // Total processing time
}
```

## MCP Protocol Messages

### Weather MCP Messages
```go
// Request
type WeatherRequest struct {
    Method string `json:"method"` // "getTemperature"
    Params struct {
        City string `json:"city"`
    } `json:"params"`
}

// Response
type WeatherResponse struct {
    Result struct {
        Temperature float64 `json:"temperature"`
        Unit        string  `json:"unit"`
        Description string  `json:"description"`
    } `json:"result"`
    Error *MCPError `json:"error,omitempty"`
}
```

### DateTime MCP Messages
```go
// Request
type DateTimeRequest struct {
    Method string `json:"method"` // "getDateTime"
    Params struct {
        City string `json:"city"`
    } `json:"params"`
}

// Response
type DateTimeResponse struct {
    Result struct {
        DateTime  string `json:"datetime"`  // ISO 8601 format
        Timezone  string `json:"timezone"`
        UTCOffset string `json:"utc_offset"`
    } `json:"result"`
    Error *MCPError `json:"error,omitempty"`
}
```

### Echo MCP Messages
```go
// Request
type EchoRequest struct {
    Method string `json:"method"` // "echo"
    Params struct {
        Text string `json:"text"`
    } `json:"params"`
}

// Response
type EchoResponse struct {
    Result struct {
        OriginalText string `json:"original_text"`
        EchoText     string `json:"echo_text"`
    } `json:"result"`
    Error *MCPError `json:"error,omitempty"`
}
```

### MCP Error
```go
type MCPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

## State Transitions

### Query Processing States
1. **Received**: Query received from user
2. **Analyzing**: LLM coordinator analyzing query
3. **Planned**: Orchestration plan created with execution strategy
4. **Dispatching**: Sending requests to sub-agents (parallel or sequential)
5. **Processing**: Sub-agents querying MCP servers
6. **Aggregating**: Collecting responses from sub-agents
7. **Completed**: Response formatted and returned
8. **Failed**: Error occurred, error message returned

### Sub-Agent States
1. **Idle**: Waiting for requests
2. **Connecting**: Establishing MCP connection
3. **Querying**: Sending request to MCP server
4. **Waiting**: Awaiting MCP response
5. **Responding**: Sending response to coordinator

## Validation Rules

### Query Validation
- Text must be non-empty
- Text length must be < 500 characters
- Must contain a recognizable US city name

### City Validation
- City name must match predefined US cities list
- Case-insensitive matching
- No special characters except spaces and hyphens

### Temperature Validation
- Temperature must be between -100°F and 150°F
- Unit must be "F" or "C"

### DateTime Validation
- DateTime must be valid time.Time
- Timezone must be valid IANA timezone
- DateTime should be within reasonable bounds (not far future/past)

### Echo Validation
- Echo text must be non-empty
- Echo text max length: 1000 characters
- Echo response must match original text exactly

### Response Validation
- At least one data field must be present if Success=true
- Error message required if Success=false
- Duration must be positive

## Relationships

1. **Query** → **City**: One-to-one (each query targets one city)
2. **Query** → **OrchestrationPlan**: One-to-one
3. **OrchestrationPlan** → **AgentTask**: One-to-many
4. **AgentTask** → **AgentRequest**: One-to-one
5. **Query** → **QueryResponse**: One-to-one
6. **QueryResponse** → **TemperatureData**: Zero-or-one
7. **QueryResponse** → **DateTimeData**: Zero-or-one
8. **QueryResponse** → **EchoData**: Zero-or-one
9. **AgentRequest** → **AgentResponse**: One-to-one
10. **City** → **TemperatureData**: One-to-many (multiple queries)
11. **City** → **DateTimeData**: One-to-many (multiple queries)

## Data Constraints

- All timestamps in UTC internally, converted for display
- City names normalized to title case
- Temperature always stored in Fahrenheit, converted as needed
- All string fields trimmed of leading/trailing whitespace
- Maximum concurrent sub-agent requests: 2
- Request timeout: 5 seconds default
- MCP server connection timeout: 2 seconds