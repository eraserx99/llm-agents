# TLS Configuration API Contract

## Configuration Endpoints

### TLS Configuration Structure
```go
type TLSConfig struct {
    CertDir       string `json:"cert_dir"`
    ServerCert    string `json:"server_cert"`
    ServerKey     string `json:"server_key"`
    ClientCert    string `json:"client_cert"`
    ClientKey     string `json:"client_key"`
    CACert        string `json:"ca_cert"`
    DemoMode      bool   `json:"demo_mode"`
    MinTLSVersion uint16 `json:"min_tls_version"`
    Port          int    `json:"port"`
}
```

### Certificate Generation Request
```go
type CertificateGenerationRequest struct {
    CertType     string `json:"cert_type"`     // "server", "client", "ca"
    CommonName   string `json:"common_name"`
    Organization string `json:"organization"`
    Country      string `json:"country"`
    ValidityDays int    `json:"validity_days"`
    KeySize      int    `json:"key_size"`
}
```

### Certificate Generation Response
```go
type CertificateGenerationResponse struct {
    CertificatePath string    `json:"certificate_path"`
    PrivateKeyPath  string    `json:"private_key_path"`
    SerialNumber    string    `json:"serial_number"`
    ExpiresAt       time.Time `json:"expires_at"`
}
```

## TLS Connection Validation

### Connection Test Request
```go
type TLSConnectionTestRequest struct {
    ServerURL  string     `json:"server_url"`
    TLSConfig  TLSConfig  `json:"tls_config"`
    Timeout    int        `json:"timeout_seconds"`
}
```

### Connection Test Response
```go
type TLSConnectionTestResponse struct {
    Success        bool     `json:"success"`
    TLSVersion     string   `json:"tls_version"`
    CipherSuite    string   `json:"cipher_suite"`
    ServerCertCN   string   `json:"server_cert_cn"`
    ClientCertCN   string   `json:"client_cert_cn"`
    Error          string   `json:"error,omitempty"`
    ConnectedAt    time.Time `json:"connected_at"`
}
```

## Error Responses

### Standard Error Format
```go
type TLSError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### Error Codes
- `CERT_NOT_FOUND`: Certificate file not found
- `CERT_INVALID`: Certificate parsing failed
- `CERT_EXPIRED`: Certificate has expired
- `KEY_MISMATCH`: Private key doesn't match certificate
- `TLS_HANDSHAKE_FAILED`: TLS handshake failed
- `CONFIG_INVALID`: Configuration validation failed
- `CONNECTION_TIMEOUT`: Connection timeout occurred