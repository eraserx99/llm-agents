// Package server provides the base MCP server implementation
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/steve/llm-agents/internal/models"
)

// Handler interface for MCP method handlers
type Handler interface {
	Handle(ctx context.Context, params json.RawMessage) (interface{}, error)
}

// Server represents an MCP server
type Server struct {
	handlers map[string]Handler
	port     int
	name     string
}

// NewServer creates a new MCP server
func NewServer(name string, port int) *Server {
	return &Server{
		handlers: make(map[string]Handler),
		port:     port,
		name:     name,
	}
}

// RegisterHandler registers a method handler
func (s *Server) RegisterHandler(method string, handler Handler) {
	s.handlers[method] = handler
}

// Start starts the MCP server
func (s *Server) Start() error {
	http.HandleFunc("/rpc", s.handleRPC)
	addr := ":" + strconv.Itoa(s.port)
	log.Printf("[%s] MCP server starting on port %d", s.name, s.port)
	return http.ListenAndServe(addr, nil)
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

// HandlerFunc is a function adapter for Handler interface
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Handle implements the Handler interface
func (f HandlerFunc) Handle(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return f(ctx, params)
}
