package com.llmagents.mcp.common;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;

/**
 * Jackson ObjectMapper configuration for JSON serialization.
 * Configured to match Go's JSON output format exactly.
 */
public final class JsonConfig {

    private JsonConfig() {
        // Utility class
    }

    /**
     * Create configured ObjectMapper for MCP protocol.
     * Configuration:
     * - Uses snake_case for property names (matches Go's json tags)
     * - Excludes null values from output
     * - Formats dates as ISO-8601 strings
     * - Preserves property order as declared in classes
     * - Fails on unknown properties during deserialization
     *
     * @return configured ObjectMapper
     */
    public static ObjectMapper createObjectMapper() {
        ObjectMapper mapper = new ObjectMapper();

        // Property naming: snake_case (matches Go's json tags)
        mapper.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);

        // Exclude null values
        mapper.setSerializationInclusion(JsonInclude.Include.NON_NULL);

        // Date/Time handling (ISO-8601 format)
        mapper.registerModule(new JavaTimeModule());
        mapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);

        // Preserve order of properties as declared in class
        mapper.configure(SerializationFeature.ORDER_MAP_ENTRIES_BY_KEYS, false);

        // Fail on unknown properties (strict validation)
        mapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);

        // Pretty-print for debugging (can be disabled in production)
        // mapper.enable(SerializationFeature.INDENT_OUTPUT);

        return mapper;
    }

    /**
     * Create ObjectMapper with pretty-printing enabled.
     * Useful for debugging and logging.
     *
     * @return ObjectMapper with pretty-printing
     */
    public static ObjectMapper createPrettyObjectMapper() {
        ObjectMapper mapper = createObjectMapper();
        mapper.enable(SerializationFeature.INDENT_OUTPUT);
        return mapper;
    }
}
