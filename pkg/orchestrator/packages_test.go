package orchestrator

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	// This test will only pass on Linux systems with a package manager
	if runtime.GOOS != "linux" {
		t.Skip("Package manager detection only supported on Linux")
	}
	
	pm, err := detectPackageManager()
	
	// We expect either a valid package manager or an error
	if err != nil && pm != nil {
		t.Error("Should not return both error and package manager")
	}
	
	// If we found a package manager, it should have the required fields
	if pm != nil {
		if pm.Name == "" {
			t.Error("Package manager name should not be empty")
		}
		if len(pm.InstallCmd) == 0 {
			t.Error("Package manager should have install command")
		}
		if !pm.Available {
			t.Error("Detected package manager should be marked as available")
		}
	}
}

func TestGetSystemPackagesForManager(t *testing.T) {
	bundle := BundleConfig{
		Name:           "test-bundle",
		SystemPackages: []string{"generic-pkg"},
		PackageManagers: map[string][]string{
			"apt": {"apt-specific-pkg"},
			"yum": {"yum-specific-pkg"},
		},
	}
	
	tests := []struct {
		name     string
		manager  string
		expected []string
	}{
		{
			name:     "apt specific packages",
			manager:  "apt",
			expected: []string{"apt-specific-pkg"},
		},
		{
			name:     "yum specific packages", 
			manager:  "yum",
			expected: []string{"yum-specific-pkg"},
		},
		{
			name:     "fallback to generic packages",
			manager:  "dnf",
			expected: []string{"generic-pkg"},
		},
		{
			name:     "unknown manager falls back to generic",
			manager:  "unknown",
			expected: []string{"generic-pkg"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bundle.getSystemPackagesForManager(tt.manager)
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d packages, got %d", len(tt.expected), len(result))
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("expected package %s, got %s", expected, result[i])
				}
			}
		})
	}
}

func TestExpandSystemPackages(t *testing.T) {
	o := &Orchestrator{}
	
	bundles := []BundleConfig{
		{
			Name:           "base",
			SystemPackages: []string{"base-pkg1", "base-pkg2"},
		},
		{
			Name:            "extended",
			SystemPackages:  []string{"ext-pkg1"},
			IncludesBundles: []string{"base"},
		},
		{
			Name: "manager-specific",
			PackageManagers: map[string][]string{
				"apt": {"apt-pkg1", "apt-pkg2"},
				"yum": {"yum-pkg1"},
			},
		},
	}
	
	tests := []struct {
		name        string
		bundleName  string
		manager     string
		expected    []string
		shouldError bool
	}{
		{
			name:       "base bundle only",
			bundleName: "base",
			manager:    "apt",
			expected:   []string{"base-pkg1", "base-pkg2"},
		},
		{
			name:       "extended bundle includes base",
			bundleName: "extended",
			manager:    "apt", 
			expected:   []string{"base-pkg1", "base-pkg2", "ext-pkg1"},
		},
		{
			name:       "manager-specific packages for apt",
			bundleName: "manager-specific",
			manager:    "apt",
			expected:   []string{"apt-pkg1", "apt-pkg2"},
		},
		{
			name:       "manager-specific packages for yum",
			bundleName: "manager-specific", 
			manager:    "yum",
			expected:   []string{"yum-pkg1"},
		},
		{
			name:        "non-existent bundle",
			bundleName:  "nonexistent",
			manager:     "apt",
			shouldError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[string]bool)
			result, err := o.expandSystemPackages(tt.bundleName, bundles, visited, tt.manager)
			
			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			
			if !tt.shouldError {
				if len(result) != len(tt.expected) {
					t.Errorf("expected %d packages, got %d", len(tt.expected), len(result))
				}
				
				// Check that all expected packages are present
				pkgMap := make(map[string]bool)
				for _, pkg := range result {
					pkgMap[pkg] = true
				}
				for _, expected := range tt.expected {
					if !pkgMap[expected] {
						t.Errorf("expected package %s not found in result", expected)
					}
				}
			}
		})
	}
}

func TestSystemPackageCircularDependency(t *testing.T) {
	o := &Orchestrator{}
	
	bundles := []BundleConfig{
		{
			Name:            "bundle1",
			SystemPackages:  []string{"pkg1"},
			IncludesBundles: []string{"bundle2"},
		},
		{
			Name:            "bundle2", 
			SystemPackages:  []string{"pkg2"},
			IncludesBundles: []string{"bundle1"},
		},
	}
	
	visited := make(map[string]bool)
	_, err := o.expandSystemPackages("bundle1", bundles, visited, "apt")
	
	if err == nil {
		t.Error("expected circular dependency error but got none")
	}
	if err != nil && !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("expected circular dependency error, got: %v", err)
	}
}

func TestPackageDeduplication(t *testing.T) {
	o := &Orchestrator{}
	
	bundles := []BundleConfig{
		{
			Name:           "base",
			SystemPackages: []string{"shared-pkg", "base-pkg"},
		},
		{
			Name:            "extended",
			SystemPackages:  []string{"shared-pkg", "ext-pkg"}, // shared-pkg appears in both
			IncludesBundles: []string{"base"},
		},
	}
	
	visited := make(map[string]bool)
	result, err := o.expandSystemPackages("extended", bundles, visited, "apt")
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	// Count occurrences of shared-pkg
	count := 0
	for _, pkg := range result {
		if pkg == "shared-pkg" {
			count++
		}
	}
	
	if count != 1 {
		t.Errorf("expected shared-pkg to appear once, but appeared %d times", count)
	}
	
	// Should have 3 unique packages: shared-pkg, base-pkg, ext-pkg
	if len(result) != 3 {
		t.Errorf("expected 3 unique packages, got %d", len(result))
	}
}