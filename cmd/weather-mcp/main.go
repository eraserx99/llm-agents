// Weather MCP Server - provides temperature data via Model Context Protocol
package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/server"
	"github.com/steve/llm-agents/internal/mcp/weather"
	"github.com/steve/llm-agents/internal/utils"
)

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

	var srv *server.Server

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

		// Create TLS configuration
		tlsConfig := config.NewTLSConfig(certDir, demoMode)
		tlsConfig.Port = tlsPort

		// Create TLS-enabled server
		srv = server.NewTLSServer("weather-mcp", httpPort, tlsPort, tlsConfig)
		utils.Info("Weather MCP Server configured with TLS support")
		utils.Info("HTTP port: %d, HTTPS port: %d", httpPort, tlsPort)
		utils.Info("TLS demo mode: %v", demoMode)
		utils.Info("Certificate directory: %s", certDir)
	} else {
		// HTTP-only mode
		srv = server.NewServer("weather-mcp", httpPort)
		utils.Info("Weather MCP Server configured for HTTP only")
		utils.Info("HTTP port: %d", httpPort)
	}

	// Register weather handler
	weatherHandler := weather.NewHandler()
	srv.RegisterHandler("getTemperature", weatherHandler)

	// Start server(s)
	if *useTLS {
		// Start both HTTP and HTTPS servers
		if err := srv.StartBoth(); err != nil {
			log.Fatal("Failed to start weather MCP servers:", err)
		}
		utils.Info("Weather MCP Server started with both HTTP and HTTPS support")
	} else {
		// Start HTTP server only
		if err := srv.Start(); err != nil {
			log.Fatal("Failed to start weather MCP server:", err)
		}
		utils.Info("Weather MCP Server started (HTTP only)")
	}

	// Keep the main goroutine alive
	select {}
}
