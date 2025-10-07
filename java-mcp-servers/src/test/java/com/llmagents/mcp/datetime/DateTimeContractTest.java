package com.llmagents.mcp.datetime;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.llmagents.mcp.common.JsonConfig;
import com.llmagents.mcp.common.protocol.JsonRpcError;
import com.llmagents.mcp.common.protocol.JsonRpcRequest;
import com.llmagents.mcp.common.protocol.JsonRpcResponse;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

/**
 * Contract test for DateTime MCP Server.
 * Tests compliance with datetime-api.json contract.
 */
@DisplayName("DateTime API Contract Tests")
class DateTimeContractTest {

    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = JsonConfig.createObjectMapper();
    }

    @Test
    @DisplayName("DateTimeData should have correct JSON structure")
    @SuppressWarnings("unchecked")
    void testDateTimeDataJsonStructure() throws Exception {
        DateTimeData data = DateTimeData.forCity("New York");

        String json = objectMapper.writeValueAsString(data);
        Map<String, Object> map = objectMapper.readValue(json, Map.class);

        // Verify required fields
        assertThat(map).containsKeys("local_time", "timezone", "utc_offset", "city", "timestamp");
        assertThat(map.get("city")).isEqualTo("New York");
        assertThat(map.get("timezone")).isEqualTo("America/New_York");
    }

    @Test
    @DisplayName("DateTime response format matches Go implementation")
    void testDateTimeResponseFormat() {
        DateTimeData data = DateTimeData.forCity("Los Angeles");
        String formatted = data.format();

        // Format: "Time in {city}: {localTime} ({timezone}, UTC{offset})"
        assertThat(formatted).matches(
            "Time in Los Angeles: \\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} \\(America/Los_Angeles, UTC[+-]\\d{2}:\\d{2}\\)"
        );
    }

    @Test
    @DisplayName("Unsupported city throws IllegalArgumentException")
    void testUnsupportedCity() {
        assertThatThrownBy(() -> DateTimeData.forCity("UnknownCity"))
            .isInstanceOf(IllegalArgumentException.class)
            .hasMessageContaining("Unsupported city");
    }

    @Test
    @DisplayName("City alias NYC resolves to New York timezone")
    void testCityAlias() {
        DateTimeData data = DateTimeData.forCity("NYC");
        assertThat(data.timezone()).isEqualTo("America/New_York");
    }

    @Test
    @DisplayName("DateTimeTool handles valid request")
    void testDateTimeToolValidRequest() {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of(
                "name", "getDateTime",
                "arguments", Map.of("city", "London")
            ),
            1
        );

        DateTimeTool tool = new DateTimeTool();
        JsonRpcResponse response = tool.handleToolCall(request);

        assertThat(response.isSuccess()).isTrue();
        assertThat(response.id()).isEqualTo(1);
    }

    @Test
    @DisplayName("DateTimeTool returns error for unsupported city")
    void testDateTimeToolUnsupportedCity() {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of(
                "name", "getDateTime",
                "arguments", Map.of("city", "UnknownCity")
            ),
            2
        );

        DateTimeTool tool = new DateTimeTool();
        JsonRpcResponse response = tool.handleToolCall(request);

        assertThat(response.isSuccess()).isFalse();
        assertThat(response.error()).isNotNull();
        assertThat(response.error().code()).isEqualTo(JsonRpcError.INVALID_PARAMS);
    }
}
