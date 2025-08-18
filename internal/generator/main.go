package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

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
	CustomCommands   map[string]string `json:"custom_commands,omitempty"`
	BuildConfig      BuildConfig       `json:"build_config,omitempty"`
}

// BuildConfig represents language-specific build configuration
type BuildConfig struct {
	PreBuildSteps    []string          `json:"pre_build_steps,omitempty"`
	PostBuildSteps   []string          `json:"post_build_steps,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	InstallPath      string            `json:"install_path,omitempty"`
	BinaryPath       string            `json:"binary_path,omitempty"`
	ConfigFiles      []string          `json:"config_files,omitempty"`
	SystemDeps       []string          `json:"system_deps,omitempty"`
}

// LanguageConfig represents language-specific configuration
type LanguageConfig struct {
	MinVersion    string            `json:"min_version"`
	BuildTool     string            `json:"build_tool"`
	InstallCmd    string            `json:"install_cmd"`
	TestCmd       string            `json:"test_cmd"`
	CleanCmd      string            `json:"clean_cmd"`
	Environment   map[string]string `json:"environment,omitempty"`
	Dependencies  []string          `json:"dependencies,omitempty"`
	SetupSteps    []string          `json:"setup_steps,omitempty"`
}

// Config represents the complete configuration structure
type Config struct {
	SchemaVersion    string                     `json:"schema_version"`
	DefaultBuildType string                     `json:"default_build_type"`
	Tools            []ToolConfig               `json:"tools"`
	Categories       map[string]string          `json:"categories"`
	Languages        map[string]LanguageConfig  `json:"languages"`
}

// TemplateData represents data passed to templates
type TemplateData struct {
	Tool           ToolConfig
	Language       LanguageConfig
	BuildTypes     []string
	HasPCRE2       bool
	HasShell       bool
	InstallPath    string
	BinaryPath     string
	RepoName       string
	ScriptName     string
}

// GeneratorOptions represents generation options
type GeneratorOptions struct {
	ConfigPath   string
	OutputDir    string
	TemplateDir  string
	Language     string
	Tools        []string
	Force        bool
	Validate     bool
	DryRun       bool
}

// Global variables
var (
	configPath  string
	outputDir   string
	templateDir string
	repoDir     string
)

// Main is the entry point for the script generator command-line tool.
// It provides functionality to generate installation scripts from templates.
func Main() {
	var rootCmd = &cobra.Command{
		Use:   "script-generator",
		Short: "Gearbox script generator",
		Long:  "Template-based generator for optimized tool installation scripts",
	}

	// Add commands
	rootCmd.AddCommand(generateCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(cleanCmd())

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to tools.json configuration file")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Output directory for generated scripts")
	rootCmd.PersistentFlags().StringVarP(&templateDir, "templates", "t", "", "Template directory")
	rootCmd.PersistentFlags().StringVar(&repoDir, "repo-dir", "", "Repository directory (default: auto-detect)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// generateCmd creates the generate command
func generateCmd() *cobra.Command {
	var opts GeneratorOptions

	cmd := &cobra.Command{
		Use:   "generate [tools...]",
		Short: "Generate optimized installation scripts from templates",
		Long: `Generate tool installation scripts using language-specific templates.
If no tools are specified, all tools will be generated.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			generator, err := NewGenerator(opts)
			if err != nil {
				return fmt.Errorf("failed to initialize generator: %w", err)
			}

			if len(args) == 0 {
				// Generate all tools
				return generator.GenerateAll()
			} else {
				// Generate specific tools
				return generator.GenerateTools(args)
			}
		},
	}

	// Generation options
	cmd.Flags().StringVarP(&opts.Language, "language", "l", "", "Generate scripts for specific language only")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing scripts")
	cmd.Flags().BoolVar(&opts.Validate, "validate", true, "Validate generated scripts")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be generated without creating files")

	return cmd
}

// listCmd creates the list command
func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available tools and templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator, err := NewGenerator(GeneratorOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize generator: %w", err)
			}

			return generator.ListTools()
		},
	}
}

// validateCmd creates the validate command
func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [scripts...]",
		Short: "Validate generated scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator, err := NewGenerator(GeneratorOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize generator: %w", err)
			}

			return generator.ValidateScripts(args)
		},
	}
}

// cleanCmd creates the clean command
func cleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean generated scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator, err := NewGenerator(GeneratorOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize generator: %w", err)
			}

			return generator.CleanGenerated()
		},
	}
}

// Generator manages script generation
type Generator struct {
	config      Config
	options     GeneratorOptions
	repoDir     string
	outputDir   string
	templateDir string
	templates   map[string]*template.Template
}

// NewGenerator creates a new script generator instance with the given options.
// It loads configuration and templates, then initializes the generator for script creation.
func NewGenerator(options GeneratorOptions) (*Generator, error) {
	// Set up repository directory
	localRepoDir := repoDir
	if localRepoDir == "" {
		if wd, err := os.Getwd(); err == nil {
			localRepoDir = wd
		} else {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Use provided paths or auto-detect
	localConfigPath := options.ConfigPath
	if localConfigPath == "" {
		localConfigPath = filepath.Join(localRepoDir, "config", "tools.json")
	}
	
	localOutputDir := options.OutputDir
	if localOutputDir == "" {
		localOutputDir = filepath.Join(localRepoDir, "scripts")
	}
	
	localTemplateDir := options.TemplateDir
	if localTemplateDir == "" {
		localTemplateDir = filepath.Join(localRepoDir, "templates")
	}

	// Load configuration
	config, err := loadConfig(localConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	generator := &Generator{
		config:      config,
		options:     options,
		repoDir:     localRepoDir,
		outputDir:   localOutputDir,
		templateDir: localTemplateDir,
		templates:   make(map[string]*template.Template),
	}

	// Load templates
	if err := generator.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return generator, nil
}

// loadConfig loads the configuration from file
func loadConfig(path string) (Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return config, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse JSON configuration: %w", err)
	}

	return config, nil
}

// templateFuncs returns template functions for string manipulation
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"join":  strings.Join,
		"split": strings.Split,
		"printf": fmt.Sprintf,
	}
}

// loadTemplates loads all template files
func (g *Generator) loadTemplates() error {
	// Ensure template directory exists
	if _, err := os.Stat(g.templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory not found: %s", g.templateDir)
	}

	// Load base template
	baseTemplate := filepath.Join(g.templateDir, "base.sh.tmpl")
	if _, err := os.Stat(baseTemplate); err == nil {
		tmpl, err := template.New("base.sh.tmpl").Funcs(templateFuncs()).ParseFiles(baseTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse base template: %w", err)
		}
		g.templates["base"] = tmpl
	}

	// Load language-specific templates
	languages := []string{"rust", "go", "python", "c"}
	for _, lang := range languages {
		templatePath := filepath.Join(g.templateDir, fmt.Sprintf("%s.sh.tmpl", lang))
		if _, err := os.Stat(templatePath); err == nil {
			tmpl, err := template.New(fmt.Sprintf("%s.sh.tmpl", lang)).Funcs(templateFuncs()).ParseFiles(templatePath)
			if err != nil {
				return fmt.Errorf("failed to parse %s template: %w", lang, err)
			}
			g.templates[lang] = tmpl
		}
	}

	return nil
}

// GenerateAll generates scripts for all tools
func (g *Generator) GenerateAll() error {
	fmt.Printf("🔧 Generating scripts for %d tools\n\n", len(g.config.Tools))

	// Group tools by language for optimal processing
	languageGroups := make(map[string][]ToolConfig)
	for _, tool := range g.config.Tools {
		languageGroups[tool.Language] = append(languageGroups[tool.Language], tool)
	}

	var generated, failed int

	// Process each language group
	for lang, tools := range languageGroups {
		fmt.Printf("📦 Generating %s tools (%d scripts)\n", strings.Title(lang), len(tools))
		
		for _, tool := range tools {
			if err := g.generateScript(tool); err != nil {
				fmt.Printf("❌ Failed to generate %s: %v\n", tool.Name, err)
				failed++
			} else {
				fmt.Printf("✅ Generated %s\n", tool.Name)
				generated++
			}
		}
		fmt.Println()
	}

	fmt.Printf("📊 Generation Summary\n")
	fmt.Printf("Generated: %d\n", generated)
	fmt.Printf("Failed: %d\n", failed)

	if failed > 0 {
		return fmt.Errorf("%d scripts failed to generate", failed)
	}

	fmt.Printf("\n🎉 All scripts generated successfully!\n")
	return nil
}

// GenerateTools generates scripts for specific tools
func (g *Generator) GenerateTools(toolNames []string) error {
	fmt.Printf("🔧 Generating scripts for %d specified tools\n\n", len(toolNames))

	var generated, failed int

	for _, name := range toolNames {
		tool, found := g.findTool(name)
		if !found {
			fmt.Printf("❌ Tool not found: %s\n", name)
			failed++
			continue
		}

		if err := g.generateScript(tool); err != nil {
			fmt.Printf("❌ Failed to generate %s: %v\n", name, err)
			failed++
		} else {
			fmt.Printf("✅ Generated %s\n", name)
			generated++
		}
	}

	fmt.Printf("\n📊 Generation Summary\n")
	fmt.Printf("Generated: %d\n", generated)
	fmt.Printf("Failed: %d\n", failed)

	if failed > 0 {
		return fmt.Errorf("%d scripts failed to generate", failed)
	}

	return nil
}

// generateScript generates a script for a single tool
func (g *Generator) generateScript(tool ToolConfig) error {
	// Get template for this tool's language
	tmpl, exists := g.templates[tool.Language]
	if !exists {
		// Fall back to base template
		tmpl, exists = g.templates["base"]
		if !exists {
			return fmt.Errorf("no template found for language %s and no base template", tool.Language)
		}
	}

	// Prepare template data
	data := g.prepareTemplateData(tool)

	// Generate script content
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Determine output file path
	scriptName := fmt.Sprintf("install-%s.sh", tool.Name)
	outputPath := filepath.Join(g.outputDir, scriptName)

	// Check if file exists and force flag
	if _, err := os.Stat(outputPath); err == nil && !g.options.Force {
		return fmt.Errorf("script already exists (use --force to overwrite): %s", outputPath)
	}

	// Write to file
	if g.options.DryRun {
		fmt.Printf("Would generate: %s\n", outputPath)
		return nil
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0755); err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	// Validate if requested
	if g.options.Validate {
		if err := g.validateScript(outputPath); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// prepareTemplateData prepares data for template execution
func (g *Generator) prepareTemplateData(tool ToolConfig) TemplateData {
	langConfig := g.config.Languages[tool.Language]
	
	// Extract build types
	var buildTypes []string
	for buildType := range tool.BuildTypes {
		buildTypes = append(buildTypes, buildType)
	}
	sort.Strings(buildTypes)

	// Determine paths
	installPath := tool.BuildConfig.InstallPath
	if installPath == "" {
		installPath = "/usr/local/bin"
	}

	binaryPath := tool.BuildConfig.BinaryPath
	if binaryPath == "" {
		binaryPath = filepath.Join(installPath, tool.BinaryName)
	}

	// Extract repository name
	repoName := filepath.Base(tool.Repository)
	if strings.HasSuffix(repoName, ".git") {
		repoName = strings.TrimSuffix(repoName, ".git")
	}

	return TemplateData{
		Tool:         tool,
		Language:     langConfig,
		BuildTypes:   buildTypes,
		HasPCRE2:     strings.Contains(strings.Join(buildTypes, " "), "pcre2"),
		HasShell:     tool.ShellIntegration,
		InstallPath:  installPath,
		BinaryPath:   binaryPath,
		RepoName:     repoName,
		ScriptName:   fmt.Sprintf("install-%s.sh", tool.Name),
	}
}

// findTool finds a tool by name in the configuration
func (g *Generator) findTool(name string) (ToolConfig, bool) {
	for _, tool := range g.config.Tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return ToolConfig{}, false
}

// ListTools lists all available tools
func (g *Generator) ListTools() error {
	// Group tools by language
	languageGroups := make(map[string][]ToolConfig)
	for _, tool := range g.config.Tools {
		languageGroups[tool.Language] = append(languageGroups[tool.Language], tool)
	}

	fmt.Printf("📋 Available Tools for Script Generation\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for lang, tools := range languageGroups {
		fmt.Printf("📦 %s (%d tools)\n", strings.Title(lang), len(tools))
		
		sort.Slice(tools, func(i, j int) bool {
			return tools[i].Name < tools[j].Name
		})

		for _, tool := range tools {
			fmt.Printf("  %-15s %s\n", tool.Name, tool.Description)
		}
		fmt.Println()
	}

	fmt.Printf("🎯 Template Status\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	for lang := range languageGroups {
		status := "❌ Missing"
		if _, exists := g.templates[lang]; exists {
			status = "✅ Available"
		}
		fmt.Printf("  %-10s %s\n", strings.Title(lang), status)
	}

	return nil
}

// ValidateScripts validates generated scripts
func (g *Generator) ValidateScripts(scriptPaths []string) error {
	// If no scripts specified, validate all in output directory
	if len(scriptPaths) == 0 {
		files, err := filepath.Glob(filepath.Join(g.outputDir, "install-*.sh"))
		if err != nil {
			return fmt.Errorf("failed to find scripts: %w", err)
		}
		scriptPaths = files
	}

	fmt.Printf("🔍 Validating %d scripts\n\n", len(scriptPaths))

	var validated, failed int

	for _, scriptPath := range scriptPaths {
		if err := g.validateScript(scriptPath); err != nil {
			fmt.Printf("❌ %s: %v\n", filepath.Base(scriptPath), err)
			failed++
		} else {
			fmt.Printf("✅ %s\n", filepath.Base(scriptPath))
			validated++
		}
	}

	fmt.Printf("\n📊 Validation Summary\n")
	fmt.Printf("Validated: %d\n", validated)
	fmt.Printf("Failed: %d\n", failed)

	if failed > 0 {
		return fmt.Errorf("%d scripts failed validation", failed)
	}

	return nil
}

// validateScript validates a single script
func (g *Generator) validateScript(scriptPath string) error {
	// Check if file exists and is executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		return fmt.Errorf("script not found: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return fmt.Errorf("script not executable")
	}

	// Basic syntax check (bash -n)
	// This would run: bash -n scriptPath
	// For now, just check basic script structure
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}

	scriptContent := string(content)

	// Check for shebang
	if !strings.HasPrefix(scriptContent, "#!/bin/bash") {
		return fmt.Errorf("missing bash shebang")
	}

	// Check for basic structure
	requiredElements := []string{
		"set -e",
		"show_help()",
		"error()",
		"log()",
	}

	for _, element := range requiredElements {
		if !strings.Contains(scriptContent, element) {
			return fmt.Errorf("missing required element: %s", element)
		}
	}

	return nil
}

// CleanGenerated removes all generated scripts
func (g *Generator) CleanGenerated() error {
	pattern := filepath.Join(g.outputDir, "install-*.sh")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find generated scripts: %w", err)
	}

	if len(files) == 0 {
		fmt.Printf("No generated scripts found in %s\n", g.outputDir)
		return nil
	}

	fmt.Printf("🧹 Cleaning %d generated scripts\n", len(files))

	var removed int
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			fmt.Printf("❌ Failed to remove %s: %v\n", filepath.Base(file), err)
		} else {
			fmt.Printf("✅ Removed %s\n", filepath.Base(file))
			removed++
		}
	}

	fmt.Printf("\n📊 Cleanup Summary: %d scripts removed\n", removed)
	return nil
}