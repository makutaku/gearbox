package orchestrator

import (
	"os/exec"
	"strings"
)

// isToolInstalled checks if a tool is installed
func isToolInstalled(tool ToolConfig) bool {
	// Special handling for nerd-fonts
	if tool.Name == "nerd-fonts" {
		return isNerdFontsInstalled()
	}
	
	_, err := exec.LookPath(tool.BinaryName)
	return err == nil
}

// verifyTool verifies a tool installation by running its test command
func verifyTool(tool ToolConfig) bool {
	if !isToolInstalled(tool) {
		return false
	}

	// Special handling for nerd-fonts
	if tool.Name == "nerd-fonts" {
		return isNerdFontsInstalled()
	}

	if tool.TestCommand == "" {
		return true // If no test command, just check if binary exists
	}

	// Parse and execute test command
	parts := strings.Fields(tool.TestCommand)
	if len(parts) == 0 {
		return true
	}

	// Use binary_name instead of the first part of test_command for tools with different binary names
	binaryName := tool.BinaryName
	if binaryName == "" {
		binaryName = tool.Name // Fallback to tool name if binary_name not specified
	}
	
	// For test commands, use ALL parts as arguments (don't skip the first one)
	// Most test commands are just "--version", not "tool_name --version"
	cmdArgs := parts

	cmd := exec.Command(binaryName, cmdArgs...)
	err := cmd.Run()
	return err == nil
}

// getToolVersion gets the version of an installed tool
func getToolVersion(tool ToolConfig) string {
	// Special handling for nerd-fonts
	if tool.Name == "nerd-fonts" {
		return getNerdFontsVersion()
	}
	if tool.TestCommand == "" {
		return "installed"
	}

	parts := strings.Fields(tool.TestCommand)
	if len(parts) == 0 {
		return "installed"
	}

	// Use binary_name instead of the first part of test_command for tools with different binary names
	// This handles cases like bottom (binary: btm), ripgrep (binary: rg), etc.
	binaryName := tool.BinaryName
	if binaryName == "" {
		binaryName = tool.Name // Fallback to tool name if binary_name not specified
	}
	
	// For test commands, use ALL parts as arguments (don't skip the first one)
	// Most test commands are just "--version", not "tool_name --version"
	cmdArgs := parts
	
	cmd := exec.Command(binaryName, cmdArgs...)
	output, err := cmd.Output()
	if err != nil {
		return "installed"
	}

	// Extract version information from output
	return extractVersionFromOutput(string(output))
}

