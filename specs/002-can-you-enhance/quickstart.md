# Quick Start: mTLS Enhancement for MCP Servers

**Date**: 2025-09-25
**Feature**: mTLS Enhancement for MCP Servers
**Estimated Setup Time**: 10 minutes

## Prerequisites
- Go 1.25.1+ installed
- Access to the llm-agents repository
- Network ports 8081, 8082, 8083 available for HTTP
- Network ports 8443, 8444, 8445 available for HTTPS/TLS

## Quick Setup Steps

### 1. Generate Self-Signed Certificates (2 minutes)
```bash
# Navigate to project root
cd llm-agents

# Generate certificates using built-in tool
go run cmd/cert-gen/main.go --demo-mode

# Verify certificates were created
ls -la certs/
# Should show: ca.crt, ca.key, server.crt, server.key, client.crt, client.key
```

### 2. Configure Environment Variables (1 minute)
```bash
# Set environment variables for TLS configuration
export TLS_ENABLED=true
export TLS_DEMO_MODE=true
export TLS_CERT_DIR=./certs
export WEATHER_MCP_TLS_PORT=8443
export DATETIME_MCP_TLS_PORT=8444
export ECHO_MCP_TLS_PORT=8445
```

### 3. Start TLS-Enabled MCP Servers (1 minute)
```bash
# Terminal 1: Start weather MCP server with TLS
go run cmd/weather-mcp/main.go --tls

# Terminal 2: Start datetime MCP server with TLS
go run cmd/datetime-mcp/main.go --tls

# Terminal 3: Start echo MCP server with TLS
go run cmd/echo-mcp/main.go --tls
```

### 4. Test TLS Connection (2 minutes)
```bash
# Test TLS connectivity
go run cmd/test-tls/main.go --server weather --demo-mode

# Expected output:
# ✅ TLS connection established
# ✅ Mutual authentication successful
# ✅ Weather data retrieved securely
# TLS Version: TLS 1.3
# Cipher Suite: TLS_AES_256_GCM_SHA384
# Server Certificate: CN=weather-mcp
# Client Certificate: CN=mcp-client
```

### 5. Run Coordinator with TLS (2 minutes)
```bash
# Start the main coordinator with TLS support
go run cmd/main/main.go --use-tls

# Test query
echo "What's the weather in New York?" | go run cmd/main/main.go --use-tls
```

### 6. Verify Security (2 minutes)
```bash
# Try connecting without client certificate (should fail)
curl -k https://localhost:8443/rpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getTemperature","params":{"city":"Boston"},"id":1}'

# Expected: Connection refused or TLS handshake failure

# Try with HTTP on TLS port (should fail)
curl http://localhost:8443/rpc

# Expected: Connection refused
```

## Validation Checklist

### ✅ Certificate Generation
- [ ] CA certificate created (`certs/ca.crt`)
- [ ] Server certificates created for all 3 servers
- [ ] Client certificate created
- [ ] Private keys have correct permissions (600)
- [ ] Certificates have correct permissions (644)

### ✅ Server Startup
- [ ] Weather MCP server starts on TLS port 8443
- [ ] DateTime MCP server starts on TLS port 8444
- [ ] Echo MCP server starts on TLS port 8445
- [ ] All servers log "TLS enabled" message
- [ ] No certificate loading errors in logs

### ✅ Client Connection
- [ ] Coordinator connects successfully to all TLS servers
- [ ] TLS handshake completes within 5 seconds
- [ ] Mutual authentication succeeds
- [ ] JSON-RPC calls work over TLS connections
- [ ] Response times < 100ms additional overhead

### ✅ Security Validation
- [ ] Connections without client certificates are rejected
- [ ] HTTP requests to TLS ports are rejected
- [ ] Invalid certificates are rejected
- [ ] Demo mode allows self-signed certificates
- [ ] TLS version is 1.2 or higher

## Troubleshooting

### Common Issues

**"Certificate not found" error:**
```bash
# Ensure certificates were generated
ls -la certs/
# If missing, re-run certificate generation
go run cmd/cert-gen/main.go --demo-mode
```

**"TLS handshake failed" error:**
```bash
# Check if demo mode is enabled
echo $TLS_DEMO_MODE
# Should be "true" for development

# Verify certificate paths
echo $TLS_CERT_DIR
# Should point to certs directory
```

**"Port already in use" error:**
```bash
# Check if ports are available
lsof -i :8443
lsof -i :8444
lsof -i :8445

# Kill existing processes if needed
pkill -f "weather-mcp"
```

**Connection timeout:**
```bash
# Verify servers are listening on TLS ports
netstat -an | grep 844[3-5]

# Check firewall settings
# Ensure ports 8443-8445 are not blocked
```

## Performance Verification

### Expected Metrics
- **TLS Handshake Time**: < 10ms
- **Request Latency Overhead**: < 5ms
- **Memory Usage Increase**: < 10MB per server
- **CPU Overhead**: < 5% under normal load

### Performance Test
```bash
# Run performance comparison
go run cmd/perf-test/main.go --compare-http-tls

# Expected output shows minimal overhead:
# HTTP Average: 2.5ms
# TLS Average: 7.2ms
# Overhead: 4.7ms (188%)
```

## Next Steps

After successful quickstart:
1. Review security logs: `tail -f logs/security.log`
2. Monitor certificate expiration: `go run cmd/cert-check/main.go`
3. Configure production settings: Update environment variables
4. Set up certificate renewal: Schedule certificate regeneration
5. Enable strict validation: Set `TLS_DEMO_MODE=false` for production

## Support

For issues:
1. Check logs in `logs/` directory
2. Verify certificate validity with `openssl x509 -in certs/server.crt -text -noout`
3. Test connectivity with `openssl s_client -connect localhost:8443 -cert certs/client.crt -key certs/client.key`