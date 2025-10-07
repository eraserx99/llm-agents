// Test MCP Client - Direct test without OpenRouter
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/mcp/client"
	"github.com/steve/llm-agents/internal/utils"
)

func main() {
	// Initialize logging
	utils.InitLogger("DEBUG", true)

	// Configure TLS
	tlsConfig := config.NewTLSConfig("./certs", true)

	// Create weather MCP client
	weatherClient, err := client.NewTLSClient("https://localhost:8443/mcp", 30*time.Second, tlsConfig)
	if err != nil {
		fmt.Printf("Error creating weather client: %v\n", err)
		os.Exit(1)
	}
	defer weatherClient.Close()

	utils.Info("Testing weather MCP client...")

	// Call GetTemperature
	ctx := context.Background()
	tempData, err := weatherClient.CallWeather(ctx, "Chicago")
	if err != nil {
		fmt.Printf("Error calling temperature: %v\n", err)
		os.Exit(1)
	}

	utils.Info("Successfully retrieved weather data: %+v", tempData)
	fmt.Printf("\nWeather in %s: %.1f%s, %s\n", tempData.City, tempData.Temperature, tempData.Unit, tempData.Description)
}
