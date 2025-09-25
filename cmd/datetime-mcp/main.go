// DateTime MCP Server - provides datetime data via Model Context Protocol
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/steve/llm-agents/internal/mcp/datetime"
	"github.com/steve/llm-agents/internal/mcp/server"
	"github.com/steve/llm-agents/internal/utils"
)

func main() {
	// Initialize logging
	utils.InitLogger("INFO", false)

	// Get port from environment or use default
	port := 8082
	if portStr := os.Getenv("DATETIME_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Create server
	srv := server.NewServer("datetime-mcp", port)

	// Register datetime handler
	datetimeHandler := datetime.NewHandler()
	srv.RegisterHandler("getDateTime", datetimeHandler)

	// Start server
	utils.Info("Starting DateTime MCP Server on port %d", port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start datetime MCP server:", err)
	}
}
