#!/bin/bash
# Test script to capture Go MCP server responses

export OPENROUTER_API_KEY="dummy-key-for-testing"
export LOG_LEVEL="debug"
export TLS_ENABLED=true
export TLS_DEMO_MODE=true
export TLS_CERT_DIR=./certs
export MCP_WEATHER_URL=https://localhost:8443/mcp
export MCP_DATETIME_URL=https://localhost:8444/mcp
export MCP_ECHO_URL=https://localhost:8445/mcp

# Run with verbose logging
./bin/llm-agents -city "Chicago" -query "What is the temperature?" -verbose 2>&1 | tee test-go-mcp-output.log
