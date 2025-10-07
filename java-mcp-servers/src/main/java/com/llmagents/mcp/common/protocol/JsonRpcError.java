package com.llmagents.mcp.common.protocol;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * JSON-RPC 2.0 error matching MCP protocol.
 * Standard error codes:
 * - -32700: Parse error
 * - -32600: Invalid Request
 * - -32601: Method not found
 * - -32602: Invalid params
 * - -32603: Internal error
 * - -32000 to -32099: Server error (reserved for implementation-defined errors)
 */
public record JsonRpcError(
    @JsonProperty("code") int code,
    @JsonProperty("message") String message,
    @JsonProperty("data") Object data
) {
    // Standard JSON-RPC error codes
    public static final int PARSE_ERROR = -32700;
    public static final int INVALID_REQUEST = -32600;
    public static final int METHOD_NOT_FOUND = -32601;
    public static final int INVALID_PARAMS = -32602;
    public static final int INTERNAL_ERROR = -32603;

    /**
     * Validate error.
     */
    public JsonRpcError {
        if (message == null || message.isBlank()) {
            throw new IllegalArgumentException("Error message cannot be null or blank");
        }
    }

    /**
     * Create parse error (-32700).
     *
     * @param detail error detail
     * @return JsonRpcError for parse error
     */
    public static JsonRpcError parseError(String detail) {
        return new JsonRpcError(
            PARSE_ERROR,
            "Parse error",
            detail
        );
    }

    /**
     * Create invalid request error (-32600).
     *
     * @param detail error detail
     * @return JsonRpcError for invalid request
     */
    public static JsonRpcError invalidRequest(String detail) {
        return new JsonRpcError(
            INVALID_REQUEST,
            "Invalid Request",
            detail
        );
    }

    /**
     * Create method not found error (-32601).
     *
     * @param method method name
     * @return JsonRpcError for method not found
     */
    public static JsonRpcError methodNotFound(String method) {
        return new JsonRpcError(
            METHOD_NOT_FOUND,
            "Method not found",
            "Method '" + method + "' is not supported"
        );
    }

    /**
     * Create invalid params error (-32602).
     *
     * @param detail error detail
     * @return JsonRpcError for invalid params
     */
    public static JsonRpcError invalidParams(String detail) {
        return new JsonRpcError(
            INVALID_PARAMS,
            "Invalid params",
            detail
        );
    }

    /**
     * Create internal error (-32603).
     *
     * @param detail error detail
     * @return JsonRpcError for internal error
     */
    public static JsonRpcError internalError(String detail) {
        return new JsonRpcError(
            INTERNAL_ERROR,
            "Internal error",
            detail
        );
    }
}
