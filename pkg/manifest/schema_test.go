package manifest

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewManifest(t *testing.T) {
	manifest := NewManifest()
	
	if manifest.SchemaVersion != SchemaVersion {
		t.Errorf("NewManifest() SchemaVersion = %v, want %v", manifest.SchemaVersion, SchemaVersion)
	}
	
	if manifest.Installations == nil {
		t.Error("NewManifest() should initialize Installations map")
	}
	
	if manifest.Dependencies == nil {
		t.Error("NewManifest() should initialize Dependencies map")
	}
	
	if manifest.CreatedAt.IsZero() {
		t.Error("NewManifest() should set CreatedAt")
	}
	
	if manifest.UpdatedAt.IsZero() {
		t.Error("NewManifest() should set UpdatedAt")
	}
	
	if len(manifest.Installations) != 0 {
		t.Error("NewManifest() should start with empty installations")
	}
	
	if len(manifest.Dependencies) != 0 {
		t.Error("NewManifest() should start with empty dependencies")
	}
}

func TestInstallationManifest_AddInstallation(t *testing.T) {
	manifest := NewManifest()
	oldUpdateTime := manifest.UpdatedAt
	
	// Add slight delay to ensure UpdatedAt changes
	time.Sleep(1 * time.Millisecond)
	
	record := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
		BinaryPaths: []string{"/usr/local/bin/test"},
	}
	
	manifest.AddInstallation("test-tool", record)
	
	// Check that installation was added
	if len(manifest.Installations) != 1 {
		t.Errorf("AddInstallation() installations count = %d, want 1", len(manifest.Installations))
	}
	
	// Check that record was stored correctly
	storedRecord, exists := manifest.Installations["test-tool"]
	if !exists {
		t.Error("AddInstallation() should store the record")
	}
	
	if storedRecord != record {
		t.Error("AddInstallation() should store the exact record provided")
	}
	
	// Check that UpdatedAt was updated
	if !manifest.UpdatedAt.After(oldUpdateTime) {
		t.Error("AddInstallation() should update UpdatedAt timestamp")
	}
}

func TestInstallationManifest_AddDependency(t *testing.T) {
	manifest := NewManifest()
	oldUpdateTime := manifest.UpdatedAt
	
	// Add slight delay to ensure UpdatedAt changes
	time.Sleep(1 * time.Millisecond)
	
	record := &DependencyRecord{
		InstalledBy: "gearbox",
		Version:     "1.0.0",
		PreExisting: false,
		Dependents:  []string{"test-tool"},
		InstalledAt: time.Now(),
	}
	
	manifest.AddDependency("rust", record)
	
	// Check that dependency was added
	if len(manifest.Dependencies) != 1 {
		t.Errorf("AddDependency() dependencies count = %d, want 1", len(manifest.Dependencies))
	}
	
	// Check that record was stored correctly
	storedRecord, exists := manifest.Dependencies["rust"]
	if !exists {
		t.Error("AddDependency() should store the record")
	}
	
	if storedRecord != record {
		t.Error("AddDependency() should store the exact record provided")
	}
	
	// Check that UpdatedAt was updated
	if !manifest.UpdatedAt.After(oldUpdateTime) {
		t.Error("AddDependency() should update UpdatedAt timestamp")
	}
}

func TestInstallationManifest_GetInstallation(t *testing.T) {
	manifest := NewManifest()
	
	record := &InstallationRecord{
		Method:      MethodCargoInstall,
		Version:     "2.0.0",
		InstalledAt: time.Now(),
	}
	
	manifest.AddInstallation("ripgrep", record)
	
	// Test existing installation
	retrievedRecord, exists := manifest.GetInstallation("ripgrep")
	if !exists {
		t.Error("GetInstallation() should return true for existing installation")
	}
	
	if retrievedRecord != record {
		t.Error("GetInstallation() should return the correct record")
	}
	
	// Test non-existing installation
	_, exists = manifest.GetInstallation("nonexistent")
	if exists {
		t.Error("GetInstallation() should return false for non-existing installation")
	}
}

func TestInstallationManifest_GetDependency(t *testing.T) {
	manifest := NewManifest()
	
	record := &DependencyRecord{
		InstalledBy: "gearbox",
		Version:     "stable",
		Dependents:  []string{"ripgrep", "fd"},
		InstalledAt: time.Now(),
	}
	
	manifest.AddDependency("rust", record)
	
	// Test existing dependency
	retrievedRecord, exists := manifest.GetDependency("rust")
	if !exists {
		t.Error("GetDependency() should return true for existing dependency")
	}
	
	if retrievedRecord != record {
		t.Error("GetDependency() should return the correct record")
	}
	
	// Test non-existing dependency
	_, exists = manifest.GetDependency("nonexistent")
	if exists {
		t.Error("GetDependency() should return false for non-existing dependency")
	}
}

func TestInstallationManifest_IsInstalled(t *testing.T) {
	manifest := NewManifest()
	
	record := &InstallationRecord{
		Method:      MethodGoInstall,
		Version:     "latest",
		InstalledAt: time.Now(),
	}
	
	manifest.AddInstallation("fzf", record)
	
	// Test existing installation
	if !manifest.IsInstalled("fzf") {
		t.Error("IsInstalled() should return true for existing installation")
	}
	
	// Test non-existing installation
	if manifest.IsInstalled("nonexistent") {
		t.Error("IsInstalled() should return false for non-existing installation")
	}
}

func TestInstallationManifest_GetDependents(t *testing.T) {
	manifest := NewManifest()
	
	// Add dependency with multiple dependents
	record := &DependencyRecord{
		InstalledBy: "gearbox",
		Version:     "stable",
		Dependents:  []string{"ripgrep", "fd", "bat"},
		InstalledAt: time.Now(),
	}
	
	manifest.AddDependency("rust", record)
	
	// Test existing dependency
	dependents := manifest.GetDependents("rust")
	if len(dependents) != 3 {
		t.Errorf("GetDependents() count = %d, want 3", len(dependents))
	}
	
	expectedDependents := []string{"ripgrep", "fd", "bat"}
	for _, expected := range expectedDependents {
		found := false
		for _, dependent := range dependents {
			if dependent == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetDependents() missing dependent: %s", expected)
		}
	}
	
	// Test non-existing dependency
	nonExistentDependents := manifest.GetDependents("nonexistent")
	if len(nonExistentDependents) != 0 {
		t.Error("GetDependents() should return empty slice for non-existing dependency")
	}
}

func TestInstallationManifest_ToJSON(t *testing.T) {
	manifest := NewManifest()
	
	record := &InstallationRecord{
		Method:           MethodSourceBuild,
		Version:          "1.0.0",
		InstalledAt:      time.Now(),
		BinaryPaths:      []string{"/usr/local/bin/test"},
		Dependencies:     []string{"rust"},
		UserRequested:    true,
		ConfigFiles:      []string{"/home/user/.config/test.conf"},
		SystemPackages:   []string{"build-essential"},
	}
	
	manifest.AddInstallation("test-tool", record)
	
	data, err := manifest.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() error = %v", err)
		return
	}
	
	if len(data) == 0 {
		t.Error("ToJSON() should return non-empty data")
	}
	
	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("ToJSON() should produce valid JSON: %v", err)
	}
	
	// Check that indentation is present (pretty-printed)
	if !strings.Contains(string(data), "\n") {
		t.Error("ToJSON() should produce indented/pretty-printed JSON")
	}
}

func TestFromJSON(t *testing.T) {
	// Create a test manifest
	original := NewManifest()
	
	record := &InstallationRecord{
		Method:           MethodCargoInstall,
		Version:          "2.1.0",
		InstalledAt:      time.Now().Truncate(time.Second), // Truncate for JSON precision
		BinaryPaths:      []string{"/usr/local/bin/rg"},
		Dependencies:     []string{"rust"},
		UserRequested:    true,
		ConfigFiles:      []string{},
		SystemPackages:   []string{},
	}
	
	original.AddInstallation("ripgrep", record)
	
	// Serialize to JSON
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize original manifest: %v", err)
	}
	
	// Deserialize from JSON
	parsed, err := FromJSON(data)
	if err != nil {
		t.Errorf("FromJSON() error = %v", err)
		return
	}
	
	// Verify deserialized manifest
	if parsed.SchemaVersion != original.SchemaVersion {
		t.Errorf("FromJSON() SchemaVersion = %v, want %v", parsed.SchemaVersion, original.SchemaVersion)
	}
	
	if len(parsed.Installations) != len(original.Installations) {
		t.Errorf("FromJSON() installations count = %d, want %d", len(parsed.Installations), len(original.Installations))
	}
	
	// Check specific installation
	parsedRecord, exists := parsed.GetInstallation("ripgrep")
	if !exists {
		t.Error("FromJSON() should preserve installations")
	}
	
	if parsedRecord.Method != record.Method {
		t.Errorf("FromJSON() Method = %v, want %v", parsedRecord.Method, record.Method)
	}
	
	if parsedRecord.Version != record.Version {
		t.Errorf("FromJSON() Version = %v, want %v", parsedRecord.Version, record.Version)
	}
	
	if len(parsedRecord.BinaryPaths) != len(record.BinaryPaths) {
		t.Errorf("FromJSON() BinaryPaths count = %d, want %d", len(parsedRecord.BinaryPaths), len(record.BinaryPaths))
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	invalidJSON := `{"invalid": json}`
	
	_, err := FromJSON([]byte(invalidJSON))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON")
	}
}

func TestInstallationManifest_Validate(t *testing.T) {
	tests := []struct {
		name          string
		manifest      *InstallationManifest
		expectedError bool
		errorContains string
	}{
		{
			name:          "valid manifest",
			manifest:      NewManifest(),
			expectedError: false,
		},
		{
			name: "unsupported schema version",
			manifest: &InstallationManifest{
				SchemaVersion: "2.0",
				Installations: make(map[string]*InstallationRecord),
				Dependencies:  make(map[string]*DependencyRecord),
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
			expectedError: true,
			errorContains: "Unsupported schema version",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			
			if tt.expectedError {
				if err == nil {
					t.Error("Validate() should return error")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Validate() error = %v, should contain %s", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() should not return error, got %v", err)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Message: "test error"}
	
	expected := "Manifest validation error: test error"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}

func TestInstallationMethods(t *testing.T) {
	// Verify that all installation methods are defined as constants
	methods := []InstallationMethod{
		MethodSourceBuild,
		MethodCargoInstall,
		MethodGoInstall,
		MethodSystemPackage,
		MethodPipx,
		MethodNpmGlobal,
		MethodManualDownload,
		MethodBundle,
		MethodPreExisting,
	}
	
	for _, method := range methods {
		if string(method) == "" {
			t.Errorf("Installation method should not be empty: %v", method)
		}
	}
	
	// Test method usage in records
	record := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}
	
	if record.Method != MethodSourceBuild {
		t.Error("InstallationRecord should accept InstallationMethod values")
	}
}

func TestInstallationRecord_Fields(t *testing.T) {
	now := time.Now()
	
	record := &InstallationRecord{
		Method:              MethodCargoInstall,
		Version:             "1.2.3",
		InstalledAt:         now,
		BinaryPaths:         []string{"/usr/local/bin/tool", "/usr/local/bin/tool-alias"},
		BuildDir:            "/tmp/build/tool",
		SourceRepo:          "https://github.com/example/tool",
		Dependencies:        []string{"rust", "git"},
		InstalledByBundle:   "developer-tools",
		UserRequested:       true,
		InstallationContext: []string{"manual", "bundle:developer-tools"},
		ConfigFiles:         []string{"/home/user/.config/tool.conf"},
		SystemPackages:      []string{"libssl-dev", "pkg-config"},
	}
	
	// Verify all fields can be set and retrieved
	if record.Method != MethodCargoInstall {
		t.Error("InstallationRecord Method field not working")
	}
	
	if record.Version != "1.2.3" {
		t.Error("InstallationRecord Version field not working")
	}
	
	if !record.InstalledAt.Equal(now) {
		t.Error("InstallationRecord InstalledAt field not working")
	}
	
	if len(record.BinaryPaths) != 2 {
		t.Error("InstallationRecord BinaryPaths field not working")
	}
	
	if record.BuildDir != "/tmp/build/tool" {
		t.Error("InstallationRecord BuildDir field not working")
	}
	
	if record.SourceRepo != "https://github.com/example/tool" {
		t.Error("InstallationRecord SourceRepo field not working")
	}
	
	if len(record.Dependencies) != 2 {
		t.Error("InstallationRecord Dependencies field not working")
	}
	
	if record.InstalledByBundle != "developer-tools" {
		t.Error("InstallationRecord InstalledByBundle field not working")
	}
	
	if !record.UserRequested {
		t.Error("InstallationRecord UserRequested field not working")
	}
	
	if len(record.InstallationContext) != 2 {
		t.Error("InstallationRecord InstallationContext field not working")
	}
	
	if len(record.ConfigFiles) != 1 {
		t.Error("InstallationRecord ConfigFiles field not working")
	}
	
	if len(record.SystemPackages) != 2 {
		t.Error("InstallationRecord SystemPackages field not working")
	}
}

func TestDependencyRecord_Fields(t *testing.T) {
	now := time.Now()
	
	record := &DependencyRecord{
		InstalledBy: "gearbox",
		Version:     "stable",
		PreExisting: false,
		Dependents:  []string{"tool1", "tool2", "tool3"},
		InstallPath: "/usr/local",
		InstalledAt: now,
	}
	
	// Verify all fields can be set and retrieved
	if record.InstalledBy != "gearbox" {
		t.Error("DependencyRecord InstalledBy field not working")
	}
	
	if record.Version != "stable" {
		t.Error("DependencyRecord Version field not working")
	}
	
	if record.PreExisting {
		t.Error("DependencyRecord PreExisting field not working")
	}
	
	if len(record.Dependents) != 3 {
		t.Error("DependencyRecord Dependents field not working")
	}
	
	if record.InstallPath != "/usr/local" {
		t.Error("DependencyRecord InstallPath field not working")
	}
	
	if !record.InstalledAt.Equal(now) {
		t.Error("DependencyRecord InstalledAt field not working")
	}
}

// Test JSON serialization/deserialization with complex data
func TestComplexJSONSerialization(t *testing.T) {
	manifest := NewManifest()
	
	// Add multiple installations with different methods
	installations := map[string]*InstallationRecord{
		"ripgrep": {
			Method:              MethodCargoInstall,
			Version:             "13.0.0",
			InstalledAt:         time.Now().Truncate(time.Second),
			BinaryPaths:         []string{"/usr/local/bin/rg"},
			Dependencies:        []string{"rust"},
			UserRequested:       true,
			InstallationContext: []string{"manual"},
		},
		"fzf": {
			Method:              MethodGoInstall,
			Version:             "0.44.1",
			InstalledAt:         time.Now().Truncate(time.Second),
			BinaryPaths:         []string{"/usr/local/bin/fzf"},
			Dependencies:        []string{"go"},
			UserRequested:       false,
			InstalledByBundle:   "essential-tools",
			InstallationContext: []string{"bundle:essential-tools"},
		},
		"git": {
			Method:              MethodPreExisting,
			Version:             "2.34.1",
			InstalledAt:         time.Now().Truncate(time.Second),
			BinaryPaths:         []string{"/usr/bin/git"},
			Dependencies:        []string{},
			UserRequested:       false,
			InstallationContext: []string{"pre_existing"},
		},
	}
	
	for name, record := range installations {
		manifest.AddInstallation(name, record)
	}
	
	// Add dependencies
	dependencies := map[string]*DependencyRecord{
		"rust": {
			InstalledBy: "gearbox",
			Version:     "1.75.0",
			PreExisting: false,
			Dependents:  []string{"ripgrep", "fd", "bat"},
			InstallPath: "/usr/local",
			InstalledAt: time.Now().Truncate(time.Second),
		},
		"go": {
			InstalledBy: "gearbox",
			Version:     "1.21.5",
			PreExisting: false,
			Dependents:  []string{"fzf", "lazygit"},
			InstallPath: "/usr/local",
			InstalledAt: time.Now().Truncate(time.Second),
		},
	}
	
	for name, record := range dependencies {
		manifest.AddDependency(name, record)
	}
	
	// Serialize to JSON
	data, err := manifest.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize complex manifest: %v", err)
	}
	
	// Deserialize from JSON
	parsed, err := FromJSON(data)
	if err != nil {
		t.Fatalf("Failed to deserialize complex manifest: %v", err)
	}
	
	// Validate deserialized manifest
	if err := parsed.Validate(); err != nil {
		t.Errorf("Deserialized manifest failed validation: %v", err)
	}
	
	// Check installations
	if len(parsed.Installations) != len(manifest.Installations) {
		t.Errorf("Installation count mismatch: got %d, want %d", len(parsed.Installations), len(manifest.Installations))
	}
	
	// Check dependencies
	if len(parsed.Dependencies) != len(manifest.Dependencies) {
		t.Errorf("Dependency count mismatch: got %d, want %d", len(parsed.Dependencies), len(manifest.Dependencies))
	}
	
	// Verify specific records
	ripgrepRecord, exists := parsed.GetInstallation("ripgrep")
	if !exists {
		t.Error("ripgrep installation should exist in parsed manifest")
	} else {
		if ripgrepRecord.Method != MethodCargoInstall {
			t.Errorf("ripgrep method = %v, want %v", ripgrepRecord.Method, MethodCargoInstall)
		}
		if len(ripgrepRecord.Dependencies) != 1 || ripgrepRecord.Dependencies[0] != "rust" {
			t.Errorf("ripgrep dependencies = %v, want [rust]", ripgrepRecord.Dependencies)
		}
	}
	
	rustRecord, exists := parsed.GetDependency("rust")
	if !exists {
		t.Error("rust dependency should exist in parsed manifest")
	} else {
		if len(rustRecord.Dependents) != 3 {
			t.Errorf("rust dependents count = %d, want 3", len(rustRecord.Dependents))
		}
	}
}

// Benchmark tests
func BenchmarkNewManifest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewManifest()
	}
}

func BenchmarkAddInstallation(b *testing.B) {
	manifest := NewManifest()
	record := &InstallationRecord{
		Method:      MethodSourceBuild,
		Version:     "1.0.0",
		InstalledAt: time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manifest.AddInstallation("test-tool", record)
		// Reset for next iteration
		delete(manifest.Installations, "test-tool")
	}
}

func BenchmarkToJSON(b *testing.B) {
	manifest := NewManifest()
	
	// Add some data to make it realistic
	for i := 0; i < 10; i++ {
		record := &InstallationRecord{
			Method:      MethodSourceBuild,
			Version:     "1.0.0",
			InstalledAt: time.Now(),
		}
		manifest.AddInstallation("tool-"+string(rune(i)), record)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manifest.ToJSON()
	}
}