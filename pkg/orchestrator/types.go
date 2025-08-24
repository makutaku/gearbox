package orchestrator

import (
	"fmt"
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
	configMgr     *ConfigManager
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

// ConfigManager handles configuration management without global state
type ConfigManager struct {
	config        Config
	bundleConfigs map[string][]string
	mu            sync.RWMutex
}

// NewConfigManager creates a new configuration manager that handles
// loading and thread-safe access to tool and bundle configurations.
// It replaces the use of global variables for better maintainability.
//
// Parameters:
//   - configPath: Path to the tools.json configuration file
//
// Returns:
//   - *ConfigManager: Configured manager instance
//   - error: Error if configuration loading fails
func NewConfigManager(configPath string) (*ConfigManager, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	
	return &ConfigManager{
		config:        config,
		bundleConfigs: make(map[string][]string),
	}, nil
}

// GetConfig returns the current configuration in a thread-safe manner.
// This method uses a read lock to allow concurrent access while preventing
// data races with configuration updates.
//
// Returns:
//   - Config: Copy of the current configuration
func (cm *ConfigManager) GetConfig() Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// SetBundleConfigs sets the bundle configurations (thread-safe)
func (cm *ConfigManager) SetBundleConfigs(configs map[string][]string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.bundleConfigs = configs
}

// GetBundleConfigs returns the bundle configurations (thread-safe)
func (cm *ConfigManager) GetBundleConfigs() map[string][]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	result := make(map[string][]string)
	for k, v := range cm.bundleConfigs {
		result[k] = v
	}
	return result
}