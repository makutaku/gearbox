package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewShowCmd creates the show command
func NewShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show bundle <name>",
		Short: "Show details about a bundle",
		Long: `Display detailed information about a specific bundle.

Shows:
- Bundle description and category
- Complete list of included tools
- Other bundles it includes
- Tags and metadata`,
		Args: cobra.ExactArgs(2),
		RunE: runShow,
	}

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	if args[0] != "bundle" {
		return fmt.Errorf("only 'show bundle' is supported at this time")
	}

	bundleName := args[1]

	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "orchestrator")

	// Check if the orchestrator is available
	if _, err := os.Stat(orchestratorPath); err != nil {
		return fmt.Errorf("orchestrator not found. Please run 'make build' to compile all components")
	}

	// Use the orchestrator to show bundle details
	orchestratorCmd := exec.Command(orchestratorPath, "show", "bundle", bundleName)
	orchestratorCmd.Stdout = os.Stdout
	orchestratorCmd.Stderr = os.Stderr
	
	return orchestratorCmd.Run()
}