package uninstall

import (
	"fmt"
	"strings"

	"gearbox/pkg/manifest"
)

// SafetyLevel defines the level of safety for removal operations
type SafetyLevel int

const (
	SafetyConservative SafetyLevel = iota // Maximum safety - preserve everything possible
	SafetyStandard                        // Standard safety - reasonable defaults
	SafetyAggressive                      // Minimal safety - remove more aggressively
)

// RemovalMethod defines how a tool should be removed
type RemovalMethod string

const (
	RemovalSourceBuild  RemovalMethod = "source_build"
	RemovalCargoInstall RemovalMethod = "cargo_uninstall"
	RemovalGoInstall    RemovalMethod = "go_clean"
	RemovalSystemPackage RemovalMethod = "system_uninstall"
	RemovalPipx         RemovalMethod = "pipx_uninstall"
	RemovalNpmGlobal    RemovalMethod = "npm_uninstall"
	RemovalManualDelete RemovalMethod = "manual_delete"
	RemovalBundle       RemovalMethod = "bundle_remove"
	RemovalPreExisting  RemovalMethod = "preserve" // Never remove pre-existing
)

// RemovalAction represents a single removal operation
type RemovalAction struct {
	Target      string        `json:"target"`
	Method      RemovalMethod `json:"method"`
	Paths       []string      `json:"paths"`
	Reason      string        `json:"reason"`
	Dependencies []string     `json:"dependencies"`
	IsSafe      bool          `json:"is_safe"`
}

// KeepReason explains why a tool should not be removed
type KeepReason struct {
	Target  string   `json:"target"`
	Reasons []string `json:"reasons"`
}

// SafetyWarning represents a potential issue with removal
type SafetyWarning struct {
	Target  string `json:"target"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// DependencyAction represents an action on a dependency
type DependencyAction struct {
	Dependency string        `json:"dependency"`
	Action     RemovalMethod `json:"action"`
	Reason     string        `json:"reason"`
	Affected   []string      `json:"affected"`
}

// RemovalPlan contains all information about what will be removed
type RemovalPlan struct {
	ToRemove     []RemovalAction    `json:"to_remove"`
	ToKeep       []KeepReason       `json:"to_keep"`
	Warnings     []SafetyWarning    `json:"warnings"`
	Dependencies []DependencyAction `json:"dependencies"`
	Summary      RemovalSummary     `json:"summary"`
}

// RemovalSummary provides high-level statistics about the removal plan
type RemovalSummary struct {
	TotalRequested     int                       `json:"total_requested"`
	WillRemove         int                       `json:"will_remove"`
	WillKeep           int                       `json:"will_keep"`
	WarningCount       int                       `json:"warning_count"`
	MethodBreakdown    map[RemovalMethod]int     `json:"method_breakdown"`
	DependencyActions  map[string]int            `json:"dependency_actions"`
	EstimatedSpaceFreed string                   `json:"estimated_space_freed"`
}

// RemovalEngine handles the analysis and planning of tool removal
type RemovalEngine struct {
	tracker     *manifest.Tracker
	safetyLevel SafetyLevel
}

// NewRemovalEngine creates a new removal engine
func NewRemovalEngine(safetyLevel SafetyLevel) (*RemovalEngine, error) {
	tracker, err := manifest.NewTracker()
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest tracker: %w", err)
	}

	return &RemovalEngine{
		tracker:     tracker,
		safetyLevel: safetyLevel,
	}, nil
}

// PlanRemoval analyzes the requested removals and creates a safe removal plan
func (r *RemovalEngine) PlanRemoval(targets []string, options RemovalOptions) (*RemovalPlan, error) {
	plan := &RemovalPlan{
		ToRemove:     []RemovalAction{},
		ToKeep:       []KeepReason{},
		Warnings:     []SafetyWarning{},
		Dependencies: []DependencyAction{},
	}

	// Process each target
	for _, target := range targets {
		if err := r.analyzeTarget(target, plan, options); err != nil {
			return nil, fmt.Errorf("failed to analyze target %s: %w", target, err)
		}
	}

	// Analyze dependencies
	if err := r.analyzeDependencies(plan, options); err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	// Generate summary
	r.generateSummary(plan)

	return plan, nil
}

// analyzeTarget analyzes a single target for removal
func (r *RemovalEngine) analyzeTarget(target string, plan *RemovalPlan, options RemovalOptions) error {
	// Check if target is a bundle first
	if r.isBundle(target) {
		return r.analyzeBundleTarget(target, plan, options)
	}
	
	// Check if target is tracked
	record, exists := r.tracker.GetInstallation(target)
	if !exists {
		plan.Warnings = append(plan.Warnings, SafetyWarning{
			Target:  target,
			Level:   "info",
			Message: "Tool is not tracked by gearbox - may not be installed or is pre-existing",
		})
		return nil
	}

	// Never remove pre-existing tools
	if record.Method == manifest.MethodPreExisting {
		plan.ToKeep = append(plan.ToKeep, KeepReason{
			Target:  target,
			Reasons: []string{"Tool was pre-existing before gearbox installation"},
		})
		return nil
	}

	// Check if removal is safe
	canRemove, reasons, err := r.tracker.CanSafelyRemove(target)
	if err != nil {
		return fmt.Errorf("failed to check removal safety: %w", err)
	}

	if !canRemove && !options.Force {
		plan.ToKeep = append(plan.ToKeep, KeepReason{
			Target:  target,
			Reasons: reasons,
		})
		return nil
	}

	// If forced removal with dependents, add warnings
	if !canRemove && options.Force {
		plan.Warnings = append(plan.Warnings, SafetyWarning{
			Target:  target,
			Level:   "warning",
			Message: "Forcing removal despite dependencies: " + strings.Join(reasons, ", "),
		})
	}

	// Create removal action
	action := RemovalAction{
		Target:       target,
		Method:       r.getRemovalMethod(record.Method),
		Paths:        record.BinaryPaths,
		Dependencies: record.Dependencies,
		IsSafe:       canRemove,
		Reason:       "User requested removal",
	}

	// Add build directory if it exists
	if record.BuildDir != "" {
		action.Paths = append(action.Paths, record.BuildDir)
	}

	// Add config files if removing
	if options.RemoveConfig && len(record.ConfigFiles) > 0 {
		action.Paths = append(action.Paths, record.ConfigFiles...)
	}

	plan.ToRemove = append(plan.ToRemove, action)
	return nil
}

// analyzeDependencies analyzes shared dependencies and determines actions
func (r *RemovalEngine) analyzeDependencies(plan *RemovalPlan, options RemovalOptions) error {
	dependencyUsage := make(map[string][]string) // dependency -> list of tools using it
	
	// Collect all dependencies from tools being removed
	var removingTools []string
	for _, action := range plan.ToRemove {
		removingTools = append(removingTools, action.Target)
		
		for _, dep := range action.Dependencies {
			dependencyUsage[dep] = r.tracker.GetDependents(dep)
		}
	}

	// Analyze each dependency
	for dependency, dependents := range dependencyUsage {
		var remainingDependents []string
		
		// Find dependents that are NOT being removed
		for _, dependent := range dependents {
			isBeingRemoved := false
			for _, removingTool := range removingTools {
				if dependent == removingTool {
					isBeingRemoved = true
					break
				}
			}
			if !isBeingRemoved {
				remainingDependents = append(remainingDependents, dependent)
			}
		}

		var action DependencyAction
		action.Dependency = dependency
		action.Affected = dependents

		if len(remainingDependents) == 0 {
			// No remaining dependents - safe to remove if cascade enabled
			if options.Cascade {
				action.Action = RemovalManualDelete
				action.Reason = "No remaining dependents - removing with cascade"
			} else {
				action.Action = "preserve"
				action.Reason = "No remaining dependents but cascade not enabled - keeping"
			}
		} else {
			// Has remaining dependents - preserve
			action.Action = "preserve"
			action.Reason = fmt.Sprintf("Still needed by: %s", strings.Join(remainingDependents, ", "))
		}

		plan.Dependencies = append(plan.Dependencies, action)
	}

	return nil
}

// getRemovalMethod maps installation method to removal method
func (r *RemovalEngine) getRemovalMethod(installMethod manifest.InstallationMethod) RemovalMethod {
	switch installMethod {
	case manifest.MethodSourceBuild:
		return RemovalSourceBuild
	case manifest.MethodCargoInstall:
		return RemovalCargoInstall
	case manifest.MethodGoInstall:
		return RemovalGoInstall
	case manifest.MethodSystemPackage:
		return RemovalSystemPackage
	case manifest.MethodPipx:
		return RemovalPipx
	case manifest.MethodNpmGlobal:
		return RemovalNpmGlobal
	case manifest.MethodManualDownload:
		return RemovalManualDelete
	case manifest.MethodBundle:
		return RemovalBundle
	case manifest.MethodPreExisting:
		return RemovalPreExisting
	default:
		return RemovalManualDelete
	}
}

// generateSummary creates a summary of the removal plan
func (r *RemovalEngine) generateSummary(plan *RemovalPlan) {
	summary := RemovalSummary{
		TotalRequested:    len(plan.ToRemove) + len(plan.ToKeep),
		WillRemove:        len(plan.ToRemove),
		WillKeep:          len(plan.ToKeep),
		WarningCount:      len(plan.Warnings),
		MethodBreakdown:   make(map[RemovalMethod]int),
		DependencyActions: make(map[string]int),
	}

	// Count removal methods
	for _, action := range plan.ToRemove {
		summary.MethodBreakdown[action.Method]++
	}

	// Count dependency actions
	for _, dep := range plan.Dependencies {
		summary.DependencyActions[string(dep.Action)]++
	}

	plan.Summary = summary
}

// RemovalOptions configures how removal should be performed
type RemovalOptions struct {
	Force               bool   // Force removal even if there are dependents
	Cascade             bool   // Remove unused dependencies
	RemoveConfig        bool   // Remove configuration files
	DryRun              bool   // Only plan, don't execute
	Backup              bool   // Create backup before removal
	BackupSuffix        string // Suffix for backup files
	RemoveBundleContents bool   // Remove all tools in bundle, not just bundle tracking
}

// ValidatePlan checks if a removal plan is safe to execute
func (r *RemovalEngine) ValidatePlan(plan *RemovalPlan) []SafetyWarning {
	var warnings []SafetyWarning

	// Check for forced removals
	for _, action := range plan.ToRemove {
		if !action.IsSafe {
			warnings = append(warnings, SafetyWarning{
				Target:  action.Target,
				Level:   "error",
				Message: "Forced removal may break other tools",
			})
		}
	}

	// Check for dependency removals
	for _, dep := range plan.Dependencies {
		if dep.Action != "preserve" && len(dep.Affected) > 1 {
			warnings = append(warnings, SafetyWarning{
				Target:  dep.Dependency,
				Level:   "warning",
				Message: fmt.Sprintf("Removing shared dependency used by %d tools", len(dep.Affected)),
			})
		}
	}

	// Check safety level constraints
	switch r.safetyLevel {
	case SafetyConservative:
		// In conservative mode, warn about any removal
		for _, action := range plan.ToRemove {
			warnings = append(warnings, SafetyWarning{
				Target:  action.Target,
				Level:   "info",
				Message: "Conservative mode: double-check removal is necessary",
			})
		}
	case SafetyAggressive:
		// In aggressive mode, warn if we're keeping too much
		if plan.Summary.WillKeep > plan.Summary.WillRemove {
			warnings = append(warnings, SafetyWarning{
				Target:  "general",
				Level:   "info",
				Message: "Aggressive mode: consider using --force to remove more tools",
			})
		}
	}

	return warnings
}

// GetRemovalMethodDescription returns a human-readable description of a removal method
func GetRemovalMethodDescription(method RemovalMethod) string {
	switch method {
	case RemovalSourceBuild:
		return "Remove binaries and build artifacts from source installation"
	case RemovalCargoInstall:
		return "Use 'cargo uninstall' to remove Rust tool"
	case RemovalGoInstall:
		return "Remove Go tool binary and clean module cache"
	case RemovalSystemPackage:
		return "Use system package manager to uninstall"
	case RemovalPipx:
		return "Use 'pipx uninstall' to remove Python tool"
	case RemovalNpmGlobal:
		return "Use 'npm uninstall -g' to remove Node.js tool"
	case RemovalManualDelete:
		return "Manually delete files and directories"
	case RemovalBundle:
		return "Remove bundle tracking and optionally contained tools"
	case RemovalPreExisting:
		return "Preserve pre-existing installation"
	default:
		return "Unknown removal method"
	}
}

// isBundle checks if a target is a bundle
func (r *RemovalEngine) isBundle(target string) bool {
	// Check if the target is tracked as a bundle
	record, exists := r.tracker.GetInstallation(target)
	if exists && record.Method == manifest.MethodBundle {
		return true
	}
	
	// Check if it's a known bundle suffix
	if strings.HasSuffix(target, "_bundle") || strings.HasSuffix(target, "-bundle") {
		return true
	}
	
	return false
}

// analyzeBundleTarget analyzes a bundle for removal
func (r *RemovalEngine) analyzeBundleTarget(bundleName string, plan *RemovalPlan, options RemovalOptions) error {
	// Get bundle record
	record, exists := r.tracker.GetInstallation(bundleName)
	if !exists {
		plan.Warnings = append(plan.Warnings, SafetyWarning{
			Target:  bundleName,
			Level:   "info",
			Message: "Bundle is not tracked by gearbox",
		})
		return nil
	}
	
	if record.Method != manifest.MethodBundle {
		plan.Warnings = append(plan.Warnings, SafetyWarning{
			Target:  bundleName,
			Level:   "warning",
			Message: "Target is not a bundle",
		})
		return nil
	}
	
	// Add bundle removal action
	action := RemovalAction{
		Target:       bundleName,
		Method:       RemovalBundle,
		Paths:        []string{}, // Bundles don't have file paths
		Dependencies: record.Dependencies,
		IsSafe:       true, // Bundle removal is generally safe
		Reason:       "User requested bundle removal",
	}
	plan.ToRemove = append(plan.ToRemove, action)
	
	// If removing bundle tools is requested, analyze each tool in the bundle
	if options.RemoveBundleContents {
		for _, toolName := range record.Dependencies {
			if err := r.analyzeTarget(toolName, plan, options); err != nil {
				return fmt.Errorf("failed to analyze bundle tool %s: %w", toolName, err)
			}
		}
	} else {
		// Just removing bundle tracking, warn about contained tools
		if len(record.Dependencies) > 0 {
			plan.Warnings = append(plan.Warnings, SafetyWarning{
				Target:  bundleName,
				Level:   "info",
				Message: fmt.Sprintf("Bundle contains %d tools that will remain installed: %s", 
					len(record.Dependencies), strings.Join(record.Dependencies, ", ")),
			})
		}
	}
	
	return nil
}