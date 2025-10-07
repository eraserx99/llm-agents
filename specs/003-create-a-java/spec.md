# Feature Specification: Java MCP Servers

**Feature Branch**: `003-create-a-java`
**Created**: 2025-10-07
**Status**: Draft
**Input**: User description: "create a Java version of each of the MCP servers with exactly the same features and input/output data & format so that I can connect the agents to Java-based MCP servers without changing any of the agents code. please stick to Java 21 or greater, Gradle, and the latest version of MCP SDK for Jave, https://github.com/modelcontextprotocol/java-sdk. make sure you will update or extend Makefile, README.md, and Visual Code Studio configirations"

## Execution Flow (main)
```
1. Parse user description from Input
   � Feature description provided: Java MCP servers compatible with existing Go agents
2. Extract key concepts from description
   � Actors: Go coordinator agent (existing), Java MCP servers (new)
   � Actions: Create Java equivalents, maintain API compatibility, configure build system
   � Data: Weather, DateTime, Echo data with identical JSON format
   � Constraints: Java 21+, Gradle, official MCP Java SDK, no agent code changes
3. For each unclear aspect:
   � All requirements are clear from user description
4. Fill User Scenarios & Testing section
   � User flow: Replace Go servers with Java servers without changing agents
5. Generate Functional Requirements
   � Each requirement is testable through integration tests
6. Identify Key Entities
   � Weather data, DateTime data, Echo data entities
7. Run Review Checklist
   � Spec focused on WHAT, not HOW
8. Return: SUCCESS (spec ready for planning)
```

---

## � Quick Guidelines
-  Focus on WHAT users need and WHY
- L Avoid HOW to implement (no tech stack, APIs, code structure)
- =e Written for business stakeholders, not developers

---

## Clarifications

### Session 2025-10-07
- Q: What happens when a Java server is started with `--tls` flag but required certificates don't exist in the `TLS_CERT_DIR`? → A: Server fails immediately on startup with clear error message listing missing certificate files
- Q: How does the system handle port conflicts when Java servers attempt to bind to same ports as Go servers? → A: Server fails immediately on startup with error message indicating port already in use and suggesting alternate port configuration
- Q: What happens when the coordinator agent sends malformed JSON to a Java server? → A: Server returns JSON-RPC error response with code -32700 (Parse error) and descriptive error message, maintains connection
- Q: How does the Java datetime server respond when queried for a city name that is not in the supported timezone mappings? → A: Server returns JSON-RPC error with code -32602 (Invalid params) indicating unsupported city
- Q: What happens when a Java server process terminates unexpectedly while an active query is being processed? → A: Coordinator agent receives connection error, retries request once, then returns error to user if still failed
- Additional requirement: Gradle must support running MCP servers with command-line parameters (e.g., `./gradlew runWeatherServer --args='--tls'`), with corresponding Makefile targets, README documentation, and VS Code launch configurations

---

## User Scenarios & Testing *(mandatory)*

### Primary User Story
A developer working with the LLM multi-agent system wants to replace the existing Go-based MCP servers (weather, datetime, echo) with Java-based equivalents. The replacement must be seamless - the coordinator agent and all sub-agents must continue to work without any code modifications. The developer switches between Go and Java servers by simply stopping one set and starting the other, with all queries producing identical results.

### Acceptance Scenarios
1. **Given** Go coordinator agent is running with Java weather MCP server, **When** user queries "What's the temperature in New York?", **Then** system returns temperature data in the same format as Go weather server
2. **Given** Go coordinator agent is running with Java datetime MCP server, **When** user queries "What time is it in Los Angeles?", **Then** system returns datetime information in the same format as Go datetime server
3. **Given** Go coordinator agent is running with Java echo MCP server, **When** user queries "echo hello world", **Then** system echoes the text back in the same format as Go echo server
4. **Given** all three Java MCP servers are running, **When** user queries "What's the weather and time in Chicago?", **Then** system invokes both Java servers in parallel and combines results successfully
5. **Given** Java MCP servers with TLS enabled, **When** coordinator agent connects using mTLS, **Then** mutual TLS authentication succeeds and queries work as expected
6. **Given** developer runs build command, **When** Gradle build completes, **Then** all three Java server executables are created in expected locations
7. **Given** developer opens project in VS Code, **When** viewing Java server code, **Then** IDE provides proper Java syntax highlighting, code completion, and debugging support
8. **Given** developer wants to run Java weather server with TLS enabled, **When** developer executes `./gradlew runWeatherServer --args='--tls'`, **Then** server starts successfully in HTTPS mode on port 8443
9. **Given** developer wants convenient server startup, **When** developer executes `make run-java-weather-tls`, **Then** Makefile invokes Gradle with correct arguments and server starts in TLS mode

### Edge Cases
- **Missing TLS Certificates**: When Java server starts with `--tls` flag but required certificates (ca.crt, server.crt, server.key) don't exist in TLS_CERT_DIR, server MUST fail immediately on startup with clear error message listing specific missing certificate files
- **Port Conflicts**: When Java server attempts to bind to port already in use (e.g., Go server running on same port), server MUST fail immediately on startup with error message indicating which port is unavailable and suggesting environment variable configuration to use alternate port
- **Malformed JSON Requests**: When coordinator agent sends malformed JSON (invalid syntax, missing required JSON-RPC fields), server MUST return JSON-RPC error response with code -32700 (Parse error) and descriptive error message, while maintaining the connection for subsequent valid requests
- **Unsupported City Names**: When datetime server receives query for city not in supported timezone mappings (New York, Los Angeles, Chicago, Denver, London, Tokyo, etc.), server MUST return JSON-RPC error response with code -32602 (Invalid params) indicating the city is not supported
- **Server Crash During Query**: When Java server process terminates unexpectedly while processing active query, coordinator agent MUST detect connection error, retry request once automatically, and if retry fails, return error message to user indicating server unavailability

## Requirements *(mandatory)*

### Functional Requirements
- **FR-001**: Java weather MCP server MUST accept getTemperature requests with city parameter and return temperature, unit, description, city, and timestamp fields in same JSON format as Go weather server
- **FR-002**: Java datetime MCP server MUST accept getDateTime requests with city parameter and return local_time, timezone, utc_offset, city, and timestamp fields in same JSON format as Go datetime server
- **FR-003**: Java echo MCP server MUST accept echo requests with text parameter and return original_text, echo_text, and timestamp fields in same JSON format as Go echo server
- **FR-004**: Java MCP servers MUST implement MCP Streaming Protocol using official MCP Java SDK with StreamableHTTPHandler equivalent
- **FR-005**: Java MCP servers MUST expose /mcp endpoint for all MCP protocol communication (JSON-RPC 2.0 over HTTP/SSE)
- **FR-006**: Java MCP servers MUST support optional mTLS authentication using same certificate files as Go servers (ca.crt, server.crt, server.key, client.crt)
- **FR-007**: Java MCP servers MUST run on same default ports as Go servers (weather: 8081/8443, datetime: 8082/8444, echo: 8083/8445 for HTTP/HTTPS)
- **FR-008**: Java MCP servers MUST support command-line flag for enabling TLS mode (equivalent to Go's --tls flag)
- **FR-009**: Java MCP servers MUST read same environment variables as Go servers (TLS_ENABLED, TLS_DEMO_MODE, TLS_CERT_DIR, port configurations)
- **FR-010**: Build system MUST support "make build-java" command to compile all Java MCP servers
- **FR-011**: Build system MUST support "make run-java-servers" and "make run-java-servers-tls" commands to start Java servers
- **FR-012**: README MUST document how to build, configure, and run Java MCP servers with examples
- **FR-013**: VS Code workspace MUST include Java language server configuration and debugging launch configurations for Java servers
- **FR-026**: Gradle build MUST support running individual MCP servers with command-line parameters using application plugin (e.g., `./gradlew runWeatherServer --args='--tls --verbose'`)
- **FR-027**: Makefile MUST provide wrapper targets for Gradle run commands (e.g., `make run-java-weather-tls` maps to `./gradlew runWeatherServer --args='--tls'`)
- **FR-028**: VS Code launch configurations MUST include Gradle-based run configurations with support for passing command-line arguments to Java servers
- **FR-014**: Java weather server MUST return simulated weather data with same temperature range and condition options as Go version
- **FR-015**: Java datetime server MUST support same city timezone mappings as Go version (New York, Los Angeles, Chicago, Denver, London, Tokyo, etc.)
- **FR-016**: Java MCP servers MUST log requests and responses at same verbosity levels as Go servers (support for verbose logging flag)
- **FR-017**: Java MCP servers MUST implement JSON-RPC 2.0 protocol exactly as specified in MCP Streaming Protocol specification
- **FR-018**: Java MCP servers MUST handle connection lifecycle (initialization, tool discovery, tool execution, cleanup) identically to Go servers
- **FR-019**: Java MCP servers MUST return error responses in same JSON-RPC error format when queries fail
- **FR-020**: Build artifacts MUST be organized in consistent directory structure (bin/ or build/ for executables, matching project conventions)
- **FR-021**: Java MCP servers MUST fail immediately on startup when TLS mode enabled but required certificate files missing, displaying clear error message listing each missing file (ca.crt, server.crt, server.key)
- **FR-022**: Java MCP servers MUST fail immediately on startup when unable to bind to configured port, displaying error message with port number and suggesting environment variable to configure alternate port
- **FR-023**: Java MCP servers MUST return JSON-RPC error response with code -32700 (Parse error) when receiving malformed JSON, include descriptive error message, and maintain connection for subsequent requests
- **FR-024**: Java datetime server MUST return JSON-RPC error response with code -32602 (Invalid params) when queried for city not in supported timezone mappings, with error message indicating unsupported city name
- **FR-025**: Go coordinator agent MUST implement retry logic that detects connection errors to Java servers, automatically retries failed request once, and returns descriptive error to user if retry also fails

### Key Entities *(include if feature involves data)*
- **Weather Data**: Represents current weather conditions for a city with temperature (floating point number), unit (string: "�C"), description (string: weather condition), city (string: city name), timestamp (string: ISO 8601 format)
- **DateTime Data**: Represents current date and time for a city with local_time (string: formatted local time), timezone (string: timezone name), utc_offset (string: UTC offset format), city (string: city name), timestamp (string: ISO 8601 format)
- **Echo Data**: Represents echoed text with original_text (string: input text), echo_text (string: echoed output), timestamp (string: ISO 8601 format)
- **MCP Tool Definition**: Describes available tools with name (string: tool identifier), description (string: tool purpose), input schema (JSON schema: parameter definitions)
- **MCP Request**: JSON-RPC 2.0 request with jsonrpc (string: "2.0"), method (string: "tools/list" or "tools/call"), params (object: method parameters), id (number/string: request identifier)
- **MCP Response**: JSON-RPC 2.0 response with jsonrpc (string: "2.0"), result (object: response data) or error (object: error details), id (number/string: matching request identifier)

---

## Review & Acceptance Checklist
*GATE: Automated checks run during main() execution*

### Content Quality
- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness
- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status
*Updated by main() during processing*

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked (none found)
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed

---
