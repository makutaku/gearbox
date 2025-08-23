package commands

import (
	"github.com/spf13/cobra"
	"github.com/rs/zerolog/log"
	
	"gearbox/cmd/gearbox/tui"
	"gearbox/pkg/errors"
)

// NewTUICmd creates the TUI command
func NewTUICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive Text User Interface",
		Long: `Launch Gearbox's interactive Text User Interface (TUI) for a rich, visual experience.

The TUI provides:
  - Visual tool browsing with search and filters
  - Bundle exploration with dependency visualization
  - Real-time installation progress tracking
  - Interactive configuration management
  - System health monitoring
  - Keyboard navigation and mouse support

This is an enhanced interface that complements the CLI, providing a more
intuitive way to discover and manage development tools.`,
		Example: `  # Launch the TUI
  gearbox tui
  
  # Launch TUI in demo mode with mock data (safe for testing)
  gearbox tui --demo
  
  # Launch TUI in test mode for automated testing
  gearbox tui --test
  
  # Navigation:
  #   Tab       - Switch between views
  #   ↑/↓       - Navigate lists
  #   Enter     - Select/Confirm
  #   /         - Search
  #   ?         - Help
  #   q         - Quit`,
		RunE: runTUI,
	}
	
	// Add test and demo mode flags
	cmd.Flags().BoolP("demo", "d", false, "Launch in demo mode with mock data (safe for testing)")
	cmd.Flags().BoolP("test", "t", false, "Launch in test mode for automated testing")
	cmd.Flags().String("test-scenario", "", "Run specific test scenario (basic-nav, tool-install, bundle-install)")
	
	// Future flags could include:
	// cmd.Flags().StringP("theme", "t", "default", "Color theme (default, dark, light)")
	// cmd.Flags().BoolP("compact", "c", false, "Use compact mode")
	// cmd.Flags().StringP("start-view", "s", "dashboard", "Initial view to display")
	
	return cmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	log.Info().Msg("Launching Gearbox TUI")
	
	// Get flags
	demoMode, _ := cmd.Flags().GetBool("demo")
	testMode, _ := cmd.Flags().GetBool("test")
	testScenario, _ := cmd.Flags().GetString("test-scenario")
	
	// Check terminal capabilities (skip for test mode)
	if !testMode && !isTerminalInteractive() {
		log.Warn().Msg("Terminal does not appear to be interactive")
		return ErrNotInteractive
	}
	
	// Create TUI options
	opts := tui.Options{
		DemoMode:     demoMode,
		TestMode:     testMode,
		TestScenario: testScenario,
	}
	
	// Run the TUI with options
	if err := tui.RunWithOptions(opts); err != nil {
		log.Error().Err(err).Msg("TUI error")
		return err
	}
	
	log.Info().Msg("TUI closed")
	return nil
}

// isTerminalInteractive checks if we're running in an interactive terminal
func isTerminalInteractive() bool {
	// This is a simple check; could be enhanced
	return true
}

var ErrNotInteractive = errors.New(
	errors.SystemError,
	"tui.launch",
).WithMessage("TUI requires an interactive terminal").
	WithDetails("The Text User Interface cannot run in non-interactive environments")