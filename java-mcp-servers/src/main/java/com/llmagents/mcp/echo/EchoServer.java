package com.llmagents.mcp.echo;

import com.llmagents.mcp.common.SSLContextFactory;
import com.llmagents.mcp.common.ServerConfig;
import com.llmagents.mcp.transport.MCPServlet;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.ServerConnector;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHolder;
import org.eclipse.jetty.util.ssl.SslContextFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import picocli.CommandLine;
import picocli.CommandLine.Command;
import picocli.CommandLine.Option;

import javax.net.ssl.SSLContext;
import java.util.List;
import java.util.concurrent.Callable;

/**
 * Echo MCP Server main class.
 * Provides echo tool via MCP protocol.
 */
@Command(name = "echo-mcp-server", mixinStandardHelpOptions = true,
         version = "1.0.0", description = "Echo MCP Server")
public class EchoServer implements Callable<Integer> {
    private static final Logger logger = LoggerFactory.getLogger(EchoServer.class);

    @Option(names = {"--tls"}, description = "Enable TLS mode")
    private boolean tlsFlag;

    @Option(names = {"--verbose"}, description = "Enable verbose logging")
    private boolean verbose;

    public static void main(String[] args) {
        int exitCode = new CommandLine(new EchoServer()).execute(args);
        System.exit(exitCode);
    }

    @Override
    public Integer call() throws Exception {
        // Load configuration
        ServerConfig config = ServerConfig.create(
            8083,  // default HTTP port
            8445,  // default HTTPS port
            "ECHO_MCP_PORT",
            "ECHO_MCP_TLS_PORT",
            tlsFlag,
            verbose
        );

        config.configureLogging();
        config.validate();

        logger.info("Starting Echo MCP Server");
        logger.info("HTTP port: {}, HTTPS port: {}, TLS enabled: {}",
            config.httpPort(), config.httpsPort(), config.tlsEnabled());

        // Create tool handler
        EchoTool echoTool = new EchoTool();

        // Create MCP servlet
        MCPServlet mcpServlet = new MCPServlet(
            "echo-mcp",
            echoTool::handleToolCall,
            List.of(EchoTool.getToolDefinition())
        );

        // Create Jetty server
        Server server = new Server();

        // Configure connector (HTTP or HTTPS)
        ServerConnector connector;
        if (config.tlsEnabled()) {
            // HTTPS with mTLS
            SSLContext sslContext = SSLContextFactory.createServerContext(config.tlsConfig());
            SslContextFactory.Server sslContextFactory = new SslContextFactory.Server();
            sslContextFactory.setSslContext(sslContext);
            sslContextFactory.setNeedClientAuth(true);  // Require client certificate (mTLS)

            connector = new ServerConnector(server, sslContextFactory);
            connector.setPort(config.httpsPort());
            logger.info("Configured HTTPS connector on port {} with mTLS", config.httpsPort());
        } else {
            // HTTP
            connector = new ServerConnector(server);
            connector.setPort(config.httpPort());
            logger.info("Configured HTTP connector on port {}", config.httpPort());
        }

        server.addConnector(connector);

        // Configure servlet
        ServletContextHandler context = new ServletContextHandler(ServletContextHandler.SESSIONS);
        context.setContextPath("/");
        server.setHandler(context);

        context.addServlet(new ServletHolder(mcpServlet), "/mcp");

        // Start server
        try {
            server.start();
            logger.info("Echo MCP Server started successfully on {}://localhost:{}",
                config.effectiveProtocol(), config.effectivePort());
            server.join();
            return 0;
        } catch (Exception e) {
            logger.error("Failed to start Echo MCP Server", e);
            return 1;
        }
    }
}
