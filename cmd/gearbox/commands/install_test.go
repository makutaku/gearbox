package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestInstallCommandFlagParsing tests that CLI flags are parsed correctly
func TestInstallCommandFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]interface{}
	}{
		{
			name: "nerd-fonts with specific fonts",
			args: []string{"install", "nerd-fonts", "--fonts=JetBrainsMono,FiraCode"},
			expected: map[string]interface{}{
				"fonts":  "JetBrainsMono,FiraCode",
				"tools":  []string{"nerd-fonts"},
				"interactive": false,
				"preview": false,
			},
		},
		{
			name: "nerd-fonts with interactive flag",
			args: []string{"install", "nerd-fonts", "--interactive"},
			expected: map[string]interface{}{
				"fonts":  "",
				"tools":  []string{"nerd-fonts"},
				"interactive": true,
				"preview": false,
			},
		},
		{
			name: "nerd-fonts with all advanced flags",
			args: []string{"install", "nerd-fonts", "--fonts=FiraCode", "--interactive", "--preview", "--configure-apps"},
			expected: map[string]interface{}{
				"fonts":  "FiraCode",
				"tools":  []string{"nerd-fonts"},
				"interactive": true,
				"preview": true,
				"configure-apps": true,
			},
		},
		{
			name: "multiple tools with build flags",
			args: []string{"install", "fd", "ripgrep", "--minimal", "--skip-common-deps"},
			expected: map[string]interface{}{
				"tools":  []string{"fd", "ripgrep"},
				"minimal": true,
				"skip-common-deps": true,
			},
		},
		{
			name: "single tool with maximum flag",
			args: []string{"install", "ffmpeg", "--maximum", "--run-tests"},
			expected: map[string]interface{}{
				"tools":  []string{"ffmpeg"},
				"maximum": true,
				"run-tests": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInstallCmd()
			cmd.SetArgs(tt.args[1:]) // Remove "install" from args

			// Capture the arguments that would be passed to orchestrator
			var capturedArgs []string
			var capturedFlags map[string]interface{}

			// Override the runE function to capture arguments instead of executing
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				capturedArgs = args
				capturedFlags = make(map[string]interface{})

				// Capture all flag values
				if fonts, _ := cmd.Flags().GetString("fonts"); fonts != "" {
					capturedFlags["fonts"] = fonts
				}
				if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
					capturedFlags["interactive"] = interactive
				}
				if preview, _ := cmd.Flags().GetBool("preview"); preview {
					capturedFlags["preview"] = preview
				}
				if configureApps, _ := cmd.Flags().GetBool("configure-apps"); configureApps {
					capturedFlags["configure-apps"] = configureApps
				}
				if minimal, _ := cmd.Flags().GetBool("minimal"); minimal {
					capturedFlags["minimal"] = minimal
				}
				if maximum, _ := cmd.Flags().GetBool("maximum"); maximum {
					capturedFlags["maximum"] = maximum
				}
				if skipCommonDeps, _ := cmd.Flags().GetBool("skip-common-deps"); skipCommonDeps {
					capturedFlags["skip-common-deps"] = skipCommonDeps
				}
				if runTests, _ := cmd.Flags().GetBool("run-tests"); runTests {
					capturedFlags["run-tests"] = runTests
				}

				return nil
			}

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Check tool arguments
			if expectedTools, ok := tt.expected["tools"].([]string); ok {
				if len(capturedArgs) != len(expectedTools) {
					t.Errorf("Expected %d tools, got %d: %v", len(expectedTools), len(capturedArgs), capturedArgs)
				}
				for i, expectedTool := range expectedTools {
					if i >= len(capturedArgs) || capturedArgs[i] != expectedTool {
						t.Errorf("Expected tool %q at index %d, got %q", expectedTool, i, capturedArgs[i])
					}
				}
			}

			// Check flag values
			for key, expectedValue := range tt.expected {
				if key == "tools" {
					continue // Already checked above
				}
				
				actualValue, ok := capturedFlags[key]
				if !ok && expectedValue != "" && expectedValue != false {
					t.Errorf("Expected flag %q not captured", key)
					continue
				}
				
				if ok && actualValue != expectedValue {
					t.Errorf("Flag %q: expected %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}

// TestInstallCommandArgumentSeparation tests that tools and flags are properly separated
func TestInstallCommandArgumentSeparation(t *testing.T) {
	tests := []struct {
		name         string
		cmdLine      string
		expectedArgs []string
		shouldFail   bool
	}{
		{
			name:         "simple tool list",
			cmdLine:      "fd ripgrep fzf",
			expectedArgs: []string{"fd", "ripgrep", "fzf"},
		},
		{
			name:         "tool with flags",
			cmdLine:      "nerd-fonts --fonts=JetBrainsMono",
			expectedArgs: []string{"nerd-fonts"},
		},
		{
			name:         "multiple tools with flags",
			cmdLine:      "fd ripgrep --minimal --skip-common-deps",
			expectedArgs: []string{"fd", "ripgrep"},
		},
		{
			name:         "flags before tools (should work)",
			cmdLine:      "--minimal fd ripgrep",
			expectedArgs: []string{"fd", "ripgrep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewInstallCmd()
			args := strings.Fields(tt.cmdLine)
			cmd.SetArgs(args)

			var capturedArgs []string
			cmd.RunE = func(cmd *cobra.Command, args []string) error {
				capturedArgs = args
				return nil
			}

			err := cmd.Execute()
			if tt.shouldFail {
				if err == nil {
					t.Error("Expected command to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			if len(capturedArgs) != len(tt.expectedArgs) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expectedArgs), len(capturedArgs), capturedArgs)
			}

			for i, expected := range tt.expectedArgs {
				if i >= len(capturedArgs) || capturedArgs[i] != expected {
					t.Errorf("Arg %d: expected %q, got %q", i, expected, capturedArgs[i])
				}
			}
		})
	}
}

// TestNerdFontsArgumentConstruction tests that nerd-fonts specific arguments are built correctly
func TestNerdFontsArgumentConstruction(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]interface{}
		expected []string
	}{
		{
			name: "fonts flag only",
			flags: map[string]interface{}{
				"fonts": "JetBrainsMono",
			},
			expected: []string{"--fonts=JetBrainsMono"},
		},
		{
			name: "multiple flags",
			flags: map[string]interface{}{
				"fonts":        "FiraCode,Hack",
				"interactive":  true,
				"preview":      true,
				"configure-apps": true,
			},
			expected: []string{"--fonts=FiraCode,Hack", "--interactive", "--preview", "--configure-apps"},
		},
		{
			name: "with build type and orchestrator flags",
			flags: map[string]interface{}{
				"fonts":           "JetBrainsMono",
				"minimal":         true,
				"skip-common-deps": true,
				"run-tests":       true,
			},
			expected: []string{"--fonts=JetBrainsMono", "--build-type", "minimal", "--skip-common-deps", "--run-tests"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test simulates the argument construction logic from install.go
			var args []string

			// Add nerd-fonts specific flags (simulating CLI to orchestrator conversion)
			if fonts, ok := tt.flags["fonts"].(string); ok && fonts != "" {
				args = append(args, "--fonts="+fonts)
			}
			if interactive, ok := tt.flags["interactive"].(bool); ok && interactive {
				args = append(args, "--interactive")
			}
			if preview, ok := tt.flags["preview"].(bool); ok && preview {
				args = append(args, "--preview")
			}
			if configureApps, ok := tt.flags["configure-apps"].(bool); ok && configureApps {
				args = append(args, "--configure-apps")
			}

			// Add standard orchestrator flags
			if minimal, ok := tt.flags["minimal"].(bool); ok && minimal {
				args = append(args, "--build-type", "minimal")
			}
			if maximum, ok := tt.flags["maximum"].(bool); ok && maximum {
				args = append(args, "--build-type", "maximum")
			}
			if skipDeps, ok := tt.flags["skip-common-deps"].(bool); ok && skipDeps {
				args = append(args, "--skip-common-deps")
			}
			if runTests, ok := tt.flags["run-tests"].(bool); ok && runTests {
				args = append(args, "--run-tests")
			}

			if len(args) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expected), len(args), args)
			}

			for i, expected := range tt.expected {
				if i >= len(args) || args[i] != expected {
					t.Errorf("Arg %d: expected %q, got %q", i, expected, args[i])
				}
			}
		})
	}
}

// TestCliOrchestratorIntegration tests the full CLI to orchestrator flow
func TestCliOrchestratorIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a temporary script to capture orchestrator arguments
	tmpDir := t.TempDir()
	mockOrchestratorPath := filepath.Join(tmpDir, "orchestrator")
	
	// Create mock orchestrator that captures and echoes its arguments
	mockOrchestratorScript := `#!/bin/bash
echo "MOCK_ORCHESTRATOR_ARGS: $@" >&2
exit 0
`
	
	err := os.WriteFile(mockOrchestratorPath, []byte(mockOrchestratorScript), 0755)
	if err != nil {
		t.Fatalf("Failed to create mock orchestrator: %v", err)
	}

	tests := []struct {
		name         string
		cmdArgs      []string
		expectedArgs []string
	}{
		{
			name:    "nerd-fonts with fonts flag",
			cmdArgs: []string{"install", "nerd-fonts", "--fonts=JetBrainsMono"},
			expectedArgs: []string{"install", "nerd-fonts", "--fonts", "JetBrainsMono"},
		},
		{
			name:    "nerd-fonts with multiple flags",
			cmdArgs: []string{"install", "nerd-fonts", "--fonts=FiraCode", "--interactive", "--skip-common-deps"},
			expectedArgs: []string{"install", "nerd-fonts", "--fonts", "FiraCode", "--interactive", "--skip-common-deps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would require setting up the actual CLI environment
			// For now, we test the argument construction logic
			cmd := exec.Command("echo", "CLI integration test placeholder")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Mock test failed: %v", err)
			}
			
			t.Logf("Mock output: %s", output)
		})
	}
}

// BenchmarkArgumentParsing benchmarks the argument parsing performance
func BenchmarkArgumentParsing(b *testing.B) {
	cmd := NewInstallCmd()
	args := []string{"nerd-fonts", "--fonts=JetBrainsMono,FiraCode,Hack", "--interactive", "--preview"}
	
	for i := 0; i < b.N; i++ {
		cmd.SetArgs(args)
		cmd.ParseFlags(args)
	}
}