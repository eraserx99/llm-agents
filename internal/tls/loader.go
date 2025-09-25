// Package tls provides TLS certificate loading and configuration utilities
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/steve/llm-agents/internal/config"
)

// TLSLoader handles loading and configuring TLS certificates
type TLSLoader struct {
	config *config.TLSConfig
}

// NewTLSLoader creates a new TLS loader with the given configuration
func NewTLSLoader(cfg *config.TLSConfig) *TLSLoader {
	return &TLSLoader{
		config: cfg,
	}
}

// LoadServerTLSConfig loads and creates a TLS configuration for servers
func (loader *TLSLoader) LoadServerTLSConfig() (*tls.Config, error) {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(loader.config.ServerCert, loader.config.ServerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load CA certificate for client verification
	caCert, err := os.ReadFile(loader.config.CACert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   loader.config.MinTLSVersion,
		MaxVersion:   tls.VersionTLS13,
	}

	// Configure for demo mode if enabled
	if loader.config.DemoMode {
		// In demo mode, accept any client cert and do minimal custom verification
		tlsConfig.ClientAuth = tls.RequireAnyClientCert
		tlsConfig.InsecureSkipVerify = true // Skip built-in validation
		tlsConfig.VerifyConnection = func(cs tls.ConnectionState) error {
			// Custom demo-mode verification that's more permissive
			if len(cs.PeerCertificates) == 0 {
				return fmt.Errorf("no client certificate provided")
			}
			// In demo mode, just accept any certificate that's present and parseable
			return nil
		}
	}

	return tlsConfig, nil
}

// LoadClientTLSConfig loads and creates a TLS configuration for clients
func (loader *TLSLoader) LoadClientTLSConfig(serverName string) (*tls.Config, error) {
	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(loader.config.ClientCert, loader.config.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load CA certificate for server verification
	caCert, err := os.ReadFile(loader.config.CACert)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   serverName,
		MinVersion:   loader.config.MinTLSVersion,
		MaxVersion:   tls.VersionTLS13,
	}

	// Configure for demo mode if enabled
	if loader.config.DemoMode {
		tlsConfig.InsecureSkipVerify = true // Skip hostname verification in demo mode
		tlsConfig.VerifyPeerCertificate = loader.demoModeVerifyPeerCertificate
	}

	return tlsConfig, nil
}

// demoModeVerifyPeerCertificate provides custom certificate verification for demo mode
func (loader *TLSLoader) demoModeVerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	// In demo mode, we perform minimal verification:
	// 1. Certificate must be parseable
	// 2. Certificate must not be expired
	// 3. Certificate must be signed by our CA

	if len(rawCerts) == 0 {
		return fmt.Errorf("no certificates provided")
	}

	// Parse the peer certificate
	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("failed to parse peer certificate: %w", err)
	}

	// Load CA certificate
	caCertPEM, err := os.ReadFile(loader.config.CACert)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		return fmt.Errorf("failed to parse CA certificate")
	}

	// Verify certificate against CA
	opts := x509.VerifyOptions{
		Roots: caCertPool,
	}

	_, err = cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// ValidateCertificatePair validates that a certificate and private key pair match
func (loader *TLSLoader) ValidateCertificatePair(certPath, keyPath string) error {
	_, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("certificate and private key do not match: %w", err)
	}
	return nil
}

// GetTLSConnectionInfo extracts information from a TLS connection
func (loader *TLSLoader) GetTLSConnectionInfo(conn *tls.Conn) (*TLSConnectionInfo, error) {
	state := conn.ConnectionState()

	tlsVersion := "Unknown"
	switch state.Version {
	case tls.VersionTLS10:
		tlsVersion = "TLS 1.0"
	case tls.VersionTLS11:
		tlsVersion = "TLS 1.1"
	case tls.VersionTLS12:
		tlsVersion = "TLS 1.2"
	case tls.VersionTLS13:
		tlsVersion = "TLS 1.3"
	}

	cipherSuite := getCipherSuiteName(state.CipherSuite)

	var clientCertCN string
	if len(state.PeerCertificates) > 0 {
		clientCertCN = state.PeerCertificates[0].Subject.CommonName
	}

	return &TLSConnectionInfo{
		RemoteAddr:        conn.RemoteAddr().String(),
		TLSVersion:        tlsVersion,
		CipherSuite:       cipherSuite,
		ClientCertCN:      clientCertCN,
		HandshakeComplete: state.HandshakeComplete,
	}, nil
}

// TLSConnectionInfo holds information about a TLS connection
type TLSConnectionInfo struct {
	RemoteAddr        string `json:"remote_addr"`
	TLSVersion        string `json:"tls_version"`
	CipherSuite       string `json:"cipher_suite"`
	ClientCertCN      string `json:"client_cert_cn"`
	HandshakeComplete bool   `json:"handshake_complete"`
}

// getCipherSuiteName converts cipher suite ID to human readable name
func getCipherSuiteName(id uint16) string {
	suites := map[uint16]string{
		tls.TLS_RSA_WITH_RC4_128_SHA:                "TLS_RSA_WITH_RC4_128_SHA",
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:           "TLS_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA:            "TLS_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_RSA_WITH_AES_256_CBC_SHA:            "TLS_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256:         "TLS_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256:         "TLS_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384:         "TLS_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:        "TLS_ECDHE_ECDSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:    "TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:          "TLS_ECDHE_RSA_WITH_RC4_128_SHA",
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:     "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:      "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:   "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256: "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:   "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384: "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305:    "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305:  "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		tls.TLS_AES_128_GCM_SHA256:                  "TLS_AES_128_GCM_SHA256",
		tls.TLS_AES_256_GCM_SHA384:                  "TLS_AES_256_GCM_SHA384",
		tls.TLS_CHACHA20_POLY1305_SHA256:            "TLS_CHACHA20_POLY1305_SHA256",
	}

	if name, ok := suites[id]; ok {
		return name
	}
	return fmt.Sprintf("Unknown cipher suite (0x%04x)", id)
}