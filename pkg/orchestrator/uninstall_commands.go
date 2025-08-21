package orchestrator

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gearbox/pkg/uninstall"
)

// trackInstallationCmd creates the track-installation command
func trackInstallationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "track-installation TOOL_NAME METHOD VERSION [OPTIONS...]",
		Short:              "Track a tool installation in the manifest",
		Args:               cobra.MinimumNArgs(3),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleTrackInstallation(args)
		},
	}
	return cmd
}

// trackBundleCmd creates the track-bundle command
func trackBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "track-bundle BUNDLE_NAME TOOLS [OPTIONS...]",
		Short:              "Track a bundle installation in the manifest",
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleTrackBundle(args)
		},
	}
}

// isTrackedCmd creates the is-tracked command
func isTrackedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "is-tracked TOOL_NAME",
		Short: "Check if a tool is tracked in the manifest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleIsTracked(args)
		},
	}
}

// trackPreexistingCmd creates the track-preexisting command
func trackPreexistingCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "track-preexisting TOOL_NAME BINARY_PATH VERSION",
		Short:              "Track a pre-existing tool installation",
		Args:               cobra.ExactArgs(3),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleTrackPreexisting(args)
		},
	}
}

// initManifestCmd creates the init-manifest command
func initManifestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init-manifest",
		Short: "Initialize a new installation manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleInitManifest(args)
		},
	}
}

// manifestStatusCmd creates the manifest-status command
func manifestStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manifest-status",
		Short: "Show installation manifest status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleManifestStatus(args)
		},
	}
}

// listDependentsCmd creates the list-dependents command
func listDependentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-dependents TOOL_NAME",
		Short: "List tools that depend on the given tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleListDependents(args)
		},
	}
}

// canRemoveCmd creates the can-remove command
func canRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "can-remove TOOL_NAME",
		Short: "Check if a tool can be safely removed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleCanRemove(args)
		},
	}
}

// showRemovalPlan displays a removal plan to the user
func showRemovalPlan(plan *uninstall.RemovalPlan) error {
	fmt.Printf("🗑️  Removal Plan\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Show tools to be removed
	if len(plan.ToRemove) > 0 {
		fmt.Printf("📋 Tools to be removed (%d):\n", len(plan.ToRemove))
		for _, action := range plan.ToRemove {
			safety := "🟢"
			if !action.IsSafe {
				safety = "🔴"
			}
			fmt.Printf("  %s %-15s (%s) - %s\n", 
				safety, action.Target, action.Method, action.Reason)
		}
		fmt.Printf("\n")
	}

	// Show tools to be kept
	if len(plan.ToKeep) > 0 {
		fmt.Printf("🛡️  Tools to be kept (%d):\n", len(plan.ToKeep))
		for _, keep := range plan.ToKeep {
			fmt.Printf("  %-15s - %s\n", keep.Target, strings.Join(keep.Reasons, ", "))
		}
		fmt.Printf("\n")
	}

	// Show dependency actions
	if len(plan.Dependencies) > 0 {
		fmt.Printf("🔗 Dependency actions:\n")
		for _, dep := range plan.Dependencies {
			action := "preserve"
			if dep.Action != "preserve" {
				action = string(dep.Action)
			}
			fmt.Printf("  %-15s %s - %s\n", dep.Dependency, action, dep.Reason)
		}
		fmt.Printf("\n")
	}

	// Show warnings
	if len(plan.Warnings) > 0 {
		fmt.Printf("⚠️  Warnings (%d):\n", len(plan.Warnings))
		for _, warning := range plan.Warnings {
			level := "ℹ️"
			if warning.Level == "warning" {
				level = "⚠️"
			} else if warning.Level == "error" {
				level = "❌"
			}
			fmt.Printf("  %s %s: %s\n", level, warning.Target, warning.Message)
		}
		fmt.Printf("\n")
	}

	// Show summary
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  Total requested: %d\n", plan.Summary.TotalRequested)
	fmt.Printf("  Will remove: %d\n", plan.Summary.WillRemove)
	fmt.Printf("  Will keep: %d\n", plan.Summary.WillKeep)
	fmt.Printf("  Warnings: %d\n", plan.Summary.WarningCount)

	return nil
}

// showDetailedRemovalPlan displays a detailed removal plan with validation
func showDetailedRemovalPlan(plan *uninstall.RemovalPlan, engine *uninstall.RemovalEngine) error {
	if err := showRemovalPlan(plan); err != nil {
		return err
	}

	// Validate plan and show additional warnings
	validationWarnings := engine.ValidatePlan(plan)
	if len(validationWarnings) > 0 {
		fmt.Printf("\n🔍 Validation Results:\n")
		for _, warning := range validationWarnings {
			level := "ℹ️"
			if warning.Level == "warning" {
				level = "⚠️"
			} else if warning.Level == "error" {
				level = "❌"
			}
			fmt.Printf("  %s %s: %s\n", level, warning.Target, warning.Message)
		}
	}

	// Show method breakdown
	if len(plan.Summary.MethodBreakdown) > 0 {
		fmt.Printf("\n🔧 Removal methods:\n")
		for method, count := range plan.Summary.MethodBreakdown {
			description := uninstall.GetRemovalMethodDescription(method)
			fmt.Printf("  %-20s %d tools - %s\n", string(method), count, description)
		}
	}

	return nil
}