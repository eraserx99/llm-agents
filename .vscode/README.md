# VSCode Configuration

This directory contains VSCode configuration files for debugging and running the LLM Multi-Agent System.

## Files

- **`launch.json`** - Debug configurations for all components
- **`tasks.json`** - Build and test tasks
- **`settings.json`** - Go-specific workspace settings

## Setting Up Your API Key

You have **3 options** for providing the `OPENROUTER_API_KEY`:

### Option 1: Using `.env` File (Recommended for Git)

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and set your API key:
   ```bash
   OPENROUTER_API_KEY=your-actual-api-key-here
   ```

3. The `.env` file is already gitignored, so your key won't be committed

4. VSCode will automatically load it via `envFile` in `launch.json`

### Option 2: Using VSCode Settings (Simplest)

1. Edit `.vscode/settings.json`

2. Replace `"your-api-key-here"` with your actual key:
   ```json
   "terminal.integrated.env.osx": {
       "OPENROUTER_API_KEY": "sk-or-v1-..."
   }
   ```

3. **Warning**: Be careful not to commit this file with your real API key

### Option 3: Using System Environment (Most Secure)

1. Add to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.):
   ```bash
   export OPENROUTER_API_KEY="your-api-key-here"
   ```

2. Restart VSCode or run:
   ```bash
   source ~/.zshrc  # or ~/.bashrc
   code .
   ```

3. VSCode will inherit the environment variable via `${env:OPENROUTER_API_KEY}`

## Debug Configurations

### Coordinator Agent

- **Launch Main (Coordinator Agent)** - Run the coordinator with a temperature query
- **Launch Main (Combined Query)** - Run with both weather and time query
- **Launch Main (Echo Query)** - Run with echo query

### Go MCP Servers

- **Launch Weather MCP Server** - HTTP mode
- **Launch Weather MCP Server (TLS)** - HTTPS mode with mTLS
- **Launch DateTime MCP Server** - HTTP mode
- **Launch DateTime MCP Server (TLS)** - HTTPS mode with mTLS
- **Launch Echo MCP Server** - HTTP mode
- **Launch Echo MCP Server (TLS)** - HTTPS mode with mTLS

### Java MCP Servers

- **Java Weather MCP Server** - HTTP mode
- **Java Weather MCP Server (TLS)** - HTTPS mode with mTLS
- **Java DateTime MCP Server** - HTTP mode
- **Java DateTime MCP Server (TLS)** - HTTPS mode with mTLS
- **Java Echo MCP Server** - HTTP mode
- **Java Echo MCP Server (TLS)** - HTTPS mode with mTLS

### Utilities

- **Launch Certificate Generator** - Generate TLS certificates

### Compound Configurations

- **All MCP Servers (HTTP)** - Launch all three Go servers in HTTP mode simultaneously
- **All MCP Servers (TLS)** - Launch all three Go servers in TLS mode simultaneously
- **All Java MCP Servers (HTTP)** - Launch all three Java servers in HTTP mode simultaneously
- **All Java MCP Servers (TLS)** - Launch all three Java servers in TLS mode simultaneously

## How to Debug

### Quick Start

1. **Set your API key** (see options above)

2. **Start all MCP servers**:
   - Open Run & Debug panel (`Cmd+Shift+D` on Mac, `Ctrl+Shift+D` on Windows/Linux)
   - Select **"All MCP Servers (HTTP)"** for Go servers, or **"All Java MCP Servers (HTTP)"** for Java servers
   - Press `F5` or click green play button
   - Wait for all servers to start

3. **Run the coordinator**:
   - Open Run & Debug panel again
   - Select "Launch Main (Coordinator Agent)"
   - Press `F5`
   - Set breakpoints as needed

### Debugging Java Servers

**Prerequisites**:
- Java 21+ installed
- Java servers built: `make build-java` (from project root)
- [Extension Pack for Java](https://marketplace.visualstudio.com/items?itemName=vscjava.vscode-java-pack) installed in VS Code

**To debug Java servers**:
1. Select any Java server configuration (e.g., "Java Weather MCP Server")
2. Press `F5` to launch with debugging
3. Set breakpoints in Java source files (`.java` files in `java-mcp-servers/src/`)
4. Java servers are fully protocol-compatible with Go servers

### Tips

- **Setting Breakpoints**: Click in the gutter (left of line numbers) to set breakpoints
- **Viewing Variables**: Use the Variables panel during debugging
- **Debug Console**: Use the Debug Console to evaluate expressions
- **Multiple Debug Sessions**: You can run multiple debug sessions simultaneously (servers + coordinator)

## Tasks

Press `Cmd+Shift+B` (Mac) or `Ctrl+Shift+B` (Windows/Linux) to run build tasks.

### Available Tasks

**Build Tasks:**
- `Build All` (default) - Builds all components via Makefile
- `Build Main (Coordinator)` - Build coordinator only
- `Build Weather MCP` - Build weather server only
- `Build DateTime MCP` - Build datetime server only
- `Build Echo MCP` - Build echo server only
- `Build Certificate Generator` - Build cert-gen only

**Test Tasks:**
- `Test All` (default test) - Run all tests
- `Test with Race Detector` - Run tests with race detection
- `Test Verbose` - Run tests with verbose output

**Quality Tasks:**
- `Format Code` - Run `go fmt`
- `Vet Code` - Run `go vet`
- `Lint Code` - Run `golangci-lint` (requires installation)
- `Quality Checks` - Run fmt, vet, and test together

**Other Tasks:**
- `Clean Build` - Remove build artifacts
- `Tidy Dependencies` - Run `go mod tidy`
- `Generate Certificates` - Create TLS certificates for mTLS
- `Start MCP Servers (HTTP)` - Start all servers in background (HTTP)
- `Start MCP Servers (TLS)` - Start all servers in background (TLS)
- `Stop MCP Servers` - Stop all running servers

## Environment Variables in launch.json

Each debug configuration can have custom environment variables. The coordinator configurations use:

```json
"env": {
    "OPENROUTER_API_KEY": "${env:OPENROUTER_API_KEY}"
},
"envFile": "${workspaceFolder}/.env"
```

This means:
1. First, it loads variables from `.env` file (if exists)
2. Then, it uses `${env:OPENROUTER_API_KEY}` which can come from:
   - The `.env` file
   - Your system environment
   - VSCode terminal environment (from `settings.json`)

## Troubleshooting

### "OPENROUTER_API_KEY not set" error

1. Verify your key is set using one of the 3 options above
2. If using `.env`, make sure it's in the project root (same level as `go.mod`)
3. If using system environment, restart VSCode after setting it
4. Check the Debug Console for the actual error message

### Servers not starting

1. Check if ports 8081-8083 are available: `lsof -i :8081`
2. Check server logs in the Debug Console
3. For TLS mode, ensure certificates exist: `ls -la certs/`

### Breakpoints not hitting

1. Make sure you're running in debug mode (not run mode)
2. Verify the code is compiled with debug symbols (default in VSCode)
3. Check that the source file path matches the running binary

### Need to change query/city/args?

Edit the `args` array in `launch.json` for any coordinator configuration:

```json
"args": [
    "-city", "Your City",
    "-query", "Your query here",
    "-verbose"
]
```

## Go Extension Settings

The `settings.json` file configures the Go extension with:

- **Auto-formatting** on save with `gofmt`
- **Auto-organize imports** on save
- **Linting** with `golangci-lint` on save (if installed)
- **Vetting** on save
- **Race detector** enabled for tests by default
- **Delve debugger** configured for optimal debugging

## Additional Resources

- [VSCode Go Extension Docs](https://github.com/golang/vscode-go/wiki)
- [Debugging in VSCode](https://code.visualstudio.com/docs/editor/debugging)
- [OpenRouter API](https://openrouter.ai)
