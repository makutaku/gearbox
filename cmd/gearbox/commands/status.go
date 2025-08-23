package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [TOOLS...]",
		Short: "Show installation status of tools",
		Long: `Display the installation status of development tools.

Shows which tools are installed, their versions, and installation paths.
If specific tools are provided, only those tools are checked.

Special tool status:
- nerd-fonts: Shows installed font families and variants

Examples:
  gearbox status                  # Show status for all tools
  gearbox status fd ripgrep       # Show status for specific tools
  gearbox status nerd-fonts       # Show detailed Nerd Fonts status`,
		RunE: runStatus,
	}

	cmd.Flags().BoolP("all", "a", false, "Show status for all tools (default)")
	cmd.Flags().BoolP("installed", "i", false, "Show only installed tools")
	cmd.Flags().BoolP("missing", "m", false, "Show only missing tools")
	cmd.Flags().BoolP("detailed", "d", false, "Show detailed information")
	cmd.Flags().Bool("manifest-only", false, "Show only manifest-tracked tools")
	cmd.Flags().Bool("unified", false, "Show unified view (manifest + live detection)")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "orchestrator")

	// This command requires the orchestrator for proper status tracking
	if _, err := os.Stat(orchestratorPath); err == nil {
		return runStatusWithOrchestrator(orchestratorPath, cmd, args)
	}

	// Fallback to basic status check
	return runBasicStatus(args)
}

func runStatusWithOrchestrator(orchestratorPath string, cmd *cobra.Command, args []string) error {
	// Check if nerd-fonts status is requested
	if len(args) == 1 && args[0] == "nerd-fonts" {
		return showNerdFontsStatus()
	}
	
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
	if manifestOnly, _ := cmd.Flags().GetBool("manifest-only"); manifestOnly {
		statusCmd.Args = append(statusCmd.Args, "--manifest-only")
	}
	if unified, _ := cmd.Flags().GetBool("unified"); unified {
		statusCmd.Args = append(statusCmd.Args, "--unified")
	}

	statusCmd.Stdout = os.Stdout
	statusCmd.Stderr = os.Stderr
	
	return statusCmd.Run()
}

func runBasicStatus(tools []string) error {
	// If specific tools provided, check only those
	if len(tools) > 0 {
		for _, tool := range tools {
			if tool == "nerd-fonts" {
				return showNerdFontsStatus()
			} else {
				checkToolStatus(tool)
			}
		}
		return nil
	}

	fmt.Println("📊 Tool Installation Status")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Note: This is a basic status check. For comprehensive status")
	fmt.Println("information, please rebuild the project to enable the orchestrator.")
	fmt.Println()

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

// showNerdFontsStatus displays detailed information about installed Nerd Fonts
func showNerdFontsStatus() error {
	fmt.Println("🎨 Nerd Fonts Installation Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Get installed fonts
	installedFonts, err := getNerdFontsList()
	if err != nil {
		fmt.Printf("❌ Error checking fonts: %v\n", err)
		return err
	}

	if len(installedFonts) == 0 {
		fmt.Println("❌ No Nerd Fonts detected")
		fmt.Println("   Run: gearbox install nerd-fonts")
		return nil
	}

	// Group fonts by family
	fontFamilies := groupFontsByFamily(installedFonts)
	
	fmt.Printf("✅ Found %d Nerd Fonts from %d font families\n", len(installedFonts), len(fontFamilies))
	fmt.Println()

	// Show fonts directory info
	fontsDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "fonts")
	if size := getDirSize(fontsDir); size > 0 {
		fmt.Printf("📁 Location: %s (%s)\n", fontsDir, humanReadableSize(size))
	} else {
		fmt.Printf("📁 Location: %s\n", fontsDir)
	}
	fmt.Println()

	// List font families
	fmt.Println("📋 Installed Font Families:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	familyNames := make([]string, 0, len(fontFamilies))
	for family := range fontFamilies {
		familyNames = append(familyNames, family)
	}
	sort.Strings(familyNames)

	for i, family := range familyNames {
		variants := fontFamilies[family]
		if len(variants) == 1 {
			fmt.Printf("%2d. %s\n", i+1, family)
		} else {
			fmt.Printf("%2d. %s (%d variants)\n", i+1, family, len(variants))
		}
	}

	fmt.Println()
	fmt.Println("💡 Usage Tips:")
	fmt.Println("   • Use 'gearbox doctor nerd-fonts' for health checks")
	fmt.Println("   • Use 'gearbox doctor nerd-fonts --verbose' for detailed analysis") 
	fmt.Println("   • Use 'fc-list | grep -i nerd' to see all font variants")

	return nil
}

// getNerdFontsList returns a list of installed Nerd Fonts
func getNerdFontsList() ([]string, error) {
	cmd := exec.Command("fc-list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var fonts []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "nerd") {
			// Extract font name from fc-list output: path: family,style
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

// groupFontsByFamily groups fonts by their family name
func groupFontsByFamily(fonts []string) map[string][]string {
	families := make(map[string][]string)
	
	for _, font := range fonts {
		// Extract family name (before the first comma if present)
		family := font
		if commaIndex := strings.Index(font, ","); commaIndex != -1 {
			family = strings.TrimSpace(font[:commaIndex])
		}
		
		// Remove common suffixes to group similar fonts
		family = strings.TrimSuffix(family, " Nerd Font")
		family = strings.TrimSuffix(family, " Nerd Font Mono")
		family = strings.TrimSuffix(family, " Nerd Font Propo")
		family = strings.TrimSuffix(family, " NF")
		family = strings.TrimSuffix(family, " NFM")
		family = strings.TrimSuffix(family, " NFP")
		
		families[family] = append(families[family], font)
	}
	
	return families
}

// Helper functions - these are declared in doctor.go
// getDirSize and humanReadableSize are available from doctor.go