package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/tls"
)

// TestCertificateGeneration tests the complete certificate generation workflow
func TestCertificateGeneration(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "cert_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create TLS configuration
	tlsConfig := config.NewTLSConfig(tempDir, true)

	// Create certificate manager
	certManager := tls.NewCertificateManager(tlsConfig)

	t.Run("generate_ca_certificate", func(t *testing.T) {
		// Test CA certificate generation
		err := certManager.GenerateCA()
		if err != nil {
			t.Fatalf("Failed to generate CA certificate: %v", err)
		}

		// Verify CA certificate file exists
		if _, err := os.Stat(tlsConfig.CACert); os.IsNotExist(err) {
			t.Errorf("CA certificate file was not created: %s", tlsConfig.CACert)
		}

		// Verify CA private key file exists
		caKeyPath := filepath.Join(tlsConfig.CertDir, "ca.key")
		if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
			t.Errorf("CA private key file was not created: %s", caKeyPath)
		}

		// Check file permissions on private key
		keyInfo, err := os.Stat(caKeyPath)
		if err != nil {
			t.Errorf("Failed to stat CA key file: %v", err)
		} else if keyInfo.Mode().Perm() != 0600 {
			t.Errorf("CA private key has incorrect permissions: got %o, want 0600", keyInfo.Mode().Perm())
		}
	})

	t.Run("generate_server_certificate", func(t *testing.T) {
		// First generate CA (prerequisite)
		err := certManager.GenerateCA()
		if err != nil {
			t.Fatalf("Failed to generate CA certificate: %v", err)
		}

		// Test server certificate generation
		err = certManager.GenerateServerCert("test-server")
		if err != nil {
			t.Fatalf("Failed to generate server certificate: %v", err)
		}

		// Verify server certificate file exists
		if _, err := os.Stat(tlsConfig.ServerCert); os.IsNotExist(err) {
			t.Errorf("Server certificate file was not created: %s", tlsConfig.ServerCert)
		}

		// Verify server private key file exists
		if _, err := os.Stat(tlsConfig.ServerKey); os.IsNotExist(err) {
			t.Errorf("Server private key file was not created: %s", tlsConfig.ServerKey)
		}
	})

	t.Run("generate_client_certificate", func(t *testing.T) {
		// First generate CA (prerequisite)
		err := certManager.GenerateCA()
		if err != nil {
			t.Fatalf("Failed to generate CA certificate: %v", err)
		}

		// Test client certificate generation
		err = certManager.GenerateClientCert("test-client")
		if err != nil {
			t.Fatalf("Failed to generate client certificate: %v", err)
		}

		// Verify client certificate file exists
		if _, err := os.Stat(tlsConfig.ClientCert); os.IsNotExist(err) {
			t.Errorf("Client certificate file was not created: %s", tlsConfig.ClientCert)
		}

		// Verify client private key file exists
		if _, err := os.Stat(tlsConfig.ClientKey); os.IsNotExist(err) {
			t.Errorf("Client private key file was not created: %s", tlsConfig.ClientKey)
		}
	})

	t.Run("generate_all_certificates", func(t *testing.T) {
		// Test generating all certificates in one call
		err := certManager.GenerateAllCerts()
		if err != nil {
			t.Fatalf("Failed to generate all certificates: %v", err)
		}

		// Verify all certificate files exist
		requiredFiles := []string{
			tlsConfig.CACert,
			filepath.Join(tlsConfig.CertDir, "ca.key"),
			tlsConfig.ServerCert,
			tlsConfig.ServerKey,
			tlsConfig.ClientCert,
			tlsConfig.ClientKey,
		}

		for _, file := range requiredFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Required certificate file missing: %s", file)
			}
		}
	})

	t.Run("validate_generated_certificates", func(t *testing.T) {
		// Generate certificates first
		err := certManager.GenerateAllCerts()
		if err != nil {
			t.Fatalf("Failed to generate certificates: %v", err)
		}

		// Validate CA certificate
		err = certManager.ValidateCertificate(tlsConfig.CACert)
		if err != nil {
			t.Errorf("CA certificate validation failed: %v", err)
		}

		// Validate server certificate
		err = certManager.ValidateCertificate(tlsConfig.ServerCert)
		if err != nil {
			t.Errorf("Server certificate validation failed: %v", err)
		}

		// Validate client certificate
		err = certManager.ValidateCertificate(tlsConfig.ClientCert)
		if err != nil {
			t.Errorf("Client certificate validation failed: %v", err)
		}
	})

	t.Run("certificate_information_retrieval", func(t *testing.T) {
		// Generate certificates first
		err := certManager.GenerateAllCerts()
		if err != nil {
			t.Fatalf("Failed to generate certificates: %v", err)
		}

		// Get CA certificate info
		caInfo, err := certManager.GetCertificateInfo(tlsConfig.CACert)
		if err != nil {
			t.Errorf("Failed to get CA certificate info: %v", err)
		} else {
			if !caInfo.IsCA {
				t.Errorf("CA certificate should be marked as CA")
			}
			if caInfo.Subject == "" {
				t.Errorf("CA certificate should have subject")
			}
		}

		// Get server certificate info
		serverInfo, err := certManager.GetCertificateInfo(tlsConfig.ServerCert)
		if err != nil {
			t.Errorf("Failed to get server certificate info: %v", err)
		} else {
			if serverInfo.IsCA {
				t.Errorf("Server certificate should not be marked as CA")
			}
			if time.Now().After(serverInfo.NotAfter) {
				t.Errorf("Server certificate should not be expired")
			}
		}

		// Get client certificate info
		clientInfo, err := certManager.GetCertificateInfo(tlsConfig.ClientCert)
		if err != nil {
			t.Errorf("Failed to get client certificate info: %v", err)
		} else {
			if clientInfo.IsCA {
				t.Errorf("Client certificate should not be marked as CA")
			}
			if time.Now().After(clientInfo.NotAfter) {
				t.Errorf("Client certificate should not be expired")
			}
		}
	})
}

// TestCertificateGenerationErrorCases tests error handling in certificate generation
func TestCertificateGenerationErrorCases(t *testing.T) {
	t.Run("invalid_certificate_directory", func(t *testing.T) {
		// Create TLS configuration with non-existent directory
		invalidConfig := config.NewTLSConfig("/nonexistent/path", false)
		certManager := tls.NewCertificateManager(invalidConfig)

		// This should fail because the directory doesn't exist
		err := certManager.GenerateCA()
		if err == nil {
			t.Error("Expected error when generating CA with invalid directory")
		}
	})

	t.Run("server_cert_without_ca", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "cert_error_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		tlsConfig := config.NewTLSConfig(tempDir, true)
		certManager := tls.NewCertificateManager(tlsConfig)

		// Try to generate server certificate without CA
		err = certManager.GenerateServerCert("test-server")
		if err == nil {
			t.Error("Expected error when generating server certificate without CA")
		}
	})

	t.Run("client_cert_without_ca", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "cert_error_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		tlsConfig := config.NewTLSConfig(tempDir, true)
		certManager := tls.NewCertificateManager(tlsConfig)

		// Try to generate client certificate without CA
		err = certManager.GenerateClientCert("test-client")
		if err == nil {
			t.Error("Expected error when generating client certificate without CA")
		}
	})
}

// TestCertificateValidationErrorCases tests certificate validation error handling
func TestCertificateValidationErrorCases(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cert_validation_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tlsConfig := config.NewTLSConfig(tempDir, true)
	certManager := tls.NewCertificateManager(tlsConfig)

	t.Run("validate_nonexistent_certificate", func(t *testing.T) {
		err := certManager.ValidateCertificate("/nonexistent/cert.pem")
		if err == nil {
			t.Error("Expected error when validating nonexistent certificate")
		}
	})

	t.Run("validate_invalid_certificate_format", func(t *testing.T) {
		// Create a file with invalid certificate content
		invalidCertPath := filepath.Join(tempDir, "invalid.crt")
		err := os.WriteFile(invalidCertPath, []byte("not a certificate"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid certificate file: %v", err)
		}

		err = certManager.ValidateCertificate(invalidCertPath)
		if err == nil {
			t.Error("Expected error when validating invalid certificate format")
		}
	})

	t.Run("get_info_nonexistent_certificate", func(t *testing.T) {
		_, err := certManager.GetCertificateInfo("/nonexistent/cert.pem")
		if err == nil {
			t.Error("Expected error when getting info for nonexistent certificate")
		}
	})
}