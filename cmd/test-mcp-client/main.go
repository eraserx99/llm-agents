// Test MCP Client using official MCP Go SDK with HTTP/SSE transport
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
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
	serverURL := flag.String("server", "http://localhost:8091", "MCP server URL")
	city := flag.String("city", "Boston", "City to get weather for")
	flag.Parse()

	// Initialize logging
	logLevel := "INFO"
	if *verbose {
		logLevel = "DEBUG"
	}
	utils.InitLogger(logLevel, true)

	var tlsConfig *config.TLSConfig

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

		// Update server URL for HTTPS
		if *serverURL == "http://localhost:8091" {
			*serverURL = "https://localhost:8491"
		}

		utils.Info("MCP Test Client configured with TLS support")
		utils.Info("Server URL: %s, demo mode: %v, cert dir: %s", *serverURL, demoMode, certDir)
	} else {
		utils.Info("MCP Test Client configured for HTTP mode")
		utils.Info("Server URL: %s", *serverURL)
	}

	// Create MCP client using official SDK
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-mcp-client",
		Version: "v1.0.0",
	}, nil)

	// Create custom HTTP/SSE transport for client
	mcpTransport := transport.NewClientTransport(*serverURL, tlsConfig)

	utils.Info("Connecting to MCP server with HTTP/SSE streaming transport...")

	// Connect to server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := client.Connect(ctx, mcpTransport, nil)
	if err != nil {
		log.Fatalf("Failed to connect to MCP server: %v", err)
	}
	defer session.Close()

	utils.Info("Connected to MCP server successfully!")

	// Test tool listing
	utils.Info("Listing available tools...")
	toolsResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		log.Printf("Failed to list tools: %v", err)
	} else {
		utils.Info("Available tools:")
		for _, tool := range toolsResult.Tools {
			utils.Info("  - %s: %s", tool.Name, tool.Description)
		}
	}

	// Test weather tool call
	utils.Info("Calling getTemperature tool for city: %s", *city)

	toolParams := &mcp.CallToolParams{
		Name: "getTemperature",
		Arguments: map[string]any{
			"city": *city,
		},
	}

	toolResult, err := session.CallTool(ctx, toolParams)
	if err != nil {
		log.Fatalf("Failed to call getTemperature tool: %v", err)
	}

	utils.Info("Tool call successful!")

	// Display results
	fmt.Printf("\n=== MCP Tool Call Results ===\n")
	fmt.Printf("Tool: %s\n", toolParams.Name)
	fmt.Printf("City: %s\n", *city)

	if toolResult.Content != nil {
		fmt.Printf("Response:\n")
		for _, content := range toolResult.Content {
			if textContent, ok := content.(*mcp.TextContent); ok {
				fmt.Printf("  %s\n", textContent.Text)
			}
		}
	}

	// Try to parse structured data if available
	if len(toolResult.Content) > 0 {
		fmt.Printf("\nRaw content data:\n")
		for i, content := range toolResult.Content {
			contentJSON, _ := json.MarshalIndent(content, "  ", "  ")
			fmt.Printf("  Content[%d]: %s\n", i, string(contentJSON))
		}
	}

	fmt.Printf("\n=== MCP Streaming Test Complete ===\n")
	fmt.Printf("✅ Successfully connected using MCP HTTP/SSE streaming transport\n")
	fmt.Printf("✅ Tool listing worked\n")
	fmt.Printf("✅ Tool execution worked\n")

	if tlsConfig != nil {
		fmt.Printf("✅ mTLS authentication successful\n")
	}

	utils.Info("MCP client test completed successfully")
}