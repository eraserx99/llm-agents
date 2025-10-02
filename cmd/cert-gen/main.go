// Certificate Generation Utility for MCP Servers
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/tls"
	"github.com/steve/llm-agents/internal/utils"
)

var (
	certDir    = flag.String("cert-dir", "./certs", "Directory to store certificates")
	demoMode   = flag.Bool("demo-mode", false, "Enable demo mode (relaxed validation)")
	serverName = flag.String("server-name", "mcp-server", "Common name for server certificate")
	clientName = flag.String("client-name", "mcp-client", "Common name for client certificate")
	force      = flag.Bool("force", false, "Overwrite existing certificates")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()

	// Initialize logging
	logLevel := "INFO"
	if *verbose {
		logLevel = "DEBUG"
	}
	utils.InitLogger(logLevel, true)

	utils.Info("Starting certificate generation...")
	utils.Info("Certificate directory: %s", *certDir)
	utils.Info("Demo mode: %v", *demoMode)

	// Create certificate directory if it doesn't exist
	if err := os.MkdirAll(*certDir, 0755); err != nil {
		log.Fatalf("Failed to create certificate directory: %v", err)
	}

	// Create TLS configuration
	tlsConfig := config.NewTLSConfig(*certDir, *demoMode)

	// Check if certificates already exist
	if !*force && certificatesExist(tlsConfig) {
		fmt.Println("Certificates already exist. Use --force to overwrite.")
		listExistingCertificates(tlsConfig)
		return
	}

	// Create certificate manager
	certManager := tls.NewCertificateManager(tlsConfig)

	// Generate CA certificate
	utils.Info("Generating Certificate Authority...")
	if err := certManager.GenerateCA(); err != nil {
		log.Fatalf("Failed to generate CA certificate: %v", err)
	}
	utils.Info("✓ CA certificate generated: %s", tlsConfig.CACert)

	// Generate server certificate
	utils.Info("Generating server certificate for: %s", *serverName)
	if err := certManager.GenerateServerCert(*serverName); err != nil {
		log.Fatalf("Failed to generate server certificate: %v", err)
	}
	utils.Info("✓ Server certificate generated: %s", tlsConfig.ServerCert)

	// Generate client certificate
	utils.Info("Generating client certificate for: %s", *clientName)
	if err := certManager.GenerateClientCert(*clientName); err != nil {
		log.Fatalf("Failed to generate client certificate: %v", err)
	}
	utils.Info("✓ Client certificate generated: %s", tlsConfig.ClientCert)

	// Set proper permissions
	if err := setPermissions(tlsConfig); err != nil {
		log.Fatalf("Failed to set certificate permissions: %v", err)
	}

	// Display certificate information
	fmt.Println("\n🎉 Certificate generation completed successfully!")
	fmt.Println("\nGenerated certificates:")
	displayCertificateInfo(certManager, tlsConfig)

	// Display next steps
	fmt.Println("\n📝 Next steps:")
	fmt.Println("1. Set environment variables:")
	fmt.Printf("   export TLS_ENABLED=true\n")
	fmt.Printf("   export TLS_DEMO_MODE=%v\n", *demoMode)
	fmt.Printf("   export TLS_CERT_DIR=%s\n", *certDir)
	fmt.Println("2. Start MCP servers with --tls flag")
	fmt.Println("3. Test connections with cmd/test-tls utility")
}

// certificatesExist checks if certificates already exist
func certificatesExist(cfg *config.TLSConfig) bool {
	files := []string{cfg.CACert, cfg.ServerCert, cfg.ClientCert}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// listExistingCertificates lists existing certificates
func listExistingCertificates(cfg *config.TLSConfig) {
	fmt.Println("\nExisting certificates:")
	files := map[string]string{
		"CA Certificate":     cfg.CACert,
		"Server Certificate": cfg.ServerCert,
		"Client Certificate": cfg.ClientCert,
	}

	for name, path := range files {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("  ✓ %s: %s\n", name, path)
		} else {
			fmt.Printf("  ✗ %s: %s (missing)\n", name, path)
		}
	}
}

// setPermissions sets appropriate permissions for certificate files
func setPermissions(cfg *config.TLSConfig) error {
	// Private keys should be readable only by owner
	keyFiles := []string{
		filepath.Join(cfg.CertDir, "ca.key"),
		cfg.ServerKey,
		cfg.ClientKey,
	}

	for _, keyFile := range keyFiles {
		if err := os.Chmod(keyFile, 0600); err != nil {
			return fmt.Errorf("failed to set permissions for %s: %w", keyFile, err)
		}
		utils.Debug("Set permissions 600 for %s", keyFile)
	}

	// Certificates can be readable by group
	certFiles := []string{
		cfg.CACert,
		cfg.ServerCert,
		cfg.ClientCert,
	}

	for _, certFile := range certFiles {
		if err := os.Chmod(certFile, 0644); err != nil {
			return fmt.Errorf("failed to set permissions for %s: %w", certFile, err)
		}
		utils.Debug("Set permissions 644 for %s", certFile)
	}

	return nil
}

// displayCertificateInfo displays information about generated certificates
func displayCertificateInfo(certManager *tls.CertificateManager, cfg *config.TLSConfig) {
	certificates := map[string]string{
		"CA Certificate":     cfg.CACert,
		"Server Certificate": cfg.ServerCert,
		"Client Certificate": cfg.ClientCert,
	}

	for name, path := range certificates {
		fmt.Printf("\n%s (%s):\n", name, filepath.Base(path))

		info, err := certManager.GetCertificateInfo(path)
		if err != nil {
			fmt.Printf("  Error reading certificate: %v\n", err)
			continue
		}

		fmt.Printf("  Subject: %s\n", info.Subject)
		fmt.Printf("  Valid from: %s\n", info.NotBefore.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Valid until: %s\n", info.NotAfter.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Serial: %s\n", info.SerialNumber)
		if info.IsCA {
			fmt.Printf("  Type: Certificate Authority\n")
		}
		if len(info.KeyUsage) > 0 {
			fmt.Printf("  Key Usage: %v\n", info.KeyUsage)
		}
	}
}