package orchestrator

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// PackageManager represents a system package manager
type PackageManager struct {
	Name          string
	InstallCmd    []string
	CheckCmd      []string
	UpdateCmd     []string
	Available     bool
}

// SystemPackage represents a system package to be installed
type SystemPackage struct {
	Name        string
	Manager     string
	Installed   bool
	Available   bool
}

// detectPackageManager detects the available package manager on the system
func detectPackageManager() (*PackageManager, error) {
	if runtime.GOOS != "linux" {
		return nil, fmt.Errorf("system package installation only supported on Linux")
	}

	managers := []PackageManager{
		{
			Name:       "apt",
			InstallCmd: []string{"apt-get", "install", "-y"},
			CheckCmd:   []string{"dpkg", "-l"},
			UpdateCmd:  []string{"apt-get", "update"},
		},
		{
			Name:       "yum",
			InstallCmd: []string{"yum", "install", "-y"},
			CheckCmd:   []string{"rpm", "-qa"},
			UpdateCmd:  []string{"yum", "update"},
		},
		{
			Name:       "dnf",
			InstallCmd: []string{"dnf", "install", "-y"},
			CheckCmd:   []string{"rpm", "-qa"},
			UpdateCmd:  []string{"dnf", "update"},
		},
	}

	for _, mgr := range managers {
		if _, err := exec.LookPath(mgr.InstallCmd[0]); err == nil {
			mgr.Available = true
			return &mgr, nil
		}
	}

	return nil, fmt.Errorf("no supported package manager found (apt, yum, dnf)")
}

// isPackageInstalled checks if a system package is installed
func (pm *PackageManager) isPackageInstalled(packageName string) (bool, error) {
	var cmd *exec.Cmd
	
	switch pm.Name {
	case "apt":
		cmd = exec.Command("dpkg", "-l", packageName)
	case "yum", "dnf":
		cmd = exec.Command("rpm", "-q", packageName)
	default:
		return false, fmt.Errorf("unsupported package manager: %s", pm.Name)
	}
	
	err := cmd.Run()
	return err == nil, nil
}

// installPackages installs system packages using the detected package manager
func (pm *PackageManager) installPackages(packages []string, dryRun bool) error {
	if len(packages) == 0 {
		return nil
	}

	if dryRun {
		fmt.Printf("📦 Would install system packages (%s): %s\n", pm.Name, strings.Join(packages, ", "))
		return nil
	}

	fmt.Printf("📦 Installing system packages via %s: %s\n", pm.Name, strings.Join(packages, ", "))
	
	// Update package lists first
	if len(pm.UpdateCmd) > 0 {
		updateCmd := exec.Command(pm.UpdateCmd[0], pm.UpdateCmd[1:]...)
		if err := updateCmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: Failed to update package lists: %v\n", err)
		}
	}
	
	// Install packages
	args := append(pm.InstallCmd[1:], packages...)
	installCmd := exec.Command(pm.InstallCmd[0], args...)
	
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages %v: %w", packages, err)
	}
	
	fmt.Printf("✅ Successfully installed system packages: %s\n", strings.Join(packages, ", "))
	return nil
}

// getSystemPackagesForManager returns the appropriate package list for the current package manager
func (b *BundleConfig) getSystemPackagesForManager(managerName string) []string {
	// First check if there's a manager-specific list
	if packages, exists := b.PackageManagers[managerName]; exists {
		return packages
	}
	
	// Fall back to generic system_packages
	return b.SystemPackages
}

// expandSystemPackages collects all system packages from a bundle and its included bundles
func (o *Orchestrator) expandSystemPackages(bundleName string, bundles []BundleConfig, visited map[string]bool, managerName string) ([]string, error) {
	// Check for circular dependencies
	if visited[bundleName] {
		return nil, fmt.Errorf("circular dependency detected in bundle: %s", bundleName)
	}
	visited[bundleName] = true

	bundle, found := o.findBundle(bundleName, bundles)
	if !found {
		return nil, fmt.Errorf("bundle not found: %s", bundleName)
	}

	var packages []string
	
	// First, expand included bundles
	for _, includedBundle := range bundle.IncludesBundles {
		pathVisited := make(map[string]bool)
		for k, v := range visited {
			pathVisited[k] = v
		}
		
		expandedPackages, err := o.expandSystemPackages(includedBundle, bundles, pathVisited, managerName)
		if err != nil {
			return nil, fmt.Errorf("failed to expand system packages from included bundle %s: %w", includedBundle, err)
		}
		packages = append(packages, expandedPackages...)
	}
	
	// Then add direct system packages
	packages = append(packages, bundle.getSystemPackagesForManager(managerName)...)
	
	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	var uniquePackages []string
	for _, pkg := range packages {
		if !seen[pkg] {
			seen[pkg] = true
			uniquePackages = append(uniquePackages, pkg)
		}
	}
	
	return uniquePackages, nil
}