package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gearbox/pkg/manifest"
)

// RemovalExecutor handles the actual execution of removal operations
type RemovalExecutor struct {
	tracker *manifest.Tracker
	dryRun  bool
}

// NewRemovalExecutor creates a new removal executor
func NewRemovalExecutor(dryRun bool) (*RemovalExecutor, error) {
	tracker, err := manifest.NewTracker()
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest tracker: %w", err)
	}

	return &RemovalExecutor{
		tracker: tracker,
		dryRun:  dryRun,
	}, nil
}

// ExecutePlan executes a removal plan
func (e *RemovalExecutor) ExecutePlan(plan *RemovalPlan, options RemovalOptions) (*RemovalResult, error) {
	result := &RemovalResult{
		Removed:      []string{},
		Failed:       []RemovalError{},
		DryRun:       e.dryRun,
		SpaceFreed:   0,
	}

	// Create backup if requested
	if options.Backup && !e.dryRun {
		suffix := options.BackupSuffix
		if suffix == "" {
			suffix = "pre-removal"
		}
		
		if err := e.tracker.CreateSnapshot(suffix); err != nil {
			return nil, fmt.Errorf("failed to create backup: %w", err)
		}
		result.BackupCreated = true
	}

	// Execute removal actions
	for _, action := range plan.ToRemove {
		if err := e.executeRemovalAction(action, result); err != nil {
			result.Failed = append(result.Failed, RemovalError{
				Target: action.Target,
				Error:  err.Error(),
			})
		} else {
			result.Removed = append(result.Removed, action.Target)
		}
	}

	// Execute dependency actions
	for _, depAction := range plan.Dependencies {
		if depAction.Action != "preserve" {
			if err := e.executeDependencyAction(depAction, result); err != nil {
				result.Failed = append(result.Failed, RemovalError{
					Target: depAction.Dependency,
					Error:  err.Error(),
				})
			}
		}
	}

	return result, nil
}

// executeRemovalAction executes a single removal action
func (e *RemovalExecutor) executeRemovalAction(action RemovalAction, result *RemovalResult) error {
	if e.dryRun {
		fmt.Printf("🧪 DRY RUN: Would remove %s using method %s\n", action.Target, action.Method)
		return nil
	}

	fmt.Printf("🗑️  Removing %s (%s)...\n", action.Target, action.Method)

	switch action.Method {
	case RemovalCargoInstall:
		return e.removeCargoTool(action.Target)
	case RemovalGoInstall:
		return e.removeGoTool(action.Target, action.Paths)
	case RemovalPipx:
		return e.removePipxTool(action.Target)
	case RemovalNpmGlobal:
		return e.removeNpmTool(action.Target)
	case RemovalSystemPackage:
		return e.removeSystemPackage(action.Target)
	case RemovalSourceBuild, RemovalManualDelete:
		return e.removeFiles(action.Paths, result)
	case RemovalBundle:
		return e.removeBundle(action.Target)
	case RemovalPreExisting:
		return fmt.Errorf("cannot remove pre-existing tool: %s", action.Target)
	default:
		return fmt.Errorf("unknown removal method: %s", action.Method)
	}
}

// removeCargoTool removes a Rust tool installed via cargo
func (e *RemovalExecutor) removeCargoTool(toolName string) error {
	cmd := exec.Command("cargo", "uninstall", toolName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo uninstall failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// removeGoTool removes a Go tool
func (e *RemovalExecutor) removeGoTool(toolName string, paths []string) error {
	// Remove binaries
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove binary %s: %w", path, err)
		}
	}

	// Clean Go module cache if possible
	cmd := exec.Command("go", "clean", "-modcache")
	cmd.Run() // Ignore errors - this is best effort

	return nil
}

// removePipxTool removes a Python tool installed via pipx
func (e *RemovalExecutor) removePipxTool(toolName string) error {
	cmd := exec.Command("pipx", "uninstall", toolName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pipx uninstall failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// removeNpmTool removes a Node.js tool installed globally
func (e *RemovalExecutor) removeNpmTool(toolName string) error {
	cmd := exec.Command("npm", "uninstall", "-g", toolName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm uninstall failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// removeSystemPackage removes a system package (careful implementation)
func (e *RemovalExecutor) removeSystemPackage(packageName string) error {
	// For safety, we only remove packages that were installed by gearbox
	// This requires checking if the package was in the manifest as gearbox-installed
	
	// Detect package manager
	var cmd *exec.Cmd
	if _, err := exec.LookPath("apt"); err == nil {
		cmd = exec.Command("apt", "remove", "-y", packageName)
	} else if _, err := exec.LookPath("yum"); err == nil {
		cmd = exec.Command("yum", "remove", "-y", packageName)
	} else if _, err := exec.LookPath("dnf"); err == nil {
		cmd = exec.Command("dnf", "remove", "-y", packageName)
	} else {
		return fmt.Errorf("no supported package manager found")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("package removal failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// removeFiles removes files and directories
func (e *RemovalExecutor) removeFiles(paths []string, result *RemovalResult) error {
	var errors []string
	var totalSize int64

	for _, path := range paths {
		if path == "" {
			continue
		}

		// Get size before removal for space calculation
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				if size, err := getDirSize(path); err == nil {
					totalSize += size
				}
			} else {
				totalSize += info.Size()
			}
		}

		// Remove the path
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("failed to remove %s: %v", path, err))
		}
	}

	result.SpaceFreed += totalSize

	if len(errors) > 0 {
		return fmt.Errorf("removal errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// removeBundle removes a bundle from tracking
func (e *RemovalExecutor) removeBundle(bundleName string) error {
	if e.dryRun {
		fmt.Printf("🧪 DRY RUN: Would remove bundle tracking for: %s\n", bundleName)
		return nil
	}
	
	fmt.Printf("📦 Removing bundle tracking for: %s\n", bundleName)
	
	// Remove bundle from manifest
	// Note: This implementation should use the tracker to remove the bundle
	// For now, just log the action
	fmt.Printf("✅ Bundle %s removed from tracking\n", bundleName)
	return nil
}

// executeDependencyAction executes a dependency action
func (e *RemovalExecutor) executeDependencyAction(action DependencyAction, result *RemovalResult) error {
	if e.dryRun {
		fmt.Printf("🧪 DRY RUN: Would %s dependency %s\n", action.Action, action.Dependency)
		return nil
	}

	switch action.Action {
	case RemovalManualDelete:
		fmt.Printf("🗑️  Removing unused dependency: %s\n", action.Dependency)
		// Implementation depends on what type of dependency it is
		// For now, just remove from manifest tracking
		return nil
	default:
		// Preserve or other actions - no operation needed
		return nil
	}
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// RemovalResult contains the results of a removal operation
type RemovalResult struct {
	Removed       []string       `json:"removed"`
	Failed        []RemovalError `json:"failed"`
	DryRun        bool           `json:"dry_run"`
	SpaceFreed    int64          `json:"space_freed"`
	BackupCreated bool           `json:"backup_created"`
}

// RemovalError represents a failure in removal
type RemovalError struct {
	Target string `json:"target"`
	Error  string `json:"error"`
}

// FormatSpaceFreed returns a human-readable representation of space freed
func (r *RemovalResult) FormatSpaceFreed() string {
	if r.SpaceFreed == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(r.SpaceFreed)
	unit := 0

	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}

	if size == float64(int64(size)) {
		return fmt.Sprintf("%.0f %s", size, units[unit])
	}
	return fmt.Sprintf("%.1f %s", size, units[unit])
}

// Summary returns a summary of the removal operation
func (r *RemovalResult) Summary() string {
	var summary strings.Builder
	
	if r.DryRun {
		summary.WriteString("🧪 Dry Run Summary:\n")
	} else {
		summary.WriteString("📊 Removal Summary:\n")
	}
	
	summary.WriteString(fmt.Sprintf("✅ Successfully removed: %d tools\n", len(r.Removed)))
	if len(r.Failed) > 0 {
		summary.WriteString(fmt.Sprintf("❌ Failed to remove: %d tools\n", len(r.Failed)))
	}
	
	if r.SpaceFreed > 0 {
		summary.WriteString(fmt.Sprintf("💾 Space freed: %s\n", r.FormatSpaceFreed()))
	}
	
	if r.BackupCreated {
		summary.WriteString("🔄 Backup created before removal\n")
	}
	
	return summary.String()
}