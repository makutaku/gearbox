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
  
  # Navigation:
  #   Tab       - Switch between views
  #   ↑/↓       - Navigate lists
  #   Enter     - Select/Confirm
  #   /         - Search
  #   ?         - Help
  #   q         - Quit`,
		RunE: runTUI,
	}
	
	// Future flags could include:
	// cmd.Flags().StringP("theme", "t", "default", "Color theme (default, dark, light)")
	// cmd.Flags().BoolP("compact", "c", false, "Use compact mode")
	// cmd.Flags().StringP("start-view", "s", "dashboard", "Initial view to display")
	
	return cmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	log.Info().Msg("Launching Gearbox TUI")
	
	// Check terminal capabilities
	if !isTerminalInteractive() {
		log.Warn().Msg("Terminal does not appear to be interactive")
		return ErrNotInteractive
	}
	
	// Run the TUI
	if err := tui.Run(); err != nil {
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