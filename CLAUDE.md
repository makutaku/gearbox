# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Essential Tools Installer - a collection of automated installation scripts for essential command-line tools on Debian Linux. The project focuses on building tools from source with various optimization levels and proper dependency management.

## Common Commands

### Installation Commands
- `gearbox install` - Show confirmation prompt and install all 31 tools
- `gearbox install fd ripgrep fzf` - Install only specified tools (recommended approach)
- `gearbox install --minimal fd ripgrep` - Install with minimal/fast builds  
- `gearbox install --maximum ffmpeg` - Install with full-featured builds
- `gearbox install nerd-fonts` - Install standard font collection (8 fonts) with cross-tool suggestions
- `gearbox install nerd-fonts --fonts="FiraCode"` - Install specific font via CLI
- `gearbox install nerd-fonts --interactive` - Interactive font selection with previews

### Bundle Installation (NEW!)

**Language Ecosystem Bundles:**
- `gearbox install --bundle python-ecosystem` - Python runtime + pipx, black, flake8, mypy, poetry, pytest, jupyter + essential tools
- `gearbox install --bundle nodejs-ecosystem` - Node.js runtime + TypeScript, ESLint, Angular/Vue/React CLIs, jest + essential tools
- `gearbox install --bundle rust-ecosystem` - Rust compiler + rustfmt, clippy, rust-analyzer, cargo tools + essential tools
- `gearbox install --bundle go-ecosystem` - Go compiler + gopls, golangci-lint, air, staticcheck, delve + essential tools
- `gearbox install --bundle java-ecosystem` - Java 17 + Maven, Gradle + essential tools
- `gearbox install --bundle ruby-ecosystem` - Ruby runtime + Rails, RSpec, RuboCop, Solargraph + essential tools
- `gearbox install --bundle cpp-ecosystem` - GCC/Clang + CMake, Ninja, GDB, Valgrind, Conan, vcpkg + essential tools

**AI-Powered Development:**
- `gearbox install --bundle ai-coding-agent` - Serena MCP server for AI-assisted coding with semantic code analysis (multi-language)

**Core Development Bundles:**
- `gearbox install --bundle essential` - Install curated bundle of essential tools
- `gearbox install --bundle developer` - Complete development environment
- `gearbox install --bundle quickstart` - Recommended starter bundle

**Infrastructure & DevOps Bundles:**
- `gearbox install --bundle web-dev` - Web development (nginx, nodejs + tools)  
- `gearbox install --bundle docker-dev` - Docker development environment
- `gearbox install --bundle database-admin` - Database administration tools
- `gearbox install --bundle netadmin` - Network administration and monitoring toolkit

**Specialized Bundles:**
- `gearbox list bundles` - Show available bundles with descriptions
- `gearbox show bundle web-dev` - Show bundle contents including system packages

### General Commands  
- `gearbox list` - Show available tools with descriptions
- `gearbox help` - Show detailed help and usage information

### Configuration Management
- `gearbox config show` - Display current configuration settings
- `gearbox config set DEFAULT_BUILD_TYPE maximum` - Set configuration values
- `gearbox config wizard` - Interactive configuration setup
- `gearbox config reset` - Reset to default configuration
- `gearbox config help` - Show configuration help

### Health Checks & Diagnostics
- `gearbox doctor` - Run comprehensive health checks
- `gearbox doctor nerd-fonts` - Advanced font diagnostics (cache, terminal, VS Code, starship)
- `gearbox status nerd-fonts` - Detailed font status with individual fonts and disk usage
- `gearbox doctor system` - Check system requirements only
- `gearbox doctor tools` - Check installed tools status
- `gearbox doctor env` - Check environment variables
- `gearbox doctor help` - Show diagnostic help

### Advanced Options
- `gearbox install --skip-common-deps` - Skip common dependency installation
- `gearbox install --run-tests` - Run test suites for tools that support it
- `gearbox install --no-shell` - Skip shell integration setup (fzf)

### Building
- `make build` - Build all components (CLI and tools)
- `make cli` - Build just the Go CLI
- `make tools` - Build orchestrator, script-generator, config-manager
- `make deps` - Install Go dependencies
- `make dev-setup` - Setup development environment
- `make dev` - Quick development setup and test
- `make clean` - Clean build artifacts
- `make install` - Install system-wide (requires sudo)
- `make info` - Show build information (version, build time, Go version)

### Testing
- `make test` - Run all tests (Go and shell)
- `./tests/test-runner.sh` - Run basic validation tests for installation scripts
- `./tests/framework/test-framework.sh` - Comprehensive testing framework for shell functions and integrations
- Individual tests: `test_unit_common.sh`, `test_integration_tools.sh`, `test_template_validation.sh`

### Individual Tool Installation
Use the main CLI for all installations (recommended):
- `gearbox install fd` - Install fd with standard build
- `gearbox install ripgrep --maximum` - Install ripgrep with optimized build
- `gearbox install fzf --no-shell` - Install fzf without shell integration
- `gearbox install nerd-fonts --minimal` - Install essential fonts (3 fonts)
- `gearbox install nerd-fonts` - Install standard font collection (8 fonts)  
- `gearbox install nerd-fonts --maximum` - Install complete font collection (15+ fonts)

Direct scripts are available for advanced use cases but discouraged for normal usage.

### Advanced Nerd Fonts Features

The nerd-fonts implementation provides sophisticated font management with professional-grade UX features:

#### Installation Modes & Build Types
```bash
# CLI-based installations (recommended)
gearbox install nerd-fonts                    # Standard collection (8 fonts, ~80MB)
gearbox install nerd-fonts --minimal          # Essential fonts (3 fonts, ~30MB) 
gearbox install nerd-fonts --maximum          # Complete collection (15+ fonts, ~200MB)

# Advanced font selection via CLI
gearbox install nerd-fonts --fonts="FiraCode"               # Install specific font
gearbox install nerd-fonts --fonts="FiraCode,JetBrainsMono" # Install multiple fonts
gearbox install nerd-fonts --interactive                    # Interactive selection with previews
gearbox install nerd-fonts --preview --fonts="FiraCode"     # Preview before installation

# Combined options
gearbox install nerd-fonts --fonts="FiraCode" --configure-apps  # Install + auto-configure apps
gearbox install nerd-fonts --interactive --configure-apps       # Interactive + configuration
gearbox install nerd-fonts --dry-run --fonts="Hack"             # Preview specific font
```

#### Smart Application Configuration
```bash
# CLI with automatic application configuration
gearbox install nerd-fonts --configure-apps                     # Install + configure apps
gearbox install nerd-fonts --fonts="FiraCode" --configure-apps  # Specific font + configure
gearbox status nerd-fonts                                       # Check installation status and health

# Configuration examples (automatically applied with --configure-apps):
# VS Code: "editor.fontFamily": "FiraCode Nerd Font", "editor.fontLigatures": true
# Kitty: font_family JetBrains Mono Nerd Font  
# Alacritty: family: JetBrains Mono Nerd Font
```

#### Font Collections by Build Type

**Minimal Collection (3 fonts, ~30MB):**
- FiraCode - Programming ligatures + icons
- JetBrains Mono - Clean, readable monospace  
- Hack - Terminal-optimized

**Standard Collection (8 fonts, ~80MB):**
- All minimal fonts plus:
- Source Code Pro - Adobe's programming font
- Inconsolata - Classic monospace
- Cascadia Code - Microsoft's font with ligatures
- Ubuntu Mono - Ubuntu's official monospace
- DejaVu Sans Mono - Popular open source font

**Maximum Collection (15+ fonts, ~200MB):**
- All standard fonts plus:
- Victor Mono - Cursive italic programming font
- Menlo - Apple's monospace font
- Anonymous Pro - Fixed-width font for coders
- Space Mono - Google's monospace font  
- IBM Plex Mono - IBM's corporate font
- Roboto Mono - Google's robot-themed font
- Terminus - Bitmap font optimized for coding

#### Interactive Font Preview System
```bash
# Rich previews with actual character samples
scripts/install-nerd-fonts.sh --preview --fonts="FiraCode,JetBrainsMono"

# Interactive selection with live previews
scripts/install-nerd-fonts.sh --interactive
# Navigation: ↑/↓ arrows, SPACE to select, 'p' to preview, ENTER to confirm
```

**Preview includes:**
- **Programming Ligatures**: `=> != >= <= && || -> <-`
- **File Icons**: `  󰈙  󰗀  󰅴  󰘳  󰊕`
- **Git Symbols**: `  󰊢  󰊰  󰜘  󰊦`  
- **System Icons**: `  󰍹  󰻀  󰍛  󰈎  󰏗`
- **Real Code Samples**: JavaScript/TypeScript examples with ligatures
- **Font Configuration Names**: Exact family names for terminal/editor config

#### Advanced Health Checks & Diagnostics
```bash
# Comprehensive nerd-fonts health check
gearbox doctor nerd-fonts                     # CLI integration
./bin/orchestrator doctor nerd-fonts          # Direct orchestrator access

# Status with detailed font information
gearbox status nerd-fonts                     # Shows installed fonts, disk usage
./bin/orchestrator status nerd-fonts          # Detailed status display
```

**Health Check Coverage:**
- **Font Installation**: Verify specific fonts are properly installed
- **Font Cache System**: Validate fontconfig cache integrity  
- **Terminal Support**: Check Unicode/UTF-8 compatibility
- **VS Code Integration**: Detect and validate editor configuration
- **Terminal Configuration**: Check terminal-specific font settings
- **Starship Integration**: Assess cross-tool compatibility and suggestions
- **System Requirements**: Verify fontconfig, fc-cache, and dependencies

#### Cross-Tool Dependency Intelligence
```bash
# Smart suggestions during installation
gearbox install nerd-fonts --dry-run
# → Suggests: "⭐ Consider adding 'starship' - A customizable prompt that works great with Nerd Fonts"

gearbox install starship --dry-run  
# → Suggests: "🎨 Consider adding 'nerd-fonts' - Starship displays icons and symbols much better with Nerd Fonts"

# Bundle suggestions for related tools
gearbox install fzf --dry-run
# → Suggests terminal enhancement bundle (bat, eza) for complete experience
```

**Intelligent Relationships:**
- **Starship ↔ Nerd Fonts**: Bi-directional suggestions for optimal prompt experience
- **Terminal Enhancement Bundle**: fzf, bat, eza recommendations  
- **Development Tools Bundle**: ripgrep, fd, sd, tokei groupings
- **Git Workflow**: delta + lazygit pairings

#### Font Management & Maintenance
```bash
# Installation with testing and verification
scripts/install-nerd-fonts.sh --run-tests     # Verify fonts after installation
scripts/install-nerd-fonts.sh --force         # Force reinstall existing fonts

# Build configuration and caching  
scripts/install-nerd-fonts.sh --skip-deps     # Skip dependency installation
scripts/install-nerd-fonts.sh --config-only  # Prepare config without installing

# Different installation phases
scripts/install-nerd-fonts.sh --build-only    # Download and prepare, don't install
scripts/install-nerd-fonts.sh --install       # Complete installation (default)
```

#### Professional Configuration Examples

**VS Code settings.json:**
```json
{
  "editor.fontFamily": "FiraCode Nerd Font",
  "editor.fontLigatures": true,
  "terminal.integrated.fontFamily": "JetBrains Mono Nerd Font"
}
```

**Terminal Configurations:**
```bash
# Kitty (auto-configured)
echo 'font_family JetBrains Mono Nerd Font' >> ~/.config/kitty/kitty.conf

# Alacritty (auto-configured) 
# ~/.config/alacritty/alacritty.yml
font:
  normal:
    family: JetBrains Mono Nerd Font
```

**Starship Integration:**
```toml
# ~/.config/starship.toml - enhanced with Nerd Font symbols
[character]
success_symbol = "[](bold green)"
error_symbol = "[](bold red)"

[git_branch] 
symbol = " "

[directory]
substitutions = { "~" = "󰜴" }
```

## Architecture

### Directory Structure
- `scripts/` - All shell code organized in logical structure:
  * `lib/` - Modular shared libraries with lazy loading system
  * `installation/` - All installation scripts categorized by functionality
- `scripts/lib/` - Modular shared library system:
  * `common.sh` - Modular entry point with lazy loading
  * `core/` - Essential modules (logging, validation, security, utilities)
  * `build/` - Build system modules (dependencies, execution, cache, cleanup)
  * `system/` - System integration modules (installation, backup, environment)
  * `config.sh` - Configuration management system (~/.gearboxrc)
  * `doctor.sh` - Health check and diagnostic system
- `scripts/installation/` - Installation scripts organized by category:
  * `common/` - install-all-tools.sh, install-common-deps.sh
  * `categories/core/` - fd, ripgrep, fzf, jq, zoxide
  * `categories/development/` - gh, lazygit, delta, difftastic, etc.
  * `categories/system/` - bottom, procs, bandwhich, dust, fclones
  * `categories/text/` - bat, sd, xsv, tealdeer, eza, choose
  * `categories/media/` - ffmpeg, imagemagick, 7zip
  * `categories/ui/` - nerd-fonts, starship, yazi
- `cmd/gearbox/` - Go CLI source code (Cobra framework):
  * `main.go` - CLI entry point with version info and global flags
  * `commands/` - Command implementations (install, list, config, doctor, status, generate)
  * `internal/` - Internal packages (config, tools, ui)
  * `pkg/` - Shared packages for CLI functionality
- `tools/` - Go tools source code:
  * `orchestrator/` - Advanced installation orchestrator
  * `script-generator/` - Template-based script generator  
  * `config-manager/` - Configuration management tool
- `templates/` - Script generation templates:
  * `base.sh.tmpl` - Base template for all installation scripts
  * Language-specific templates: `rust.sh.tmpl`, `go.sh.tmpl`, `c.sh.tmpl`, `python.sh.tmpl`
- `config/` - Configuration files:
  * `tools.json` - Tool definitions with metadata, build types, and dependencies
- `bin/` - Compiled Go binaries (orchestrator, script-generator, config-manager)
- `gearbox` - Main CLI binary (Go) with commands: install, list, config, doctor, help, status, generate
- `docs/` - Documentation files
- `examples/` - Example usage scripts
- `tests/` - Basic validation tests

### Build System Architecture

**Modular Library System (`scripts/lib/`)**:
- `scripts/lib/common.sh` - Modular entry point with lazy loading:
  * Load essential core modules automatically (logging, validation, security, utilities)
  * Lazy loading functions for optional modules (build, system)
  * Module loading validation and duplicate prevention
  * Comprehensive error handling and module status tracking
- `scripts/lib/core/` - Essential core modules:
  * `logging.sh` - Unified logging (log, error, warning, success, debug, progress, spinners)
  * `validation.sh` - Input validation (tool names, file paths, URLs, versions, sanitization)
  * `security.sh` - Security functions (root prevention, safe execution, injection protection)
  * `utilities.sh` - Core utilities (optimal jobs, file operations, human readable sizes)
- `scripts/lib/build/` - Build system modules:
  * `dependencies.sh` - Dependency management and installation
  * `execution.sh` - Safe command execution with caching
  * `cache.sh` - Build cache system for performance optimization
  * `cleanup.sh` - Build artifact cleanup and management
- `scripts/lib/system/` - System integration modules:
  * `installation.sh` - Installation patterns and verification
  * `backup.sh` - File backup and rollback functionality
  * `environment.sh` - Environment setup and validation
- `scripts/lib/config.sh` - Configuration management system:
  * User preferences in ~/.gearboxrc (10 configurable settings)
  * Default build types, parallel job limits, caching options
  * Interactive configuration wizard and CLI management
- `scripts/lib/doctor.sh` - Health check and diagnostic system:
  * Comprehensive system validation (OS, memory, disk, internet)
  * Installed tool verification and coverage analysis
  * Environment and permission checks

**Main Installation Script (`scripts/install-all-tools.sh`)**:
- Orchestrates tool installation in optimal dependency order for all 31 tools
- Supports three build types: minimal, standard, maximum (configurable via ~/.gearboxrc)
- Handles common dependency installation via `install-common-deps.sh`
- Installation order optimized for shared toolchains: Go tools → Rust tools → C/C++ tools
- Progress indicators for multi-tool installations
- Includes confirmation prompt when installing all tools (30-60 minute process)
- Configuration-aware defaults (respects user preferences)

**Individual Tool Scripts (`scripts/installation/categories/*/install-*.sh`)**:
- Each tool has a dedicated installation script following consistent patterns
- All scripts use modular `scripts/lib/common.sh` system (eliminated code duplication)
- Organized by category: core, development, system, text, media, ui
- Common command-line interface with build type flags (-m, -r, -o, etc.)
- Support for --skip-deps, --run-tests, --force flags
- Build from source with proper dependency validation
- Integrated build cache system for faster reinstallations
- Safe command execution (no eval usage, array-based commands)
- Root prevention checks and comprehensive security validation

### Available Tools (31 total)

**Core Development Tools:**
- **fd** - Fast file finder (Rust) - build flags: -m minimal, -r release
- **ripgrep** - Fast text search (Rust) - build flags: --no-pcre2 minimal, -r release  
- **fzf** - Fuzzy finder (Go) - build flags: -s standard, -p profiling
- **jq** - JSON processor (C) - build flags: -m minimal, -s standard, -o optimized

**Navigation & File Management:**
- **zoxide** - Smart cd command (Rust)
- **yazi** - Terminal file manager (Rust) 
- **fclones** - Duplicate file finder (Rust)
- **bat** - Enhanced cat with syntax highlighting (Rust)
- **eza** - Modern ls replacement (Rust)
- **dust** - Better disk usage analyzer (Rust)

**Development Tools:**
- **serena** - Coding agent toolkit (Python)
- **uv** - Python package manager (Rust)
- **ruff** - Python linter & formatter (Rust)
- **starship** - Customizable shell prompt (Rust)
- **nerd-fonts** - Patched fonts with icons and glyphs for developers (C) - advanced UX features
- **delta** - Syntax-highlighting pager (Rust)
- **lazygit** - Terminal UI for Git (Go)
- **gh** - GitHub CLI (Go)
- **difftastic** - Structural diff tool (Rust)

**System Monitoring:**
- **bottom** - Cross-platform system monitor (Rust)
- **procs** - Modern ps replacement (Rust)
- **bandwhich** - Network bandwidth monitor (Rust)

**Text Processing:**
- **sd** - Find & replace CLI (Rust)
- **xsv** - CSV data toolkit (Rust)
- **choose** - Cut/awk alternative (Rust)
- **tealdeer** - Fast tldr client (Rust)

**Analysis Tools:**
- **tokei** - Code statistics tool (Rust)
- **hyperfine** - Command-line benchmarking (Rust)

**Media Processing:**
- **ffmpeg** - Video/audio processing (C/C++) - build flags: -m minimal, -g standard, -x maximum
- **imagemagick** - Image manipulation (C/C++)
- **7zip** - Compression tool (C/C++) - build flags: -b basic, -o optimized, -a all-features

### Dependency Management
- `install-common-deps.sh` installs shared dependencies (Rust 1.88.0+, Go 1.23.4+, build tools)
- Individual scripts can skip dependency installation with `--skip-deps`
- Source repositories cloned to `~/tools/build/` to keep scripts directory clean
- Binaries installed to `/usr/local/bin/` with proper PATH management

### Build Configuration
The system supports three build types across all tools:
- **minimal**: Fast builds with basic features
- **standard**: Balanced builds with reasonable features (default, configurable)
- **maximum**: Full-featured builds with all optimizations

Default build type can be configured via:
- `gearbox config set DEFAULT_BUILD_TYPE maximum`
- Interactive wizard: `gearbox config wizard`
- Configuration file: `~/.gearboxrc`

### Performance Optimizations
- **Build Cache System**: Automated caching of compiled binaries by tool and build type
  * Cache stored in `~/tools/cache/` with automatic cleanup (configurable retention)
  * Significant speed improvement for reinstallations and testing
  * Supports multiple build types per tool
- **Parallel Builds**: Optimal CPU core usage with memory-aware job limits
  * Auto-detection of system resources (CPU cores, available memory)
  * Configurable via `MAX_PARALLEL_JOBS` setting
  * Safety limits prevent system overload
- **Progress Indicators**: Real-time progress tracking for multi-tool installations
  * Step-by-step progress reporting (e.g., "Installing 3/30 tools...")
  * Time estimates and status updates during long builds

### Security Enhancements
- **Command Injection Prevention**: All dangerous eval usage eliminated
  * Migrated to safe array-based command execution
  * Input validation and sanitization
- **Root Prevention**: Scripts refuse to run as root for security
- **Safe Configuration**: Input validation for all user configuration values

### User Experience Improvements
- **Configuration Management**: Comprehensive user preference system
  * 10 configurable settings (build types, caching, dependencies, etc.)
  * Interactive configuration wizard
  * Command-line configuration management
- **Health Check System**: Comprehensive diagnostic capabilities
  * System requirements validation (OS, memory, disk, internet)
  * Tool installation verification with coverage analysis
  * Environment and permission checks
  * Actionable recommendations for issues

### Testing Strategy
- **Comprehensive Test Suite**: Multi-layered testing system with 50+ function coverage
  * `test_core_functions.sh` - Quick validation of essential functions across all modules
  * `test_unit_comprehensive.sh` - Detailed unit tests for all shell functions
  * `test_workflow_integration.sh` - Multi-tool workflow and integration testing
  * `test_performance_benchmarks.sh` - Performance analysis and optimization identification
  * `test_error_handling.sh` - Security and resilience testing with 25+ scenarios
- **Security Testing**: Command injection, path traversal, privilege escalation prevention
- **Performance Benchmarking**: Function timing, memory usage, parallel execution analysis
- **Shell Testing Framework**: Test framework system (`tests/framework/test-framework.sh`)
  * Unit tests for shared library functions
  * Integration tests for tool installations (`test_integration_tools.sh`)
  * Template validation tests (`test_template_validation.sh`)
  * Test result tracking with pass/fail/skip counts and timing
- **Basic Validation**: `test-runner.sh` validates script executability and configuration loading
- **Go Testing**: Standard Go test framework for CLI and tools (`go test ./...`)
- **Post-Install Validation**: Individual tools support `--run-tests` flag for post-build validation  
- **Runtime Verification**: All installed tools checked via command line accessibility
- **Health Check System**: Ongoing validation via `gearbox doctor` with comprehensive diagnostics

## Key Implementation Patterns

### Script Structure
All installation scripts follow consistent patterns:
- Source `scripts/lib/common.sh` for modular shared functions and utilities
- Automatic loading of core modules (logging, validation, security, utilities)
- Lazy loading of optional modules (build, system) when needed
- Standardized command-line argument parsing with build type flags
- Root user prevention checks and comprehensive security validation
- Build cache integration for performance optimization
- Safe command execution (no eval, array-based commands, injection protection)
- Parallel build support with optimal CPU utilization
- Progress indicators and real-time status reporting
- Error handling with cleanup traps and rollback functionality
- Install dependencies (unless `--skip-deps` specified)
- Clone source to build directories outside script directory
- Configure build based on build type (minimal/standard/maximum, configurable)
- Build and install to `/usr/local/bin/` or `~/.cargo/bin/`
- Cache successful builds for future use with integrity verification
- Optional post-install testing and shell integration
- Comprehensive input validation and sanitization

### CLI Interface Design
- **Safety first**: No command auto-executes without user awareness
- **Confirmation prompts**: Installing all tools shows impact and requires confirmation  
- **Specific over general**: `gearbox install fd ripgrep` recommended over `gearbox install`
- **Clear help**: `gearbox list` shows all tools, `gearbox help` shows usage
- **Configuration-aware**: Respects user preferences from `~/.gearboxrc`
- **Comprehensive diagnostics**: `gearbox doctor` provides health checks and recommendations
- **User-friendly configuration**: Interactive wizard and CLI management via `gearbox config`

### Build System Integration
- **Shared toolchains**: Common dependencies installed once (Rust 1.88.0+, Go 1.23.4+)
- **Optimized order**: Go tools first, then Rust tools, then C/C++ tools
- **Multiple build types**: Tools support different optimization levels via standardized flags
- **Clean separation**: Source builds in `~/tools/build/`, final binaries in `/usr/local/bin/`

### Go CLI Architecture
The main CLI (`gearbox`) is built using Cobra framework:
- **Type-safe commands**: Each command in `cmd/gearbox/commands/` with proper validation
- **Global flags**: `--verbose`/`-v` and `--quiet`/`-q` for output control
- **Version embedding**: Build-time version info via ldflags
- **Orchestrator integration**: CLI delegates complex operations to Go tools
- **Shell script fallback**: Commands can fall back to existing shell scripts when needed
- **Error handling**: Structured error reporting with proper exit codes

### Template System
Script generation via Go templates (`templates/`):
- **Base template** (`base.sh.tmpl`): Common structure for all installation scripts
- **Language-specific templates**: Specialized patterns for Rust, Go, C/C++, Python tools
- **Metadata-driven**: Tool definitions in `config/tools.json` drive template rendering
- **Template variables**: Tool name, repository, build types, dependencies, shell integration
- **Generated scripts**: Follow same patterns as hand-written scripts, use modular `scripts/lib/common.sh`
- **Validation**: Template output validated by test framework

### Advanced Tool Patterns

The nerd-fonts implementation showcases advanced architectural patterns for complex tools:

#### Sophisticated UX Architecture
- **Multi-Modal Interface**: Supports CLI args, interactive selection, and preview modes
- **Rich Preview System**: Real character samples with categorized symbol display
- **Cross-Tool Intelligence**: Bi-directional relationship awareness (starship ↔ nerd-fonts)
- **Contextual Suggestions**: Bundle recommendations based on installation patterns

#### Font Management Architecture
```bash
# Font collection management via associative arrays
declare -A MINIMAL_FONTS=( [FiraCode]="..." [JetBrainsMono]="..." )
declare -A STANDARD_FONTS=( ... )  # Inherits + extends minimal
declare -A MAXIMUM_FONTS=( ... )   # Inherits + extends standard

# Dynamic font collection resolution
get_font_collection "$BUILD_TYPE" selected_fonts
```

#### Interactive System Architecture
```bash
# Navigation state management
current_index=0
font_selected["FontName"]=true/false

# Live preview integration
case "$key" in
    'p'|'P') show_font_preview "${fonts_array[$current_index]}" ;;
    $'\x1b') handle_arrow_keys ;;  # ESC sequences for ↑/↓
    ' ')     toggle_font_selection ;;
    '')      confirm_selection ;;
esac
```

#### Health Check Integration
```go
// Orchestrator integration for advanced diagnostics
func (o *Orchestrator) runNerdFontsDoctor() error {
    // Multi-faceted health assessment
    fontStatus := isNerdFontsInstalled()
    cacheHealth, issues := checkFontcacheHealth()
    terminalSupport := checkTerminalSupport()
    vscodeConfigured, fontFamily := checkVSCodeFontConfig()
    
    // Cross-tool relationship analysis
    if isStarshipInstalled() && !fontStatus {
        recommendations = append(recommendations, starshipNeedsFontsMessage)
    }
    
    return formatHealthReport(status, issues, recommendations)
}
```

#### Dependency Intelligence System
```go
// Cross-tool relationship mapping
func (o *Orchestrator) suggestRelatedTools(tools []ToolConfig) {
    relationships := map[string][]string{
        "starship":    {"nerd-fonts"},     // Prompt needs icons
        "nerd-fonts":  {"starship"},       // Fonts enhance prompts
        "delta":       {"lazygit"},        // Git workflow bundle
        "fzf":         {"bat", "eza"},     // Terminal enhancement bundle
    }
    
    // Dynamic suggestion generation based on installation context
    generateIntelligentSuggestions(tools, relationships)
}
```

#### Application Configuration Architecture
```bash
# Multi-application auto-configuration
configure_applications() {
    # VS Code JSON manipulation with jq
    jq '. + {"editor.fontFamily": "FiraCode Nerd Font"}' settings.json
    
    # Terminal-specific configuration detection and setup
    configure_terminal_fonts()  # Handles Kitty, Alacritty, GNOME Terminal
}

# Dynamic terminal detection and configuration
configure_terminal_fonts() {
    case "$(detect_terminal)" in
        kitty)     configure_kitty_font ;;
        alacritty) configure_alacritty_font ;;
        gnome)     suggest_gnome_configuration ;;
    esac
}
```

#### Status Display Architecture
```go
// Enhanced status reporting for complex tools
func (o *Orchestrator) showNerdFontsDetailedStatus() {
    status := getNerdFontsDetailedStatus()
    
    fmt.Printf("📋 Nerd Fonts Status\n")
    fmt.Printf("Installed: %d fonts\n", status["installed_count"])
    fmt.Printf("Disk Usage: %s\n", status["disk_usage"])
    
    // Individual font listing with metadata
    for font, info := range status["fonts"].(map[string]interface{}) {
        fmt.Printf("  ✅ %s (%s)\n", font, info["description"])
    }
}
```

This architecture enables:
- **Professional UX**: Rich previews, intelligent suggestions, comprehensive diagnostics
- **Cross-Tool Intelligence**: Relationship awareness between tools for optimal workflow
- **Scalable Patterns**: Template for implementing advanced features in other tools
- **Maintainable Code**: Clean separation between UI, business logic, and system integration

## Development Workflow

### Adding New Tools
1. **Define tool metadata** in `config/tools.json`:
   * Basic info: name, description, category, repository
   * Build configuration: language, build types, dependencies
   * Integration: binary name, test command, shell integration
2. **Generate installation script**: Use `gearbox generate <tool-name>` 
3. **Test script**: Run generated script with `--dry-run` and validation tests
4. **Update documentation**: Tool automatically appears in `gearbox list`

### Modifying CLI Commands
1. **Edit command files** in `cmd/gearbox/commands/`
2. **Build and test**: `make cli && ./gearbox --help`
3. **Run tests**: `make test` for Go tests, `./tests/test-runner.sh` for shell tests
4. **Update help text**: Ensure examples and descriptions are current

### Tool Metadata Schema (`config/tools.json`)
Each tool entry contains:
- **Core fields**: `name`, `description`, `category`, `repository`, `binary_name`, `language`
- **Build types**: `minimal`, `standard`, `maximum` with corresponding flags
- **Dependencies**: Array of system packages and language toolchains
- **Integration**: `shell_integration` boolean, `test_command` for validation
- **Versioning**: `min_version` for compatibility requirements