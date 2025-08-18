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

func Main() {
	var rootCmd = &cobra.Command{
		Use:   "orchestrator",
		Short: "Gearbox installation orchestrator",
		Long:  "Advanced orchestration engine for managing gearbox tool installations with dependency resolution and parallel execution",
	}

	// Add commands
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(verifyCmd())

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

	return cmd
}

// listCmd creates the list command
func listCmd() *cobra.Command {
	var category string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			return orchestrator.ListTools(category, verbose)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
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

// NewOrchestrator creates a new orchestrator instance
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
	fmt.Printf("🔧 Gearbox Orchestrator - Installing %d tools\n\n", len(toolNames))

	// Validate tool names
	var validTools []ToolConfig
	for _, name := range toolNames {
		tool, found := o.findTool(name)
		if !found {
			return fmt.Errorf("tool not found: %s", name)
		}
		validTools = append(validTools, tool)
	}

	// Resolve dependencies and determine installation order
	installOrder, err := o.resolveDependencies(validTools)
	if err != nil {
		return fmt.Errorf("dependency resolution failed: %w", err)
	}

	if o.options.DryRun {
		return o.showDryRun(installOrder)
	}

	// Show installation plan
	o.showInstallationPlan(installOrder)

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
func (o *Orchestrator) showDryRun(tools []ToolConfig) error {
	fmt.Printf("🔍 Dry Run - Installation Plan\n\n")
	fmt.Printf("Build Type: %s\n", o.options.BuildType)
	fmt.Printf("Max Parallel Jobs: %d\n", o.options.MaxParallelJobs)
	fmt.Printf("Skip Common Deps: %v\n", o.options.SkipCommonDeps)
	fmt.Printf("Run Tests: %v\n", o.options.RunTests)
	fmt.Printf("Shell Integration: %v\n", !o.options.NoShell)
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
		if isToolInstalled(tool) {
			installed++
			version := getToolVersion(tool)
			fmt.Printf("✅ %-15s %s\n", tool.Name, version)
		} else {
			notInstalled++
			fmt.Printf("❌ %-15s Not installed\n", tool.Name)
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
	_, err := exec.LookPath(tool.BinaryName)
	return err == nil
}

// verifyTool verifies a tool installation by running its test command
func verifyTool(tool ToolConfig) bool {
	if !isToolInstalled(tool) {
		return false
	}

	if tool.TestCommand == "" {
		return true // If no test command, just check if binary exists
	}

	// Parse and execute test command
	parts := strings.Fields(tool.TestCommand)
	if len(parts) == 0 {
		return true
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	err := cmd.Run()
	return err == nil
}

// getToolVersion gets the version of an installed tool
func getToolVersion(tool ToolConfig) string {
	if tool.TestCommand == "" {
		return "installed"
	}

	parts := strings.Fields(tool.TestCommand)
	if len(parts) == 0 {
		return "installed"
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "installed"
	}

	// Return first line of output, trimmed
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return "installed"
}