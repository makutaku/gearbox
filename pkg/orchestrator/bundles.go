package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BundleConfig represents a tool bundle configuration
type BundleConfig struct {
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Category        string              `json:"category"`
	Tools           []string            `json:"tools"`
	SystemPackages  []string            `json:"system_packages"`
	PackageManagers map[string][]string `json:"package_managers"`
	IncludesBundles []string            `json:"includes_bundles"`
	Tags            []string            `json:"tags"`
}

// BundleConfiguration represents the complete bundle configuration file
type BundleConfiguration struct {
	SchemaVersion string         `json:"schema_version"`
	Bundles       []BundleConfig `json:"bundles"`
}

// loadBundles loads bundle configuration from bundles.json
func (o *Orchestrator) loadBundles() (*BundleConfiguration, error) {
	bundlesPath := filepath.Join(o.repoDir, "config", "bundles.json")
	
	file, err := os.Open(bundlesPath)
	if err != nil {
		// Bundles are optional, so return empty config if file doesn't exist
		if os.IsNotExist(err) {
			return &BundleConfiguration{
				SchemaVersion: "1.0",
				Bundles:       []BundleConfig{},
			}, nil
		}
		return nil, fmt.Errorf("failed to open bundles.json: %w", err)
	}
	defer file.Close()

	var bundleConfig BundleConfiguration
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&bundleConfig); err != nil {
		return nil, fmt.Errorf("failed to decode bundles.json: %w", err)
	}

	return &bundleConfig, nil
}

// findBundle finds a bundle by name
func (o *Orchestrator) findBundle(name string, bundles []BundleConfig) (*BundleConfig, bool) {
	for _, bundle := range bundles {
		if bundle.Name == name {
			return &bundle, true
		}
	}
	return nil, false
}

// expandBundle recursively expands a bundle into a list of tool names
func (o *Orchestrator) expandBundle(bundleName string, bundles []BundleConfig, visited map[string]bool) ([]string, error) {
	// Check for circular dependencies
	if visited[bundleName] {
		return nil, fmt.Errorf("circular dependency detected in bundle: %s", bundleName)
	}
	visited[bundleName] = true

	bundle, found := o.findBundle(bundleName, bundles)
	if !found {
		return nil, fmt.Errorf("bundle not found: %s", bundleName)
	}

	var tools []string
	
	// First, expand included bundles
	for _, includedBundle := range bundle.IncludesBundles {
		// Create a new visited map for each recursive call to track this path
		pathVisited := make(map[string]bool)
		for k, v := range visited {
			pathVisited[k] = v
		}
		
		expandedTools, err := o.expandBundle(includedBundle, bundles, pathVisited)
		if err != nil {
			return nil, fmt.Errorf("failed to expand included bundle %s: %w", includedBundle, err)
		}
		tools = append(tools, expandedTools...)
	}
	
	// Then add direct tools
	tools = append(tools, bundle.Tools...)
	
	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	var uniqueTools []string
	for _, tool := range tools {
		if !seen[tool] {
			seen[tool] = true
			uniqueTools = append(uniqueTools, tool)
		}
	}
	
	return uniqueTools, nil
}

// isBundle checks if a given name is a bundle
func (o *Orchestrator) isBundle(name string, bundles []BundleConfig) bool {
	_, found := o.findBundle(name, bundles)
	return found
}

// expandBundlesAndTools takes a list of mixed bundle and tool names and returns expanded tool list
func (o *Orchestrator) expandBundlesAndTools(names []string) ([]string, error) {
	var bundleConfig *BundleConfiguration
	var err error
	
	// Use already loaded config if available, otherwise load it
	if o.bundleConfig != nil {
		bundleConfig = o.bundleConfig
	} else {
		bundleConfig, err = o.loadBundles()
		if err != nil {
			return nil, fmt.Errorf("failed to load bundles: %w", err)
		}
	}

	var allTools []string
	
	for _, name := range names {
		// Check if it's a bundle
		if o.isBundle(name, bundleConfig.Bundles) {
			visited := make(map[string]bool)
			expandedTools, err := o.expandBundle(name, bundleConfig.Bundles, visited)
			if err != nil {
				return nil, fmt.Errorf("failed to expand bundle %s: %w", name, err)
			}
			allTools = append(allTools, expandedTools...)
		} else {
			// It's a regular tool
			allTools = append(allTools, name)
		}
	}
	
	// Remove duplicates while preserving order
	seen := make(map[string]bool)
	var uniqueTools []string
	for _, tool := range allTools {
		if !seen[tool] {
			seen[tool] = true
			uniqueTools = append(uniqueTools, tool)
		}
	}
	
	return uniqueTools, nil
}

// getBundleInfo returns formatted information about a bundle
func (o *Orchestrator) getBundleInfo(bundleName string) (string, error) {
	bundleConfig, err := o.loadBundles()
	if err != nil {
		return "", fmt.Errorf("failed to load bundles: %w", err)
	}

	bundle, found := o.findBundle(bundleName, bundleConfig.Bundles)
	if !found {
		return "", fmt.Errorf("bundle not found: %s", bundleName)
	}

	// Expand the bundle to get all tools
	visited := make(map[string]bool)
	tools, err := o.expandBundle(bundleName, bundleConfig.Bundles, visited)
	if err != nil {
		return "", fmt.Errorf("failed to expand bundle: %w", err)
	}

	// Get system packages
	var systemPackages []string
	var managerName string
	if o.packageMgr != nil {
		managerName = o.packageMgr.Name
		visitedPkg := make(map[string]bool)
		systemPackages, _ = o.expandSystemPackages(bundleName, bundleConfig.Bundles, visitedPkg, managerName)
	}

	info := fmt.Sprintf("Bundle: %s\n", bundle.Name)
	info += fmt.Sprintf("Description: %s\n", bundle.Description)
	info += fmt.Sprintf("Category: %s\n", bundle.Category)
	if len(bundle.Tags) > 0 {
		info += fmt.Sprintf("Tags: %v\n", bundle.Tags)
	}
	if len(bundle.IncludesBundles) > 0 {
		info += fmt.Sprintf("Includes bundles: %v\n", bundle.IncludesBundles)
	}
	
	info += fmt.Sprintf("Total tools: %d\n", len(tools))
	if len(systemPackages) > 0 {
		info += fmt.Sprintf("System packages: %d (%s)\n", len(systemPackages), managerName)
	}
	
	info += "Tools:\n"
	for _, tool := range tools {
		info += fmt.Sprintf("  - %s\n", tool)
	}
	
	if len(systemPackages) > 0 {
		info += fmt.Sprintf("System packages (%s):\n", managerName)
		for _, pkg := range systemPackages {
			info += fmt.Sprintf("  - %s\n", pkg)
		}
	}

	return info, nil
}

// ListBundles lists all available bundles
func (o *Orchestrator) ListBundles(verbose bool) error {
	bundleConfig, err := o.loadBundles()
	if err != nil {
		return fmt.Errorf("failed to load bundles: %w", err)
	}

	if len(bundleConfig.Bundles) == 0 {
		fmt.Println("No bundles available.")
		return nil
	}

	fmt.Printf("📦 Available Bundles (%d)\n\n", len(bundleConfig.Bundles))

	// Group bundles by category
	categories := make(map[string][]BundleConfig)
	for _, bundle := range bundleConfig.Bundles {
		categories[bundle.Category] = append(categories[bundle.Category], bundle)
	}

	// Sort categories for consistent output
	var catNames []string
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	// Display bundles by category
	for _, category := range catNames {
		fmt.Printf("🔧 %s\n", strings.Title(category))
		bundles := categories[category]
		
		// Sort bundles by name
		sort.Slice(bundles, func(i, j int) bool {
			return bundles[i].Name < bundles[j].Name
		})

		for _, bundle := range bundles {
			fmt.Printf("  • %-20s - %s\n", bundle.Name, bundle.Description)
			
			if verbose {
				// Show tools count
				visited := make(map[string]bool)
				tools, _ := o.expandBundle(bundle.Name, bundleConfig.Bundles, visited)
				fmt.Printf("    Tools: %d", len(tools))
				
				if len(bundle.IncludesBundles) > 0 {
					fmt.Printf(" (includes: %s)", strings.Join(bundle.IncludesBundles, ", "))
				}
				
				if len(bundle.Tags) > 0 {
					fmt.Printf("\n    Tags: %s", strings.Join(bundle.Tags, ", "))
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}

	return nil
}

// ShowBundle displays detailed information about a specific bundle
func (o *Orchestrator) ShowBundle(bundleName string) error {
	info, err := o.getBundleInfo(bundleName)
	if err != nil {
		return err
	}
	
	fmt.Println(info)
	return nil
}