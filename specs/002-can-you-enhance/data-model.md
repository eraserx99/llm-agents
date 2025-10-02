# Phase 1: Data Model - mTLS Enhancement for MCP Servers

**Date**: 2025-09-25
**Feature**: mTLS Enhancement for MCP Servers

## Core Entities

### TLSConfig
Configuration entity for TLS settings across servers and clients.

**Fields**:
- `CertDir` (string): Directory path for certificate storage
- `ServerCert` (string): Path to server certificate file
- `ServerKey` (string): Path to server private key file
- `ClientCert` (string): Path to client certificate file
- `ClientKey` (string): Path to client private key file
- `CACert` (string): Path to Certificate Authority certificate
- `DemoMode` (bool): Enable relaxed validation for demo purposes
- `MinTLSVersion` (uint16): Minimum TLS version (default: TLS 1.2)
- `Port` (int): TLS-enabled port number

**Validation Rules**:
- All certificate paths must be readable files when DemoMode is false
- CertDir must be a valid directory path
- Port must be in valid range (1024-65535)
- MinTLSVersion must be >= TLS 1.2

### Certificate
Represents a TLS certificate with its metadata.

**Fields**:
- `Type` (CertificateType): SERVER, CLIENT, or CA
- `CommonName` (string): Certificate subject common name
- `Organization` (string): Organization name
- `Country` (string): Country code
- `Validity` (Duration): Certificate validity period
- `KeySize` (int): RSA key size (default: 2048)
- `SerialNumber` (int64): Unique certificate serial number

**Validation Rules**:
- CommonName is required and must be valid hostname or CN
- KeySize must be >= 2048 bits
- Validity must be > 0 and <= 10 years
- SerialNumber must be unique within CA

### CertificateType
Enumeration for certificate types.

**Values**:
- `SERVER`: Server authentication certificates
- `CLIENT`: Client authentication certificates
- `CA`: Certificate Authority certificates

### TLSConnection
Represents an active mTLS connection between client and server.

**Fields**:
- `ServerName` (string): Target server hostname
- `Port` (int): Target server port
- `ClientCertificate` (Certificate): Client certificate used for authentication
- `ServerCertificate` (Certificate): Server certificate received
- `TLSVersion` (uint16): Negotiated TLS version
- `CipherSuite` (uint16): Negotiated cipher suite
- `Verified` (bool): Whether mutual authentication succeeded
- `EstablishedAt` (time.Time): Connection establishment timestamp

**State Transitions**:
- `INITIATING` → `HANDSHAKING` → `VERIFIED` (success path)
- `INITIATING` → `HANDSHAKING` → `FAILED` (failure path)
- `VERIFIED` → `CLOSED` (normal termination)

### MCPServerConfig
Extended configuration for MCP servers with TLS support.

**Fields**:
- `Name` (string): Server identifier (weather-mcp, datetime-mcp, echo-mcp)
- `HTTPPort` (int): Traditional HTTP port
- `TLSPort` (int): mTLS-enabled port
- `TLSEnabled` (bool): Whether TLS is active
- `TLSConfig` (TLSConfig): TLS configuration settings
- `Handlers` (map[string]Handler): MCP method handlers

**Validation Rules**:
- Name must be unique across servers
- HTTPPort and TLSPort must be different
- When TLSEnabled is true, TLSConfig must be valid

### MCPClientConfig
Configuration for MCP clients with TLS support.

**Fields**:
- `ServerURL` (string): Target server URL (http:// or https://)
- `UseTLS` (bool): Whether to use TLS connection
- `TLSConfig` (TLSConfig): Client TLS configuration
- `Timeout` (Duration): Connection timeout
- `RetryAttempts` (int): Number of connection retry attempts

**Validation Rules**:
- ServerURL must be valid URL format
- When UseTLS is true, TLSConfig must be valid
- Timeout must be > 0
- RetryAttempts must be >= 0

## Relationships

```
TLSConfig ──┐
           │
           ├─→ MCPServerConfig
           │
           └─→ MCPClientConfig

Certificate ──→ TLSConnection

MCPServerConfig ──→ TLSConnection (server side)
MCPClientConfig ──→ TLSConnection (client side)
```

## Data Flow

### Certificate Generation Flow
1. Generate CA certificate and private key
2. Generate server certificate signed by CA
3. Generate client certificate signed by CA
4. Store all certificates in configured directory
5. Set appropriate file permissions (600 for keys, 644 for certs)

### Connection Establishment Flow
1. Client loads client certificate and CA certificate
2. Client initiates TLS connection to server
3. Server presents server certificate
4. Client validates server certificate against CA
5. Server requests client certificate
6. Client presents client certificate
7. Server validates client certificate against CA
8. TLS handshake completes with mutual authentication
9. JSON-RPC communication proceeds over encrypted channel

### Configuration Loading Flow
1. Load TLS configuration from environment variables
2. Validate certificate file paths and accessibility
3. Create TLS configuration objects for servers and clients
4. Initialize servers with TLS-enabled listeners
5. Initialize clients with TLS-enabled HTTP transports

## Error Handling

### Certificate Errors
- `ErrCertificateNotFound`: Certificate file missing
- `ErrCertificateInvalid`: Certificate parsing failed
- `ErrCertificateExpired`: Certificate past validity period
- `ErrPrivateKeyMismatch`: Private key doesn't match certificate

### TLS Connection Errors
- `ErrTLSHandshakeFailed`: TLS handshake failed
- `ErrClientCertRequired`: Server requires client certificate
- `ErrServerCertInvalid`: Server certificate validation failed
- `ErrUnsupportedTLSVersion`: TLS version not supported

### Configuration Errors
- `ErrInvalidTLSConfig`: TLS configuration validation failed
- `ErrCertificateDirectoryNotFound`: Certificate directory missing
- `ErrPortConflict`: HTTP and TLS ports are the same