package com.llmagents.mcp.datetime;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;

import java.time.Instant;
import java.time.ZoneId;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Map;

/**
 * DateTime data matching Go implementation's format.
 * Example: {"local_time":"2024-09-23 11:30:45","timezone":"America/Los_Angeles","utc_offset":"-07:00","city":"Los Angeles","timestamp":"2024-09-23T18:30:45Z"}
 */
@JsonPropertyOrder({"local_time", "timezone", "utc_offset", "city", "timestamp"})
public record DateTimeData(
    String localTime,
    String timezone,
    String utcOffset,
    String city,
    String timestamp
) {
    // Timezone mappings matching Go implementation
    private static final Map<String, String> CITY_TIMEZONES = Map.of(
        "New York", "America/New_York",
        "NYC", "America/New_York",
        "Los Angeles", "America/Los_Angeles",
        "LA", "America/Los_Angeles",
        "Chicago", "America/Chicago",
        "Denver", "America/Denver",
        "London", "Europe/London",
        "Tokyo", "Asia/Tokyo"
    );

    // Default timezone if city not found
    private static final String DEFAULT_TIMEZONE = "America/New_York";

    // Date/time formatters matching Go implementation
    private static final DateTimeFormatter LOCAL_TIME_FORMATTER =
        DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
    private static final DateTimeFormatter OFFSET_FORMATTER =
        DateTimeFormatter.ofPattern("XXX");  // Formats as ±HH:MM

    /**
     * Validate datetime data.
     */
    public DateTimeData {
        if (localTime == null || localTime.isBlank()) {
            throw new IllegalArgumentException("Local time cannot be null or blank");
        }
        if (timezone == null || timezone.isBlank()) {
            throw new IllegalArgumentException("Timezone cannot be null or blank");
        }
        if (utcOffset == null || utcOffset.isBlank()) {
            throw new IllegalArgumentException("UTC offset cannot be null or blank");
        }
        if (city == null || city.isBlank()) {
            throw new IllegalArgumentException("City cannot be null or blank");
        }
        if (timestamp == null || timestamp.isBlank()) {
            throw new IllegalArgumentException("Timestamp cannot be null or blank");
        }
    }

    /**
     * Get current datetime for a city.
     * Matches Go implementation's timezone mapping logic.
     *
     * @param city city name
     * @return DateTimeData for the city
     * @throws IllegalArgumentException if city is not supported
     */
    public static DateTimeData forCity(String city) {
        // Get timezone for city (case-insensitive lookup)
        String timezoneId = CITY_TIMEZONES.entrySet().stream()
            .filter(entry -> entry.getKey().equalsIgnoreCase(city))
            .map(Map.Entry::getValue)
            .findFirst()
            .orElse(null);

        if (timezoneId == null) {
            throw new IllegalArgumentException(
                "Unsupported city: " + city +
                ". Supported cities: " + String.join(", ", CITY_TIMEZONES.keySet())
            );
        }

        ZoneId zoneId = ZoneId.of(timezoneId);
        ZonedDateTime now = ZonedDateTime.now(zoneId);

        String localTime = now.format(LOCAL_TIME_FORMATTER);
        String timezone = timezoneId;
        String utcOffset = now.format(OFFSET_FORMATTER);
        String timestamp = Instant.now().toString();

        return new DateTimeData(localTime, timezone, utcOffset, city, timestamp);
    }

    /**
     * Format datetime data as human-readable string.
     * Matches Go implementation's format: "Time in {city}: {localTime} ({timezone}, UTC{offset})"
     *
     * @return formatted string
     */
    public String format() {
        return String.format("Time in %s: %s (%s, UTC%s)", city, localTime, timezone, utcOffset);
    }

    /**
     * Check if a city is supported.
     *
     * @param city city name
     * @return true if city has timezone mapping
     */
    public static boolean isCitySupported(String city) {
        return CITY_TIMEZONES.keySet().stream()
            .anyMatch(key -> key.equalsIgnoreCase(city));
    }
}
