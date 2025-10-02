# Makefile for LLM Multi-Agent System

.PHONY: build build-sdk build-all clean test lint fmt vet run-servers run-servers-tls stop-servers generate-certs demo-streaming demo-streaming-tls demo dev-setup install-tools help

# Build all binaries
build:
	@echo "Building all components..."
	@echo "All MCP servers now use official MCP Go SDK with StreamableHTTPHandler"
	@mkdir -p bin
	go build -o bin/llm-agents ./cmd/main
	go build -o bin/weather-mcp ./cmd/weather-mcp
	go build -o bin/datetime-mcp ./cmd/datetime-mcp
	go build -o bin/echo-mcp ./cmd/echo-mcp
	go build -o bin/cert-gen ./cmd/cert-gen
	@echo "Build complete! Binaries are in ./bin/"

# Build MCP SDK components (Official MCP Go SDK with HTTP/SSE)
build-sdk:
	@echo "Building MCP SDK components..."
	@mkdir -p bin
	go build -o bin/weather-mcp-sdk ./cmd/weather-mcp-sdk
	go build -o bin/test-mcp-client ./cmd/test-mcp-client
	@echo "MCP SDK build complete! Binaries are in ./bin/"

# Build everything including SDK components
build-all: build build-sdk
	@echo "All components built successfully!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run

# Run all quality checks
quality: fmt vet test
	@echo "All quality checks complete!"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Start MCP servers in background (all now use official MCP SDK)
run-servers:
	@echo "Starting MCP servers with official MCP Go SDK..."
	@echo "All servers now use mcp.NewServer() and mcp.NewStreamableHTTPHandler()"
	@echo "Starting weather MCP server on port 8081..."
	@./bin/weather-mcp > weather-mcp.log 2>&1 &
	@echo "Starting datetime MCP server on port 8082..."
	@./bin/datetime-mcp > datetime-mcp.log 2>&1 &
	@echo "Starting echo MCP server on port 8083..."
	@./bin/echo-mcp > echo-mcp.log 2>&1 &
	@echo "All servers started! Check *.log files for output."
	@echo "Use 'make stop-servers' to stop them."

# Start MCP servers with TLS in background
run-servers-tls: generate-certs
	@echo "Starting MCP servers with TLS enabled..."
	@echo "All servers will run both HTTP and HTTPS endpoints"
	@echo "Starting weather MCP server (HTTP: 8081, HTTPS: 8443)..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/weather-mcp --tls > weather-mcp-tls.log 2>&1 &
	@echo "Starting datetime MCP server (HTTP: 8082, HTTPS: 8444)..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/datetime-mcp --tls > datetime-mcp-tls.log 2>&1 &
	@echo "Starting echo MCP server (HTTP: 8083, HTTPS: 8445)..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/echo-mcp --tls > echo-mcp-tls.log 2>&1 &
	@echo "All TLS servers started! Check *-tls.log files for output."
	@echo "Use 'make stop-servers' to stop them."

# Stop MCP servers
stop-servers:
	@echo "Stopping MCP servers..."
	@pkill -f weather-mcp || true
	@pkill -f datetime-mcp || true
	@pkill -f echo-mcp || true
	@echo "All servers stopped!"

# Generate TLS certificates for mTLS demos
generate-certs: build
	@echo "Generating TLS certificates..."
	@./bin/cert-gen
	@echo "Certificates generated in ./certs/ directory!"

# Run MCP Streaming Protocol demo (HTTP mode)
demo-streaming: build-sdk
	@echo "=== MCP Streaming Protocol Demo (HTTP) ==="
	@echo ""
	@echo "Starting MCP SDK server on port 8091..."
	@./bin/weather-mcp-sdk -verbose > weather-sdk.log 2>&1 &
	@sleep 2
	@echo "Testing MCP client with streaming protocol..."
	@./bin/test-mcp-client -verbose -city "New York" || true
	@echo ""
	@echo "Demo complete! Check weather-sdk.log for server output."
	@pkill -f weather-mcp-sdk || true

# Run MCP Streaming Protocol demo (mTLS mode)
demo-streaming-tls: build-sdk generate-certs
	@echo "=== MCP Streaming Protocol Demo (mTLS) ==="
	@echo ""
	@echo "Starting MCP SDK server with mTLS on port 8491..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/weather-mcp-sdk -tls -verbose > weather-sdk-tls.log 2>&1 &
	@sleep 2
	@echo "Testing MCP client with mTLS streaming protocol..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/test-mcp-client -tls -verbose -city "Tokyo" || true
	@echo ""
	@echo "mTLS demo complete! Check weather-sdk-tls.log for server output."
	@pkill -f weather-mcp-sdk || true

# Example usage demonstrations
demo: build
	@echo "=== LLM Multi-Agent System Demo ==="
	@echo ""
	@echo "First, make sure you have OPENROUTER_API_KEY set:"
	@echo "export OPENROUTER_API_KEY=\"your-key-here\""
	@echo ""
	@echo "Then start the MCP servers:"
	@echo "make run-servers"
	@echo ""
	@echo "Example queries:"
	@echo ""
	@echo "1. Temperature query:"
	@echo "./bin/llm-agents -city \"New York\" -query \"What's the temperature?\""
	@echo ""
	@echo "2. DateTime query:"
	@echo "./bin/llm-agents -city \"Los Angeles\" -query \"What time is it?\""
	@echo ""
	@echo "3. Combined query (parallel execution):"
	@echo "./bin/llm-agents -city \"Chicago\" -query \"What's the weather and time?\""
	@echo ""
	@echo "4. Echo query:"
	@echo "./bin/llm-agents -query \"echo hello world\""
	@echo ""
	@echo "5. Verbose mode:"
	@echo "./bin/llm-agents -city \"Miami\" -query \"temperature please\" -verbose"

# Development setup
dev-setup: deps build
	@echo "Development setup complete!"
	@echo "Run 'make demo' for usage examples."

# Install tools (optional)
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed!"

# Help
help:
	@echo "LLM Multi-Agent System Makefile"
	@echo ""
	@echo "Build Targets:"
	@echo "  make build             - Build all legacy components"
	@echo "  make build-sdk         - Build MCP SDK components (official SDK)"
	@echo "  make build-all         - Build everything (legacy + SDK)"
	@echo "  make clean             - Clean build artifacts"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test              - Run tests"
	@echo "  make test-race         - Run tests with race detection"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make fmt               - Format code"
	@echo "  make vet               - Vet code"
	@echo "  make lint              - Lint code (requires golangci-lint)"
	@echo "  make quality           - Run fmt, vet, and test"
	@echo ""
	@echo "Server Management:"
	@echo "  make run-servers       - Start legacy MCP servers in background (HTTP only)"
	@echo "  make run-servers-tls   - Start legacy MCP servers with TLS enabled (HTTP + HTTPS)"
	@echo "  make stop-servers      - Stop MCP servers"
	@echo ""
	@echo "MCP Streaming Protocol Demos:"
	@echo "  make demo-streaming    - Run MCP SDK demo (HTTP mode)"
	@echo "  make demo-streaming-tls - Run MCP SDK demo (mTLS mode)"
	@echo "  make generate-certs    - Generate TLS certificates for mTLS"
	@echo ""
	@echo "Legacy Demos:"
	@echo "  make demo              - Show legacy multi-agent usage examples"
	@echo ""
	@echo "Setup & Tools:"
	@echo "  make deps              - Download and tidy dependencies"
	@echo "  make dev-setup         - Complete development setup"
	@echo "  make install-tools     - Install optional development tools"
	@echo "  make help              - Show this help message"