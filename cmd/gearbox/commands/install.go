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

// NewInstallCmd creates the install command
func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [TOOLS...]",
		Short: "Install development tools",
		Long: `Install one or more development tools with advanced orchestration.

If no tools are specified, you will be prompted to install all available tools.
The orchestrator provides parallel installation, dependency resolution, and 
comprehensive progress tracking.`,
		Example: `  gearbox install fd ripgrep fzf             # Install specific tools
  gearbox install --bundle essential         # Install essential bundle
  gearbox install --bundle developer         # Install developer bundle
  gearbox install --minimal fd               # Fast installation
  gearbox install --maximum ffmpeg           # Full-featured build
  gearbox install nerd-fonts --fonts="FiraCode"    # Install specific font
  gearbox install nerd-fonts --interactive   # Interactive font selection
  gearbox install                            # Install all tools (with confirmation)`,
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

	// Nerd-fonts specific options
	cmd.Flags().String("fonts", "", "Install specific fonts (comma-separated, e.g. 'FiraCode,JetBrainsMono')")
	cmd.Flags().Bool("interactive", false, "Interactive font selection with previews")
	cmd.Flags().Bool("preview", false, "Show font previews before installation")
	cmd.Flags().Bool("configure-apps", false, "Automatically configure VS Code, terminals, etc.")
	
	// Bundle options
	cmd.Flags().String("bundle", "", "Install a predefined bundle (e.g. 'essential', 'developer', 'data-science')")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	start := time.Now()
	log := logger.GetGlobalLogger().Operation("install")
	
	// Handle bundle flag
	if bundleName, _ := cmd.Flags().GetString("bundle"); bundleName != "" {
		// Add bundle to args list
		args = append(args, bundleName)
	}
	
	// Debug: Show detailed argument parsing info
	fmt.Printf("🔍 CLI Debug - Raw args from Cobra: %v\n", args)
	fmt.Printf("🔍 CLI Debug - Command line: %v\n", os.Args)
	if fonts, _ := cmd.Flags().GetString("fonts"); fonts != "" {
		fmt.Printf("🔍 CLI Debug - fonts flag value: '%s'\n", fonts)
	}
	
	log.Infof("Starting installation of %d tools", len(args))
	
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

	result := runWithOrchestrator(orchestratorPath, cmd, args)
	log.Duration("install", time.Since(start))
	return result
}

func runWithOrchestrator(orchestratorPath string, cmd *cobra.Command, args []string) error {
	log := logger.GetGlobalLogger().Operation("orchestrator")
	
	log.Debug("Delegating to orchestrator")
	
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

	// Add nerd-fonts specific flags
	if fonts, _ := cmd.Flags().GetString("fonts"); fonts != "" {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--fonts", fonts)
	}
	if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--interactive")
	}
	if preview, _ := cmd.Flags().GetBool("preview"); preview {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--preview")
	}
	if configureApps, _ := cmd.Flags().GetBool("configure-apps"); configureApps {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--configure-apps")
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

