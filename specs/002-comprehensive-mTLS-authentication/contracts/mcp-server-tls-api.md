# MCP Server TLS API Contract

## Server Interface Extensions

### TLS-Enabled Server Interface
```go
type TLSServer interface {
    Start() error                    // Start HTTP server
    StartTLS(config TLSConfig) error // Start TLS server
    RegisterHandler(method string, handler Handler)
    GetTLSConfig() *TLSConfig
    IsSecure() bool
}
```

### Server Constructor
```go
func NewTLSServer(name string, httpPort, tlsPort int, tlsConfig TLSConfig) *Server
```

## Configuration Methods

### TLS Configuration Update
```go
type UpdateTLSConfigRequest struct {
    ServerName string    `json:"server_name"`
    TLSConfig  TLSConfig `json:"tls_config"`
}

type UpdateTLSConfigResponse struct {
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
}
```

### Server Status Check
```go
type ServerStatusResponse struct {
    ServerName  string `json:"server_name"`
    HTTPPort    int    `json:"http_port"`
    TLSPort     int    `json:"tls_port"`
    TLSEnabled  bool   `json:"tls_enabled"`
    Secure      bool   `json:"secure"`
    Uptime      string `json:"uptime"`
    ActiveConns int    `json:"active_connections"`
}
```

## TLS-Specific Endpoints

### Certificate Information
```go
type CertificateInfoResponse struct {
    Subject     string    `json:"subject"`
    Issuer      string    `json:"issuer"`
    SerialNumber string   `json:"serial_number"`
    NotBefore   time.Time `json:"not_before"`
    NotAfter    time.Time `json:"not_after"`
    IsCA        bool      `json:"is_ca"`
    KeyUsage    []string  `json:"key_usage"`
}
```

### TLS Connection Info
```go
type TLSConnectionInfo struct {
    RemoteAddr       string `json:"remote_addr"`
    TLSVersion       string `json:"tls_version"`
    CipherSuite      string `json:"cipher_suite"`
    ClientCertCN     string `json:"client_cert_cn"`
    HandshakeComplete bool  `json:"handshake_complete"`
    EstablishedAt    time.Time `json:"established_at"`
}
```

## Error Handling

### TLS-Specific Errors
```go
type TLSServerError struct {
    ServerName string `json:"server_name"`
    Code       string `json:"code"`
    Message    string `json:"message"`
    Timestamp  time.Time `json:"timestamp"`
}
```

### Error Categories
- `TLS_CONFIG_ERROR`: Invalid TLS configuration
- `CERT_LOAD_ERROR`: Failed to load certificates
- `TLS_BIND_ERROR`: Failed to bind to TLS port
- `CLIENT_AUTH_ERROR`: Client certificate authentication failed
- `HANDSHAKE_ERROR`: TLS handshake failed