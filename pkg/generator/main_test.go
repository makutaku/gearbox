package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test data for script generator
var testGeneratorConfig = Config{
	SchemaVersion:    "1.0",
	DefaultBuildType: "standard",
	Tools: []ToolConfig{
		{
			Name:        "test-rust-tool",
			Description: "Test Rust tool for generation",
			Category:    "development",
			Repository:  "https://github.com/test/rust-tool.git",
			BinaryName:  "rust-tool",
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
			Description: "Test Go tool for generation",
			Category:    "development",
			Repository:  "https://github.com/test/go-tool.git",
			BinaryName:  "go-tool",
			Language:    "go",
			BuildTypes: map[string]string{
				"minimal":  "-m",
				"standard": "-s",
				"maximum":  "-o",
			},
			Dependencies:     []string{"go"},
			MinVersion:       "",
			ShellIntegration: true,
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

// setupTestGeneratorEnv creates test environment for script generator
func setupTestGeneratorEnv(t *testing.T) (string, string, string) {
	tempDir := t.TempDir()
	
	// Create config file
	configPath := filepath.Join(tempDir, "config", "tools.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}
	
	configData, err := json.MarshalIndent(testGeneratorConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	
	// Create templates directory with basic templates
	templatesDir := filepath.Join(tempDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}
	
	// Create basic rust template
	rustTemplate := `#!/bin/bash

# {{.Tool.Name}} Installation Script
# {{.Tool.Description}}

set -e  # Exit on any error

BUILD_TYPE="{{if .Tool.BuildTypes.standard}}standard{{else}}{{index .BuildTypes 0}}{{end}}"

# Logging functions
log() {
    echo "[INFO] $1"
}

error() {
    echo "[ERROR] $1" >&2
    exit 1
}

show_help() {
    echo "{{.Tool.Description}}"
    echo "Usage: $0 [OPTIONS]"
    echo "Build Types:"{{range $buildType, $flag := .Tool.BuildTypes}}
    echo "  {{$flag|printf "%-20s"}} {{$buildType}} build"{{end}}
}

# Configuration
{{.Tool.Name|upper}}_DIR="{{.RepoName}}"
{{.Tool.Name|upper}}_REPO="{{.Tool.Repository}}"

log "Installing {{.Tool.Name}}..."
success "{{.Tool.Name}} installation completed"
`
	
	rustTemplatePath := filepath.Join(templatesDir, "rust.sh.tmpl")
	if err := os.WriteFile(rustTemplatePath, []byte(rustTemplate), 0644); err != nil {
		t.Fatalf("Failed to write rust template: %v", err)
	}
	
	// Create basic go template
	goTemplate := `#!/bin/bash

# {{.Tool.Name}} Installation Script for Go
# {{.Tool.Description}}

set -e  # Exit on any error

BUILD_TYPE="{{if .Tool.BuildTypes.standard}}standard{{else}}{{index .BuildTypes 0}}{{end}}"

# Logging functions
log() {
    echo "[INFO] $1"
}

error() {
    echo "[ERROR] $1" >&2
    exit 1
}

show_help() {
    echo "{{.Tool.Description}}"
    echo "Usage: $0 [OPTIONS]"
}

log "Installing Go tool {{.Tool.Name}}..."
{{if .HasShell}}
# Shell integration enabled
{{end}}
success "{{.Tool.Name}} installation completed"
`
	
	goTemplatePath := filepath.Join(templatesDir, "go.sh.tmpl")
	if err := os.WriteFile(goTemplatePath, []byte(goTemplate), 0644); err != nil {
		t.Fatalf("Failed to write go template: %v", err)
	}
	
	// Create output directory
	outputDir := filepath.Join(tempDir, "scripts")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	
	return tempDir, configPath, templatesDir
}

// TestLoadConfig tests configuration loading for script generator
func TestGeneratorLoadConfig(t *testing.T) {
	_, configPath, _ := setupTestGeneratorEnv(t)
	
	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if len(config.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(config.Tools))
	}
	
	if config.Tools[0].Language != "rust" {
		t.Errorf("Expected first tool to be rust, got %s", config.Tools[0].Language)
	}
	
	if config.Tools[1].Language != "go" {
		t.Errorf("Expected second tool to be go, got %s", config.Tools[1].Language)
	}
}

// TestNewGenerator tests generator initialization
func TestNewGenerator(t *testing.T) {
	repoDir, configPath, templatesDir := setupTestGeneratorEnv(t)
	
	options := GeneratorOptions{
		ConfigPath:  configPath,
		OutputDir:   filepath.Join(repoDir, "scripts"),
		TemplateDir: templatesDir,
		Force:       false,
		Validate:    false,
		DryRun:      false,
	}
	
	// Set global variables
	configPath = options.ConfigPath
	outputDir = options.OutputDir
	templateDir = options.TemplateDir
	repoDir = repoDir
	
	generator, err := NewGenerator(options)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	if len(generator.config.Tools) != 2 {
		t.Errorf("Expected 2 tools in generator config, got %d", len(generator.config.Tools))
	}
	
	if len(generator.templates) < 2 {
		t.Errorf("Expected at least 2 templates loaded, got %d", len(generator.templates))
	}
	
	// Check that rust and go templates are loaded
	if _, exists := generator.templates["rust"]; !exists {
		t.Error("Expected rust template to be loaded")
	}
	
	if _, exists := generator.templates["go"]; !exists {
		t.Error("Expected go template to be loaded")
	}
}

// TestFindTool tests tool finding functionality
func TestGeneratorFindTool(t *testing.T) {
	testRepoDir, testConfigPath, testTemplatesDir := setupTestGeneratorEnv(t)
	
	options := GeneratorOptions{
		ConfigPath:  testConfigPath,
		OutputDir:   filepath.Join(testRepoDir, "scripts"),
		TemplateDir: testTemplatesDir,
	}
	
	generator, err := NewGenerator(options)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Test finding existing tool
	tool, found := generator.findTool("test-rust-tool")
	if !found {
		t.Error("Expected to find test-rust-tool")
	}
	
	if tool.Name != "test-rust-tool" {
		t.Errorf("Expected tool name 'test-rust-tool', got %s", tool.Name)
	}
	
	if tool.Language != "rust" {
		t.Errorf("Expected language 'rust', got %s", tool.Language)
	}
	
	// Test finding non-existent tool
	_, found = generator.findTool("nonexistent-tool")
	if found {
		t.Error("Expected not to find nonexistent-tool")
	}
}

// TestPrepareTemplateData tests template data preparation
func TestPrepareTemplateData(t *testing.T) {
	testRepoDir, testConfigPath, testTemplatesDir := setupTestGeneratorEnv(t)
	
	options := GeneratorOptions{
		ConfigPath:  testConfigPath,
		OutputDir:   filepath.Join(testRepoDir, "scripts"),
		TemplateDir: testTemplatesDir,
	}
	
	generator, err := NewGenerator(options)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	tool := generator.config.Tools[1] // test-go-tool with shell integration
	data := generator.prepareTemplateData(tool)
	
	// Verify template data fields
	if data.Tool.Name != "test-go-tool" {
		t.Errorf("Expected tool name 'test-go-tool', got %s", data.Tool.Name)
	}
	
	if !data.HasShell {
		t.Error("Expected HasShell to be true for go tool")
	}
	
	if data.InstallPath == "" {
		t.Error("Expected InstallPath to be set")
	}
	
	if data.BinaryPath == "" {
		t.Error("Expected BinaryPath to be set")
	}
	
	if data.RepoName != "go-tool" {
		t.Errorf("Expected RepoName 'go-tool', got %s", data.RepoName)
	}
	
	if data.ScriptName != "install-test-go-tool.sh" {
		t.Errorf("Expected ScriptName 'install-test-go-tool.sh', got %s", data.ScriptName)
	}
	
	// Check build types
	expectedBuildTypes := []string{"maximum", "minimal", "standard"}
	if len(data.BuildTypes) != len(expectedBuildTypes) {
		t.Errorf("Expected %d build types, got %d", len(expectedBuildTypes), len(data.BuildTypes))
	}
}

// TestGenerateScript tests script generation
func TestGenerateScript(t *testing.T) {
	testRepoDir, testConfigPath, testTemplatesDir := setupTestGeneratorEnv(t)
	
	options := GeneratorOptions{
		ConfigPath:  testConfigPath,
		OutputDir:   filepath.Join(testRepoDir, "scripts"),
		TemplateDir: testTemplatesDir,
		Force:       true,
		Validate:    false,
	}
	
	generator, err := NewGenerator(options)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	tool := generator.config.Tools[0] // test-rust-tool
	err = generator.generateScript(tool)
	if err != nil {
		t.Fatalf("Failed to generate script: %v", err)
	}
	
	// Check that script was created
	scriptPath := filepath.Join(options.OutputDir, "install-test-rust-tool.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Generated script does not exist")
	}
	
	// Check script content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read generated script: %v", err)
	}
	
	scriptContent := string(content)
	
	// Verify basic script structure
	if !strings.Contains(scriptContent, "#!/bin/bash") {
		t.Error("Generated script should have bash shebang")
	}
	
	if !strings.Contains(scriptContent, "test-rust-tool") {
		t.Error("Generated script should contain tool name")
	}
	
	if !strings.Contains(scriptContent, "Test Rust tool for generation") {
		t.Error("Generated script should contain tool description")
	}
	
	if !strings.Contains(scriptContent, "show_help()") {
		t.Error("Generated script should have show_help function")
	}
	
	// Check build type flags are present
	if !strings.Contains(scriptContent, "-m") || !strings.Contains(scriptContent, "-s") || !strings.Contains(scriptContent, "-r") {
		t.Error("Generated script should contain build type flags")
	}
}

// TestGenerateTools tests generating multiple tools
func TestGenerateTools(t *testing.T) {
	testRepoDir, testConfigPath, testTemplatesDir := setupTestGeneratorEnv(t)
	
	options := GeneratorOptions{
		ConfigPath:  testConfigPath,
		OutputDir:   filepath.Join(testRepoDir, "scripts"),
		TemplateDir: testTemplatesDir,
		Force:       true,
		Validate:    false,
	}
	
	generator, err := NewGenerator(options)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	toolNames := []string{"test-rust-tool", "test-go-tool"}
	err = generator.GenerateTools(toolNames)
	if err != nil {
		t.Fatalf("Failed to generate tools: %v", err)
	}
	
	// Check that both scripts were created
	for _, toolName := range toolNames {
		scriptPath := filepath.Join(options.OutputDir, "install-"+toolName+".sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			t.Errorf("Generated script does not exist for %s", toolName)
		}
		
		// Check script is executable
		info, err := os.Stat(scriptPath)
		if err != nil {
			t.Errorf("Failed to stat script for %s: %v", toolName, err)
		}
		
		if info.Mode()&0111 == 0 {
			t.Errorf("Generated script is not executable for %s", toolName)
		}
	}
}

// TestValidateScript tests script validation
func TestValidateScript(t *testing.T) {
	testRepoDir, testConfigPath, testTemplatesDir := setupTestGeneratorEnv(t)
	
	options := GeneratorOptions{
		ConfigPath:  testConfigPath,
		OutputDir:   filepath.Join(testRepoDir, "scripts"),
		TemplateDir: testTemplatesDir,
		Force:       true,
		Validate:    false,
	}
	
	generator, err := NewGenerator(options)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Generate a script first
	tool := generator.config.Tools[0]
	err = generator.generateScript(tool)
	if err != nil {
		t.Fatalf("Failed to generate script: %v", err)
	}
	
	scriptPath := filepath.Join(options.OutputDir, "install-test-rust-tool.sh")
	
	// Test validation
	err = generator.validateScript(scriptPath)
	if err != nil {
		t.Errorf("Script validation failed: %v", err)
	}
	
	// Test with non-existent script
	err = generator.validateScript("/nonexistent/script.sh")
	if err == nil {
		t.Error("Expected validation to fail for non-existent script")
	}
	
	// Test with non-executable script
	nonExecScript := filepath.Join(options.OutputDir, "non-exec.sh")
	if err := os.WriteFile(nonExecScript, []byte("#!/bin/bash\necho test"), 0644); err != nil {
		t.Fatalf("Failed to create non-executable script: %v", err)
	}
	
	err = generator.validateScript(nonExecScript)
	if err == nil {
		t.Error("Expected validation to fail for non-executable script")
	}
}

// TestTemplateFunctions tests template helper functions
func TestTemplateFunctions(t *testing.T) {
	funcs := templateFuncs()
	
	// Test upper function
	if upperFunc, exists := funcs["upper"]; exists {
		if fn, ok := upperFunc.(func(string) string); ok {
			result := fn("test")
			if result != "TEST" {
				t.Errorf("Expected 'TEST', got '%s'", result)
			}
		} else {
			t.Error("upper function has wrong type")
		}
	} else {
		t.Error("upper function not found in template functions")
	}
	
	// Test lower function
	if lowerFunc, exists := funcs["lower"]; exists {
		if fn, ok := lowerFunc.(func(string) string); ok {
			result := fn("TEST")
			if result != "test" {
				t.Errorf("Expected 'test', got '%s'", result)
			}
		} else {
			t.Error("lower function has wrong type")
		}
	} else {
		t.Error("lower function not found in template functions")
	}
}

// TestErrorHandling tests various error conditions
func TestGeneratorErrorHandling(t *testing.T) {
	// Test with non-existent config
	_, err := NewGenerator(GeneratorOptions{ConfigPath: "/nonexistent/config.json"})
	if err == nil {
		t.Error("Expected error with non-existent config")
	}
	
	// Test with non-existent template directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configData, _ := json.MarshalIndent(testGeneratorConfig, "", "  ")
	os.WriteFile(configPath, configData, 0644)
	
	configPath = configPath
	templateDir = "/nonexistent/templates"
	repoDir = tempDir
	
	_, err = NewGenerator(GeneratorOptions{})
	if err == nil {
		t.Error("Expected error with non-existent template directory")
	}
}

// BenchmarkGenerateScript benchmarks script generation
func BenchmarkGenerateScript(b *testing.B) {
	repoDir, configPath, templatesDir := setupTestGeneratorEnv(&testing.T{})
	
	configPath = configPath
	outputDir = filepath.Join(repoDir, "scripts")
	templateDir = templatesDir
	repoDir = repoDir
	
	generator, err := NewGenerator(GeneratorOptions{Force: true, Validate: false})
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}
	
	tool := generator.config.Tools[0]
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := generator.generateScript(tool)
		if err != nil {
			b.Fatalf("Failed to generate script: %v", err)
		}
	}
}