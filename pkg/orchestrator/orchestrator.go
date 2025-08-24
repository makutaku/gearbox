package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	
	"gearbox/pkg/manifest"
)

// NewOrchestrator creates a new orchestrator instance with the given options.
// It auto-detects repository and configuration paths, loads configuration,
// and initializes the orchestrator with proper settings for parallel execution.
// OrchestratorBuilder provides a builder pattern for creating orchestrators
type OrchestratorBuilder struct {
	options      InstallationOptions
	repoDir      string
	configPath   string
	configMgr    *ConfigManager
	bundleConfig *BundleConfiguration
	packageMgr   *PackageManager
}

// NewOrchestratorBuilder creates a new orchestrator builder using the builder pattern.
// This replaces the monolithic NewOrchestrator function with a more maintainable
// approach that separates concerns and makes testing easier.
//
// Parameters:
//   - options: Installation configuration options
//
// Returns:
//   - *OrchestratorBuilder: Builder instance for chaining configuration methods
func NewOrchestratorBuilder(options InstallationOptions) *OrchestratorBuilder {
	return &OrchestratorBuilder{
		options: options,
	}
}

// WithRepoDir sets the repository directory
func (b *OrchestratorBuilder) WithRepoDir(dir string) *OrchestratorBuilder {
	b.repoDir = dir
	return b
}

// WithConfigPath sets the configuration file path
func (b *OrchestratorBuilder) WithConfigPath(path string) *OrchestratorBuilder {
	b.configPath = path
	return b
}

// autoDetectPaths automatically detects repository and config paths
func (b *OrchestratorBuilder) autoDetectPaths() error {
	// Auto-detect repository directory if not provided
	if b.repoDir == "" {
		if wd, err := os.Getwd(); err == nil {
			b.repoDir = wd
		} else {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Auto-detect config path if not provided
	if b.configPath == "" {
		b.configPath = filepath.Join(b.repoDir, "config", "tools.json")
	}

	return nil
}

// loadConfiguration loads and validates the configuration
func (b *OrchestratorBuilder) loadConfiguration() error {
	configMgr, err := NewConfigManager(b.configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	b.configMgr = configMgr
	
	// Set default build type from config if not specified
	if b.options.BuildType == "" {
		b.options.BuildType = configMgr.GetConfig().DefaultBuildType
	}

	return nil
}

// setupParallelism configures parallel job limits
func (b *OrchestratorBuilder) setupParallelism() {
	// Auto-detect max parallel jobs with intelligent resource calculation
	if b.options.MaxParallelJobs == 0 {
		cpuCount := runtime.NumCPU()
		
		// Get available memory (basic implementation)
		memoryLimitedJobs := b.calculateMemoryBasedJobs()
		
		// Take the minimum of CPU-based and memory-based limits
		b.options.MaxParallelJobs = min(cpuCount, memoryLimitedJobs)
		
		// Ensure we have at least 1 job but not more than 8 (reasonable upper bound)
		if b.options.MaxParallelJobs < 1 {
			b.options.MaxParallelJobs = 1
		} else if b.options.MaxParallelJobs > 8 {
			b.options.MaxParallelJobs = 8
		}
		
		if b.options.Verbose {
			fmt.Printf("Auto-detected parallel jobs: %d (CPU cores: %d, memory-limited: %d)\n", 
				b.options.MaxParallelJobs, cpuCount, memoryLimitedJobs)
		}
	}
}

// calculateMemoryBasedJobs estimates max parallel jobs based on available memory
func (b *OrchestratorBuilder) calculateMemoryBasedJobs() int {
	// Conservative estimates for memory usage per build type
	// These are rough estimates - could be made more sophisticated
	memoryPerJobMB := map[string]int{
		"minimal":  200,  // 200MB per minimal build
		"standard": 500,  // 500MB per standard build
		"maximum":  1000, // 1GB per maximum build
	}
	
	buildType := b.options.BuildType
	if buildType == "" {
		buildType = "standard"
	}
	
	memPerJob := memoryPerJobMB[buildType]
	if memPerJob == 0 {
		memPerJob = 500 // default
	}
	
	// Try to get system memory (Linux-specific basic implementation)
	availableMemoryMB := b.getAvailableMemoryMB()
	if availableMemoryMB <= 0 {
		// Fallback: assume 4GB available memory
		availableMemoryMB = 4096
	}
	
	// Reserve 1GB for system processes
	usableMemoryMB := availableMemoryMB - 1024
	if usableMemoryMB < memPerJob {
		return 1 // Can only run one job
	}
	
	return usableMemoryMB / memPerJob
}

// getAvailableMemoryMB gets available memory in MB (basic Linux implementation)
func (b *OrchestratorBuilder) getAvailableMemoryMB() int {
	// Read /proc/meminfo for available memory
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return -1 // Unknown
	}
	
	lines := strings.Split(string(content), "\n")
	var availableKB int
	
	for _, line := range lines {
		if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if kb, err := strconv.Atoi(fields[1]); err == nil {
					availableKB = kb
					break
				}
			}
		}
	}
	
	if availableKB == 0 {
		return -1 // Could not determine
	}
	
	return availableKB / 1024 // Convert KB to MB
}

// (min function moved to utils.go to avoid redeclaration)

// loadBundleConfig loads bundle configuration (optional)
func (b *OrchestratorBuilder) loadBundleConfig() error {
	// Use the real bundle loading implementation
	bundleConfigPath := filepath.Join(b.repoDir, "config", "bundles.json")
	
	// Read and parse the bundle configuration file directly
	file, err := os.Open(bundleConfigPath)
	if err != nil {
		// Bundles are optional, so return empty config if file doesn't exist
		if os.IsNotExist(err) {
			b.bundleConfig = &BundleConfiguration{
				SchemaVersion: "1.0",
				Bundles:       []BundleConfig{},
			}
			return nil
		}
		if b.options.Verbose {
			fmt.Printf("⚠️  Warning: Failed to open bundles.json: %v\n", err)
		}
		b.bundleConfig = &BundleConfiguration{
			SchemaVersion: "1.0",
			Bundles:       []BundleConfig{},
		}
		return nil
	}
	defer file.Close()

	var bundleConfig BundleConfiguration
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&bundleConfig); err != nil {
		if b.options.Verbose {
			fmt.Printf("⚠️  Warning: Failed to decode bundles.json: %v\n", err)
		}
		b.bundleConfig = &BundleConfiguration{
			SchemaVersion: "1.0",
			Bundles:       []BundleConfig{},
		}
		return nil
	}

	b.bundleConfig = &bundleConfig
	return nil
}

// detectPackageManager detects the system package manager
func (b *OrchestratorBuilder) detectPackageManager() error {
	packageMgr, err := detectPackageManager()
	if err != nil && b.options.Verbose {
		fmt.Printf("⚠️  System package manager not detected: %v\n", err)
	}
	b.packageMgr = packageMgr
	return nil
}

// Build creates the final orchestrator instance
func (b *OrchestratorBuilder) Build() (*Orchestrator, error) {
	// Execute build steps in order
	if err := b.autoDetectPaths(); err != nil {
		return nil, err
	}

	if err := b.loadConfiguration(); err != nil {
		return nil, err
	}

	b.setupParallelism()

	if err := b.loadBundleConfig(); err != nil {
		return nil, err
	}

	if err := b.detectPackageManager(); err != nil {
		return nil, err
	}

	// Create orchestrator instance
	orchestrator := &Orchestrator{
		configMgr:    b.configMgr,
		bundleConfig: b.bundleConfig,
		packageMgr:   b.packageMgr,
		options:      b.options,
		repoDir:      b.repoDir,
		scriptsDir:   filepath.Join(b.repoDir, "scripts"),
		results:      make([]InstallationResult, 0),
	}

	// Initialize memory pool for result objects
	orchestrator.resultPool = sync.Pool{
		New: func() interface{} {
			return &InstallationResult{}
		},
	}

	return orchestrator, nil
}





// findTool finds a tool by name in the configuration
func (o *Orchestrator) findTool(name string) (ToolConfig, bool) {
	config := o.configMgr.GetConfig()
	for _, tool := range config.Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolConfig{}, false
}

// ListTools lists available tools
func (o *Orchestrator) ListTools(category string, verbose bool) error {
	// Group tools by category
	toolsByCategory := make(map[string][]ToolConfig)
	config := o.configMgr.GetConfig()
	for _, tool := range config.Tools {
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
		if description, exists := config.Categories[cat]; exists {
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

// ShowStatus shows installation status of tools with different modes
func (o *Orchestrator) ShowStatus(toolNames []string, manifestOnly bool, unified bool) error {
	config := o.configMgr.GetConfig()
	var tools []ToolConfig
	
	if len(toolNames) == 0 {
		tools = config.Tools
	} else {
		for _, name := range toolNames {
			tool, found := o.findTool(name)
			if !found {
				return fmt.Errorf("tool not found: %s", name)
			}
			tools = append(tools, tool)
		}
	}

	// Load manifest data for manifest-only or unified modes
	var manifestData map[string]bool
	if manifestOnly || unified {
		manifestData = make(map[string]bool)
		
		// Load manifest data from ~/.gearbox/manifest.json
		manifestMgr := manifest.NewManager()
		if manifestContent, err := manifestMgr.Load(); err == nil && manifestContent.Installations != nil {
			for toolName := range manifestContent.Installations {
				manifestData[toolName] = true
			}
		}
	}
	
	// Show different headers based on mode
	if manifestOnly {
		fmt.Printf("📊 Tool Installation Status (Manifest Only)\n")
	} else if unified {
		fmt.Printf("📊 Tool Installation Status (Unified: Manifest + Live)\n")
	} else {
		fmt.Printf("📊 Tool Installation Status\n")
	}
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
			continue
		}
		
		// Different logic based on mode
		if manifestOnly {
			// Only show tools that are in manifest
			if _, inManifest := manifestData[tool.Name]; inManifest {
				installed++
				fmt.Printf("✅ %-15s manifest-tracked\n", tool.Name)
			} else {
				notInstalled++
				fmt.Printf("❌ %-15s Not in manifest\n", tool.Name)
			}
		} else if unified {
			// Show unified view with indicators
			liveInstalled := isToolInstalled(tool)
			inManifest := false
			if manifestData != nil {
				_, inManifest = manifestData[tool.Name]
			}
			
			var status, source string
			if liveInstalled && inManifest {
				status = "✅"
				source = "gearbox"
				installed++
			} else if liveInstalled && !inManifest {
				status = "✅"
				source = "system"
				installed++
			} else if !liveInstalled && inManifest {
				status = "⚠️"
				source = "missing"
				notInstalled++
			} else {
				status = "❌"
				source = "not installed"
				notInstalled++
			}
			
			version := ""
			if liveInstalled {
				version = getToolVersion(tool)
			}
			
			fmt.Printf("%s %-15s %-12s (%s)\n", status, tool.Name, version, source)
		} else {
			// Default live detection mode
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
	config := o.configMgr.GetConfig()
	var tools []ToolConfig
	
	if len(toolNames) == 0 {
		tools = config.Tools
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

// GetConfig returns the orchestrator's configuration
func (o *Orchestrator) GetConfig() *Config {
	config := o.configMgr.GetConfig()
	return &config
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

// Cleanup performs cleanup operations for the orchestrator instance.
// This method should be called when the orchestrator is no longer needed
// to free resources and prevent memory leaks.
//
// Thread-safe: Uses mutex to prevent concurrent access during cleanup.
//
// Returns:
//   - error: Combined error if any cleanup operations fail
func (o *Orchestrator) Cleanup() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	var cleanupErrors []error

	// Clean up progress bar
	if o.progressBar != nil {
		if err := o.progressBar.Close(); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to close progress bar: %w", err))
		}
		o.progressBar = nil
	}

	// Clear results to free memory
	for i := range o.results {
		// Clear individual result data
		o.results[i] = InstallationResult{}
	}
	o.results = nil

	// Reset result pool
	o.resultPool = sync.Pool{
		New: func() interface{} {
			return &InstallationResult{}
		},
	}

	return combineErrors(cleanupErrors)
}

// CleanupOnFailure performs cleanup when installation fails
func (o *Orchestrator) CleanupOnFailure(failedTool string, tempDirs []string) error {
	var cleanupErrors []error

	// Log cleanup action
	fmt.Printf("🧹 Cleaning up after failed installation of %s...\n", failedTool)

	// Clean up temporary directories
	for _, dir := range tempDirs {
		if err := os.RemoveAll(dir); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("failed to remove temp dir %s: %w", dir, err))
		} else {
			fmt.Printf("✓ Removed temporary directory: %s\n", dir)
		}
	}

	// Mark the failed tool in results
	o.mu.Lock()
	for i, result := range o.results {
		if result.Tool.Name == failedTool {
			o.results[i].Success = false
			o.results[i].Error = fmt.Errorf("installation failed and cleaned up")
			break
		}
	}
	o.mu.Unlock()

	return combineErrors(cleanupErrors)
}

// AddCleanupHandler adds a cleanup function to be called on failure
type CleanupHandler func() error

// CleanupContext manages cleanup handlers for an operation
type CleanupContext struct {
	handlers []CleanupHandler
	mu       sync.Mutex
}

// NewCleanupContext creates a new cleanup context
func NewCleanupContext() *CleanupContext {
	return &CleanupContext{
		handlers: make([]CleanupHandler, 0),
	}
}

// AddHandler adds a cleanup handler
func (ctx *CleanupContext) AddHandler(handler CleanupHandler) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.handlers = append(ctx.handlers, handler)
}

// RunCleanup runs all cleanup handlers
func (ctx *CleanupContext) RunCleanup() error {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	var errors []error
	// Run cleanup handlers in reverse order (LIFO)
	for i := len(ctx.handlers) - 1; i >= 0; i-- {
		if err := ctx.handlers[i](); err != nil {
			errors = append(errors, err)
		}
	}

	return combineErrors(errors)
}

// combineErrors combines multiple errors into a single error
func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	if len(errors) == 1 {
		return errors[0]
	}

	var messages []string
	for _, err := range errors {
		messages = append(messages, err.Error())
	}
	return fmt.Errorf("multiple cleanup errors: %s", strings.Join(messages, "; "))
}
