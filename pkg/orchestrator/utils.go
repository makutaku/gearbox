package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// isToolInConfig checks if a tool exists in the configuration
func isToolInConfig(configMgr *ConfigManager, toolName string) bool {
	config := configMgr.GetConfig()
	for _, tool := range config.Tools {
		if tool.Name == toolName {
			return true
		}
	}
	return false
}

// extractVersionFromOutput extracts version information from command output
func extractVersionFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Try various patterns to extract version
		patterns := []string{
			// Standard version patterns
			`(\d+\.\d+\.\d+(?:-[\w\.-]+)?)`,
			`(\d+\.\d+\.\d+)`,
			`(\d+\.\d+)`,
			`(\d+)`,
			// Tool-specific patterns
			`version\s*:?\s*([v]?\d+\.\d+\.\d+(?:-[\w\.-]+)?)`,
			`([v]?\d+\.\d+\.\d+(?:-[\w\.-]+)?)`,
			// Git-style versions
			`([a-f0-9]{7,40})`,
		}
		
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				version := matches[1]
				// Clean up common prefixes
				version = strings.TrimPrefix(version, "v")
				version = strings.TrimPrefix(version, "Version")
				version = strings.TrimSpace(version)
				if version != "" {
					return version
				}
			}
		}
		
		// If no pattern matches, try to extract from verbose output
		if version := extractVersionFromVerboseOutput(line); version != "" {
			return version
		}
	}
	
	return "unknown"
}

// extractVersionFromVerboseOutput handles verbose output patterns
func extractVersionFromVerboseOutput(line string) string {
	// Handle specific tools with verbose output
	verbosePatterns := map[string][]string{
		"go": {
			`go version go(\d+\.\d+\.?\d*)\s`,
			`go(\d+\.\d+\.?\d*)`,
		},
		"rust": {
			`rustc (\d+\.\d+\.\d+)`,
			`cargo (\d+\.\d+\.\d+)`,
		},
		"node": {
			`v(\d+\.\d+\.\d+)`,
		},
		"python": {
			`Python (\d+\.\d+\.\d+)`,
		},
	}
	
	for _, patterns := range verbosePatterns {
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				return matches[1]
			}
		}
	}
	
	return ""
}

// loadBundleConfig loads bundle configurations from JSON
func loadBundleConfig(configPath string) (map[string][]string, error) {
	bundlePath := filepath.Join(filepath.Dir(configPath), "bundles.json")
	
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read bundle config: %w", err)
	}
	
	var bundles map[string][]string
	if err := json.Unmarshal(data, &bundles); err != nil {
		return nil, fmt.Errorf("failed to parse bundle config: %w", err)
	}
	
	return bundles, nil
}

// getOptimalJobs calculates optimal number of parallel jobs based on system resources
func getOptimalJobs() int {
	numCPU := runtime.NumCPU()
	
	// Use 75% of available CPUs, minimum 1, maximum 8
	jobs := max(1, min(8, numCPU*3/4))
	
	return jobs
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sortToolsByDependencies sorts tools to ensure dependencies are installed first
func sortToolsByDependencies(tools []ToolConfig) []ToolConfig {
	sorted := make([]ToolConfig, 0, len(tools))
	processed := make(map[string]bool)
	
	var addTool func(ToolConfig)
	addTool = func(tool ToolConfig) {
		if processed[tool.Name] {
			return
		}
		
		// Add dependencies first
		for _, dep := range tool.Dependencies {
			for _, t := range tools {
				if t.Name == dep && !processed[t.Name] {
					addTool(t)
				}
			}
		}
		
		sorted = append(sorted, tool)
		processed[tool.Name] = true
	}
	
	for _, tool := range tools {
		addTool(tool)
	}
	
	return sorted
}

// formatList formats a string slice into a readable list
func formatList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}
	if len(items) == 2 {
		return items[0] + " and " + items[1]
	}
	
	return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
}

// createProgressBar creates a new progress bar with consistent styling
func createProgressBar(max int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

// isValidBuildType checks if the given build type is valid
func isValidBuildType(buildType string) bool {
	validTypes := []string{"minimal", "standard", "maximum"}
	for _, valid := range validTypes {
		if buildType == valid {
			return true
		}
	}
	return false
}

// expandBundleNames expands bundle names to their constituent tools
func expandBundleNames(names []string, bundles map[string][]string) []string {
	var expanded []string
	seen := make(map[string]bool)
	
	for _, name := range names {
		if bundleTools, isBundle := bundles[name]; isBundle {
			for _, tool := range bundleTools {
				if !seen[tool] {
					expanded = append(expanded, tool)
					seen[tool] = true
				}
			}
		} else {
			if !seen[name] {
				expanded = append(expanded, name)
				seen[name] = true
			}
		}
	}
	
	return expanded
}