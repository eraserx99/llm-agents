// Echo MCP Server - provides echo functionality via Model Context Protocol
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/steve/llm-agents/internal/mcp/echo"
	"github.com/steve/llm-agents/internal/mcp/server"
	"github.com/steve/llm-agents/internal/utils"
)

func main() {
	// Initialize logging
	utils.InitLogger("INFO", false)

	// Get port from environment or use default
	port := 8083
	if portStr := os.Getenv("ECHO_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Create server
	srv := server.NewServer("echo-mcp", port)

	// Register echo handler
	echoHandler := echo.NewHandler()
	srv.RegisterHandler("echo", echoHandler)

	// Start server
	utils.Info("Starting Echo MCP Server on port %d", port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start echo MCP server:", err)
	}
}
