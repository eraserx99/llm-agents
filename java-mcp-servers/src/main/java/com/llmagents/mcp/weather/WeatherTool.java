package com.llmagents.mcp.weather;

import com.llmagents.mcp.common.protocol.JsonRpcError;
import com.llmagents.mcp.common.protocol.JsonRpcRequest;
import com.llmagents.mcp.common.protocol.JsonRpcResponse;
import com.llmagents.mcp.common.protocol.ToolCallResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;

/**
 * MCP tool handler for weather queries.
 * Implements getTemperature tool matching Go implementation.
 */
public class WeatherTool {
    private static final Logger logger = LoggerFactory.getLogger(WeatherTool.class);

    public static final String TOOL_NAME = "getTemperature";
    public static final String TOOL_DESCRIPTION = "Get current temperature and weather conditions for a city";

    /**
     * Handle MCP tool call request.
     *
     * @param request JSON-RPC request
     * @return JSON-RPC response with weather data
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

            // Extract city parameter
            Object cityObj = arguments.get("city");
            if (cityObj == null) {
                logger.warn("Missing city parameter");
                return JsonRpcResponse.error(
                    JsonRpcError.invalidParams("Missing required parameter: city"),
                    request.id()
                );
            }

            String city = cityObj.toString();
            logger.info("Getting weather for city: {}", city);

            // Simulate weather data
            WeatherData weatherData = WeatherData.simulate(city);

            logger.debug("Weather data: {}", weatherData);

            // Return structured result with both text and JSON content
            ToolCallResult result = ToolCallResult.withStructuredContent(
                weatherData.format(),  // Human-readable text
                weatherData            // Structured JSON data
            );
            return JsonRpcResponse.success(result, request.id());

        } catch (Exception e) {
            logger.error("Error handling weather tool call", e);
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
                    "city", Map.of(
                        "type", "string",
                        "description", "the city to get weather for"
                    )
                ),
                "required", new String[]{"city"}
            )
        );
    }
}
