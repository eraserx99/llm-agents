# Tasks: Java MCP Servers

**Feature**: Java MCP Servers (003-create-a-java)
**Input**: Design documents from `/specs/003-create-a-java/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → SUCCESS: Loaded with tech stack (Java 21, Gradle, MCP SDK v0.14.1)
2. Load optional design documents:
   → data-model.md: 6 entities (WeatherData, DateTimeData, EchoData, MCPRequest, MCPResponse, MCPError)
   → contracts/: 4 files (mcp-protocol, weather-api, datetime-api, echo-api)
   → research.md: 8 technical decisions (MCP SDK, transport, mTLS, build, JSON, logging, CLI, testing)
   → quickstart.md: Build/run scenarios with HTTP and TLS modes
3. Generate tasks by category:
   → Setup: Gradle project, dependencies, shared utilities
   → Tests: 4 contract tests, 6 integration tests (TDD)
   → Core: 6 data models, 3 tool implementations, 3 server assemblies
   → Integration: mTLS support, Gradle run tasks, Makefile targets
   → Polish: Unit tests, Go compatibility tests, documentation
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001-T035)
6. Generate dependency graph
7. Validate: All contracts have tests, all entities have models, tests before impl
8. Return: SUCCESS (35 tasks ready for execution)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Java module**: `java-mcp-servers/` (separate module alongside Go code)
- **Source**: `java-mcp-servers/src/main/java/com/llmagents/mcp/`
- **Tests**: `java-mcp-servers/src/test/java/com/llmagents/mcp/`
- **Resources**: `java-mcp-servers/src/main/resources/`

---

## Phase 3.1: Project Setup & Configuration

- [ ] **T001** Create Gradle project structure with settings.gradle and root build.gradle
  - Path: `java-mcp-servers/` (new directory)
  - Create directory structure: src/main/java, src/main/resources, src/test/java
  - Initialize Gradle wrapper (./gradlew)
  - Configure Java 21 toolchain

- [ ] **T002** Configure Gradle dependencies in build.gradle
  - Path: `java-mcp-servers/build.gradle`
  - Add MCP Java SDK v0.14.1 (io.github.modelcontextprotocol:mcp-server)
  - Add Jakarta Servlet API, Jetty embedded server
  - Add BouncyCastle for PEM certificate loading
  - Add Jackson for JSON serialization
  - Add Picocli for CLI parsing
  - Add SLF4J + Logback for logging
  - Add JUnit 5, AssertJ, WireMock for testing

- [ ] **T003** [P] Configure Gradle application plugin for running servers with command-line args
  - Path: `java-mcp-servers/build.gradle`
  - Add application plugin
  - Define tasks: runWeatherServer, runDateTimeServer, runEchoServer
  - Configure to support `--args='--tls --verbose'` parameter passing

- [ ] **T004** [P] Configure code quality tools (Checkstyle, SpotBugs)
  - Path: `java-mcp-servers/build.gradle`
  - Add checkstyle plugin with Java coding standards
  - Add spotbugs plugin for static analysis
  - Configure linting to run before tests

- [ ] **T005** [P] Create logging configuration
  - Path: `java-mcp-servers/src/main/resources/logback.xml`
  - Configure SLF4J patterns matching Go log format
  - Set up MDC for request IDs
  - Configure INFO level default, DEBUG when --verbose flag

---

## Phase 3.2: Shared Utilities & Common Code

- [ ] **T006** [P] Create TLS configuration loader
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/tls/TLSConfig.java`
  - Load environment variables: TLS_ENABLED, TLS_DEMO_MODE, TLS_CERT_DIR
  - Provide certificate file paths
  - Validate configuration

- [ ] **T007** [P] Create PEM certificate loader using BouncyCastle
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/tls/PEMCertificateLoader.java`
  - Load ca.crt, server.crt, server.key from PEM format
  - Convert to Java KeyStore and TrustStore
  - Handle missing certificate files with clear error messages (FR-021)

- [ ] **T008** [P] Create SSLContext factory for mTLS
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/tls/SSLContextFactory.java`
  - Configure KeyManagerFactory with server certificate
  - Configure TrustManagerFactory with CA certificate
  - Create SSLContext for Jetty server
  - Support demo mode (relaxed validation)

- [ ] **T009** [P] Create server configuration class
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/config/ServerConfig.java`
  - Read environment variables for ports (WEATHER_MCP_PORT, WEATHER_MCP_TLS_PORT, etc.)
  - Provide default ports (8081/8443, 8082/8444, 8083/8445)
  - Parse command-line flags with Picocli (--tls, --verbose)

- [ ] **T010** [P] Create Jackson ObjectMapper configurator
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/json/JacksonConfig.java`
  - Configure field ordering (@JsonPropertyOrder support)
  - Configure ISO 8601 timestamp format
  - Configure null omission
  - Ensure byte-exact JSON compatibility with Go

---

## Phase 3.3: Data Models (TDD - Models First)

- [ ] **T011** [P] Create WeatherData record with validation
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/weather/model/WeatherData.java`
  - Implement Java record with fields: temperature, unit, description, city, timestamp
  - Add @JsonPropertyOrder annotation
  - Validate in compact constructor (temperature range, unit="°C", valid descriptions)
  - Add factory method: create(double, String, String)

- [ ] **T012** [P] Create DateTimeData record with validation
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/datetime/model/DateTimeData.java`
  - Implement Java record with fields: localTime, timezone, utcOffset, city, timestamp
  - Add @JsonProperty for snake_case fields
  - Validate IANA timezone, UTC offset format
  - Add factory method: create(String city, String timezoneId)

- [ ] **T013** [P] Create EchoData record with validation
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/echo/model/EchoData.java`
  - Implement Java record with fields: originalText, echoText, timestamp
  - Add @JsonProperty for snake_case fields
  - Validate echoText equals originalText
  - Add factory method: create(String text)

- [ ] **T014** [P] Create MCP protocol models (MCPRequest, MCPResponse, MCPError)
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/model/MCPRequest.java`
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/model/MCPResponse.java`
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/model/MCPError.java`
  - Implement JSON-RPC 2.0 message structures
  - Add error code constants (-32700, -32600, -32601, -32602, -32603)
  - Add factory methods for success/error responses

---

## Phase 3.4: Contract Tests (TDD - MUST FAIL BEFORE IMPLEMENTATION)

**CRITICAL: These tests MUST be written and MUST FAIL before ANY tool implementation**

- [ ] **T015** [P] Contract test for weather-api.json
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/WeatherContractTest.java`
  - Test JSON-RPC request/response format for getTemperature tool
  - Verify response matches weather-api.json schema
  - Test with cities: New York, Tokyo
  - MUST FAIL initially (no server implementation yet)

- [ ] **T016** [P] Contract test for datetime-api.json
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/DateTimeContractTest.java`
  - Test JSON-RPC request/response format for getDateTime tool
  - Verify response matches datetime-api.json schema
  - Test with cities: New York, Los Angeles, London, Tokyo
  - Test unsupported city returns -32602 error (FR-024)
  - MUST FAIL initially (no server implementation yet)

- [ ] **T017** [P] Contract test for echo-api.json
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/EchoContractTest.java`
  - Test JSON-RPC request/response format for echo tool
  - Verify response matches echo-api.json schema
  - Test with text: "hello world", "", "Special chars 🌍"
  - MUST FAIL initially (no server implementation yet)

- [ ] **T018** [P] Contract test for mcp-protocol.json (general protocol)
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/MCPProtocolTest.java`
  - Test tools/list method returns correct tool definitions
  - Test malformed JSON returns -32700 error (FR-023)
  - Test invalid JSON-RPC returns -32600 error
  - Test unknown method returns -32601 error
  - MUST FAIL initially (no server implementation yet)

---

## Phase 3.5: Tool Implementations (Make Tests Pass)

- [ ] **T019** Create WeatherTool handler using MCP SDK
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/weather/WeatherTool.java`
  - Implement getTemperature tool using mcp.AddTool()
  - Generate simulated temperature (20.0-45.0°C range)
  - Select random weather condition (Sunny, Partly cloudy, Cloudy, Light rain, Clear)
  - Return WeatherData with current timestamp
  - Format response text: "Weather in {city}: {temp}°C, {description}"
  - MUST make T015 pass

- [ ] **T020** Create DateTimeTool handler with city timezone mapping
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/datetime/DateTimeTool.java`
  - Implement getDateTime tool using mcp.AddTool()
  - Map cities to IANA timezones (New York→America/New_York, etc.)
  - Calculate local time using ZonedDateTime
  - Format localTime as "yyyy-MM-dd HH:mm:ss"
  - Format utcOffset as "±HH:MM"
  - Return JSON-RPC error -32602 for unsupported cities (FR-024)
  - MUST make T016 pass

- [ ] **T021** Create EchoTool handler
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/echo/EchoTool.java`
  - Implement echo tool using mcp.AddTool()
  - Return EchoData with text echoed back
  - Format response text: "Echo: {text}"
  - MUST make T017 pass

---

## Phase 3.6: MCP Server Assembly with Streamable HTTP Transport

- [ ] **T022** Create MCPServlet for HTTP/SSE transport
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/common/transport/MCPServlet.java`
  - Extend HttpServlet
  - Handle POST requests to /mcp endpoint
  - Parse JSON-RPC 2.0 requests
  - Delegate to MCP SDK server
  - Return SSE responses (Content-Type: text/event-stream)
  - Handle malformed JSON with -32700 error (FR-023)
  - Use AsyncContext for non-blocking SSE
  - MUST make T018 pass

- [ ] **T023** Create WeatherServer main class
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/weather/WeatherServer.java`
  - Create McpServer using official SDK (mcp.NewServer)
  - Register WeatherTool using mcp.AddTool()
  - Configure Jetty embedded server on port 8081 (HTTP) and 8443 (HTTPS)
  - Register MCPServlet on /mcp endpoint
  - Parse command-line args with Picocli (--tls, --verbose)
  - Fail fast if port unavailable (FR-022)
  - Start HTTP server
  - If --tls flag: load certificates, fail fast if missing (FR-021), start HTTPS server
  - Log server startup with port info

- [ ] **T024** Create DateTimeServer main class
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/datetime/DateTimeServer.java`
  - Create McpServer using official SDK
  - Register DateTimeTool using mcp.AddTool()
  - Configure Jetty on ports 8082 (HTTP) and 8444 (HTTPS)
  - Same servlet registration, CLI parsing, TLS support as WeatherServer
  - Fail fast on port conflict (FR-022) or missing certificates (FR-021)

- [ ] **T025** Create EchoServer main class
  - Path: `java-mcp-servers/src/main/java/com/llmagents/mcp/echo/EchoServer.java`
  - Create McpServer using official SDK
  - Register EchoTool using mcp.AddTool()
  - Configure Jetty on ports 8083 (HTTP) and 8445 (HTTPS)
  - Same servlet registration, CLI parsing, TLS support as other servers
  - Fail fast on port conflict (FR-022) or missing certificates (FR-021)

---

## Phase 3.7: Integration Tests (Go Compatibility)

- [ ] **T026** [P] Integration test: Java weather server with Go coordinator
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/GoCompatibilityWeatherTest.java`
  - Start Java weather server on port 8081
  - Send same JSON-RPC request that Go coordinator would send
  - Compare JSON response byte-by-byte with expected Go format
  - Verify acceptance scenario 1 from spec (New York temperature query)

- [ ] **T027** [P] Integration test: Java datetime server with Go coordinator
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/GoCompatibilityDateTimeTest.java`
  - Start Java datetime server on port 8082
  - Compare response format with Go server
  - Test city aliases (NYC → New York)
  - Verify acceptance scenario 2 from spec (Los Angeles datetime query)

- [ ] **T028** [P] Integration test: Java echo server with Go coordinator
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/GoCompatibilityEchoTest.java`
  - Start Java echo server on port 8083
  - Test with various text inputs
  - Verify acceptance scenario 3 from spec (echo hello world)

- [ ] **T029** [P] Integration test: mTLS authentication
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/MTLSAuthTest.java`
  - Generate test certificates using Go cert-gen
  - Start Java servers with --tls flag
  - Connect using client certificate
  - Verify TLS handshake succeeds
  - Verify queries work over HTTPS
  - Verify acceptance scenario 5 from spec (mTLS connection)

- [ ] **T030** [P] Integration test: Error handling edge cases
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/integration/ErrorHandlingTest.java`
  - Test missing certificates with --tls → server fails fast (FR-021)
  - Test port conflict → server fails fast (FR-022)
  - Test malformed JSON → -32700 error (FR-023)
  - Test unsupported city → -32602 error (FR-024)

---

## Phase 3.8: Build System Integration

- [ ] **T031** Update Makefile with Java-specific targets
  - Path: `Makefile` (root of repository)
  - Add target: `build-java` → `cd java-mcp-servers && ./gradlew build`
  - Add target: `run-java-servers` → run all three servers via `java -jar`
  - Add target: `run-java-servers-tls` → run with TLS environment variables
  - Add target: `run-java-weather-tls` → `cd java-mcp-servers && ./gradlew runWeatherServer --args='--tls'` (FR-027)
  - Add target: `run-java-datetime-tls` → similar Gradle run command (FR-027)
  - Add target: `run-java-echo-tls` → similar Gradle run command (FR-027)
  - Add target: `stop-java-servers` → `pkill -f 'mcp-server.jar'`
  - Add target: `test-java` → `cd java-mcp-servers && ./gradlew test`
  - Verify FR-026, FR-027 requirements satisfied

- [ ] **T032** Update VS Code launch configurations for Java servers
  - Path: `.vscode/launch.json`
  - Add configuration: "Launch Java Weather MCP Server (HTTP)" using Gradle run task
  - Add configuration: "Launch Java Weather MCP Server (TLS)" with --args='--tls'
  - Add configuration: "Launch Java DateTime MCP Server (HTTP)"
  - Add configuration: "Launch Java DateTime MCP Server (TLS)" with --args='--tls'
  - Add configuration: "Launch Java Echo MCP Server (HTTP)"
  - Add configuration: "Launch Java Echo MCP Server (TLS)" with --args='--tls'
  - Add compound: "All Java MCP Servers (HTTP)"
  - Add compound: "All Java MCP Servers (TLS)"
  - Verify FR-028 requirement satisfied

- [ ] **T033** Update README.md with Java server documentation
  - Path: `README.md`
  - Add "Java MCP Servers" section after Go servers section
  - Document prerequisites: Java 21+, Gradle
  - Document build commands: `make build-java`, `./gradlew build`
  - Document run commands: `make run-java-servers`, Gradle run tasks with --args
  - Document TLS mode setup (same certificates as Go)
  - Document examples: `./gradlew runWeatherServer --args='--tls --verbose'` (FR-026)
  - Add troubleshooting section for Java-specific issues
  - Verify FR-012 requirement satisfied

---

## Phase 3.9: Unit Tests & Polish

- [ ] **T034** [P] Unit tests for data model validation
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/unit/WeatherDataTest.java`
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/unit/DateTimeDataTest.java`
  - Path: `java-mcp-servers/src/test/java/com/llmagents/mcp/unit/EchoDataTest.java`
  - Test compact constructor validation rules
  - Test invalid inputs throw IllegalArgumentException
  - Use @ParameterizedTest for table-driven tests (Go standard T-1)

- [ ] **T035** [P] Update .gitignore for Java artifacts
  - Path: `.gitignore`
  - Add java-mcp-servers/build/
  - Add java-mcp-servers/.gradle/
  - Add *.jar exclusion exception for distribution
  - Add IDE-specific files (.idea/, *.iml)

---

## Dependencies

**Setup Phase**:
- T001 (project structure) must complete before all other tasks
- T002 (dependencies) must complete before any code tasks

**Shared Utilities**:
- T006-T010 can run in parallel [P]
- T011-T014 (data models) depend on T010 (Jackson config)
- T011-T014 can run in parallel [P] after T010

**Tests Before Implementation** (TDD):
- T015-T018 (contract tests) must be written before T019-T021 (tools)
- T015-T018 can run in parallel [P]
- T019 (WeatherTool) must make T015 pass
- T020 (DateTimeTool) must make T016 pass
- T021 (EchoTool) must make T017 pass
- T022 (MCPServlet) must make T018 pass

**Server Assembly**:
- T022 (servlet) must complete before T023-T025 (servers)
- T023-T025 can run in parallel [P] (different server classes)

**Integration Tests**:
- T026-T030 depend on T023-T025 (servers must exist)
- T026-T030 can run in parallel [P] (different test classes)

**Build System**:
- T031-T033 depend on T023-T025 (servers must be runnable)
- T031-T033 can run in parallel [P] (different files)

**Polish**:
- T034-T035 can run in parallel [P]
- T034-T035 should be last tasks

---

## Parallel Execution Examples

### Example 1: Shared Utilities (after T002)
```bash
# Launch T006-T010 together (different files):
Task: "Create TLS configuration loader in java-mcp-servers/src/main/java/com/llmagents/mcp/common/tls/TLSConfig.java"
Task: "Create PEM certificate loader in java-mcp-servers/src/main/java/com/llmagents/mcp/common/tls/PEMCertificateLoader.java"
Task: "Create SSLContext factory in java-mcp-servers/src/main/java/com/llmagents/mcp/common/tls/SSLContextFactory.java"
Task: "Create server configuration in java-mcp-servers/src/main/java/com/llmagents/mcp/common/config/ServerConfig.java"
Task: "Create Jackson configurator in java-mcp-servers/src/main/java/com/llmagents/mcp/common/json/JacksonConfig.java"
```

### Example 2: Data Models (after T010)
```bash
# Launch T011-T014 together (different files):
Task: "Create WeatherData record in java-mcp-servers/src/main/java/com/llmagents/mcp/weather/model/WeatherData.java"
Task: "Create DateTimeData record in java-mcp-servers/src/main/java/com/llmagents/mcp/datetime/model/DateTimeData.java"
Task: "Create EchoData record in java-mcp-servers/src/main/java/com/llmagents/mcp/echo/model/EchoData.java"
Task: "Create MCP protocol models in java-mcp-servers/src/main/java/com/llmagents/mcp/common/model/"
```

### Example 3: Contract Tests (after T014, before T019)
```bash
# Launch T015-T018 together (MUST FAIL initially):
Task: "Contract test for weather-api.json in WeatherContractTest.java"
Task: "Contract test for datetime-api.json in DateTimeContractTest.java"
Task: "Contract test for echo-api.json in EchoContractTest.java"
Task: "Contract test for mcp-protocol.json in MCPProtocolTest.java"
```

### Example 4: Server Main Classes (after T022)
```bash
# Launch T023-T025 together (different server classes):
Task: "Create WeatherServer main class with Jetty and MCP SDK"
Task: "Create DateTimeServer main class with Jetty and MCP SDK"
Task: "Create EchoServer main class with Jetty and MCP SDK"
```

### Example 5: Integration Tests (after T025)
```bash
# Launch T026-T030 together (different test classes):
Task: "Go compatibility test for weather server"
Task: "Go compatibility test for datetime server"
Task: "Go compatibility test for echo server"
Task: "mTLS authentication integration test"
Task: "Error handling edge cases integration test"
```

---

## Validation Checklist
*GATE: Checked before execution*

- [x] All contracts (4 files) have corresponding tests (T015-T018)
- [x] All entities (6 models) have model tasks (T011-T014)
- [x] All tests come before implementation (T015-T018 before T019-T025)
- [x] Parallel tasks [P] truly independent (different files, no shared state)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] TDD flow enforced: contract tests → tool implementations → integration tests
- [x] Gradle run tasks with --args support (FR-026)
- [x] Makefile wrapper targets for Gradle (FR-027)
- [x] VS Code launch configurations with args (FR-028)

---

## Notes

- **TDD Critical**: Tests T015-T018 MUST fail before tools T019-T021 are implemented
- **Parallel Optimization**: 22 tasks marked [P] can run concurrently (reduce wall-clock time by ~50%)
- **Certificate Reuse**: No Java cert-gen needed, reuse Go-generated certificates in ./certs/
- **Gradle Command-Line**: Application plugin supports `./gradlew runWeatherServer --args='--tls'` (FR-026)
- **Port Conflicts**: Developers can run Go and Java servers simultaneously on different ports
- **Commit Strategy**: Commit after each task, especially after making tests pass
- **Avoid**: Vague tasks, same file conflicts, implementation before tests

---

## Task Count Summary

- **Setup**: 5 tasks (T001-T005)
- **Shared Utilities**: 5 tasks (T006-T010)
- **Data Models**: 4 tasks (T011-T014)
- **Contract Tests**: 4 tasks (T015-T018) - TDD phase 1
- **Tool Implementations**: 3 tasks (T019-T021) - TDD phase 2
- **Server Assembly**: 4 tasks (T022-T025)
- **Integration Tests**: 5 tasks (T026-T030)
- **Build Integration**: 3 tasks (T031-T033)
- **Polish**: 2 tasks (T034-T035)

**Total**: 35 tasks (22 marked [P] for parallel execution)

---

*Generated: 2025-10-07*
*Ready for execution with TDD workflow and parallel optimization*
