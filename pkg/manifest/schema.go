package manifest

import (
	"encoding/json"
	"time"
)

// SchemaVersion defines the current manifest schema version
const SchemaVersion = "1.0"

// InstallationManifest represents the complete installation state
type InstallationManifest struct {
	SchemaVersion string                           `json:"schema_version"`
	Installations map[string]*InstallationRecord  `json:"installations"`
	Dependencies  map[string]*DependencyRecord    `json:"dependencies"`
	CreatedAt     time.Time                       `json:"created_at"`
	UpdatedAt     time.Time                       `json:"updated_at"`
}

// InstallationRecord tracks a single tool installation
type InstallationRecord struct {
	Method           InstallationMethod `json:"method"`
	Version          string             `json:"version"`
	InstalledAt      time.Time          `json:"installed_at"`
	BinaryPaths      []string           `json:"binary_paths"`
	BuildDir         string             `json:"build_dir,omitempty"`
	SourceRepo       string             `json:"source_repo,omitempty"`
	Dependencies     []string           `json:"dependencies"`
	InstalledByBundle string            `json:"installed_by_bundle,omitempty"`
	UserRequested    bool               `json:"user_requested"`
	InstallationContext []string        `json:"installation_context"`
	ConfigFiles      []string           `json:"config_files,omitempty"`
	SystemPackages   []string           `json:"system_packages,omitempty"`
}

// DependencyRecord tracks shared dependencies
type DependencyRecord struct {
	InstalledBy   string    `json:"installed_by"`
	Version       string    `json:"version"`
	PreExisting   bool      `json:"pre_existing"`
	Dependents    []string  `json:"dependents"`
	InstallPath   string    `json:"install_path,omitempty"`
	InstalledAt   time.Time `json:"installed_at"`
}

// InstallationMethod represents how a tool was installed
type InstallationMethod string

const (
	MethodSourceBuild  InstallationMethod = "source_build"
	MethodCargoInstall InstallationMethod = "cargo_install"
	MethodGoInstall    InstallationMethod = "go_install"
	MethodSystemPackage InstallationMethod = "system_package"
	MethodPipx         InstallationMethod = "pipx"
	MethodNpmGlobal    InstallationMethod = "npm_global"
	MethodManualDownload InstallationMethod = "manual_download"
	MethodBundle       InstallationMethod = "bundle"
	MethodPreExisting  InstallationMethod = "pre_existing"
)

// NewManifest creates a new empty manifest
func NewManifest() *InstallationManifest {
	now := time.Now()
	return &InstallationManifest{
		SchemaVersion: SchemaVersion,
		Installations: make(map[string]*InstallationRecord),
		Dependencies:  make(map[string]*DependencyRecord),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddInstallation adds a new installation record
func (m *InstallationManifest) AddInstallation(name string, record *InstallationRecord) {
	m.Installations[name] = record
	m.UpdatedAt = time.Now()
}

// AddDependency adds a new dependency record
func (m *InstallationManifest) AddDependency(name string, record *DependencyRecord) {
	m.Dependencies[name] = record
	m.UpdatedAt = time.Now()
}

// GetInstallation retrieves an installation record
func (m *InstallationManifest) GetInstallation(name string) (*InstallationRecord, bool) {
	record, exists := m.Installations[name]
	return record, exists
}

// GetDependency retrieves a dependency record
func (m *InstallationManifest) GetDependency(name string) (*DependencyRecord, bool) {
	record, exists := m.Dependencies[name]
	return record, exists
}

// IsInstalled checks if a tool is installed by gearbox
func (m *InstallationManifest) IsInstalled(name string) bool {
	_, exists := m.Installations[name]
	return exists
}

// GetDependents returns tools that depend on a given dependency
func (m *InstallationManifest) GetDependents(dependency string) []string {
	if dep, exists := m.Dependencies[dependency]; exists {
		return dep.Dependents
	}
	return []string{}
}

// ToJSON serializes the manifest to JSON
func (m *InstallationManifest) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// FromJSON deserializes the manifest from JSON
func FromJSON(data []byte) (*InstallationManifest, error) {
	var manifest InstallationManifest
	err := json.Unmarshal(data, &manifest)
	return &manifest, err
}

// Validate checks if the manifest is valid
func (m *InstallationManifest) Validate() error {
	if m.SchemaVersion != SchemaVersion {
		return &ValidationError{
			Message: "Unsupported schema version: " + m.SchemaVersion,
		}
	}
	return nil
}

// ValidationError represents a manifest validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "Manifest validation error: " + e.Message
}