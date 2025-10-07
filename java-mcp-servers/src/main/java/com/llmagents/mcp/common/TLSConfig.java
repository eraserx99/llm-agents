package com.llmagents.mcp.common;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * TLS configuration loaded from environment variables.
 * Matches Go implementation's TLS settings.
 */
public record TLSConfig(
    boolean enabled,
    boolean demoMode,
    String certDir,
    Path caCertPath,
    Path serverCertPath,
    Path serverKeyPath,
    Path clientCertPath
) {
    private static final Logger logger = LoggerFactory.getLogger(TLSConfig.class);

    private static final String DEFAULT_CERT_DIR = "./certs";

    /**
     * Load TLS configuration from environment variables.
     * Environment variables:
     * - TLS_ENABLED: Enable TLS mode (default: false)
     * - TLS_DEMO_MODE: Use relaxed validation for demo (default: true)
     * - TLS_CERT_DIR: Certificate directory path (default: ./certs)
     */
    public static TLSConfig fromEnvironment() {
        boolean enabled = Boolean.parseBoolean(
            System.getenv().getOrDefault("TLS_ENABLED", "false")
        );
        boolean demoMode = Boolean.parseBoolean(
            System.getenv().getOrDefault("TLS_DEMO_MODE", "true")
        );
        String certDir = System.getenv().getOrDefault("TLS_CERT_DIR", DEFAULT_CERT_DIR);

        Path basePath = Paths.get(certDir);
        Path caCert = basePath.resolve("ca.crt");
        Path serverCert = basePath.resolve("server.crt");
        Path serverKey = basePath.resolve("server.key");
        Path clientCert = basePath.resolve("client.crt");

        TLSConfig config = new TLSConfig(
            enabled,
            demoMode,
            certDir,
            caCert,
            serverCert,
            serverKey,
            clientCert
        );

        logger.debug("TLS Configuration loaded: enabled={}, demoMode={}, certDir={}",
            enabled, demoMode, certDir);

        return config;
    }

    /**
     * Validate that all required certificate files exist.
     *
     * @throws IllegalStateException if any required certificate file is missing
     */
    public void validate() {
        if (!enabled) {
            return;
        }

        StringBuilder missing = new StringBuilder();

        if (!Files.exists(caCertPath)) {
            missing.append("  - CA certificate: ").append(caCertPath).append("\n");
        }
        if (!Files.exists(serverCertPath)) {
            missing.append("  - Server certificate: ").append(serverCertPath).append("\n");
        }
        if (!Files.exists(serverKeyPath)) {
            missing.append("  - Server private key: ").append(serverKeyPath).append("\n");
        }

        if (!missing.isEmpty()) {
            throw new IllegalStateException(
                "TLS enabled but required certificate files are missing:\n" +
                missing +
                "Please generate certificates using: make generate-certs\n" +
                "Or set TLS_CERT_DIR environment variable to the correct path."
            );
        }

        logger.info("TLS certificates validated successfully");
    }

    /**
     * Check if certificate files are readable.
     *
     * @return true if all certificate files exist and are readable
     */
    public boolean areCertificatesReadable() {
        if (!enabled) {
            return true;
        }

        return Files.isReadable(caCertPath) &&
               Files.isReadable(serverCertPath) &&
               Files.isReadable(serverKeyPath);
    }
}
