package com.llmagents.mcp.echo;

import com.llmagents.mcp.common.protocol.JsonRpcError;
import com.llmagents.mcp.common.protocol.JsonRpcRequest;
import com.llmagents.mcp.common.protocol.JsonRpcResponse;
import com.llmagents.mcp.common.protocol.ToolCallResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;

/**
 * MCP tool handler for echo functionality.
 * Implements echo tool matching Go implementation.
 */
public class EchoTool {
    private static final Logger logger = LoggerFactory.getLogger(EchoTool.class);

    public static final String TOOL_NAME = "echo";
    public static final String TOOL_DESCRIPTION = "Echo back the provided text";

    /**
     * Handle MCP tool call request.
     *
     * @param request JSON-RPC request
     * @return JSON-RPC response with echoed text
     */
    public JsonRpcResponse handleToolCall(JsonRpcRequest request) {
        try {
            // Extract tool name and arguments
            String toolName = request.getParamAsString("name");
            Map<String, Object> arguments = request.getParamAsMap("arguments");

            if (toolName == null || !toolName.equals(TOOL_NAME)) {
                logger.warn("Invalid tool name: {}", toolName);
                return JsonRpcResponse.error(
                    JsonRpcError.methodNotFound(toolName),
                    request.id()
                );
            }

            if (arguments == null) {
                logger.warn("Missing arguments in request");
                return JsonRpcResponse.error(
                    JsonRpcError.invalidParams("Missing arguments"),
                    request.id()
                );
            }

            // Extract text parameter
            Object textObj = arguments.get("text");
            if (textObj == null) {
                logger.warn("Missing text parameter");
                return JsonRpcResponse.error(
                    JsonRpcError.invalidParams("Missing required parameter: text"),
                    request.id()
                );
            }

            String text = textObj.toString();
            logger.info("Echoing text: {}", text);

            // Echo the text
            EchoData echoData = EchoData.echo(text);

            logger.debug("Echo result: {}", echoData);

            // Return structured result with both text and JSON content
            ToolCallResult result = ToolCallResult.withStructuredContent(
                echoData.format(),  // Human-readable text
                echoData            // Structured JSON data
            );
            return JsonRpcResponse.success(result, request.id());

        } catch (Exception e) {
            logger.error("Error handling echo tool call", e);
            return JsonRpcResponse.error(
                JsonRpcError.internalError(e.getMessage()),
                request.id()
            );
        }
    }

    /**
     * Get tool definition for MCP tools/list response.
     *
     * @return tool definition map
     */
    public static Map<String, Object> getToolDefinition() {
        return Map.of(
            "name", TOOL_NAME,
            "description", TOOL_DESCRIPTION,
            "inputSchema", Map.of(
                "type", "object",
                "properties", Map.of(
                    "text", Map.of(
                        "type", "string",
                        "description", "the text to echo back"
                    )
                ),
                "required", new String[]{"text"}
            )
        );
    }
}
