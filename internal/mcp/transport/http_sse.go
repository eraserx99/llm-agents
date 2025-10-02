// Package transport provides HTTP/SSE transport for MCP SDK with streaming support
package transport

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/steve/llm-agents/internal/config"
	mcptls "github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

// HTTPSSETransport implements MCP Streaming Protocol over HTTP with SSE support
type HTTPSSETransport struct {
	ServerURL   string
	TLSConfig   *config.TLSConfig
	tlsLoader   *mcptls.TLSLoader
	httpClient  *http.Client
	isClient    bool
	serverPort  int
	mu          sync.RWMutex
}

// NewClientTransport creates a new HTTP/SSE transport for MCP clients
func NewClientTransport(serverURL string, tlsConfig *config.TLSConfig) *HTTPSSETransport {
	transport := &HTTPSSETransport{
		ServerURL: serverURL,
		TLSConfig: tlsConfig,
		isClient:  true,
	}

	if tlsConfig != nil {
		transport.tlsLoader = mcptls.NewTLSLoader(tlsConfig)
		clientTLSConfig, err := transport.tlsLoader.LoadClientTLSConfig("localhost")
		if err == nil {
			transport.httpClient = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig:     clientTLSConfig,
					MaxIdleConns:        10,
					MaxIdleConnsPerHost: 5,
					IdleConnTimeout:     30 * time.Second,
				},
				Timeout: 30 * time.Second,
			}
			utils.Info("MCP HTTP/SSE client transport created with mTLS")
		} else {
			utils.Error("Failed to create TLS client config: %v", err)
			transport.httpClient = &http.Client{Timeout: 30 * time.Second}
		}
	} else {
		transport.httpClient = &http.Client{Timeout: 30 * time.Second}
		utils.Info("MCP HTTP/SSE client transport created without TLS")
	}

	return transport
}

// NewServerTransport creates a new HTTP/SSE transport for MCP servers
func NewServerTransport(port int, tlsConfig *config.TLSConfig) *HTTPSSETransport {
	transport := &HTTPSSETransport{
		TLSConfig:  tlsConfig,
		isClient:   false,
		serverPort: port,
	}

	if tlsConfig != nil {
		transport.tlsLoader = mcptls.NewTLSLoader(tlsConfig)
		utils.Info("MCP HTTP/SSE server transport created with mTLS on port %d", port)
	} else {
		utils.Info("MCP HTTP/SSE server transport created without TLS on port %d", port)
	}

	return transport
}

// Connect implements the mcp.Transport interface
func (t *HTTPSSETransport) Connect(ctx context.Context) (mcp.Connection, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isClient {
		return t.connectClient(ctx)
	}
	return t.connectServer(ctx)
}

// connectClient establishes connection for client mode
func (t *HTTPSSETransport) connectClient(ctx context.Context) (mcp.Connection, error) {
	utils.Info("Connecting MCP client to %s with HTTP/SSE transport", t.ServerURL)

	conn := &HTTPSSEConnection{
		transport:    t,
		isClient:     true,
		serverURL:    t.ServerURL,
		httpClient:   t.httpClient,
		messageQueue: make(chan jsonrpc.Message, 100),
		closeSignal:  make(chan struct{}),
		sessionID:    fmt.Sprintf("client-%d", time.Now().UnixNano()),
	}

	// Start SSE event stream reader
	if err := conn.startSSEReader(ctx); err != nil {
		return nil, fmt.Errorf("failed to start SSE reader: %w", err)
	}

	utils.Info("MCP client connection established with session ID: %s", conn.sessionID)
	return conn, nil
}

// connectServer establishes connection for server mode
func (t *HTTPSSETransport) connectServer(ctx context.Context) (mcp.Connection, error) {
	utils.Info("Starting MCP server HTTP/SSE transport on port %d", t.serverPort)

	conn := &HTTPSSEConnection{
		transport:    t,
		isClient:     false,
		messageQueue: make(chan jsonrpc.Message, 100),
		closeSignal:  make(chan struct{}),
		sessionID:    fmt.Sprintf("server-%d", time.Now().UnixNano()),
		clients:      make(map[string]*SSEClient),
	}

	// Start HTTP server with SSE support
	if err := conn.startHTTPServer(ctx); err != nil {
		return nil, fmt.Errorf("failed to start HTTP server: %w", err)
	}

	utils.Info("MCP server connection established with session ID: %s", conn.sessionID)
	return conn, nil
}

// HTTPSSEConnection implements mcp.Connection for HTTP/SSE transport
type HTTPSSEConnection struct {
	transport    *HTTPSSETransport
	isClient     bool
	serverURL    string
	httpClient   *http.Client
	httpServer   *http.Server
	messageQueue chan jsonrpc.Message
	closeSignal  chan struct{}
	sessionID    string
	clients      map[string]*SSEClient // For server mode
	mu           sync.RWMutex
	closed       bool
}

// SSEClient represents a connected SSE client
type SSEClient struct {
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	Request  *http.Request
	ClientID string
}

// Read implements mcp.Connection.Read
func (c *HTTPSSEConnection) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.closeSignal:
		return nil, mcp.ErrConnectionClosed
	case msg := <-c.messageQueue:
		return msg, nil
	}
}

// Write implements mcp.Connection.Write
func (c *HTTPSSEConnection) Write(ctx context.Context, msg jsonrpc.Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return mcp.ErrConnectionClosed
	}

	if c.isClient {
		return c.writeClientMessage(ctx, msg)
	}
	return c.writeServerMessage(ctx, msg)
}

// writeClientMessage sends message from client to server
func (c *HTTPSSEConnection) writeClientMessage(ctx context.Context, msg jsonrpc.Message) error {
	jsonData, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Use POST for client-to-server messages
	req, err := http.NewRequestWithContext(ctx, "POST", c.serverURL+"/mcp", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server responded with status %d", resp.StatusCode)
	}

	return nil
}

// writeServerMessage sends message from server to client via SSE
func (c *HTTPSSEConnection) writeServerMessage(ctx context.Context, msg jsonrpc.Message) error {
	jsonData, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	c.mu.RLock()
	clients := make([]*SSEClient, 0, len(c.clients))
	for _, client := range c.clients {
		clients = append(clients, client)
	}
	c.mu.RUnlock()

	// Send to all connected SSE clients
	for _, client := range clients {
		sseData := fmt.Sprintf("data: %s\n\n", string(jsonData))
		if _, err := fmt.Fprint(client.Writer, sseData); err != nil {
			utils.Error("Failed to write SSE data to client %s: %v", client.ClientID, err)
			continue
		}
		client.Flusher.Flush()
	}

	return nil
}

// Close implements mcp.Connection.Close
func (c *HTTPSSEConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	close(c.closeSignal)
	close(c.messageQueue)

	if c.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c.httpServer.Shutdown(ctx)
	}

	utils.Info("MCP HTTP/SSE connection closed: %s", c.sessionID)
	return nil
}

// SessionID implements mcp.Connection.SessionID
func (c *HTTPSSEConnection) SessionID() string {
	return c.sessionID
}

// startSSEReader starts reading SSE events for client mode
func (c *HTTPSSEConnection) startSSEReader(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverURL+"/sse", nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("SSE endpoint responded with status %d", resp.StatusCode)
	}

	go func() {
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "" {
					continue
				}

				msg, err := jsonrpc.DecodeMessage([]byte(data))
				if err != nil {
					utils.Error("Failed to parse SSE message: %v", err)
					continue
				}

				select {
				case c.messageQueue <- msg:
				case <-c.closeSignal:
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			utils.Error("SSE scanner error: %v", err)
		}
	}()

	return nil
}

// startHTTPServer starts HTTP server with SSE support for server mode
func (c *HTTPSSEConnection) startHTTPServer(ctx context.Context) error {
	mux := http.NewServeMux()

	// MCP endpoint for receiving messages
	mux.HandleFunc("/mcp", c.handleMCPRequest)

	// SSE endpoint for sending messages to clients
	mux.HandleFunc("/sse", c.handleSSERequest)

	var tlsConfig *tls.Config
	if c.transport.TLSConfig != nil {
		var err error
		tlsConfig, err = c.transport.tlsLoader.LoadServerTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to load TLS config: %w", err)
		}
	}

	c.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", c.transport.serverPort),
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		var err error
		if tlsConfig != nil {
			err = c.httpServer.ListenAndServeTLS("", "")
		} else {
			err = c.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			utils.Error("HTTP server error: %v", err)
		}
	}()

	return nil
}

// handleMCPRequest handles incoming MCP messages
func (c *HTTPSSEConnection) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the body first
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Decode the message
	msg, err := jsonrpc.DecodeMessage(body)
	if err != nil {
		http.Error(w, "Invalid JSON-RPC message", http.StatusBadRequest)
		return
	}

	select {
	case c.messageQueue <- msg:
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	case <-c.closeSignal:
		http.Error(w, "Connection closed", http.StatusServiceUnavailable)
	default:
		http.Error(w, "Message queue full", http.StatusServiceUnavailable)
	}
}

// handleSSERequest handles SSE connections from clients
func (c *HTTPSSEConnection) handleSSERequest(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	client := &SSEClient{
		Writer:   w,
		Flusher:  flusher,
		Request:  r,
		ClientID: clientID,
	}

	c.mu.Lock()
	c.clients[clientID] = client
	c.mu.Unlock()

	utils.Info("SSE client connected: %s", clientID)

	// Send initial connection message
	fmt.Fprintf(w, "data: {\"type\":\"connection\",\"clientId\":\"%s\"}\n\n", clientID)
	flusher.Flush()

	// Keep connection alive until client disconnects
	<-r.Context().Done()

	c.mu.Lock()
	delete(c.clients, clientID)
	c.mu.Unlock()

	utils.Info("SSE client disconnected: %s", clientID)
}