# Quickstart Guide: Multi-Agent System

## Prerequisites

1. **Go 1.21+** installed
2. **OpenRouter API Key** for Claude 3.5 Sonnet
3. **Git** for cloning the repository

## Installation

### 1. Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd llm-agents

# Install dependencies
go mod download

# Set your OpenRouter API key
export OPENROUTER_API_KEY="your-api-key-here"
```

### 2. Build the System
```bash
# Build all components
go build -o bin/llm-agents ./cmd/main
go build -o bin/weather-mcp ./cmd/weather-mcp
go build -o bin/datetime-mcp ./cmd/datetime-mcp
go build -o bin/echo-mcp ./cmd/echo-mcp

# Or use make if available
make build
```

### 3. Start MCP Servers
Open three terminal windows:

**Terminal 1 - Weather MCP Server:**
```bash
./bin/weather-mcp
# Server starting on port 8081...
```

**Terminal 2 - DateTime MCP Server:**
```bash
./bin/datetime-mcp
# Server starting on port 8082...
```

**Terminal 3 - Echo MCP Server:**
```bash
./bin/echo-mcp
# Server starting on port 8083...
```

## Usage Examples

### Basic Queries

**Temperature Query:**
```bash
./bin/llm-agents "What is the temperature in New York City right now?"
# Expected: The current temperature in New York City is 72.5°F with partly cloudy conditions.
```

**DateTime Query:**
```bash
./bin/llm-agents "What time is it in Los Angeles?"
# Expected: The current time in Los Angeles is 12:30 PM PST (September 23, 2025).
```

**Combined Query:**
```bash
./bin/llm-agents "What is the datetime and temperature of Chicago now?"
# Expected: In Chicago, it is currently 2:30 PM CST (September 23, 2025) with a temperature of 45.0°F and cloudy conditions.
```

**Echo Query:**
```bash
./bin/llm-agents "Please echo: hello world!"
# Expected: hello world!

./bin/llm-agents "Echo this text: The system is working perfectly"
# Expected: The system is working perfectly
```

### Using Command Line Flags

**Specify City Explicitly:**
```bash
./bin/llm-agents -city "Seattle" "What's the weather and time?"
# Uses Seattle regardless of query text
```

**Verbose Output:**
```bash
./bin/llm-agents -verbose "Temperature in Miami?"
# Shows detailed processing steps and timing
```

**Custom Timeout:**
```bash
./bin/llm-agents -timeout 10s "What is the temperature in Boston?"
# Sets 10 second timeout for the query
```

## Testing the System

### 1. Unit Tests
```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...
```

### 2. Integration Tests
```bash
# Ensure MCP servers are running first
go test ./test/integration -tags=integration
```

### 3. Manual Validation Tests

**Test 1: Single Temperature Query**
```bash
./bin/llm-agents "What is the temperature in Houston?"
# ✓ Should return temperature for Houston
# ✓ Response time should be < 1 second
```

**Test 2: Single DateTime Query**
```bash
./bin/llm-agents "What time is it in Phoenix?"
# ✓ Should return current time for Phoenix
# ✓ Time should be in MST/PDT depending on season
```

**Test 3: Parallel Processing**
```bash
./bin/llm-agents "Tell me the time and temperature in Denver"
# ✓ Should return both datetime and temperature
# ✓ Should complete faster than sequential queries
```

**Test 4: Error Handling - Invalid City**
```bash
./bin/llm-agents "Temperature in Faketown?"
# ✓ Should return error: City not found
# ✓ Exit code should be 2
```

**Test 5: Error Handling - MCP Server Down**
```bash
# Stop weather MCP server, then:
./bin/llm-agents "What's the temperature in Dallas?"
# ✓ Should return error: Weather service unavailable
# ✓ Exit code should be 3
```

**Test 6: City Flag Override**
```bash
./bin/llm-agents -city "San Francisco" "What's the weather in New York?"
# ✓ Should return San Francisco weather (not New York)
```

**Test 7: Echo Agent Only**
```bash
./bin/llm-agents "Please echo: testing echo functionality"
# ✓ Should return: testing echo functionality
# ✓ Should NOT invoke weather or datetime agents
# ✓ Verbose mode should show: Selected agents: [echo]
```

**Test 8: Agent Selection Visibility**
```bash
./bin/llm-agents -verbose "What's the temperature in Boston?"
# ✓ Should show: Selected agents: [temperature]
# ✓ Should NOT show datetime or echo agents in the log

./bin/llm-agents -verbose "Echo this: hello world"
# ✓ Should show: Selected agents: [echo]
# ✓ Should NOT show weather or datetime agents in the log
```

## Troubleshooting

### Common Issues

**Issue: "Connection refused" errors**
- Solution: Ensure MCP servers are running on ports 8081, 8082, and 8083

**Issue: "API key not found" error**
- Solution: Set `OPENROUTER_API_KEY` environment variable

**Issue: "City not found" for valid cities**
- Solution: Check city name spelling, use full names (e.g., "New York City" not "NYC")

**Issue: Timeout errors**
- Solution: Increase timeout with `-timeout 10s` flag

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=DEBUG
./bin/llm-agents -verbose "Temperature in Atlanta?"
```

### Health Check
```bash
# Check if MCP servers are responding
curl -X POST http://localhost:8081/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getTemperature","params":{"city":"Boston"},"id":1}'

curl -X POST http://localhost:8082/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getDateTime","params":{"city":"Boston"},"id":1}'

curl -X POST http://localhost:8083/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"echo","params":{"text":"test"},"id":1}'
```

## Performance Benchmarks

Expected performance metrics:
- **Single agent query**: < 500ms
- **Dual agent query (parallel)**: < 700ms
- **MCP server response time**: < 200ms
- **Memory usage**: < 50MB per component

## Next Steps

1. **Add more cities**: Extend the US cities database
2. **Custom queries**: Modify agent prompts for different response styles
3. **Additional data sources**: Integrate more MCP servers for extended functionality
4. **Monitoring**: Add Prometheus metrics for production deployment