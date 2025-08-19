package tests

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNerdFontsArgumentFlow tests the complete argument flow from CLI to script
func TestNerdFontsArgumentFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Could not find project root: %v", err)
	}

	tests := []struct {
		name           string
		cliArgs        []string
		expectedScript []string
		shouldFail     bool
	}{
		{
			name:    "nerd-fonts with specific font",
			cliArgs: []string{"install", "nerd-fonts", "--fonts=JetBrainsMono", "--skip-common-deps", "--dry-run"},
			expectedScript: []string{"--fonts=JetBrainsMono", "--skip-deps", "--config-only"},
		},
		{
			name:    "nerd-fonts interactive mode",
			cliArgs: []string{"install", "nerd-fonts", "--interactive", "--skip-common-deps", "--dry-run"},
			expectedScript: []string{"--interactive", "--skip-deps", "--config-only"},
		},
		{
			name:    "nerd-fonts with multiple flags",
			cliArgs: []string{"install", "nerd-fonts", "--fonts=FiraCode,Hack", "--preview", "--configure-apps", "--skip-common-deps", "--dry-run"},
			expectedScript: []string{"--fonts=FiraCode,Hack", "--preview", "--configure-apps", "--skip-deps", "--config-only"},
		},
		{
			name:    "nerd-fonts with build type",
			cliArgs: []string{"install", "nerd-fonts", "--fonts=SourceCodePro", "--minimal", "--skip-common-deps", "--dry-run"},
			expectedScript: []string{"--fonts=SourceCodePro", "--minimal", "--skip-deps", "--config-only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock script that captures its arguments
			tmpDir := t.TempDir()
			mockScript := filepath.Join(tmpDir, "install-nerd-fonts.sh")
			
			scriptContent := `#!/bin/bash
echo "MOCK_SCRIPT_ARGS_COUNT: $#" >&2
echo "MOCK_SCRIPT_ARGS: $*" >&2
for i in "$@"; do
    echo "MOCK_SCRIPT_ARG: $i" >&2
done
exit 0
`
			err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
			if err != nil {
				t.Fatalf("Failed to create mock script: %v", err)
			}

			// Temporarily replace the real script with our mock
			originalScript := filepath.Join(projectRoot, "scripts", "install-nerd-fonts.sh")
			backupScript := originalScript + ".backup"
			
			// Backup original script
			if _, err := os.Stat(originalScript); err == nil {
				err = os.Rename(originalScript, backupScript)
				if err != nil {
					t.Fatalf("Failed to backup original script: %v", err)
				}
				defer func() {
					os.Rename(backupScript, originalScript)
				}()
			}
			
			// Copy mock script to original location
			err = copyFile(mockScript, originalScript)
			if err != nil {
				t.Fatalf("Failed to copy mock script: %v", err)
			}

			// Run the CLI command
			gearboxPath := filepath.Join(projectRoot, "gearbox")
			cmd := exec.Command(gearboxPath, tt.cliArgs...)
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Dir = projectRoot

			err = cmd.Run()
			if tt.shouldFail {
				if err == nil {
					t.Error("Expected command to fail, but it succeeded")
				}
				return
			}

			stderrStr := stderr.String()
			
			// Parse the mock script output to extract arguments
			argsRegex := regexp.MustCompile(`MOCK_SCRIPT_ARGS: (.*)`)
			matches := argsRegex.FindStringSubmatch(stderrStr)
			if len(matches) < 2 {
				t.Fatalf("Could not find script arguments in output: %s", stderrStr)
			}
			
			actualArgs := strings.Fields(matches[1])
			
			// Verify arguments match expectations
			if len(actualArgs) != len(tt.expectedScript) {
				t.Errorf("Expected %d script args, got %d: %v", len(tt.expectedScript), len(actualArgs), actualArgs)
			}
			
			for i, expected := range tt.expectedScript {
				if i >= len(actualArgs) || actualArgs[i] != expected {
					t.Errorf("Script arg %d: expected %q, got %q", i, expected, actualArgs[i])
				}
			}
			
			// Verify no mysterious extra arguments
			for i, arg := range actualArgs {
				if arg == "2" || arg == "1" || arg == "0" {
					t.Errorf("Found suspicious numeric argument %q at position %d", arg, i)
				}
			}
		})
	}
}

// TestArgumentContaminationPrevention tests that the fix prevents argument contamination
func TestArgumentContaminationPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test specifically checks for the "2" argument contamination issue
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Could not find project root: %v", err)
	}

	// Create a monitoring script that fails if it receives unexpected arguments
	tmpDir := t.TempDir()
	monitorScript := filepath.Join(tmpDir, "install-nerd-fonts.sh")
	
	scriptContent := `#!/bin/bash
echo "MONITOR_SCRIPT_RECEIVED_ARGS: $*" >&2

# Check for suspicious arguments
for arg in "$@"; do
    case "$arg" in
        "2"|"1"|"0")
            echo "ERROR: Received suspicious numeric argument: $arg" >&2
            exit 1
            ;;
        "")
            echo "ERROR: Received empty argument" >&2
            exit 1
            ;;
    esac
done

# Check for correct nerd-fonts arguments
valid_pattern="^(--fonts=|--interactive|--preview|--configure-apps|--minimal|--standard|--maximum|--skip-deps|--run-tests|--force|--config-only).*"
for arg in "$@"; do
    if [[ ! "$arg" =~ $valid_pattern ]]; then
        echo "ERROR: Received invalid argument: $arg" >&2
        exit 1
    fi
done

echo "SUCCESS: All arguments are valid" >&2
exit 0
`
	
	err = os.WriteFile(monitorScript, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create monitor script: %v", err)
	}

	// Temporarily replace the real script
	originalScript := filepath.Join(projectRoot, "scripts", "install-nerd-fonts.sh")
	backupScript := originalScript + ".backup"
	
	// Backup and replace
	if _, err := os.Stat(originalScript); err == nil {
		err = os.Rename(originalScript, backupScript)
		if err != nil {
			t.Fatalf("Failed to backup original script: %v", err)
		}
		defer os.Rename(backupScript, originalScript)
	}
	
	err = copyFile(monitorScript, originalScript)
	if err != nil {
		t.Fatalf("Failed to copy monitor script: %v", err)
	}

	// Test various command combinations that previously caused issues
	testCommands := [][]string{
		{"install", "nerd-fonts", "--fonts=JetBrainsMono", "--skip-common-deps", "--dry-run"},
		{"install", "nerd-fonts", "--interactive", "--skip-common-deps", "--dry-run"},
		{"install", "nerd-fonts", "--fonts=FiraCode,Hack", "--preview", "--skip-common-deps", "--dry-run"},
		{"install", "nerd-fonts", "--minimal", "--skip-common-deps", "--dry-run"},
	}

	gearboxPath := filepath.Join(projectRoot, "gearbox")
	
	for i, cmdArgs := range testCommands {
		t.Run(fmt.Sprintf("contamination_test_%d", i), func(t *testing.T) {
			cmd := exec.Command(gearboxPath, cmdArgs...)
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Dir = projectRoot

			err := cmd.Run()
			
			stderrStr := stderr.String()
			
			// The monitor script should succeed if no contamination occurs
			if strings.Contains(stderrStr, "ERROR: Received suspicious numeric argument") {
				t.Errorf("Argument contamination detected: %s", stderrStr)
			}
			
			if strings.Contains(stderrStr, "SUCCESS: All arguments are valid") {
				t.Logf("✅ No argument contamination detected for command: %v", cmdArgs)
			} else {
				t.Errorf("Monitor script did not report success for command: %v. Output: %s", cmdArgs, stderrStr)
			}
		})
	}
}

// TestNerdFontsScriptArgumentParsing tests the script's argument parsing directly
func TestNerdFontsScriptArgumentParsing(t *testing.T) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Could not find project root: %v", err)
	}

	scriptPath := filepath.Join(projectRoot, "scripts", "install-nerd-fonts.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Skipf("Script not found: %s", scriptPath)
	}

	tests := []struct {
		name      string
		args      []string
		shouldFail bool
		expectMsg  string
	}{
		{
			name: "valid fonts argument",
			args: []string{"--fonts=JetBrainsMono", "--skip-deps", "--config-only"},
			shouldFail: false,
		},
		{
			name: "valid interactive argument",
			args: []string{"--interactive", "--skip-deps", "--config-only"},
			shouldFail: false,
		},
		{
			name: "invalid numeric argument",
			args: []string{"--fonts=JetBrainsMono", "--skip-deps", "2"},
			shouldFail: true,
			expectMsg: "Unknown option: 2",
		},
		{
			name: "invalid argument",
			args: []string{"--fonts=JetBrainsMono", "--invalid-flag"},
			shouldFail: true,
			expectMsg: "Unknown option: --invalid-flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("bash", scriptPath)
			cmd.Args = append(cmd.Args, tt.args...)
			
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			
			if tt.shouldFail {
				if err == nil {
					t.Error("Expected script to fail, but it succeeded")
				}
				if tt.expectMsg != "" && !strings.Contains(stderr.String(), tt.expectMsg) {
					t.Errorf("Expected error message %q, got: %s", tt.expectMsg, stderr.String())
				}
			} else {
				if err != nil {
					t.Errorf("Script failed unexpectedly: %v. Stderr: %s", err, stderr.String())
				}
			}
		})
	}
}

// Helper functions

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	
	return "", fmt.Errorf("could not find project root (no go.mod found)")
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0755)
}

import "fmt"