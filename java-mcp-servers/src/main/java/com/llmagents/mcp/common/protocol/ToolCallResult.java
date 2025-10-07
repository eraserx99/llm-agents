package com.llmagents.mcp.common.protocol;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

/**
 * MCP tool call result matching protocol format.
 * Wraps tool output in MCP content format.
 * The structuredContent field is at the same level as content to match the official SDK format.
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record ToolCallResult(
    @JsonProperty("content") List<ContentItem> content,
    @JsonProperty("structuredContent") Object structuredContent
) {
    /**
     * Create tool call result from text.
     *
     * @param text result text
     * @return ToolCallResult with text content
     */
    public static ToolCallResult fromText(String text) {
        return new ToolCallResult(
            List.of(new ContentItem("text", text, null)),
            null
        );
    }

    /**
     * Create tool call result with structured content.
     * The structured content is placed at the top level to match the official MCP SDK format.
     *
     * @param text human-readable text
     * @param structuredData structured JSON data
     * @return ToolCallResult with both text and structured content
     */
    public static ToolCallResult withStructuredContent(String text, Object structuredData) {
        return new ToolCallResult(
            List.of(new ContentItem("text", text, null)),
            structuredData
        );
    }

    /**
     * Content item in MCP protocol.
     */
    @JsonInclude(JsonInclude.Include.NON_NULL)
    public record ContentItem(
        @JsonProperty("type") String type,
        @JsonProperty("text") String text,
        @JsonProperty("data") Object data
    ) {
        public ContentItem {
            if (type == null || type.isBlank()) {
                throw new IllegalArgumentException("Content type cannot be null or blank");
            }
        }
    }

}
