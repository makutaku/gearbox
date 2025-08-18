package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewDoctorCmd creates the doctor command
func NewDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks and diagnostics",
		Long: `Run comprehensive health checks to validate the system state,
tool installations, and configuration.

This command checks:
- System requirements and dependencies
- Tool installation status and integrity  
- Configuration validity
- Build environment setup
- Network connectivity for downloads`,
		RunE: runDoctor,
	}

	cmd.Flags().String("check", "", "Run specific check (system, tools, env, config)")
	cmd.Flags().Bool("fix", false, "Attempt to fix detected issues automatically")
	cmd.Flags().Bool("verbose", false, "Show detailed diagnostic output")

	return cmd
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	
	// Try to use advanced orchestrator health checks if available
	orchestratorPath := filepath.Join(repoDir, "bin", "orchestrator")
	if _, err := os.Stat(orchestratorPath); err == nil {
		return runWithOrchestratorDoctor(orchestratorPath, cmd, args)
	}

	// Fallback to shell-based doctor if available
	doctorScript := filepath.Join(repoDir, "lib", "doctor.sh")
	if _, err := os.Stat(doctorScript); err == nil {
		return runWithShellDoctor(repoDir, cmd, args)
	}

	// Basic built-in health checks
	return runBasicHealthChecks()
}

func runWithOrchestratorDoctor(orchestratorPath string, cmd *cobra.Command, args []string) error {
	doctorCmd := exec.Command(orchestratorPath, "doctor")
	
	// Pass through flags
	if check, _ := cmd.Flags().GetString("check"); check != "" {
		doctorCmd.Args = append(doctorCmd.Args, "--check", check)
	}
	if fix, _ := cmd.Flags().GetBool("fix"); fix {
		doctorCmd.Args = append(doctorCmd.Args, "--fix")
	}
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		doctorCmd.Args = append(doctorCmd.Args, "--verbose")
	}

	doctorCmd.Stdout = os.Stdout
	doctorCmd.Stderr = os.Stderr
	doctorCmd.Stdin = os.Stdin
	
	return doctorCmd.Run()
}

func runWithShellDoctor(repoDir string, cmd *cobra.Command, args []string) error {
	// Source the doctor library and run checks
	fmt.Println("Running shell-based health checks...")
	
	// This is a simplified version - the actual implementation would need
	// to properly source and execute the shell doctor functions
	doctorScript := fmt.Sprintf(`
		cd %s
		source lib/doctor.sh
		init_doctor
		run_all_checks
	`, repoDir)
	
	shellCmd := exec.Command("bash", "-c", doctorScript)
	shellCmd.Stdout = os.Stdout
	shellCmd.Stderr = os.Stderr
	
	return shellCmd.Run()
}

func runBasicHealthChecks() error {
	fmt.Println("🏥 Gearbox Health Check")
	fmt.Println("=====================")
	fmt.Println()

	// Basic system checks
	fmt.Println("✅ Go runtime available")
	
	// Check for common tools
	commonTools := []string{"git", "curl", "wget", "make", "gcc"}
	for _, tool := range commonTools {
		if _, err := exec.LookPath(tool); err == nil {
			fmt.Printf("✅ %s available\n", tool)
		} else {
			fmt.Printf("❌ %s not found\n", tool)
		}
	}

	fmt.Println()
	fmt.Println("Note: For comprehensive health checks, please rebuild the project")
	fmt.Println("to enable the advanced orchestrator and full diagnostic suite.")

	return nil
}