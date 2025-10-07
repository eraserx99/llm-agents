package com.llmagents.mcp.common.protocol;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.llmagents.mcp.common.JsonConfig;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

/**
 * Contract test for JSON-RPC 2.0 protocol.
 * Tests compliance with mcp-protocol.json contract.
 */
@DisplayName("JSON-RPC 2.0 Protocol Tests")
class JsonRpcProtocolTest {

    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = JsonConfig.createObjectMapper();
    }

    @Test
    @DisplayName("JsonRpcRequest serializes correctly")
    void testRequestSerialization() throws Exception {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of("name", "getTemperature", "arguments", Map.of("city", "New York")),
            1
        );

        String json = objectMapper.writeValueAsString(request);
        assertThat(json).contains("\"jsonrpc\":\"2.0\"");
        assertThat(json).contains("\"method\":\"tools/call\"");
    }

    @Test
    @DisplayName("JsonRpcRequest requires jsonrpc 2.0")
    void testRequestValidation() {
        assertThatThrownBy(() -> new JsonRpcRequest("1.0", "test", null, 1))
            .isInstanceOf(IllegalArgumentException.class)
            .hasMessageContaining("jsonrpc must be '2.0'");
    }

    @Test
    @DisplayName("JsonRpcResponse success format")
    void testSuccessResponse() throws Exception {
        JsonRpcResponse response = JsonRpcResponse.success(
            ToolCallResult.fromText("Weather in Tokyo: 37.3°C, Light rain"),
            1
        );

        assertThat(response.isSuccess()).isTrue();
        String json = objectMapper.writeValueAsString(response);
        assertThat(json).contains("\"jsonrpc\":\"2.0\"");
        assertThat(json).contains("\"result\":");
        assertThat(json).doesNotContain("\"error\":");
    }

    @Test
    @DisplayName("JsonRpcResponse error format")
    void testErrorResponse() throws Exception {
        JsonRpcResponse response = JsonRpcResponse.error(
            JsonRpcError.invalidParams("Missing city parameter"),
            1
        );

        assertThat(response.isSuccess()).isFalse();
        String json = objectMapper.writeValueAsString(response);
        assertThat(json).contains("\"jsonrpc\":\"2.0\"");
        assertThat(json).contains("\"error\":");
        assertThat(json).doesNotContain("\"result\":");
    }

    @Test
    @DisplayName("JsonRpcError parse error (-32700)")
    void testParseError() {
        JsonRpcError error = JsonRpcError.parseError("Invalid JSON");
        assertThat(error.code()).isEqualTo(JsonRpcError.PARSE_ERROR);
        assertThat(error.code()).isEqualTo(-32700);
    }

    @Test
    @DisplayName("JsonRpcError invalid params (-32602)")
    void testInvalidParamsError() {
        JsonRpcError error = JsonRpcError.invalidParams("City not supported");
        assertThat(error.code()).isEqualTo(JsonRpcError.INVALID_PARAMS);
        assertThat(error.code()).isEqualTo(-32602);
    }

    @Test
    @DisplayName("ToolCallResult content format")
    @SuppressWarnings("unchecked")
    void testToolCallResultFormat() throws Exception {
        ToolCallResult result = ToolCallResult.fromText("Echo: hello");

        String json = objectMapper.writeValueAsString(result);
        Map<String, Object> map = objectMapper.readValue(json, Map.class);

        assertThat(map).containsKey("content");
    }
}
