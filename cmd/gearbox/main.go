package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gearbox/cmd/gearbox/commands"
)

var version = "dev" // This will be set during build

func main() {
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
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewDoctorCmd())
	rootCmd.AddCommand(commands.NewStatusCmd())
	rootCmd.AddCommand(commands.NewGenerateCmd())

	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress non-error output")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}