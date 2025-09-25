# Makefile for LLM Multi-Agent System

.PHONY: build clean test lint fmt vet run-servers help

# Build all binaries
build:
	@echo "Building all components..."
	@mkdir -p bin
	go build -o bin/llm-agents ./cmd/main
	go build -o bin/weather-mcp ./cmd/weather-mcp
	go build -o bin/datetime-mcp ./cmd/datetime-mcp
	go build -o bin/echo-mcp ./cmd/echo-mcp
	@echo "Build complete! Binaries are in ./bin/"

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

# Start MCP servers in background
run-servers:
	@echo "Starting MCP servers..."
	@echo "Starting weather MCP server on port 8081..."
	@./bin/weather-mcp > weather-mcp.log 2>&1 &
	@echo "Starting datetime MCP server on port 8082..."
	@./bin/datetime-mcp > datetime-mcp.log 2>&1 &
	@echo "Starting echo MCP server on port 8083..."
	@./bin/echo-mcp > echo-mcp.log 2>&1 &
	@echo "All servers started! Check *.log files for output."
	@echo "Use 'make stop-servers' to stop them."

# Stop MCP servers
stop-servers:
	@echo "Stopping MCP servers..."
	@pkill -f weather-mcp || true
	@pkill -f datetime-mcp || true
	@pkill -f echo-mcp || true
	@echo "All servers stopped!"

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
	@echo "Usage:"
	@echo "  make build         - Build all binaries"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make test-race     - Run tests with race detection"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Vet code"
	@echo "  make lint          - Lint code (requires golangci-lint)"
	@echo "  make quality       - Run fmt, vet, and test"
	@echo "  make deps          - Download and tidy dependencies"
	@echo "  make run-servers   - Start MCP servers in background"
	@echo "  make stop-servers  - Stop MCP servers"
	@echo "  make demo          - Show usage examples"
	@echo "  make dev-setup     - Complete development setup"
	@echo "  make install-tools - Install optional development tools"
	@echo "  make help          - Show this help message"