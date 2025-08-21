package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTracker(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Errorf("NewTracker() error = %v", err)
		return
	}
	
	if tracker == nil {
		t.Fatal("NewTracker() should return a tracker instance")
	}
	
	if tracker.manager == nil {
		t.Error("NewTracker() should initialize manager")
	}
	
	if tracker.manifest == nil {
		t.Error("NewTracker() should initialize manifest")
	}
}

func TestTracker_TrackInstallation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	config := TrackingConfig{
		Method:              MethodSourceBuild,
		Version:             "1.0.0",
		BinaryPaths:         []string{"/usr/local/bin/test-tool"},
		BuildDir:            "/tmp/build/test-tool",
		SourceRepo:          "https://github.com/example/test-tool",
		Dependencies:        []string{"rust", "git"},
		InstalledByBundle:   "",
		UserRequested:       true,
		InstallationContext: []string{"manual"},
		ConfigFiles:         []string{"/home/user/.config/test-tool.conf"},
		SystemPackages:      []string{"build-essential"},
	}
	
	// Track installation
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Errorf("TrackInstallation() error = %v", err)
		return
	}
	
	// Verify installation was tracked
	if !tracker.IsInstalled("test-tool") {
		t.Error("TrackInstallation() should mark tool as installed")
	}
	
	// Verify installation record
	record, exists := tracker.GetInstallation("test-tool")
	if !exists {
		t.Error("TrackInstallation() should create installation record")
		return
	}
	
	if record.Method != config.Method {
		t.Errorf("Installation method = %v, want %v", record.Method, config.Method)
	}
	
	if record.Version != config.Version {
		t.Errorf("Installation version = %v, want %v", record.Version, config.Version)
	}
	
	if len(record.BinaryPaths) != len(config.BinaryPaths) {
		t.Errorf("Installation binary paths count = %d, want %d", len(record.BinaryPaths), len(config.BinaryPaths))
	}
	
	if record.BuildDir != config.BuildDir {
		t.Errorf("Installation build dir = %v, want %v", record.BuildDir, config.BuildDir)
	}
	
	if record.SourceRepo != config.SourceRepo {
		t.Errorf("Installation source repo = %v, want %v", record.SourceRepo, config.SourceRepo)
	}
	
	if len(record.Dependencies) != len(config.Dependencies) {
		t.Errorf("Installation dependencies count = %d, want %d", len(record.Dependencies), len(config.Dependencies))
	}
	
	if record.UserRequested != config.UserRequested {
		t.Errorf("Installation user requested = %v, want %v", record.UserRequested, config.UserRequested)
	}
	
	// Verify dependencies were tracked
	for _, dep := range config.Dependencies {
		dependents := tracker.GetDependents(dep)
		if len(dependents) == 0 {
			t.Errorf("Dependency %s should have dependents", dep)
		}
		
		found := false
		for _, dependent := range dependents {
			if dependent == "test-tool" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Dependency %s should list test-tool as dependent", dep)
		}
	}
}

func TestTracker_TrackInstallation_AlreadyTracked(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	config := TrackingConfig{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		BinaryPaths: []string{"/usr/local/bin/test-tool"},
	}
	
	// Track installation first time
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("First TrackInstallation() error = %v", err)
	}
	
	// Try to track again - should fail
	err = tracker.TrackInstallation("test-tool", config)
	if err == nil {
		t.Error("TrackInstallation() should error when tool is already tracked")
	}
	
	if !strings.Contains(err.Error(), "already tracked") {
		t.Errorf("TrackInstallation() error should mention already tracked, got: %v", err)
	}
}

func TestTracker_TrackBundle(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Track some tools first
	tools := []string{"ripgrep", "fd", "fzf"}
	for _, tool := range tools {
		config := TrackingConfig{
			Method:        MethodSourceBuild,
			Version:       "1.0.0",
			UserRequested: false,
		}
		if err := tracker.TrackInstallation(tool, config); err != nil {
			t.Fatalf("Failed to track tool %s: %v", tool, err)
		}
	}
	
	// Track bundle
	bundleName := "essential-tools"
	if err := tracker.TrackBundle(bundleName, tools, true); err != nil {
		t.Errorf("TrackBundle() error = %v", err)
		return
	}
	
	// Verify bundle was tracked
	bundleKey := bundleName + "_bundle"
	if !tracker.IsInstalled(bundleKey) {
		t.Error("TrackBundle() should track bundle installation")
	}
	
	// Verify bundle record
	bundleRecord, exists := tracker.GetInstallation(bundleKey)
	if !exists {
		t.Error("TrackBundle() should create bundle record")
		return
	}
	
	if bundleRecord.Method != MethodBundle {
		t.Errorf("Bundle method = %v, want %v", bundleRecord.Method, MethodBundle)
	}
	
	if bundleRecord.UserRequested != true {
		t.Error("Bundle should be marked as user requested")
	}
	
	if len(bundleRecord.Dependencies) != len(tools) {
		t.Errorf("Bundle dependencies count = %d, want %d", len(bundleRecord.Dependencies), len(tools))
	}
	
	// Verify tools were updated to reference bundle
	for _, tool := range tools {
		record, exists := tracker.GetInstallation(tool)
		if !exists {
			t.Errorf("Tool %s should still exist", tool)
			continue
		}
		
		if record.InstalledByBundle != bundleName {
			t.Errorf("Tool %s should reference bundle %s, got %s", tool, bundleName, record.InstalledByBundle)
		}
		
		// Check installation context
		bundleContext := "bundle:" + bundleName
		found := false
		for _, context := range record.InstallationContext {
			if context == bundleContext {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Tool %s should have bundle context %s", tool, bundleContext)
		}
	}
}

func TestTracker_IsInstalled(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Should not be installed initially
	if tracker.IsInstalled("test-tool") {
		t.Error("IsInstalled() should return false for non-tracked tool")
	}
	
	// Track installation
	config := TrackingConfig{
		Method:  MethodSourceBuild,
		Version: "1.0.0",
	}
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	// Should be installed now
	if !tracker.IsInstalled("test-tool") {
		t.Error("IsInstalled() should return true for tracked tool")
	}
}

func TestTracker_GetInstallation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Test non-existent installation
	_, exists := tracker.GetInstallation("nonexistent")
	if exists {
		t.Error("GetInstallation() should return false for non-tracked tool")
	}
	
	// Track installation
	config := TrackingConfig{
		Method:  MethodCargoInstall,
		Version: "2.0.0",
	}
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	// Test existing installation
	record, exists := tracker.GetInstallation("test-tool")
	if !exists {
		t.Error("GetInstallation() should return true for tracked tool")
		return
	}
	
	if record.Method != config.Method {
		t.Errorf("GetInstallation() method = %v, want %v", record.Method, config.Method)
	}
	
	if record.Version != config.Version {
		t.Errorf("GetInstallation() version = %v, want %v", record.Version, config.Version)
	}
}

func TestTracker_GetAllInstallations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Should be empty initially
	installations := tracker.GetAllInstallations()
	if len(installations) != 0 {
		t.Error("GetAllInstallations() should return empty map initially")
	}
	
	// Track multiple installations
	tools := []string{"tool1", "tool2", "tool3"}
	for _, tool := range tools {
		config := TrackingConfig{
			Method:  MethodSourceBuild,
			Version: "1.0.0",
		}
		if err := tracker.TrackInstallation(tool, config); err != nil {
			t.Fatalf("Failed to track tool %s: %v", tool, err)
		}
	}
	
	// Should return all installations
	installations = tracker.GetAllInstallations()
	if len(installations) != len(tools) {
		t.Errorf("GetAllInstallations() count = %d, want %d", len(installations), len(tools))
	}
	
	// Verify all tools are present
	for _, tool := range tools {
		if _, exists := installations[tool]; !exists {
			t.Errorf("GetAllInstallations() should include %s", tool)
		}
	}
}

func TestTracker_DetectPreExisting(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Test with a tool that likely exists on the system (sh)
	exists, path, err := tracker.DetectPreExisting("sh", "sh")
	if err != nil {
		t.Errorf("DetectPreExisting() error = %v", err)
		return
	}
	
	if exists {
		if path == "" {
			t.Error("DetectPreExisting() should return path when tool exists")
		}
	}
	
	// Test with a tool that doesn't exist
	exists, path, err = tracker.DetectPreExisting("nonexistent-tool-12345", "nonexistent-tool-12345")
	if err != nil {
		t.Errorf("DetectPreExisting() error = %v", err)
		return
	}
	
	if exists {
		t.Error("DetectPreExisting() should return false for non-existent tool")
	}
	
	if path != "" {
		t.Error("DetectPreExisting() should return empty path for non-existent tool")
	}
	
	// Test with already tracked tool
	config := TrackingConfig{
		Method:  MethodSourceBuild,
		Version: "1.0.0",
	}
	if err := tracker.TrackInstallation("tracked-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	exists, path, err = tracker.DetectPreExisting("tracked-tool", "tracked-tool")
	if err != nil {
		t.Errorf("DetectPreExisting() error = %v", err)
		return
	}
	
	if exists {
		t.Error("DetectPreExisting() should return false for already tracked tool")
	}
}

func TestTracker_TrackPreExisting(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Track pre-existing tool
	toolName := "git"
	binaryPath := "/usr/bin/git"
	version := "2.34.1"
	
	if err := tracker.TrackPreExisting(toolName, binaryPath, version); err != nil {
		t.Errorf("TrackPreExisting() error = %v", err)
		return
	}
	
	// Verify tool was tracked
	if !tracker.IsInstalled(toolName) {
		t.Error("TrackPreExisting() should mark tool as installed")
	}
	
	// Verify installation record
	record, exists := tracker.GetInstallation(toolName)
	if !exists {
		t.Error("TrackPreExisting() should create installation record")
		return
	}
	
	if record.Method != MethodPreExisting {
		t.Errorf("Pre-existing tool method = %v, want %v", record.Method, MethodPreExisting)
	}
	
	if record.Version != version {
		t.Errorf("Pre-existing tool version = %v, want %v", record.Version, version)
	}
	
	if len(record.BinaryPaths) != 1 || record.BinaryPaths[0] != binaryPath {
		t.Errorf("Pre-existing tool binary paths = %v, want [%s]", record.BinaryPaths, binaryPath)
	}
	
	if record.UserRequested {
		t.Error("Pre-existing tool should not be marked as user requested")
	}
	
	// Check installation context
	if len(record.InstallationContext) != 1 || record.InstallationContext[0] != "pre_existing" {
		t.Errorf("Pre-existing tool context = %v, want [pre_existing]", record.InstallationContext)
	}
}

func TestTracker_GetDependents(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Track tools with shared dependency
	tools := []string{"ripgrep", "fd", "bat"}
	dependency := "rust"
	
	for _, tool := range tools {
		config := TrackingConfig{
			Method:       MethodCargoInstall,
			Version:      "1.0.0",
			Dependencies: []string{dependency},
		}
		if err := tracker.TrackInstallation(tool, config); err != nil {
			t.Fatalf("Failed to track tool %s: %v", tool, err)
		}
	}
	
	// Get dependents
	dependents := tracker.GetDependents(dependency)
	if len(dependents) != len(tools) {
		t.Errorf("GetDependents() count = %d, want %d", len(dependents), len(tools))
	}
	
	// Verify all tools are listed as dependents
	for _, tool := range tools {
		found := false
		for _, dependent := range dependents {
			if dependent == tool {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetDependents() should include %s", tool)
		}
	}
	
	// Test non-existent dependency
	nonExistentDependents := tracker.GetDependents("nonexistent-dep")
	if len(nonExistentDependents) != 0 {
		t.Error("GetDependents() should return empty slice for non-existent dependency")
	}
}

func TestTracker_CanSafelyRemove(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Test non-tracked tool
	_, _, err = tracker.CanSafelyRemove("nonexistent")
	if err == nil {
		t.Error("CanSafelyRemove() should error for non-tracked tool")
	}
	
	// Track pre-existing tool (should not be removable)
	if err := tracker.TrackPreExisting("git", "/usr/bin/git", "2.34.1"); err != nil {
		t.Fatalf("TrackPreExisting() error = %v", err)
	}
	
	canRemove, reasons, err := tracker.CanSafelyRemove("git")
	if err != nil {
		t.Errorf("CanSafelyRemove() error = %v", err)
		return
	}
	
	if canRemove {
		t.Error("CanSafelyRemove() should return false for pre-existing tool")
	}
	
	if len(reasons) == 0 {
		t.Error("CanSafelyRemove() should provide reasons for pre-existing tool")
	}
	
	// Track tool with no dependents (should be removable)
	config := TrackingConfig{
		Method:  MethodSourceBuild,
		Version: "1.0.0",
	}
	if err := tracker.TrackInstallation("standalone-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	canRemove, reasons, err = tracker.CanSafelyRemove("standalone-tool")
	if err != nil {
		t.Errorf("CanSafelyRemove() error = %v", err)
		return
	}
	
	if !canRemove {
		t.Error("CanSafelyRemove() should return true for standalone tool")
	}
	
	if len(reasons) != 0 {
		t.Error("CanSafelyRemove() should not provide reasons for removable tool")
	}
	
	// Track tool with dependents (should not be removable)
	rustConfig := TrackingConfig{
		Method:  MethodSystemPackage,
		Version: "stable",
	}
	if err := tracker.TrackInstallation("rust", rustConfig); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	ripgrepConfig := TrackingConfig{
		Method:       MethodCargoInstall,
		Version:      "13.0.0",
		Dependencies: []string{"rust"},
	}
	if err := tracker.TrackInstallation("ripgrep", ripgrepConfig); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	canRemove, reasons, err = tracker.CanSafelyRemove("rust")
	if err != nil {
		t.Errorf("CanSafelyRemove() error = %v", err)
		return
	}
	
	if canRemove {
		t.Error("CanSafelyRemove() should return false for tool with dependents")
	}
	
	if len(reasons) == 0 {
		t.Error("CanSafelyRemove() should provide reasons for tool with dependents")
	}
	
	// Check that reasons mention dependents
	found := false
	for _, reason := range reasons {
		if strings.Contains(reason, "Required by") && strings.Contains(reason, "ripgrep") {
			found = true
			break
		}
	}
	if !found {
		t.Error("CanSafelyRemove() reasons should mention dependent tools")
	}
}

func TestTracker_CreateSnapshot(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Track some installations to have data to backup
	config := TrackingConfig{
		Method:  MethodSourceBuild,
		Version: "1.0.0",
	}
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	// Create snapshot
	suffix := "test-snapshot"
	if err := tracker.CreateSnapshot(suffix); err != nil {
		t.Errorf("CreateSnapshot() error = %v", err)
		return
	}
	
	// Verify snapshot was created by checking backup files
	backups, err := tracker.manager.ListBackups()
	if err != nil {
		t.Errorf("Failed to list backups: %v", err)
		return
	}
	
	if len(backups) == 0 {
		t.Error("CreateSnapshot() should create a backup file")
		return
	}
	
	// Check that one of the backups contains our suffix
	found := false
	for _, backup := range backups {
		if strings.Contains(backup, suffix) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("CreateSnapshot() should create backup with suffix %s", suffix)
	}
}

func TestTracker_GetBinaryPaths(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Test non-tracked tool
	_, err = tracker.GetBinaryPaths("nonexistent")
	if err == nil {
		t.Error("GetBinaryPaths() should error for non-tracked tool")
	}
	
	// Track tool with binary paths
	binaryPaths := []string{"/usr/local/bin/tool", "/usr/local/bin/tool-alias"}
	config := TrackingConfig{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		BinaryPaths: binaryPaths,
	}
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	// Get binary paths
	paths, err := tracker.GetBinaryPaths("test-tool")
	if err != nil {
		t.Errorf("GetBinaryPaths() error = %v", err)
		return
	}
	
	if len(paths) != len(binaryPaths) {
		t.Errorf("GetBinaryPaths() count = %d, want %d", len(paths), len(binaryPaths))
	}
	
	for i, path := range paths {
		if path != binaryPaths[i] {
			t.Errorf("GetBinaryPaths()[%d] = %s, want %s", i, path, binaryPaths[i])
		}
	}
}

func TestTracker_GetBuildDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Test non-tracked tool
	_, err = tracker.GetBuildDir("nonexistent")
	if err == nil {
		t.Error("GetBuildDir() should error for non-tracked tool")
	}
	
	// Track tool with build directory
	buildDir := "/tmp/build/test-tool"
	config := TrackingConfig{
		Method:   MethodSourceBuild,
		Version:  "1.0.0",
		BuildDir: buildDir,
	}
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("TrackInstallation() error = %v", err)
	}
	
	// Get build directory
	dir, err := tracker.GetBuildDir("test-tool")
	if err != nil {
		t.Errorf("GetBuildDir() error = %v", err)
		return
	}
	
	if dir != buildDir {
		t.Errorf("GetBuildDir() = %s, want %s", dir, buildDir)
	}
}

func TestContains(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}
	
	// Test existing item
	if !contains(slice, "banana") {
		t.Error("contains() should return true for existing item")
	}
	
	// Test non-existing item
	if contains(slice, "grape") {
		t.Error("contains() should return false for non-existing item")
	}
	
	// Test empty slice
	if contains([]string{}, "item") {
		t.Error("contains() should return false for empty slice")
	}
}

func TestDetectBinaryPaths(t *testing.T) {
	// Test with a tool that likely exists (sh)
	paths := DetectBinaryPaths("sh", []string{"bash"})
	
	// Should find at least sh
	if len(paths) == 0 {
		t.Skip("No shell binaries found in PATH, skipping test")
	}
	
	// Verify all returned paths are absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("DetectBinaryPaths() should return absolute paths, got %s", path)
		}
	}
	
	// Test with non-existent tool
	paths = DetectBinaryPaths("nonexistent-tool-12345", []string{"also-nonexistent"})
	if len(paths) != 0 {
		t.Error("DetectBinaryPaths() should return empty slice for non-existent tools")
	}
}

func TestTracker_GetInstallationStats(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		t.Fatalf("NewTracker() error = %v", err)
	}
	
	// Track various types of installations
	installations := map[string]TrackingConfig{
		"ripgrep": {
			Method:  MethodCargoInstall,
			Version: "13.0.0",
		},
		"fzf": {
			Method:  MethodGoInstall,
			Version: "0.44.1",
		},
		"jq": {
			Method:  MethodSourceBuild,
			Version: "1.6",
		},
		"git": {
			Method:  MethodPreExisting,
			Version: "2.34.1",
		},
	}
	
	for name, config := range installations {
		if config.Method == MethodPreExisting {
			if err := tracker.TrackPreExisting(name, "/usr/bin/"+name, config.Version); err != nil {
				t.Fatalf("Failed to track pre-existing tool %s: %v", name, err)
			}
		} else {
			if err := tracker.TrackInstallation(name, config); err != nil {
				t.Fatalf("Failed to track tool %s: %v", name, err)
			}
		}
	}
	
	// Track a bundle
	bundleTools := []string{"ripgrep", "fzf", "jq"}
	if err := tracker.TrackBundle("essential-tools", bundleTools, true); err != nil {
		t.Fatalf("Failed to track bundle: %v", err)
	}
	
	// Get stats
	stats := tracker.GetInstallationStats()
	
	// Verify total count (4 tools + 1 bundle = 5)
	expectedTotal := 5
	if stats["total"] != expectedTotal {
		t.Errorf("GetInstallationStats() total = %d, want %d", stats["total"], expectedTotal)
	}
	
	// Verify tool count (4 tools)
	expectedTools := 4
	if stats["tools"] != expectedTools {
		t.Errorf("GetInstallationStats() tools = %d, want %d", stats["tools"], expectedTools)
	}
	
	// Verify bundle count (1 bundle)
	expectedBundles := 1
	if stats["bundles"] != expectedBundles {
		t.Errorf("GetInstallationStats() bundles = %d, want %d", stats["bundles"], expectedBundles)
	}
	
	// Verify method counts
	if stats["cargo_install"] != 1 {
		t.Errorf("GetInstallationStats() cargo_install = %d, want 1", stats["cargo_install"])
	}
	
	if stats["go_install"] != 1 {
		t.Errorf("GetInstallationStats() go_install = %d, want 1", stats["go_install"])
	}
	
	if stats["source_build"] != 1 {
		t.Errorf("GetInstallationStats() source_build = %d, want 1", stats["source_build"])
	}
	
	if stats["pre_existing"] != 1 {
		t.Errorf("GetInstallationStats() pre_existing = %d, want 1", stats["pre_existing"])
	}
	
	if stats["bundle"] != 1 {
		t.Errorf("GetInstallationStats() bundle = %d, want 1", stats["bundle"])
	}
}

// Benchmark tests
func BenchmarkTracker_TrackInstallation(b *testing.B) {
	tempDir := b.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		b.Fatalf("NewTracker() error = %v", err)
	}
	
	config := TrackingConfig{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		BinaryPaths: []string{"/usr/local/bin/test"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toolName := "tool-" + string(rune(i))
		_ = tracker.TrackInstallation(toolName, config)
	}
}

func BenchmarkTracker_IsInstalled(b *testing.B) {
	tempDir := b.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)
	
	tracker, err := NewTracker()
	if err != nil {
		b.Fatalf("NewTracker() error = %v", err)
	}
	
	// Track some tools for testing
	config := TrackingConfig{
		Method:  MethodSourceBuild,
		Version: "1.0.0",
	}
	for i := 0; i < 100; i++ {
		toolName := "tool-" + string(rune(i))
		_ = tracker.TrackInstallation(toolName, config)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toolName := "tool-" + string(rune(i%100))
		_ = tracker.IsInstalled(toolName)
	}
}