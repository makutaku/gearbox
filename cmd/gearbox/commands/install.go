package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewInstallCmd creates the install command
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [TOOLS...]",
		Short: "Install development tools",
		Long: `Install one or more development tools with advanced orchestration.

If no tools are specified, you will be prompted to install all available tools.
The orchestrator provides parallel installation, dependency resolution, and 
comprehensive progress tracking.`,
		Example: `  gearbox install fd ripgrep fzf       # Install specific tools
  gearbox install --minimal fd         # Fast installation
  gearbox install --maximum ffmpeg     # Full-featured build
  gearbox install                      # Install all tools (with confirmation)`,
		RunE: runInstall,
	}

	// Build type flags
	cmd.Flags().Bool("minimal", false, "Fast builds with essential features")
	cmd.Flags().Bool("standard", false, "Balanced builds with reasonable features (default)")
	cmd.Flags().Bool("maximum", false, "Full-featured builds with all optimizations")

	// Installation options
	cmd.Flags().Bool("skip-common-deps", false, "Skip common dependency installation")
	cmd.Flags().Bool("run-tests", false, "Run test suites for validation")
	cmd.Flags().Bool("no-shell", false, "Skip shell integration setup (fzf, zoxide, etc.)")
	cmd.Flags().Bool("force", false, "Force reinstallation if already installed")

	// Performance options
	cmd.Flags().IntP("jobs", "j", 0, "Number of parallel jobs (0 = auto-detect)")
	cmd.Flags().Bool("no-cache", false, "Disable build cache")
	cmd.Flags().Bool("dry-run", false, "Show what would be installed without executing")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "bin", "orchestrator")

	// Check if the advanced orchestrator is available
	if _, err := os.Stat(orchestratorPath); err == nil {
		return runWithOrchestrator(orchestratorPath, cmd, args)
	}

	// Fallback to legacy shell-based installation
	return runWithLegacyScripts(repoDir, cmd, args)
}

func runWithOrchestrator(orchestratorPath string, cmd *cobra.Command, args []string) error {
	// Build the orchestrator command
	orchestratorCmd := exec.Command(orchestratorPath, "install")
	
	// Add tool arguments
	orchestratorCmd.Args = append(orchestratorCmd.Args, args...)

	// Convert flags to orchestrator arguments
	if minimal, _ := cmd.Flags().GetBool("minimal"); minimal {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--build-type", "minimal")
	}
	if maximum, _ := cmd.Flags().GetBool("maximum"); maximum {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--build-type", "maximum")
	}
	if skipDeps, _ := cmd.Flags().GetBool("skip-common-deps"); skipDeps {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--skip-common-deps")
	}
	if runTests, _ := cmd.Flags().GetBool("run-tests"); runTests {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--run-tests")
	}
	if noShell, _ := cmd.Flags().GetBool("no-shell"); noShell {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--no-shell")
	}
	if force, _ := cmd.Flags().GetBool("force"); force {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--force")
	}
	if jobs, _ := cmd.Flags().GetInt("jobs"); jobs > 0 {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--jobs", fmt.Sprintf("%d", jobs))
	}
	if noCache, _ := cmd.Flags().GetBool("no-cache"); noCache {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--no-cache")
	}
	if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--dry-run")
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

func runWithLegacyScripts(repoDir string, cmd *cobra.Command, args []string) error {
	fmt.Println("Note: Using legacy shell-based installation (orchestrator not available)")
	fmt.Println("For better performance and features, consider rebuilding the project.")
	fmt.Println()

	// If no tools specified, delegate to install-all-tools.sh
	if len(args) == 0 {
		allToolsScript := filepath.Join(repoDir, "scripts", "install-all-tools.sh")
		if _, err := os.Stat(allToolsScript); err != nil {
			return fmt.Errorf("install-all-tools.sh not found: %w", err)
		}

		legacyCmd := exec.Command(allToolsScript)
		
		// Convert flags to legacy script arguments
		var legacyArgs []string
		if minimal, _ := cmd.Flags().GetBool("minimal"); minimal {
			legacyArgs = append(legacyArgs, "--minimal")
		}
		if maximum, _ := cmd.Flags().GetBool("maximum"); maximum {
			legacyArgs = append(legacyArgs, "--maximum")
		}
		if skipDeps, _ := cmd.Flags().GetBool("skip-common-deps"); skipDeps {
			legacyArgs = append(legacyArgs, "--skip-common-deps")
		}
		if runTests, _ := cmd.Flags().GetBool("run-tests"); runTests {
			legacyArgs = append(legacyArgs, "--run-tests")
		}
		if noShell, _ := cmd.Flags().GetBool("no-shell"); noShell {
			legacyArgs = append(legacyArgs, "--no-shell")
		}

		legacyCmd.Args = append(legacyCmd.Args, legacyArgs...)
		legacyCmd.Stdout = os.Stdout
		legacyCmd.Stderr = os.Stderr
		legacyCmd.Stdin = os.Stdin
		
		return legacyCmd.Run()
	}

	// Install individual tools using their scripts
	for _, tool := range args {
		scriptPath := filepath.Join(repoDir, "scripts", fmt.Sprintf("install-%s.sh", tool))
		if _, err := os.Stat(scriptPath); err != nil {
			fmt.Printf("Warning: Script for tool '%s' not found: %s\n", tool, scriptPath)
			continue
		}

		fmt.Printf("Installing %s...\n", tool)
		
		toolCmd := exec.Command(scriptPath)
		
		// Convert flags to script arguments
		var scriptArgs []string
		if minimal, _ := cmd.Flags().GetBool("minimal"); minimal {
			scriptArgs = append(scriptArgs, "-m")
		}
		if maximum, _ := cmd.Flags().GetBool("maximum"); maximum {
			// Try different flag variations for maximum build
			scriptArgs = append(scriptArgs, "-r") // Most tools use -r for release/maximum
		}
		if skipDeps, _ := cmd.Flags().GetBool("skip-common-deps"); skipDeps {
			scriptArgs = append(scriptArgs, "--skip-deps")
		}
		if runTests, _ := cmd.Flags().GetBool("run-tests"); runTests {
			scriptArgs = append(scriptArgs, "--run-tests")
		}
		if force, _ := cmd.Flags().GetBool("force"); force {
			scriptArgs = append(scriptArgs, "--force")
		}

		toolCmd.Args = append(toolCmd.Args, scriptArgs...)
		toolCmd.Stdout = os.Stdout
		toolCmd.Stderr = os.Stderr
		
		if err := toolCmd.Run(); err != nil {
			fmt.Printf("Error installing %s: %v\n", tool, err)
			// Continue with other tools instead of failing completely
		}
	}

	return nil
}