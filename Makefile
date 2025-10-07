# Makefile for LLM Multi-Agent System

.PHONY: build clean test lint fmt vet run-servers run-servers-tls stop-servers generate-certs query query-tls demo dev-setup install-tools help build-java run-java-servers run-java-servers-tls stop-java-servers test-java clean-java run-java-weather run-java-weather-tls run-java-datetime run-java-datetime-tls run-java-echo run-java-echo-tls

# Build all binaries
build:
	@echo "Building all components..."
	@mkdir -p bin
	go build -o bin/llm-agents ./cmd/main
	go build -o bin/weather-mcp ./cmd/weather-mcp
	go build -o bin/datetime-mcp ./cmd/datetime-mcp
	go build -o bin/echo-mcp ./cmd/echo-mcp
	go build -o bin/cert-gen ./cmd/cert-gen
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
	@echo "Starting MCP servers (HTTP mode)..."
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
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/weather-mcp --tls --verbose > weather-mcp-tls.log 2>&1 &
	@echo "Starting datetime MCP server (HTTP: 8082, HTTPS: 8444)..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/datetime-mcp --tls --verbose > datetime-mcp-tls.log 2>&1 &
	@echo "Starting echo MCP server (HTTP: 8083, HTTPS: 8445)..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && ./bin/echo-mcp --tls --verbose > echo-mcp-tls.log 2>&1 &
	@echo "All TLS servers started! Check *-tls.log files for output."
	@echo "Use 'make stop-servers' to stop them."

# Stop MCP servers
stop-servers:
	@echo "Stopping MCP servers..."
	@pkill -f weather-mcp || true
	@pkill -f datetime-mcp || true
	@pkill -f echo-mcp || true
	@echo "All servers stopped!"

# Generate TLS certificates for mTLS
generate-certs: build
	@echo "Generating TLS certificates..."
	@./bin/cert-gen
	@echo "Certificates generated in ./certs/ directory!"

# Run coordinator agent with TLS demo mode
query-tls:
	@if [ -z "$$OPENROUTER_API_KEY" ]; then \
		echo "Error: OPENROUTER_API_KEY environment variable is required"; \
		exit 1; \
	fi
	@echo "Running query with TLS demo mode enabled..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs \
		MCP_WEATHER_URL=https://localhost:8443/mcp \
		MCP_DATETIME_URL=https://localhost:8444/mcp \
		MCP_ECHO_URL=https://localhost:8445/mcp && \
		./bin/llm-agents $(ARGS)

# Run coordinator agent query (HTTP mode)
query:
	@if [ -z "$$OPENROUTER_API_KEY" ]; then \
		echo "Error: OPENROUTER_API_KEY environment variable is required"; \
		exit 1; \
	fi
	@./bin/llm-agents $(ARGS)

# Example usage demonstrations
demo: build
	@echo "=== LLM Multi-Agent System Demo ==="
	@echo ""
	@echo "First, make sure you have OPENROUTER_API_KEY set:"
	@echo "export OPENROUTER_API_KEY=\"your-key-here\""
	@echo ""
	@echo "Then start the MCP servers:"
	@echo "make run-servers       # HTTP mode"
	@echo "make run-servers-tls   # HTTPS mode with mTLS"
	@echo ""
	@echo "Example queries (HTTP mode):"
	@echo ""
	@echo "1. Temperature query:"
	@echo "make query ARGS='-city \"New York\" -query \"What is the temperature?\"'"
	@echo ""
	@echo "2. DateTime query:"
	@echo "make query ARGS='-city \"Los Angeles\" -query \"What time is it?\"'"
	@echo ""
	@echo "3. Combined query (parallel execution):"
	@echo "make query ARGS='-city \"Chicago\" -query \"What is the weather and time?\"'"
	@echo ""
	@echo "Example queries (TLS mode - requires run-servers-tls or run-java-servers-tls):"
	@echo ""
	@echo "1. Temperature query with TLS:"
	@echo "make query-tls ARGS='-city \"Tokyo\" -query \"What is the temperature?\"'"
	@echo ""
	@echo "2. Combined query with TLS:"
	@echo "make query-tls ARGS='-city \"Chicago\" -query \"What is the weather and time?\"'"
	@echo ""
	@echo "Or run directly:"
	@echo "export TLS_DEMO_MODE=true"
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
	@echo "  make build             - Build all components"
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
	@echo "  make run-servers       - Start MCP servers in background (HTTP mode)"
	@echo "  make run-servers-tls   - Start MCP servers with mTLS enabled"
	@echo "  make stop-servers      - Stop all MCP servers"
	@echo "  make generate-certs    - Generate TLS certificates for mTLS"
	@echo ""
	@echo "Queries:"
	@echo "  make query ARGS='...'  - Run coordinator query (HTTP mode)"
	@echo "  make query-tls ARGS='...' - Run coordinator query (TLS mode with demo mode)"
	@echo ""
	@echo "Demos:"
	@echo "  make demo              - Show multi-agent usage examples"
	@echo ""
	@echo "Setup & Tools:"
	@echo "  make deps              - Download and tidy dependencies"
	@echo "  make dev-setup         - Complete development setup"
	@echo "  make install-tools     - Install optional development tools"
	@echo "  make help              - Show this help message"
	@echo ""
	@echo "Java MCP Servers:"
	@echo "  make build-java        - Build Java MCP servers (requires Java 21+)"
	@echo "  make run-java-servers  - Start Java MCP servers (HTTP mode)"
	@echo "  make run-java-servers-tls - Start Java servers with mTLS"
	@echo "  make stop-java-servers - Stop Java MCP servers"
	@echo "  make test-java         - Run Java tests"
	@echo "  make clean-java        - Clean Java build artifacts"

# ========================================
# Java MCP Servers
# ========================================

# Build Java MCP servers
build-java:
	@echo "Building Java MCP servers..."
	cd java-mcp-servers && ./gradlew buildAllJars
	@echo "Java build complete! JARs are in ./java-mcp-servers/build/libs/"

# Run Java MCP servers (HTTP mode)
run-java-servers:
	@echo "Starting Java MCP servers (HTTP mode)..."
	@java -jar java-mcp-servers/build/libs/weather-mcp-server-1.0.0.jar > weather-mcp-java.log 2>&1 &
	@echo "Started Weather MCP Server (Java) on port 8081"
	@java -jar java-mcp-servers/build/libs/datetime-mcp-server-1.0.0.jar > datetime-mcp-java.log 2>&1 &
	@echo "Started DateTime MCP Server (Java) on port 8082"
	@java -jar java-mcp-servers/build/libs/echo-mcp-server-1.0.0.jar > echo-mcp-java.log 2>&1 &
	@echo "Started Echo MCP Server (Java) on port 8083"
	@echo "All Java servers started! Check *-java.log files for output."

# Run Java MCP servers with TLS
run-java-servers-tls: generate-certs
	@echo "Starting Java MCP servers with TLS enabled..."
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && \
		java -jar java-mcp-servers/build/libs/weather-mcp-server-1.0.0.jar --tls > weather-mcp-java-tls.log 2>&1 &
	@echo "Started Weather MCP Server (Java) with TLS on port 8443"
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && \
		java -jar java-mcp-servers/build/libs/datetime-mcp-server-1.0.0.jar --tls > datetime-mcp-java-tls.log 2>&1 &
	@echo "Started DateTime MCP Server (Java) with TLS on port 8444"
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && \
		java -jar java-mcp-servers/build/libs/echo-mcp-server-1.0.0.jar --tls > echo-mcp-java-tls.log 2>&1 &
	@echo "Started Echo MCP Server (Java) with TLS on port 8445"
	@echo "All Java TLS servers started! Check *-java-tls.log files for output."

# Run individual Java servers via Gradle (with CLI args support)
run-java-weather:
	cd java-mcp-servers && ./gradlew runWeatherServer

run-java-weather-tls: generate-certs
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && \
		cd java-mcp-servers && ./gradlew runWeatherServer --args='--tls'

run-java-datetime:
	cd java-mcp-servers && ./gradlew runDateTimeServer

run-java-datetime-tls: generate-certs
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && \
		cd java-mcp-servers && ./gradlew runDateTimeServer --args='--tls'

run-java-echo:
	cd java-mcp-servers && ./gradlew runEchoServer

run-java-echo-tls: generate-certs
	@export TLS_ENABLED=true TLS_DEMO_MODE=true TLS_CERT_DIR=./certs && \
		cd java-mcp-servers && ./gradlew runEchoServer --args='--tls'

# Stop Java MCP servers
stop-java-servers:
	@echo "Stopping Java MCP servers..."
	@pkill -f "weather-mcp-server" || true
	@pkill -f "datetime-mcp-server" || true
	@pkill -f "echo-mcp-server" || true
	@echo "Java servers stopped!"

# Test Java MCP servers
test-java:
	@echo "Running Java tests..."
	cd java-mcp-servers && ./gradlew test

# Clean Java build artifacts
clean-java:
	@echo "Cleaning Java build artifacts..."
	cd java-mcp-servers && ./gradlew clean
	@echo "Java clean complete!"