package com.llmagents.mcp.common.protocol;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * JSON-RPC 2.0 response matching MCP protocol.
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record JsonRpcResponse(
    @JsonProperty("jsonrpc") String jsonrpc,
    @JsonProperty("result") Object result,
    @JsonProperty("error") JsonRpcError error,
    @JsonProperty("id") Object id
) {
    public JsonRpcResponse {
        if (jsonrpc == null || !jsonrpc.equals("2.0")) {
            throw new IllegalArgumentException("jsonrpc must be '2.0'");
        }
        // Either result or error must be present, but not both
        if (result == null && error == null) {
            throw new IllegalArgumentException("Either result or error must be present");
        }
        if (result != null && error != null) {
            throw new IllegalArgumentException("Cannot have both result and error");
        }
    }

    /**
     * Create success response.
     *
     * @param result result object
     * @param id request ID
     * @return JsonRpcResponse with result
     */
    public static JsonRpcResponse success(Object result, Object id) {
        return new JsonRpcResponse("2.0", result, null, id);
    }

    /**
     * Create error response.
     *
     * @param error error object
     * @param id request ID
     * @return JsonRpcResponse with error
     */
    public static JsonRpcResponse error(JsonRpcError error, Object id) {
        return new JsonRpcResponse("2.0", null, error, id);
    }

    /**
     * Check if response is successful.
     *
     * @return true if result is present and error is null
     */
    public boolean isSuccess() {
        return result != null && error == null;
    }
}
