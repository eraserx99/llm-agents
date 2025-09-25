# CLI Interface Contract

## Command Line Arguments

### Basic Usage
```bash
llm-agents [OPTIONS] "query"
```

### Options
- `-city string`: Specify the city name (overrides city detection in query)
- `-help`: Display usage information
- `-version`: Display version information
- `-timeout duration`: Set query timeout (default: 5s)
- `-verbose`: Enable verbose output with timing information

### Examples

#### Temperature Query
```bash
llm-agents "What is the temperature in New York City right now?"
# Output: The current temperature in New York City is 72.5°F with partly cloudy conditions.

llm-agents -city "Los Angeles" "What's the temperature?"
# Output: The current temperature in Los Angeles is 68.0°F with clear skies.
```

#### DateTime Query
```bash
llm-agents "What time is it in Chicago?"
# Output: The current time in Chicago is 2:30 PM CST (September 23, 2025).

llm-agents -city "Seattle" "What's the datetime?"
# Output: The current time in Seattle is 12:30 PM PST (September 23, 2025).
```

#### Combined Query
```bash
llm-agents "What is the datetime and temperature of Miami now?"
# Output: In Miami, it is currently 3:30 PM EST (September 23, 2025) with a temperature of 85.0°F and sunny conditions.
```

#### Echo Query
```bash
llm-agents "Please echo my sentence: hello world!"
# Output: hello world!

llm-agents "Echo this text: The weather is nice today"
# Output: The weather is nice today
```

#### Error Cases
```bash
llm-agents "What is the temperature in InvalidCity?"
# Output: Error: City "InvalidCity" not found. Please specify a valid US city.

llm-agents "What is the weather?"
# Output: Error: No city specified. Please include a city name in your query or use the -city flag.
```

### Exit Codes
- `0`: Success
- `1`: Invalid arguments or flags
- `2`: City not found
- `3`: Service unavailable (MCP server connection failed)
- `4`: Timeout
- `5`: Other runtime error

### Input Validation
- Query text: Required, max 500 characters
- City flag: Optional, must match US city list
- Timeout: Optional, must be positive duration (e.g., "10s", "1m")

### Output Format

#### Standard Output
Natural language response on stdout:
```
The current temperature in [City] is [Temp]°F with [conditions].
The current time in [City] is [Time] [TZ] ([Date]).
In [City], it is currently [Time] [TZ] ([Date]) with a temperature of [Temp]°F and [conditions].
```

#### Verbose Output
When `-verbose` flag is used:
```
[TIMESTAMP] Starting query processing...
[TIMESTAMP] LLM Coordinator analyzing query: "What is the datetime and temperature of NYC?"
[TIMESTAMP] Detected city: New York City
[TIMESTAMP] Query analysis: Requires datetime AND temperature data
[TIMESTAMP] Orchestration decision: PARALLEL execution (independent data sources)
[TIMESTAMP] Selected agents: [datetime, temperature]
[TIMESTAMP] Dispatching to datetime agent...
[TIMESTAMP] Dispatching to temperature agent...
[TIMESTAMP] DateTime data received (took 95ms)
[TIMESTAMP] Temperature data received (took 120ms)
[TIMESTAMP] Invoked agents: [datetime, temperature]
[TIMESTAMP] Total processing time: 125ms

The current time in New York City is 2:30 PM EST with a temperature of 72.5°F and partly cloudy conditions.
```

#### Echo Query Verbose Output
```
[TIMESTAMP] Starting query processing...
[TIMESTAMP] LLM Coordinator analyzing query: "Please echo: hello world"
[TIMESTAMP] Query analysis: Echo request detected
[TIMESTAMP] Orchestration decision: SEQUENTIAL execution (single agent)
[TIMESTAMP] Selected agents: [echo]
[TIMESTAMP] Dispatching to echo agent...
[TIMESTAMP] Echo data received (took 15ms)
[TIMESTAMP] Invoked agents: [echo]
[TIMESTAMP] Total processing time: 18ms

hello world
```

#### Error Output
Errors written to stderr:
```
Error: [error message]
```

### Environment Variables
- `OPENROUTER_API_KEY`: Required for Claude 3.5 Sonnet access
- `MCP_WEATHER_URL`: Override default weather MCP server URL (default: http://localhost:8081)
- `MCP_DATETIME_URL`: Override default datetime MCP server URL (default: http://localhost:8082)
- `MCP_ECHO_URL`: Override default echo MCP server URL (default: http://localhost:8083)
- `LOG_LEVEL`: Set logging level (DEBUG, INFO, WARN, ERROR)