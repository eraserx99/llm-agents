# Data Model: Java MCP Servers

**Feature**: Java MCP Servers (003-create-a-java)
**Date**: 2025-10-07
**Purpose**: Define data structures for Java MCP servers to match Go implementation

## Overview

This document defines the data models (Java records) for the three MCP servers. All models must produce JSON output identical to the Go implementation to ensure protocol compatibility with existing Go coordinator agents.

## Design Principles

1. **Immutability**: Use Java records (final fields, no setters)
2. **Validation**: Validate in compact constructors
3. **JSON Compatibility**: Field names/order match Go structs exactly
4. **Type Safety**: Use appropriate Java types (no `Object` where possible)
5. **ISO 8601 Timestamps**: All timestamps in RFC3339 format

## Entity Definitions

### 1. WeatherData

**Purpose**: Represents weather conditions for a city (simulated data)

**Go Equivalent**:
```go
type WeatherResult struct {
    Temperature float64 `json:"temperature"`
    Unit        string  `json:"unit"`
    Description string  `json:"description"`
    City        string  `json:"city"`
    Timestamp   string  `json:"timestamp"`
}
```

**Java Definition**:
```java
package com.llmagents.mcp.weather.model;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import java.time.Instant;

/**
 * Represents current weather conditions for a city.
 * Produces JSON compatible with Go WeatherResult struct.
 */
@JsonPropertyOrder({"temperature", "unit", "description", "city", "timestamp"})
public record WeatherData(
    double temperature,    // Temperature in Celsius (range: 20.0-45.0)
    String unit,           // Temperature unit (always "°C")
    String description,    // Weather condition
    String city,           // City name
    String timestamp       // ISO 8601 format (e.g., "2024-09-23T14:30:45Z")
) {
    // Validation in compact constructor
    public WeatherData {
        // Validate temperature range
        if (temperature < -50.0 || temperature > 60.0) {
            throw new IllegalArgumentException(
                "Temperature must be between -50°C and 60°C, got: " + temperature
            );
        }

        // Validate unit
        if (unit == null || !unit.equals("°C")) {
            throw new IllegalArgumentException(
                "Unit must be '°C', got: " + unit
            );
        }

        // Validate description
        if (description == null || description.isBlank()) {
            throw new IllegalArgumentException("Description cannot be null or blank");
        }

        // Validate description is one of allowed values
        var validDescriptions = java.util.Set.of(
            "Sunny", "Partly cloudy", "Cloudy", "Light rain", "Clear"
        );
        if (!validDescriptions.contains(description)) {
            throw new IllegalArgumentException(
                "Description must be one of: " + validDescriptions + ", got: " + description
            );
        }

        // Validate city
        if (city == null || city.isBlank()) {
            throw new IllegalArgumentException("City cannot be null or blank");
        }

        // Validate timestamp format
        if (timestamp == null) {
            throw new IllegalArgumentException("Timestamp cannot be null");
        }
        try {
            Instant.parse(timestamp); // Validate ISO 8601 format
        } catch (Exception e) {
            throw new IllegalArgumentException(
                "Timestamp must be ISO 8601 format, got: " + timestamp, e
            );
        }
    }

    /**
     * Creates WeatherData with current timestamp.
     */
    public static WeatherData create(
        double temperature,
        String description,
        String city
    ) {
        return new WeatherData(
            temperature,
            "°C",
            description,
            city,
            Instant.now().toString()
        );
    }
}
```

**Field Constraints**:
| Field | Type | Required | Constraints | Go Equivalent |
|-------|------|----------|-------------|---------------|
| temperature | double | Yes | -50.0 to 60.0 (simulated range: 20.0-45.0) | float64 |
| unit | String | Yes | Must be "°C" | string |
| description | String | Yes | One of: Sunny, Partly cloudy, Cloudy, Light rain, Clear | string |
| city | String | Yes | Non-blank | string |
| timestamp | String | Yes | ISO 8601 format (RFC3339) | string |

**Example JSON**:
```json
{
  "temperature": 37.3,
  "unit": "°C",
  "description": "Light rain",
  "city": "Tokyo",
  "timestamp": "2024-09-23T14:30:45Z"
}
```

### 2. DateTimeData

**Purpose**: Represents current date/time information for a city with timezone

**Go Equivalent**:
```go
type DateTimeResult struct {
    LocalTime   string `json:"local_time"`
    Timezone    string `json:"timezone"`
    UTCOffset   string `json:"utc_offset"`
    City        string `json:"city"`
    Timestamp   string `json:"timestamp"`
}
```

**Java Definition**:
```java
package com.llmagents.mcp.datetime.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import java.time.Instant;
import java.time.ZoneId;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;

/**
 * Represents current date/time information for a city.
 * Produces JSON compatible with Go DateTimeResult struct.
 */
@JsonPropertyOrder({"local_time", "timezone", "utc_offset", "city", "timestamp"})
public record DateTimeData(
    @JsonProperty("local_time") String localTime,   // Local time in "yyyy-MM-dd HH:mm:ss" format
    String timezone,                                  // IANA timezone (e.g., "America/New_York")
    @JsonProperty("utc_offset") String utcOffset,    // UTC offset (e.g., "-05:00")
    String city,                                      // City name
    String timestamp                                  // ISO 8601 format
) {
    // Validation in compact constructor
    public DateTimeData {
        // Validate localTime format
        if (localTime == null || localTime.isBlank()) {
            throw new IllegalArgumentException("Local time cannot be null or blank");
        }

        // Validate timezone
        if (timezone == null || timezone.isBlank()) {
            throw new IllegalArgumentException("Timezone cannot be null or blank");
        }
        try {
            ZoneId.of(timezone); // Validate IANA timezone
        } catch (Exception e) {
            throw new IllegalArgumentException(
                "Invalid timezone: " + timezone, e
            );
        }

        // Validate utcOffset format
        if (utcOffset == null || !utcOffset.matches("[+-]\\d{2}:\\d{2}")) {
            throw new IllegalArgumentException(
                "UTC offset must match format ±HH:MM, got: " + utcOffset
            );
        }

        // Validate city
        if (city == null || city.isBlank()) {
            throw new IllegalArgumentException("City cannot be null or blank");
        }

        // Validate timestamp
        if (timestamp == null) {
            throw new IllegalArgumentException("Timestamp cannot be null");
        }
        try {
            Instant.parse(timestamp);
        } catch (Exception e) {
            throw new IllegalArgumentException(
                "Timestamp must be ISO 8601 format, got: " + timestamp, e
            );
        }
    }

    /**
     * Creates DateTimeData for the given city and timezone.
     */
    public static DateTimeData create(String city, String timezoneId) {
        ZoneId zoneId = ZoneId.of(timezoneId);
        ZonedDateTime now = ZonedDateTime.now(zoneId);

        // Format local time as "yyyy-MM-dd HH:mm:ss" (matches Go format)
        String localTime = now.format(
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")
        );

        // Format UTC offset as "±HH:MM"
        String utcOffset = now.format(DateTimeFormatter.ofPattern("XXX"));

        return new DateTimeData(
            localTime,
            timezoneId,
            utcOffset,
            city,
            Instant.now().toString()
        );
    }
}
```

**Field Constraints**:
| Field | Type | Required | Constraints | Go Equivalent |
|-------|------|----------|-------------|---------------|
| local_time | String | Yes | Format: "yyyy-MM-dd HH:mm:ss" | string |
| timezone | String | Yes | Valid IANA timezone | string |
| utc_offset | String | Yes | Format: "±HH:MM" (e.g., "-05:00") | string |
| city | String | Yes | Non-blank | string |
| timestamp | String | Yes | ISO 8601 format (RFC3339) | string |

**Supported City Timezone Mappings** (matches Go implementation):
| City | Timezone | UTC Offset (example) |
|------|----------|---------------------|
| New York, NYC | America/New_York | -05:00 or -04:00 (DST) |
| Los Angeles, LA | America/Los_Angeles | -08:00 or -07:00 (DST) |
| Chicago | America/Chicago | -06:00 or -05:00 (DST) |
| Denver | America/Denver | -07:00 or -06:00 (DST) |
| London | Europe/London | +00:00 or +01:00 (BST) |
| Tokyo | Asia/Tokyo | +09:00 |

**Example JSON**:
```json
{
  "local_time": "2024-09-23 14:30:45",
  "timezone": "America/New_York",
  "utc_offset": "-04:00",
  "city": "New York",
  "timestamp": "2024-09-23T18:30:45Z"
}
```

### 3. EchoData

**Purpose**: Represents echoed text (simple pass-through)

**Go Equivalent**:
```go
type EchoResult struct {
    OriginalText string `json:"original_text"`
    EchoText     string `json:"echo_text"`
    Timestamp    string `json:"timestamp"`
}
```

**Java Definition**:
```java
package com.llmagents.mcp.echo.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import java.time.Instant;

/**
 * Represents echoed text.
 * Produces JSON compatible with Go EchoResult struct.
 */
@JsonPropertyOrder({"original_text", "echo_text", "timestamp"})
public record EchoData(
    @JsonProperty("original_text") String originalText,   // Input text
    @JsonProperty("echo_text") String echoText,            // Same as originalText
    String timestamp                                        // ISO 8601 format
) {
    // Validation in compact constructor
    public EchoData {
        // Validate originalText
        if (originalText == null) {
            throw new IllegalArgumentException("Original text cannot be null");
        }

        // Validate echoText matches originalText
        if (!originalText.equals(echoText)) {
            throw new IllegalArgumentException(
                "Echo text must match original text"
            );
        }

        // Validate timestamp
        if (timestamp == null) {
            throw new IllegalArgumentException("Timestamp cannot be null");
        }
        try {
            Instant.parse(timestamp);
        } catch (Exception e) {
            throw new IllegalArgumentException(
                "Timestamp must be ISO 8601 format, got: " + timestamp, e
            );
        }
    }

    /**
     * Creates EchoData with current timestamp.
     */
    public static EchoData create(String text) {
        return new EchoData(
            text,
            text,  // Echo is identical to original
            Instant.now().toString()
        );
    }
}
```

**Field Constraints**:
| Field | Type | Required | Constraints | Go Equivalent |
|-------|------|----------|-------------|---------------|
| original_text | String | Yes | None (can be empty string) | string |
| echo_text | String | Yes | Must equal original_text | string |
| timestamp | String | Yes | ISO 8601 format (RFC3339) | string |

**Example JSON**:
```json
{
  "original_text": "hello world",
  "echo_text": "hello world",
  "timestamp": "2024-09-23T14:30:45Z"
}
```

## MCP Protocol Models

### 4. MCP Request (JSON-RPC 2.0)

**Purpose**: Represents incoming JSON-RPC 2.0 requests from Go coordinator

**Specification**: https://www.jsonrpc.org/specification

**Java Definition**:
```java
package com.llmagents.mcp.common.model;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;

/**
 * JSON-RPC 2.0 request message.
 */
@JsonPropertyOrder({"jsonrpc", "method", "params", "id"})
public record MCPRequest(
    String jsonrpc,       // Always "2.0"
    String method,        // "tools/list" or "tools/call"
    Object params,        // Method-specific parameters (Map or null)
    Object id             // Request ID (String, Integer, or null)
) {
    public MCPRequest {
        if (!"2.0".equals(jsonrpc)) {
            throw new IllegalArgumentException(
                "JSON-RPC version must be '2.0', got: " + jsonrpc
            );
        }
        if (method == null || method.isBlank()) {
            throw new IllegalArgumentException("Method cannot be null or blank");
        }
    }
}
```

**Example**:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "getTemperature",
    "arguments": {
      "city": "New York"
    }
  },
  "id": 1
}
```

### 5. MCP Response (JSON-RPC 2.0)

**Purpose**: Represents outgoing JSON-RPC 2.0 responses to Go coordinator

**Java Definition**:
```java
package com.llmagents.mcp.common.model;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;

/**
 * JSON-RPC 2.0 response message.
 */
@JsonPropertyOrder({"jsonrpc", "result", "error", "id"})
@JsonInclude(JsonInclude.Include.NON_NULL)
public record MCPResponse(
    String jsonrpc,       // Always "2.0"
    Object result,        // Response data (null if error)
    MCPError error,       // Error details (null if success)
    Object id             // Matches request ID
) {
    public MCPResponse {
        if (!"2.0".equals(jsonrpc)) {
            throw new IllegalArgumentException(
                "JSON-RPC version must be '2.0', got: " + jsonrpc
            );
        }
        // Either result or error must be present, but not both
        if ((result == null && error == null) || (result != null && error != null)) {
            throw new IllegalArgumentException(
                "Response must have either result or error, but not both"
            );
        }
    }

    /**
     * Creates success response.
     */
    public static MCPResponse success(Object result, Object id) {
        return new MCPResponse("2.0", result, null, id);
    }

    /**
     * Creates error response.
     */
    public static MCPResponse error(MCPError error, Object id) {
        return new MCPResponse("2.0", null, error, id);
    }
}
```

**Example Success**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [{
      "type": "text",
      "text": "Weather in Tokyo: 37.3°C, Light rain"
    }]
  },
  "id": 1
}
```

**Example Error**:
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32600,
    "message": "Invalid Request"
  },
  "id": null
}
```

### 6. MCP Error

**Purpose**: Represents JSON-RPC 2.0 error object

**Java Definition**:
```java
package com.llmagents.mcp.common.model;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;

/**
 * JSON-RPC 2.0 error object.
 */
@JsonPropertyOrder({"code", "message", "data"})
public record MCPError(
    int code,          // Error code (JSON-RPC standard codes)
    String message,    // Error message
    Object data        // Optional additional error data
) {
    // Standard JSON-RPC error codes
    public static final int PARSE_ERROR = -32700;
    public static final int INVALID_REQUEST = -32600;
    public static final int METHOD_NOT_FOUND = -32601;
    public static final int INVALID_PARAMS = -32602;
    public static final int INTERNAL_ERROR = -32603;

    public static MCPError parseError(String message) {
        return new MCPError(PARSE_ERROR, message, null);
    }

    public static MCPError invalidRequest(String message) {
        return new MCPError(INVALID_REQUEST, message, null);
    }

    public static MCPError methodNotFound(String methodName) {
        return new MCPError(METHOD_NOT_FOUND, "Method not found: " + methodName, null);
    }

    public static MCPError invalidParams(String message) {
        return new MCPError(INVALID_PARAMS, message, null);
    }

    public static MCPError internalError(String message) {
        return new MCPError(INTERNAL_ERROR, message, null);
    }
}
```

## Validation Summary

| Model | Validation Rules | Test Coverage |
|-------|------------------|---------------|
| WeatherData | Temperature range, unit="°C", valid description, non-blank city, ISO 8601 timestamp | Unit tests for all validation rules |
| DateTimeData | Valid IANA timezone, UTC offset format, non-blank city, ISO 8601 timestamp | Unit tests for all validation rules |
| EchoData | originalText equals echoText, ISO 8601 timestamp | Unit tests for validation rules |
| MCPRequest | jsonrpc="2.0", non-blank method | Unit tests for validation rules |
| MCPResponse | jsonrpc="2.0", exclusive result/error | Unit tests for success and error cases |
| MCPError | Valid error codes | Unit tests for standard error codes |

## JSON Compatibility Checklist

- [x] Field names match Go struct tags exactly (snake_case with @JsonProperty)
- [x] Field order matches Go structs (@JsonPropertyOrder annotation)
- [x] Timestamp format is ISO 8601 (RFC3339)
- [x] Null fields omitted in JSON output (@JsonInclude.NON_NULL)
- [x] Number formatting matches Go (no trailing zeros)
- [x] Record types provide immutability
- [x] Validation in compact constructors
- [x] Factory methods for convenience

## Next Steps

1. ✅ Data models defined
2. → Create API contracts in /contracts/ directory
3. → Write contract tests (must fail initially - TDD)
4. → Implement tool handlers to make tests pass
5. → Assemble servers with MCP SDK integration

---
*Data model design completed: 2025-10-07*
*Ready for API contract generation*
