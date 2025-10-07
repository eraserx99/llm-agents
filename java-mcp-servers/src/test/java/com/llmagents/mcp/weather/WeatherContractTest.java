package com.llmagents.mcp.weather;

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

/**
 * Contract test for Weather MCP Server.
 * Tests compliance with weather-api.json contract.
 * These tests MUST FAIL initially (TDD approach) until WeatherTool is implemented.
 */
@DisplayName("Weather API Contract Tests")
class WeatherContractTest {

    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = JsonConfig.createObjectMapper();
    }

    @Test
    @DisplayName("WeatherData should have correct JSON structure")
    @SuppressWarnings("unchecked")
    void testWeatherDataJsonStructure() throws Exception {
        // This test will pass because WeatherData is already implemented
        WeatherData data = WeatherData.simulate("New York");

        String json = objectMapper.writeValueAsString(data);
        Map<String, Object> map = objectMapper.readValue(json, Map.class);

        // Verify required fields
        assertThat(map).containsKeys("temperature", "unit", "description", "city", "timestamp");
        assertThat(map.get("unit")).isEqualTo("°C");
        assertThat(map.get("city")).isEqualTo("New York");
    }

    @Test
    @DisplayName("Weather response format matches Go implementation")
    void testWeatherResponseFormat() {
        WeatherData data = WeatherData.simulate("Tokyo");
        String formatted = data.format();

        // Format: "Weather in {city}: {temp}°C, {description}"
        assertThat(formatted).matches("Weather in Tokyo: \\d+\\.\\d°C, .+");
    }

    @Test
    @DisplayName("Temperature in valid range (20-45°C)")
    void testTemperatureRange() {
        for (int i = 0; i < 100; i++) {
            WeatherData data = WeatherData.simulate("Test City");
            assertThat(data.temperature())
                .isBetween(20.0, 45.0);
        }
    }

    @Test
    @DisplayName("WeatherTool handles valid request")
    void testWeatherToolValidRequest() {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of(
                "name", "getTemperature",
                "arguments", Map.of("city", "Tokyo")
            ),
            1
        );

        WeatherTool tool = new WeatherTool();
        JsonRpcResponse response = tool.handleToolCall(request);

        assertThat(response.isSuccess()).isTrue();
        assertThat(response.id()).isEqualTo(1);
    }

    @Test
    @DisplayName("WeatherTool returns error for missing city")
    void testWeatherToolMissingCity() {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of("name", "getTemperature", "arguments", Map.of()),
            2
        );

        WeatherTool tool = new WeatherTool();
        JsonRpcResponse response = tool.handleToolCall(request);

        assertThat(response.isSuccess()).isFalse();
        assertThat(response.error()).isNotNull();
        assertThat(response.error().code()).isEqualTo(JsonRpcError.INVALID_PARAMS);
    }
}
