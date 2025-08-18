package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [TOOLS...]",
		Short: "Show installation status of tools",
		Long: `Display the installation status of development tools.

Shows which tools are installed, their versions, and installation paths.
If specific tools are provided, only those tools are checked.`,
		RunE: runStatus,
	}

	cmd.Flags().BoolP("all", "a", false, "Show status for all tools (default)")
	cmd.Flags().BoolP("installed", "i", false, "Show only installed tools")
	cmd.Flags().BoolP("missing", "m", false, "Show only missing tools")
	cmd.Flags().BoolP("detailed", "d", false, "Show detailed information")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "bin", "orchestrator")

	// This command requires the orchestrator for proper status tracking
	if _, err := os.Stat(orchestratorPath); err == nil {
		return runStatusWithOrchestrator(orchestratorPath, cmd, args)
	}

	// Fallback to basic status check
	return runBasicStatus(args)
}

func runStatusWithOrchestrator(orchestratorPath string, cmd *cobra.Command, args []string) error {
	statusCmd := exec.Command(orchestratorPath, "status")
	
	// Add tool arguments
	statusCmd.Args = append(statusCmd.Args, args...)

	// Pass through flags
	if all, _ := cmd.Flags().GetBool("all"); all {
		statusCmd.Args = append(statusCmd.Args, "--all")
	}
	if installed, _ := cmd.Flags().GetBool("installed"); installed {
		statusCmd.Args = append(statusCmd.Args, "--installed")
	}
	if missing, _ := cmd.Flags().GetBool("missing"); missing {
		statusCmd.Args = append(statusCmd.Args, "--missing")
	}
	if detailed, _ := cmd.Flags().GetBool("detailed"); detailed {
		statusCmd.Args = append(statusCmd.Args, "--detailed")
	}

	statusCmd.Stdout = os.Stdout
	statusCmd.Stderr = os.Stderr
	
	return statusCmd.Run()
}

func runBasicStatus(tools []string) error {
	fmt.Println("📊 Tool Installation Status")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Note: This is a basic status check. For comprehensive status")
	fmt.Println("information, please rebuild the project to enable the orchestrator.")
	fmt.Println()

	// If specific tools provided, check only those
	if len(tools) > 0 {
		for _, tool := range tools {
			checkToolStatus(tool)
		}
		return nil
	}

	// Check common tools
	commonTools := []string{
		"fd", "ripgrep", "fzf", "jq", "zoxide", "yazi", "bat", "eza",
		"starship", "delta", "lazygit", "bottom", "procs", "tokei",
		"hyperfine", "gh", "dust", "sd", "tealdeer", "choose",
	}

	for _, tool := range commonTools {
		checkToolStatus(tool)
	}

	return nil
}

func checkToolStatus(tool string) {
	// Simple check using 'which' or 'command -v'
	if _, err := exec.LookPath(tool); err == nil {
		fmt.Printf("✅ %-12s installed\n", tool)
	} else {
		fmt.Printf("❌ %-12s not found\n", tool)
	}
}