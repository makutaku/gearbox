package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gearbox/pkg/errors"
	"gearbox/pkg/logger"
)

// NewUninstallCmd creates the uninstall command
func NewUninstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [TOOLS...]",
		Short: "Uninstall development tools safely",
		Long: `Uninstall one or more development tools with dependency analysis and safe removal.

The uninstall command analyzes dependencies and provides a removal plan before execution.
It ensures that removing tools won't break other installed tools unless forced.`,
		Example: `  gearbox uninstall fd ripgrep              # Uninstall specific tools
  gearbox uninstall fd --force              # Force removal despite dependencies
  gearbox uninstall fd --cascade            # Remove unused dependencies
  gearbox uninstall fd --dry-run            # Show what would be removed
  gearbox uninstall fd --remove-config      # Remove configuration files too
  gearbox uninstall fd --no-backup          # Skip backup creation`,
		RunE: runUninstall,
	}

	// Removal options
	cmd.Flags().Bool("force", false, "Force removal even if there are dependents")
	cmd.Flags().Bool("cascade", false, "Remove unused dependencies")
	cmd.Flags().Bool("remove-config", false, "Remove configuration files")
	cmd.Flags().Bool("dry-run", false, "Show what would be removed without executing")
	cmd.Flags().Bool("backup", true, "Create backup before removal")
	cmd.Flags().Bool("no-backup", false, "Skip backup creation")
	cmd.Flags().String("backup-suffix", "", "Suffix for backup files")

	// Bundle options
	cmd.Flags().Bool("bundle-contents", false, "Remove all tools in bundle, not just bundle tracking")

	// Safety options
	cmd.Flags().String("safety", "standard", "Safety level (conservative, standard, aggressive)")

	return cmd
}

func runUninstall(cmd *cobra.Command, args []string) error {
	start := time.Now()
	log := logger.GetGlobalLogger().Operation("uninstall")
	
	if len(args) == 0 {
		return fmt.Errorf("no tools specified for removal")
	}
	
	log.Infof("Starting uninstallation of %d tools", len(args))
	
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return errors.NewSystemError("get executable path", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "orchestrator")

	// Check if the orchestrator is available
	if _, err := os.Stat(orchestratorPath); err != nil {
		return errors.NewDependencyError("check orchestrator", "orchestrator binary", err).
			WithMessage("Orchestrator not found. Please run 'make build' to compile all components.").
			WithContext("path", orchestratorPath)
	}

	result := runUninstallWithOrchestrator(orchestratorPath, cmd, args)
	log.Duration("uninstall", time.Since(start))
	return result
}

func runUninstallWithOrchestrator(orchestratorPath string, cmd *cobra.Command, args []string) error {
	log := logger.GetGlobalLogger().Operation("orchestrator")
	
	log.Debug("Delegating to orchestrator")
	
	// Build the orchestrator command
	orchestratorCmd := exec.Command(orchestratorPath, "uninstall")
	
	// Add tool arguments  
	orchestratorCmd.Args = append(orchestratorCmd.Args, args...)

	// Convert flags to orchestrator arguments
	if force, _ := cmd.Flags().GetBool("force"); force {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--force")
	}
	if cascade, _ := cmd.Flags().GetBool("cascade"); cascade {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--cascade")
	}
	if removeConfig, _ := cmd.Flags().GetBool("remove-config"); removeConfig {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--remove-config")
	}
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--dry-run")
	}
	
	// Handle backup flags
	if noBackup, _ := cmd.Flags().GetBool("no-backup"); noBackup {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--backup=false")
	} else if backup, _ := cmd.Flags().GetBool("backup"); backup {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--backup=true")
	}
	
	if backupSuffix, _ := cmd.Flags().GetString("backup-suffix"); backupSuffix != "" {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--backup-suffix", backupSuffix)
	}
	
	if bundleContents, _ := cmd.Flags().GetBool("bundle-contents"); bundleContents {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--bundle-contents")
	}

	// Add global flags
	if verbose, _ := cmd.Parent().PersistentFlags().GetBool("verbose"); verbose {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--verbose")
	}

	// Connect stdio
	orchestratorCmd.Stdout = os.Stdout
	orchestratorCmd.Stderr = os.Stderr
	orchestratorCmd.Stdin = os.Stdin

	return orchestratorCmd.Run()
}