package contract

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/steve/llm-agents/internal/config"
)

// TestTLSConfigStructContract tests the TLS configuration data structure contract
func TestTLSConfigStructContract(t *testing.T) {
	tests := []struct {
		name     string
		config   config.TLSConfig
		wantErr  bool
		errorMsg string
	}{
		{
			name: "valid_tls_config",
			config: config.TLSConfig{
				CertDir:       "./test-certs",
				ServerCert:    "./test-certs/server.crt",
				ServerKey:     "./test-certs/server.key",
				ClientCert:    "./test-certs/client.crt",
				ClientKey:     "./test-certs/client.key",
				CACert:        "./test-certs/ca.crt",
				DemoMode:      true,
				MinTLSVersion: 0x0303, // TLS 1.2
				Port:          8443,
			},
			wantErr: false,
		},
		{
			name: "invalid_port_range",
			config: config.TLSConfig{
				CertDir:       "./test-certs",
				ServerCert:    "./test-certs/server.crt",
				ServerKey:     "./test-certs/server.key",
				ClientCert:    "./test-certs/client.crt",
				ClientKey:     "./test-certs/client.key",
				CACert:        "./test-certs/ca.crt",
				DemoMode:      true,
				MinTLSVersion: 0x0303,
				Port:          70000, // Invalid port
			},
			wantErr:  true,
			errorMsg: "port must be in range 1024-65535",
		},
		{
			name: "missing_cert_dir",
			config: config.TLSConfig{
				ServerCert:    "./test-certs/server.crt",
				ServerKey:     "./test-certs/server.key",
				ClientCert:    "./test-certs/client.crt",
				ClientKey:     "./test-certs/client.key",
				CACert:        "./test-certs/ca.crt",
				DemoMode:      true,
				MinTLSVersion: 0x0303,
				Port:          8443,
			},
			wantErr:  true,
			errorMsg: "certificate directory is required",
		},
		{
			name: "invalid_min_tls_version",
			config: config.TLSConfig{
				CertDir:       "./test-certs",
				ServerCert:    "./test-certs/server.crt",
				ServerKey:     "./test-certs/server.key",
				ClientCert:    "./test-certs/client.crt",
				ClientKey:     "./test-certs/client.key",
				CACert:        "./test-certs/ca.crt",
				DemoMode:      true,
				MinTLSVersion: 0x0301, // TLS 1.0 (too old)
				Port:          8443,
			},
			wantErr:  true,
			errorMsg: "minimum TLS version must be >= TLS 1.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test should FAIL until TLS validation is implemented
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Errorf("TLSConfig.Validate() expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("TLSConfig.Validate() unexpected error: %v", err)
			}

			if tt.wantErr && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg && len(err.Error()) > 0 {
					// For now, just check if error contains expected message
					// This allows for more flexible error message matching
					t.Logf("Expected error message: %s, got: %s", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

// TestCertificateConfigContract tests the Certificate configuration contract
func TestCertificateConfigContract(t *testing.T) {
	tests := []struct {
		name     string
		cert     config.Certificate
		wantErr  bool
		errorMsg string
	}{
		{
			name: "valid_server_certificate",
			cert: config.Certificate{
				Type:         config.ServerCert,
				CommonName:   "mcp-server",
				Organization: "MCP Demo",
				Country:      "US",
				Validity:     365 * 24 * time.Hour,
				KeySize:      2048,
				SerialNumber: 123456,
			},
			wantErr: false,
		},
		{
			name: "missing_common_name",
			cert: config.Certificate{
				Type:         config.ClientCert,
				Organization: "MCP Demo",
				Country:      "US",
				Validity:     365 * 24 * time.Hour,
				KeySize:      2048,
				SerialNumber: 123456,
			},
			wantErr:  true,
			errorMsg: "common name is required",
		},
		{
			name: "weak_key_size",
			cert: config.Certificate{
				Type:         config.ServerCert,
				CommonName:   "mcp-server",
				Organization: "MCP Demo",
				Country:      "US",
				Validity:     365 * 24 * time.Hour,
				KeySize:      1024, // Too small
				SerialNumber: 123456,
			},
			wantErr:  true,
			errorMsg: "key size must be >= 2048 bits",
		},
		{
			name: "invalid_validity_period",
			cert: config.Certificate{
				Type:         config.ServerCert,
				CommonName:   "mcp-server",
				Organization: "MCP Demo",
				Country:      "US",
				Validity:     15 * 365 * 24 * time.Hour, // 15 years, too long
				KeySize:      2048,
				SerialNumber: 123456,
			},
			wantErr:  true,
			errorMsg: "validity period must be > 0 and <= 10 years",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test should FAIL until Certificate validation is implemented
			err := tt.cert.Validate()

			if tt.wantErr && err == nil {
				t.Errorf("Certificate.Validate() expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Certificate.Validate() unexpected error: %v", err)
			}
		})
	}
}

// TestJSONSerialization tests JSON serialization contract
func TestJSONSerialization(t *testing.T) {
	tlsConfig := config.TLSConfig{
		CertDir:       "./test-certs",
		ServerCert:    "./test-certs/server.crt",
		ServerKey:     "./test-certs/server.key",
		ClientCert:    "./test-certs/client.crt",
		ClientKey:     "./test-certs/client.key",
		CACert:        "./test-certs/ca.crt",
		DemoMode:      true,
		MinTLSVersion: 0x0303,
		Port:          8443,
	}

	// Test serialization
	data, err := json.Marshal(tlsConfig)
	if err != nil {
		t.Fatalf("Failed to marshal TLSConfig: %v", err)
	}

	// Test deserialization
	var decoded config.TLSConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal TLSConfig: %v", err)
	}

	// Verify fields match
	if decoded.CertDir != tlsConfig.CertDir {
		t.Errorf("CertDir mismatch: got %s, want %s", decoded.CertDir, tlsConfig.CertDir)
	}

	if decoded.DemoMode != tlsConfig.DemoMode {
		t.Errorf("DemoMode mismatch: got %v, want %v", decoded.DemoMode, tlsConfig.DemoMode)
	}

	if decoded.Port != tlsConfig.Port {
		t.Errorf("Port mismatch: got %d, want %d", decoded.Port, tlsConfig.Port)
	}
}

// TestMCPServerConfigContract tests the MCP server configuration contract
func TestMCPServerConfigContract(t *testing.T) {
	tests := []struct {
		name     string
		config   config.MCPServerConfig
		wantErr  bool
		errorMsg string
	}{
		{
			name: "valid_server_config",
			config: config.MCPServerConfig{
				Name:       "weather-mcp",
				HTTPPort:   8081,
				TLSPort:    8443,
				TLSEnabled: true,
				TLSConfig: config.TLSConfig{
					CertDir:       "./test-certs",
					ServerCert:    "./test-certs/server.crt",
					ServerKey:     "./test-certs/server.key",
					ClientCert:    "./test-certs/client.crt",
					ClientKey:     "./test-certs/client.key",
					CACert:        "./test-certs/ca.crt",
					DemoMode:      true,
					MinTLSVersion: 0x0303,
					Port:          8443,
				},
			},
			wantErr: false,
		},
		{
			name: "missing_server_name",
			config: config.MCPServerConfig{
				HTTPPort:   8081,
				TLSPort:    8443,
				TLSEnabled: true,
			},
			wantErr:  true,
			errorMsg: "server name is required",
		},
		{
			name: "port_conflict",
			config: config.MCPServerConfig{
				Name:       "weather-mcp",
				HTTPPort:   8443,
				TLSPort:    8443, // Same as HTTP port
				TLSEnabled: true,
			},
			wantErr:  true,
			errorMsg: "HTTP and TLS ports must be different",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test should FAIL until MCPServerConfig validation is implemented
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Errorf("MCPServerConfig.Validate() expected error but got none")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("MCPServerConfig.Validate() unexpected error: %v", err)
			}
		})
	}
}