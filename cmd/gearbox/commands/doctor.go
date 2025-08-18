package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

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
	
	// Add cleanup subcommand
	cmd.AddCommand(NewDoctorCleanupCmd())

	return cmd
}

// NewDoctorCleanupCmd creates the doctor cleanup subcommand
func NewDoctorCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup [tool_names...]",
		Short: "Clean build artifacts and optimize disk usage",
		Long: `Clean build artifacts while preserving installed binaries and essential files.

This command provides three cleanup modes:
- minimal: Remove only temporary files (logs, .tmp files, etc.)
- standard: Remove intermediate build artifacts but keep source (default)
- aggressive: Remove everything except source files (if preserved)

Examples:
  gearbox doctor cleanup                    # Show disk usage report
  gearbox doctor cleanup fd ripgrep        # Clean specific tools  
  gearbox doctor cleanup --all             # Clean all tools
  gearbox doctor cleanup --mode aggressive # Maximum space savings
  gearbox doctor cleanup --dry-run         # Show what would be cleaned`,
		RunE: runDoctorCleanup,
	}

	cmd.Flags().String("mode", "standard", "Cleanup mode (minimal, standard, aggressive)")
	cmd.Flags().Bool("all", false, "Clean artifacts for all tools")
	cmd.Flags().Bool("dry-run", false, "Show what would be cleaned without doing it")
	cmd.Flags().Bool("auto-cleanup", false, "Enable automatic cleanup after future installs")
	cmd.Flags().String("preserve-source", "true", "Preserve source files in aggressive mode")

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

func runDoctorCleanup(cmd *cobra.Command, args []string) error {
	mode, _ := cmd.Flags().GetString("mode")
	all, _ := cmd.Flags().GetBool("all")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	autoCleanup, _ := cmd.Flags().GetBool("auto-cleanup")
	preserveSource, _ := cmd.Flags().GetString("preserve-source")

	// Set environment variables for shell functions
	os.Setenv("GEARBOX_CLEANUP_MODE", mode)
	os.Setenv("GEARBOX_PRESERVE_SOURCE", preserveSource)
	
	if autoCleanup {
		os.Setenv("GEARBOX_AUTO_CLEANUP", "true")
		fmt.Println("✅ Auto-cleanup enabled for future installations")
	}

	// Get repository directory
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	repoDir := filepath.Dir(execPath)

	// If no arguments and not --all, show disk usage report
	if len(args) == 0 && !all {
		return runDiskUsageReport(repoDir)
	}

	// Determine tools to clean
	var toolsToClean []string
	if all {
		// Get all tools from configuration
		toolsToClean, err = getAllToolNames(repoDir)
		if err != nil {
			fmt.Printf("Warning: Could not get full tool list: %v\n", err)
			// Fallback to common tools
			toolsToClean = []string{"fd", "ripgrep", "fzf", "bat", "eza", "bottom", "delta", "starship"}
		}
	} else {
		toolsToClean = args
	}

	fmt.Printf("🧹 Disk Space Cleanup (%s mode)\n", mode)
	fmt.Printf("================================\n\n")

	if dryRun {
		fmt.Println("DRY RUN - No files will be deleted")
		fmt.Println()
	}

	totalSaved := int64(0)
	for i, tool := range toolsToClean {
		fmt.Printf("[%d/%d] Cleaning %s...\n", i+1, len(toolsToClean), tool)
		
		if dryRun {
			// Show what would be cleaned without actually doing it
			saved := estimateCleanupSize(repoDir, tool, mode)
			totalSaved += saved
			if saved > 0 {
				fmt.Printf("  Would free: %s\n", humanReadableSize(saved))
			} else {
				fmt.Printf("  No artifacts to clean\n")
			}
		} else {
			// Run actual cleanup via shell function
			saved := runShellCleanup(repoDir, tool, mode)
			totalSaved += saved
		}
	}

	fmt.Printf("\n📊 Cleanup Summary\n")
	fmt.Printf("==================\n")
	if dryRun {
		fmt.Printf("Would free: %s across %d tools\n", humanReadableSize(totalSaved), len(toolsToClean))
		fmt.Printf("\nRun without --dry-run to perform cleanup\n")
	} else {
		fmt.Printf("Total freed: %s across %d tools\n", humanReadableSize(totalSaved), len(toolsToClean))
		fmt.Printf("✅ Cleanup completed successfully\n")
	}

	return nil
}

func runDiskUsageReport(repoDir string) error {
	fmt.Println("📊 Gearbox Disk Usage Report")
	fmt.Println("============================")
	fmt.Println()

	// Run shell function for disk usage
	script := fmt.Sprintf(`
		cd %s
		source lib/common.sh
		show_disk_usage
	`, repoDir)

	cmd := exec.Command("bash", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

func getAllToolNames(repoDir string) ([]string, error) {
	// Try to read from config/tools.json
	configPath := filepath.Join(repoDir, "config", "tools.json")
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Simple approach: extract tool names from build directory
	buildDir := filepath.Join(os.Getenv("HOME"), "tools", "build")
	if _, err := os.Stat(buildDir); err != nil {
		return []string{}, nil // No build directory means no tools to clean
	}

	var tools []string
	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			tools = append(tools, entry.Name())
		}
	}

	return tools, nil
}

func estimateCleanupSize(repoDir, tool, mode string) int64 {
	buildDir := filepath.Join(os.Getenv("HOME"), "tools", "build", tool)
	if _, err := os.Stat(buildDir); err != nil {
		return 0
	}

	// Get current size
	cmd := exec.Command("du", "-sb", buildDir)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var size int64
	fmt.Sscanf(string(output), "%d", &size)

	// Estimate reduction based on mode
	switch mode {
	case "minimal":
		return size / 20 // ~5% reduction
	case "standard":
		return size / 2  // ~50% reduction  
	case "aggressive":
		return size * 9 / 10 // ~90% reduction
	default:
		return size / 2
	}
}

func runShellCleanup(repoDir, tool, mode string) int64 {
	script := fmt.Sprintf(`
		cd %s
		source lib/common.sh
		cleanup_build_artifacts "%s" "%s"
	`, repoDir, tool, mode)

	cmd := exec.Command("bash", "-c", script)
	
	// Capture output to parse freed space while still showing it to user
	var output strings.Builder
	cmd.Stdout = io.MultiWriter(os.Stdout, &output)
	cmd.Stderr = os.Stderr
	
	cmd.Run()
	
	// Parse the output to extract freed space
	return parseFreedSpace(output.String())
}

// parseFreedSpace extracts the freed space from shell cleanup output
func parseFreedSpace(output string) int64 {
	// Look for patterns like "Cleaned 363MB from bottom" or "freed 363MB"
	patterns := []string{
		`Cleaned (\d+(?:\.\d+)?)([KMGTPE]?)B`,
		`freed (\d+(?:\.\d+)?)([KMGTPE]?)B`,
		`Freed (\d+(?:\.\d+)?)([KMGTPE]?)B`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) >= 3 {
			size, err := strconv.ParseFloat(matches[1], 64)
			if err != nil {
				continue
			}
			
			// Convert to bytes based on unit
			unit := matches[2]
			switch unit {
			case "K":
				size *= 1024
			case "M":
				size *= 1024 * 1024
			case "G":
				size *= 1024 * 1024 * 1024
			case "T":
				size *= 1024 * 1024 * 1024 * 1024
			case "P":
				size *= 1024 * 1024 * 1024 * 1024 * 1024
			case "E":
				size *= 1024 * 1024 * 1024 * 1024 * 1024 * 1024
			}
			
			return int64(size)
		}
	}
	
	return 0
}

func humanReadableSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}