package orchestrator

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gearbox/pkg/uninstall"
)

// installCmd creates the install command
func installCmd() *cobra.Command {
	var opts InstallationOptions

	cmd := &cobra.Command{
		Use:   "install [tools...]",
		Short: "Install tools with advanced orchestration",
		Long: `Install one or more tools with dependency resolution, parallel execution,
and comprehensive progress tracking. If no tools are specified, all tools will be installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(opts)
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			var toolsToInstall []string
			if len(args) == 0 {
				// Install all tools
				for _, tool := range orchestrator.config.Tools {
					toolsToInstall = append(toolsToInstall, tool.Name)
				}
			} else {
				toolsToInstall = args
			}

			return orchestrator.InstallTools(toolsToInstall)
		},
	}

	// Installation options
	cmd.Flags().StringVarP(&opts.BuildType, "build-type", "b", "standard", "Build type (minimal, standard, maximum)")
	cmd.Flags().BoolVar(&opts.SkipCommonDeps, "skip-common-deps", false, "Skip common dependency installation")
	cmd.Flags().BoolVar(&opts.RunTests, "run-tests", false, "Run test suites for validation")
	cmd.Flags().BoolVar(&opts.NoShell, "no-shell", false, "Skip shell integration setup")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force reinstallation if already installed")
	cmd.Flags().IntVarP(&opts.MaxParallelJobs, "jobs", "j", 0, "Maximum parallel jobs (0 = auto-detect)")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be installed without executing")

	// Nerd-fonts specific options
	cmd.Flags().StringVar(&opts.Fonts, "fonts", "", "Install specific fonts (comma-separated, e.g. 'FiraCode,JetBrainsMono')")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", false, "Interactive font selection with previews")
	cmd.Flags().BoolVar(&opts.Preview, "preview", false, "Show font previews before installation")
	cmd.Flags().BoolVar(&opts.ConfigureApps, "configure-apps", false, "Automatically configure VS Code, terminals, etc.")

	return cmd
}

// listCmd creates the list command
func listCmd() *cobra.Command {
	var category string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list [bundles]",
		Short: "List available tools or bundles",
		Long: `List available tools or bundles.
		
Without arguments, lists all available tools.
Use 'list bundles' to list all available bundles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}
			
			// Check if user wants to list bundles
			if len(args) > 0 && args[0] == "bundles" {
				return orchestrator.ListBundles(verbose)
			}

			return orchestrator.ListTools(category, verbose)
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	return cmd
}

// showCmd creates the show command
func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show bundle <name>",
		Short: "Show details about a bundle",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "bundle" {
				return fmt.Errorf("only 'show bundle' is supported")
			}
			
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}
			
			return orchestrator.ShowBundle(args[1])
		},
	}
	
	return cmd
}

// statusCmd creates the status command
func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [tools...]",
		Short: "Show installation status of tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			manifestOnly, _ := cmd.Flags().GetBool("manifest-only")
			unified, _ := cmd.Flags().GetBool("unified")
			
			return orchestrator.ShowStatus(args, manifestOnly, unified)
		},
	}
	
	cmd.Flags().Bool("manifest-only", false, "Show only manifest-tracked tools")
	cmd.Flags().Bool("unified", false, "Show unified view (manifest + live detection)")
	cmd.Flags().Bool("all", false, "Show status for all tools (default)")
	cmd.Flags().Bool("installed", false, "Show only installed tools")
	cmd.Flags().Bool("missing", false, "Show only missing tools")
	
	return cmd
}

// verifyCmd creates the verify command
func verifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify [tools...]",
		Short: "Verify tool installations",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			return orchestrator.VerifyTools(args)
		},
	}
}

// doctorCmd creates the doctor command
func doctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor [tool]",
		Short: "Run health checks and diagnostics",
		RunE: func(cmd *cobra.Command, args []string) error {
			orchestrator, err := NewOrchestrator(InstallationOptions{})
			if err != nil {
				return fmt.Errorf("failed to initialize orchestrator: %w", err)
			}

			return orchestrator.RunDoctor(args)
		},
	}
	
	cmd.Flags().Bool("fix", false, "Attempt to fix detected issues automatically")
	return cmd
}

// uninstallCmd creates the uninstall command
func uninstallCmd() *cobra.Command {
	var opts uninstall.RemovalOptions

	cmd := &cobra.Command{
		Use:   "uninstall [tools...]",
		Short: "Uninstall tools with safe removal",
		Long: `Uninstall one or more tools with dependency analysis and safe removal.
Analyzes dependencies and provides a removal plan before execution.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no tools specified for removal")
			}

			// Create removal engine with standard safety level
			engine, err := uninstall.NewRemovalEngine(uninstall.SafetyStandard)
			if err != nil {
				return fmt.Errorf("failed to create removal engine: %w", err)
			}

			// Plan removal
			plan, err := engine.PlanRemoval(args, opts)
			if err != nil {
				return fmt.Errorf("failed to plan removal: %w", err)
			}

			// Show plan and get confirmation
			if err := showRemovalPlan(plan); err != nil {
				return err
			}

			if !opts.DryRun {
				fmt.Printf("\nProceed with removal? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Printf("❌ Removal cancelled\n")
					return nil
				}
			}

			// Execute removal
			executor, err := uninstall.NewRemovalExecutor(opts.DryRun)
			if err != nil {
				return fmt.Errorf("failed to create removal executor: %w", err)
			}

			result, err := executor.ExecutePlan(plan, opts)
			if err != nil {
				return fmt.Errorf("failed to execute removal: %w", err)
			}

			// Show results
			fmt.Printf("\n%s", result.Summary())
			return nil
		},
	}

	// Removal options
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force removal even if there are dependents")
	cmd.Flags().BoolVar(&opts.Cascade, "cascade", false, "Remove unused dependencies")
	cmd.Flags().BoolVar(&opts.RemoveConfig, "remove-config", false, "Remove configuration files")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be removed without executing")
	cmd.Flags().BoolVar(&opts.Backup, "backup", true, "Create backup before removal")
	cmd.Flags().StringVar(&opts.BackupSuffix, "backup-suffix", "", "Suffix for backup files")
	cmd.Flags().BoolVar(&opts.RemoveBundleContents, "bundle-contents", false, "Remove all tools in bundle, not just bundle tracking")

	return cmd
}

// uninstallPlanCmd creates the uninstall-plan command
func uninstallPlanCmd() *cobra.Command {
	var safetyLevel string

	cmd := &cobra.Command{
		Use:   "uninstall-plan [tools...]",
		Short: "Show removal plan without executing",
		Long: `Analyze what would be removed for the specified tools.
Shows dependency analysis and safety warnings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no tools specified for analysis")
			}

			// Parse safety level
			var safety uninstall.SafetyLevel
			switch strings.ToLower(safetyLevel) {
			case "conservative":
				safety = uninstall.SafetyConservative
			case "standard":
				safety = uninstall.SafetyStandard
			case "aggressive":
				safety = uninstall.SafetyAggressive
			default:
				safety = uninstall.SafetyStandard
			}

			// Create removal engine
			engine, err := uninstall.NewRemovalEngine(safety)
			if err != nil {
				return fmt.Errorf("failed to create removal engine: %w", err)
			}

			// Plan removal
			opts := uninstall.RemovalOptions{DryRun: true}
			plan, err := engine.PlanRemoval(args, opts)
			if err != nil {
				return fmt.Errorf("failed to plan removal: %w", err)
			}

			// Show detailed plan
			return showDetailedRemovalPlan(plan, engine)
		},
	}

	cmd.Flags().StringVar(&safetyLevel, "safety", "standard", "Safety level (conservative, standard, aggressive)")

	return cmd
}

// uninstallExecuteCmd creates the uninstall-execute command
func uninstallExecuteCmd() *cobra.Command {
	var opts uninstall.RemovalOptions

	cmd := &cobra.Command{
		Use:   "uninstall-execute [tools...]",
		Short: "Execute removal without confirmation",
		Long: `Execute tool removal without interactive confirmation.
Use with caution - this bypasses safety prompts.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no tools specified for removal")
			}

			// Create removal engine with aggressive safety for non-interactive use
			engine, err := uninstall.NewRemovalEngine(uninstall.SafetyAggressive)
			if err != nil {
				return fmt.Errorf("failed to create removal engine: %w", err)
			}

			// Plan removal
			plan, err := engine.PlanRemoval(args, opts)
			if err != nil {
				return fmt.Errorf("failed to plan removal: %w", err)
			}

			// Execute removal immediately
			executor, err := uninstall.NewRemovalExecutor(opts.DryRun)
			if err != nil {
				return fmt.Errorf("failed to create removal executor: %w", err)
			}

			fmt.Printf("🗑️  Executing removal of %d tools...\n", len(plan.ToRemove))

			result, err := executor.ExecutePlan(plan, opts)
			if err != nil {
				return fmt.Errorf("failed to execute removal: %w", err)
			}

			// Show results
			fmt.Printf("\n%s", result.Summary())
			return nil
		},
	}

	// Removal options
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force removal even if there are dependents")
	cmd.Flags().BoolVar(&opts.Cascade, "cascade", false, "Remove unused dependencies")
	cmd.Flags().BoolVar(&opts.RemoveConfig, "remove-config", false, "Remove configuration files")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be removed without executing")
	cmd.Flags().BoolVar(&opts.Backup, "backup", true, "Create backup before removal")
	cmd.Flags().StringVar(&opts.BackupSuffix, "backup-suffix", "", "Suffix for backup files")

	return cmd
}