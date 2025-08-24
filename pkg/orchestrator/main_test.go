package orchestrator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test data
var testConfig = Config{
	SchemaVersion:    "1.0",
	DefaultBuildType: "standard",
	Tools: []ToolConfig{
		{
			Name:        "test-tool",
			Description: "Test tool for unit testing",
			Category:    "development",
			Repository:  "https://github.com/test/test-tool.git",
			BinaryName:  "test-tool",
			Language:    "rust",
			BuildTypes: map[string]string{
				"minimal":  "-m",
				"standard": "-s",
				"maximum":  "-r",
			},
			Dependencies:     []string{"rust", "build-essential"},
			MinVersion:       "",
			ShellIntegration: false,
			TestCommand:      "--version",
		},
		{
			Name:        "test-go-tool",
			Description: "Test Go tool",
			Category:    "development",
			Repository:  "https://github.com/test/go-tool.git",
			BinaryName:  "go-tool",
			Language:    "go",
			BuildTypes: map[string]string{
				"minimal":  "-m",
				"standard": "-s",
				"maximum":  "-r",
			},
			Dependencies:     []string{"go"},
			MinVersion:       "",
			ShellIntegration: false,
			TestCommand:      "version",
		},
	},
	Categories: map[string]string{
		"test": "Test tools for development",
	},
	Languages: map[string]LanguageConfig{
		"rust": {
			MinVersion: "1.88.0",
			BuildTool:  "cargo",
		},
		"go": {
			MinVersion: "1.23.4",
			BuildTool:  "go",
		},
	},
}

// setupTestConfig creates a temporary config file for testing
func setupTestConfig(t *testing.T) string {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	configPath := filepath.Join(configDir, "tools.json")
	
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	
	return configPath
}

// TestLoadConfig tests configuration loading
func TestLoadConfig(t *testing.T) {
	configPath := setupTestConfig(t)
	
	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify basic config fields
	if config.SchemaVersion != testConfig.SchemaVersion {
		t.Errorf("Expected schema version %s, got %s", testConfig.SchemaVersion, config.SchemaVersion)
	}
	
	if config.DefaultBuildType != testConfig.DefaultBuildType {
		t.Errorf("Expected default build type %s, got %s", testConfig.DefaultBuildType, config.DefaultBuildType)
	}
	
	if len(config.Tools) != len(testConfig.Tools) {
		t.Errorf("Expected %d tools, got %d", len(testConfig.Tools), len(config.Tools))
	}
	
	// Test with non-existent file
	_, err = loadConfig("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error when loading non-existent config")
	}
}

// TestNewOrchestrator tests orchestrator initialization
func TestNewOrchestrator(t *testing.T) {
	testConfigPath := setupTestConfig(t)
	tempDir := filepath.Dir(filepath.Dir(testConfigPath)) // Go up one level from config/tools.json
	
	// Set global variables for test
	repoDir = tempDir
	configPath = testConfigPath
	
	options := InstallationOptions{
		BuildType:       "standard",
		MaxParallelJobs: 2,
		SkipCommonDeps:  false,
		RunTests:        false,
		Force:           false,
		Verbose:         false,
		DryRun:          false,
	}
	
	orchestrator, err := NewOrchestrator(options)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	// Verify orchestrator fields
	config := orchestrator.configMgr.GetConfig()
	if config.SchemaVersion != testConfig.SchemaVersion {
		t.Errorf("Expected schema version %s, got %s", testConfig.SchemaVersion, config.SchemaVersion)
	}
	
	if orchestrator.options.BuildType != "standard" {
		t.Errorf("Expected build type 'standard', got %s", orchestrator.options.BuildType)
	}
	
	if orchestrator.options.MaxParallelJobs != 2 {
		t.Errorf("Expected max parallel jobs 2, got %d", orchestrator.options.MaxParallelJobs)
	}
}

// TestFindTool tests tool lookup functionality
func TestFindTool(t *testing.T) {
	testConfigPath := setupTestConfig(t)
	tempDir := filepath.Dir(filepath.Dir(testConfigPath))
	
	repoDir = tempDir
	configPath = testConfigPath
	
	orchestrator, err := NewOrchestrator(InstallationOptions{})
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	// Test finding existing tool
	tool, found := orchestrator.findTool("test-tool")
	if !found {
		t.Error("Expected to find test-tool")
	}
	
	if tool.Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got %s", tool.Name)
	}
	
	if tool.Language != "rust" {
		t.Errorf("Expected language 'rust', got %s", tool.Language)
	}
	
	// Test finding non-existent tool
	_, found = orchestrator.findTool("nonexistent-tool")
	if found {
		t.Error("Expected not to find nonexistent-tool")
	}
}

// TestResolveDependencies tests dependency resolution
func TestResolveDependencies(t *testing.T) {
	testConfigPath := setupTestConfig(t)
	tempDir := filepath.Dir(filepath.Dir(testConfigPath))
	
	repoDir = tempDir
	configPath = testConfigPath
	
	orchestrator, err := NewOrchestrator(InstallationOptions{})
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	config := orchestrator.configMgr.GetConfig()
	
	// Test dependency resolution with multiple tools
	tools := []ToolConfig{
		config.Tools[0], // test-tool (rust)
		config.Tools[1], // test-go-tool (go)
	}
	
	resolvedOrder, err := orchestrator.resolveDependencies(tools)
	if err != nil {
		t.Fatalf("Failed to resolve dependencies: %v", err)
	}
	
	if len(resolvedOrder) != 2 {
		t.Errorf("Expected 2 tools in resolved order, got %d", len(resolvedOrder))
	}
	
	// Go tools should come before Rust tools in the resolved order
	if resolvedOrder[0].Language != "go" {
		t.Errorf("Expected Go tool first, got %s", resolvedOrder[0].Language)
	}
	
	if resolvedOrder[1].Language != "rust" {
		t.Errorf("Expected Rust tool second, got %s", resolvedOrder[1].Language)
	}
}

// TestInstallationOptions tests option validation and defaults
func TestInstallationOptions(t *testing.T) {
	testConfigPath := setupTestConfig(t)
	tempDir := filepath.Dir(filepath.Dir(testConfigPath))
	
	repoDir = tempDir
	configPath = testConfigPath
	
	// Test with empty options (should use defaults)
	orchestrator, err := NewOrchestrator(InstallationOptions{})
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	// Should use default build type from config
	if orchestrator.options.BuildType != testConfig.DefaultBuildType {
		t.Errorf("Expected default build type %s, got %s", testConfig.DefaultBuildType, orchestrator.options.BuildType)
	}
	
	// Should auto-detect max parallel jobs
	if orchestrator.options.MaxParallelJobs <= 0 {
		t.Error("Expected positive max parallel jobs")
	}
	
	// Should not exceed reasonable limit
	if orchestrator.options.MaxParallelJobs > 4 {
		t.Errorf("Expected max parallel jobs <= 4, got %d", orchestrator.options.MaxParallelJobs)
	}
}

// TestInstallationResult tests result structures
func TestInstallationResult(t *testing.T) {
	tool := testConfig.Tools[0]
	
	// Test successful result
	successResult := InstallationResult{
		Tool:     tool,
		Success:  true,
		Error:    nil,
		Duration: time.Second * 5,
		Output:   "Installation completed successfully",
	}
	
	if !successResult.Success {
		t.Error("Expected successful result")
	}
	
	if successResult.Error != nil {
		t.Error("Expected no error in successful result")
	}
	
	if successResult.Duration != time.Second*5 {
		t.Errorf("Expected duration 5s, got %v", successResult.Duration)
	}
	
	// Test failed result
	failResult := InstallationResult{
		Tool:     tool,
		Success:  false,
		Error:    &InstallationError{Tool: "test-tool", Message: "Build failed"},
		Duration: time.Second * 2,
		Output:   "Build process failed",
	}
	
	if failResult.Success {
		t.Error("Expected failed result")
	}
	
	if failResult.Error == nil {
		t.Error("Expected error in failed result")
	}
}

// InstallationError represents an installation error for testing
type InstallationError struct {
	Tool    string
	Message string
}

func (e *InstallationError) Error() string {
	return e.Tool + ": " + e.Message
}

// TestTemplateData tests template data preparation
func TestTemplateData(t *testing.T) {
	testConfigPath := setupTestConfig(t)
	tempDir := filepath.Dir(filepath.Dir(testConfigPath))
	
	repoDir = tempDir
	configPath = testConfigPath
	
	orchestrator, err := NewOrchestrator(InstallationOptions{})
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	config := orchestrator.configMgr.GetConfig()
	tool := config.Tools[0] // test-tool
	
	// Note: prepareTemplateData is not exported in the actual code,
	// so this test would need the function to be exported or moved to a testable form
	// This is a placeholder for the structure we would test
	
	// Verify tool has expected properties
	if tool.Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got %s", tool.Name)
	}
	
	if len(tool.BuildTypes) != 3 {
		t.Errorf("Expected 3 build types, got %d", len(tool.BuildTypes))
	}
	
	expectedBuildTypes := []string{"minimal", "standard", "maximum"}
	for _, buildType := range expectedBuildTypes {
		if _, exists := tool.BuildTypes[buildType]; !exists {
			t.Errorf("Expected build type %s to exist", buildType)
		}
	}
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	// Test with invalid JSON
	tempDir := t.TempDir()
	invalidConfigPath := filepath.Join(tempDir, "invalid.json")
	
	invalidJSON := `{
		"schema_version": "1.0"
		"invalid": "json"
	}`
	
	if err := os.WriteFile(invalidConfigPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	
	_, err := loadConfig(invalidConfigPath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
	
	// Test with empty file
	emptyConfigPath := filepath.Join(tempDir, "empty.json")
	if err := os.WriteFile(emptyConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write empty config: %v", err)
	}
	
	_, err = loadConfig(emptyConfigPath)
	if err == nil {
		t.Error("Expected error when loading empty config")
	}
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	// Test with missing config file
	_, err := loadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("Expected error when config file doesn't exist")
	}
	
	// Test orchestrator creation with invalid repo dir
	repoDir = "/nonexistent/repo"
	configPath = ""
	
	_, err = NewOrchestrator(InstallationOptions{})
	if err == nil {
		t.Error("Expected error when repo directory is invalid")
	}
}

// BenchmarkLoadConfig benchmarks configuration loading
func BenchmarkLoadConfig(b *testing.B) {
	// Create a temporary config file
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "tools.json")
	
	configData, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		b.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		b.Fatalf("Failed to write test config: %v", err)
	}
	
	// Benchmark the loading
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loadConfig(configPath)
		if err != nil {
			b.Fatalf("Failed to load config: %v", err)
		}
	}
}

// BenchmarkResolveDependencies benchmarks dependency resolution
func BenchmarkResolveDependencies(b *testing.B) {
	configPath := setupTestConfig(&testing.T{})
	tempDir := filepath.Dir(configPath)
	
	repoDir = tempDir
	configPath = configPath
	
	orchestrator, err := NewOrchestrator(InstallationOptions{})
	if err != nil {
		b.Fatalf("Failed to create orchestrator: %v", err)
	}
	
	config := orchestrator.configMgr.GetConfig()
	tools := config.Tools
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := orchestrator.resolveDependencies(tools)
		if err != nil {
			b.Fatalf("Failed to resolve dependencies: %v", err)
		}
	}
}