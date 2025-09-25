// Package tls provides certificate management and TLS configuration utilities
package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/steve/llm-agents/internal/config"
)

// CertificateManager handles certificate generation, loading, and validation
type CertificateManager struct {
	config *config.TLSConfig
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(cfg *config.TLSConfig) *CertificateManager {
	return &CertificateManager{
		config: cfg,
	}
}

// GenerateCA generates a Certificate Authority certificate and private key
func (cm *CertificateManager) GenerateCA() error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"MCP Demo CA"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// Write certificate to file
	certOut, err := os.Create(cm.config.CACert)
	if err != nil {
		return fmt.Errorf("failed to create CA certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	// Write private key to file
	keyOut, err := os.Create(filepath.Join(cm.config.CertDir, "ca.key"))
	if err != nil {
		return fmt.Errorf("failed to create CA key file: %w", err)
	}
	defer keyOut.Close()

	// Set restrictive permissions for private key
	if err := keyOut.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set CA key permissions: %w", err)
	}

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal CA private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return fmt.Errorf("failed to write CA private key: %w", err)
	}

	return nil
}

// GenerateServerCert generates a server certificate signed by the CA
func (cm *CertificateManager) GenerateServerCert(commonName string) error {
	return cm.generateCert(commonName, config.ServerCert, cm.config.ServerCert, cm.config.ServerKey, true)
}

// GenerateClientCert generates a client certificate signed by the CA
func (cm *CertificateManager) GenerateClientCert(commonName string) error {
	return cm.generateCert(commonName, config.ClientCert, cm.config.ClientCert, cm.config.ClientKey, false)
}

// generateCert is a helper function to generate certificates
func (cm *CertificateManager) generateCert(commonName string, certType config.CertificateType, certPath, keyPath string, isServer bool) error {
	// Load CA certificate and key
	caCert, caKey, err := cm.loadCA()
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization:  []string{"MCP Demo"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    commonName,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{},
	}

	if isServer {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		// Add localhost and 127.0.0.1 for demo purposes
		template.DNSNames = []string{"localhost", commonName}
		template.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	} else {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key to file
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	// Set restrictive permissions for private key
	if err := keyOut.Chmod(0600); err != nil {
		return fmt.Errorf("failed to set key permissions: %w", err)
	}

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// loadCA loads the CA certificate and private key
func (cm *CertificateManager) loadCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	// Load CA certificate
	caCertPEM, err := os.ReadFile(cm.config.CACert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Load CA private key
	caKeyPath := filepath.Join(cm.config.CertDir, "ca.key")
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CA private key: %w", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return nil, nil, fmt.Errorf("failed to parse CA private key PEM")
	}

	caKey, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	rsaKey, ok := caKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("CA private key is not RSA")
	}

	return caCert, rsaKey, nil
}

// GenerateAllCerts generates all required certificates (CA, server, client)
func (cm *CertificateManager) GenerateAllCerts() error {
	// Generate CA first
	if err := cm.GenerateCA(); err != nil {
		return fmt.Errorf("failed to generate CA: %w", err)
	}

	// Generate server certificate
	if err := cm.GenerateServerCert("mcp-server"); err != nil {
		return fmt.Errorf("failed to generate server certificate: %w", err)
	}

	// Generate client certificate
	if err := cm.GenerateClientCert("mcp-client"); err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	return nil
}

// ValidateCertificate validates a certificate file
func (cm *CertificateManager) ValidateCertificate(certPath string) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if certificate is expired
	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	if time.Now().Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}

	return nil
}

// CertificateInfo returns information about a certificate
type CertificateInfo struct {
	Subject      string    `json:"subject"`
	Issuer       string    `json:"issuer"`
	SerialNumber string    `json:"serial_number"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	IsCA         bool      `json:"is_ca"`
	KeyUsage     []string  `json:"key_usage"`
}

// GetCertificateInfo returns information about a certificate
func (cm *CertificateManager) GetCertificateInfo(certPath string) (*CertificateInfo, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	keyUsage := []string{}
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		keyUsage = append(keyUsage, "Digital Signature")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		keyUsage = append(keyUsage, "Key Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		keyUsage = append(keyUsage, "Certificate Sign")
	}

	return &CertificateInfo{
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		IsCA:         cert.IsCA,
		KeyUsage:     keyUsage,
	}, nil
}