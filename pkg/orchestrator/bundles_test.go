package orchestrator

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBundleExpansion(t *testing.T) {
	// Create a test orchestrator
	o := &Orchestrator{
		repoDir: "../..", // Adjust path as needed for tests
		bundleConfig: &BundleConfiguration{
			SchemaVersion: "1.0",
			Bundles: []BundleConfig{
				{
					Name:        "minimal",
					Description: "Minimal bundle",
					Category:    "core",
					Tools:       []string{"fd", "ripgrep", "fzf"},
				},
				{
					Name:            "essential",
					Description:     "Essential bundle",
					Category:        "core",
					Tools:           []string{"bat", "eza"},
					IncludesBundles: []string{"minimal"},
				},
				{
					Name:            "developer",
					Description:     "Developer bundle",
					Category:        "dev",
					Tools:           []string{"starship", "delta"},
					IncludesBundles: []string{"essential"},
				},
			},
		},
	}

	tests := []struct {
		name         string
		bundleName   string
		expectedTools []string
		shouldError  bool
	}{
		{
			name:         "expand minimal bundle",
			bundleName:   "minimal",
			expectedTools: []string{"fd", "ripgrep", "fzf"},
			shouldError:  false,
		},
		{
			name:         "expand essential bundle with includes",
			bundleName:   "essential",
			expectedTools: []string{"fd", "ripgrep", "fzf", "bat", "eza"},
			shouldError:  false,
		},
		{
			name:         "expand developer bundle with nested includes",
			bundleName:   "developer",
			expectedTools: []string{"fd", "ripgrep", "fzf", "bat", "eza", "starship", "delta"},
			shouldError:  false,
		},
		{
			name:         "non-existent bundle",
			bundleName:   "nonexistent",
			expectedTools: nil,
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[string]bool)
			tools, err := o.expandBundle(tt.bundleName, o.bundleConfig.Bundles, visited)

			if tt.shouldError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.shouldError {
				if len(tools) != len(tt.expectedTools) {
					t.Errorf("expected %d tools, got %d", len(tt.expectedTools), len(tools))
				}

				// Check that all expected tools are present
				toolMap := make(map[string]bool)
				for _, tool := range tools {
					toolMap[tool] = true
				}
				for _, expected := range tt.expectedTools {
					if !toolMap[expected] {
						t.Errorf("expected tool %s not found in expansion", expected)
					}
				}
			}
		})
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	o := &Orchestrator{
		bundleConfig: &BundleConfiguration{
			Bundles: []BundleConfig{
				{
					Name:            "bundle1",
					IncludesBundles: []string{"bundle2"},
				},
				{
					Name:            "bundle2",
					IncludesBundles: []string{"bundle3"},
				},
				{
					Name:            "bundle3",
					IncludesBundles: []string{"bundle1"}, // Creates circular dependency
				},
			},
		},
	}

	visited := make(map[string]bool)
	_, err := o.expandBundle("bundle1", o.bundleConfig.Bundles, visited)
	
	if err == nil {
		t.Error("expected circular dependency error but got none")
	}
	// Check that error contains circular dependency message
	if err != nil && !strings.Contains(err.Error(), "circular dependency detected in bundle: bundle1") {
		t.Errorf("expected circular dependency error, got: %v", err)
	}
}

func TestExpandBundlesAndTools(t *testing.T) {
	// Set up test orchestrator with bundles
	o := &Orchestrator{
		repoDir: filepath.Join("..", ".."),
		bundleConfig: &BundleConfiguration{
			Bundles: []BundleConfig{
				{
					Name:  "test-bundle",
					Tools: []string{"fd", "ripgrep"},
				},
			},
		},
	}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "only tools",
			input:    []string{"bat", "eza"},
			expected: []string{"bat", "eza"},
		},
		{
			name:     "only bundle",
			input:    []string{"test-bundle"},
			expected: []string{"fd", "ripgrep"},
		},
		{
			name:     "mix of tools and bundles",
			input:    []string{"bat", "test-bundle", "jq"},
			expected: []string{"bat", "fd", "ripgrep", "jq"},
		},
		{
			name:     "duplicate tools removed",
			input:    []string{"fd", "test-bundle"},
			expected: []string{"fd", "ripgrep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := o.expandBundlesAndTools(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}


			if len(result) != len(tt.expected) {
				t.Errorf("expected %d tools, got %d", len(tt.expected), len(result))
			}

			// Check all expected tools are present
			resultMap := make(map[string]bool)
			for _, tool := range result {
				resultMap[tool] = true
			}
			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("expected tool %s not found", expected)
				}
			}
		})
	}
}

func TestIsBundle(t *testing.T) {
	bundles := []BundleConfig{
		{Name: "essential"},
		{Name: "developer"},
	}

	o := &Orchestrator{}

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"existing bundle", "essential", true},
		{"another bundle", "developer", true},
		{"not a bundle", "fd", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.isBundle(tt.input, bundles)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}