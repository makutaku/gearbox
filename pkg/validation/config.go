package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ToolConfig represents a single tool configuration with validation
type ToolConfig struct {
	Name             string            `json:"name" validate:"required,alphanum_dash,min=2,max=30"`
	Description      string            `json:"description" validate:"required,min=10,max=200"`
	Category         string            `json:"category" validate:"required,oneof=core navigation development system-monitoring text-processing analysis media"`
	Repository       string            `json:"repository" validate:"required,url"`
	BinaryName       string            `json:"binary_name" validate:"required,alphanum_dash,min=2,max=30"`
	Language         string            `json:"language" validate:"required,oneof=rust go python c"`
	BuildTypes       map[string]string `json:"build_types" validate:"required,min=1"`
	Dependencies     []string          `json:"dependencies" validate:"dive,alphanum_dash"`
	MinVersion       string            `json:"min_version"`
	ShellIntegration bool              `json:"shell_integration"`
	TestCommand      string            `json:"test_command" validate:"required"`
}

// Config represents the complete configuration structure with validation
type Config struct {
	SchemaVersion    string                    `json:"schema_version" validate:"required,semver"`
	DefaultBuildType string                    `json:"default_build_type" validate:"required,oneof=minimal standard maximum"`
	Tools            []ToolConfig              `json:"tools" validate:"required,min=1,dive"`
	Categories       map[string]string         `json:"categories"`
	Languages        map[string]LanguageConfig `json:"languages"`
}

// LanguageConfig represents language-specific configuration
type LanguageConfig struct {
	MinVersion string `json:"min_version" validate:"semver"`
	BuildTool  string `json:"build_tool" validate:"required"`
}

// ValidateConfig validates the entire configuration structure
func ValidateConfig(config Config) error {
	// Validate schema version format
	if err := validateSemVer(config.SchemaVersion); err != nil {
		return fmt.Errorf("invalid schema_version: %w", err)
	}

	// Validate default build type
	if !isValidBuildType(config.DefaultBuildType) {
		return fmt.Errorf("invalid default_build_type: %s", config.DefaultBuildType)
	}

	// Validate tools
	if len(config.Tools) == 0 {
		return fmt.Errorf("tools array cannot be empty")
	}

	toolNames := make(map[string]bool)
	binaryNames := make(map[string]bool)

	for i, tool := range config.Tools {
		if err := ValidateToolConfig(tool); err != nil {
			return fmt.Errorf("invalid tool at index %d: %w", i, err)
		}

		// Check for duplicate tool names
		if toolNames[tool.Name] {
			return fmt.Errorf("duplicate tool name: %s", tool.Name)
		}
		toolNames[tool.Name] = true

		// Check for duplicate binary names
		if binaryNames[tool.BinaryName] {
			return fmt.Errorf("duplicate binary name: %s", tool.BinaryName)
		}
		binaryNames[tool.BinaryName] = true
	}

	return nil
}

// ValidateToolConfig validates a single tool configuration
func ValidateToolConfig(tool ToolConfig) error {
	// Validate name
	if err := validateToolName(tool.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	// Validate description
	if len(tool.Description) < 10 || len(tool.Description) > 200 {
		return fmt.Errorf("description must be between 10 and 200 characters")
	}

	// Validate category
	if !isValidCategory(tool.Category) {
		return fmt.Errorf("invalid category: %s", tool.Category)
	}

	// Validate repository URL
	if err := validateURL(tool.Repository); err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Validate binary name
	if err := validateToolName(tool.BinaryName); err != nil {
		return fmt.Errorf("invalid binary_name: %w", err)
	}

	// Validate language
	if !isValidLanguage(tool.Language) {
		return fmt.Errorf("invalid language: %s", tool.Language)
	}

	// Validate build types
	if len(tool.BuildTypes) == 0 {
		return fmt.Errorf("build_types cannot be empty")
	}

	for buildType, flag := range tool.BuildTypes {
		if !isValidBuildType(buildType) {
			return fmt.Errorf("invalid build type: %s", buildType)
		}
		if flag == "" {
			return fmt.Errorf("build flag for %s cannot be empty", buildType)
		}
	}

	// Validate dependencies
	for _, dep := range tool.Dependencies {
		if err := validateDependencyName(dep); err != nil {
			return fmt.Errorf("invalid dependency: %w", err)
		}
	}

	// Validate test command
	if tool.TestCommand == "" {
		return fmt.Errorf("test_command cannot be empty")
	}

	return nil
}

// validateToolName validates tool and binary names
func validateToolName(name string) error {
	if len(name) < 2 || len(name) > 30 {
		return fmt.Errorf("name must be between 2 and 30 characters")
	}

	// Allow alphanumeric, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	if !matched {
		return fmt.Errorf("name can only contain alphanumeric characters, hyphens, and underscores")
	}

	return nil
}

// validateURL validates repository URLs
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// validateSemVer validates semantic version strings
func validateSemVer(version string) error {
	matched, _ := regexp.MatchString(`^\d+\.\d+(\.\d+)?$`, version)
	if !matched {
		return fmt.Errorf("must be in semantic version format (e.g., 1.0, 1.0.0)")
	}
	return nil
}

// validateDependencyName validates dependency names
func validateDependencyName(name string) error {
	if name == "" {
		return fmt.Errorf("dependency name cannot be empty")
	}

	// Allow alphanumeric, hyphens, underscores, and periods for package names
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._-]+$`, name)
	if !matched {
		return fmt.Errorf("dependency name can only contain alphanumeric characters, periods, hyphens, and underscores")
	}

	return nil
}

// isValidBuildType checks if a build type is valid
func isValidBuildType(buildType string) bool {
	validTypes := map[string]bool{
		"minimal":  true,
		"standard": true,
		"maximum":  true,
	}
	return validTypes[buildType]
}

// isValidCategory checks if a category is valid
func isValidCategory(category string) bool {
	validCategories := map[string]bool{
		"core":        true,
		"navigation":  true,
		"development": true,
		"system":      true,
		"media":       true,
	}
	return validCategories[category]
}

// isValidLanguage checks if a language is valid
func isValidLanguage(language string) bool {
	validLanguages := map[string]bool{
		"rust":   true,
		"go":     true,
		"python": true,
		"c":      true,
	}
	return validLanguages[strings.ToLower(language)]
}