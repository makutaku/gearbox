package uninstall

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gearbox/pkg/manifest"
)

// setupTestTracker creates a test manifest tracker in a temporary directory
func setupTestTracker(t *testing.T) (*manifest.Tracker, func()) {
	tempDir := t.TempDir()
	
	// Set up environment to use temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	
	tracker, err := manifest.NewTracker()
	if err != nil {
		t.Fatalf("Failed to create test tracker: %v", err)
	}
	
	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}
	
	return tracker, cleanup
}

func TestNewRemovalExecutor(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	tests := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "dry run executor",
			dryRun: true,
		},
		{
			name:   "real executor",
			dryRun: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewRemovalExecutor(tt.dryRun)
			if err != nil {
				t.Errorf("NewRemovalExecutor() error = %v", err)
				return
			}
			
			if executor == nil {
				t.Fatal("NewRemovalExecutor() should return an executor instance")
			}
			
			if executor.dryRun != tt.dryRun {
				t.Errorf("NewRemovalExecutor() dryRun = %v, want %v", executor.dryRun, tt.dryRun)
			}
			
			if executor.tracker == nil {
				t.Error("NewRemovalExecutor() should initialize tracker")
			}
		})
	}
}

func TestRemovalExecutor_ExecutePlan_DryRun(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(true) // dry run
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	plan := &RemovalPlan{
		ToRemove: []RemovalAction{
			{
				Target: "test-tool",
				Method: RemovalSourceBuild,
				Paths:  []string{"/usr/local/bin/test-tool"},
			},
		},
		Dependencies: []DependencyAction{
			{
				Dependency: "rust",
				Action:     "preserve",
				Reason:     "Still needed by other tools",
			},
		},
	}
	
	options := RemovalOptions{
		Backup: false,
	}
	
	result, err := executor.ExecutePlan(plan, options)
	if err != nil {
		t.Errorf("ExecutePlan() error = %v", err)
		return
	}
	
	if result == nil {
		t.Fatal("ExecutePlan() should return a result")
	}
	
	if !result.DryRun {
		t.Error("ExecutePlan() result should indicate dry run")
	}
	
	if len(result.Removed) != 1 {
		t.Errorf("ExecutePlan() removed count = %d, want 1", len(result.Removed))
	}
	
	if result.Removed[0] != "test-tool" {
		t.Errorf("ExecutePlan() removed[0] = %s, want test-tool", result.Removed[0])
	}
	
	if len(result.Failed) != 0 {
		t.Errorf("ExecutePlan() failed count = %d, want 0", len(result.Failed))
	}
}

func TestRemovalExecutor_ExecutePlan_WithBackup(t *testing.T) {
	tracker, cleanup := setupTestTracker(t)
	defer cleanup()
	
	// Add a test installation to have something to backup
	config := manifest.TrackingConfig{
		Method:  manifest.MethodSourceBuild,
		Version: "1.0.0",
	}
	if err := tracker.TrackInstallation("test-tool", config); err != nil {
		t.Fatalf("Failed to track test installation: %v", err)
	}
	
	executor, err := NewRemovalExecutor(true) // dry run to avoid actual removal
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	plan := &RemovalPlan{
		ToRemove: []RemovalAction{
			{
				Target: "test-tool",
				Method: RemovalSourceBuild,
				Paths:  []string{"/usr/local/bin/test-tool"},
			},
		},
	}
	
	options := RemovalOptions{
		Backup:       true,
		BackupSuffix: "test-backup",
	}
	
	result, err := executor.ExecutePlan(plan, options)
	if err != nil {
		t.Errorf("ExecutePlan() error = %v", err)
		return
	}
	
	if !result.DryRun {
		t.Error("Dry run should not create actual backup")
	}
}

func TestRemovalExecutor_ExecutePlan_WithErrors(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(true) // dry run
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	plan := &RemovalPlan{
		ToRemove: []RemovalAction{
			{
				Target: "valid-tool",
				Method: RemovalSourceBuild,
				Paths:  []string{"/usr/local/bin/valid-tool"},
			},
			{
				Target: "invalid-tool",
				Method: RemovalMethod("invalid_method"),
				Paths:  []string{},
			},
		},
	}
	
	options := RemovalOptions{}
	
	result, err := executor.ExecutePlan(plan, options)
	if err != nil {
		t.Errorf("ExecutePlan() should not return error for dry run: %v", err)
		return
	}
	
	// Should have one successful and one failed
	if len(result.Removed) != 1 {
		t.Errorf("ExecutePlan() removed count = %d, want 1", len(result.Removed))
	}
	
	if len(result.Failed) != 1 {
		t.Errorf("ExecutePlan() failed count = %d, want 1", len(result.Failed))
	}
	
	if result.Failed[0].Target != "invalid-tool" {
		t.Errorf("ExecutePlan() failed[0].Target = %s, want invalid-tool", result.Failed[0].Target)
	}
}

func TestRemovalExecutor_executeRemovalAction_DryRun(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(true) // dry run
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	tests := []struct {
		name   string
		action RemovalAction
	}{
		{
			name: "source build removal",
			action: RemovalAction{
				Target: "test-tool",
				Method: RemovalSourceBuild,
				Paths:  []string{"/usr/local/bin/test-tool"},
			},
		},
		{
			name: "cargo install removal",
			action: RemovalAction{
				Target: "ripgrep",
				Method: RemovalCargoInstall,
			},
		},
		{
			name: "go install removal",
			action: RemovalAction{
				Target: "fzf",
				Method: RemovalGoInstall,
				Paths:  []string{"/usr/local/bin/fzf"},
			},
		},
		{
			name: "bundle removal",
			action: RemovalAction{
				Target: "essential-tools",
				Method: RemovalBundle,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RemovalResult{}
			
			err := executor.executeRemovalAction(tt.action, result)
			
			// Dry run should never error for valid methods
			if err != nil {
				t.Errorf("executeRemovalAction() dry run error = %v", err)
			}
		})
	}
}

func TestRemovalExecutor_executeRemovalAction_PreExisting(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(false) // real execution
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	action := RemovalAction{
		Target: "pre-existing-tool",
		Method: RemovalPreExisting,
	}
	
	result := &RemovalResult{}
	err = executor.executeRemovalAction(action, result)
	
	if err == nil {
		t.Error("executeRemovalAction() should error for pre-existing tools")
	}
	
	if !strings.Contains(err.Error(), "cannot remove pre-existing") {
		t.Errorf("executeRemovalAction() error should mention pre-existing, got: %v", err)
	}
}

func TestRemovalExecutor_executeRemovalAction_UnknownMethod(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(false) // real execution
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	action := RemovalAction{
		Target: "test-tool",
		Method: RemovalMethod("unknown_method"),
	}
	
	result := &RemovalResult{}
	err = executor.executeRemovalAction(action, result)
	
	if err == nil {
		t.Error("executeRemovalAction() should error for unknown methods")
	}
	
	if !strings.Contains(err.Error(), "unknown removal method") {
		t.Errorf("executeRemovalAction() error should mention unknown method, got: %v", err)
	}
}

func TestRemovalExecutor_removeFiles(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(false) // real execution
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	// Create temporary files for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-file.txt")
	testContent := "test content"
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	paths := []string{testFile}
	result := &RemovalResult{}
	
	err = executor.removeFiles(paths, result)
	if err != nil {
		t.Errorf("removeFiles() error = %v", err)
	}
	
	// Verify file was removed
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("removeFiles() should have removed the test file")
	}
	
	// Verify space was calculated
	if result.SpaceFreed <= 0 {
		t.Error("removeFiles() should calculate space freed")
	}
	
	expectedSize := int64(len(testContent))
	if result.SpaceFreed != expectedSize {
		t.Errorf("removeFiles() space freed = %d, want %d", result.SpaceFreed, expectedSize)
	}
}

func TestRemovalExecutor_removeFiles_NonExistentFile(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(false) // real execution
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	paths := []string{"/nonexistent/file/path"}
	result := &RemovalResult{}
	
	// Should not error for non-existent files
	err = executor.removeFiles(paths, result)
	if err != nil {
		t.Errorf("removeFiles() should not error for non-existent files: %v", err)
	}
}

func TestRemovalExecutor_removeFiles_Directory(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(false) // real execution
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	// Create temporary directory with files
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test-dir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	testFile := filepath.Join(testDir, "file.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	paths := []string{testDir}
	result := &RemovalResult{}
	
	err = executor.removeFiles(paths, result)
	if err != nil {
		t.Errorf("removeFiles() error = %v", err)
	}
	
	// Verify directory was removed
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("removeFiles() should have removed the test directory")
	}
	
	// Verify space was calculated for directory contents
	if result.SpaceFreed <= 0 {
		t.Error("removeFiles() should calculate space freed for directories")
	}
}

func TestRemovalExecutor_removeFiles_EmptyPaths(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(false) // real execution
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	paths := []string{"", "", ""} // Empty paths should be skipped
	result := &RemovalResult{}
	
	err = executor.removeFiles(paths, result)
	if err != nil {
		t.Errorf("removeFiles() should not error for empty paths: %v", err)
	}
	
	if result.SpaceFreed != 0 {
		t.Error("removeFiles() should not report space freed for empty paths")
	}
}

func TestRemovalExecutor_executeDependencyAction_DryRun(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(true) // dry run
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	tests := []struct {
		name   string
		action DependencyAction
	}{
		{
			name: "preserve dependency",
			action: DependencyAction{
				Dependency: "rust",
				Action:     "preserve",
				Reason:     "Still needed by other tools",
			},
		},
		{
			name: "remove dependency",
			action: DependencyAction{
				Dependency: "unused-dep",
				Action:     RemovalManualDelete,
				Reason:     "No longer needed",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RemovalResult{}
			
			err := executor.executeDependencyAction(tt.action, result)
			
			// Dry run should never error
			if err != nil {
				t.Errorf("executeDependencyAction() dry run error = %v", err)
			}
		})
	}
}

func TestGetDirSize(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create nested directories with files
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	
	content1 := "Hello, World!"
	content2 := "This is a test file."
	
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	
	size, err := getDirSize(tempDir)
	if err != nil {
		t.Errorf("getDirSize() error = %v", err)
		return
	}
	
	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("getDirSize() = %d, want %d", size, expectedSize)
	}
}

func TestGetDirSize_NonExistentDir(t *testing.T) {
	size, err := getDirSize("/nonexistent/directory")
	if err == nil {
		t.Error("getDirSize() should error for non-existent directory")
	}
	
	if size != 0 {
		t.Errorf("getDirSize() should return 0 for non-existent directory, got %d", size)
	}
}

func TestRemovalResult_FormatSpaceFreed(t *testing.T) {
	tests := []struct {
		name       string
		spaceFreed int64
		expected   string
	}{
		{
			name:       "zero bytes",
			spaceFreed: 0,
			expected:   "0 B",
		},
		{
			name:       "bytes",
			spaceFreed: 512,
			expected:   "512 B",
		},
		{
			name:       "kilobytes",
			spaceFreed: 1024,
			expected:   "1 KB",
		},
		{
			name:       "kilobytes with decimal",
			spaceFreed: 1536, // 1.5 KB
			expected:   "1.5 KB",
		},
		{
			name:       "megabytes",
			spaceFreed: 1024 * 1024,
			expected:   "1 MB",
		},
		{
			name:       "gigabytes",
			spaceFreed: 1024 * 1024 * 1024,
			expected:   "1 GB",
		},
		{
			name:       "large value",
			spaceFreed: 2 * 1024 * 1024 * 1024, // 2 GB
			expected:   "2 GB",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &RemovalResult{
				SpaceFreed: tt.spaceFreed,
			}
			
			formatted := result.FormatSpaceFreed()
			if formatted != tt.expected {
				t.Errorf("FormatSpaceFreed() = %s, want %s", formatted, tt.expected)
			}
		})
	}
}

func TestRemovalResult_Summary(t *testing.T) {
	tests := []struct {
		name     string
		result   RemovalResult
		contains []string
	}{
		{
			name: "dry run summary",
			result: RemovalResult{
				Removed:       []string{"tool1", "tool2"},
				Failed:        []RemovalError{},
				DryRun:        true,
				SpaceFreed:    1024 * 1024, // 1 MB
				BackupCreated: false,
			},
			contains: []string{
				"🧪 Dry Run Summary",
				"✅ Successfully removed: 2 tools",
				"💾 Space freed: 1 MB",
			},
		},
		{
			name: "real execution summary with failures",
			result: RemovalResult{
				Removed: []string{"tool1"},
				Failed: []RemovalError{
					{Target: "tool2", Error: "permission denied"},
				},
				DryRun:        false,
				SpaceFreed:    512 * 1024, // 512 KB
				BackupCreated: true,
			},
			contains: []string{
				"📊 Removal Summary",
				"✅ Successfully removed: 1 tools",
				"❌ Failed to remove: 1 tools",
				"💾 Space freed: 512 KB",
				"🔄 Backup created before removal",
			},
		},
		{
			name: "no space freed",
			result: RemovalResult{
				Removed:       []string{"tool1"},
				Failed:        []RemovalError{},
				DryRun:        false,
				SpaceFreed:    0,
				BackupCreated: false,
			},
			contains: []string{
				"📊 Removal Summary",
				"✅ Successfully removed: 1 tools",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.Summary()
			
			for _, expectedText := range tt.contains {
				if !strings.Contains(summary, expectedText) {
					t.Errorf("Summary() should contain %q, got:\n%s", expectedText, summary)
				}
			}
		})
	}
}

func TestRemovalError_Struct(t *testing.T) {
	err := RemovalError{
		Target: "test-tool",
		Error:  "permission denied",
	}
	
	if err.Target != "test-tool" {
		t.Errorf("RemovalError.Target = %s, want test-tool", err.Target)
	}
	
	if err.Error != "permission denied" {
		t.Errorf("RemovalError.Error = %s, want permission denied", err.Error)
	}
}

// Test removal methods with mock commands (dry run only for safety)
func TestRemovalExecutor_RemovalMethods_DryRun(t *testing.T) {
	_, cleanup := setupTestTracker(t)
	defer cleanup()
	
	executor, err := NewRemovalExecutor(true) // dry run only
	if err != nil {
		t.Fatalf("NewRemovalExecutor() error = %v", err)
	}
	
	tests := []struct {
		name   string
		method RemovalMethod
		target string
		paths  []string
	}{
		{
			name:   "cargo install removal",
			method: RemovalCargoInstall,
			target: "ripgrep",
			paths:  []string{},
		},
		{
			name:   "go install removal",
			method: RemovalGoInstall,
			target: "fzf",
			paths:  []string{"/usr/local/bin/fzf"},
		},
		{
			name:   "pipx removal",
			method: RemovalPipx,
			target: "black",
			paths:  []string{},
		},
		{
			name:   "npm global removal",
			method: RemovalNpmGlobal,
			target: "typescript",
			paths:  []string{},
		},
		{
			name:   "system package removal",
			method: RemovalSystemPackage,
			target: "curl",
			paths:  []string{},
		},
		{
			name:   "source build removal",
			method: RemovalSourceBuild,
			target: "custom-tool",
			paths:  []string{"/usr/local/bin/custom-tool", "/tmp/build/custom-tool"},
		},
		{
			name:   "manual delete removal",
			method: RemovalManualDelete,
			target: "downloaded-tool",
			paths:  []string{"/opt/downloaded-tool"},
		},
		{
			name:   "bundle removal",
			method: RemovalBundle,
			target: "essential-tools",
			paths:  []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := RemovalAction{
				Target: tt.target,
				Method: tt.method,
				Paths:  tt.paths,
			}
			
			result := &RemovalResult{}
			
			// Dry run should not error for any valid method
			err := executor.executeRemovalAction(action, result)
			if err != nil {
				t.Errorf("executeRemovalAction() dry run should not error for %s: %v", tt.method, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkRemovalResult_FormatSpaceFreed(b *testing.B) {
	result := &RemovalResult{
		SpaceFreed: 1024 * 1024 * 512, // 512 MB
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.FormatSpaceFreed()
	}
}

func BenchmarkRemovalResult_Summary(b *testing.B) {
	result := &RemovalResult{
		Removed: []string{"tool1", "tool2", "tool3"},
		Failed: []RemovalError{
			{Target: "tool4", Error: "error1"},
			{Target: "tool5", Error: "error2"},
		},
		DryRun:        false,
		SpaceFreed:    1024 * 1024 * 100, // 100 MB
		BackupCreated: true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.Summary()
	}
}

func BenchmarkGetDirSize(b *testing.B) {
	// Create a temporary directory with some files
	tempDir := b.TempDir()
	
	for i := 0; i < 10; i++ {
		file := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
		content := strings.Repeat("test content ", 100) // ~1.3KB per file
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = getDirSize(tempDir)
	}
}