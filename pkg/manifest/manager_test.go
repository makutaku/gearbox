package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	
	if manager == nil {
		t.Fatal("NewManager() should return a manager instance")
	}
	
	// Check that paths are set
	if manager.manifestPath == "" {
		t.Error("NewManager() should set manifestPath")
	}
	
	if manager.backupDir == "" {
		t.Error("NewManager() should set backupDir")
	}
	
	// Check that paths contain expected components
	if !strings.Contains(manager.manifestPath, ManifestDir) {
		t.Errorf("manifestPath should contain %s, got %s", ManifestDir, manager.manifestPath)
	}
	
	if !strings.Contains(manager.manifestPath, ManifestFile) {
		t.Errorf("manifestPath should contain %s, got %s", ManifestFile, manager.manifestPath)
	}
	
	if !strings.Contains(manager.backupDir, BackupDir) {
		t.Errorf("backupDir should contain %s, got %s", BackupDir, manager.backupDir)
	}
}

func TestManager_GetManifestPath(t *testing.T) {
	manager := NewManager()
	
	path := manager.GetManifestPath()
	if path == "" {
		t.Error("GetManifestPath() should return non-empty path")
	}
	
	if path != manager.manifestPath {
		t.Error("GetManifestPath() should return the same path as stored internally")
	}
}

func TestManager_EnsureManifestDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, "test", ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, "test", BackupDir),
	}
	
	// Ensure directory doesn't exist initially
	manifestDir := filepath.Dir(manifestPath)
	if _, err := os.Stat(manifestDir); !os.IsNotExist(err) {
		t.Skip("Test directory already exists")
	}
	
	// Test directory creation
	if err := manager.EnsureManifestDir(); err != nil {
		t.Errorf("EnsureManifestDir() error = %v", err)
	}
	
	// Verify directory was created
	if _, err := os.Stat(manifestDir); os.IsNotExist(err) {
		t.Error("EnsureManifestDir() should create the directory")
	}
	
	// Test that calling again doesn't cause error
	if err := manager.EnsureManifestDir(); err != nil {
		t.Errorf("EnsureManifestDir() should not error when directory exists: %v", err)
	}
}

func TestManager_EnsureBackupDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	backupDir := filepath.Join(tempDir, "test", BackupDir)
	manager := &Manager{
		manifestPath: filepath.Join(tempDir, "test", ManifestFile),
		backupDir:    backupDir,
	}
	
	// Ensure directory doesn't exist initially
	if _, err := os.Stat(backupDir); !os.IsNotExist(err) {
		t.Skip("Test backup directory already exists")
	}
	
	// Test directory creation
	if err := manager.EnsureBackupDir(); err != nil {
		t.Errorf("EnsureBackupDir() error = %v", err)
	}
	
	// Verify directory was created
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("EnsureBackupDir() should create the directory")
	}
	
	// Test that calling again doesn't cause error
	if err := manager.EnsureBackupDir(); err != nil {
		t.Errorf("EnsureBackupDir() should not error when directory exists: %v", err)
	}
}

func TestManager_Load_NewManifest(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Ensure manifest file doesn't exist
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Skip("Test manifest file already exists")
	}
	
	// Load should create new manifest
	manifest, err := manager.Load()
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}
	
	if manifest == nil {
		t.Fatal("Load() should return a manifest instance")
	}
	
	// Verify it's a new manifest
	if manifest.SchemaVersion != SchemaVersion {
		t.Errorf("Load() SchemaVersion = %v, want %v", manifest.SchemaVersion, SchemaVersion)
	}
	
	if len(manifest.Installations) != 0 {
		t.Error("Load() should create empty manifest")
	}
	
	// Verify file was created
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Load() should create manifest file")
	}
}

func TestManager_Load_ExistingManifest(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Create and save a manifest first
	originalManifest := NewManifest()
	record := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}
	originalManifest.AddInstallation("test-tool", record)
	
	if err := manager.Save(originalManifest); err != nil {
		t.Fatalf("Failed to save test manifest: %v", err)
	}
	
	// Load the existing manifest
	loadedManifest, err := manager.Load()
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}
	
	if loadedManifest == nil {
		t.Fatal("Load() should return a manifest instance")
	}
	
	// Verify loaded manifest has the same data
	if len(loadedManifest.Installations) != 1 {
		t.Errorf("Load() installations count = %d, want 1", len(loadedManifest.Installations))
	}
	
	loadedRecord, exists := loadedManifest.GetInstallation("test-tool")
	if !exists {
		t.Error("Load() should preserve installation records")
	}
	
	if loadedRecord.Method != MethodSourceBuild {
		t.Errorf("Load() method = %v, want %v", loadedRecord.Method, MethodSourceBuild)
	}
}

func TestManager_Save(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Create test manifest
	manifest := NewManifest()
	record := &InstallationRecord{
		Method:      MethodCargoInstall,
		Version:     "2.0.0",
		InstalledAt: time.Now(),
		BinaryPaths: []string{"/usr/local/bin/test"},
	}
	manifest.AddInstallation("test-tool", record)
	
	// Save manifest
	if err := manager.Save(manifest); err != nil {
		t.Errorf("Save() error = %v", err)
		return
	}
	
	// Verify file was created
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Save() should create manifest file")
	}
	
	// Verify file contents by loading
	loadedManifest, err := manager.Load()
	if err != nil {
		t.Errorf("Failed to load saved manifest: %v", err)
		return
	}
	
	if len(loadedManifest.Installations) != 1 {
		t.Errorf("Saved manifest installations count = %d, want 1", len(loadedManifest.Installations))
	}
	
	loadedRecord, exists := loadedManifest.GetInstallation("test-tool")
	if !exists {
		t.Error("Save() should preserve installation records")
	}
	
	if loadedRecord.Version != "2.0.0" {
		t.Errorf("Save() version = %v, want %v", loadedRecord.Version, "2.0.0")
	}
}

func TestManager_Save_AtomicWrite(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Create test manifest
	manifest := NewManifest()
	
	// Save manifest
	if err := manager.Save(manifest); err != nil {
		t.Errorf("Save() error = %v", err)
		return
	}
	
	// Verify temporary file was cleaned up
	tempPath := manifestPath + ".tmp"
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Save() should clean up temporary file")
	}
	
	// Verify final file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Save() should create final manifest file")
	}
}

func TestManager_Backup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	backupDir := filepath.Join(tempDir, BackupDir)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    backupDir,
	}
	
	// Create and save a test manifest
	manifest := NewManifest()
	record := &InstallationRecord{
		Method:      MethodGoInstall,
		Version:     "1.5.0",
		InstalledAt: time.Now(),
	}
	manifest.AddInstallation("test-tool", record)
	
	if err := manager.Save(manifest); err != nil {
		t.Fatalf("Failed to save test manifest: %v", err)
	}
	
	// Create backup
	suffix := "test-backup"
	if err := manager.Backup(suffix); err != nil {
		t.Errorf("Backup() error = %v", err)
		return
	}
	
	// Verify backup directory was created
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("Backup() should create backup directory")
	}
	
	// List backup files
	backups, err := manager.ListBackups()
	if err != nil {
		t.Errorf("Failed to list backups: %v", err)
		return
	}
	
	if len(backups) != 1 {
		t.Errorf("Backup count = %d, want 1", len(backups))
	}
	
	// Verify backup filename contains suffix
	if !strings.Contains(backups[0], suffix) {
		t.Errorf("Backup filename should contain suffix %s, got %s", suffix, backups[0])
	}
	
	// Verify backup filename has timestamp format
	if !strings.HasPrefix(backups[0], "manifest-") {
		t.Errorf("Backup filename should start with 'manifest-', got %s", backups[0])
	}
	
	if !strings.HasSuffix(backups[0], ".json") {
		t.Errorf("Backup filename should end with '.json', got %s", backups[0])
	}
}

func TestManager_Backup_NoManifest(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path (no manifest file exists)
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Backup should succeed even if no manifest exists
	if err := manager.Backup("test"); err != nil {
		t.Errorf("Backup() should not error when no manifest exists: %v", err)
	}
	
	// Should not create any backup files
	backups, err := manager.ListBackups()
	if err != nil {
		t.Errorf("Failed to list backups: %v", err)
		return
	}
	
	if len(backups) != 0 {
		t.Errorf("Backup count = %d, want 0 when no manifest exists", len(backups))
	}
}

func TestManager_ListBackups(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manager := &Manager{
		manifestPath: filepath.Join(tempDir, ManifestFile),
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Test with no backup directory
	backups, err := manager.ListBackups()
	if err != nil {
		t.Errorf("ListBackups() should not error when backup dir doesn't exist: %v", err)
		return
	}
	
	if len(backups) != 0 {
		t.Errorf("ListBackups() should return empty slice when no backups exist")
	}
	
	// Create backup directory with test files
	if err := manager.EnsureBackupDir(); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	
	// Create test backup files
	testFiles := []string{
		"manifest-20240101-120000.json",
		"manifest-20240101-130000-test.json",
		"not-a-backup.txt",
		"manifest-20240101-140000.json",
	}
	
	for _, filename := range testFiles {
		filepath := filepath.Join(manager.backupDir, filename)
		if err := os.WriteFile(filepath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}
	
	// List backups
	backups, err = manager.ListBackups()
	if err != nil {
		t.Errorf("ListBackups() error = %v", err)
		return
	}
	
	// Should only include .json files (exclude .txt file)
	expectedCount := 3
	if len(backups) != expectedCount {
		t.Errorf("ListBackups() count = %d, want %d", len(backups), expectedCount)
	}
	
	// Verify all returned files are JSON files
	for _, backup := range backups {
		if !strings.HasSuffix(backup, ".json") {
			t.Errorf("ListBackups() should only return .json files, got %s", backup)
		}
	}
}

func TestManager_RestoreBackup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	backupDir := filepath.Join(tempDir, BackupDir)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    backupDir,
	}
	
	// Create original manifest
	originalManifest := NewManifest()
	originalRecord := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}
	originalManifest.AddInstallation("original-tool", originalRecord)
	
	if err := manager.Save(originalManifest); err != nil {
		t.Fatalf("Failed to save original manifest: %v", err)
	}
	
	// Create backup
	if err := manager.Backup("test-restore"); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}
	
	// Modify manifest
	modifiedManifest := NewManifest()
	modifiedRecord := &InstallationRecord{
		Method:      MethodCargoInstall,
		Version:     "2.0.0",
		InstalledAt: time.Now(),
	}
	modifiedManifest.AddInstallation("modified-tool", modifiedRecord)
	
	if err := manager.Save(modifiedManifest); err != nil {
		t.Fatalf("Failed to save modified manifest: %v", err)
	}
	
	// Get backup filename
	backups, err := manager.ListBackups()
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}
	
	if len(backups) == 0 {
		t.Fatal("No backups found")
	}
	
	var testBackup string
	for _, backup := range backups {
		if strings.Contains(backup, "test-restore") {
			testBackup = backup
			break
		}
	}
	
	if testBackup == "" {
		t.Fatal("Test backup not found")
	}
	
	// Restore backup
	if err := manager.RestoreBackup(testBackup); err != nil {
		t.Errorf("RestoreBackup() error = %v", err)
		return
	}
	
	// Verify restoration
	restoredManifest, err := manager.Load()
	if err != nil {
		t.Errorf("Failed to load restored manifest: %v", err)
		return
	}
	
	// Should have original tool, not modified tool
	if !restoredManifest.IsInstalled("original-tool") {
		t.Error("RestoreBackup() should restore original tool")
	}
	
	if restoredManifest.IsInstalled("modified-tool") {
		t.Error("RestoreBackup() should not have modified tool")
	}
	
	// Verify there's now a pre-restore backup
	backupsAfterRestore, err := manager.ListBackups()
	if err != nil {
		t.Errorf("Failed to list backups after restore: %v", err)
		return
	}
	
	// Should have more backups now (original + pre-restore)
	if len(backupsAfterRestore) <= len(backups) {
		t.Error("RestoreBackup() should create a pre-restore backup")
	}
	
	// Check for pre-restore backup
	hasPreRestoreBackup := false
	for _, backup := range backupsAfterRestore {
		if strings.Contains(backup, "pre-restore") {
			hasPreRestoreBackup = true
			break
		}
	}
	
	if !hasPreRestoreBackup {
		t.Error("RestoreBackup() should create a pre-restore backup")
	}
}

func TestManager_RestoreBackup_NonExistentBackup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manager := &Manager{
		manifestPath: filepath.Join(tempDir, ManifestFile),
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Try to restore non-existent backup
	err := manager.RestoreBackup("nonexistent-backup.json")
	if err == nil {
		t.Error("RestoreBackup() should error for non-existent backup")
	}
	
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("RestoreBackup() error should mention file doesn't exist, got: %v", err)
	}
}

func TestManager_Exists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Should not exist initially
	if manager.Exists() {
		t.Error("Exists() should return false when manifest doesn't exist")
	}
	
	// Create manifest
	manifest := NewManifest()
	if err := manager.Save(manifest); err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}
	
	// Should exist now
	if !manager.Exists() {
		t.Error("Exists() should return true when manifest exists")
	}
}

// Test error conditions
func TestManager_Load_CorruptedManifest(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Create corrupted manifest file
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0755); err != nil {
		t.Fatalf("Failed to create manifest directory: %v", err)
	}
	
	corruptedJSON := `{"invalid": json, missing quotes}`
	if err := os.WriteFile(manifestPath, []byte(corruptedJSON), 0644); err != nil {
		t.Fatalf("Failed to write corrupted manifest: %v", err)
	}
	
	// Load should fail
	_, err := manager.Load()
	if err == nil {
		t.Error("Load() should error for corrupted manifest")
	}
	
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Load() error should mention parsing issue, got: %v", err)
	}
}

func TestManager_Constants(t *testing.T) {
	// Verify constants are defined correctly
	if ManifestDir == "" {
		t.Error("ManifestDir should not be empty")
	}
	
	if ManifestFile == "" {
		t.Error("ManifestFile should not be empty")
	}
	
	if BackupDir == "" {
		t.Error("BackupDir should not be empty")
	}
	
	// Verify expected values
	if ManifestDir != ".gearbox" {
		t.Errorf("ManifestDir = %s, want .gearbox", ManifestDir)
	}
	
	if ManifestFile != "manifest.json" {
		t.Errorf("ManifestFile = %s, want manifest.json", ManifestFile)
	}
	
	if BackupDir != "backups" {
		t.Errorf("BackupDir = %s, want backups", BackupDir)
	}
}

// Test concurrent access safety
func TestManager_ConcurrentSave(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create manager with custom path
	manifestPath := filepath.Join(tempDir, ManifestFile)
	manager := &Manager{
		manifestPath: manifestPath,
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Test concurrent saves (atomic writes should prevent corruption)
	const numGoroutines = 10
	done := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			manifest := NewManifest()
			record := &InstallationRecord{
				Method:      MethodSourceBuild,
				Version:     "1.0.0",
				InstalledAt: time.Now(),
			}
			manifest.AddInstallation("tool-"+string(rune(id)), record)
			done <- manager.Save(manifest)
		}(i)
	}
	
	// Wait for all goroutines to complete
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}
	
	// All saves should succeed (last one wins)
	if len(errors) > 0 {
		t.Errorf("Concurrent saves failed: %v", errors)
	}
	
	// Verify final manifest is valid
	manifest, err := manager.Load()
	if err != nil {
		t.Errorf("Failed to load manifest after concurrent saves: %v", err)
	}
	
	if manifest == nil {
		t.Error("Manifest should not be nil after concurrent saves")
	}
}

// Benchmark tests
func BenchmarkManager_Save(b *testing.B) {
	tempDir := b.TempDir()
	manager := &Manager{
		manifestPath: filepath.Join(tempDir, ManifestFile),
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	manifest := NewManifest()
	record := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}
	manifest.AddInstallation("test-tool", record)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Save(manifest)
	}
}

func BenchmarkManager_Load(b *testing.B) {
	tempDir := b.TempDir()
	manager := &Manager{
		manifestPath: filepath.Join(tempDir, ManifestFile),
		backupDir:    filepath.Join(tempDir, BackupDir),
	}
	
	// Create test manifest
	manifest := NewManifest()
	record := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}
	manifest.AddInstallation("test-tool", record)
	_ = manager.Save(manifest)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Load()
	}
}