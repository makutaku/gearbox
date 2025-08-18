package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
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

// Global configuration
var config Config
var configPath string

// Main is the entry point for the configuration manager command-line tool.
// It provides functionality to manage gearbox configuration settings.
func Main() {
	var rootCmd = &cobra.Command{
		Use:   "config-manager",
		Short: "Gearbox configuration management utility",
		Long:  "A command-line tool for managing gearbox tool configurations and build flags",
	}

	// Add commands
	rootCmd.AddCommand(loadCmd())
	rootCmd.AddCommand(getBuildFlagCmd())
	rootCmd.AddCommand(listToolsCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(generateCmd())

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to tools.json configuration file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// loadCmd loads and parses the configuration file
func loadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "load",
		Short: "Load and validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			fmt.Printf("Configuration loaded successfully: %d tools, %d categories\n", 
				len(config.Tools), len(config.Categories))
			return nil
		},
	}
}

// getBuildFlagCmd gets build flag for a specific tool and build type
func getBuildFlagCmd() *cobra.Command {
	var buildType string
	
	cmd := &cobra.Command{
		Use:   "build-flag [tool-name]",
		Short: "Get build flag for a tool and build type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			
			toolName := args[0]
			if buildType == "" {
				buildType = config.DefaultBuildType
			}
			
			flag, err := getBuildFlag(toolName, buildType)
			if err != nil {
				return err
			}
			
			fmt.Print(flag)
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&buildType, "build-type", "b", "", "Build type (minimal, standard, maximum)")
	return cmd
}

// listToolsCmd lists all available tools
func listToolsCmd() *cobra.Command {
	var category string
	var verbose bool
	
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			
			return listTools(category, verbose)
		},
	}
	
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	return cmd
}

// validateCmd validates the configuration file
func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			
			if err := validateConfig(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			
			fmt.Println("Configuration is valid")
			return nil
		},
	}
}

// generateCmd generates shell script helper functions
func generateCmd() *cobra.Command {
	var output string
	
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate shell script helper functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			
			return generateShellHelpers(output)
		},
	}
	
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")
	return cmd
}

// loadConfig loads the configuration from file
func loadConfig() error {
	if configPath == "" {
		// Try to find config file in standard locations
		possiblePaths := []string{
			"config/tools.json",
			"../config/tools.json",
			"../../config/tools.json",
		}
		
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
		
		if configPath == "" {
			return fmt.Errorf("configuration file not found. Use --config to specify path")
		}
	}
	
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file %s: %w", configPath, err)
	}
	defer file.Close()
	
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse JSON configuration: %w", err)
	}
	
	return nil
}

// getBuildFlag returns the build flag for a tool and build type
func getBuildFlag(toolName, buildType string) (string, error) {
	for _, tool := range config.Tools {
		if tool.Name == toolName {
			if flag, exists := tool.BuildTypes[buildType]; exists {
				return flag, nil
			}
			return "", fmt.Errorf("build type '%s' not supported for tool '%s'", buildType, toolName)
		}
	}
	return "", fmt.Errorf("tool '%s' not found", toolName)
}

// listTools lists all tools, optionally filtered by category
func listTools(category string, verbose bool) error {
	// Group tools by category
	toolsByCategory := make(map[string][]ToolConfig)
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
	
	// Print tools by category
	for _, cat := range categories {
		if description, exists := config.Categories[cat]; exists {
			fmt.Printf("\n%s:\n", description)
		} else {
			fmt.Printf("\n%s:\n", strings.Title(cat))
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

// validateConfig validates the loaded configuration
func validateConfig() error {
	var errors []string
	
	// Validate schema version
	if config.SchemaVersion == "" {
		errors = append(errors, "schema_version is required")
	}
	
	// Validate default build type
	validBuildTypes := map[string]bool{"minimal": true, "standard": true, "maximum": true}
	if !validBuildTypes[config.DefaultBuildType] {
		errors = append(errors, fmt.Sprintf("invalid default_build_type: %s", config.DefaultBuildType))
	}
	
	// Validate tools
	toolNames := make(map[string]bool)
	for i, tool := range config.Tools {
		// Check for duplicate tool names
		if toolNames[tool.Name] {
			errors = append(errors, fmt.Sprintf("duplicate tool name: %s", tool.Name))
		}
		toolNames[tool.Name] = true
		
		// Validate required fields
		if tool.Name == "" {
			errors = append(errors, fmt.Sprintf("tool %d: name is required", i))
		}
		if tool.Repository == "" {
			errors = append(errors, fmt.Sprintf("tool %s: repository is required", tool.Name))
		}
		if tool.BinaryName == "" {
			errors = append(errors, fmt.Sprintf("tool %s: binary_name is required", tool.Name))
		}
		
		// Validate build types
		for buildType := range tool.BuildTypes {
			if !validBuildTypes[buildType] {
				errors = append(errors, fmt.Sprintf("tool %s: invalid build_type: %s", tool.Name, buildType))
			}
		}
		
		// Validate category exists
		if _, exists := config.Categories[tool.Category]; !exists {
			errors = append(errors, fmt.Sprintf("tool %s: unknown category: %s", tool.Name, tool.Category))
		}
		
		// Validate language exists
		if _, exists := config.Languages[tool.Language]; !exists {
			errors = append(errors, fmt.Sprintf("tool %s: unknown language: %s", tool.Name, tool.Language))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  %s", strings.Join(errors, "\n  "))
	}
	
	return nil
}

// generateShellHelpers generates shell script helper functions
func generateShellHelpers(outputPath string) error {
	var output io.Writer = os.Stdout
	
	if outputPath != "" {
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		output = file
	}
	
	configManagerPath, _ := filepath.Abs(os.Args[0])
	
	fmt.Fprintf(output, "#!/bin/bash\n")
	fmt.Fprintf(output, "# Generated shell helpers for gearbox configuration\n")
	fmt.Fprintf(output, "# Generated from: %s\n\n", configPath)
	
	fmt.Fprintf(output, "# Configuration manager path\n")
	fmt.Fprintf(output, "CONFIG_MANAGER=\"%s\"\n", configManagerPath)
	fmt.Fprintf(output, "CONFIG_FILE=\"%s\"\n\n", configPath)
	
	fmt.Fprintf(output, "# Get build flag for a tool and build type\n")
	fmt.Fprintf(output, "get_build_flag() {\n")
	fmt.Fprintf(output, "    local tool=\"$1\"\n")
	fmt.Fprintf(output, "    local build_type=\"${2:-standard}\"\n")
	fmt.Fprintf(output, "    \n")
	fmt.Fprintf(output, "    \"$CONFIG_MANAGER\" --config=\"$CONFIG_FILE\" build-flag \"$tool\" --build-type=\"$build_type\" 2>/dev/null\n")
	fmt.Fprintf(output, "}\n\n")
	
	fmt.Fprintf(output, "# List all available tools\n")
	fmt.Fprintf(output, "list_tools() {\n")
	fmt.Fprintf(output, "    \"$CONFIG_MANAGER\" --config=\"$CONFIG_FILE\" list \"$@\"\n")
	fmt.Fprintf(output, "}\n\n")
	
	fmt.Fprintf(output, "# Validate configuration\n")
	fmt.Fprintf(output, "validate_config() {\n")
	fmt.Fprintf(output, "    \"$CONFIG_MANAGER\" --config=\"$CONFIG_FILE\" validate\n")
	fmt.Fprintf(output, "}\n\n")
	
	fmt.Fprintf(output, "# Get tool info as JSON (requires jq)\n")
	fmt.Fprintf(output, "get_tool_info() {\n")
	fmt.Fprintf(output, "    local tool=\"$1\"\n")
	fmt.Fprintf(output, "    [[ -f \"$CONFIG_FILE\" ]] && jq -r \".tools[] | select(.name == \\\"$tool\\\")\" \"$CONFIG_FILE\"\n")
	fmt.Fprintf(output, "}\n\n")
	
	fmt.Fprintf(output, "# Get all tools in category\n")
	fmt.Fprintf(output, "get_tools_in_category() {\n")
	fmt.Fprintf(output, "    local category=\"$1\"\n")
	fmt.Fprintf(output, "    [[ -f \"$CONFIG_FILE\" ]] && jq -r \".tools[] | select(.category == \\\"$category\\\") | .name\" \"$CONFIG_FILE\"\n")
	fmt.Fprintf(output, "}\n\n")
	
	// Generate validation function for available tools
	fmt.Fprintf(output, "# Validate tool name (generated from config)\n")
	fmt.Fprintf(output, "validate_tool_name() {\n")
	fmt.Fprintf(output, "    local tool=\"$1\"\n")
	fmt.Fprintf(output, "    case \"$tool\" in\n")
	
	var toolNames []string
	for _, tool := range config.Tools {
		toolNames = append(toolNames, tool.Name)
	}
	sort.Strings(toolNames)
	
	fmt.Fprintf(output, "        %s)\n", strings.Join(toolNames, "|"))
	fmt.Fprintf(output, "            return 0\n")
	fmt.Fprintf(output, "            ;;\n")
	fmt.Fprintf(output, "        *)\n")
	fmt.Fprintf(output, "            echo \"ERROR: Invalid tool name: '$tool'. Use 'gearbox list' to see available tools.\" >&2\n")
	fmt.Fprintf(output, "            return 1\n")
	fmt.Fprintf(output, "            ;;\n")
	fmt.Fprintf(output, "    esac\n")
	fmt.Fprintf(output, "}\n\n")
	
	if outputPath != "" {
		fmt.Printf("Shell helpers generated: %s\n", outputPath)
	}
	
	return nil
}