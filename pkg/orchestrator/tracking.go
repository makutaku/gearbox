package orchestrator

import (
	"fmt"
	"os"
	"strings"

	"gearbox/pkg/manifest"
)

// handleTrackInstallation processes track-installation command
func handleTrackInstallation(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: track-installation TOOL_NAME METHOD VERSION [OPTIONS...]")
	}

	toolName := args[0]
	method := manifest.InstallationMethod(args[1])
	version := args[2]

	// Parse options
	config := manifest.TrackingConfig{
		Method:              method,
		Version:             version,
		UserRequested:       true,
		InstallationContext: []string{},
	}

	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--binary-paths":
			if i+1 < len(args) {
				config.BinaryPaths = strings.Split(args[i+1], ",")
				i++
			}
		case "--build-dir":
			if i+1 < len(args) {
				config.BuildDir = args[i+1]
				i++
			}
		case "--source-repo":
			if i+1 < len(args) {
				config.SourceRepo = args[i+1]
				i++
			}
		case "--dependencies":
			if i+1 < len(args) {
				config.Dependencies = strings.Split(args[i+1], ",")
				i++
			}
		case "--installed-by-bundle":
			if i+1 < len(args) {
				config.InstalledByBundle = args[i+1]
				i++
			}
		case "--config-files":
			if i+1 < len(args) {
				config.ConfigFiles = strings.Split(args[i+1], ",")
				i++
			}
		case "--system-packages":
			if i+1 < len(args) {
				config.SystemPackages = strings.Split(args[i+1], ",")
				i++
			}
		case "--not-user-requested":
			config.UserRequested = false
		}
	}

	// Add installation context
	if config.InstalledByBundle != "" {
		config.InstallationContext = append(config.InstallationContext, "bundle:"+config.InstalledByBundle)
	}
	if config.UserRequested {
		config.InstallationContext = append(config.InstallationContext, "user_request")
	}

	// Create tracker and track installation
	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	if err := tracker.TrackInstallation(toolName, config); err != nil {
		return fmt.Errorf("failed to track installation: %w", err)
	}

	fmt.Printf("Successfully tracked installation: %s\n", toolName)
	return nil
}

// handleTrackBundle processes track-bundle command
func handleTrackBundle(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: track-bundle BUNDLE_NAME TOOLS [OPTIONS...]")
	}

	bundleName := args[0]
	toolsStr := args[1]
	tools := strings.Split(toolsStr, ",")

	userRequested := true
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--not-user-requested":
			userRequested = false
		}
	}

	// Create tracker and track bundle
	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	if err := tracker.TrackBundle(bundleName, tools, userRequested); err != nil {
		return fmt.Errorf("failed to track bundle: %w", err)
	}

	fmt.Printf("Successfully tracked bundle: %s\n", bundleName)
	return nil
}

// handleIsTracked checks if a tool is tracked
func handleIsTracked(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: is-tracked TOOL_NAME")
	}

	toolName := args[0]

	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	if tracker.IsInstalled(toolName) {
		fmt.Printf("true\n")
		os.Exit(0)
	} else {
		fmt.Printf("false\n")
		os.Exit(1)
	}

	return nil
}

// handleTrackPreexisting tracks a pre-existing tool
func handleTrackPreexisting(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: track-preexisting TOOL_NAME BINARY_PATH VERSION")
	}

	toolName := args[0]
	binaryPath := args[1]
	version := args[2]

	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	if err := tracker.TrackPreExisting(toolName, binaryPath, version); err != nil {
		return fmt.Errorf("failed to track pre-existing tool: %w", err)
	}

	fmt.Printf("Successfully tracked pre-existing tool: %s\n", toolName)
	return nil
}

// handleInitManifest initializes a new manifest
func handleInitManifest(args []string) error {
	manager := manifest.NewManager()
	newManifest := manifest.NewManifest()

	if err := manager.Save(newManifest); err != nil {
		return fmt.Errorf("failed to initialize manifest: %w", err)
	}

	fmt.Printf("Successfully initialized manifest at: %s\n", manager.GetManifestPath())
	return nil
}

// handleManifestStatus shows manifest status
func handleManifestStatus(args []string) error {
	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	stats := tracker.GetInstallationStats()
	installations := tracker.GetAllInstallations()

	fmt.Printf("📋 Installation Manifest Status\n")
	fmt.Printf("Total installations: %d\n", stats["total"])
	fmt.Printf("Tools: %d\n", stats["tools"])
	fmt.Printf("Bundles: %d\n", stats["bundles"])
	fmt.Printf("\n")

	fmt.Printf("📊 By installation method:\n")
	for method, count := range stats {
		if method != "total" && method != "tools" && method != "bundles" {
			fmt.Printf("  %s: %d\n", method, count)
		}
	}
	fmt.Printf("\n")

	fmt.Printf("🔧 Tracked installations:\n")
	for name, record := range installations {
		if !strings.HasSuffix(name, "_bundle") {
			userReq := ""
			if record.UserRequested {
				userReq = " (user requested)"
			}
			bundle := ""
			if record.InstalledByBundle != "" {
				bundle = fmt.Sprintf(" [bundle: %s]", record.InstalledByBundle)
			}
			fmt.Printf("  ✅ %s (%s)%s%s\n", name, record.Method, userReq, bundle)
		}
	}

	// Show bundles
	fmt.Printf("\n📦 Tracked bundles:\n")
	for name, record := range installations {
		if strings.HasSuffix(name, "_bundle") {
			bundleName := strings.TrimSuffix(name, "_bundle")
			userReq := ""
			if record.UserRequested {
				userReq = " (user requested)"
			}
			fmt.Printf("  📦 %s%s\n", bundleName, userReq)
		}
	}

	return nil
}

// handleListDependents shows what depends on a tool/dependency
func handleListDependents(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: list-dependents TOOL_NAME")
	}

	toolName := args[0]

	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	dependents := tracker.GetDependents(toolName)
	if len(dependents) == 0 {
		fmt.Printf("No dependents found for: %s\n", toolName)
		return nil
	}

	fmt.Printf("Tools that depend on %s:\n", toolName)
	for _, dependent := range dependents {
		fmt.Printf("  - %s\n", dependent)
	}

	return nil
}

// handleCanRemove checks if a tool can be safely removed
func handleCanRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: can-remove TOOL_NAME")
	}

	toolName := args[0]

	tracker, err := manifest.NewTracker()
	if err != nil {
		return fmt.Errorf("failed to create tracker: %w", err)
	}

	canRemove, reasons, err := tracker.CanSafelyRemove(toolName)
	if err != nil {
		return fmt.Errorf("failed to check removal safety: %w", err)
	}

	if canRemove {
		fmt.Printf("✅ %s can be safely removed\n", toolName)
		os.Exit(0)
	} else {
		fmt.Printf("❌ %s cannot be safely removed:\n", toolName)
		for _, reason := range reasons {
			fmt.Printf("  - %s\n", reason)
		}
		os.Exit(1)
	}

	return nil
}