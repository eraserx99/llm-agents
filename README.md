# LLM Multi-Agent System

A Go-based demonstration of intelligent multi-agent coordination using Claude 3.5 Sonnet via OpenRouter. The system coordinates specialized sub-agents for temperature, datetime, and echo queries using Model Context Protocol (MCP) servers with optional **mutual TLS (mTLS) authentication**.

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
go build -o bin/cert-gen ./cmd/cert-gen          # Certificate generator (for TLS)
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

## 🔐 TLS/mTLS Security (Optional)

The system supports **mutual TLS (mTLS) authentication** for secure communication between the coordinator and MCP servers. Both HTTP and HTTPS modes are supported.

### Quick mTLS Setup

1. **Generate Certificates** (one-time setup):
```bash
# Build certificate generator
go build -o bin/cert-gen ./cmd/cert-gen

# Generate CA, server, and client certificates
./bin/cert-gen
# Creates certificates in ./certs/ directory
```

2. **Run with mTLS Enabled**:
```bash
# Set TLS environment variables
export TLS_ENABLED=true
export TLS_DEMO_MODE=true
export TLS_CERT_DIR=./certs

# Start MCP servers with TLS (in separate terminals)
./bin/weather-mcp --tls    # HTTP: 8080, HTTPS: 8443
./bin/datetime-mcp --tls   # HTTP: 8081, HTTPS: 8444
./bin/echo-mcp --tls       # HTTP: 8082, HTTPS: 8445

# Run queries (coordinator auto-detects TLS mode)
./bin/llm-agents -city "New York" -query "What's the temperature?"
```

### TLS Modes Comparison

| Mode | Security | Setup | Use Case |
|------|----------|--------|----------|
| **HTTP** | None | Simple | Development, testing |
| **mTLS** | Full mutual auth | Certificates required | Production, demos |

### Certificate Details

The system uses **self-signed certificates** with a custom Certificate Authority:

```bash
certs/
├── ca.crt          # Certificate Authority (used to sign other certs)
├── ca.key          # CA private key
├── server.crt      # Server certificate (for MCP servers)
├── server.key      # Server private key
├── client.crt      # Client certificate (for coordinator)
└── client.key      # Client private key
```

**Certificate Properties:**
- **Validity**: 1 year from generation
- **Key Size**: 2048-bit RSA
- **Algorithm**: SHA-256 with RSA
- **Extensions**: Proper key usage for TLS client/server authentication
- **SAN**: Includes localhost, 127.0.0.1 for local development

### TLS Environment Variables

```bash
# TLS Control
TLS_ENABLED=true           # Enable/disable TLS mode
TLS_DEMO_MODE=true         # Relaxed validation for self-signed certs
TLS_CERT_DIR=./certs       # Certificate directory path

# Port Configuration
WEATHER_MCP_PORT=8080      # HTTP port for weather server
WEATHER_MCP_TLS_PORT=8443  # HTTPS port for weather server
DATETIME_MCP_PORT=8081     # HTTP port for datetime server
DATETIME_MCP_TLS_PORT=8444 # HTTPS port for datetime server
ECHO_MCP_PORT=8082         # HTTP port for echo server
ECHO_MCP_TLS_PORT=8445     # HTTPS port for echo server
```

### Running HTTP vs HTTPS

**HTTP Mode (Default)**:
```bash
# No TLS variables needed
./bin/weather-mcp          # Runs on HTTP port only
./bin/datetime-mcp
./bin/echo-mcp

# Coordinator uses HTTP clients
./bin/llm-agents -city "Boston" -query "temperature"
```

**HTTPS Mode (mTLS)**:
```bash
# Set TLS environment
export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs

# Servers run both HTTP and HTTPS
./bin/weather-mcp --tls    # HTTP: 8080, HTTPS: 8443
./bin/datetime-mcp --tls   # HTTP: 8081, HTTPS: 8444
./bin/echo-mcp --tls       # HTTP: 8082, HTTPS: 8445

# Coordinator auto-detects and uses HTTPS clients with mTLS
./bin/llm-agents -city "Boston" -query "temperature"
```

### TLS Verification

Test your mTLS setup:
```bash
# Check certificates are valid
openssl verify -CAfile certs/ca.crt certs/server.crt
openssl verify -CAfile certs/ca.crt certs/client.crt

# Test HTTPS endpoints directly
curl -k --cert certs/client.crt --key certs/client.key \
     --cacert certs/ca.crt \
     https://localhost:8443/rpc

# Or use the built-in test
go run test-mtls.go
```

### Security Notes

- **Demo Mode**: Uses relaxed certificate validation suitable for development
- **Production**: Disable `TLS_DEMO_MODE` and use properly signed certificates
- **Mutual Authentication**: Both client and server verify each other's certificates
- **Certificate Rotation**: Regenerate certificates before they expire (1 year)

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
MCP_WEATHER_URL="http://localhost:8081"    # HTTP mode
MCP_DATETIME_URL="http://localhost:8082"   # HTTP mode
MCP_ECHO_URL="http://localhost:8083"       # HTTP mode

# TLS/mTLS Configuration (Optional)
TLS_ENABLED="false"           # Enable TLS mode (true/false)
TLS_DEMO_MODE="true"          # Relaxed validation for self-signed certs
TLS_CERT_DIR="./certs"        # Certificate directory path

# TLS Port Configuration (when TLS_ENABLED=true)
WEATHER_MCP_PORT="8080"       # HTTP port
WEATHER_MCP_TLS_PORT="8443"   # HTTPS port
DATETIME_MCP_PORT="8081"      # HTTP port
DATETIME_MCP_TLS_PORT="8444"  # HTTPS port
ECHO_MCP_PORT="8082"          # HTTP port
ECHO_MCP_TLS_PORT="8445"      # HTTPS port

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
│   ├── echo-mcp/       # Echo MCP server
│   └── cert-gen/       # Certificate generator for TLS
├── internal/
│   ├── agents/         # Agent implementations
│   ├── mcp/           # MCP server framework
│   ├── config/        # Configuration (including TLS)
│   ├── tls/           # TLS certificate management
│   └── utils/         # Utilities
├── test/              # Test files
└── certs/             # TLS certificates (generated)
    ├── ca.crt         # Certificate Authority
    ├── server.crt     # Server certificate
    └── client.crt     # Client certificate
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
- Check if ports 8081-8083 (HTTP) or 8443-8445 (HTTPS) are available
- Look for error messages in server output
- For TLS mode, ensure certificates exist: `ls -la certs/`

**TLS/Certificate issues?**
- Generate certificates: `./bin/cert-gen`
- Verify certificates: `openssl verify -CAfile certs/ca.crt certs/server.crt`
- Check TLS environment variables: `echo $TLS_ENABLED $TLS_DEMO_MODE`
- Run TLS test: `go run test-mtls.go`

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
- For TLS mode, check if servers started with `--tls` flag

**Certificate validation errors?**
- Ensure `TLS_DEMO_MODE=true` for self-signed certificates
- Check certificate expiry: `openssl x509 -in certs/server.crt -text -noout | grep "Not After"`
- Regenerate certificates if expired: `./bin/cert-gen`

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