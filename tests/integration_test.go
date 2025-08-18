package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullWorkflow tests the complete gearbox workflow
func TestFullWorkflow(t *testing.T) {
	// Create temporary directory for test (for future use)
	_ = t.TempDir()
	
	// Test that we can build the CLI and tools
	t.Run("BuildSystem", func(t *testing.T) {
		// This test verifies the build system works
		// In a real integration test, we would run make build here
		// For now, we just check that the source files exist
		
		expectedFiles := []string{
			"../cmd/gearbox/main.go",
			"../internal/orchestrator/main.go",
			"../internal/script-generator/main.go",
			"../internal/config-manager/main.go",
			"../config/tools.json",
		}
		
		for _, file := range expectedFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected file does not exist: %s", file)
			}
		}
	})
	
	// Test configuration validation
	t.Run("ConfigurationValidation", func(t *testing.T) {
		configPath := "../config/tools.json"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Skip("tools.json not found, skipping validation test")
		}
		
		// In a real test, we would load and validate the configuration
		// For now, we just verify the file exists and is readable
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}
		
		if len(content) == 0 {
			t.Error("Config file is empty")
		}
		
		// Basic JSON structure check
		if !strings.Contains(string(content), "tools") {
			t.Error("Config file should contain tools section")
		}
	})
	
	// Test shared library functions
	t.Run("SharedLibrary", func(t *testing.T) {
		libPath := "../lib/common.sh"
		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			t.Skip("common.sh not found, skipping library test")
		}
		
		content, err := os.ReadFile(libPath)
		if err != nil {
			t.Fatalf("Failed to read library file: %v", err)
		}
		
		libContent := string(content)
		
		// Check for essential functions
		requiredFunctions := []string{
			"log()",
			"error()",
			"success()",
			"warning()",
			"add_rollback_action()",
			"execute_rollback()",
		}
		
		for _, fn := range requiredFunctions {
			if !strings.Contains(libContent, fn) {
				t.Errorf("Library should contain function: %s", fn)
			}
		}
	})
	
	// Test script directory structure
	t.Run("ScriptStructure", func(t *testing.T) {
		scriptsDir := "../scripts"
		if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
			t.Skip("scripts directory not found, skipping structure test")
		}
		
		// Check for essential scripts
		essentialScripts := []string{
			"install-all-tools.sh",
			"install-common-deps.sh",
			"install-fd.sh",
			"install-ripgrep.sh",
		}
		
		for _, script := range essentialScripts {
			scriptPath := filepath.Join(scriptsDir, script)
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				t.Errorf("Essential script not found: %s", script)
			}
		}
	})
	
	// Test template system
	t.Run("TemplateSystem", func(t *testing.T) {
		templatesDir := "../templates"
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			t.Skip("templates directory not found, skipping template test")
		}
		
		// Check for template files
		expectedTemplates := []string{
			"base.sh.tmpl",
			"rust.sh.tmpl",
			"go.sh.tmpl",
			"c.sh.tmpl",
			"python.sh.tmpl",
		}
		
		for _, template := range expectedTemplates {
			templatePath := filepath.Join(templatesDir, template)
			if _, err := os.Stat(templatePath); os.IsNotExist(err) {
				t.Errorf("Template not found: %s", template)
			}
		}
	})
}

// TestProjectStructure verifies the overall project structure
func TestProjectStructure(t *testing.T) {
	expectedDirectories := []string{
		"../cmd",
		"../internal",
		"../pkg",
		"../lib",
		"../scripts",
		"../templates",
		"../config",
		"../tests",
		"../docs",
	}
	
	for _, dir := range expectedDirectories {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Expected directory does not exist: %s", dir)
		}
	}
	
	// Check for essential files
	essentialFiles := []string{
		"../go.mod",
		"../Makefile",
		"../README.md",
		"../CLAUDE.md",
	}
	
	for _, file := range essentialFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Essential file does not exist: %s", file)
		}
	}
}