package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gearbox/cmd/gearbox/commands"
	"gearbox/pkg/errors"
	"gearbox/pkg/logger"
)

var version = "dev" // This will be set during build

func main() {
	// Initialize logging first
	initLogging()
	
	rootCmd := &cobra.Command{
		Use:   "gearbox",
		Short: "Essential Tools Installer",
		Long: `Gearbox - Essential Tools Installer

A powerful command-line tool for installing and managing essential development tools.
Supports parallel installation, dependency resolution, and comprehensive health checks.`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(commands.NewInstallCmd())
	rootCmd.AddCommand(commands.NewListCmd())
	rootCmd.AddCommand(commands.NewShowCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewDoctorCmd())
	rootCmd.AddCommand(commands.NewStatusCmd())
	rootCmd.AddCommand(commands.NewGenerateCmd())

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-error output")

	// Set up pre-execution hook for logging configuration
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureLogging(cmd)
	}
	
	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		handleError(err)
		os.Exit(1)
	}
}

// initLogging initializes the global logger with default settings.
func initLogging() {
	logger.SetGlobalLogger(logger.NewDefault())
}

// configureLogging configures logging based on command flags.
func configureLogging(cmd *cobra.Command) {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")
	
	var log *logger.Logger
	switch {
	case verbose:
		log = logger.NewVerbose()
	case quiet:
		log = logger.NewQuiet()
	default:
		log = logger.NewDefault()
	}
	
	logger.SetGlobalLogger(log)
}

// handleError provides enhanced error handling with suggestions.
func handleError(err error) {
	if gearboxErr, ok := err.(*errors.GearboxError); ok {
		// Handle structured gearbox errors
		fmt.Fprintf(os.Stderr, "❌ %s\n", gearboxErr.Error())
		
		if suggestion := gearboxErr.GetSuggestion(); suggestion != "" {
			fmt.Fprintf(os.Stderr, "💡 %s\n", suggestion)
		}
		
		// Show debug information if verbose
		if verbose, _ := os.LookupEnv("GEARBOX_DEBUG"); verbose == "true" {
			fmt.Fprintf(os.Stderr, "\n🐛 Debug Information:\n%s\n", gearboxErr.String())
		}
	} else {
		// Handle regular errors
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "💡 Run 'gearbox doctor' for system diagnostics\n")
	}
}