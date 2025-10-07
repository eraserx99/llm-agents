// Weather MCP Server using official MCP Go SDK with StreamableHTTPHandler
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/steve/llm-agents/internal/config"
	mcptls "github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

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

// responseCapture wraps http.ResponseWriter to capture response data
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body = append(rc.body, b...)
	return rc.ResponseWriter.Write(b)
}

func main() {
	// Parse command line flags
	useTLS := flag.Bool("tls", false, "Enable TLS support")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Initialize logging
	logLevel := "INFO"
	if *verbose {
		logLevel = "DEBUG"
	}
	utils.InitLogger(logLevel, true)

	// Get ports from environment or use defaults
	httpPort := 8081
	if portStr := os.Getenv("WEATHER_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			httpPort = p
		}
	}

	tlsPort := 8443
	if portStr := os.Getenv("WEATHER_MCP_TLS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			tlsPort = p
		}
	}

	// Create MCP server using official SDK
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "weather-mcp",
		Version: "v1.0.0",
	}, nil)

	// Add weather tool using the official SDK's generic AddTool function
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

		callToolResult := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Weather in %s: %.1f%s, %s",
						result.City, result.Temperature, result.Unit, result.Description),
				},
			},
		}

		// Log the complete response structure for debugging
		if resultJSON, err := json.MarshalIndent(map[string]interface{}{
			"callToolResult": callToolResult,
			"structuredData": result,
		}, "", "  "); err == nil {
			utils.Debug("Complete tool response payload:\n%s", string(resultJSON))
		}

		return callToolResult, result, nil
	})

	// Create StreamableHTTPHandler using official SDK
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{JSONResponse: true})

	// Wrap handler to log responses
	loggingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer wrapper to capture the response
		responseWriter := &responseCapture{
			ResponseWriter: w,
			statusCode:     200,
			body:           []byte{},
		}

		handler.ServeHTTP(responseWriter, r)

		// Log the complete HTTP response for debugging
		utils.Debug("HTTP Response Status: %d", responseWriter.statusCode)
		utils.Debug("HTTP Response Body:\n%s", string(responseWriter.body))
	})

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.Handle("/mcp", loggingHandler)

	var tlsConfig *config.TLSConfig

	if *useTLS {
		// TLS mode - configure TLS
		tlsEnabled := os.Getenv("TLS_ENABLED") == "true"
		if !tlsEnabled {
			log.Fatal("TLS flag provided but TLS_ENABLED environment variable not set")
		}

		// Get TLS configuration from environment
		certDir := os.Getenv("TLS_CERT_DIR")
		if certDir == "" {
			certDir = "./certs"
		}

		demoMode := os.Getenv("TLS_DEMO_MODE") == "true"
		tlsConfig = config.NewTLSConfig(certDir, demoMode)

		utils.Info("Weather MCP Server configured with TLS support")
		utils.Info("HTTP port: %d, HTTPS port: %d", httpPort, tlsPort)
		utils.Info("TLS demo mode: %v", demoMode)
		utils.Info("Certificate directory: %s", certDir)
	} else {
		utils.Info("Weather MCP Server configured for HTTP only")
		utils.Info("HTTP port: %d", httpPort)
	}

	// Start HTTP server
	go func() {
		addr := fmt.Sprintf(":%d", httpPort)
		utils.Info("Starting Weather MCP Server (HTTP) on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Start HTTPS server if TLS is enabled
	if *useTLS && tlsConfig != nil {
		go func() {
			addr := fmt.Sprintf(":%d", tlsPort)
			utils.Info("Starting Weather MCP Server (HTTPS) on %s", addr)

			tlsLoader := mcptls.NewTLSLoader(tlsConfig)
			serverTLSConfig, err := tlsLoader.LoadServerTLSConfig()
			if err != nil {
				log.Fatal("Failed to load TLS config:", err)
			}

			server := &http.Server{
				Addr:      addr,
				Handler:   mux,
				TLSConfig: serverTLSConfig,
			}

			if err := server.ListenAndServeTLS("", ""); err != nil {
				log.Fatal("Failed to start HTTPS server:", err)
			}
		}()
	}

	utils.Info("Weather MCP Server started with official SDK StreamableHTTPHandler")

	// Keep the main goroutine alive
	select {}
}