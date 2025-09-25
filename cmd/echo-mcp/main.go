// Echo MCP Server - provides echo functionality via Model Context Protocol
package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/echo"
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
	httpPort := 8083
	if portStr := os.Getenv("ECHO_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			httpPort = p
		}
	}

	tlsPort := 8445
	if portStr := os.Getenv("ECHO_MCP_TLS_PORT"); portStr != "" {
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

		srv = server.NewTLSServer("echo-mcp", httpPort, tlsPort, tlsConfig)
		utils.Info("Echo MCP Server configured with TLS support")
	} else {
		srv = server.NewServer("echo-mcp", httpPort)
		utils.Info("Echo MCP Server configured for HTTP only")
	}

	// Register echo handler
	echoHandler := echo.NewHandler()
	srv.RegisterHandler("echo", echoHandler)

	// Start server(s)
	if *useTLS {
		if err := srv.StartBoth(); err != nil {
			log.Fatal("Failed to start echo MCP servers:", err)
		}
		utils.Info("Echo MCP Server started with TLS support")
	} else {
		if err := srv.Start(); err != nil {
			log.Fatal("Failed to start echo MCP server:", err)
		}
		utils.Info("Echo MCP Server started (HTTP only)")
	}

	select {}
}
