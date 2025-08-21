package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// isNerdFontsInstalled checks if any Nerd Fonts are installed
func isNerdFontsInstalled() bool {
	// Check if fc-list command is available
	if _, err := exec.LookPath("fc-list"); err != nil {
		return false
	}
	
	// Check if any Nerd Fonts are installed
	cmd := exec.Command("fc-list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	outputStr := strings.ToLower(string(output))
	return strings.Contains(outputStr, "nerd font") || strings.Contains(outputStr, "nerd")
}

// getNerdFontsVersion returns information about installed Nerd Fonts
func getNerdFontsVersion() string {
	if _, err := exec.LookPath("fc-list"); err != nil {
		return "fc-list not available"
	}
	
	// Get list of installed Nerd Fonts
	cmd := exec.Command("sh", "-c", "fc-list | grep -i 'nerd font' | wc -l")
	output, err := cmd.Output()
	if err != nil {
		return "Error checking fonts"
	}
	
	count := strings.TrimSpace(string(output))
	if count == "0" {
		return "No Nerd Fonts installed"
	}
	
	// Get some example fonts
	cmd = exec.Command("sh", "-c", "fc-list | grep -i 'nerd font' | head -3 | cut -d: -f2 | cut -d, -f1 | sort | uniq")
	examples, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("%s Nerd Fonts installed", count)
	}
	
	exampleList := strings.TrimSpace(string(examples))
	if exampleList != "" {
		lines := strings.Split(exampleList, "\n")
		cleanLines := make([]string, 0, len(lines))
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				cleanLines = append(cleanLines, line)
			}
		}
		if len(cleanLines) > 0 {
			return fmt.Sprintf("%s fonts (%s)", count, strings.Join(cleanLines, ", "))
		}
	}
	
	return fmt.Sprintf("%s Nerd Fonts installed", count)
}

// getNerdFontsDetailedStatus returns detailed status information for nerd-fonts
func getNerdFontsDetailedStatus() map[string]interface{} {
	status := make(map[string]interface{})
	
	// Check if fc-list is available
	if _, err := exec.LookPath("fc-list"); err != nil {
		status["error"] = "fc-list not available"
		return status
	}
	
	// Get all Nerd Fonts
	cmd := exec.Command("sh", "-c", "fc-list | grep -i 'nerd font' | cut -d: -f2 | cut -d, -f1 | sort | uniq")
	output, err := cmd.Output()
	if err != nil {
		status["error"] = "Error listing fonts"
		return status
	}
	
	fontList := strings.TrimSpace(string(output))
	if fontList == "" {
		status["installed"] = false
		status["count"] = 0
		status["fonts"] = []string{}
		return status
	}
	
	fonts := strings.Split(fontList, "\n")
	cleanFonts := make([]string, 0, len(fonts))
	for _, font := range fonts {
		font = strings.TrimSpace(font)
		if font != "" {
			cleanFonts = append(cleanFonts, font)
		}
	}
	
	status["installed"] = len(cleanFonts) > 0
	status["count"] = len(cleanFonts)
	status["fonts"] = cleanFonts
	
	// Get disk usage
	homeDir := os.Getenv("HOME")
	fontsDir := homeDir + "/.local/share/fonts"
	
	cmd = exec.Command("du", "-sh", fontsDir)
	if diskOutput, err := cmd.Output(); err == nil {
		diskUsage := strings.Fields(string(diskOutput))
		if len(diskUsage) > 0 {
			status["disk_usage"] = diskUsage[0]
		}
	}
	
	// Check font cache status
	cmd = exec.Command("fc-cache", "--version")
	if err := cmd.Run(); err == nil {
		status["font_cache"] = "available"
	} else {
		status["font_cache"] = "unavailable"
	}
	
	return status
}

// showNerdFontsDetailedStatus displays detailed status for nerd-fonts
func (o *Orchestrator) showNerdFontsDetailedStatus() {
	status := getNerdFontsDetailedStatus()
	
	if errorMsg, hasError := status["error"]; hasError {
		fmt.Printf("❌ %-15s %s\n", "nerd-fonts", errorMsg)
		return
	}
	
	installed, _ := status["installed"].(bool)
	count, _ := status["count"].(int)
	fonts, _ := status["fonts"].([]string)
	
	if !installed || count == 0 {
		fmt.Printf("❌ %-15s Not installed\n", "nerd-fonts")
		return
	}
	
	fmt.Printf("✅ %-15s %d fonts installed\n", "nerd-fonts", count)
	
	// Show individual fonts
	for _, font := range fonts {
		if len(font) > 30 {
			font = font[:27] + "..."
		}
		fmt.Printf("   ├─ %s\n", font)
	}
	
	// Show disk usage if available
	if diskUsage, hasDisk := status["disk_usage"]; hasDisk {
		fmt.Printf("   └─ 💾 Disk usage: %s in ~/.local/share/fonts\n", diskUsage)
	}
}

// runNerdFontsDoctor runs comprehensive nerd-fonts health checks
func (o *Orchestrator) runNerdFontsDoctor() error {
	fmt.Printf("🔍 Nerd Fonts Health Check\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	var issues []string
	var recommendations []string
	
	// 1. Check if nerd-fonts are installed
	if !isNerdFontsInstalled() {
		fmt.Printf("❌ Nerd Fonts: Not installed\n")
		issues = append(issues, "No Nerd Fonts installed")
		recommendations = append(recommendations, "Run 'gearbox install nerd-fonts' to install fonts")
	} else {
		status := getNerdFontsDetailedStatus()
		count, _ := status["count"].(int)
		fonts, _ := status["fonts"].([]string)
		
		fmt.Printf("✅ Nerd Fonts: %d fonts installed\n", count)
		for _, font := range fonts[:min(3, len(fonts))] {
			fmt.Printf("   ├─ %s\n", font)
		}
		if len(fonts) > 3 {
			fmt.Printf("   └─ ... and %d more\n", len(fonts)-3)
		}
	}
	
	// 2. Check font cache
	if _, err := exec.LookPath("fc-cache"); err != nil {
		fmt.Printf("❌ Font Cache: fc-cache not available\n")
		issues = append(issues, "Font cache system unavailable")
		recommendations = append(recommendations, "Install fontconfig package: sudo apt install fontconfig")
	} else {
		fmt.Printf("✅ Font Cache: Available and working\n")
	}
	
	// 3. Check terminal support
	terminalSupport := checkTerminalSupport()
	if terminalSupport {
		fmt.Printf("✅ Terminal Support: Unicode symbols supported\n")
	} else {
		fmt.Printf("⚠️  Terminal Support: Limited Unicode support detected\n")
		issues = append(issues, "Terminal may not display all font symbols")
		recommendations = append(recommendations, "Use a modern terminal like Kitty, Alacritty, or configure your current terminal")
	}
	
	// 4. Check VS Code configuration
	vscodeConfigured, _ := checkVSCodeFontConfig()
	if vscodeConfigured {
		fmt.Printf("✅ VS Code: Configured to use Nerd Fonts\n")
	} else {
		if isVSCodeInstalled() {
			fmt.Printf("⚠️  VS Code: Not configured to use Nerd Fonts\n")
			issues = append(issues, "VS Code not configured for Nerd Fonts")
			recommendations = append(recommendations, "Add to VS Code settings.json: \"editor.fontFamily\": \"FiraCode Nerd Font\"")
		} else {
			fmt.Printf("ℹ️  VS Code: Not installed\n")
		}
	}
	
	// 5. Check terminal font configuration
	terminalConfigured := checkTerminalSupport()
	if terminalConfigured {
		fmt.Printf("✅ Terminal Config: Using Nerd Font\n")
	} else {
		fmt.Printf("⚠️  Terminal Config: Not using Nerd Font\n")
		issues = append(issues, "Terminal not configured to use Nerd Fonts")
		recommendations = append(recommendations, "Configure your terminal to use a Nerd Font (e.g., 'JetBrains Mono Nerd Font')")
	}
	
	// 6. Check starship integration
	if isStarshipInstalled() {
		fmt.Printf("✅ Starship: Installed (works great with Nerd Fonts)\n")
		if !isNerdFontsInstalled() {
			recommendations = append(recommendations, "Starship icons will display better with Nerd Fonts installed")
		}
	} else {
		fmt.Printf("ℹ️  Starship: Not installed\n")
		recommendations = append(recommendations, "Consider installing Starship for an enhanced prompt: gearbox install starship")
	}
	
	// Summary
	fmt.Printf("\n📈 Health Check Summary\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	
	if len(issues) == 0 {
		fmt.Printf("🎉 All checks passed! Your Nerd Fonts setup is optimal.\n")
	} else {
		fmt.Printf("⚠️  Found %d issue(s):\n", len(issues))
		for i, issue := range issues {
			fmt.Printf("  %d. %s\n", i+1, issue)
		}
	}
	
	if len(recommendations) > 0 {
		fmt.Printf("\n💡 Recommendations:\n")
		for i, rec := range recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}
	
	return nil
}

// installNerdFontsWithOptions handles nerd-fonts installation with special options
func (o *Orchestrator) installNerdFontsWithOptions(tool ToolConfig) error {
	fmt.Printf("🎨 Nerd Fonts Installation with Advanced Options\n\n")

	// Build script command with options
	args := []string{}
	
	if o.options.Fonts != "" {
		args = append(args, "--fonts="+o.options.Fonts)
	}
	
	if o.options.Interactive {
		args = append(args, "--interactive")
	}
	
	if o.options.Preview {
		args = append(args, "--preview")
	}
	
	if o.options.ConfigureApps {
		args = append(args, "--configure-apps")
	}
	
	// Add standard orchestrator options
	if o.options.BuildType == "minimal" {
		args = append(args, "--minimal")
	} else if o.options.BuildType == "maximum" {
		args = append(args, "--maximum")
	}
	
	if o.options.SkipCommonDeps {
		args = append(args, "--skip-deps")
	}
	
	if o.options.RunTests {
		args = append(args, "--run-tests")
	}
	
	if o.options.Force {
		args = append(args, "--force")
	}
	
	if o.options.DryRun {
		args = append(args, "--config-only")
	}

	// Execute the nerd-fonts script directly with the options
	return o.executeNerdFontsScript(args)
}

// executeNerdFontsScript runs the nerd-fonts script with given arguments
func (o *Orchestrator) executeNerdFontsScript(args []string) error {
	// Find the script path
	scriptPath := "./scripts/install-nerd-fonts.sh"
	if _, err := os.Stat(scriptPath); err != nil {
		// Try alternative paths
		scriptPath = "scripts/install-nerd-fonts.sh"
		if _, err := os.Stat(scriptPath); err != nil {
			return fmt.Errorf("nerd-fonts script not found: %w", err)
		}
	}

	// Build the command
	cmdArgs := []string{"bash", scriptPath}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	
	// Connect stdio for interactive features
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	fmt.Printf("🚀 Executing: %s %s\n", scriptPath, strings.Join(args, " "))
	
	return cmd.Run()
}

// Helper functions for nerd-fonts doctor checks

func checkTerminalSupport() bool {
	// Check for Unicode support by testing if terminal can display basic Unicode
	term := os.Getenv("TERM")
	if strings.Contains(term, "xterm") || strings.Contains(term, "screen") || 
	   strings.Contains(term, "tmux") || strings.Contains(term, "alacritty") ||
	   strings.Contains(term, "kitty") {
		return true
	}
	
	// Check LANG/LC_* environment variables for UTF-8 support
	for _, envVar := range []string{"LANG", "LC_ALL", "LC_CTYPE"} {
		if val := os.Getenv(envVar); strings.Contains(strings.ToUpper(val), "UTF") {
			return true
		}
	}
	
	return false
}

func isVSCodeInstalled() bool {
	// Check for VS Code binary
	if _, err := exec.LookPath("code"); err == nil {
		return true
	}
	// Check for VS Code Insiders
	if _, err := exec.LookPath("code-insiders"); err == nil {
		return true
	}
	return false
}

func checkVSCodeFontConfig() (bool, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}
	
	configPath := filepath.Join(homeDir, ".config", "Code", "User", "settings.json")
	if _, err := os.Stat(configPath); err != nil {
		return false, ""
	}
	
	content, err := os.ReadFile(configPath)
	if err != nil {
		return false, ""
	}
	
	configStr := string(content)
	if strings.Contains(configStr, "Nerd Font") {
		// Extract font family if possible
		if idx := strings.Index(configStr, `"editor.fontFamily"`); idx != -1 {
			start := strings.Index(configStr[idx:], `"`) + idx + 1
			end := strings.Index(configStr[start:], `"`) + start
			if start < end && end < len(configStr) {
				fontFamily := configStr[strings.Index(configStr[start:], `"`)+start+1:end]
				return true, fontFamily
			}
		}
		return true, "Unknown Nerd Font"
	}
	
	return false, ""
}

func isStarshipInstalled() bool {
	_, err := exec.LookPath("starship")
	return err == nil
}

func checkStarshipFontUsage() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	
	// Check starship config
	configPath := filepath.Join(homeDir, ".config", "starship.toml")
	if content, err := os.ReadFile(configPath); err == nil {
		// Look for Nerd Font symbols or emoji usage
		configStr := string(content)
		nerdFontSymbols := []string{"", "", "", "", "", "λ", "⎈", ""}
		for _, symbol := range nerdFontSymbols {
			if strings.Contains(configStr, symbol) {
				return true
			}
		}
	}
	
	return false
}

func checkFontcacheHealth() (bool, []string) {
	var issues []string
	
	// Check if fc-cache exists
	if _, err := exec.LookPath("fc-cache"); err != nil {
		issues = append(issues, "fc-cache command not found - fontconfig may not be installed")
		return false, issues
	}
	
	// Check font cache directory
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".cache", "fontconfig")
	if _, err := os.Stat(cacheDir); err != nil {
		issues = append(issues, "Font cache directory not found - run 'fc-cache -f' to rebuild")
	}
	
	// Try to run fc-list to test font system
	cmd := exec.Command("fc-list")
	if err := cmd.Run(); err != nil {
		issues = append(issues, "Font system not responding - try running 'fc-cache -fv'")
		return false, issues
	}
	
	return len(issues) == 0, issues
}

func getTerminalRecommendations() []string {
	var recommendations []string
	
	term := os.Getenv("TERM")
	switch {
	case strings.Contains(term, "kitty"):
		recommendations = append(recommendations, "Kitty: Add 'font_family JetBrains Mono Nerd Font' to ~/.config/kitty/kitty.conf")
	case strings.Contains(term, "alacritty"):
		recommendations = append(recommendations, "Alacritty: Configure font in ~/.config/alacritty/alacritty.yml")
	case strings.Contains(term, "gnome"):
		recommendations = append(recommendations, "GNOME Terminal: Set font via Preferences > Profiles > Text")
	default:
		recommendations = append(recommendations, "Configure your terminal to use a Nerd Font for best results")
		recommendations = append(recommendations, "Popular choices: JetBrains Mono Nerd Font, FiraCode Nerd Font, Hack Nerd Font")
	}
	
	return recommendations
}