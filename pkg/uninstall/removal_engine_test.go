package uninstall

import (
	"os"
	"strings"
	"testing"

	"gearbox/pkg/manifest"
)

func TestNewRemovalEngine(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tests := []struct {
		name        string
		safetyLevel SafetyLevel
	}{
		{
			name:        "conservative safety",
			safetyLevel: SafetyConservative,
		},
		{
			name:        "standard safety",
			safetyLevel: SafetyStandard,
		},
		{
			name:        "aggressive safety",
			safetyLevel: SafetyAggressive,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewRemovalEngine(tt.safetyLevel)
			if err != nil {
				t.Errorf("NewRemovalEngine() error = %v", err)
				return
			}
			
			if engine == nil {
				t.Fatal("NewRemovalEngine() should return an engine instance")
			}
			
			if engine.safetyLevel != tt.safetyLevel {
				t.Errorf("NewRemovalEngine() safetyLevel = %v, want %v", engine.safetyLevel, tt.safetyLevel)
			}
			
			if engine.tracker == nil {
				t.Error("NewRemovalEngine() should initialize tracker")
			}
		})
	}
}

func TestRemovalEngine_PlanRemoval_NonExistentTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	targets := []string{"nonexistent-tool"}
	options := RemovalOptions{}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	if plan == nil {
		t.Fatal("PlanRemoval() should return a plan")
	}
	
	// Should have warnings but no removal actions
	if len(plan.ToRemove) != 0 {
		t.Error("PlanRemoval() should not plan removal of non-existent tool")
	}
	
	if len(plan.Warnings) == 0 {
		t.Error("PlanRemoval() should warn about non-existent tool")
	}
	
	// Check warning content
	found := false
	for _, warning := range plan.Warnings {
		if warning.Target == "nonexistent-tool" && strings.Contains(warning.Message, "not tracked") {
			found = true
			break
		}
	}
	if !found {
		t.Error("PlanRemoval() should warn that tool is not tracked")
	}
}

func TestRemovalEngine_PlanRemoval_PreExistingTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track a pre-existing tool
	if err := engine.tracker.TrackPreExisting("git", "/usr/bin/git", "2.34.1"); err != nil {
		t.Fatalf("Failed to track pre-existing tool: %v", err)
	}
	
	targets := []string{"git"}
	options := RemovalOptions{}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should not plan removal of pre-existing tool
	if len(plan.ToRemove) != 0 {
		t.Error("PlanRemoval() should not plan removal of pre-existing tool")
	}
	
	if len(plan.ToKeep) == 0 {
		t.Error("PlanRemoval() should plan to keep pre-existing tool")
	}
	
	// Check keep reason
	found := false
	for _, keep := range plan.ToKeep {
		if keep.Target == "git" {
			for _, reason := range keep.Reasons {
				if strings.Contains(reason, "pre-existing") {
					found = true
					break
				}
			}
			break
		}
	}
	if !found {
		t.Error("PlanRemoval() should mention pre-existing as reason to keep")
	}
}

func TestRemovalEngine_PlanRemoval_ToolWithDependents(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track rust as a dependency
	rustConfig := manifest.TrackingConfig{
		Method:  manifest.MethodSystemPackage,
		Version: "stable",
	}
	if err := engine.tracker.TrackInstallation("rust", rustConfig); err != nil {
		t.Fatalf("Failed to track rust: %v", err)
	}
	
	// Track ripgrep that depends on rust
	ripgrepConfig := manifest.TrackingConfig{
		Method:       manifest.MethodCargoInstall,
		Version:      "13.0.0",
		Dependencies: []string{"rust"},
	}
	if err := engine.tracker.TrackInstallation("ripgrep", ripgrepConfig); err != nil {
		t.Fatalf("Failed to track ripgrep: %v", err)
	}
	
	targets := []string{"rust"}
	options := RemovalOptions{Force: false}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should not plan removal due to dependents
	if len(plan.ToRemove) != 0 {
		t.Error("PlanRemoval() should not plan removal of tool with dependents")
	}
	
	if len(plan.ToKeep) == 0 {
		t.Error("PlanRemoval() should plan to keep tool with dependents")
	}
	
	// Check keep reason mentions dependents
	found := false
	for _, keep := range plan.ToKeep {
		if keep.Target == "rust" {
			for _, reason := range keep.Reasons {
				if strings.Contains(reason, "ripgrep") {
					found = true
					break
				}
			}
			break
		}
	}
	if !found {
		t.Error("PlanRemoval() should mention dependents as reason to keep")
	}
}

func TestRemovalEngine_PlanRemoval_ForceRemoval(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track rust as a dependency
	rustConfig := manifest.TrackingConfig{
		Method:  manifest.MethodSystemPackage,
		Version: "stable",
	}
	if err := engine.tracker.TrackInstallation("rust", rustConfig); err != nil {
		t.Fatalf("Failed to track rust: %v", err)
	}
	
	// Track ripgrep that depends on rust
	ripgrepConfig := manifest.TrackingConfig{
		Method:       manifest.MethodCargoInstall,
		Version:      "13.0.0",
		Dependencies: []string{"rust"},
	}
	if err := engine.tracker.TrackInstallation("ripgrep", ripgrepConfig); err != nil {
		t.Fatalf("Failed to track ripgrep: %v", err)
	}
	
	targets := []string{"rust"}
	options := RemovalOptions{Force: true}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should plan removal when forced
	if len(plan.ToRemove) != 1 {
		t.Errorf("PlanRemoval() should plan forced removal, got %d removals", len(plan.ToRemove))
	}
	
	if plan.ToRemove[0].Target != "rust" {
		t.Errorf("PlanRemoval() removal target = %s, want rust", plan.ToRemove[0].Target)
	}
	
	if plan.ToRemove[0].IsSafe {
		t.Error("PlanRemoval() forced removal should be marked as unsafe")
	}
	
	// Should have warning about forced removal
	foundWarning := false
	for _, warning := range plan.Warnings {
		if warning.Target == "rust" && strings.Contains(warning.Message, "Forcing removal") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("PlanRemoval() should warn about forced removal")
	}
}

func TestRemovalEngine_PlanRemoval_StandaloneTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track a standalone tool
	config := manifest.TrackingConfig{
		Method:      manifest.MethodSourceBuild,
		Version:     "1.0.0",
		BinaryPaths: []string{"/usr/local/bin/standalone-tool"},
		BuildDir:    "/tmp/build/standalone-tool",
	}
	if err := engine.tracker.TrackInstallation("standalone-tool", config); err != nil {
		t.Fatalf("Failed to track standalone tool: %v", err)
	}
	
	targets := []string{"standalone-tool"}
	options := RemovalOptions{}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should plan removal of standalone tool
	if len(plan.ToRemove) != 1 {
		t.Errorf("PlanRemoval() should plan removal of standalone tool, got %d removals", len(plan.ToRemove))
	}
	
	removal := plan.ToRemove[0]
	if removal.Target != "standalone-tool" {
		t.Errorf("PlanRemoval() removal target = %s, want standalone-tool", removal.Target)
	}
	
	if removal.Method != RemovalSourceBuild {
		t.Errorf("PlanRemoval() removal method = %s, want %s", removal.Method, RemovalSourceBuild)
	}
	
	if !removal.IsSafe {
		t.Error("PlanRemoval() standalone tool removal should be safe")
	}
	
	// Check that binary and build paths are included
	expectedPaths := []string{"/usr/local/bin/standalone-tool", "/tmp/build/standalone-tool"}
	if len(removal.Paths) != len(expectedPaths) {
		t.Errorf("PlanRemoval() paths count = %d, want %d", len(removal.Paths), len(expectedPaths))
	}
	
	for _, expectedPath := range expectedPaths {
		found := false
		for _, path := range removal.Paths {
			if path == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PlanRemoval() should include path %s", expectedPath)
		}
	}
}

func TestRemovalEngine_PlanRemoval_WithConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track a tool with config files
	config := manifest.TrackingConfig{
		Method:      manifest.MethodSourceBuild,
		Version:     "1.0.0",
		BinaryPaths: []string{"/usr/local/bin/tool-with-config"},
		ConfigFiles: []string{"/home/user/.config/tool.conf", "/etc/tool/config.ini"},
	}
	if err := engine.tracker.TrackInstallation("tool-with-config", config); err != nil {
		t.Fatalf("Failed to track tool with config: %v", err)
	}
	
	targets := []string{"tool-with-config"}
	options := RemovalOptions{RemoveConfig: true}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	if len(plan.ToRemove) != 1 {
		t.Errorf("PlanRemoval() should plan removal, got %d removals", len(plan.ToRemove))
		return
	}
	
	removal := plan.ToRemove[0]
	
	// Check that config files are included when RemoveConfig is true
	expectedPaths := []string{
		"/usr/local/bin/tool-with-config",
		"/home/user/.config/tool.conf",
		"/etc/tool/config.ini",
	}
	
	if len(removal.Paths) != len(expectedPaths) {
		t.Errorf("PlanRemoval() paths count = %d, want %d", len(removal.Paths), len(expectedPaths))
	}
	
	for _, expectedPath := range expectedPaths {
		found := false
		for _, path := range removal.Paths {
			if path == expectedPath {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("PlanRemoval() should include config path %s", expectedPath)
		}
	}
}

func TestRemovalEngine_PlanRemoval_Bundle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track some tools
	tools := []string{"ripgrep", "fd", "fzf"}
	for _, tool := range tools {
		config := manifest.TrackingConfig{
			Method:  manifest.MethodSourceBuild,
			Version: "1.0.0",
		}
		if err := engine.tracker.TrackInstallation(tool, config); err != nil {
			t.Fatalf("Failed to track tool %s: %v", tool, err)
		}
	}
	
	// Track bundle
	bundleName := "essential-tools"
	if err := engine.tracker.TrackBundle(bundleName, tools, true); err != nil {
		t.Fatalf("Failed to track bundle: %v", err)
	}
	
	targets := []string{bundleName + "_bundle"}
	options := RemovalOptions{RemoveBundleContents: false}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should plan bundle removal
	if len(plan.ToRemove) != 1 {
		t.Errorf("PlanRemoval() should plan bundle removal, got %d removals", len(plan.ToRemove))
		return
	}
	
	removal := plan.ToRemove[0]
	if removal.Method != RemovalBundle {
		t.Errorf("PlanRemoval() bundle removal method = %s, want %s", removal.Method, RemovalBundle)
	}
	
	// Should warn about contained tools
	foundWarning := false
	for _, warning := range plan.Warnings {
		if strings.Contains(warning.Message, "Bundle contains") && strings.Contains(warning.Message, "tools that will remain") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Error("PlanRemoval() should warn about tools remaining after bundle removal")
	}
}

func TestRemovalEngine_PlanRemoval_BundleWithContents(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track some tools
	tools := []string{"ripgrep", "fd"}
	for _, tool := range tools {
		config := manifest.TrackingConfig{
			Method:  manifest.MethodSourceBuild,
			Version: "1.0.0",
		}
		if err := engine.tracker.TrackInstallation(tool, config); err != nil {
			t.Fatalf("Failed to track tool %s: %v", tool, err)
		}
	}
	
	// Track bundle
	bundleName := "essential-tools"
	if err := engine.tracker.TrackBundle(bundleName, tools, true); err != nil {
		t.Fatalf("Failed to track bundle: %v", err)
	}
	
	targets := []string{bundleName + "_bundle"}
	options := RemovalOptions{RemoveBundleContents: true}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should plan removal of bundle + tools
	expectedRemovals := 3 // bundle + 2 tools
	if len(plan.ToRemove) != expectedRemovals {
		t.Errorf("PlanRemoval() should plan removal of bundle and tools, got %d removals", len(plan.ToRemove))
	}
	
	// Check that bundle removal is included
	foundBundle := false
	foundTools := 0
	
	for _, removal := range plan.ToRemove {
		if removal.Target == bundleName+"_bundle" && removal.Method == RemovalBundle {
			foundBundle = true
		} else if removal.Method == RemovalSourceBuild {
			foundTools++
		}
	}
	
	if !foundBundle {
		t.Error("PlanRemoval() should include bundle removal")
	}
	
	if foundTools != len(tools) {
		t.Errorf("PlanRemoval() should plan removal of %d tools, got %d", len(tools), foundTools)
	}
}

func TestRemovalEngine_PlanRemoval_DependencyAnalysis(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track dependency
	depConfig := manifest.TrackingConfig{
		Method:  manifest.MethodSystemPackage,
		Version: "stable",
	}
	if err := engine.tracker.TrackInstallation("shared-dep", depConfig); err != nil {
		t.Fatalf("Failed to track dependency: %v", err)
	}
	
	// Track two tools that depend on it
	tool1Config := manifest.TrackingConfig{
		Method:       manifest.MethodSourceBuild,
		Version:      "1.0.0",
		Dependencies: []string{"shared-dep"},
	}
	if err := engine.tracker.TrackInstallation("tool1", tool1Config); err != nil {
		t.Fatalf("Failed to track tool1: %v", err)
	}
	
	tool2Config := manifest.TrackingConfig{
		Method:       manifest.MethodSourceBuild,
		Version:      "1.0.0",
		Dependencies: []string{"shared-dep"},
	}
	if err := engine.tracker.TrackInstallation("tool2", tool2Config); err != nil {
		t.Fatalf("Failed to track tool2: %v", err)
	}
	
	// Remove only tool1
	targets := []string{"tool1"}
	options := RemovalOptions{Cascade: true}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should plan removal of tool1 only
	if len(plan.ToRemove) != 1 {
		t.Errorf("PlanRemoval() should plan removal of 1 tool, got %d", len(plan.ToRemove))
	}
	
	// Should analyze dependency but preserve it (tool2 still needs it)
	if len(plan.Dependencies) == 0 {
		t.Error("PlanRemoval() should analyze dependencies")
	}
	
	foundDep := false
	for _, depAction := range plan.Dependencies {
		if depAction.Dependency == "shared-dep" {
			foundDep = true
			if depAction.Action != "preserve" {
				t.Errorf("PlanRemoval() should preserve shared dependency, got action: %s", depAction.Action)
			}
			if !strings.Contains(depAction.Reason, "tool2") {
				t.Error("PlanRemoval() dependency reason should mention remaining dependent")
			}
			break
		}
	}
	if !foundDep {
		t.Error("PlanRemoval() should analyze shared dependency")
	}
}

func TestRemovalEngine_PlanRemoval_CascadeRemoval(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track dependency
	depConfig := manifest.TrackingConfig{
		Method:  manifest.MethodSystemPackage,
		Version: "stable",
	}
	if err := engine.tracker.TrackInstallation("unused-dep", depConfig); err != nil {
		t.Fatalf("Failed to track dependency: %v", err)
	}
	
	// Track tool that depends on it
	toolConfig := manifest.TrackingConfig{
		Method:       manifest.MethodSourceBuild,
		Version:      "1.0.0",
		Dependencies: []string{"unused-dep"},
	}
	if err := engine.tracker.TrackInstallation("only-tool", toolConfig); err != nil {
		t.Fatalf("Failed to track tool: %v", err)
	}
	
	// Remove the tool with cascade
	targets := []string{"only-tool"}
	options := RemovalOptions{Cascade: true}
	
	plan, err := engine.PlanRemoval(targets, options)
	if err != nil {
		t.Errorf("PlanRemoval() error = %v", err)
		return
	}
	
	// Should plan removal of tool
	if len(plan.ToRemove) != 1 {
		t.Errorf("PlanRemoval() should plan removal of 1 tool, got %d", len(plan.ToRemove))
	}
	
	// Should plan removal of dependency (cascade)
	foundDepRemoval := false
	for _, depAction := range plan.Dependencies {
		if depAction.Dependency == "unused-dep" && depAction.Action == RemovalManualDelete {
			foundDepRemoval = true
			if !strings.Contains(depAction.Reason, "cascade") {
				t.Error("PlanRemoval() dependency removal reason should mention cascade")
			}
			break
		}
	}
	if !foundDepRemoval {
		t.Error("PlanRemoval() should plan cascade removal of unused dependency")
	}
}

func TestRemovalEngine_getRemovalMethod(t *testing.T) {
	engine := &RemovalEngine{}
	
	tests := []struct {
		name           string
		installMethod  manifest.InstallationMethod
		expectedMethod RemovalMethod
	}{
		{
			name:           "source build",
			installMethod:  manifest.MethodSourceBuild,
			expectedMethod: RemovalSourceBuild,
		},
		{
			name:           "cargo install",
			installMethod:  manifest.MethodCargoInstall,
			expectedMethod: RemovalCargoInstall,
		},
		{
			name:           "go install",
			installMethod:  manifest.MethodGoInstall,
			expectedMethod: RemovalGoInstall,
		},
		{
			name:           "system package",
			installMethod:  manifest.MethodSystemPackage,
			expectedMethod: RemovalSystemPackage,
		},
		{
			name:           "pipx",
			installMethod:  manifest.MethodPipx,
			expectedMethod: RemovalPipx,
		},
		{
			name:           "npm global",
			installMethod:  manifest.MethodNpmGlobal,
			expectedMethod: RemovalNpmGlobal,
		},
		{
			name:           "manual download",
			installMethod:  manifest.MethodManualDownload,
			expectedMethod: RemovalManualDelete,
		},
		{
			name:           "bundle",
			installMethod:  manifest.MethodBundle,
			expectedMethod: RemovalBundle,
		},
		{
			name:           "pre-existing",
			installMethod:  manifest.MethodPreExisting,
			expectedMethod: RemovalPreExisting,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := engine.getRemovalMethod(tt.installMethod)
			if method != tt.expectedMethod {
				t.Errorf("getRemovalMethod() = %s, want %s", method, tt.expectedMethod)
			}
		})
	}
}

func TestRemovalEngine_ValidatePlan(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tests := []struct {
		name           string
		safetyLevel    SafetyLevel
		plan           *RemovalPlan
		expectedWarnings int
		checkMessages  []string
	}{
		{
			name:        "safe plan",
			safetyLevel: SafetyStandard,
			plan: &RemovalPlan{
				ToRemove: []RemovalAction{
					{Target: "safe-tool", IsSafe: true},
				},
				Dependencies: []DependencyAction{
					{Dependency: "dep1", Action: "preserve", Affected: []string{"tool1"}},
				},
				Summary: RemovalSummary{WillRemove: 1, WillKeep: 0},
			},
			expectedWarnings: 0,
		},
		{
			name:        "forced removal",
			safetyLevel: SafetyStandard,
			plan: &RemovalPlan{
				ToRemove: []RemovalAction{
					{Target: "unsafe-tool", IsSafe: false},
				},
				Summary: RemovalSummary{WillRemove: 1, WillKeep: 0},
			},
			expectedWarnings: 1,
			checkMessages:    []string{"Forced removal may break"},
		},
		{
			name:        "shared dependency removal",
			safetyLevel: SafetyStandard,
			plan: &RemovalPlan{
				ToRemove: []RemovalAction{},
				Dependencies: []DependencyAction{
					{
						Dependency: "shared-dep",
						Action:     RemovalManualDelete,
						Affected:   []string{"tool1", "tool2", "tool3"},
					},
				},
				Summary: RemovalSummary{WillRemove: 0, WillKeep: 0},
			},
			expectedWarnings: 1,
			checkMessages:    []string{"Removing shared dependency"},
		},
		{
			name:        "conservative mode",
			safetyLevel: SafetyConservative,
			plan: &RemovalPlan{
				ToRemove: []RemovalAction{
					{Target: "any-tool", IsSafe: true},
				},
				Summary: RemovalSummary{WillRemove: 1, WillKeep: 0},
			},
			expectedWarnings: 1,
			checkMessages:    []string{"Conservative mode"},
		},
		{
			name:        "aggressive mode",
			safetyLevel: SafetyAggressive,
			plan: &RemovalPlan{
				ToRemove: []RemovalAction{},
				ToKeep: []KeepReason{
					{Target: "tool1"},
					{Target: "tool2"},
					{Target: "tool3"},
				},
				Summary: RemovalSummary{WillRemove: 0, WillKeep: 3},
			},
			expectedWarnings: 1,
			checkMessages:    []string{"Aggressive mode", "consider using --force"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewRemovalEngine(tt.safetyLevel)
			if err != nil {
				t.Fatalf("NewRemovalEngine() error = %v", err)
			}
			
			warnings := engine.ValidatePlan(tt.plan)
			
			if len(warnings) != tt.expectedWarnings {
				t.Errorf("ValidatePlan() warnings count = %d, want %d", len(warnings), tt.expectedWarnings)
			}
			
			// Check specific warning messages
			for _, expectedMsg := range tt.checkMessages {
				found := false
				for _, warning := range warnings {
					if strings.Contains(warning.Message, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidatePlan() should contain warning with message: %s", expectedMsg)
				}
			}
		})
	}
}

func TestGetRemovalMethodDescription(t *testing.T) {
	tests := []struct {
		name     string
		method   RemovalMethod
		contains string
	}{
		{
			name:     "source build",
			method:   RemovalSourceBuild,
			contains: "source installation",
		},
		{
			name:     "cargo install",
			method:   RemovalCargoInstall,
			contains: "cargo uninstall",
		},
		{
			name:     "go install",
			method:   RemovalGoInstall,
			contains: "Go tool",
		},
		{
			name:     "system package",
			method:   RemovalSystemPackage,
			contains: "system package manager",
		},
		{
			name:     "pipx",
			method:   RemovalPipx,
			contains: "pipx uninstall",
		},
		{
			name:     "npm global",
			method:   RemovalNpmGlobal,
			contains: "npm uninstall -g",
		},
		{
			name:     "manual delete",
			method:   RemovalManualDelete,
			contains: "Manually delete",
		},
		{
			name:     "bundle",
			method:   RemovalBundle,
			contains: "bundle tracking",
		},
		{
			name:     "pre-existing",
			method:   RemovalPreExisting,
			contains: "Preserve pre-existing",
		},
		{
			name:     "unknown method",
			method:   RemovalMethod("unknown"),
			contains: "Unknown removal method",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			description := GetRemovalMethodDescription(tt.method)
			
			if !strings.Contains(description, tt.contains) {
				t.Errorf("GetRemovalMethodDescription() = %s, should contain %s", description, tt.contains)
			}
		})
	}
}

func TestRemovalEngine_isBundle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		t.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track a bundle
	tools := []string{"tool1", "tool2"}
	for _, tool := range tools {
		config := manifest.TrackingConfig{
			Method:  manifest.MethodSourceBuild,
			Version: "1.0.0",
		}
		if err := engine.tracker.TrackInstallation(tool, config); err != nil {
			t.Fatalf("Failed to track tool %s: %v", tool, err)
		}
	}
	
	bundleName := "test-bundle"
	if err := engine.tracker.TrackBundle(bundleName, tools, true); err != nil {
		t.Fatalf("Failed to track bundle: %v", err)
	}
	
	tests := []struct {
		name     string
		target   string
		expected bool
	}{
		{
			name:     "tracked bundle",
			target:   bundleName + "_bundle",
			expected: true,
		},
		{
			name:     "bundle suffix",
			target:   "unknown-bundle",
			expected: true,
		},
		{
			name:     "bundle dash suffix",
			target:   "unknown-bundle",
			expected: true,
		},
		{
			name:     "regular tool",
			target:   "tool1",
			expected: false,
		},
		{
			name:     "non-existent",
			target:   "nonexistent",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.isBundle(tt.target)
			if result != tt.expected {
				t.Errorf("isBundle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSafetyLevel_Constants(t *testing.T) {
	// Verify safety level constants are defined
	if SafetyConservative < 0 {
		t.Error("SafetyConservative should be >= 0")
	}
	
	if SafetyStandard <= SafetyConservative {
		t.Error("SafetyStandard should be > SafetyConservative")
	}
	
	if SafetyAggressive <= SafetyStandard {
		t.Error("SafetyAggressive should be > SafetyStandard")
	}
}

func TestRemovalMethod_Constants(t *testing.T) {
	// Verify removal method constants are defined and non-empty
	methods := []RemovalMethod{
		RemovalSourceBuild,
		RemovalCargoInstall,
		RemovalGoInstall,
		RemovalSystemPackage,
		RemovalPipx,
		RemovalNpmGlobal,
		RemovalManualDelete,
		RemovalBundle,
		RemovalPreExisting,
	}
	
	for _, method := range methods {
		if string(method) == "" {
			t.Errorf("Removal method should not be empty: %v", method)
		}
	}
}

func TestRemovalOptions_Struct(t *testing.T) {
	options := RemovalOptions{
		Force:                true,
		Cascade:              true,
		RemoveConfig:         true,
		DryRun:               true,
		Backup:               true,
		BackupSuffix:         "test",
		RemoveBundleContents: true,
	}
	
	// Verify all fields can be set and retrieved
	if !options.Force {
		t.Error("RemovalOptions Force field not working")
	}
	
	if !options.Cascade {
		t.Error("RemovalOptions Cascade field not working")
	}
	
	if !options.RemoveConfig {
		t.Error("RemovalOptions RemoveConfig field not working")
	}
	
	if !options.DryRun {
		t.Error("RemovalOptions DryRun field not working")
	}
	
	if !options.Backup {
		t.Error("RemovalOptions Backup field not working")
	}
	
	if options.BackupSuffix != "test" {
		t.Error("RemovalOptions BackupSuffix field not working")
	}
	
	if !options.RemoveBundleContents {
		t.Error("RemovalOptions RemoveBundleContents field not working")
	}
}

// Benchmark tests
func BenchmarkRemovalEngine_PlanRemoval(b *testing.B) {
	tempDir := b.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	engine, err := NewRemovalEngine(SafetyStandard)
	if err != nil {
		b.Fatalf("NewRemovalEngine() error = %v", err)
	}
	
	// Track some tools for benchmarking
	for i := 0; i < 10; i++ {
		toolName := "tool-" + string(rune(i))
		config := manifest.TrackingConfig{
			Method:  manifest.MethodSourceBuild,
			Version: "1.0.0",
		}
		_ = engine.tracker.TrackInstallation(toolName, config)
	}
	
	targets := []string{"tool-0", "tool-1", "tool-2"}
	options := RemovalOptions{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.PlanRemoval(targets, options)
	}
}

func BenchmarkGetRemovalMethodDescription(b *testing.B) {
	methods := []RemovalMethod{
		RemovalSourceBuild,
		RemovalCargoInstall,
		RemovalGoInstall,
		RemovalSystemPackage,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		method := methods[i%len(methods)]
		_ = GetRemovalMethodDescription(method)
	}
}