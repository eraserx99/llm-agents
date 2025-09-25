// Weather MCP Server - provides temperature data via Model Context Protocol
package main

import (
	"log"
	"os"
	"strconv"

	"github.com/steve/llm-agents/internal/mcp/server"
	"github.com/steve/llm-agents/internal/mcp/weather"
	"github.com/steve/llm-agents/internal/utils"
)

func main() {
	// Initialize logging
	utils.InitLogger("INFO", false)

	// Get port from environment or use default
	port := 8081
	if portStr := os.Getenv("WEATHER_MCP_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Create server
	srv := server.NewServer("weather-mcp", port)

	// Register weather handler
	weatherHandler := weather.NewHandler()
	srv.RegisterHandler("getTemperature", weatherHandler)

	// Start server
	utils.Info("Starting Weather MCP Server on port %d", port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start weather MCP server:", err)
	}
}
