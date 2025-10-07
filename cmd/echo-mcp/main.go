// Echo MCP Server using official MCP Go SDK with StreamableHTTPHandler
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/steve/llm-agents/internal/config"
	mcptls "github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

type EchoArgs struct {
	Text string `json:"text" jsonschema:"the text to echo back"`
}

type EchoResult struct {
	OriginalText string `json:"original_text"`
	EchoText     string `json:"echo_text"`
	Timestamp    string `json:"timestamp"`
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

	// Create MCP server using official SDK
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "echo-mcp",
		Version: "v1.0.0",
	}, nil)

	// Add echo tool using the official SDK's generic AddTool function
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echo back the provided text",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args EchoArgs) (*mcp.CallToolResult, EchoResult, error) {
		utils.Info("Handling echo request for text: %s", args.Text)

		result := EchoResult{
			OriginalText: args.Text,
			EchoText:     args.Text,
			Timestamp:    time.Now().Format(time.RFC3339),
		}

		utils.Info("Returning echo data: %+v", result)

		callToolResult := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Echo: %s", result.EchoText),
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

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

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

		utils.Info("Echo MCP Server configured with TLS support")
		utils.Info("HTTP port: %d, HTTPS port: %d", httpPort, tlsPort)
		utils.Info("TLS demo mode: %v", demoMode)
		utils.Info("Certificate directory: %s", certDir)
	} else {
		utils.Info("Echo MCP Server configured for HTTP only")
		utils.Info("HTTP port: %d", httpPort)
	}

	// Start HTTP server
	go func() {
		addr := fmt.Sprintf(":%d", httpPort)
		utils.Info("Starting Echo MCP Server (HTTP) on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal("Failed to start HTTP server:", err)
		}
	}()

	// Start HTTPS server if TLS is enabled
	if *useTLS && tlsConfig != nil {
		go func() {
			addr := fmt.Sprintf(":%d", tlsPort)
			utils.Info("Starting Echo MCP Server (HTTPS) on %s", addr)

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

	utils.Info("Echo MCP Server started with official SDK StreamableHTTPHandler")

	// Keep the main goroutine alive
	select {}
}