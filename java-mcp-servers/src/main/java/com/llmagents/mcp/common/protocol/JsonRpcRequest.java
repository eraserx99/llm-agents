package com.llmagents.mcp.common.protocol;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

/**
 * JSON-RPC 2.0 request matching MCP protocol.
 */
public record JsonRpcRequest(
    @JsonProperty("jsonrpc") String jsonrpc,
    @JsonProperty("method") String method,
    @JsonProperty("params") Map<String, Object> params,
    @JsonProperty("id") Object id
) {
    public JsonRpcRequest {
        if (jsonrpc == null || !jsonrpc.equals("2.0")) {
            throw new IllegalArgumentException("jsonrpc must be '2.0'");
        }
        if (method == null || method.isBlank()) {
            throw new IllegalArgumentException("method cannot be null or blank");
        }
    }

    /**
     * Get parameter value by name.
     *
     * @param name parameter name
     * @return parameter value or null if not present
     */
    public Object getParam(String name) {
        return params != null ? params.get(name) : null;
    }

    /**
     * Get parameter value as String.
     *
     * @param name parameter name
     * @return parameter value as String or null
     */
    @SuppressWarnings("unchecked")
    public String getParamAsString(String name) {
        Object value = getParam(name);
        return value != null ? value.toString() : null;
    }

    /**
     * Get parameter value as Map.
     *
     * @param name parameter name
     * @return parameter value as Map or null
     */
    @SuppressWarnings("unchecked")
    public Map<String, Object> getParamAsMap(String name) {
        Object value = getParam(name);
        return value instanceof Map ? (Map<String, Object>) value : null;
    }
}
