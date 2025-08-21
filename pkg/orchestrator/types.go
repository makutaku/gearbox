package orchestrator

import (
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
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
	Tool     ToolConfig
	Success  bool
	Error    error
	Duration time.Duration
	Output   string
}

// Orchestrator handles tool installation orchestration
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
	// Global configuration loaded at startup
	globalConfig Config
	// Available bundle configurations
	bundleConfigs map[string][]string
)