package manifest

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Tracker provides high-level operations for tracking installations
type Tracker struct {
	manager  *Manager
	manifest *InstallationManifest
}

// NewTracker creates a new installation tracker
func NewTracker() (*Tracker, error) {
	manager := NewManager()
	manifest, err := manager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}
	
	return &Tracker{
		manager:  manager,
		manifest: manifest,
	}, nil
}

// TrackInstallation records a new tool installation
func (t *Tracker) TrackInstallation(name string, config TrackingConfig) error {
	// Check if already tracked
	if _, exists := t.manifest.Installations[name]; exists {
		return fmt.Errorf("tool %s is already tracked", name)
	}
	
	// Create installation record
	record := &InstallationRecord{
		Method:              config.Method,
		Version:             config.Version,
		InstalledAt:         time.Now(),
		BinaryPaths:         config.BinaryPaths,
		BuildDir:            config.BuildDir,
		SourceRepo:          config.SourceRepo,
		Dependencies:        config.Dependencies,
		InstalledByBundle:   config.InstalledByBundle,
		UserRequested:       config.UserRequested,
		InstallationContext: config.InstallationContext,
		ConfigFiles:         config.ConfigFiles,
		SystemPackages:      config.SystemPackages,
	}
	
	// Add to manifest
	t.manifest.AddInstallation(name, record)
	
	// Track dependencies
	for _, dep := range config.Dependencies {
		if err := t.trackDependency(dep, name, config.Method); err != nil {
			return fmt.Errorf("failed to track dependency %s: %w", dep, err)
		}
	}
	
	// Save manifest
	if err := t.manager.Save(t.manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}
	
	return nil
}

// TrackBundle records a bundle installation
func (t *Tracker) TrackBundle(bundleName string, tools []string, userRequested bool) error {
	// Create bundle record
	bundleRecord := &InstallationRecord{
		Method:              MethodBundle,
		InstalledAt:         time.Now(),
		UserRequested:       userRequested,
		InstallationContext: []string{fmt.Sprintf("bundle:%s", bundleName)},
		Dependencies:        tools, // Bundle dependencies are the tools it contains
	}
	
	bundleKey := bundleName + "_bundle"
	t.manifest.AddInstallation(bundleKey, bundleRecord)
	
	// Update existing tool records to reference this bundle
	for _, tool := range tools {
		if record, exists := t.manifest.Installations[tool]; exists {
			if record.InstalledByBundle == "" {
				record.InstalledByBundle = bundleName
			}
			// Add bundle context if not already present
			bundleContext := fmt.Sprintf("bundle:%s", bundleName)
			if !contains(record.InstallationContext, bundleContext) {
				record.InstallationContext = append(record.InstallationContext, bundleContext)
			}
		}
	}
	
	// Save manifest
	return t.manager.Save(t.manifest)
}

// IsInstalled checks if a tool is tracked as installed
func (t *Tracker) IsInstalled(name string) bool {
	return t.manifest.IsInstalled(name)
}

// GetInstallation retrieves installation information
func (t *Tracker) GetInstallation(name string) (*InstallationRecord, bool) {
	return t.manifest.GetInstallation(name)
}

// GetAllInstallations returns all tracked installations
func (t *Tracker) GetAllInstallations() map[string]*InstallationRecord {
	return t.manifest.Installations
}

// DetectPreExisting checks if a tool exists before gearbox installation
func (t *Tracker) DetectPreExisting(toolName, binaryName string) (bool, string, error) {
	if binaryName == "" {
		binaryName = toolName
	}
	
	// Check if already tracked by gearbox
	if t.IsInstalled(toolName) {
		return false, "", nil
	}
	
	// Check if binary exists in PATH
	path, err := exec.LookPath(binaryName)
	if err != nil {
		return false, "", nil // Not found, not pre-existing
	}
	
	// Tool exists but not tracked by gearbox - it's pre-existing
	return true, path, nil
}

// TrackPreExisting records a pre-existing tool
func (t *Tracker) TrackPreExisting(toolName, binaryPath, version string) error {
	record := &InstallationRecord{
		Method:              MethodPreExisting,
		Version:             version,
		InstalledAt:         time.Now(),
		BinaryPaths:         []string{binaryPath},
		UserRequested:       false,
		InstallationContext: []string{"pre_existing"},
	}
	
	t.manifest.AddInstallation(toolName, record)
	return t.manager.Save(t.manifest)
}

// trackDependency handles dependency tracking
func (t *Tracker) trackDependency(depName, dependent string, method InstallationMethod) error {
	dep, exists := t.manifest.Dependencies[depName]
	if !exists {
		// Create new dependency record
		dep = &DependencyRecord{
			InstalledBy: "gearbox",
			Dependents:  []string{dependent},
			InstalledAt: time.Now(),
		}
		t.manifest.AddDependency(depName, dep)
	} else {
		// Add to dependents if not already present
		if !contains(dep.Dependents, dependent) {
			dep.Dependents = append(dep.Dependents, dependent)
		}
	}
	
	return nil
}

// TrackingConfig contains parameters for tracking an installation
type TrackingConfig struct {
	Method              InstallationMethod
	Version             string
	BinaryPaths         []string
	BuildDir            string
	SourceRepo          string
	Dependencies        []string
	InstalledByBundle   string
	UserRequested       bool
	InstallationContext []string
	ConfigFiles         []string
	SystemPackages      []string
}

// GetDependents returns tools that depend on a given dependency
func (t *Tracker) GetDependents(dependency string) []string {
	return t.manifest.GetDependents(dependency)
}

// CanSafelyRemove checks if a tool can be safely removed
func (t *Tracker) CanSafelyRemove(toolName string) (bool, []string, error) {
	record, exists := t.manifest.GetInstallation(toolName)
	if !exists {
		return false, nil, fmt.Errorf("tool %s is not tracked", toolName)
	}
	
	// Never remove pre-existing tools
	if record.Method == MethodPreExisting {
		return false, []string{"Tool was pre-existing before gearbox"}, nil
	}
	
	// Check if other tools depend on this one
	var dependents []string
	for name, otherRecord := range t.manifest.Installations {
		if name != toolName && contains(otherRecord.Dependencies, toolName) {
			dependents = append(dependents, name)
		}
	}
	
	if len(dependents) > 0 {
		reasons := []string{fmt.Sprintf("Required by other tools: %s", strings.Join(dependents, ", "))}
		return false, reasons, nil
	}
	
	return true, nil, nil
}

// CreateSnapshot creates a backup of the current manifest
func (t *Tracker) CreateSnapshot(suffix string) error {
	return t.manager.Backup(suffix)
}

// GetBinaryPaths returns all binary paths for a tool
func (t *Tracker) GetBinaryPaths(toolName string) ([]string, error) {
	record, exists := t.manifest.GetInstallation(toolName)
	if !exists {
		return nil, fmt.Errorf("tool %s is not tracked", toolName)
	}
	
	return record.BinaryPaths, nil
}

// GetBuildDir returns the build directory for a tool
func (t *Tracker) GetBuildDir(toolName string) (string, error) {
	record, exists := t.manifest.GetInstallation(toolName)
	if !exists {
		return "", fmt.Errorf("tool %s is not tracked", toolName)
	}
	
	return record.BuildDir, nil
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// DetectBinaryPaths automatically detects binary paths for a tool
func DetectBinaryPaths(toolName string, aliases []string) []string {
	var paths []string
	
	// Check main tool name
	if path, err := exec.LookPath(toolName); err == nil {
		paths = append(paths, path)
	}
	
	// Check aliases
	for _, alias := range aliases {
		if path, err := exec.LookPath(alias); err == nil {
			paths = append(paths, path)
		}
	}
	
	return paths
}

// GetInstallationStats returns statistics about tracked installations
func (t *Tracker) GetInstallationStats() map[string]int {
	stats := make(map[string]int)
	
	// Count by method
	for _, record := range t.manifest.Installations {
		stats[string(record.Method)]++
	}
	
	// Total count
	stats["total"] = len(t.manifest.Installations)
	
	// Bundle count
	bundleCount := 0
	for name := range t.manifest.Installations {
		if strings.HasSuffix(name, "_bundle") {
			bundleCount++
		}
	}
	stats["bundles"] = bundleCount
	stats["tools"] = stats["total"] - bundleCount
	
	return stats
}