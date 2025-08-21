package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

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

// findTool finds a tool by name in the configuration
func (o *Orchestrator) findTool(name string) (ToolConfig, bool) {
	for _, tool := range o.config.Tools {
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