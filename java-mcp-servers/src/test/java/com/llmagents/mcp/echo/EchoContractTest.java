package com.llmagents.mcp.echo;

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
 * Contract test for Echo MCP Server.
 * Tests compliance with echo-api.json contract.
 */
@DisplayName("Echo API Contract Tests")
class EchoContractTest {

    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = JsonConfig.createObjectMapper();
    }

    @Test
    @DisplayName("EchoData should have correct JSON structure")
    @SuppressWarnings("unchecked")
    void testEchoDataJsonStructure() throws Exception {
        EchoData data = EchoData.echo("hello world");

        String json = objectMapper.writeValueAsString(data);
        Map<String, Object> map = objectMapper.readValue(json, Map.class);

        // Verify required fields
        assertThat(map).containsKeys("original_text", "echo_text", "timestamp");
        assertThat(map.get("original_text")).isEqualTo("hello world");
        assertThat(map.get("echo_text")).isEqualTo("hello world");
    }

    @Test
    @DisplayName("Echo response format matches Go implementation")
    void testEchoResponseFormat() {
        EchoData data = EchoData.echo("test message");
        String formatted = data.format();

        // Format: "Echo: {text}"
        assertThat(formatted).isEqualTo("Echo: test message");
    }

    @Test
    @DisplayName("Echo handles empty string")
    void testEchoEmptyString() {
        EchoData data = EchoData.echo("");
        assertThat(data.echoText()).isEmpty();
        assertThat(data.format()).isEqualTo("Echo: ");
    }

    @Test
    @DisplayName("Echo handles special characters")
    void testEchoSpecialCharacters() {
        String special = "Hello, World! 🌍 @#$%^&*()";
        EchoData data = EchoData.echo(special);
        assertThat(data.echoText()).isEqualTo(special);
    }

    @Test
    @DisplayName("EchoTool handles valid request")
    void testEchoToolValidRequest() {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of(
                "name", "echo",
                "arguments", Map.of("text", "hello world")
            ),
            1
        );

        EchoTool tool = new EchoTool();
        JsonRpcResponse response = tool.handleToolCall(request);

        assertThat(response.isSuccess()).isTrue();
        assertThat(response.id()).isEqualTo(1);
    }

    @Test
    @DisplayName("EchoTool returns error for missing text")
    void testEchoToolMissingText() {
        JsonRpcRequest request = new JsonRpcRequest(
            "2.0",
            "tools/call",
            Map.of("name", "echo", "arguments", Map.of()),
            2
        );

        EchoTool tool = new EchoTool();
        JsonRpcResponse response = tool.handleToolCall(request);

        assertThat(response.isSuccess()).isFalse();
        assertThat(response.error()).isNotNull();
        assertThat(response.error().code()).isEqualTo(JsonRpcError.INVALID_PARAMS);
    }
}
