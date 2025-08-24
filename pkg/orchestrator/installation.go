package orchestrator

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// InstallTools orchestrates the installation of specified tools
func (o *Orchestrator) InstallTools(toolNames []string) error {
	// Track which tools come from which bundles for progress display
	bundleToolMap := make(map[string][]string)
	var directTools []string
	
	// First, expand any bundles and track their tools
	for _, name := range toolNames {
		if o.isBundle(name, o.bundleConfig.Bundles) {
			visited := make(map[string]bool)
			expandedTools, err := o.expandBundle(name, o.bundleConfig.Bundles, visited)
			if err != nil {
				return fmt.Errorf("failed to expand bundle %s: %w", name, err)
			}
			bundleToolMap[name] = expandedTools
		} else {
			directTools = append(directTools, name)
		}
	}
	
	// Get all unique tools
	expandedToolNames, err := o.expandBundlesAndTools(toolNames)
	if err != nil {
		return fmt.Errorf("failed to expand bundles: %w", err)
	}
	
	// Show installation header with bundle context
	if len(bundleToolMap) > 0 {
		fmt.Printf("🔧 Gearbox Orchestrator - Installing %d tools", len(expandedToolNames))
		if len(bundleToolMap) == 1 {
			for bundleName := range bundleToolMap {
				fmt.Printf(" (bundle: %s)", bundleName)
			}
		} else {
			fmt.Printf(" (from %d bundles)", len(bundleToolMap))
		}
		if len(directTools) > 0 {
			fmt.Printf(" + %d direct tools", len(directTools))
		}
		fmt.Printf("\n\n")
		
		// Show bundle breakdown
		for bundleName, tools := range bundleToolMap {
			fmt.Printf("📦 Bundle '%s': %d tools\n", bundleName, len(tools))
		}
		if len(directTools) > 0 {
			fmt.Printf("🔧 Direct tools: %d\n", len(directTools))
		}
		fmt.Printf("\n")
	} else {
		fmt.Printf("🔧 Gearbox Orchestrator - Installing %d tools\n\n", len(expandedToolNames))
	}

	// Validate tool names
	var validTools []ToolConfig
	for _, name := range expandedToolNames {
		tool, found := o.findTool(name)
		if !found {
			return fmt.Errorf("tool not found: %s", name)
		}
		validTools = append(validTools, tool)
	}

	// Handle nerd-fonts specific options
	if len(validTools) == 1 && validTools[0].Name == "nerd-fonts" {
		if o.options.Fonts != "" || o.options.Interactive || o.options.Preview || o.options.ConfigureApps {
			return o.installNerdFontsWithOptions(validTools[0])
		}
	}

	// Check for cross-tool recommendations
	o.suggestRelatedTools(validTools)

	// Resolve dependencies and determine installation order
	installOrder, err := o.resolveDependencies(validTools)
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	if o.options.DryRun {
		return o.showDryRun(installOrder, toolNames)
	}

	// Show installation plan
	o.showInstallationPlan(installOrder)

	// Install system packages first (if any)
	if err := o.installSystemPackagesFromBundles(toolNames); err != nil {
		return fmt.Errorf("failed to install system packages: %w", err)
	}

	// Install common dependencies first (unless skipped)
	if !o.options.SkipCommonDeps {
		fmt.Printf("📦 Installing common dependencies...\n")
		if err := o.installCommonDependencies(); err != nil {
			return fmt.Errorf("failed to install common dependencies: %w", err)
		}
		fmt.Printf("✅ Common dependencies installed\n\n")
	}

	// Execute installations with progress tracking
	fmt.Printf("🚀 Starting installations...\n")
	o.progressBar = progressbar.NewOptions(len(installOrder),
		progressbar.OptionSetDescription("Installing tools"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	if err := o.executeInstallations(installOrder); err != nil {
		return err
	}

	// Show results
	return o.showResults()
}

// installSystemPackagesFromBundles installs system packages for any bundles in the tool list
func (o *Orchestrator) installSystemPackagesFromBundles(toolNames []string) error {
	if o.packageMgr == nil {
		return nil // No package manager available, skip system packages
	}
	
	var allSystemPackages []string
	
	// Check each tool name to see if it's a bundle with system packages
	for _, name := range toolNames {
		if o.isBundle(name, o.bundleConfig.Bundles) {
			visited := make(map[string]bool)
			packages, err := o.expandSystemPackages(name, o.bundleConfig.Bundles, visited, o.packageMgr.Name)
			if err != nil {
				return fmt.Errorf("failed to expand system packages from bundle %s: %w", name, err)
			}
			allSystemPackages = append(allSystemPackages, packages...)
		}
	}
	
	if len(allSystemPackages) == 0 {
		return nil // No system packages to install
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	var uniquePackages []string
	for _, pkg := range allSystemPackages {
		if !seen[pkg] {
			seen[pkg] = true
			uniquePackages = append(uniquePackages, pkg)
		}
	}
	
	if o.options.DryRun {
		fmt.Printf("📦 System packages that would be installed (%s): %s\n\n", o.packageMgr.Name, strings.Join(uniquePackages, ", "))
		return nil
	}
	
	// Install system packages
	return o.packageMgr.installPackages(uniquePackages, o.options.DryRun)
}

// resolveDependencies resolves dependencies and determines optimal installation order
func (o *Orchestrator) resolveDependencies(tools []ToolConfig) ([]ToolConfig, error) {
	// Group tools by language for optimal installation order
	languageGroups := make(map[string][]ToolConfig)
	for _, tool := range tools {
		languageGroups[tool.Language] = append(languageGroups[tool.Language], tool)
	}

	// Define language priority order (Go first for bootstrapping, then Rust, then others)
	languageOrder := []string{"go", "rust", "python", "c"}
	
	var installOrder []ToolConfig
	addedTools := make(map[string]bool)

	// Add tools in language priority order
	for _, lang := range languageOrder {
		if langTools, exists := languageGroups[lang]; exists {
			// Sort tools within language group by name for deterministic order
			sort.Slice(langTools, func(i, j int) bool {
				return langTools[i].Name < langTools[j].Name
			})
			
			for _, tool := range langTools {
				if !addedTools[tool.Name] {
					installOrder = append(installOrder, tool)
					addedTools[tool.Name] = true
				}
			}
		}
	}

	// Add any remaining tools not in the predefined language groups
	for _, tool := range tools {
		if !addedTools[tool.Name] {
			installOrder = append(installOrder, tool)
			addedTools[tool.Name] = true
		}
	}

	return installOrder, nil
}

// showDryRun displays what would be installed without executing
func (o *Orchestrator) showDryRun(tools []ToolConfig, originalToolNames []string) error {
	fmt.Printf("🔍 Dry Run - Installation Plan\n\n")
	fmt.Printf("Build Type: %s\n", o.options.BuildType)
	fmt.Printf("Max Parallel Jobs: %d\n", o.options.MaxParallelJobs)
	fmt.Printf("Skip Common Deps: %v\n", o.options.SkipCommonDeps)
	fmt.Printf("Run Tests: %v\n", o.options.RunTests)
	fmt.Printf("Shell Integration: %v\n", !o.options.NoShell)
	
	// Show system packages if any
	o.showSystemPackagesPlan(originalToolNames)
	
	fmt.Printf("\nInstallation Order:\n")

	for i, tool := range tools {
		buildFlag := tool.BuildTypes[o.options.BuildType]
		if buildFlag == "" {
			buildFlag = "(default)"
		}
		fmt.Printf("  %2d. %-15s (%s) - Build flag: %s\n", 
			i+1, tool.Name, tool.Language, buildFlag)
	}

	fmt.Printf("\nTotal tools to install: %d\n", len(tools))
	return nil
}

// showSystemPackagesPlan displays system packages that would be installed in dry-run mode
func (o *Orchestrator) showSystemPackagesPlan(toolNames []string) {
	if o.packageMgr == nil {
		return
	}
	
	var allSystemPackages []string
	
	// Check each tool name to see if it's a bundle with system packages
	for _, name := range toolNames {
		if o.isBundle(name, o.bundleConfig.Bundles) {
			visited := make(map[string]bool)
			packages, err := o.expandSystemPackages(name, o.bundleConfig.Bundles, visited, o.packageMgr.Name)
			if err == nil {
				allSystemPackages = append(allSystemPackages, packages...)
			}
		}
	}
	
	if len(allSystemPackages) > 0 {
		// Remove duplicates
		seen := make(map[string]bool)
		var uniquePackages []string
		for _, pkg := range allSystemPackages {
			if !seen[pkg] {
				seen[pkg] = true
				uniquePackages = append(uniquePackages, pkg)
			}
		}
		
		fmt.Printf("\nSystem packages that would be installed (%s):\n", o.packageMgr.Name)
		for _, pkg := range uniquePackages {
			fmt.Printf("  - %s\n", pkg)
		}
	}
}

// showInstallationPlan displays the installation plan
func (o *Orchestrator) showInstallationPlan(tools []ToolConfig) {
	fmt.Printf("📋 Installation Plan\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Build Type: %s\n", o.options.BuildType)
	fmt.Printf("Parallel Jobs: %d\n", o.options.MaxParallelJobs)
	fmt.Printf("Total Tools: %d\n\n", len(tools))

	// Group by language for display
	languageGroups := make(map[string][]ToolConfig)
	for _, tool := range tools {
		languageGroups[tool.Language] = append(languageGroups[tool.Language], tool)
	}

	for lang, langTools := range languageGroups {
		fmt.Printf("📦 %s (%d tools): ", strings.Title(lang), len(langTools))
		var names []string
		for _, tool := range langTools {
			names = append(names, tool.Name)
		}
		fmt.Printf("%s\n", strings.Join(names, ", "))
	}
	fmt.Println()
}

// installCommonDependencies installs common dependencies
func (o *Orchestrator) installCommonDependencies() error {
	commonDepsScript := filepath.Join(o.scriptsDir, "installation", "common", "install-common-deps.sh")
	
	if _, err := os.Stat(commonDepsScript); os.IsNotExist(err) {
		return fmt.Errorf("common dependencies script not found: %s", commonDepsScript)
	}

	cmd := exec.Command("bash", commonDepsScript)
	cmd.Dir = o.repoDir
	
	if o.options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// executeInstallations executes tool installations with parallel processing
func (o *Orchestrator) executeInstallations(tools []ToolConfig) error {
	semaphore := make(chan struct{}, o.options.MaxParallelJobs)
	var wg sync.WaitGroup
	errorChan := make(chan error, len(tools))

	for _, tool := range tools {
		wg.Add(1)
		go func(t ToolConfig) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := o.installTool(t)
			
			o.mu.Lock()
			o.results = append(o.results, result)
			o.mu.Unlock()

			if result.Error != nil {
				errorChan <- fmt.Errorf("failed to install %s: %w", t.Name, result.Error)
			}

			// Update progress
			o.progressBar.Add(1)
		}(tool)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		fmt.Printf("\n❌ Installation completed with %d errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  • %v\n", err)
		}
		return fmt.Errorf("%d tools failed to install", len(errors))
	}

	return nil
}

// findToolScript finds the installation script for a tool by searching category directories
func (o *Orchestrator) findToolScript(toolName string) string {
	// List of category directories to search
	categories := []string{"core", "development", "system", "text", "media", "ui"}
	
	scriptName := fmt.Sprintf("install-%s.sh", toolName)
	
	// Search in each category directory
	for _, category := range categories {
		scriptPath := filepath.Join(o.scriptsDir, "installation", "categories", category, scriptName)
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath
		}
	}
	
	// Fallback: try the old path (directly in scripts dir)
	fallbackPath := filepath.Join(o.scriptsDir, scriptName)
	return fallbackPath
}

// installTool installs a single tool
func (o *Orchestrator) installTool(tool ToolConfig) InstallationResult {
	start := time.Now()
	
	// Find the script in the appropriate category directory
	scriptPath := o.findToolScript(tool.Name)
	
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return InstallationResult{
			Tool:     tool,
			Success:  false,
			Error:    fmt.Errorf("installation script not found: %s", scriptPath),
			Duration: time.Since(start),
		}
	}

	// Build command arguments using standard protocol
	var args []string
	args = append(args, scriptPath)

	// Add standardized build type flag
	switch o.options.BuildType {
	case "minimal":
		args = append(args, "--minimal")
	case "standard":
		args = append(args, "--standard")
	case "maximum":
		args = append(args, "--maximum")
	default:
		// Default to standard if unknown build type
		args = append(args, "--standard")
	}

	// Add common options using standard protocol
	args = append(args, "--skip-deps") // Dependencies handled separately
	args = append(args, "--force")     // Always force to avoid interactive prompts
	
	if o.options.RunTests {
		args = append(args, "--run-tests")
	}
	
	if o.options.NoShell && tool.ShellIntegration {
		args = append(args, "--no-shell")
	}
	
	if o.options.Verbose {
		args = append(args, "--verbose")
	}
	
	if o.options.DryRun {
		args = append(args, "--dry-run")
	}

	// Execute installation
	cmd := exec.Command("bash", args...)
	
	// Set working directory to build directory (~/tools/build)
	buildDir := os.ExpandEnv("$HOME/tools/build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return InstallationResult{
			Tool:     tool,
			Success:  false,
			Error:    fmt.Errorf("failed to create build directory: %w", err),
			Duration: time.Since(start),
		}
	}
	cmd.Dir = buildDir

	var output strings.Builder
	if o.options.Verbose {
		cmd.Stdout = io.MultiWriter(os.Stdout, &output)
		cmd.Stderr = io.MultiWriter(os.Stderr, &output)
	} else {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}

	// Provide automatic "yes" responses to avoid interactive prompts
	cmd.Stdin = strings.NewReader("y\ny\ny\ny\ny\ny\ny\ny\ny\ny\n")

	err := cmd.Run()
	
	return InstallationResult{
		Tool:     tool,
		Success:  err == nil,
		Error:    err,
		Duration: time.Since(start),
		Output:   output.String(),
	}
}

// showResults displays installation results
func (o *Orchestrator) showResults() error {
	fmt.Printf("\n\n📊 Installation Results\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	var successful, failed int
	var totalDuration time.Duration

	// Sort results by success status (successful first)
	sort.Slice(o.results, func(i, j int) bool {
		if o.results[i].Success != o.results[j].Success {
			return o.results[i].Success
		}
		return o.results[i].Tool.Name < o.results[j].Tool.Name
	})

	for _, result := range o.results {
		totalDuration += result.Duration
		
		if result.Success {
			successful++
			fmt.Printf("✅ %-15s (%6.1fs) - %s\n", 
				result.Tool.Name, 
				result.Duration.Seconds(),
				result.Tool.Description)
		} else {
			failed++
			fmt.Printf("❌ %-15s (%6.1fs) - %v\n", 
				result.Tool.Name, 
				result.Duration.Seconds(),
				result.Error)
		}
	}

	fmt.Printf("\n📈 Summary\n")
	fmt.Printf("Successful: %d\n", successful)
	fmt.Printf("Failed: %d\n", failed)
	fmt.Printf("Total Duration: %.1fs\n", totalDuration.Seconds())
	fmt.Printf("Average Duration: %.1fs\n", totalDuration.Seconds()/float64(len(o.results)))

	if failed == 0 {
		fmt.Printf("\n🎉 All tools installed successfully!\n")
		return nil
	} else {
		return fmt.Errorf("%d tools failed to install", failed)
	}
}

// suggestRelatedTools analyzes the tools being installed and suggests related tools
func (o *Orchestrator) suggestRelatedTools(tools []ToolConfig) {
	var suggestions []string
	var installingNames []string
	
	// Get names of tools being installed
	for _, tool := range tools {
		installingNames = append(installingNames, tool.Name)
	}
	
	// Check for specific tool relationships
	hasStarship := contains(installingNames, "starship")
	hasNerdFonts := contains(installingNames, "nerd-fonts")
	hasFzf := contains(installingNames, "fzf")
	hasBat := contains(installingNames, "bat")
	hasEza := contains(installingNames, "eza")
	hasDelta := contains(installingNames, "delta")
	
	// Starship + Nerd Fonts relationship
	if hasStarship && !hasNerdFonts && !isNerdFontsInstalled() {
		suggestions = append(suggestions, "🎨 Consider adding 'nerd-fonts' - Starship displays icons and symbols much better with Nerd Fonts")
	}
	if hasNerdFonts && !hasStarship && !isStarshipInstalled() {
		suggestions = append(suggestions, "⭐ Consider adding 'starship' - A customizable prompt that works great with Nerd Fonts")
	}
	
	// Terminal enhancement bundle
	if (hasFzf || hasBat || hasEza) && len(installingNames) == 1 {
		missing := []string{}
		if !hasFzf && !isToolInConfig(o.configMgr, "fzf") {
			missing = append(missing, "fzf")
		}
		if !hasBat && !isToolInConfig(o.configMgr, "bat") {
			missing = append(missing, "bat")
		}
		if !hasEza && !isToolInConfig(o.configMgr, "eza") {
			missing = append(missing, "eza")
		}
		
		if len(missing) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("🚀 Terminal bundle: Consider also installing %s for a complete terminal experience", strings.Join(missing, ", ")))
		}
	}
	
	// Git workflow enhancement
	if hasDelta && !contains(installingNames, "lazygit") && !isToolInConfig(o.configMgr, "lazygit") {
		suggestions = append(suggestions, "🔧 Consider adding 'lazygit' - A terminal UI for Git that pairs well with Delta")
	}
	
	// Development tools bundle
	developmentTools := []string{"ripgrep", "fd", "sd", "tokei"}
	installingDev := 0
	for _, devTool := range developmentTools {
		if contains(installingNames, devTool) {
			installingDev++
		}
	}
	
	if installingDev >= 1 && installingDev < len(developmentTools) {
		missing := []string{}
		for _, devTool := range developmentTools {
			if !contains(installingNames, devTool) && !isToolInConfig(o.configMgr, devTool) {
				missing = append(missing, devTool)
			}
		}
		if len(missing) > 0 && len(missing) <= 2 {
			suggestions = append(suggestions, fmt.Sprintf("💻 Development bundle: Consider also installing %s", strings.Join(missing, ", ")))
		}
	}
	
	// Show suggestions if any
	if len(suggestions) > 0 {
		fmt.Printf("\n💡 Related Tool Suggestions\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		for _, suggestion := range suggestions {
			fmt.Printf("   %s\n", suggestion)
		}
		fmt.Printf("\n   Use: gearbox install <additional_tools>\n\n")
	}
}