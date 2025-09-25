// Package config provides configuration management for the multi-agent system
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// OpenRouter API configuration
	OpenRouterAPIKey string

	// MCP Server URLs
	WeatherMCPURL  string
	DateTimeMCPURL string
	EchoMCPURL     string

	// Timeouts
	QueryTimeout time.Duration
	MCPTimeout   time.Duration
	LLMTimeout   time.Duration

	// Logging
	LogLevel string
	Verbose  bool

	// CLI specific
	City string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	config := &Config{
		// OpenRouter API
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),

		// MCP Server URLs
		WeatherMCPURL:  getEnv("MCP_WEATHER_URL", "http://localhost:8081"),
		DateTimeMCPURL: getEnv("MCP_DATETIME_URL", "http://localhost:8082"),
		EchoMCPURL:     getEnv("MCP_ECHO_URL", "http://localhost:8083"),

		// Timeouts
		QueryTimeout: getDurationEnv("QUERY_TIMEOUT", 30*time.Second),
		MCPTimeout:   getDurationEnv("MCP_TIMEOUT", 10*time.Second),
		LLMTimeout:   getDurationEnv("LLM_TIMEOUT", 15*time.Second),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "INFO"),
		Verbose:  getBoolEnv("VERBOSE", false),

		// CLI
		City: getEnv("DEFAULT_CITY", ""),
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.OpenRouterAPIKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY is required")
	}

	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
