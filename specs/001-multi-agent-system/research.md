# Research Report: Multi-Agent System Implementation

**Date**: 2025-09-23
**Feature**: Multi-Agent System for Temperature and DateTime Queries

## Executive Summary
Research completed for implementing a Go-based multi-agent system using specified libraries with free, non-API-key weather and datetime services via MCP protocol.

## Key Decisions

### 1. Agent Architecture Framework
**Decision**: Use github.com/nlpodyssey/openai-agents-go
**Rationale**:
- Specified in requirements
- Provides structured agent development framework
- Supports Claude 3.5 Sonnet via OpenRouter
- Built-in support for sub-agent coordination
**Alternatives considered**:
- Direct OpenAI API calls - rejected due to lack of agent coordination features
- Custom agent framework - rejected due to development overhead

### 2. MCP Implementation
**Decision**: Use github.com/modelcontextprotocol/go-sdk
**Rationale**:
- Specified in requirements
- Standard protocol for model-context communication
- Go SDK provides client and server implementations
- Enables structured data exchange between agents and external services
**Alternatives considered**:
- REST APIs - rejected in favor of MCP standard
- gRPC - rejected for simplicity with MCP

### 3. Free Weather Data Service
**Decision**: Implement MCP server wrapping free weather API (e.g., wttr.in or Open-Meteo)
**Rationale**:
- wttr.in provides free weather data without API keys
- Open-Meteo offers free tier with no authentication
- Can be wrapped in MCP server for protocol compliance
**Alternatives considered**:
- OpenWeatherMap - requires API key
- WeatherAPI - requires registration

### 4. Free DateTime Service
**Decision**: Use system time with timezone calculations via MCP server
**Rationale**:
- No external API needed for datetime
- Go's time package supports timezone operations
- Can calculate local time for US cities using timezone database
**Alternatives considered**:
- TimeZoneDB API - requires key
- WorldTimeAPI - rate limited

### 5. Orchestration Strategy
**Decision**: LLM-driven dynamic orchestration (sequential vs parallel)
**Rationale**:
- The coordinator agent (using Claude 3.5 Sonnet) analyzes the query and determines optimal execution strategy
- For independent data requests (e.g., "time AND temperature"), LLM decides to run parallel
- For dependent or single requests, LLM chooses sequential execution
- More flexible and intelligent than hard-coded parallel execution
- Allows for future expansion with more complex agent dependencies
**Implementation**:
- Coordinator agent returns orchestration plan with execution strategy
- Execution engine uses goroutines when parallel is specified
- Sequential execution when dependencies exist or parallel not beneficial
**Alternatives considered**:
- Always parallel - wastes resources for single queries
- Always sequential - slower for independent queries
- Hard-coded rules - less flexible than LLM decision

### 6. Error Handling Strategy
**Decision**: Partial failure tolerance with clear error messages
**Rationale**:
- If one sub-agent fails in combined query, return successful data with error note
- Maintains user experience when possible
- Clear communication of failures
**Alternatives considered**:
- Fail-fast on any error - poor user experience
- Silent failure - lacks transparency

### 7. City Name Resolution
**Decision**: Simple exact-match lookup with predefined US city list
**Rationale**:
- Meets requirement for no fuzzy matching
- Fast and deterministic
- Can maintain curated list of major US cities with coordinates
**Alternatives considered**:
- Geocoding API - requires API key
- Fuzzy matching - explicitly not required

### 8. Command Line Interface
**Decision**: Use flag package for CLI arguments
**Rationale**:
- Standard Go approach
- Simple and sufficient for requirements
- Supports city name as parameter
**Alternatives considered**:
- Cobra CLI framework - overhead for simple needs
- Interactive prompts - not required

## Technical Specifications

### MCP Server Specifications
- **Weather MCP Server**:
  - Port: 8081
  - Methods: GetTemperature(city string) → Temperature
  - Data source: wttr.in or Open-Meteo API

- **DateTime MCP Server**:
  - Port: 8082
  - Methods: GetDateTime(city string) → DateTime
  - Data source: System time with timezone calculations

- **Echo MCP Server**:
  - Port: 8083
  - Methods: Echo(text string) → EchoResponse
  - Data source: Simple text echo (no external API)

### Agent Communication Flow
1. User query → Coordinator Agent (Claude 3.5 Sonnet)
2. Coordinator analyzes query and determines:
   - Which sub-agents are needed (temperature, datetime, echo, or combination)
   - Execution strategy (parallel vs sequential)
   - Any dependencies between agents
3. Coordinator returns orchestration plan with agent selection reasoning
4. Execution engine implements the plan:
   - Dispatches only to selected sub-agents according to strategy
   - Sub-agents query respective MCP servers
   - Agent invocation logged for visibility
5. Results aggregated and formatted
6. Natural language response returned with agent invocation details

### Performance Considerations
- MCP servers run as separate processes
- Connection pooling for MCP clients
- Timeout handling (5 second default)
- Caching considered but not required for MVP

## Resolved Questions

1. **How to provide free weather data?**
   - Use wttr.in API (no key required) or Open-Meteo free tier

2. **How to handle datetime without external API?**
   - Use Go's time package with IANA timezone database

3. **How to coordinate parallel sub-agents?**
   - LLM-driven orchestration with dynamic execution strategy

4. **How to handle MCP protocol?**
   - Use provided go-sdk for both client and server implementations

5. **How to resolve city names to coordinates?**
   - Maintain static mapping of US cities to lat/lon and timezone

6. **How to determine which agents to invoke?**
   - LLM coordinator analyzes intent and selects only necessary agents
   - Echo agent invoked only for explicit echo requests
   - Weather/datetime agents not invoked for pure echo requests

7. **How to show agent selection transparency?**
   - Include InvokedAgents list in response
   - Add OrchestrationLog with step-by-step decisions
   - Verbose mode shows agent selection reasoning

## Next Steps
- Proceed to Phase 1: Design contracts and data models
- Define MCP message schemas
- Create failing tests for TDD approach