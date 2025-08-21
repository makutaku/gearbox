package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ManifestDir is the directory where gearbox stores state
	ManifestDir = ".gearbox"
	// ManifestFile is the name of the manifest file
	ManifestFile = "manifest.json"
	// BackupDir is where manifest backups are stored
	BackupDir = "backups"
)

// Manager handles manifest file operations
type Manager struct {
	manifestPath string
	backupDir    string
}

// NewManager creates a new manifest manager
func NewManager() *Manager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir unavailable
		homeDir = "."
	}
	
	manifestDir := filepath.Join(homeDir, ManifestDir)
	manifestPath := filepath.Join(manifestDir, ManifestFile)
	backupDir := filepath.Join(manifestDir, BackupDir)
	
	return &Manager{
		manifestPath: manifestPath,
		backupDir:    backupDir,
	}
}

// GetManifestPath returns the path to the manifest file
func (m *Manager) GetManifestPath() string {
	return m.manifestPath
}

// EnsureManifestDir creates the manifest directory if it doesn't exist
func (m *Manager) EnsureManifestDir() error {
	manifestDir := filepath.Dir(m.manifestPath)
	return os.MkdirAll(manifestDir, 0755)
}

// EnsureBackupDir creates the backup directory if it doesn't exist
func (m *Manager) EnsureBackupDir() error {
	return os.MkdirAll(m.backupDir, 0755)
}

// Load reads and parses the manifest file
func (m *Manager) Load() (*InstallationManifest, error) {
	// Check if manifest file exists
	if _, err := os.Stat(m.manifestPath); os.IsNotExist(err) {
		// Create new manifest if file doesn't exist
		manifest := NewManifest()
		if err := m.Save(manifest); err != nil {
			return nil, fmt.Errorf("failed to create new manifest: %w", err)
		}
		return manifest, nil
	}
	
	// Read existing manifest
	data, err := os.ReadFile(m.manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}
	
	manifest, err := FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	// Validate manifest
	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	
	return manifest, nil
}

// Save writes the manifest to disk
func (m *Manager) Save(manifest *InstallationManifest) error {
	// Ensure directory exists
	if err := m.EnsureManifestDir(); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}
	
	// Update timestamp
	manifest.UpdatedAt = time.Now()
	
	// Serialize to JSON
	data, err := manifest.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize manifest: %w", err)
	}
	
	// Write to temporary file first
	tempPath := m.manifestPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary manifest: %w", err)
	}
	
	// Atomic move to final location
	if err := os.Rename(tempPath, m.manifestPath); err != nil {
		return fmt.Errorf("failed to move manifest to final location: %w", err)
	}
	
	return nil
}

// Backup creates a backup of the current manifest
func (m *Manager) Backup(suffix string) error {
	// Ensure backup directory exists
	if err := m.EnsureBackupDir(); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Check if manifest exists
	if _, err := os.Stat(m.manifestPath); os.IsNotExist(err) {
		return nil // Nothing to backup
	}
	
	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("manifest-%s", timestamp)
	if suffix != "" {
		backupName += "-" + suffix
	}
	backupName += ".json"
	
	backupPath := filepath.Join(m.backupDir, backupName)
	
	// Copy manifest to backup location
	data, err := os.ReadFile(m.manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest for backup: %w", err)
	}
	
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}
	
	return nil
}

// ListBackups returns a list of available backup files
func (m *Manager) ListBackups() ([]string, error) {
	if _, err := os.Stat(m.backupDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	files, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	var backups []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			backups = append(backups, file.Name())
		}
	}
	
	return backups, nil
}

// RestoreBackup restores a manifest from a backup file
func (m *Manager) RestoreBackup(backupName string) error {
	backupPath := filepath.Join(m.backupDir, backupName)
	
	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupName)
	}
	
	// Create backup of current manifest before restore
	if err := m.Backup("pre-restore"); err != nil {
		return fmt.Errorf("failed to backup current manifest: %w", err)
	}
	
	// Copy backup to manifest location
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}
	
	// Validate backup before restoring
	manifest, err := FromJSON(data)
	if err != nil {
		return fmt.Errorf("backup file is corrupted: %w", err)
	}
	
	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("backup file is invalid: %w", err)
	}
	
	// Write to manifest location
	if err := os.WriteFile(m.manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore manifest: %w", err)
	}
	
	return nil
}

// Exists checks if the manifest file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.manifestPath)
	return !os.IsNotExist(err)
}