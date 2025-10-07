package com.llmagents.mcp.common;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Server configuration loaded from environment variables and command-line flags.
 * Matches Go implementation's server configuration pattern.
 */
public record ServerConfig(
    int httpPort,
    int httpsPort,
    boolean tlsEnabled,
    boolean verbose,
    TLSConfig tlsConfig
) {
    private static final Logger logger = LoggerFactory.getLogger(ServerConfig.class);

    /**
     * Create server configuration from environment and CLI flags.
     *
     * @param defaultHttpPort default HTTP port
     * @param defaultHttpsPort default HTTPS port
     * @param httpPortEnvVar environment variable name for HTTP port
     * @param httpsPortEnvVar environment variable name for HTTPS port
     * @param tlsFlag whether --tls flag was passed
     * @param verboseFlag whether --verbose flag was passed
     * @return configured ServerConfig
     */
    public static ServerConfig create(
            int defaultHttpPort,
            int defaultHttpsPort,
            String httpPortEnvVar,
            String httpsPortEnvVar,
            boolean tlsFlag,
            boolean verboseFlag) {

        // Load TLS configuration
        TLSConfig tlsConfig = TLSConfig.fromEnvironment();

        // Determine if TLS is enabled (both environment and flag must be true)
        boolean tlsEnabled = tlsConfig.enabled() && tlsFlag;

        // Load port configuration from environment
        int httpPort = getEnvInt(httpPortEnvVar, defaultHttpPort);
        int httpsPort = getEnvInt(httpsPortEnvVar, defaultHttpsPort);

        ServerConfig config = new ServerConfig(
            httpPort,
            httpsPort,
            tlsEnabled,
            verboseFlag,
            tlsConfig
        );

        logger.info("Server configuration loaded: httpPort={}, httpsPort={}, tlsEnabled={}, verbose={}",
            httpPort, httpsPort, tlsEnabled, verboseFlag);

        return config;
    }

    /**
     * Get effective port based on TLS mode.
     *
     * @return active port number
     */
    public int effectivePort() {
        return tlsEnabled ? httpsPort : httpPort;
    }

    /**
     * Get effective protocol based on TLS mode.
     *
     * @return "https" or "http"
     */
    public String effectiveProtocol() {
        return tlsEnabled ? "https" : "http";
    }

    /**
     * Validate server configuration.
     * Checks that ports are valid and TLS certificates exist if enabled.
     *
     * @throws IllegalStateException if configuration is invalid
     */
    public void validate() {
        if (httpPort < 1 || httpPort > 65535) {
            throw new IllegalStateException("Invalid HTTP port: " + httpPort);
        }
        if (httpsPort < 1 || httpsPort > 65535) {
            throw new IllegalStateException("Invalid HTTPS port: " + httpsPort);
        }
        if (httpPort == httpsPort) {
            throw new IllegalStateException(
                "HTTP and HTTPS ports must be different: " + httpPort
            );
        }

        if (tlsEnabled) {
            tlsConfig.validate();
        }

        logger.info("Server configuration validated successfully");
    }

    /**
     * Enable verbose logging based on config.
     */
    public void configureLogging() {
        if (verbose) {
            System.setProperty("logback.verbose", "true");
            logger.info("Verbose logging enabled");
        }
    }

    private static int getEnvInt(String envVar, int defaultValue) {
        String value = System.getenv(envVar);
        if (value == null || value.isBlank()) {
            return defaultValue;
        }

        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException e) {
            logger.warn("Invalid integer value for {}: '{}', using default: {}",
                envVar, value, defaultValue);
            return defaultValue;
        }
    }
}
