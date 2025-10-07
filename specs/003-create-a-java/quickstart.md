# Java MCP Servers Quickstart Guide

**Feature**: Java MCP Servers (003-create-a-java)
**Date**: 2025-10-07
**Purpose**: Quick setup guide for building and running Java MCP servers

## Prerequisites

Before starting, ensure you have:

- ✅ **Java 21 or greater** ([Download AdoptiumJDK](https://adoptium.net/))
  ```bash
  java -version  # Should show Java 21+
  ```
- ✅ **Git** (for cloning repository)
- ✅ **Make** (for build automation)
- ✅ **Certificates** (optional, for TLS mode)
  ```bash
  make generate-certs  # Generates certificates in ./certs/
  ```

## Quick Start (5 minutes)

### 1. Build Java Servers

```bash
# Build all Java servers using Makefile
make build-java

# OR use Gradle directly
cd java-mcp-servers
./gradlew build
cd ..

# Verify build outputs
ls -lh java-mcp-servers/build/libs/
# Expected: weather-mcp-server.jar, datetime-mcp-server.jar, echo-mcp-server.jar
```

**Build Output**:
```
java-mcp-servers/build/libs/
├── weather-mcp-server.jar    (~15 MB)
├── datetime-mcp-server.jar   (~15 MB)
└── echo-mcp-server.jar       (~15 MB)
```

### 2. Run Servers (HTTP Mode)

```bash
# Option A: Use Makefile (recommended)
make run-java-servers

# Option B: Run individually
java -jar java-mcp-servers/build/libs/weather-mcp-server.jar &
java -jar java-mcp-servers/build/libs/datetime-mcp-server.jar &
java -jar java-mcp-servers/build/libs/echo-mcp-server.jar &
```

**Expected Output**:
```
INFO  [weather-mcp] WeatherServer - Starting Weather MCP Server (HTTP) on :8081
INFO  [weather-mcp] WeatherServer - Weather MCP Server started with StreamableHTTPHandler
INFO  [datetime-mcp] DateTimeServer - Starting DateTime MCP Server (HTTP) on :8082
INFO  [datetime-mcp] DateTimeServer - DateTime MCP Server started with StreamableHTTPHandler
INFO  [echo-mcp] EchoServer - Starting Echo MCP Server (HTTP) on :8083
INFO  [echo-mcp] EchoServer - Echo MCP Server started with StreamableHTTPHandler
```

### 3. Test Connectivity

```bash
# Test with Go coordinator agent (best test - proves compatibility)
./bin/llm-agents -city "New York" -query "What's the temperature?"

# OR test directly with curl
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "getTemperature",
      "arguments": {"city": "Tokyo"}
    },
    "id": 1
  }'
```

**Expected Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Weather in Tokyo: 37.3°C, Light rain"
      }
    ]
  },
  "id": 1
}
```

### 4. Stop Servers

```bash
# Option A: Use Makefile
make stop-java-servers

# Option B: Kill by process name
pkill -f "weather-mcp-server.jar"
pkill -f "datetime-mcp-server.jar"
pkill -f "echo-mcp-server.jar"
```

## TLS Mode (mTLS Authentication)

### Prerequisites for TLS

1. **Generate Certificates** (if not already generated):
   ```bash
   make generate-certs
   # Creates certificates in ./certs/ directory
   ```

2. **Verify Certificates Exist**:
   ```bash
   ls -lh certs/
   # Expected files: ca.crt, ca.key, server.crt, server.key, client.crt, client.key
   ```

### Run Servers with TLS

```bash
# Set TLS environment variables
export TLS_ENABLED=true
export TLS_DEMO_MODE=true
export TLS_CERT_DIR=./certs

# Configure MCP server URLs for TLS ports (for Go coordinator agent)
export MCP_WEATHER_URL=https://localhost:8443/mcp
export MCP_DATETIME_URL=https://localhost:8444/mcp
export MCP_ECHO_URL=https://localhost:8445/mcp

# Option A: Use Makefile (recommended)
make run-java-servers-tls

# Option B: Run individually with --tls flag
java -jar java-mcp-servers/build/libs/weather-mcp-server.jar --tls &
java -jar java-mcp-servers/build/libs/datetime-mcp-server.jar --tls &
java -jar java-mcp-servers/build/libs/echo-mcp-server.jar --tls &
```

**Expected Output (TLS Mode)**:
```
INFO  [weather-mcp] WeatherServer - Weather MCP Server configured with TLS support
INFO  [weather-mcp] WeatherServer - HTTP port: 8081, HTTPS port: 8443
INFO  [weather-mcp] WeatherServer - TLS demo mode: true
INFO  [weather-mcp] WeatherServer - Certificate directory: ./certs
INFO  [weather-mcp] TLSLoader - Loaded CA certificate from ./certs/ca.crt
INFO  [weather-mcp] TLSLoader - Loaded server certificate from ./certs/server.crt
INFO  [weather-mcp] TLSLoader - Loaded server private key from ./certs/server.key
INFO  [weather-mcp] WeatherServer - Starting Weather MCP Server (HTTPS) on :8443
```

### Test TLS Connectivity

```bash
# Test with Go coordinator agent (best test)
./bin/llm-agents -city "Chicago" -query "What's the temperature?"

# OR test directly with curl (requires mTLS)
curl -X POST https://localhost:8443/mcp \
  --cacert certs/ca.crt \
  --cert certs/client.crt \
  --key certs/client.key \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "getTemperature",
      "arguments": {"city": "Chicago"}
    },
    "id": 1
  }'
```

## Server Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WEATHER_MCP_PORT` | 8081 | Weather server HTTP port |
| `WEATHER_MCP_TLS_PORT` | 8443 | Weather server HTTPS port |
| `DATETIME_MCP_PORT` | 8082 | DateTime server HTTP port |
| `DATETIME_MCP_TLS_PORT` | 8444 | DateTime server HTTPS port |
| `ECHO_MCP_PORT` | 8083 | Echo server HTTP port |
| `ECHO_MCP_TLS_PORT` | 8445 | Echo server HTTPS port |
| `TLS_ENABLED` | false | Enable TLS mode |
| `TLS_DEMO_MODE` | true | Use relaxed certificate validation |
| `TLS_CERT_DIR` | ./certs | Certificate directory path |

### Command-Line Flags

Each server supports:
- `--tls` : Enable TLS support (requires TLS_ENABLED=true environment variable)
- `--verbose` : Enable verbose logging (DEBUG level)
- `--help` : Show help message

**Example**:
```bash
java -jar java-mcp-servers/build/libs/weather-mcp-server.jar --tls --verbose
```

## Integration with Go Coordinator

### Seamless Server Switching

The Java servers are protocol-compatible with Go coordinator agents. You can switch between Go and Java implementations without any code changes:

1. **Stop Go servers**:
   ```bash
   make stop-servers
   ```

2. **Start Java servers**:
   ```bash
   make run-java-servers
   ```

3. **Run queries** (coordinator agent doesn't know the difference):
   ```bash
   ./bin/llm-agents -city "Boston" -query "What's the temperature and time?"
   ```

### Parallel Testing

You can run Go and Java servers on different ports simultaneously:

```bash
# Go servers on default ports (8081-8083)
make run-servers

# Java servers on alternate ports
export WEATHER_MCP_PORT=9081
export DATETIME_MCP_PORT=9082
export ECHO_MCP_PORT=9083
make run-java-servers

# Test Go weather server
curl http://localhost:8081/mcp -X POST -d '...'

# Test Java weather server
curl http://localhost:9081/mcp -X POST -d '...'
```

## Validation Checklist

After starting servers, verify:

- [ ] **Build succeeds** without errors
  ```bash
  make build-java
  # No errors, JARs created in build/libs/
  ```

- [ ] **Servers start** and bind to correct ports
  ```bash
  # Check server processes
  ps aux | grep "mcp-server.jar"

  # Check port bindings
  lsof -i :8081  # Weather HTTP
  lsof -i :8082  # DateTime HTTP
  lsof -i :8083  # Echo HTTP
  ```

- [ ] **Health check** (if implemented)
  ```bash
  curl http://localhost:8081/health
  # Expected: {"status":"ok","server":"weather-mcp","version":"1.0.0"}
  ```

- [ ] **Go coordinator** can connect and query
  ```bash
  ./bin/llm-agents -city "Miami" -query "temperature"
  # Should return temperature data
  ```

- [ ] **mTLS handshake** succeeds (TLS mode only)
  ```bash
  openssl s_client -connect localhost:8443 \
    -CAfile certs/ca.crt \
    -cert certs/client.crt \
    -key certs/client.key
  # Should show "Verify return code: 0 (ok)"
  ```

- [ ] **JSON responses** match Go server format exactly
  ```bash
  # Compare Go and Java responses
  curl http://localhost:8081/mcp -X POST -d '...' > go_response.json
  # (Switch to Java servers)
  curl http://localhost:8081/mcp -X POST -d '...' > java_response.json
  diff go_response.json java_response.json
  # Should show only timestamp differences
  ```

## Troubleshooting

### Build Failures

**Problem**: Gradle build fails with "Could not find MCP SDK"

**Solution**:
```bash
# Verify Gradle can access Maven Central
cd java-mcp-servers
./gradlew dependencies --refresh-dependencies
```

---

**Problem**: Java version error

**Solution**:
```bash
# Check Java version
java -version  # Must be 21+

# Set JAVA_HOME if needed
export JAVA_HOME=/path/to/jdk-21
```

### Runtime Failures

**Problem**: Port already in use

**Solution**:
```bash
# Find process using port
lsof -i :8081

# Kill process
kill -9 <PID>

# OR use alternate ports
export WEATHER_MCP_PORT=9081
java -jar weather-mcp-server.jar
```

---

**Problem**: Certificate not found (TLS mode)

**Solution**:
```bash
# Verify certificates exist
ls -lh certs/
# Should show: ca.crt, server.crt, server.key, client.crt, client.key

# Regenerate if missing
make generate-certs
```

---

**Problem**: mTLS handshake failure

**Solution**:
```bash
# Verify certificates are valid
openssl verify -CAfile certs/ca.crt certs/server.crt
# Should show: certs/server.crt: OK

# Check TLS environment variables
echo $TLS_ENABLED $TLS_DEMO_MODE $TLS_CERT_DIR
# Should show: true true ./certs

# Verify server is listening on HTTPS port
lsof -i :8443
```

---

**Problem**: Go coordinator can't connect to Java servers

**Solution**:
```bash
# Test Java server directly with curl
curl http://localhost:8081/mcp -X POST -d '{...}'
# If curl works but coordinator doesn't, check:

# 1. Coordinator configuration
echo $MCP_WEATHER_URL
# Should be: http://localhost:8081/mcp (or https://... for TLS)

# 2. Network connectivity
nc -zv localhost 8081
# Should show: Connection to localhost port 8081 [tcp/*] succeeded!

# 3. Server logs
# Check logs for connection errors
```

## Performance Considerations

### Memory Usage

Each Java server uses ~64-128 MB RAM (lighter than expected due to efficient MCP SDK):

```bash
# Monitor memory usage
ps aux | grep "mcp-server.jar"
# Check RSS (resident set size) column
```

### Startup Time

- **Cold start**: ~2-3 seconds (first run)
- **Warm start**: ~1-2 seconds (subsequent runs with Gradle daemon)

### Response Time

- **HTTP requests**: <10ms (simulated data)
- **HTTPS requests**: <20ms (includes TLS handshake overhead)

## Next Steps

Once servers are running and validated:

1. ✅ **Update Documentation**: Verify README.md reflects Java servers
2. ✅ **Update Makefile**: Add `build-java`, `run-java-servers`, `run-java-servers-tls` targets
3. ✅ **Update VS Code**: Add Java server debug configurations to `.vscode/launch.json`
4. ✅ **Run Integration Tests**: Execute compatibility tests with Go coordinator
5. ✅ **Performance Testing**: Benchmark response times vs Go servers
6. ✅ **Security Audit**: Verify mTLS implementation matches Go behavior

---

**Last Updated**: 2025-10-07
**Questions?** Check the main [README.md](../../../README.md) or feature spec [spec.md](./spec.md)
