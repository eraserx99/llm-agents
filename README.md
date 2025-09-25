# LLM Multi-Agent System

A Go-based demonstration of intelligent multi-agent coordination using Claude 3.5 Sonnet via OpenRouter. The system coordinates specialized sub-agents for temperature, datetime, and echo queries using Model Context Protocol (MCP) servers.

## 🏗️ Architecture

The system features **LLM-driven orchestration** where Claude 3.5 Sonnet analyzes user queries and dynamically decides:
- Which agents to invoke
- Whether to run agents in parallel or sequence
- How to coordinate multiple data requests

### Components

- **Coordinator Agent**: Main orchestrator using Claude 3.5 Sonnet
- **Temperature Agent**: Retrieves weather data via MCP weather server
- **DateTime Agent**: Handles timezone-aware datetime queries via MCP datetime server
- **Echo Agent**: Simple text echo functionality for testing orchestration
- **MCP Servers**: Three independent JSON-RPC 2.0 servers for data services

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- OpenRouter API key for Claude 3.5 Sonnet access

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd llm-agents

# Initialize Go module
go mod tidy

# Build all components
make build
# OR build manually:
go build -o bin/llm-agents ./cmd/main
go build -o bin/weather-mcp ./cmd/weather-mcp
go build -o bin/datetime-mcp ./cmd/datetime-mcp
go build -o bin/echo-mcp ./cmd/echo-mcp
```

### Setup

1. **Get OpenRouter API Key**: Sign up at [openrouter.ai](https://openrouter.ai) and get your API key

2. **Set Environment Variable**:
```bash
export OPENROUTER_API_KEY="your-api-key-here"
```

### Running the System

1. **Start MCP Servers** (in separate terminals):
```bash
# Terminal 1: Weather MCP Server (port 8081)
./bin/weather-mcp

# Terminal 2: DateTime MCP Server (port 8082)
./bin/datetime-mcp

# Terminal 3: Echo MCP Server (port 8083)
./bin/echo-mcp
```

2. **Run Queries**:
```bash
# Temperature query
./bin/llm-agents -city "New York" -query "What's the temperature?"

# DateTime query
./bin/llm-agents -city "Los Angeles" -query "What time is it?"

# Combined query (runs in parallel)
./bin/llm-agents -city "Chicago" -query "What's the weather and time?"

# Echo query (only invokes echo agent)
./bin/llm-agents -query "echo hello world"

# Verbose mode to see orchestration details
./bin/llm-agents -city "Miami" -query "temperature please" -verbose
```

## 🎯 Key Features

### Agent Transparency
The system shows exactly which agents are invoked for each query:

```bash
$ ./bin/llm-agents -city "Boston" -query "weather and time please" -verbose

Query ID: query-1695123456789
Message: Query completed successfully
Duration: 2.1s
Invoked agents: temperature, datetime

🌡️  Temperature in Boston:
   Temperature: 72.0°F
   Conditions: Partly cloudy
   Source: weather-mcp

🕐 Time in Boston:
   Local time: 2024-09-23 14:30:45
   Timezone: America/New_York
   UTC offset: -04:00

📋 Orchestration Details:
   Execution log:
     1. temperature agent: success
     2. datetime agent: success
```

### Intelligent Orchestration
Claude 3.5 Sonnet makes smart decisions about:
- **Parallel execution**: Weather + time queries run simultaneously
- **Sequential execution**: When one result depends on another
- **Agent selection**: Echo agent only used when explicitly requested

### Echo Agent Behavior
- **Weather/DateTime queries**: Echo agent is NOT invoked
- **Explicit echo requests**: Only echo agent is invoked
- **Mixed queries**: All relevant agents are invoked appropriately

## 📊 Example Queries

| Query | Invoked Agents | Execution |
|-------|---------------|-----------|
| `"temperature in NYC"` | temperature | sequential |
| `"what time in LA"` | datetime | sequential |
| `"weather and time in Chicago"` | temperature, datetime | parallel |
| `"echo hello world"` | echo | sequential |

## 🔧 Configuration

### Environment Variables

```bash
# Required
OPENROUTER_API_KEY="your-api-key"

# Optional MCP Server URLs (defaults shown)
MCP_WEATHER_URL="http://localhost:8081"
MCP_DATETIME_URL="http://localhost:8082"
MCP_ECHO_URL="http://localhost:8083"

# Optional Timeouts
MCP_TIMEOUT="10s"      # Timeout for MCP server calls
LLM_TIMEOUT="15s"      # Timeout for Claude API calls
QUERY_TIMEOUT="30s"    # Overall query timeout

# Optional Logging
LOG_LEVEL="INFO"  # DEBUG, INFO, WARN, ERROR
```

### Supported Cities
The system supports 100+ US cities with proper timezone handling:
- Major cities: New York, Los Angeles, Chicago, Houston, Phoenix
- Aliases: NYC, LA, etc.
- Timezone-aware: Handles EST, PST, CST, MST, etc.

## 🛠️ Development

### Project Structure
```
├── cmd/
│   ├── main/           # CLI application
│   ├── weather-mcp/    # Weather MCP server
│   ├── datetime-mcp/   # DateTime MCP server
│   └── echo-mcp/       # Echo MCP server
├── internal/
│   ├── agents/         # Agent implementations
│   ├── mcp/           # MCP server framework
│   ├── config/        # Configuration
│   └── utils/         # Utilities
└── test/              # Test files
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Verbose test output
go test -v ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Lint (if golangci-lint installed)
golangci-lint run
```

## 📋 API Reference

### CLI Options
```bash
Usage: llm-agents [options]

Options:
  -city string
        City name for weather/datetime queries (required for non-echo queries)
  -query string
        Query text (required)
  -verbose
        Enable verbose output with orchestration details
  -version
        Show version information
```

### MCP Protocol
All MCP servers implement JSON-RPC 2.0 protocol:

**Weather Server (port 8081)**
```json
{
  "jsonrpc": "2.0",
  "method": "getTemperature",
  "params": {"city": "New York"},
  "id": 1
}
```

**DateTime Server (port 8082)**
```json
{
  "jsonrpc": "2.0",
  "method": "getDateTime",
  "params": {"city": "Los Angeles"},
  "id": 1
}
```

**Echo Server (port 8083)**
```json
{
  "jsonrpc": "2.0",
  "method": "echo",
  "params": {"text": "hello world"},
  "id": 1
}
```

## ⚙️ How It Works

1. **User Query**: CLI accepts natural language query
2. **LLM Analysis**: Claude 3.5 Sonnet analyzes query and creates orchestration plan
3. **Agent Selection**: System determines which agents to invoke
4. **Execution**: Agents run in parallel or sequence based on LLM decision
5. **MCP Communication**: Agents call respective MCP servers via JSON-RPC
6. **Data Aggregation**: Results are combined and formatted for display
7. **Response**: User sees results with agent transparency

## 🔍 Error Handling

The system gracefully handles:
- Invalid cities (returns appropriate error)
- MCP server failures (shows which agent failed)
- Network timeouts (configurable timeouts)
- OpenRouter API issues (clear error messages)

## 🚦 Troubleshooting

**MCP servers not starting?**
- Check if ports 8081-8083 are available
- Look for error messages in server output

**OpenRouter API errors?**
- Verify your API key is set: `echo $OPENROUTER_API_KEY`
- Check your OpenRouter account balance
- Ensure API key has Claude 3.5 Sonnet access

**City not found errors?**
- Use major US city names: "New York", "Los Angeles", "Chicago"
- Try aliases: "NYC", "LA"
- Check the datetime handler for supported cities

**Agent not responding?**
- Use `-verbose` flag to see orchestration details
- Check MCP server logs for errors
- Verify network connectivity to MCP servers

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🔗 Links

- [OpenRouter](https://openrouter.ai) - Claude 3.5 Sonnet API access
- [MCP Specification](https://modelcontextprotocol.io) - Model Context Protocol
- [wttr.in](https://wttr.in) - Free weather API used by weather MCP server