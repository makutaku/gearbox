package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
		Long: `Manage gearbox configuration settings including build types, 
cache settings, and installation preferences.

Configuration is stored in ~/.gearboxrc and can be managed through
this command or edited directly.`,
		RunE: runConfig,
	}

	// Subcommands will be added here
	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE:  runConfigShow,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE:  runConfigSet,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "wizard",
		Short: "Interactive configuration setup",
		RunE:  runConfigWizard,
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE:  runConfigReset,
	})

	return cmd
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Default to show if no subcommand specified
	return runConfigShow(cmd, args)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Try to use the config-manager if available
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	configManagerPath := filepath.Join(repoDir, "bin", "config-manager")

	if _, err := os.Stat(configManagerPath); err == nil {
		configCmd := exec.Command(configManagerPath, "show")
		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr
		return configCmd.Run()
	}

	// Fallback to shell-based config
	return runLegacyConfigShow()
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]
	
	// Try to use the config-manager if available
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	configManagerPath := filepath.Join(repoDir, "bin", "config-manager")

	if _, err := os.Stat(configManagerPath); err == nil {
		configCmd := exec.Command(configManagerPath, "set", key, value)
		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr
		return configCmd.Run()
	}

	return fmt.Errorf("config management requires the config-manager tool")
}

func runConfigWizard(cmd *cobra.Command, args []string) error {
	// Try to use the config-manager if available
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	configManagerPath := filepath.Join(repoDir, "bin", "config-manager")

	if _, err := os.Stat(configManagerPath); err == nil {
		configCmd := exec.Command(configManagerPath, "wizard")
		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr
		configCmd.Stdin = os.Stdin
		return configCmd.Run()
	}

	return fmt.Errorf("config wizard requires the config-manager tool")
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	// Try to use the config-manager if available
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	configManagerPath := filepath.Join(repoDir, "bin", "config-manager")

	if _, err := os.Stat(configManagerPath); err == nil {
		configCmd := exec.Command(configManagerPath, "reset")
		configCmd.Stdout = os.Stdout
		configCmd.Stderr = os.Stderr
		configCmd.Stdin = os.Stdin
		return configCmd.Run()
	}

	return fmt.Errorf("config reset requires the config-manager tool")
}

func runLegacyConfigShow() error {
	// Simple fallback config display
	fmt.Println("Configuration (legacy mode):")
	fmt.Println("For full configuration management, please rebuild the project.")
	
	return nil
}