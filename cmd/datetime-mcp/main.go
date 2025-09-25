// DateTime MCP Server - provides datetime data via Model Context Protocol
package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/datetime"
	"github.com/steve/llm-agents/internal/mcp/server"
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
	httpPort := 8082
	if portStr := os.Getenv("DATETIME_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			httpPort = p
		}
	}

	tlsPort := 8444
	if portStr := os.Getenv("DATETIME_MCP_TLS_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			tlsPort = p
		}
	}

	var srv *server.Server

	if *useTLS {
		// TLS mode
		if os.Getenv("TLS_ENABLED") != "true" {
			log.Fatal("TLS flag provided but TLS_ENABLED environment variable not set")
		}

		certDir := os.Getenv("TLS_CERT_DIR")
		if certDir == "" {
			certDir = "./certs"
		}

		demoMode := os.Getenv("TLS_DEMO_MODE") == "true"
		tlsConfig := config.NewTLSConfig(certDir, demoMode)
		tlsConfig.Port = tlsPort

		srv = server.NewTLSServer("datetime-mcp", httpPort, tlsPort, tlsConfig)
		utils.Info("DateTime MCP Server configured with TLS support")
	} else {
		srv = server.NewServer("datetime-mcp", httpPort)
		utils.Info("DateTime MCP Server configured for HTTP only")
	}

	// Register datetime handler
	datetimeHandler := datetime.NewHandler()
	srv.RegisterHandler("getDateTime", datetimeHandler)

	// Start server(s)
	if *useTLS {
		if err := srv.StartBoth(); err != nil {
			log.Fatal("Failed to start datetime MCP servers:", err)
		}
		utils.Info("DateTime MCP Server started with TLS support")
	} else {
		if err := srv.Start(); err != nil {
			log.Fatal("Failed to start datetime MCP server:", err)
		}
		utils.Info("DateTime MCP Server started (HTTP only)")
	}

	select {}
}
