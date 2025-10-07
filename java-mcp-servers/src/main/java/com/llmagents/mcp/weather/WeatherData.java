package com.llmagents.mcp.weather;

import com.fasterxml.jackson.annotation.JsonPropertyOrder;

import java.time.Instant;
import java.util.Random;

/**
 * Weather data matching Go implementation's format.
 * Example: {"temperature":37.3,"unit":"°C","description":"Light rain","city":"Tokyo","timestamp":"2024-09-23T14:30:45Z"}
 */
@JsonPropertyOrder({"temperature", "unit", "description", "city", "timestamp"})
public record WeatherData(
    double temperature,
    String unit,
    String description,
    String city,
    String timestamp
) {
    private static final Random random = new Random();

    // Weather conditions matching Go implementation
    private static final String[] WEATHER_CONDITIONS = {
        "Sunny",
        "Partly cloudy",
        "Cloudy",
        "Light rain",
        "Heavy rain",
        "Thunderstorms",
        "Foggy",
        "Windy",
        "Clear"
    };

    // Temperature range matching Go implementation: 20.0-45.0°C
    private static final double MIN_TEMP = 20.0;
    private static final double MAX_TEMP = 45.0;

    /**
     * Validate weather data.
     */
    public WeatherData {
        if (temperature < -50.0 || temperature > 60.0) {
            throw new IllegalArgumentException(
                "Temperature out of realistic range: " + temperature
            );
        }
        if (unit == null || unit.isBlank()) {
            throw new IllegalArgumentException("Unit cannot be null or blank");
        }
        if (description == null || description.isBlank()) {
            throw new IllegalArgumentException("Description cannot be null or blank");
        }
        if (city == null || city.isBlank()) {
            throw new IllegalArgumentException("City cannot be null or blank");
        }
        if (timestamp == null || timestamp.isBlank()) {
            throw new IllegalArgumentException("Timestamp cannot be null or blank");
        }
    }

    /**
     * Create simulated weather data for a city.
     * Matches Go implementation's random generation logic.
     *
     * @param city city name
     * @return simulated WeatherData
     */
    public static WeatherData simulate(String city) {
        double temperature = MIN_TEMP + (random.nextDouble() * (MAX_TEMP - MIN_TEMP));
        // Round to 1 decimal place
        temperature = Math.round(temperature * 10.0) / 10.0;

        String condition = WEATHER_CONDITIONS[random.nextInt(WEATHER_CONDITIONS.length)];
        String timestamp = Instant.now().toString();

        return new WeatherData(temperature, "°C", condition, city, timestamp);
    }

    /**
     * Format weather data as human-readable string.
     * Matches Go implementation's format: "Weather in {city}: {temp}°C, {description}"
     *
     * @return formatted string
     */
    public String format() {
        return String.format("Weather in %s: %.1f%s, %s", city, temperature, unit, description);
    }
}
