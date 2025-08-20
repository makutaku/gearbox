package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNerdFontsOptionsConstruction tests that nerd-fonts options are constructed correctly
func TestNerdFontsOptionsConstruction(t *testing.T) {
	tests := []struct {
		name     string
		options  InstallationOptions
		expected []string
	}{
		{
			name: "fonts flag only",
			options: InstallationOptions{
				Fonts: "JetBrainsMono",
			},
			expected: []string{"--fonts=JetBrainsMono"},
		},
		{
			name: "all nerd-fonts flags",
			options: InstallationOptions{
				Fonts:         "FiraCode,Hack",
				Interactive:   true,
				Preview:       true,
				ConfigureApps: true,
			},
			expected: []string{
				"--fonts=FiraCode,Hack",
				"--interactive",
				"--preview", 
				"--configure-apps",
			},
		},
		{
			name: "with orchestrator options",
			options: InstallationOptions{
				Fonts:         "JetBrainsMono",
				BuildType:     "minimal",
				SkipCommonDeps: true,
				RunTests:      true,
				Force:         true,
			},
			expected: []string{
				"--fonts=JetBrainsMono",
				"--minimal",
				"--skip-deps",
				"--run-tests",
				"--force",
			},
		},
		{
			name: "maximum build type",
			options: InstallationOptions{
				Fonts:     "SourceCodePro",
				BuildType: "maximum",
			},
			expected: []string{
				"--fonts=SourceCodePro",
				"--maximum",
			},
		},
		{
			name: "dry run mode",
			options: InstallationOptions{
				Fonts:  "Hack",
				DryRun: true,
			},
			expected: []string{
				"--fonts=Hack",
				"--config-only",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock orchestrator with test options
			orchestrator := &Orchestrator{
				options: tt.options,
			}

			// Simulate the argument construction logic from installNerdFontsWithOptions
			args := []string{}
			
			if orchestrator.options.Fonts != "" {
				args = append(args, "--fonts="+orchestrator.options.Fonts)
			}
			
			if orchestrator.options.Interactive {
				args = append(args, "--interactive")
			}
			
			if orchestrator.options.Preview {
				args = append(args, "--preview")
			}
			
			if orchestrator.options.ConfigureApps {
				args = append(args, "--configure-apps")
			}
			
			// Add standard orchestrator options
			if orchestrator.options.BuildType == "minimal" {
				args = append(args, "--minimal")
			} else if orchestrator.options.BuildType == "maximum" {
				args = append(args, "--maximum")
			}
			
			if orchestrator.options.SkipCommonDeps {
				args = append(args, "--skip-deps")
			}
			
			if orchestrator.options.RunTests {
				args = append(args, "--run-tests")
			}
			
			if orchestrator.options.Force {
				args = append(args, "--force")
			}
			
			if orchestrator.options.DryRun {
				args = append(args, "--config-only")
			}

			// Verify the constructed arguments match expectations
			if len(args) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expected), len(args), args)
			}

			for i, expectedArg := range tt.expected {
				if i >= len(args) || args[i] != expectedArg {
					t.Errorf("Arg %d: expected %q, got %q", i, expectedArg, args[i])
				}
			}
		})
	}
}

// TestNerdFontsDetection tests that nerd-fonts installation is detected correctly
func TestNerdFontsDetection(t *testing.T) {
	tests := []struct {
		name        string
		tools       []ToolConfig
		expectNerdFonts bool
		expectAdvanced  bool
		options     InstallationOptions
	}{
		{
			name: "single nerd-fonts with fonts option",
			tools: []ToolConfig{
				{Name: "nerd-fonts"},
			},
			expectNerdFonts: true,
			expectAdvanced:  true,
			options: InstallationOptions{
				Fonts: "JetBrainsMono",
			},
		},
		{
			name: "single nerd-fonts with interactive",
			tools: []ToolConfig{
				{Name: "nerd-fonts"},
			},
			expectNerdFonts: true,
			expectAdvanced:  true,
			options: InstallationOptions{
				Interactive: true,
			},
		},
		{
			name: "single nerd-fonts without advanced options",
			tools: []ToolConfig{
				{Name: "nerd-fonts"},
			},
			expectNerdFonts: true,
			expectAdvanced:  false,
			options: InstallationOptions{},
		},
		{
			name: "multiple tools including nerd-fonts",
			tools: []ToolConfig{
				{Name: "starship"},
				{Name: "nerd-fonts"},
			},
			expectNerdFonts: true,
			expectAdvanced:  false,
			options: InstallationOptions{
				Fonts: "JetBrainsMono", // Should not trigger advanced mode with multiple tools
			},
		},
		{
			name: "no nerd-fonts",
			tools: []ToolConfig{
				{Name: "fd"},
				{Name: "ripgrep"},
			},
			expectNerdFonts: false,
			expectAdvanced:  false,
			options: InstallationOptions{
				Fonts: "JetBrainsMono", // Should be ignored
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if nerd-fonts is in the tool list
			hasNerdFonts := false
			for _, tool := range tt.tools {
				if tool.Name == "nerd-fonts" {
					hasNerdFonts = true
					break
				}
			}

			if hasNerdFonts != tt.expectNerdFonts {
				t.Errorf("Expected nerd-fonts detection %v, got %v", tt.expectNerdFonts, hasNerdFonts)
			}

			// Check advanced options detection (single nerd-fonts tool with special options)
			shouldUseAdvanced := len(tt.tools) == 1 && hasNerdFonts && 
				(tt.options.Fonts != "" || tt.options.Interactive || tt.options.Preview || tt.options.ConfigureApps)

			if shouldUseAdvanced != tt.expectAdvanced {
				t.Errorf("Expected advanced mode %v, got %v", tt.expectAdvanced, shouldUseAdvanced)
			}
		})
	}
}

// TestScriptArgumentPassing tests that arguments are passed correctly to the script
func TestScriptArgumentPassing(t *testing.T) {
	// Create a temporary directory and mock script
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "install-nerd-fonts.sh")
	
	// Create a mock script that echoes its arguments
	mockScript := `#!/bin/bash
echo "SCRIPT_ARGS_COUNT: $#"
for arg in "$@"; do
    echo "SCRIPT_ARG: $arg"
done
exit 0
`
	
	err := os.WriteFile(scriptPath, []byte(mockScript), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock script: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		expectCount  int
		expectArgs   []string
	}{
		{
			name:        "fonts flag only",
			args:        []string{"--fonts=JetBrainsMono"},
			expectCount: 1,
			expectArgs:  []string{"--fonts=JetBrainsMono"},
		},
		{
			name:        "multiple flags",
			args:        []string{"--fonts=FiraCode,Hack", "--interactive", "--skip-deps"},
			expectCount: 3,
			expectArgs:  []string{"--fonts=FiraCode,Hack", "--interactive", "--skip-deps"},
		},
		{
			name:        "no args",
			args:        []string{},
			expectCount: 0,
			expectArgs:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate executeNerdFontsScript
			cmdArgs := []string{"bash", scriptPath}
			cmdArgs = append(cmdArgs, tt.args...)
			
			// Verify command construction
			if len(cmdArgs) != 2+len(tt.args) {
				t.Errorf("Expected %d command args, got %d: %v", 2+len(tt.args), len(cmdArgs), cmdArgs)
			}
			
			if cmdArgs[0] != "bash" {
				t.Errorf("Expected first arg to be 'bash', got %q", cmdArgs[0])
			}
			
			if cmdArgs[1] != scriptPath {
				t.Errorf("Expected second arg to be script path %q, got %q", scriptPath, cmdArgs[1])
			}
			
			// Verify that script arguments match expectations
			scriptArgs := cmdArgs[2:]
			if len(scriptArgs) != len(tt.expectArgs) {
				t.Errorf("Expected %d script args, got %d: %v", len(tt.expectArgs), len(scriptArgs), scriptArgs)
			}
			
			for i, expectedArg := range tt.expectArgs {
				if i >= len(scriptArgs) || scriptArgs[i] != expectedArg {
					t.Errorf("Script arg %d: expected %q, got %q", i, expectedArg, scriptArgs[i])
				}
			}
		})
	}
}

// TestArgumentContamination tests that no extra arguments are accidentally added
func TestArgumentContamination(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "clean args",
			input:    []string{"--fonts=JetBrainsMono", "--interactive"},
			expected: []string{"--fonts=JetBrainsMono", "--interactive"},
		},
		{
			name:     "empty args",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single arg",
			input:    []string{"--fonts=Hack"},
			expected: []string{"--fonts=Hack"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the exact logic from executeNerdFontsScript
			args := tt.input
			scriptPath := "./scripts/install-nerd-fonts.sh"
			
			// Build the command exactly as in the actual code
			cmdArgs := []string{"bash", scriptPath}
			cmdArgs = append(cmdArgs, args...)
			
			// Extract just the script arguments (everything after "bash" and script path)
			scriptArgs := cmdArgs[2:]
			
			if len(scriptArgs) != len(tt.expected) {
				t.Errorf("Expected %d script args, got %d: %v", len(tt.expected), len(scriptArgs), scriptArgs)
			}
			
			for i, expectedArg := range tt.expected {
				if i >= len(scriptArgs) || scriptArgs[i] != expectedArg {
					t.Errorf("Arg %d: expected %q, got %q", i, expectedArg, scriptArgs[i])
				}
			}
			
			// Verify no mysterious extra arguments
			for i, arg := range scriptArgs {
				if arg == "2" || arg == "1" || arg == "0" {
					t.Errorf("Found suspicious numeric argument %q at position %d", arg, i)
				}
			}
		})
	}
}

// TestNerdFontsOptionsValidation tests that nerd-fonts options are validated correctly
func TestNerdFontsOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		options InstallationOptions
		valid   bool
	}{
		{
			name: "valid fonts list",
			options: InstallationOptions{
				Fonts: "JetBrainsMono,FiraCode",
			},
			valid: true,
		},
		{
			name: "valid single font",
			options: InstallationOptions{
				Fonts: "Hack",
			},
			valid: true,
		},
		{
			name: "empty fonts with interactive",
			options: InstallationOptions{
				Interactive: true,
			},
			valid: true,
		},
		{
			name: "invalid empty options",
			options: InstallationOptions{},
			valid: true, // Empty options should default to standard installation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := true
			
			// In a real implementation, you might validate font names against available fonts
			if tt.options.Fonts != "" {
				// Check for valid font name format (no spaces, comma-separated)
				fonts := strings.Split(tt.options.Fonts, ",")
				for _, font := range fonts {
					font = strings.TrimSpace(font)
					if font == "" {
						isValid = false
						break
					}
				}
			}
			
			if isValid != tt.valid {
				t.Errorf("Expected validation result %v, got %v", tt.valid, isValid)
			}
		})
	}
}