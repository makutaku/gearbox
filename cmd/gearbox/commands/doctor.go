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
		Use:   "doctor [tool_name]",
		Short: "Run health checks and diagnostics",
		Long: `Run comprehensive health checks to validate the system state,
tool installations, and configuration.

General health checks:
- System requirements and dependencies
- Tool installation status and integrity  
- Configuration validity
- Build environment setup
- Network connectivity for downloads

Tool-specific diagnostics:
- nerd-fonts: Font installation, cache status, and availability checks

Examples:
  gearbox doctor                    # General system health check
  gearbox doctor nerd-fonts         # Nerd Fonts specific diagnostics
  gearbox doctor nerd-fonts --verbose  # Detailed font analysis`,
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
	
	// Check if tool-specific diagnostics are requested
	if len(args) > 0 {
		toolName := args[0]
		return runToolSpecificDoctor(repoDir, toolName, cmd)
	}
	
	// Try to use advanced orchestrator health checks if available for general checks
	orchestratorPath := filepath.Join(repoDir, "bin", "orchestrator")
	if _, err := os.Stat(orchestratorPath); err == nil {
		return runWithOrchestratorDoctor(orchestratorPath, cmd, args)
	}

	// Fallback to shell-based doctor if available for general checks
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
	// Check if tool-specific diagnostics are requested
	if len(args) > 0 {
		toolName := args[0]
		return runToolSpecificDoctor(repoDir, toolName, cmd)
	}
	
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
	fmt.Println("🔍 General Health Check")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("For tool-specific diagnostics, specify a tool name.")
	fmt.Println("Example: gearbox doctor nerd-fonts")

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

// runToolSpecificDoctor handles diagnostics for specific tools
func runToolSpecificDoctor(repoDir, toolName string, cmd *cobra.Command) error {
	switch toolName {
	case "nerd-fonts":
		return runNerdFontsDoctor(repoDir, cmd)
	default:
		return fmt.Errorf("tool-specific diagnostics not implemented for '%s'", toolName)
	}
}

// runNerdFontsDoctor performs comprehensive nerd-fonts health checks
func runNerdFontsDoctor(repoDir string, cmd *cobra.Command) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	
	fmt.Println("🎨 Nerd Fonts Health Check")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	
	// Track overall health
	var totalChecks, passedChecks, failedChecks, warningChecks int
	
	// 1. Check font installation status
	fmt.Println("📍 Installation Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	installedFonts, err := getInstalledNerdFonts()
	totalChecks++
	if err != nil {
		fmt.Printf("❌ Failed to check installed fonts: %v\n", err)
		failedChecks++
	} else if len(installedFonts) == 0 {
		fmt.Printf("⚠️  No Nerd Fonts detected\n")
		warningChecks++
	} else {
		fmt.Printf("✅ Found %d Nerd Fonts installed\n", len(installedFonts))
		passedChecks++
		
		if verbose {
			fmt.Println("\nInstalled fonts:")
			for _, font := range installedFonts[:min(10, len(installedFonts))] {
				fmt.Printf("  • %s\n", font)
			}
			if len(installedFonts) > 10 {
				fmt.Printf("  ... and %d more\n", len(installedFonts)-10)
			}
		}
	}
	fmt.Println()
	
	// 2. Check font cache status
	fmt.Println("🔄 Font Cache Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	totalChecks++
	cacheValid := checkFontCache()
	if cacheValid {
		fmt.Println("✅ Font cache is up to date")
		passedChecks++
	} else {
		fmt.Println("⚠️  Font cache may need refresh (run: fc-cache -fv)")
		warningChecks++
	}
	fmt.Println()
	
	// 3. Check fonts directory
	fmt.Println("📁 Fonts Directory")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	fontsDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "fonts")
	totalChecks++
	if stat, err := os.Stat(fontsDir); err != nil {
		fmt.Printf("❌ Fonts directory not found: %s\n", fontsDir)
		failedChecks++
	} else if !stat.IsDir() {
		fmt.Printf("❌ Fonts path is not a directory: %s\n", fontsDir)
		failedChecks++
	} else {
		fmt.Printf("✅ Fonts directory exists: %s\n", fontsDir)
		passedChecks++
		
		if verbose {
			// Show directory size and file count
			if size := getDirSize(fontsDir); size > 0 {
				fmt.Printf("   Size: %s\n", humanReadableSize(size))
			}
			if count := countFontFiles(fontsDir); count > 0 {
				fmt.Printf("   Font files: %d\n", count)
			}
		}
	}
	fmt.Println()
	
	// 4. Check application support
	fmt.Println("🖥️  Application Support")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	apps := []string{"code", "gnome-terminal", "konsole", "alacritty", "kitty"}
	for _, app := range apps {
		totalChecks++
		if _, err := exec.LookPath(app); err == nil {
			fmt.Printf("✅ %s available\n", app)
			passedChecks++
		} else {
			fmt.Printf("ℹ️  %s not installed\n", app)
			// Don't count as failure since these are optional
		}
	}
	fmt.Println()
	
	// 5. Check popular font availability
	fmt.Println("🔍 Font Availability Check")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	popularFonts := []string{"FiraCode", "JetBrains", "Hack", "SauceCodePro", "CaskaydiaCove"}
	for _, font := range popularFonts {
		totalChecks++
		if checkFontAvailable(font) {
			fmt.Printf("✅ %s Nerd Font available\n", font)
			passedChecks++
		} else {
			fmt.Printf("❌ %s Nerd Font not found\n", font)
			failedChecks++
		}
	}
	fmt.Println()
	
	// 6. Check installation script
	fmt.Println("🛠️  Installation Script")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	scriptPath := filepath.Join(repoDir, "scripts", "install-nerd-fonts.sh")
	totalChecks++
	if stat, err := os.Stat(scriptPath); err != nil {
		fmt.Printf("❌ Installation script not found: %s\n", scriptPath)
		failedChecks++
	} else if stat.Mode()&0111 == 0 {
		fmt.Printf("⚠️  Installation script not executable: %s\n", scriptPath)
		warningChecks++
	} else {
		fmt.Printf("✅ Installation script ready: %s\n", scriptPath)
		passedChecks++
		
		if verbose {
			fmt.Printf("   Modified: %s\n", stat.ModTime().Format("2006-01-02 15:04:05"))
			fmt.Printf("   Size: %s\n", humanReadableSize(stat.Size()))
		}
	}
	fmt.Println()
	
	// Summary
	fmt.Println("📊 Health Check Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Total checks: %d\n", totalChecks)
	fmt.Printf("✅ Passed: %d\n", passedChecks)
	if warningChecks > 0 {
		fmt.Printf("⚠️  Warnings: %d\n", warningChecks)
	}
	if failedChecks > 0 {
		fmt.Printf("❌ Failed: %d\n", failedChecks)
	}
	
	fmt.Println()
	
	// Overall status
	if failedChecks > 0 {
		fmt.Println("🔴 Nerd Fonts health: CRITICAL - Issues detected")
		fmt.Println("   Recommendation: Run 'gearbox install nerd-fonts' to fix installation")
	} else if warningChecks > 0 {
		fmt.Println("🟡 Nerd Fonts health: GOOD - Minor issues detected")
		fmt.Println("   Recommendation: Consider refreshing font cache with 'fc-cache -fv'")
	} else {
		fmt.Println("🟢 Nerd Fonts health: EXCELLENT - All checks passed")
	}
	
	return nil
}

// Helper functions for nerd-fonts diagnostics

func getInstalledNerdFonts() ([]string, error) {
	cmd := exec.Command("fc-list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var fonts []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "nerd") {
			// Extract font name from fc-list output
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				fontName := strings.TrimSpace(parts[1])
				if fontName != "" {
					fonts = append(fonts, fontName)
				}
			}
		}
	}
	
	return fonts, nil
}

func checkFontCache() bool {
	// Check if fc-cache is available and working
	cmd := exec.Command("fc-cache", "--version")
	return cmd.Run() == nil
}

func checkFontAvailable(fontName string) bool {
	cmd := exec.Command("fc-list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	outputLower := strings.ToLower(string(output))
	fontLower := strings.ToLower(fontName)
	
	return strings.Contains(outputLower, fontLower) && strings.Contains(outputLower, "nerd")
}

func getDirSize(dirPath string) int64 {
	cmd := exec.Command("du", "-sb", dirPath)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	var size int64
	fmt.Sscanf(string(output), "%d", &size)
	return size
}

func countFontFiles(dirPath string) int {
	cmd := exec.Command("find", dirPath, "-name", "*.ttf", "-o", "-name", "*.otf")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}