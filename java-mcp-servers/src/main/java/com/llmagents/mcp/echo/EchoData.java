package com.llmagents.mcp.echo;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;

import java.time.Instant;

/**
 * Echo data matching Go implementation's format.
 * Example: {"original_text":"hello world","echo_text":"hello world","timestamp":"2024-09-23T14:30:45Z"}
 */
@JsonPropertyOrder({"original_text", "echo_text", "timestamp"})
public record EchoData(
    String originalText,
    String echoText,
    String timestamp
) {
    /**
     * Validate echo data.
     */
    public EchoData {
        if (originalText == null) {
            throw new IllegalArgumentException("Original text cannot be null");
        }
        if (echoText == null) {
            throw new IllegalArgumentException("Echo text cannot be null");
        }
        if (timestamp == null || timestamp.isBlank()) {
            throw new IllegalArgumentException("Timestamp cannot be null or blank");
        }
    }

    /**
     * Create echo data from input text.
     * Matches Go implementation's echo logic (text is echoed as-is).
     *
     * @param text input text to echo
     * @return EchoData with echoed text
     */
    public static EchoData echo(String text) {
        if (text == null) {
            text = "";
        }

        String timestamp = Instant.now().toString();
        return new EchoData(text, text, timestamp);
    }

    /**
     * Format echo data as human-readable string.
     * Matches Go implementation's format: "Echo: {text}"
     *
     * @return formatted string
     */
    public String format() {
        return "Echo: " + echoText;
    }
}
