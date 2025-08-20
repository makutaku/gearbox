package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [bundles]",
		Aliases: []string{"ls"},
		Short:   "Show available tools or bundles with descriptions",
		Long: `Display a comprehensive list of all available development tools or bundles.

Usage:
  gearbox list              # List all available tools
  gearbox list bundles      # List all available bundles

The list includes:
- Tool/bundle names and categories
- Brief descriptions
- Language/technology used (for tools)
- Installation status (if orchestrator is available)`,
		RunE: runList,
	}

	cmd.Flags().BoolP("installed", "i", false, "Show only installed tools")
	cmd.Flags().BoolP("available", "a", false, "Show only available (not installed) tools")
	cmd.Flags().StringP("category", "c", "", "Filter by category (core, navigation, media, etc.)")
	cmd.Flags().BoolP("verbose", "v", false, "Show detailed information")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	// Get the directory where the gearbox binary is located
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	
	repoDir := filepath.Dir(execPath)
	orchestratorPath := filepath.Join(repoDir, "orchestrator")

	// Check if the advanced orchestrator is available
	if _, err := os.Stat(orchestratorPath); err == nil {
		// Use the orchestrator for enhanced listing
		cmdArgs := []string{"list"}
		
		// Pass through arguments (e.g., "bundles")
		if len(args) > 0 {
			cmdArgs = append(cmdArgs, args...)
		}
		
		orchestratorCmd := exec.Command(orchestratorPath, cmdArgs...)
		
		// Pass through any flags
		if installed, _ := cmd.Flags().GetBool("installed"); installed {
			orchestratorCmd.Args = append(orchestratorCmd.Args, "--installed")
		}
		if available, _ := cmd.Flags().GetBool("available"); available {
			orchestratorCmd.Args = append(orchestratorCmd.Args, "--available")
		}
		if category, _ := cmd.Flags().GetString("category"); category != "" {
			orchestratorCmd.Args = append(orchestratorCmd.Args, "--category", category)
		}
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			orchestratorCmd.Args = append(orchestratorCmd.Args, "--verbose")
		}

		orchestratorCmd.Stdout = os.Stdout
		orchestratorCmd.Stderr = os.Stderr
		
		return orchestratorCmd.Run()
	}

	// Fallback to built-in tool list
	return showBuiltinToolList()
}

func showBuiltinToolList() error {
	fmt.Print(`Available Tools:

Core Development Tools:
  fd          Fast file finder (Rust)
              Alternative to 'find' with intuitive syntax and parallel search
              
  ripgrep     Fast text search (Rust)
              High-performance grep replacement with PCRE2 and multi-line support
              
  fzf         Fuzzy finder (Go)
              Interactive file/command finder with shell integration
              
  jq          JSON processor (C)
              Command-line JSON processor with powerful query capabilities

Navigation & File Management:
  zoxide      Smart cd command (Rust)
              Smarter directory navigation with frecency (frequency + recency)
              
  yazi        Terminal file manager (Rust)
              Fast terminal file manager with vim-like keybindings and preview
              
  fclones     Duplicate file finder (Rust)
              Efficient tool to find, remove, and deduplicate identical files
              
  serena      Coding agent toolkit (Python)
              Semantic retrieval and editing capabilities for codebases
              
  uv          Python package manager (Rust)
              Extremely fast Python package and project manager
              
  ruff        Python linter & formatter (Rust)
              10-100x faster than Flake8/Black, 800+ lint rules
              
  bat         Enhanced cat with syntax highlighting (Rust)
              Cat clone with wings - Git integration, themes, paging
              
  starship    Customizable shell prompt (Rust)
              Fast, minimal prompt with contextual information
              
  eza         Modern ls replacement (Rust)
              Enhanced file listings with Git integration and tree view
              
  delta       Syntax-highlighting pager (Rust)
              Enhanced Git diff and output with word-level highlighting
              
  lazygit     Terminal UI for Git (Go)
              Interactive Git operations with visual interface
              
  bottom      Cross-platform system monitor (Rust)
              Beautiful terminal-based system resource monitoring
              
  procs       Modern ps replacement (Rust)
              Enhanced process information with tree view and colors
              
  tokei       Code statistics tool (Rust)
              Fast line counting for 200+ programming languages
              
  difftastic  Structural diff tool (Rust)
              Syntax-aware diffing for better code change analysis
              
  bandwhich   Network bandwidth monitor (Rust)
              Terminal bandwidth utilization by process
              
  xsv         CSV data toolkit (Rust)
              Fast CSV processing and analysis
              
  hyperfine   Command-line benchmarking (Rust)
              Statistical analysis of command execution times
              
  gh          GitHub CLI (Go)
              Repository management, PRs, issues, and workflows
              
  dust        Better disk usage analyzer (Rust)
              Intuitive visualization of directory sizes
              
  sd          Find & replace CLI (Rust)
              Intuitive alternative to sed for text replacement
              
  tealdeer    Fast tldr client (Rust)
              Quick command help without full man pages
              
  choose      Cut/awk alternative (Rust)
              Human-friendly text column selection

Media & Image Processing:
  ffmpeg      Video/audio processing (C/C++)
              Comprehensive media processing suite with extensive codec support
              
  imagemagick Image manipulation (C/C++)
              Powerful image processing and manipulation toolkit
              
  7zip        Compression tool (C/C++)
              High-ratio compression tool with multiple format support

Usage Examples:
  gearbox install fd ripgrep fzf       # Install core development tools
  gearbox install ffmpeg               # Install just ffmpeg
  gearbox install --minimal fd ripgrep # Fast builds
  gearbox install --maximum ffmpeg     # Full-featured build

For detailed tool documentation, see: docs/USER_GUIDE.md
`)

	return nil
}