package validation

import (
	"testing"
)

func TestValidateToolConfig(t *testing.T) {
	// Test valid tool config
	validTool := ToolConfig{
		Name:        "test-tool",
		Description: "Test tool for validation testing",
		Category:    "development",
		Repository:  "https://github.com/test/tool.git",
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
	}

	if err := ValidateToolConfig(validTool); err != nil {
		t.Errorf("Valid tool config should pass validation: %v", err)
	}

	// Test invalid name
	invalidTool := validTool
	invalidTool.Name = "a" // Too short
	if err := ValidateToolConfig(invalidTool); err == nil {
		t.Error("Tool with name too short should fail validation")
	}

	// Test invalid category
	invalidTool = validTool
	invalidTool.Category = "invalid-category"
	if err := ValidateToolConfig(invalidTool); err == nil {
		t.Error("Tool with invalid category should fail validation")
	}

	// Test empty build types
	invalidTool = validTool
	invalidTool.BuildTypes = map[string]string{}
	if err := ValidateToolConfig(invalidTool); err == nil {
		t.Error("Tool with empty build types should fail validation")
	}

	// Test invalid repository URL
	invalidTool = validTool
	invalidTool.Repository = "not-a-url"
	if err := ValidateToolConfig(invalidTool); err == nil {
		t.Error("Tool with invalid repository URL should fail validation")
	}
}

func TestValidateConfig(t *testing.T) {
	// Test valid config
	validConfig := Config{
		SchemaVersion:    "1.0",
		DefaultBuildType: "standard",
		Tools: []ToolConfig{
			{
				Name:        "test-tool",
				Description: "Test tool for validation testing",
				Category:    "development",
				Repository:  "https://github.com/test/tool.git",
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
		},
		Categories: map[string]string{
			"development": "Development Tools",
		},
		Languages: map[string]LanguageConfig{
			"rust": {
				MinVersion: "1.88.0",
				BuildTool:  "cargo",
			},
		},
	}

	if err := ValidateConfig(validConfig); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Test empty tools
	invalidConfig := validConfig
	invalidConfig.Tools = []ToolConfig{}
	if err := ValidateConfig(invalidConfig); err == nil {
		t.Error("Config with empty tools should fail validation")
	}

	// Test duplicate tool names
	invalidConfig = validConfig
	invalidConfig.Tools = append(invalidConfig.Tools, invalidConfig.Tools[0])
	if err := ValidateConfig(invalidConfig); err == nil {
		t.Error("Config with duplicate tool names should fail validation")
	}
}

func TestValidationHelpers(t *testing.T) {
	// Test validateToolName
	if err := validateToolName("valid-tool_name"); err != nil {
		t.Errorf("Valid tool name should pass: %v", err)
	}

	if err := validateToolName("a"); err == nil {
		t.Error("Too short tool name should fail")
	}

	if err := validateToolName("invalid@name"); err == nil {
		t.Error("Tool name with invalid characters should fail")
	}

	// Test validateURL
	if err := validateURL("https://github.com/test/repo.git"); err != nil {
		t.Errorf("Valid URL should pass: %v", err)
	}

	if err := validateURL("not-a-url"); err == nil {
		t.Error("Invalid URL should fail")
	}

	// Test isValidBuildType
	if !isValidBuildType("minimal") {
		t.Error("minimal should be valid build type")
	}

	if isValidBuildType("invalid") {
		t.Error("invalid should not be valid build type")
	}

	// Test isValidCategory
	if !isValidCategory("development") {
		t.Error("development should be valid category")
	}

	if isValidCategory("invalid") {
		t.Error("invalid should not be valid category")
	}

	// Test isValidLanguage
	if !isValidLanguage("rust") {
		t.Error("rust should be valid language")
	}

	if isValidLanguage("invalid") {
		t.Error("invalid should not be valid language")
	}
}