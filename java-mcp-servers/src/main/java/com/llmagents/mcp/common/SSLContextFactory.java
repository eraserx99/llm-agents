package com.llmagents.mcp.common;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.net.ssl.KeyManagerFactory;
import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManagerFactory;
import javax.net.ssl.X509TrustManager;
import java.io.IOException;
import java.security.KeyManagementException;
import java.security.KeyStore;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.security.PrivateKey;
import java.security.UnrecoverableKeyException;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;

/**
 * Factory for creating SSLContext instances with mTLS support.
 * Matches Go implementation's TLS configuration.
 */
public final class SSLContextFactory {
    private static final Logger logger = LoggerFactory.getLogger(SSLContextFactory.class);

    private static final String KEY_STORE_PASSWORD = "changeit";  // Internal password

    private SSLContextFactory() {
        // Utility class
    }

    /**
     * Create SSLContext for server with mTLS support.
     *
     * @param tlsConfig TLS configuration with certificate paths
     * @return configured SSLContext for server
     * @throws IOException if certificates cannot be loaded
     * @throws CertificateException if certificates are invalid
     * @throws KeyStoreException if keystore operations fail
     * @throws NoSuchAlgorithmException if required algorithms are not available
     * @throws UnrecoverableKeyException if private key cannot be recovered
     * @throws KeyManagementException if SSLContext initialization fails
     */
    public static SSLContext createServerContext(TLSConfig tlsConfig)
            throws IOException, CertificateException, KeyStoreException,
                   NoSuchAlgorithmException, UnrecoverableKeyException, KeyManagementException {

        logger.info("Creating SSL context for server (demo mode: {})", tlsConfig.demoMode());

        // Load certificates and private key
        X509Certificate caCert = PEMCertificateLoader.loadCertificate(tlsConfig.caCertPath());
        X509Certificate serverCert = PEMCertificateLoader.loadCertificate(tlsConfig.serverCertPath());
        PrivateKey serverKey = PEMCertificateLoader.loadPrivateKey(tlsConfig.serverKeyPath());

        // Create key store with server certificate and private key
        KeyStore keyStore = KeyStore.getInstance(KeyStore.getDefaultType());
        keyStore.load(null, null);
        keyStore.setKeyEntry(
            "server",
            serverKey,
            KEY_STORE_PASSWORD.toCharArray(),
            new X509Certificate[]{serverCert, caCert}
        );

        // Create trust store with CA certificate
        KeyStore trustStore = KeyStore.getInstance(KeyStore.getDefaultType());
        trustStore.load(null, null);
        trustStore.setCertificateEntry("ca", caCert);

        // Initialize KeyManagerFactory with server key
        KeyManagerFactory kmf = KeyManagerFactory.getInstance(KeyManagerFactory.getDefaultAlgorithm());
        kmf.init(keyStore, KEY_STORE_PASSWORD.toCharArray());

        // Initialize TrustManagerFactory with CA certificate
        TrustManagerFactory tmf = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        tmf.init(trustStore);

        // Create and initialize SSLContext
        SSLContext sslContext = SSLContext.getInstance("TLS");
        sslContext.init(kmf.getKeyManagers(), tmf.getTrustManagers(), null);

        logger.info("SSL context created successfully for server");
        return sslContext;
    }

    /**
     * Create SSLContext for client with mTLS support.
     *
     * @param tlsConfig TLS configuration with certificate paths
     * @return configured SSLContext for client
     * @throws IOException if certificates cannot be loaded
     * @throws CertificateException if certificates are invalid
     * @throws KeyStoreException if keystore operations fail
     * @throws NoSuchAlgorithmException if required algorithms are not available
     * @throws UnrecoverableKeyException if private key cannot be recovered
     * @throws KeyManagementException if SSLContext initialization fails
     */
    public static SSLContext createClientContext(TLSConfig tlsConfig)
            throws IOException, CertificateException, KeyStoreException,
                   NoSuchAlgorithmException, UnrecoverableKeyException, KeyManagementException {

        logger.info("Creating SSL context for client (demo mode: {})", tlsConfig.demoMode());

        // Load certificates
        X509Certificate caCert = PEMCertificateLoader.loadCertificate(tlsConfig.caCertPath());
        X509Certificate clientCert = PEMCertificateLoader.loadCertificate(tlsConfig.clientCertPath());

        // Create trust store with CA certificate
        KeyStore trustStore = KeyStore.getInstance(KeyStore.getDefaultType());
        trustStore.load(null, null);
        trustStore.setCertificateEntry("ca", caCert);

        // Initialize TrustManagerFactory
        TrustManagerFactory tmf = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        tmf.init(trustStore);

        // Create and initialize SSLContext
        SSLContext sslContext = SSLContext.getInstance("TLS");
        sslContext.init(null, tmf.getTrustManagers(), null);

        logger.info("SSL context created successfully for client");
        return sslContext;
    }

    /**
     * Get X509TrustManager from TrustManagerFactory.
     * Useful for custom SSL configurations.
     *
     * @param trustStore trust store containing CA certificates
     * @return X509TrustManager
     * @throws NoSuchAlgorithmException if algorithm is not available
     * @throws KeyStoreException if keystore operation fails
     */
    public static X509TrustManager getTrustManager(KeyStore trustStore)
            throws NoSuchAlgorithmException, KeyStoreException {
        TrustManagerFactory tmf = TrustManagerFactory.getInstance(TrustManagerFactory.getDefaultAlgorithm());
        tmf.init(trustStore);

        for (var tm : tmf.getTrustManagers()) {
            if (tm instanceof X509TrustManager x509tm) {
                return x509tm;
            }
        }

        throw new IllegalStateException("No X509TrustManager found in TrustManagerFactory");
    }
}
