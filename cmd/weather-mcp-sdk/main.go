// Weather MCP Server using official MCP Go SDK with HTTP/SSE transport
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/transport"
	"github.com/steve/llm-agents/internal/utils"
)

func main() {
	// Parse command line flags
	useTLS := flag.Bool("tls", false, "Enable TLS support")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	port := flag.Int("port", 8091, "HTTP port for the server")
	tlsPort := flag.Int("tls-port", 8491, "HTTPS port for the server (when TLS enabled)")
	flag.Parse()

	// Initialize logging
	logLevel := "INFO"
	if *verbose {
		logLevel = "DEBUG"
	}
	utils.InitLogger(logLevel, true)

	// Get ports from environment if specified
	if portStr := os.Getenv("WEATHER_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			*port = p
		}
	}
	if portStr := os.Getenv("WEATHER_MCP_TLS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			*tlsPort = p
		}
	}

	var tlsConfig *config.TLSConfig
	selectedPort := *port

	if *useTLS {
		// TLS mode - configure TLS
		tlsEnabled := os.Getenv("TLS_ENABLED") == "true"
		if !tlsEnabled {
			log.Fatal("TLS flag provided but TLS_ENABLED environment variable not set")
		}

		certDir := os.Getenv("TLS_CERT_DIR")
		if certDir == "" {
			certDir = "./certs"
		}

		demoMode := os.Getenv("TLS_DEMO_MODE") == "true"
		tlsConfig = config.NewTLSConfig(certDir, demoMode)
		selectedPort = *tlsPort

		utils.Info("Weather MCP Server (SDK) configured with TLS support")
		utils.Info("TLS port: %d, demo mode: %v, cert dir: %s", selectedPort, demoMode, certDir)
	} else {
		utils.Info("Weather MCP Server (SDK) configured for HTTP only on port %d", selectedPort)
	}

	// Create MCP server using official SDK
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "weather-mcp-sdk",
		Version: "v1.0.0",
	}, nil)

	// Add weather tool using the official SDK's generic AddTool function
	type WeatherArgs struct {
		City string `json:"city" jsonschema:"the city to get weather for"`
	}

	type WeatherResult struct {
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
		Description string  `json:"description"`
		City        string  `json:"city"`
		Timestamp   string  `json:"timestamp"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "getTemperature",
		Description: "Get current temperature and weather conditions for a city",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args WeatherArgs) (*mcp.CallToolResult, WeatherResult, error) {
		utils.Info("Handling getTemperature request for city: %s", args.City)

		// Simulate weather data (in real implementation, call actual weather API)
		temperature := 20.0 + rand.Float64()*25.0 // 20-45°C
		conditions := []string{"Sunny", "Partly cloudy", "Cloudy", "Light rain", "Clear"}
		description := conditions[rand.Intn(len(conditions))]

		result := WeatherResult{
			Temperature: temperature,
			Unit:        "°C",
			Description: description,
			City:        args.City,
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		utils.Info("Returning weather data: %+v", result)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Weather in %s: %.1f%s, %s",
						result.City, result.Temperature, result.Unit, result.Description),
				},
			},
		}, result, nil
	})

	// Create custom HTTP/SSE transport
	mcpTransport := transport.NewServerTransport(selectedPort, tlsConfig)

	// Run server with custom transport
	utils.Info("Starting Weather MCP Server (SDK) with HTTP/SSE streaming transport...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := server.Run(ctx, mcpTransport); err != nil {
		log.Fatalf("Failed to start weather MCP server: %v", err)
	}
}