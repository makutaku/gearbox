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

// NewPlanCmd creates the plan command for removal analysis
func NewPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan uninstall [TOOLS...]",
		Short: "Show uninstall plan without executing",
		Long: `Analyze what would be removed for the specified tools without executing.

Shows dependency analysis, safety warnings, and detailed removal methods.
This is useful for understanding the impact of tool removal.`,
		Example: `  gearbox plan uninstall fd ripgrep         # Show removal plan
  gearbox plan uninstall fd --safety conservative  # Conservative analysis
  gearbox plan uninstall fd --safety aggressive    # Aggressive analysis`,
		RunE: runPlan,
	}

	// Safety options
	cmd.Flags().String("safety", "standard", "Safety level (conservative, standard, aggressive)")

	return cmd
}

func runPlan(cmd *cobra.Command, args []string) error {
	start := time.Now()
	log := logger.GetGlobalLogger().Operation("plan")
	
	if len(args) < 2 || args[0] != "uninstall" {
		return fmt.Errorf("plan command currently only supports 'uninstall'. Usage: gearbox plan uninstall [tools...]")
	}
	
	// Remove "uninstall" from args to get the tool list
	tools := args[1:]
	if len(tools) == 0 {
		return fmt.Errorf("no tools specified for analysis")
	}
	
	log.Infof("Planning uninstallation of %d tools", len(tools))
	
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return errors.NewSystemError("get executable path", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "bin", "orchestrator")

	// Check if the orchestrator is available
	if _, err := os.Stat(orchestratorPath); err != nil {
		return errors.NewDependencyError("check orchestrator", "orchestrator binary", err).
			WithMessage("Orchestrator not found. Please run 'make build' to compile all components.").
			WithContext("path", orchestratorPath)
	}

	result := runPlanWithOrchestrator(orchestratorPath, cmd, tools)
	log.Duration("plan", time.Since(start))
	return result
}

func runPlanWithOrchestrator(orchestratorPath string, cmd *cobra.Command, args []string) error {
	log := logger.GetGlobalLogger().Operation("orchestrator")
	
	log.Debug("Delegating to orchestrator")
	
	// Build the orchestrator command
	orchestratorCmd := exec.Command(orchestratorPath, "uninstall-plan")
	
	// Add tool arguments  
	orchestratorCmd.Args = append(orchestratorCmd.Args, args...)

	// Convert flags to orchestrator arguments
	if safety, _ := cmd.Flags().GetString("safety"); safety != "" {
		orchestratorCmd.Args = append(orchestratorCmd.Args, "--safety", safety)
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