// Package cli provides the command-line interface for the multi-agent system
package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/steve/llm-agents/internal/agents/coordinator"
	"github.com/steve/llm-agents/internal/config"
	"github.com/steve/llm-agents/internal/models"
	"github.com/steve/llm-agents/internal/utils"
)

// App represents the CLI application
type App struct {
	config      *config.Config
	coordinator *coordinator.Coordinator
}

// NewApp creates a new CLI application
func NewApp() *App {
	return &App{}
}

// Run runs the CLI application
func (a *App) Run(args []string) error {
	// Parse command-line flags
	fs := flag.NewFlagSet("llm-agents", flag.ExitOnError)

	var (
		city    = fs.String("city", "", "City name for weather/datetime queries (required)")
		query   = fs.String("query", "", "Query text (required)")
		verbose = fs.Bool("verbose", false, "Enable verbose output")
		version = fs.Bool("version", false, "Show version information")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", fs.Name())
		fmt.Fprintf(os.Stderr, "A multi-agent system for temperature and datetime queries using Claude 3.5 Sonnet.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -city \"New York\" -query \"What's the temperature?\"\n", fs.Name())
		fmt.Fprintf(os.Stderr, "  %s -city \"Los Angeles\" -query \"What time is it?\"\n", fs.Name())
		fmt.Fprintf(os.Stderr, "  %s -city \"Chicago\" -query \"What's the weather and time?\"\n", fs.Name())
		fmt.Fprintf(os.Stderr, "  %s -query \"echo hello world\"\n", fs.Name())
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  OPENROUTER_API_KEY    OpenRouter API key for Claude access (required)\n")
		fmt.Fprintf(os.Stderr, "  WEATHER_SERVER_URL    Weather MCP server URL (default: http://localhost:8081)\n")
		fmt.Fprintf(os.Stderr, "  DATETIME_SERVER_URL   DateTime MCP server URL (default: http://localhost:8082)\n")
		fmt.Fprintf(os.Stderr, "  ECHO_SERVER_URL       Echo MCP server URL (default: http://localhost:8083)\n")
		fmt.Fprintf(os.Stderr, "  LOG_LEVEL            Log level: debug, info, warn, error (default: info)\n")
	}

	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Show version if requested
	if *version {
		fmt.Printf("llm-agents version %s\n", getVersion())
		return nil
	}

	// Validate required parameters
	if *query == "" {
		fs.Usage()
		return fmt.Errorf("query is required")
	}

	// Check if query requires a city (not echo queries)
	if !isEchoQuery(*query) && *city == "" {
		fs.Usage()
		return fmt.Errorf("city is required for weather/datetime queries")
	}

	// Load configuration
	a.config = config.Load()

	// Set up logging
	if *verbose {
		utils.SetLogLevel("debug")
	} else {
		utils.SetLogLevel(a.config.LogLevel)
	}

	// Initialize coordinator
	if err := a.initializeCoordinator(); err != nil {
		return fmt.Errorf("failed to initialize coordinator: %w", err)
	}
	defer a.coordinator.Close()

	// Validate coordinator
	if err := a.coordinator.Validate(); err != nil {
		return fmt.Errorf("coordinator validation failed: %w", err)
	}

	// Process the query
	return a.processQuery(*query, *city, *verbose)
}

// initializeCoordinator initializes the coordinator with sub-agents
func (a *App) initializeCoordinator() error {
	// Validate OpenRouter API key
	if a.config.OpenRouterAPIKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY environment variable is required")
	}

	// Create coordinator
	a.coordinator = coordinator.NewCoordinator(
		a.config.OpenRouterAPIKey,
		a.config.WeatherMCPURL,
		a.config.DateTimeMCPURL,
		a.config.EchoMCPURL,
		a.config.MCPTimeout,
	)

	utils.Info("Coordinator initialized with servers:")
	utils.Info("  Weather: %s", a.config.WeatherMCPURL)
	utils.Info("  DateTime: %s", a.config.DateTimeMCPURL)
	utils.Info("  Echo: %s", a.config.EchoMCPURL)

	return nil
}

// processQuery processes the user query
func (a *App) processQuery(queryText, city string, verbose bool) error {
	// Create query
	query := models.Query{
		ID:        generateQueryID(),
		Text:      queryText,
		City:      city,
		Timestamp: time.Now(),
	}

	utils.Info("Processing query: %s", queryText)
	if city != "" {
		utils.Info("Target city: %s", city)
	}

	// Process query with coordinator
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := a.coordinator.ProcessQuery(ctx, query)
	if err != nil {
		return fmt.Errorf("query processing failed: %w", err)
	}

	// Display results
	a.displayResults(response, verbose)

	return nil
}

// displayResults displays the query results
func (a *App) displayResults(response *models.QueryResponse, verbose bool) {
	fmt.Printf("Query ID: %s\n", response.QueryID)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Duration: %s\n", response.Duration)
	fmt.Printf("Invoked agents: %s\n", formatAgentList(response.InvokedAgents))

	if len(response.Errors) > 0 {
		fmt.Printf("Errors: %v\n", response.Errors)
	}

	fmt.Println()

	// Display results by type
	if response.Temperature != nil {
		a.displayTemperatureData(response.Temperature)
	}

	if response.DateTime != nil {
		a.displayDateTimeData(response.DateTime)
	}

	if response.Echo != nil {
		a.displayEchoData(response.Echo)
	}

	// Display verbose information if requested
	if verbose {
		a.displayVerboseInfo(response)
	}
}

// displayTemperatureData displays temperature information
func (a *App) displayTemperatureData(data *models.TemperatureData) {
	fmt.Printf("🌡️  Temperature in %s:\n", data.City)
	fmt.Printf("   Temperature: %.1f°%s\n", data.Temperature, data.Unit)
	fmt.Printf("   Conditions: %s\n", data.Description)
	fmt.Printf("   Source: %s\n", data.Source)
	fmt.Printf("   Retrieved: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Println()
}

// displayDateTimeData displays datetime information
func (a *App) displayDateTimeData(data *models.DateTimeData) {
	fmt.Printf("🕐 Time in %s:\n", data.City)
	fmt.Printf("   Local time: %s\n", data.DateTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Timezone: %s\n", data.Timezone)
	fmt.Printf("   UTC offset: %s\n", data.UTCOffset)
	fmt.Printf("   Retrieved: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Println()
}

// displayEchoData displays echo information
func (a *App) displayEchoData(data *models.EchoData) {
	fmt.Printf("🔊 Echo result:\n")
	fmt.Printf("   Original: %s\n", data.OriginalText)
	fmt.Printf("   Echo: %s\n", data.EchoText)
	fmt.Printf("   Retrieved: %s\n", data.Timestamp.Format(time.RFC3339))
	fmt.Println()
}

// displayVerboseInfo displays verbose orchestration information
func (a *App) displayVerboseInfo(response *models.QueryResponse) {
	fmt.Println("📋 Orchestration Details:")

	fmt.Printf("   Execution log:\n")
	for i, entry := range response.OrchestrationLog {
		fmt.Printf("     %d. %s\n", i+1, entry)
	}
	fmt.Println()
}

// Helper functions

// isEchoQuery checks if the query is an echo request
func isEchoQuery(query string) bool {
	lowerQuery := strings.ToLower(query)
	return strings.Contains(lowerQuery, "echo") || strings.Contains(lowerQuery, "repeat")
}

// formatAgentList formats the list of invoked agents
func formatAgentList(agents []models.AgentType) string {
	if len(agents) == 0 {
		return "none"
	}

	agentNames := make([]string, len(agents))
	for i, agent := range agents {
		agentNames[i] = string(agent)
	}
	return strings.Join(agentNames, ", ")
}

// generateQueryID generates a unique query ID
func generateQueryID() string {
	return fmt.Sprintf("query-%d", time.Now().UnixNano())
}

// getVersion returns the application version
func getVersion() string {
	// This would typically be set via build-time ldflags
	return "1.0.0-dev"
}
