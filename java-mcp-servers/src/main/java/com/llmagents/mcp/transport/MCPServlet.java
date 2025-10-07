package com.llmagents.mcp.transport;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.llmagents.mcp.common.JsonConfig;
import com.llmagents.mcp.common.protocol.JsonRpcError;
import com.llmagents.mcp.common.protocol.JsonRpcRequest;
import com.llmagents.mcp.common.protocol.JsonRpcResponse;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.util.List;
import java.util.Map;
import java.util.function.Function;

/**
 * Base servlet for MCP protocol over HTTP/SSE.
 * Handles JSON-RPC 2.0 requests and routes them to tool handlers.
 */
public class MCPServlet extends HttpServlet {
    private static final Logger logger = LoggerFactory.getLogger(MCPServlet.class);

    private final ObjectMapper objectMapper;
    private final Function<JsonRpcRequest, JsonRpcResponse> toolHandler;
    private final List<Map<String, Object>> toolDefinitions;
    private final String serverName;

    /**
     * Create MCP servlet.
     *
     * @param serverName server name for logging
     * @param toolHandler function to handle tool calls
     * @param toolDefinitions list of tool definitions for tools/list
     */
    public MCPServlet(
            String serverName,
            Function<JsonRpcRequest, JsonRpcResponse> toolHandler,
            List<Map<String, Object>> toolDefinitions) {
        this.serverName = serverName;
        this.toolHandler = toolHandler;
        this.toolDefinitions = toolDefinitions;
        this.objectMapper = JsonConfig.createObjectMapper();
    }

    @Override
    protected void doPost(HttpServletRequest req, HttpServletResponse resp) throws IOException {
        resp.setContentType("application/json");
        resp.setCharacterEncoding("UTF-8");

        try {
            // Parse JSON-RPC request
            JsonRpcRequest request = objectMapper.readValue(req.getInputStream(), JsonRpcRequest.class);
            logger.info("[{}] Received request: method={}, id={}", serverName, request.method(), request.id());

            // Route based on method
            JsonRpcResponse response = switch (request.method()) {
                case "initialize" -> handleInitialize(request);
                case "tools/list" -> handleToolsList(request);
                case "tools/call" -> toolHandler.apply(request);
                default -> JsonRpcResponse.error(
                    JsonRpcError.methodNotFound(request.method()),
                    request.id()
                );
            };

            // Write response
            String jsonResponse = objectMapper.writeValueAsString(response);
            logger.debug("[{}] Sending response: {}", serverName, jsonResponse);
            resp.getWriter().write(jsonResponse);
            resp.setStatus(HttpServletResponse.SC_OK);

        } catch (com.fasterxml.jackson.core.JsonParseException e) {
            // Malformed JSON - return parse error (-32700)
            logger.error("[{}] Parse error: {}", serverName, e.getMessage());
            JsonRpcResponse errorResponse = JsonRpcResponse.error(
                JsonRpcError.parseError(e.getMessage()),
                null
            );
            resp.getWriter().write(objectMapper.writeValueAsString(errorResponse));
            resp.setStatus(HttpServletResponse.SC_OK);  // Still return 200 per JSON-RPC spec

        } catch (Exception e) {
            logger.error("[{}] Internal error", serverName, e);
            JsonRpcResponse errorResponse = JsonRpcResponse.error(
                JsonRpcError.internalError(e.getMessage()),
                null
            );
            resp.getWriter().write(objectMapper.writeValueAsString(errorResponse));
            resp.setStatus(HttpServletResponse.SC_INTERNAL_SERVER_ERROR);
        }
    }

    /**
     * Handle initialize request (MCP handshake).
     *
     * @param request JSON-RPC request
     * @return JSON-RPC response with server capabilities
     */
    private JsonRpcResponse handleInitialize(JsonRpcRequest request) {
        logger.info("[{}] Handling initialize request", serverName);
        Map<String, Object> result = Map.of(
            "protocolVersion", "2024-11-05",
            "capabilities", Map.of(
                "tools", Map.of()
            ),
            "serverInfo", Map.of(
                "name", serverName,
                "version", "1.0.0"
            )
        );
        return JsonRpcResponse.success(result, request.id());
    }

    /**
     * Handle tools/list request.
     *
     * @param request JSON-RPC request
     * @return JSON-RPC response with tool definitions
     */
    private JsonRpcResponse handleToolsList(JsonRpcRequest request) {
        logger.info("[{}] Handling tools/list request", serverName);
        Map<String, Object> result = Map.of("tools", toolDefinitions);
        return JsonRpcResponse.success(result, request.id());
    }

    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws IOException {
        // Return server info for health check
        resp.setContentType("application/json");
        resp.setCharacterEncoding("UTF-8");
        Map<String, String> info = Map.of(
            "server", serverName,
            "protocol", "MCP (JSON-RPC 2.0 over HTTP)",
            "status", "running"
        );
        resp.getWriter().write(objectMapper.writeValueAsString(info));
        resp.setStatus(HttpServletResponse.SC_OK);
    }
}
