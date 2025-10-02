// Package server provides the base MCP server implementation
package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/models"
	mcptls "github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

// Handler interface for MCP method handlers
type Handler interface {
	Handle(ctx context.Context, params json.RawMessage) (interface{}, error)
}

// Server represents an MCP server with optional TLS support
type Server struct {
	handlers   map[string]Handler
	port       int
	tlsPort    int
	name       string
	tlsConfig  *config.TLSConfig
	tlsLoader  *mcptls.TLSLoader
	httpServer *http.Server
	tlsServer  *http.Server
	mu         sync.RWMutex
	started    bool
}

// NewServer creates a new MCP server
func NewServer(name string, port int) *Server {
	return &Server{
		handlers: make(map[string]Handler),
		port:     port,
		name:     name,
	}
}

// NewTLSServer creates a new MCP server with TLS support
func NewTLSServer(name string, httpPort, tlsPort int, tlsConfig *config.TLSConfig) *Server {
	server := &Server{
		handlers:  make(map[string]Handler),
		port:      httpPort,
		tlsPort:   tlsPort,
		name:      name,
		tlsConfig: tlsConfig,
	}

	if tlsConfig != nil {
		server.tlsLoader = mcptls.NewTLSLoader(tlsConfig)
	}

	return server
}

// RegisterHandler registers a method handler
func (s *Server) RegisterHandler(method string, handler Handler) {
	s.handlers[method] = handler
}

// Start starts the HTTP MCP server
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", s.handleRPC)

	s.httpServer = &http.Server{
		Addr:         ":" + strconv.Itoa(s.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	s.started = true
	utils.Info("[%s] HTTP MCP server starting on port %d", s.name, s.port)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Error("[%s] HTTP server error: %v", s.name, err)
		}
	}()

	return nil
}

// StartTLS starts the HTTPS MCP server with mutual TLS
func (s *Server) StartTLS() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tlsConfig == nil {
		return fmt.Errorf("TLS configuration not provided")
	}

	if s.tlsLoader == nil {
		return fmt.Errorf("TLS loader not initialized")
	}

	// Validate TLS configuration
	if err := s.tlsConfig.Validate(); err != nil {
		return fmt.Errorf("invalid TLS configuration: %w", err)
	}

	// Load TLS configuration
	tlsConfig, err := s.tlsLoader.LoadServerTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to load TLS configuration: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/rpc", s.handleTLSRPC)

	s.tlsServer = &http.Server{
		Addr:         ":" + strconv.Itoa(s.tlsPort),
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	utils.Info("[%s] HTTPS MCP server starting on port %d (TLS enabled)", s.name, s.tlsPort)

	go func() {
		if err := s.tlsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			utils.Error("[%s] TLS server error: %v", s.name, err)
		}
	}()

	return nil
}

// StartBoth starts both HTTP and HTTPS servers
func (s *Server) StartBoth() error {
	// Start HTTP server
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start HTTPS server if TLS is configured
	if s.tlsConfig != nil {
		if err := s.StartTLS(); err != nil {
			return fmt.Errorf("failed to start TLS server: %w", err)
		}
	}

	return nil
}

// Stop stops all servers
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errors []error

	// Stop HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("HTTP server shutdown error: %w", err))
		}
		s.httpServer = nil
	}

	// Stop TLS server
	if s.tlsServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.tlsServer.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("TLS server shutdown error: %w", err))
		}
		s.tlsServer = nil
	}

	s.started = false

	if len(errors) > 0 {
		return fmt.Errorf("server shutdown errors: %v", errors)
	}

	utils.Info("[%s] Server stopped", s.name)
	return nil
}

// handleRPC handles JSON-RPC requests
func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var request struct {
		JSONRpc string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      int             `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.sendError(w, -32700, "Parse error", 0)
		return
	}

	if request.JSONRpc != "2.0" {
		s.sendError(w, -32600, "Invalid Request", request.ID)
		return
	}

	handler, exists := s.handlers[request.Method]
	if !exists {
		s.sendError(w, -32601, "Method not found", request.ID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := handler.Handle(ctx, request.Params)
	if err != nil {
		s.sendError(w, -32603, err.Error(), request.ID)
		return
	}

	response := struct {
		JSONRpc string      `json:"jsonrpc"`
		Result  interface{} `json:"result"`
		ID      int         `json:"id"`
	}{
		JSONRpc: "2.0",
		Result:  result,
		ID:      request.ID,
	}

	json.NewEncoder(w).Encode(response)
}

// sendError sends an error response
func (s *Server) sendError(w http.ResponseWriter, code int, message string, id int) {
	response := struct {
		JSONRpc string           `json:"jsonrpc"`
		Error   *models.MCPError `json:"error"`
		ID      int              `json:"id"`
	}{
		JSONRpc: "2.0",
		Error: &models.MCPError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	w.WriteHeader(http.StatusOK) // JSON-RPC errors are still HTTP 200
	json.NewEncoder(w).Encode(response)
}

// handleTLSRPC handles JSON-RPC requests over TLS with connection logging
func (s *Server) handleTLSRPC(w http.ResponseWriter, r *http.Request) {
	// Log TLS connection information
	if r.TLS != nil && s.tlsLoader != nil {
		if tlsConn, ok := r.Context().Value("tls-conn").(*tls.Conn); ok {
			connInfo, err := s.tlsLoader.GetTLSConnectionInfo(tlsConn)
			if err == nil {
				utils.Debug("[%s] TLS connection from %s (TLS %s, %s)", s.name, connInfo.RemoteAddr, connInfo.TLSVersion, connInfo.CipherSuite)
				if connInfo.ClientCertCN != "" {
					utils.Debug("[%s] Client certificate: %s", s.name, connInfo.ClientCertCN)
				}
			}
		}
	}

	// Handle the request same as HTTP
	s.handleRPC(w, r)
}

// IsSecure returns true if the server has TLS enabled
func (s *Server) IsSecure() bool {
	return s.tlsConfig != nil
}

// GetTLSConfig returns the TLS configuration (read-only)
func (s *Server) GetTLSConfig() *config.TLSConfig {
	if s.tlsConfig != nil {
		// Return a copy to prevent external modification
		configCopy := *s.tlsConfig
		return &configCopy
	}
	return nil
}

// ServerStatus represents the current server status
type ServerStatus struct {
	ServerName  string `json:"server_name"`
	HTTPPort    int    `json:"http_port"`
	TLSPort     int    `json:"tls_port"`
	TLSEnabled  bool   `json:"tls_enabled"`
	Secure      bool   `json:"secure"`
	Started     bool   `json:"started"`
	ActiveConns int    `json:"active_connections"`
}

// GetStatus returns the current server status
func (s *Server) GetStatus() *ServerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &ServerStatus{
		ServerName: s.name,
		HTTPPort:   s.port,
		TLSPort:    s.tlsPort,
		TLSEnabled: s.tlsConfig != nil,
		Secure:     s.IsSecure(),
		Started:    s.started,
		// ActiveConns would require additional connection tracking
		ActiveConns: 0,
	}
}

// HandlerFunc is a function adapter for Handler interface
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Handle implements the Handler interface
func (f HandlerFunc) Handle(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return f(ctx, params)
}
