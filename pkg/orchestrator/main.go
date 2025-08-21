package orchestrator

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Global variables
var (
	configPath string
	repoDir    string
)

// Main is the entry point for the orchestrator command-line tool.
// It sets up the CLI commands and executes the orchestrator functionality.
func Main() {
	var rootCmd = &cobra.Command{
		Use:   "orchestrator",
		Short: "Gearbox installation orchestrator",
		Long:  "Advanced orchestration engine for managing gearbox tool installations with dependency resolution and parallel execution",
	}

	// Add commands
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(showCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(verifyCmd())
	rootCmd.AddCommand(doctorCmd())
	
	// Add tracking commands
	rootCmd.AddCommand(trackInstallationCmd())
	rootCmd.AddCommand(trackBundleCmd())
	rootCmd.AddCommand(isTrackedCmd())
	rootCmd.AddCommand(trackPreexistingCmd())
	rootCmd.AddCommand(initManifestCmd())
	rootCmd.AddCommand(manifestStatusCmd())
	rootCmd.AddCommand(listDependentsCmd())
	rootCmd.AddCommand(canRemoveCmd())
	
	// Add uninstall commands
	rootCmd.AddCommand(uninstallCmd())
	rootCmd.AddCommand(uninstallPlanCmd())
	rootCmd.AddCommand(uninstallExecuteCmd())

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to tools.json configuration file")
	rootCmd.PersistentFlags().StringVar(&repoDir, "repo-dir", "", "Repository directory (default: auto-detect)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}