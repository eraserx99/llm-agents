
## ☕ Java MCP Servers

The project now includes **Java implementations** of all MCP servers, providing **protocol-compatible alternatives** to the Go servers. The Java servers use the **official MCP Java SDK** and can be used **interchangeably** with Go servers without any changes to the coordinator agent.

### Why Java Servers?

- **Protocol Compatibility**: Identical JSON-RPC 2.0 / MCP Streaming Protocol implementation
- **Drop-in Replacement**: Same ports, same endpoints, same data formats
- **Multi-Language Demo**: Shows MCP protocol works across different tech stacks
- **Production Alternative**: Choose Go for performance or Java for ecosystem integration

### Java Implementation Stack

- **Language**: Java 21 (LTS)
- **Build System**: Gradle 8.12 with Application Plugin
- **MCP SDK**: Official `io.modelcontextprotocol.sdk:mcp:0.14.1`
- **HTTP Server**: Jetty 11.0.24 with Servlet API
- **TLS/Certificates**: BouncyCastle for PEM loading (reuses Go-generated certificates)
- **JSON**: Jackson with MCP-compatible serialization
- **CLI**: Picocli for command-line argument parsing
- **Testing**: JUnit 5 + AssertJ with 27 passing tests

### Quick Start - Java Servers

**All commands should be run from the project root directory (`llm-agents/`), not from `java-mcp-servers/`.**

```bash
# Build Java servers
make build-java

# Run Java servers (HTTP mode)
make run-java-servers

# Run Java servers with TLS
make run-java-servers-tls

# Stop Java servers
make stop-java-servers

# Run individual server via Gradle
make run-java-weather
make run-java-weather-tls  # With TLS
```

## 🔄 Switching Between Go and Java MCP Servers

This project provides **two implementations** of MCP servers:
- **Go servers**: High-performance, compiled binaries
- **Java servers**: JVM-based, enterprise-ready

Both implementations are **100% protocol-compatible** and can be used **interchangeably** without any code changes to the coordinator agent.

### Choose Your Implementation

#### Option 1: Go Servers (Default)

```bash
# 1. Build Go servers
make build

# 2. Start Go MCP servers
make run-servers              # HTTP mode
# OR
make run-servers-tls          # HTTPS mode with mTLS

# 3. Run queries
./bin/llm-agents -city "Tokyo" -query "What's the temperature?"

# 4. Stop servers
make stop-servers
```

#### Option 2: Java Servers

```bash
# 1. Build Java servers (requires Java 21+)
make build-java

# 2. Start Java MCP servers
make run-java-servers         # HTTP mode
# OR
make run-java-servers-tls     # HTTPS mode with mTLS

# 3. Run queries (same command!)
./bin/llm-agents -city "Tokyo" -query "What's the temperature?"

# 4. Stop servers
make stop-java-servers
```

### Seamless Switching

**Switch from Go to Java:**
```bash
make stop-servers           # Stop Go servers
make run-java-servers       # Start Java servers
# No other changes needed!
```

**Switch from Java to Go:**
```bash
make stop-java-servers      # Stop Java servers
make run-servers            # Start Go servers
# No other changes needed!
```

### Side-by-Side Comparison

| Feature | Go Servers | Java Servers |
|---------|-----------|--------------|
| **Build Time** | <5 seconds | ~10 seconds |
| **Binary Size** | ~15MB each | ~15MB JAR each |
| **Memory Usage** | ~10-20MB | ~64-128MB |
| **Startup Time** | <100ms | ~2-3 seconds |
| **Protocol** | MCP Streaming (JSON-RPC 2.0) | MCP Streaming (JSON-RPC 2.0) |
| **TLS Support** | ✅ mTLS | ✅ mTLS |
| **Ports** | 8081-8083 (HTTP)<br>8443-8445 (HTTPS) | 8081-8083 (HTTP)<br>8443-8445 (HTTPS) |
| **Compatibility** | 100% | 100% |

### Running Both Simultaneously

Use different ports to run Go and Java servers at the same time:

```bash
# Terminal 1: Go servers on default ports (8081-8083)
make run-servers

# Terminal 2: Java servers on alternate ports (9081-9083)
export WEATHER_MCP_PORT=9081
export DATETIME_MCP_PORT=9082
export ECHO_MCP_PORT=9083
make run-java-servers

# Test Go weather server
curl http://localhost:8081/health

# Test Java weather server
curl http://localhost:9081/health

# Configure coordinator to use Java servers
export MCP_WEATHER_URL=http://localhost:9081/mcp
export MCP_DATETIME_URL=http://localhost:9082/mcp
export MCP_ECHO_URL=http://localhost:9083/mcp
./bin/llm-agents -city "London" -query "temperature"
```

### All Available Make Targets

**Go Servers:**
```bash
make build                # Build Go binaries
make run-servers          # Start Go servers (HTTP)
make run-servers-tls      # Start Go servers (HTTPS/mTLS)
make stop-servers         # Stop Go servers
make test                 # Run Go tests
make clean                # Clean Go build artifacts
```

**Java Servers:**
```bash
make build-java           # Build Java JARs
make run-java-servers     # Start Java servers (HTTP)
make run-java-servers-tls # Start Java servers (HTTPS/mTLS)
make stop-java-servers    # Stop Java servers
make test-java            # Run Java tests
make clean-java           # Clean Java build artifacts
```

**Individual Java Servers (via Gradle):**
```bash
make run-java-weather         # Weather server (HTTP)
make run-java-weather-tls     # Weather server (HTTPS)
make run-java-datetime        # DateTime server (HTTP)
make run-java-datetime-tls    # DateTime server (HTTPS)
make run-java-echo            # Echo server (HTTP)
make run-java-echo-tls        # Echo server (HTTPS)
```

**Certificates (for TLS mode):**
```bash
make generate-certs       # Generate mTLS certificates (works for both Go and Java)
```

---

### Running Java Servers Directly

```bash
# Via JAR files (from project root)
java -jar java-mcp-servers/build/libs/weather-mcp-server-1.0.0.jar
java -jar java-mcp-servers/build/libs/datetime-mcp-server-1.0.0.jar --tls
java -jar java-mcp-servers/build/libs/echo-mcp-server-1.0.0.jar --verbose

# Via Gradle (from project root, supports --args for CLI parameters)
cd java-mcp-servers && ./gradlew runWeatherServer --args='--tls --verbose'
cd java-mcp-servers && ./gradlew runDateTimeServer
cd java-mcp-servers && ./gradlew runEchoServer --args='--tls'
```

### Java Server Ports

Same ports as Go servers for seamless switching:

| Server   | HTTP Port | HTTPS Port (TLS) |
|----------|-----------|------------------|
| Weather  | 8081      | 8443             |
| DateTime | 8082      | 8444             |
| Echo     | 8083      | 8445             |

### Switching Between Go and Java

**All make commands must be run from the project root directory (`llm-agents/`).**

**Stop Go servers and start Java servers:**
```bash
make stop-servers        # Stop Go servers
make run-java-servers    # Start Java servers

# Coordinator agent works without any code changes!
./bin/llm-agents -city "Tokyo" -query "What's the temperature?"
```

**Running Both Simultaneously (Different Ports):**
```bash
# From project root directory
# Go servers on default ports
make run-servers

# Java servers on alternate ports
export WEATHER_MCP_PORT=9081
export DATETIME_MCP_PORT=9082
export ECHO_MCP_PORT=9083
make run-java-servers
```

### Java Server Environment Variables

Same environment variables as Go servers:

```bash
export TLS_ENABLED=true            # Enable TLS mode
export TLS_DEMO_MODE=true          # Relaxed certificate validation
export TLS_CERT_DIR=./certs        # Certificate directory
export WEATHER_MCP_PORT=8081       # HTTP port (default: 8081)
export WEATHER_MCP_TLS_PORT=8443   # HTTPS port (default: 8443)
```

### Java Development

**All commands should be run from the project root directory using make, or from java-mcp-servers/ using gradlew.**

```bash
# From project root (recommended)
make build-java              # Build all Java servers
make test-java               # Run all Java tests
make clean-java              # Clean Java build artifacts

# Or using Gradle directly from java-mcp-servers/
cd java-mcp-servers
./gradlew build              # Full build with tests
./gradlew test               # Run tests only
./gradlew buildAllJars       # Build standalone JARs
./gradlew checkstyleMain     # Run checkstyle
./gradlew clean              # Clean build artifacts
```

### Java Server Architecture

```
java-mcp-servers/
├── src/main/java/com/llmagents/mcp/
│   ├── common/              # Shared utilities
│   │   ├── TLSConfig.java           # TLS configuration
│   │   ├── PEMCertificateLoader    # BouncyCastle PEM loader
│   │   ├── SSLContextFactory        # mTLS setup
│   │   ├── ServerConfig.java        # Server configuration
│   │   ├── JsonConfig.java          # Jackson ObjectMapper
│   │   └── protocol/                # JSON-RPC 2.0 models
│   ├── transport/
│   │   └── MCPServlet.java          # HTTP/SSE servlet handler
│   ├── weather/
│   │   ├── WeatherData.java         # Data model
│   │   ├── WeatherTool.java         # Tool handler
│   │   └── WeatherServer.java       # Main class
│   ├── datetime/
│   │   ├── DateTimeData.java
│   │   ├── DateTimeTool.java
│   │   └── DateTimeServer.java
│   └── echo/
│       ├── EchoData.java
│       ├── EchoTool.java
│       └── EchoServer.java
└── src/test/java/             # 27 passing tests
```

### IDE Support

**VS Code**: Launch configurations available in `.vscode/launch.json`
- Java Weather Server
- Java Weather Server (TLS)
- Java DateTime Server
- Java DateTime Server (TLS)
- Java Echo Server
- Java Echo Server (TLS)

Press `F5` in VS Code to run servers with debugging support.

### Testing Java Servers

```bash
# All tests from project root (recommended)
make test-java

# Or specific test suites using Gradle from java-mcp-servers/
cd java-mcp-servers
./gradlew test --tests "*WeatherContractTest"
./gradlew test --tests "*DateTimeContractTest"
./gradlew test --tests "*EchoContractTest"
./gradlew test --tests "*JsonRpcProtocolTest"
```

**Test Coverage**: 27 tests covering:
- Data model JSON serialization
- Tool handler request/response
- JSON-RPC 2.0 protocol compliance
- Error handling (malformed JSON, invalid params)
- Edge cases (empty strings, unsupported cities)

