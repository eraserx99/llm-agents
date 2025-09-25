// Package main provides the main entry point for the LLM multi-agent system
package main

import (
	"fmt"
	"os"

	"github.com/steve/llm-agents/internal/cli"
	"github.com/steve/llm-agents/internal/utils"
)

// version is set via build-time ldflags
var version = "dev"

func main() {
	// Initialize logging
	utils.InitLogging()

	// Create and run the CLI application
	app := cli.NewApp()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
