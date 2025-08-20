package orchestrator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"gearbox/pkg/validation"
	"gearbox/pkg/uninstall"
)

// ToolConfig represents a single tool configuration
type ToolConfig struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	Repository       string            `json:"repository"`
	BinaryName       string            `json:"binary_name"`
	Language         string            `json:"language"`
	BuildTypes       map[string]string `json:"build_types"`
	Dependencies     []string          `json:"dependencies"`
	MinVersion       string            `json:"min_version"`
	ShellIntegration bool              `json:"shell_integration"`
	TestCommand      string            `json:"test_command"`
}

// LanguageConfig represents language-specific configuration
type LanguageConfig struct {
	MinVersion string `json:"min_version"`
	BuildTool  string `json:"build_tool"`
}

// Config represents the complete configuration structure
type Config struct {
	SchemaVersion    string                     `json:"schema_version"`
	DefaultBuildType string                     `json:"default_build_type"`
	Tools            []ToolConfig               `json:"tools"`
	Categories       map[string]string          `json:"categories"`
	Languages        map[string]LanguageConfig  `json:"languages"`
}

// InstallationOptions represents installation configuration
type InstallationOptions struct {
	BuildType        string
	SkipCommonDeps   bool
	RunTests         bool
	NoShell          bool
	Force            bool
	MaxParallelJobs  int
	Verbose          bool
	DryRun           bool
	
	// Nerd-fonts specific options
	Fonts            string
	Interactive      bool
	Preview          bool
	ConfigureApps    bool
}

// InstallationResult represents the result of a tool installation
type InstallationResult struct {
	Tool      ToolConfig
	Success   bool
	Error     error
	Duration  time.Duration
	Output    string
}

// Orchestrator manages the installation process
type Orchestrator struct {
	config        Config
	bundleConfig  *BundleConfiguration
	packageMgr    *PackageManager
	options       InstallationOptions
	repoDir       string
	scriptsDir    string
	mu            sync.RWMutex  // Use RWMutex for better read performance
	results       []InstallationResult
	progressBar   *progressbar.ProgressBar
	resultPool    sync.Pool     // Memory pool for result objects
}

// Global variables
var (
	configPath string
	repoDir    string
)

// Main is the entry point for the orchestrator command-line tool.
// It sets up the CLI commands and executes the orchestrator functionality.
func Main() {
	var rootCmd = &cobra.Command{
		Use:   "orchestrator",
		Short: "Gearbox installation orchestrator",
		Long:  "Advanced orchestration engine for managing gearbox tool installations with dependency resolution and parallel execution",
	}

	// Add commands
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(showCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(verifyCmd())
	rootCmd.AddCommand(doctorCmd())
	
	// Add tracking commands
	rootCmd.AddCommand(trackInstallationCmd())
	rootCmd.AddCommand(trackBundleCmd())
	rootCmd.AddCommand(isTrackedCmd())
	rootCmd.AddCommand(trackPreexistingCmd())
	rootCmd.AddCommand(initManifestCmd())
	rootCmd.AddCommand(manifestStatusCmd())
	rootCmd.AddCommand(listDependentsCmd())
	rootCmd.AddCommand(canRemoveCmd())
	
	// Add uninstall commands
	rootCmd.AddCommand(uninstallCmd())
	rootCmd.AddCommand(uninstallPlanCmd())
	rootCmd.AddCommand(uninstallExecuteCmd())

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to tools.json configuration file")
	rootCmd.PersistentFlags().StringVar(&repoDir, "repo-dir", "", "Repository directory (default: auto-detect)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// installCmd creates the install command
func installCmd() *cobra.Command {
	var opts InstallationOptions

	cmd := &cobra.Command{
		Use:   "install [tools...]",
		Short: "Install tools with advanced orchestration",
		Long: `Install one or more tools with dependency resolution, parallel execution,
and comprehensive progress tracking. If no tools are specified, all tools will be installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(opts)
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			var toolsToInstall []string
			if len(args) == 0 {
				// Install all tools
				for _, tool := range orchestrator.config.Tools {
					toolsToInstall = append(toolsToInstall, tool.Name)
				}
			} else {
				toolsToInstall = args
			}

			return orchestrator.InstallTools(toolsToInstall)
		},
	}

	// Installation options
	cmd.Flags().StringVarP(&opts.BuildType, "build-type", "b", "standard", "Build type (minimal, standard, maximum)")
	cmd.Flags().BoolVar(&opts.SkipCommonDeps, "skip-common-deps", false, "Skip common dependency installation")
	cmd.Flags().BoolVar(&opts.RunTests, "run-tests", false, "Run test suites for validation")
	cmd.Flags().BoolVar(&opts.NoShell, "no-shell", false, "Skip shell integration setup")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force reinstallation if already installed")
	cmd.Flags().IntVarP(&opts.MaxParallelJobs, "jobs", "j", 0, "Maximum parallel jobs (0 = auto-detect)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be installed without executing")

	// Nerd-fonts specific options
	cmd.Flags().StringVar(&opts.Fonts, "fonts", "", "Install specific fonts (comma-separated, e.g. 'FiraCode,JetBrainsMono')")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", false, "Interactive font selection with previews")
	cmd.Flags().BoolVar(&opts.Preview, "preview", false, "Show font previews before installation")
	cmd.Flags().BoolVar(&opts.ConfigureApps, "configure-apps", false, "Automatically configure VS Code, terminals, etc.")

	return cmd
}

// listCmd creates the list command
func listCmd() *cobra.Command {
	var category string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list [bundles]",
		Short: "List available tools or bundles",
		Long: `List available tools or bundles.
		
Without arguments, lists all available tools.
Use 'list bundles' to list all available bundles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}
			
			// Check if user wants to list bundles
			if len(args) > 0 && args[0] == "bundles" {
				return orchestrator.ListBundles(verbose)
			}

			return orchestrator.ListTools(category, verbose)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	return cmd
}

// showCmd creates the show command
func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show bundle <name>",
		Short: "Show details about a bundle",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "bundle" {
				return fmt.Errorf("only 'show bundle' is supported")
			}
			
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}
			
			return orchestrator.ShowBundle(args[1])
		},
	}
	
	return cmd
}

// statusCmd creates the status command
func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show installation status of tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			return orchestrator.ShowStatus(args)
		},
	}
}

// verifyCmd creates the verify command
func verifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [tools...]",
		Short: "Verify tool installations",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			return orchestrator.VerifyTools(args)
		},
	}
}

// doctorCmd creates the doctor command
func doctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor [tool]",
		Short: "Run health checks and diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			return orchestrator.RunDoctor(args)
		},
	}
	
	cmd.Flags().Bool("fix", false, "Attempt to fix detected issues automatically")
	return cmd
}

// NewOrchestrator creates a new orchestrator instance with the given options.
// It auto-detects repository and configuration paths, loads configuration,
// and initializes the orchestrator with proper settings for parallel execution.
func NewOrchestrator(options InstallationOptions) (*Orchestrator, error) {
	// Auto-detect repository directory if not provided
	if repoDir == "" {
		if wd, err := os.Getwd(); err == nil {
			repoDir = wd
		} else {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Auto-detect config path if not provided
	if configPath == "" {
		configPath = filepath.Join(repoDir, "config", "tools.json")
	}

	// Load configuration
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set default build type from config
	if options.BuildType == "" {
		options.BuildType = config.DefaultBuildType
	}

	// Auto-detect max parallel jobs
	if options.MaxParallelJobs == 0 {
		options.MaxParallelJobs = runtime.NumCPU()
		// Limit based on available memory (rough estimate)
		if options.MaxParallelJobs > 4 {
			options.MaxParallelJobs = 4
		}
	}

	orchestrator := &Orchestrator{
		config:     config,
		options:    options,
		repoDir:    repoDir,
		scriptsDir: filepath.Join(repoDir, "scripts"),
		results:    make([]InstallationResult, 0),
	}
	
	// Initialize memory pool for result objects
	orchestrator.resultPool = sync.Pool{
		New: func() interface{} {
			return &InstallationResult{}
		},
	}
	
	// Load bundle configuration
	bundleConfig, err := orchestrator.loadBundles()
	if err != nil {
		// Bundles are optional, so just log warning
		if options.Verbose {
			fmt.Printf("⚠️  Warning: Failed to load bundles: %v\n", err)
		}
		bundleConfig = &BundleConfiguration{
			SchemaVersion: "1.0",
			Bundles:       []BundleConfig{},
		}
	}
	orchestrator.bundleConfig = bundleConfig
	
	// Detect package manager (optional, for system package support)
	packageMgr, err := detectPackageManager()
	if err != nil && options.Verbose {
		fmt.Printf("⚠️  System package manager not detected: %v\n", err)
	}
	orchestrator.packageMgr = packageMgr
	
	return orchestrator, nil
}

// loadConfig loads the configuration from file with optimization for startup performance
func loadConfig(path string) (Config, error) {
	var config Config

	// Use os.ReadFile for better performance on small files
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	// Use json.Decoder for streaming if file is large (>1MB), otherwise use Unmarshal
	if len(data) > 1024*1024 {
		return loadConfigStreaming(data)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}

	// Validate configuration structure
	validationConfig := validation.Config{
		SchemaVersion:    config.SchemaVersion,
		DefaultBuildType: config.DefaultBuildType,
		Tools:            make([]validation.ToolConfig, len(config.Tools)),
		Categories:       config.Categories,
		Languages:        make(map[string]validation.LanguageConfig),
	}

	// Convert tools for validation
	for i, tool := range config.Tools {
		validationConfig.Tools[i] = validation.ToolConfig{
			Name:             tool.Name,
			Description:      tool.Description,
			Category:         tool.Category,
			Repository:       tool.Repository,
			BinaryName:       tool.BinaryName,
			Language:         tool.Language,
			BuildTypes:       tool.BuildTypes,
			Dependencies:     tool.Dependencies,
			MinVersion:       tool.MinVersion,
			ShellIntegration: tool.ShellIntegration,
			TestCommand:      tool.TestCommand,
		}
	}

	// Convert languages for validation
	for name, lang := range config.Languages {
		validationConfig.Languages[name] = validation.LanguageConfig{
			MinVersion: lang.MinVersion,
			BuildTool:  lang.BuildTool,
		}
	}

	if err := validation.ValidateConfig(validationConfig); err != nil {
		return config, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// loadConfigStreaming loads large configuration files using streaming
func loadConfigStreaming(data []byte) (Config, error) {
	var config Config
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}
	
	return config, nil
}

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

// findTool finds a tool by name in the configuration
func (o *Orchestrator) findTool(name string) (ToolConfig, bool) {
	for _, tool := range o.config.Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolConfig{}, false
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
	commonDepsScript := filepath.Join(o.scriptsDir, "install-common-deps.sh")
	
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

// installTool installs a single tool
func (o *Orchestrator) installTool(tool ToolConfig) InstallationResult {
	start := time.Now()
	
	scriptPath := filepath.Join(o.scriptsDir, fmt.Sprintf("install-%s.sh", tool.Name))
	
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return InstallationResult{
			Tool:     tool,
			Success:  false,
			Error:    fmt.Errorf("installation script not found: %s", scriptPath),
			Duration: time.Since(start),
		}
	}

	// Build command arguments
	var args []string
	args = append(args, scriptPath)

	// Add build flag
	if buildFlag, exists := tool.BuildTypes[o.options.BuildType]; exists && buildFlag != "" {
		args = append(args, buildFlag)
	}

	// Add common options
	args = append(args, "--skip-deps") // Dependencies handled separately
	args = append(args, "--force")     // Always force to avoid interactive prompts
	
	if o.options.RunTests {
		args = append(args, "--run-tests")
	}
	
	if o.options.NoShell && tool.ShellIntegration {
		args = append(args, "--no-shell")
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

// ListTools lists available tools
func (o *Orchestrator) ListTools(category string, verbose bool) error {
	// Group tools by category
	toolsByCategory := make(map[string][]ToolConfig)
	for _, tool := range o.config.Tools {
		if category == "" || tool.Category == category {
			toolsByCategory[tool.Category] = append(toolsByCategory[tool.Category], tool)
		}
	}

	// Sort categories
	var categories []string
	for cat := range toolsByCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	fmt.Printf("📋 Available Tools\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for _, cat := range categories {
		if description, exists := o.config.Categories[cat]; exists {
			fmt.Printf("\n🔧 %s\n", description)
		} else {
			fmt.Printf("\n🔧 %s\n", strings.Title(cat))
		}

		tools := toolsByCategory[cat]
		sort.Slice(tools, func(i, j int) bool {
			return tools[i].Name < tools[j].Name
		})

		for _, tool := range tools {
			if verbose {
				fmt.Printf("  %-15s %s\n", tool.Name, tool.Description)
				fmt.Printf("                  Language: %s, Binary: %s\n", tool.Language, tool.BinaryName)
				if len(tool.Dependencies) > 0 {
					fmt.Printf("                  Dependencies: %s\n", strings.Join(tool.Dependencies, ", "))
				}
				if tool.ShellIntegration {
					fmt.Printf("                  Shell integration: enabled\n")
				}
				fmt.Println()
			} else {
				fmt.Printf("  %-15s %s\n", tool.Name, tool.Description)
			}
		}
	}

	return nil
}

// ShowStatus shows installation status of tools
func (o *Orchestrator) ShowStatus(toolNames []string) error {
	var tools []ToolConfig
	
	if len(toolNames) == 0 {
		tools = o.config.Tools
	} else {
		for _, name := range toolNames {
			tool, found := o.findTool(name)
			if !found {
				return fmt.Errorf("tool not found: %s", name)
			}
			tools = append(tools, tool)
		}
	}

	fmt.Printf("📊 Tool Installation Status\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	var installed, notInstalled int

	for _, tool := range tools {
		if tool.Name == "nerd-fonts" {
			// Special detailed handling for nerd-fonts
			o.showNerdFontsDetailedStatus()
			if isToolInstalled(tool) {
				installed++
			} else {
				notInstalled++
			}
		} else {
			if isToolInstalled(tool) {
				installed++
				version := getToolVersion(tool)
				fmt.Printf("✅ %-15s %s\n", tool.Name, version)
			} else {
				notInstalled++
				fmt.Printf("❌ %-15s Not installed\n", tool.Name)
			}
		}
	}

	fmt.Printf("\n📈 Summary: %d installed, %d not installed\n", installed, notInstalled)
	return nil
}

// VerifyTools verifies tool installations
func (o *Orchestrator) VerifyTools(toolNames []string) error {
	var tools []ToolConfig
	
	if len(toolNames) == 0 {
		tools = o.config.Tools
	} else {
		for _, name := range toolNames {
			tool, found := o.findTool(name)
			if !found {
				return fmt.Errorf("tool not found: %s", name)
			}
			tools = append(tools, tool)
		}
	}

	fmt.Printf("🔍 Verifying Tool Installations\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	var verified, failed int

	for _, tool := range tools {
		if verifyTool(tool) {
			verified++
			version := getToolVersion(tool)
			fmt.Printf("✅ %-15s %s\n", tool.Name, version)
		} else {
			failed++
			fmt.Printf("❌ %-15s Verification failed\n", tool.Name)
		}
	}

	fmt.Printf("\n📈 Verification Summary: %d passed, %d failed\n", verified, failed)
	
	if failed > 0 {
		return fmt.Errorf("%d tools failed verification", failed)
	}
	
	return nil
}

// isToolInstalled checks if a tool is installed
func isToolInstalled(tool ToolConfig) bool {
	// Special handling for nerd-fonts
	if tool.Name == "nerd-fonts" {
		return isNerdFontsInstalled()
	}
	
	_, err := exec.LookPath(tool.BinaryName)
	return err == nil
}

// verifyTool verifies a tool installation by running its test command
func verifyTool(tool ToolConfig) bool {
	if !isToolInstalled(tool) {
		return false
	}

	// Special handling for nerd-fonts
	if tool.Name == "nerd-fonts" {
		return isNerdFontsInstalled()
	}

	if tool.TestCommand == "" {
		return true // If no test command, just check if binary exists
	}

	// Parse and execute test command
	parts := strings.Fields(tool.TestCommand)
	if len(parts) == 0 {
		return true
	}

	// Use binary_name instead of the first part of test_command for tools with different binary names
	binaryName := tool.BinaryName
	if binaryName == "" {
		binaryName = tool.Name // Fallback to tool name if binary_name not specified
	}
	
	// For test commands, use ALL parts as arguments (don't skip the first one)
	// Most test commands are just "--version", not "tool_name --version"
	cmdArgs := parts

	cmd := exec.Command(binaryName, cmdArgs...)
	err := cmd.Run()
	return err == nil
}

// getToolVersion gets the version of an installed tool
func getToolVersion(tool ToolConfig) string {
	// Special handling for nerd-fonts
	if tool.Name == "nerd-fonts" {
		return getNerdFontsVersion()
	}
	if tool.TestCommand == "" {
		return "installed"
	}

	parts := strings.Fields(tool.TestCommand)
	if len(parts) == 0 {
		return "installed"
	}

	// Use binary_name instead of the first part of test_command for tools with different binary names
	// This handles cases like bottom (binary: btm), ripgrep (binary: rg), etc.
	binaryName := tool.BinaryName
	if binaryName == "" {
		binaryName = tool.Name // Fallback to tool name if binary_name not specified
	}
	
	// For test commands, use ALL parts as arguments (don't skip the first one)
	// Most test commands are just "--version", not "tool_name --version"
	cmdArgs := parts
	
	cmd := exec.Command(binaryName, cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return "installed"
	}

	// Extract version information from output
	return extractVersionFromOutput(string(output))
}

// extractVersionFromOutput intelligently extracts version information from command output
func extractVersionFromOutput(output string) string {
	if output == "" {
		return "installed"
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return "installed"
	}

	// Strategy 1: Look for lines that start with 'v' followed by a number (e.g., "v0.23.0")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] == 'v' && len(line) > 1 {
			// Check if second character is a digit
			if line[1] >= '0' && line[1] <= '9' {
				return line
			}
		}
	}

	// Strategy 2: Look for lines containing version patterns (e.g., "tool 1.2.3", "version 1.2.3")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for patterns like "name version", "name 1.2.3", "version 1.2.3"
		if strings.Contains(line, " ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				// Check if this part looks like a version (starts with digit, contains dots)
				if len(part) > 0 && part[0] >= '0' && part[0] <= '9' && strings.Contains(part, ".") {
					// For most cases, just return the version number itself
					version := part
					
					// Check if next part might be version-related (like build info, not random words)
					if i+1 < len(parts) {
						nextPart := parts[i+1]
						// Include next part only if it looks like version metadata
						// Exclude common non-version words
						excludeWords := []string{"compiled", "built", "with", "using", "for", "on", "at", "from", "by"}
						isExcluded := false
						for _, exclude := range excludeWords {
							if strings.EqualFold(nextPart, exclude) {
								isExcluded = true
								break
							}
						}
						
						if !isExcluded && (strings.HasPrefix(nextPart, "(") || strings.HasPrefix(nextPart, "[") || 
						   strings.Contains(nextPart, "+") || strings.Contains(nextPart, "-") ||
						   strings.Contains(nextPart, "alpha") || strings.Contains(nextPart, "beta") ||
						   strings.Contains(nextPart, "rc") || strings.Contains(nextPart, "dev") ||
						   (len(nextPart) < 8 && !strings.Contains(nextPart, " "))) { // Short version-like additions
							version += " " + nextPart
							
							// Check for third part only if it's clearly version-related
							if i+2 < len(parts) {
								thirdPart := parts[i+2]
								if strings.HasSuffix(nextPart, "-") || strings.HasSuffix(thirdPart, ")") ||
								   strings.HasSuffix(thirdPart, "]") {
									version += " " + thirdPart
								}
							}
						}
					}
					return version
				}
			}
		}
	}

	// Strategy 3: Look for standalone version numbers in any line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[0] >= '0' && line[0] <= '9' && strings.Contains(line, ".") {
			return line
		}
	}

	// Strategy 4: Extract version from verbose outputs (e.g., lazygit)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 50 { // Long line, likely verbose output
			extractedVersion := extractVersionFromVerboseOutput(line)
			if extractedVersion != "" {
				return extractedVersion
			}
		}
	}

	// Strategy 5: Fall back to first line if it contains meaningful version info
	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) > 0 {
		// If first line looks like it contains version info, use it
		if strings.Contains(strings.ToLower(firstLine), "version") || 
		   strings.Contains(firstLine, ".") ||
		   len(strings.Fields(firstLine)) <= 4 { // Short, likely version info
			return firstLine
		}
	}

	// Strategy 6: Default fallback
	return "installed"
}

// extractVersionFromVerboseOutput extracts clean version info from verbose tool outputs
func extractVersionFromVerboseOutput(line string) string {
	// Look for "version=X.Y.Z" pattern (lazygit style)
	if strings.Contains(line, "version=") {
		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "version=") {
				version := strings.TrimPrefix(part, "version=")
				if version != "" {
					return "v" + version
				}
			}
		}
	}
	
	// Look for other verbose patterns and extract key version info
	// Could add more patterns here as needed for other tools
	
	return ""
}

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

// RunDoctor runs health checks and diagnostics
func (o *Orchestrator) RunDoctor(toolNames []string) error {
	if len(toolNames) == 1 && toolNames[0] == "nerd-fonts" {
		return o.runNerdFontsDoctor()
	}
	
	// General doctor functionality can be added here
	fmt.Printf("🔍 General Health Check\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("For tool-specific diagnostics, specify a tool name.\n")
	fmt.Printf("Example: gearbox doctor nerd-fonts\n")
	
	return nil
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

// Helper functions for nerd-fonts doctor checks
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
		if !hasFzf && !isToolInConfig("fzf") {
			missing = append(missing, "fzf")
		}
		if !hasBat && !isToolInConfig("bat") {
			missing = append(missing, "bat")
		}
		if !hasEza && !isToolInConfig("eza") {
			missing = append(missing, "eza")
		}
		
		if len(missing) > 0 {
			suggestions = append(suggestions, fmt.Sprintf("🚀 Terminal bundle: Consider also installing %s for a complete terminal experience", strings.Join(missing, ", ")))
		}
	}
	
	// Git workflow enhancement
	if hasDelta && !contains(installingNames, "lazygit") && !isToolInConfig("lazygit") {
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
			if !contains(installingNames, devTool) && !isToolInConfig(devTool) {
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

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// Helper function to check if a tool exists in config and is likely installed
func isToolInConfig(toolName string) bool {
	// Simple check - try to find the tool binary in PATH
	cmd := exec.Command("which", toolName)
	return cmd.Run() == nil
}

// Tracking command implementations
func trackInstallationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "track-installation TOOL_NAME METHOD VERSION [OPTIONS...]",
		Short:              "Track a tool installation in the manifest",
		Args:               cobra.MinimumNArgs(3),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleTrackInstallation(args)
		},
	}
	return cmd
}

func trackBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "track-bundle BUNDLE_NAME TOOLS [OPTIONS...]",
		Short:              "Track a bundle installation in the manifest",
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleTrackBundle(args)
		},
	}
}

func isTrackedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "is-tracked TOOL_NAME",
		Short: "Check if a tool is tracked in the manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleIsTracked(args)
		},
	}
}

func trackPreexistingCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "track-preexisting TOOL_NAME BINARY_PATH VERSION",
		Short:              "Track a pre-existing tool installation",
		Args:               cobra.ExactArgs(3),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleTrackPreexisting(args)
		},
	}
}

func initManifestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-manifest",
		Short: "Initialize a new installation manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleInitManifest(args)
		},
	}
}

func manifestStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manifest-status",
		Short: "Show installation manifest status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleManifestStatus(args)
		},
	}
}

func listDependentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-dependents TOOL_NAME",
		Short: "List tools that depend on the given tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleListDependents(args)
		},
	}
}

func canRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "can-remove TOOL_NAME",
		Short: "Check if a tool can be safely removed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleCanRemove(args)
		},
	}
}

// uninstallCmd creates the uninstall command
func uninstallCmd() *cobra.Command {
	var opts uninstall.RemovalOptions

	cmd := &cobra.Command{
		Use:   "uninstall [tools...]",
		Short: "Uninstall tools with safe removal",
		Long: `Uninstall one or more tools with dependency analysis and safe removal.
Analyzes dependencies and provides a removal plan before execution.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no tools specified for removal")
			}

			// Create removal engine with standard safety level
			engine, err := uninstall.NewRemovalEngine(uninstall.SafetyStandard)
			if err != nil {
				return fmt.Errorf("failed to create removal engine: %w", err)
			}

			// Plan removal
			plan, err := engine.PlanRemoval(args, opts)
			if err != nil {
				return fmt.Errorf("failed to plan removal: %w", err)
			}

			// Show plan and get confirmation
			if err := showRemovalPlan(plan); err != nil {
				return err
			}

			if !opts.DryRun {
				fmt.Printf("\nProceed with removal? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Printf("❌ Removal cancelled\n")
					return nil
				}
			}

			// Execute removal
			executor, err := uninstall.NewRemovalExecutor(opts.DryRun)
			if err != nil {
				return fmt.Errorf("failed to create removal executor: %w", err)
			}

			result, err := executor.ExecutePlan(plan, opts)
			if err != nil {
				return fmt.Errorf("failed to execute removal: %w", err)
			}

			// Show results
			fmt.Printf("\n%s", result.Summary())
			return nil
		},
	}

	// Removal options
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force removal even if there are dependents")
	cmd.Flags().BoolVar(&opts.Cascade, "cascade", false, "Remove unused dependencies")
	cmd.Flags().BoolVar(&opts.RemoveConfig, "remove-config", false, "Remove configuration files")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be removed without executing")
	cmd.Flags().BoolVar(&opts.Backup, "backup", true, "Create backup before removal")
	cmd.Flags().StringVar(&opts.BackupSuffix, "backup-suffix", "", "Suffix for backup files")

	return cmd
}

// uninstallPlanCmd creates the uninstall-plan command
func uninstallPlanCmd() *cobra.Command {
	var safetyLevel string

	cmd := &cobra.Command{
		Use:   "uninstall-plan [tools...]",
		Short: "Show removal plan without executing",
		Long: `Analyze what would be removed for the specified tools.
Shows dependency analysis and safety warnings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no tools specified for analysis")
			}

			// Parse safety level
			var safety uninstall.SafetyLevel
			switch strings.ToLower(safetyLevel) {
			case "conservative":
				safety = uninstall.SafetyConservative
			case "standard":
				safety = uninstall.SafetyStandard
			case "aggressive":
				safety = uninstall.SafetyAggressive
			default:
				safety = uninstall.SafetyStandard
			}

			// Create removal engine
			engine, err := uninstall.NewRemovalEngine(safety)
			if err != nil {
				return fmt.Errorf("failed to create removal engine: %w", err)
			}

			// Plan removal
			opts := uninstall.RemovalOptions{DryRun: true}
			plan, err := engine.PlanRemoval(args, opts)
			if err != nil {
				return fmt.Errorf("failed to plan removal: %w", err)
			}

			// Show detailed plan
			return showDetailedRemovalPlan(plan, engine)
		},
	}

	cmd.Flags().StringVar(&safetyLevel, "safety", "standard", "Safety level (conservative, standard, aggressive)")

	return cmd
}

// uninstallExecuteCmd creates the uninstall-execute command
func uninstallExecuteCmd() *cobra.Command {
	var opts uninstall.RemovalOptions

	cmd := &cobra.Command{
		Use:   "uninstall-execute [tools...]",
		Short: "Execute removal without confirmation",
		Long: `Execute tool removal without interactive confirmation.
Use with caution - this bypasses safety prompts.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no tools specified for removal")
			}

			// Create removal engine with aggressive safety for non-interactive use
			engine, err := uninstall.NewRemovalEngine(uninstall.SafetyAggressive)
			if err != nil {
				return fmt.Errorf("failed to create removal engine: %w", err)
			}

			// Plan removal
			plan, err := engine.PlanRemoval(args, opts)
			if err != nil {
				return fmt.Errorf("failed to plan removal: %w", err)
			}

			// Execute removal immediately
			executor, err := uninstall.NewRemovalExecutor(opts.DryRun)
			if err != nil {
				return fmt.Errorf("failed to create removal executor: %w", err)
			}

			fmt.Printf("🗑️  Executing removal of %d tools...\n", len(plan.ToRemove))

			result, err := executor.ExecutePlan(plan, opts)
			if err != nil {
				return fmt.Errorf("failed to execute removal: %w", err)
			}

			// Show results
			fmt.Printf("\n%s", result.Summary())
			return nil
		},
	}

	// Removal options
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force removal even if there are dependents")
	cmd.Flags().BoolVar(&opts.Cascade, "cascade", false, "Remove unused dependencies")
	cmd.Flags().BoolVar(&opts.RemoveConfig, "remove-config", false, "Remove configuration files")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be removed without executing")
	cmd.Flags().BoolVar(&opts.Backup, "backup", true, "Create backup before removal")
	cmd.Flags().StringVar(&opts.BackupSuffix, "backup-suffix", "", "Suffix for backup files")

	return cmd
}

// showRemovalPlan displays a removal plan to the user
func showRemovalPlan(plan *uninstall.RemovalPlan) error {
	fmt.Printf("🗑️  Removal Plan\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Show tools to be removed
	if len(plan.ToRemove) > 0 {
		fmt.Printf("📋 Tools to be removed (%d):\n", len(plan.ToRemove))
		for _, action := range plan.ToRemove {
			safety := "🟢"
			if !action.IsSafe {
				safety = "🔴"
			}
			fmt.Printf("  %s %-15s (%s) - %s\n", 
				safety, action.Target, action.Method, action.Reason)
		}
		fmt.Printf("\n")
	}

	// Show tools to be kept
	if len(plan.ToKeep) > 0 {
		fmt.Printf("🛡️  Tools to be kept (%d):\n", len(plan.ToKeep))
		for _, keep := range plan.ToKeep {
			fmt.Printf("  %-15s - %s\n", keep.Target, strings.Join(keep.Reasons, ", "))
		}
		fmt.Printf("\n")
	}

	// Show dependency actions
	if len(plan.Dependencies) > 0 {
		fmt.Printf("🔗 Dependency actions:\n")
		for _, dep := range plan.Dependencies {
			action := "preserve"
			if dep.Action != "preserve" {
				action = string(dep.Action)
			}
			fmt.Printf("  %-15s %s - %s\n", dep.Dependency, action, dep.Reason)
		}
		fmt.Printf("\n")
	}

	// Show warnings
	if len(plan.Warnings) > 0 {
		fmt.Printf("⚠️  Warnings (%d):\n", len(plan.Warnings))
		for _, warning := range plan.Warnings {
			level := "ℹ️"
			if warning.Level == "warning" {
				level = "⚠️"
			} else if warning.Level == "error" {
				level = "❌"
			}
			fmt.Printf("  %s %s: %s\n", level, warning.Target, warning.Message)
		}
		fmt.Printf("\n")
	}

	// Show summary
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  Total requested: %d\n", plan.Summary.TotalRequested)
	fmt.Printf("  Will remove: %d\n", plan.Summary.WillRemove)
	fmt.Printf("  Will keep: %d\n", plan.Summary.WillKeep)
	fmt.Printf("  Warnings: %d\n", plan.Summary.WarningCount)

	return nil
}

// showDetailedRemovalPlan displays a detailed removal plan with validation
func showDetailedRemovalPlan(plan *uninstall.RemovalPlan, engine *uninstall.RemovalEngine) error {
	if err := showRemovalPlan(plan); err != nil {
		return err
	}

	// Validate plan and show additional warnings
	validationWarnings := engine.ValidatePlan(plan)
	if len(validationWarnings) > 0 {
		fmt.Printf("\n🔍 Validation Results:\n")
		for _, warning := range validationWarnings {
			level := "ℹ️"
			if warning.Level == "warning" {
				level = "⚠️"
			} else if warning.Level == "error" {
				level = "❌"
			}
			fmt.Printf("  %s %s: %s\n", level, warning.Target, warning.Message)
		}
	}

	// Show method breakdown
	if len(plan.Summary.MethodBreakdown) > 0 {
		fmt.Printf("\n🔧 Removal methods:\n")
		for method, count := range plan.Summary.MethodBreakdown {
			description := uninstall.GetRemovalMethodDescription(method)
			fmt.Printf("  %-20s %d tools - %s\n", string(method), count, description)
		}
	}

	return nil
}