// Package config provides TLS configuration structures and validation
package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TLSConfig represents TLS configuration for MCP servers and clients
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

// CertificateType represents the type of certificate
type CertificateType int

const (
	ServerCert CertificateType = iota
	ClientCert
	CACert
)

// String returns the string representation of CertificateType
func (ct CertificateType) String() string {
	switch ct {
	case ServerCert:
		return "server"
	case ClientCert:
		return "client"
	case CACert:
		return "ca"
	default:
		return "unknown"
	}
}

// Certificate represents a TLS certificate with its metadata
type Certificate struct {
	Type         CertificateType `json:"type"`
	CommonName   string          `json:"common_name"`
	Organization string          `json:"organization"`
	Country      string          `json:"country"`
	Validity     time.Duration   `json:"validity"`
	KeySize      int             `json:"key_size"`
	SerialNumber int64           `json:"serial_number"`
}

// MCPServerConfig represents MCP server configuration with TLS support
type MCPServerConfig struct {
	Name       string    `json:"name"`
	HTTPPort   int       `json:"http_port"`
	TLSPort    int       `json:"tls_port"`
	TLSEnabled bool      `json:"tls_enabled"`
	TLSConfig  TLSConfig `json:"tls_config"`
}

// MCPClientConfig represents MCP client configuration with TLS support
type MCPClientConfig struct {
	ServerURL     string        `json:"server_url"`
	UseTLS        bool          `json:"use_tls"`
	TLSConfig     TLSConfig     `json:"tls_config"`
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
}

// NewTLSConfig creates a new TLS configuration with defaults
func NewTLSConfig(certDir string, demoMode bool) *TLSConfig {
	return &TLSConfig{
		CertDir:       certDir,
		ServerCert:    filepath.Join(certDir, "server.crt"),
		ServerKey:     filepath.Join(certDir, "server.key"),
		ClientCert:    filepath.Join(certDir, "client.crt"),
		ClientKey:     filepath.Join(certDir, "client.key"),
		CACert:        filepath.Join(certDir, "ca.crt"),
		DemoMode:      demoMode,
		MinTLSVersion: tls.VersionTLS12,
	}
}

// Validate validates the TLS configuration
func (c *TLSConfig) Validate() error {
	if c.CertDir == "" {
		return fmt.Errorf("certificate directory is required")
	}

	// Check if certificate directory exists
	if _, err := os.Stat(c.CertDir); os.IsNotExist(err) {
		return fmt.Errorf("certificate directory does not exist: %s", c.CertDir)
	}

	// In strict mode (non-demo), verify all certificate files exist
	if !c.DemoMode {
		certFiles := []string{c.ServerCert, c.ServerKey, c.ClientCert, c.ClientKey, c.CACert}
		for _, file := range certFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("certificate file does not exist: %s", file)
			}
		}
	}

	// Validate port range
	if c.Port != 0 && (c.Port < 1024 || c.Port > 65535) {
		return fmt.Errorf("port must be in range 1024-65535, got %d", c.Port)
	}

	// Validate minimum TLS version
	if c.MinTLSVersion < tls.VersionTLS12 {
		return fmt.Errorf("minimum TLS version must be >= TLS 1.2")
	}

	return nil
}

// NewCertificate creates a new certificate configuration
func NewCertificate(certType CertificateType, commonName string) *Certificate {
	return &Certificate{
		Type:         certType,
		CommonName:   commonName,
		Organization: "MCP Demo Organization",
		Country:      "US",
		Validity:     365 * 24 * time.Hour, // 1 year
		KeySize:      2048,
		SerialNumber: time.Now().Unix(),
	}
}

// Validate validates the certificate configuration
func (c *Certificate) Validate() error {
	if c.CommonName == "" {
		return fmt.Errorf("common name is required")
	}

	if c.KeySize < 2048 {
		return fmt.Errorf("key size must be >= 2048 bits, got %d", c.KeySize)
	}

	if c.Validity <= 0 || c.Validity > 10*365*24*time.Hour {
		return fmt.Errorf("validity period must be > 0 and <= 10 years")
	}

	return nil
}

// Validate validates the MCP server configuration
func (c *MCPServerConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("server name is required")
	}

	if c.HTTPPort == c.TLSPort && c.TLSEnabled {
		return fmt.Errorf("HTTP and TLS ports must be different")
	}

	if c.TLSEnabled {
		if err := c.TLSConfig.Validate(); err != nil {
			return fmt.Errorf("TLS configuration invalid: %w", err)
		}
	}

	return nil
}

// Validate validates the MCP client configuration
func (c *MCPClientConfig) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be > 0")
	}

	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts must be >= 0")
	}

	if c.UseTLS {
		if err := c.TLSConfig.Validate(); err != nil {
			return fmt.Errorf("TLS configuration invalid: %w", err)
		}
	}

	return nil
}