# MCP Client TLS API Contract

## Client Interface Extensions

### TLS-Enabled Client Interface
```go
type TLSClient interface {
    Call(ctx context.Context, method string, params interface{}) (interface{}, error)
    CallWeather(ctx context.Context, city string) (*models.TemperatureData, error)
    CallDateTime(ctx context.Context, city string) (*models.DateTimeData, error)
    CallEcho(ctx context.Context, text string) (*models.EchoData, error)
    GetConnectionInfo() *TLSConnectionInfo
    ValidateServerCert() error
    Close()
}
```

### Client Constructors
```go
func NewTLSClient(baseURL string, tlsConfig TLSConfig, timeout time.Duration) (*Client, error)
func NewClient(baseURL string, timeout time.Duration) *Client // Existing HTTP client
```

## Configuration Methods

### TLS Client Configuration
```go
type TLSClientConfig struct {
    ServerURL       string        `json:"server_url"`
    UseTLS          bool          `json:"use_tls"`
    TLSConfig       TLSConfig     `json:"tls_config"`
    Timeout         time.Duration `json:"timeout"`
    RetryAttempts   int           `json:"retry_attempts"`
    SkipVerify      bool          `json:"skip_verify"`      // Demo mode
    ServerName      string        `json:"server_name"`      // SNI override
}
```

### Client Connection Test
```go
type ClientConnectionTestRequest struct {
    TargetURL   string        `json:"target_url"`
    TLSConfig   TLSConfig     `json:"tls_config"`
    TestMethod  string        `json:"test_method"`     // "ping", "echo", etc.
    Timeout     time.Duration `json:"timeout"`
}

type ClientConnectionTestResponse struct {
    Success         bool          `json:"success"`
    ResponseTime    time.Duration `json:"response_time"`
    TLSInfo         TLSConnectionInfo `json:"tls_info"`
    ServerResponse  interface{}   `json:"server_response,omitempty"`
    Error           string        `json:"error,omitempty"`
}
```

## Certificate Management

### Client Certificate Validation
```go
type CertificateValidationRequest struct {
    CertificatePath string `json:"certificate_path"`
    PrivateKeyPath  string `json:"private_key_path"`
    CACertPath      string `json:"ca_cert_path"`
}

type CertificateValidationResponse struct {
    Valid           bool      `json:"valid"`
    ExpiresAt       time.Time `json:"expires_at"`
    DaysUntilExpiry int       `json:"days_until_expiry"`
    Subject         string    `json:"subject"`
    Issuer          string    `json:"issuer"`
    Errors          []string  `json:"errors,omitempty"`
}
```

## Connection Monitoring

### Connection Status
```go
type ClientConnectionStatus struct {
    Connected       bool          `json:"connected"`
    ServerURL       string        `json:"server_url"`
    LastCall        time.Time     `json:"last_call"`
    TotalCalls      int           `json:"total_calls"`
    FailedCalls     int           `json:"failed_calls"`
    AverageLatency  time.Duration `json:"average_latency"`
    TLSInfo         *TLSConnectionInfo `json:"tls_info,omitempty"`
}
```

### Connection Pool Info
```go
type ConnectionPoolInfo struct {
    MaxConnections    int `json:"max_connections"`
    ActiveConnections int `json:"active_connections"`
    IdleConnections   int `json:"idle_connections"`
    TotalRequests     int `json:"total_requests"`
}
```

## Error Handling

### Client-Specific TLS Errors
```go
type TLSClientError struct {
    ServerURL string    `json:"server_url"`
    Code      string    `json:"code"`
    Message   string    `json:"message"`
    Retry     bool      `json:"retry"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Error Categories
- `CONNECTION_FAILED`: Failed to establish connection
- `CERT_VERIFICATION_FAILED`: Server certificate verification failed
- `CLIENT_CERT_ERROR`: Client certificate issue
- `TLS_HANDSHAKE_TIMEOUT`: TLS handshake timeout
- `UNSUPPORTED_TLS_VERSION`: Server TLS version not supported
- `CIPHER_SUITE_MISMATCH`: No common cipher suites