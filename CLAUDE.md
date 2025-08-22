# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Essential Tools Installer - a collection of automated installation scripts for essential command-line tools on Debian Linux. The project focuses on building tools from source with various optimization levels and proper dependency management.

## Getting Started

### Quick Start
After building the project, use the CLI from the build directory:
```bash
# Build the project
make build

# Use CLI from project root (recommended)
./build/gearbox --help
./build/gearbox list bundles
./build/./build/gearbox install fd ripgrep fzf

# Or install system-wide for global access
make install
gearbox --help
```

## Common Commands

### Installation Commands
- `./build/./build/gearbox install` - Show confirmation prompt and install all tools
- `./build/./build/gearbox install fd ripgrep fzf` - Install only specified tools (recommended approach)
- `./build/./build/gearbox install --minimal fd ripgrep` - Install with minimal/fast builds  
- `./build/./build/gearbox install --maximum ffmpeg` - Install with full-featured builds
- `./build/./build/gearbox install nerd-fonts` - Install standard font collection (8 fonts) with cross-tool suggestions
- `./build/./build/gearbox install nerd-fonts --fonts="FiraCode"` - Install specific font via CLI
- `./build/./build/gearbox install nerd-fonts --interactive` - Interactive font selection with previews

### Bundle Installation - User Journey Architecture

**🎯 Foundation Tier (Start Your Journey):**
- `./build/./build/gearbox install --bundle beginner` - Perfect starting point for new developers (essential tools + beautiful terminal)
- `./build/./build/gearbox install --bundle intermediate` - Productive developer environment with git workflow and development tools
- `./build/./build/gearbox install --bundle advanced` - High-performance development environment with debugging and performance tools

**🏗️ Domain Tier (Choose Your Role):**
- `./build/./build/gearbox install --bundle polyglot-dev` - Multi-language development environment (Python + Node.js + Docker + Cloud + Editors)
- `./build/./build/gearbox install --bundle fullstack-dev` - Complete web development (frontend + backend + databases)
- `./build/./build/gearbox install --bundle mobile-dev` - Cross-platform mobile development environment
- `./build/./build/gearbox install --bundle data-dev` - Data science and machine learning development environment
- `./build/./build/gearbox install --bundle devops-dev` - Infrastructure, monitoring, deployment + modern container tools
- `./build/./build/gearbox install --bundle security-dev` - Security analysis, penetration testing + container security
- `./build/./build/gearbox install --bundle game-dev` - Game development environment with graphics and engine tools

**🚀 Language Ecosystem Tier:**
- `./build/./build/gearbox install --bundle python-dev` - Python runtime + uv, ruff, black, mypy, poetry, pytest, ipython + essential tools
- `./build/./build/gearbox install --bundle nodejs-dev` - Node.js runtime + TypeScript, ESLint, yarn, pnpm, jest + essential tools
- `./build/./build/gearbox install --bundle rust-dev` - Rust compiler + rustfmt, clippy, rust-analyzer, cargo tools + essential tools
- `./build/./build/gearbox install --bundle go-dev` - Go compiler + gopls, golangci-lint, air, staticcheck, delve + essential tools
- `./build/./build/gearbox install --bundle java-dev` - Java 17 + Maven, Gradle + essential tools
- `./build/./build/gearbox install --bundle ruby-dev` - Ruby runtime + Rails, RSpec, RuboCop, Solargraph + essential tools
- `./build/./build/gearbox install --bundle cpp-dev` - GCC/Clang + CMake, Ninja, GDB, Valgrind, Conan, vcpkg + essential tools

**⚙️ Workflow Tier:**
- `./build/./build/gearbox install --bundle debugging-tools` - Profilers, memory analyzers, and network debugging tools
- `./build/./build/gearbox install --bundle deployment-tools` - CI/CD, containers, and cloud deployment tools
- `./build/./build/gearbox install --bundle code-review-tools` - Code linting, formatting, and analysis tools (cross-language)

**🐳 Container Development (2024 Best Practice):**
- `./build/./build/gearbox install --bundle docker` - Complete Docker development environment with security and analysis tools
- `./build/./build/gearbox install --bundle docker-rootless` - Docker CE with rootless mode (maximum security)

**🤖 AI & Infrastructure:**
- `./build/./build/gearbox install --bundle ai-tools` - AI-powered coding assistance (serena + aider + mise + just)
- `./build/./build/gearbox install --bundle cloud-tools` - AWS CLI v2 and cloud platform tools
- `./build/./build/gearbox install --bundle editors` - Neovim and modern text editors
- `./build/./build/gearbox install --bundle media-tools` - Media processing tools (ffmpeg, imagemagick, 7zip)

**Bundle Management:**
- `./build/gearbox list bundles` - Show available bundles with descriptions
- `./build/gearbox show bundle fullstack-dev` - Show bundle contents including system packages

### General Commands  
- `./build/gearbox list` - Show available tools with descriptions
- `./build/gearbox help` - Show detailed help and usage information

### Configuration Management
- `./build/gearbox config show` - Display current configuration settings
- `./build/gearbox config set DEFAULT_BUILD_TYPE maximum` - Set configuration values
- `./build/gearbox config wizard` - Interactive configuration setup
- `./build/gearbox config reset` - Reset to default configuration
- `./build/gearbox config help` - Show configuration help

### Health Checks & Diagnostics
- `./build/gearbox doctor` - Run comprehensive health checks
- `./build/gearbox doctor nerd-fonts` - Advanced font diagnostics (cache, terminal, VS Code, starship)
- `./build/gearbox status nerd-fonts` - Detailed font status with individual fonts and disk usage
- `./build/gearbox doctor system` - Check system requirements only
- `./build/gearbox doctor tools` - Check installed tools status
- `./build/gearbox doctor env` - Check environment variables
- `./build/gearbox doctor help` - Show diagnostic help

### Advanced Options
- `./build/gearbox install --skip-common-deps` - Skip common dependency installation
- `./build/gearbox install --run-tests` - Run test suites for tools that support it
- `./build/gearbox install --no-shell` - Skip shell integration setup (fzf)

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
- `./build/gearbox install fd` - Install fd with standard build
- `./build/gearbox install ripgrep --maximum` - Install ripgrep with optimized build
- `./build/gearbox install fzf --no-shell` - Install fzf without shell integration
- `./build/gearbox install nerd-fonts --minimal` - Install essential fonts (3 fonts)
- `./build/gearbox install nerd-fonts` - Install standard font collection (8 fonts)  
- `./build/gearbox install nerd-fonts --maximum` - Install complete font collection (15+ fonts)

Direct scripts are available for advanced use cases but discouraged for normal usage.

### Advanced Nerd Fonts Features

The nerd-fonts implementation provides sophisticated font management with professional-grade UX features:

#### Installation Modes & Build Types
```bash
# CLI-based installations (recommended)
./build/gearbox install nerd-fonts                    # Standard collection (8 fonts, ~80MB)
./build/gearbox install nerd-fonts --minimal          # Essential fonts (3 fonts, ~30MB) 
./build/gearbox install nerd-fonts --maximum          # Complete collection (15+ fonts, ~200MB)

# Advanced font selection via CLI
./build/gearbox install nerd-fonts --fonts="FiraCode"               # Install specific font
./build/gearbox install nerd-fonts --fonts="FiraCode,JetBrainsMono" # Install multiple fonts
./build/gearbox install nerd-fonts --interactive                    # Interactive selection with previews
./build/gearbox install nerd-fonts --preview --fonts="FiraCode"     # Preview before installation

# Combined options
./build/gearbox install nerd-fonts --fonts="FiraCode" --configure-apps  # Install + auto-configure apps
./build/gearbox install nerd-fonts --interactive --configure-apps       # Interactive + configuration
./build/gearbox install nerd-fonts --dry-run --fonts="Hack"             # Preview specific font
```

#### Smart Application Configuration
```bash
# CLI with automatic application configuration
./build/gearbox install nerd-fonts --configure-apps                     # Install + configure apps
./build/gearbox install nerd-fonts --fonts="FiraCode" --configure-apps  # Specific font + configure
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
./build/gearbox doctor nerd-fonts             # CLI integration
./build/orchestrator doctor nerd-fonts        # Direct orchestrator access

# Status with detailed font information
./build/gearbox status nerd-fonts             # Shows installed fonts, disk usage
./build/orchestrator status nerd-fonts        # Detailed status display
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
./build/gearbox install nerd-fonts --dry-run
# → Suggests: "⭐ Consider adding 'starship' - A customizable prompt that works great with Nerd Fonts"

./build/gearbox install starship --dry-run  
# → Suggests: "🎨 Consider adding 'nerd-fonts' - Starship displays icons and symbols much better with Nerd Fonts"

# Bundle suggestions for related tools
./build/gearbox install fzf --dry-run
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
- `cmd/` - Go command entry points:
  * `gearbox/` - Main CLI source code (Cobra framework)
  * `orchestrator/` - Advanced installation orchestrator entry point
  * `script-generator/` - Template-based script generator entry point
  * `config-manager/` - Configuration management tool entry point
- `pkg/` - Go packages (shared, reusable code):
  * `orchestrator/` - Modular installation orchestration logic:
    - `main.go` - Entry point and CLI setup (56 lines)
    - `types.go` - Type definitions and data structures
    - `config.go` - Configuration loading and validation
    - `orchestrator.go` - Core orchestration logic
    - `commands.go` - CLI command definitions
    - `installation.go` - Installation and dependency management
    - `verification.go` - Tool verification and status reporting
    - `nerdfonts.go` - Specialized nerd-fonts functionality
    - `uninstall_commands.go` - Uninstall operations
    - `utils.go` - Common utilities and helpers
  * `generator/` - Script generation functionality
  * `config/` - Configuration management
  * `manifest/` - Installation tracking and state management with comprehensive test coverage
  * `uninstall/` - Safe removal engine with dry-run support and dependency analysis
  * `errors/` - Structured error handling with context and suggestions
  * `logger/` - Centralized structured logging with level support
  * `validation/` - Input validation utilities
- `build/` - **Compiled binaries directory (Go best practice)**:
  * `gearbox` - Main CLI binary
  * `orchestrator` - Installation orchestrator
  * `script-generator` - Template generator
  * `config-manager` - Configuration tool
- `templates/` - Script generation templates:
  * `base.sh.tmpl` - Base template for all installation scripts
  * Language-specific templates: `rust.sh.tmpl`, `go.sh.tmpl`, `c.sh.tmpl`, `python.sh.tmpl`
- `config/` - Configuration files:
  * `tools.json` - Tool definitions with metadata, build types, and dependencies
  * `bundles.json` - Bundle definitions and relationships
- `docs/` - Comprehensive documentation library:
  * `USER_GUIDE.md` - End-user installation and usage guide
  * `DEVELOPER_GUIDE.md` - Developer setup and contribution guide
  * `TESTING_GUIDE.md` - Testing framework and validation procedures
  * `TROUBLESHOOTING.md` - Common issues and solutions
  * `CONFIGURATION_MIGRATION.md` - Configuration system documentation
  * `GO_MIGRATION_PLAN.md` - Shell-to-Go migration strategy
  * `INSTALLATION_METHODS.md` - Detailed installation method documentation
  * `DISK_SPACE_MANAGEMENT.md` - Storage optimization and cache management
  * Additional technical guides and migration documentation
- `README.md` - Main project overview and quick start guide (project root)
- `examples/` - Example usage scripts:
  * `full-setup.sh` - Complete installation example
  * `minimal-setup.sh` - Minimal installation example  
  * `rust-tools-only.sh` - Rust-specific tools installation
- `tests/` - Comprehensive test suite

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
  * `cache.sh` - Complete build cache system with metadata, validation, and cleanup
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
- Orchestrates tool installation in optimal dependency order for all 42 tools
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

### Available Tools (42 total)

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
- `./build/gearbox config set DEFAULT_BUILD_TYPE maximum`
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
- **Comprehensive Go Test Coverage**: Extensive test coverage across all Go packages
  * `pkg/errors/`: 25+ test functions covering all error types, context methods, suggestion generation
  * `pkg/logger/`: Structured logging tests with JSON validation and level handling  
  * `pkg/manifest/`: Complete CRUD operations, atomic writes, backup/restore, concurrent access
  * `pkg/uninstall/`: Removal execution, dry-run support, dependency analysis, safety checks
  * **450+ total test cases** with benchmarks and edge case coverage
- **Shell Testing Framework**: Multi-layered testing system with 50+ function coverage
  * `test_core_functions.sh` - Quick validation of essential functions across all modules
  * `test_unit_comprehensive.sh` - Detailed unit tests for all shell functions
  * `test_workflow_integration.sh` - Multi-tool workflow and integration testing
  * `test_performance_benchmarks.sh` - Performance analysis and optimization identification
  * `test_error_handling.sh` - Security and resilience testing with 25+ scenarios
- **Security Testing**: Command injection, path traversal, privilege escalation prevention
- **Performance Benchmarking**: Function timing, memory usage, parallel execution analysis
- **Integration Testing**: Test framework system (`tests/framework/test-framework.sh`)
  * Unit tests for shared library functions
  * Integration tests for tool installations (`test_integration_tools.sh`)
  * Template validation tests (`test_template_validation.sh`)
  * Test result tracking with pass/fail/skip counts and timing
- **Basic Validation**: `test-runner.sh` validates script executability and configuration loading
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
- **Specific over general**: `./build/gearbox install fd ripgrep` recommended over `./build/gearbox install`
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

## Documentation Structure

### Quick Navigation
The project documentation is organized across multiple locations for different audiences:

**📋 Project Root**:
- **`README.md`** - Main project overview, quick start, and installation guide
- **`CLAUDE.md`** - This file: comprehensive technical documentation for Claude Code

**📁 Documentation Library (`docs/`)**:
- **`USER_GUIDE.md`** - Complete end-user manual with usage examples
- **`DEVELOPER_GUIDE.md`** - Developer setup, contribution guidelines, and architecture
- **`TESTING_GUIDE.md`** - Testing framework documentation and validation procedures
- **`TROUBLESHOOTING.md`** - Common issues, solutions, and debugging guides

**🔧 Technical Documentation (`docs/`)**:
- **`CONFIGURATION_MIGRATION.md`** - Configuration system and user preferences
- **`GO_MIGRATION_PLAN.md`** - Strategic shell-to-Go migration roadmap
- **`INSTALLATION_METHODS.md`** - Detailed installation method documentation
- **`DISK_SPACE_MANAGEMENT.md`** - Storage optimization and cache management
- **`CLI_MIGRATION_COMPLETED.md`** - Completed CLI migration documentation
- **Phase-specific guides**: `PHASE2_ORCHESTRATION.md`, `PHASE3_SCRIPT_GENERATION.md`

**🎯 Examples & Practical Guides (`examples/`)**:
- **`full-setup.sh`** - Complete installation with all tools
- **`minimal-setup.sh`** - Lightweight installation for basic usage
- **`rust-tools-only.sh`** - Rust ecosystem tools installation

### Documentation Usage Guide
- **New Users**: Start with `README.md` → `docs/USER_GUIDE.md`
- **Developers**: Read `docs/DEVELOPER_GUIDE.md` → `CLAUDE.md` → `docs/TESTING_GUIDE.md`
- **Issues/Problems**: Check `docs/TROUBLESHOOTING.md` first
- **Advanced Configuration**: See `docs/CONFIGURATION_MIGRATION.md`
- **Migration Planning**: Review `docs/GO_MIGRATION_PLAN.md`

## Interactive TUI Implementation

The project now includes a comprehensive Text User Interface (TUI) built with the Bubble Tea framework, providing an intuitive visual interface for all gearbox operations.

### TUI Architecture

**Framework**: Bubble Tea (Elm architecture for Go)
- Model-View-Update pattern for reactive UI
- Tea commands for async operations
- Lipgloss for consistent styling
- Full keyboard navigation support

**Package Structure** (`cmd/gearbox/tui/`):
```
tui/
├── app.go              # Main TUI application and model
├── state.go            # Global application state management
├── taskprovider.go     # Task manager adapter interface
├── styles/
│   └── theme.go        # Consistent styling and theming
├── tasks/
│   └── manager.go      # Background task management
└── views/
    ├── interfaces.go   # Common interfaces (TaskProvider, TaskStatus)
    ├── dashboard.go    # System overview and quick actions
    ├── toolbrowser.go  # Tool search and selection
    ├── bundleexplorer.go # Bundle browsing and installation
    ├── installmanager.go # Installation progress tracking
    ├── config.go       # Configuration management
    └── health.go       # System health monitoring
```

### TUI Features

**1. Dashboard View** (`views/dashboard.go`)
- System statistics (installed/available tools)
- Recent activity tracking
- Smart recommendations based on installed tools
- Quick action buttons for common operations
- System resource overview

**2. Tool Browser** (`views/toolbrowser.go`)
- Real-time search across names, descriptions, and languages
- Category-based filtering (Core, Development, System, etc.)
- Multi-selection support with Space key
- Side-by-side preview pane with detailed information
- Installation status indicators

**3. Bundle Explorer** (`views/bundleexplorer.go`)
- Hierarchical bundle display organized by tiers
- Expandable bundle details showing included tools
- Installation progress tracking per bundle
- Smart category filtering
- One-click bundle installation

**4. Install Manager** (`views/installmanager.go`)
- Real-time installation progress with stage information
- Concurrent task management (configurable parallelism)
- Live output streaming from installation processes
- Progress bars with percentage and time estimates
- Cancel/retry capabilities for failed installations

**5. Configuration View** (`views/config.go`)
- Interactive settings editing with type validation
- Support for string, number, boolean, and choice types
- Live validation with error feedback
- Reset to defaults option
- Prepared for integration with ~/.gearboxrc

**6. Health Monitor** (`views/health.go`)
- Comprehensive system health checks
- Tool installation coverage analysis
- Toolchain verification (Rust, Go, build tools)
- Smart suggestions for resolving issues
- Auto-refresh capability for monitoring

### Technical Implementation Details

**State Management**:
```go
type AppState struct {
    CurrentView    ViewType
    Tools          []orchestrator.ToolConfig
    Bundles        []orchestrator.BundleConfig
    InstalledTools map[string]*manifest.InstallationRecord
    TaskQueue      []string
}
```

**Task System Architecture**:
- Background goroutines for parallel installations
- Channel-based updates via `TaskUpdateMsg`
- TaskProvider interface for view/model decoupling
- Simulated installations with realistic progress stages
- Proper cancellation support with context

**UI Component Patterns**:
```go
// Common view interface pattern
type View interface {
    SetSize(width, height int)
    Update(msg tea.Msg) tea.Cmd
    Render() string
}
```

**Integration Points**:
- Seamless CLI/TUI integration via `gearbox tui` command
- Shared orchestrator backend for consistency
- Manifest tracking across CLI and TUI operations
- Configuration management via shared config system

### Usage Examples

```bash
# Launch TUI
gearbox tui

# Navigation
Tab         - Switch between views
↑/↓ or j/k  - Navigate lists
Enter       - Select/Confirm
Space       - Toggle selection
/           - Search (in Tool Browser)
?           - Help screen
q           - Quit

# View-specific shortcuts
D - Dashboard
T - Tool Browser
B - Bundle Explorer
I - Install Manager
C - Configuration
H - Health Monitor

# Tool Browser operations
Space - Select/deselect tool
i     - Install selected tools
c     - Cycle categories
p     - Toggle preview

# Bundle Explorer operations
Enter - Expand/collapse bundle
i     - Install bundle
c     - Cycle category filter

# Install Manager operations
s     - Start pending installations
c     - Cancel current task
o     - Toggle output display
```

### Development Guidelines for TUI

**Adding New Views**:
1. Create view struct in `views/` implementing common methods
2. Add view type to `ViewType` enum in `state.go`
3. Register view in `app.go` NewModel and update methods
4. Add navigation shortcut and help text

**Styling Consistency**:
- Use `styles.CurrentTheme` for all colors
- Apply consistent spacing and borders
- Use semantic style names (SuccessStyle, ErrorStyle, etc.)

**Task Integration**:
- All long-running operations should use TaskManager
- Provide real-time feedback via progress updates
- Handle cancellation gracefully

## Recent Improvements & Best Practices

### Code Quality Enhancements (2024)

#### Comprehensive Test Coverage
The project now features extensive test coverage across all Go packages:

- **pkg/errors**: 25+ test functions covering all error types, context methods, and suggestion generation
- **pkg/logger**: Structured logging tests with JSON validation and level handling
- **pkg/manifest**: Complete CRUD operations, atomic writes, backup/restore, and concurrent access
- **pkg/uninstall**: Removal execution, dry-run support, dependency analysis, and safety checks

**Test execution**:
```bash
go test ./... -v                    # Run all Go tests
./tests/test-runner.sh             # Shell script validation
./tests/framework/test-framework.sh # Comprehensive integration tests
```

#### Modular Architecture Refactoring  
The monolithic `pkg/orchestrator/main.go` (2250 lines) has been split into focused modules:

- **`types.go`** - All struct and type definitions
- **`config.go`** - Configuration loading and validation
- **`orchestrator.go`** - Core orchestration logic (111 lines)
- **`commands.go`** - CLI command definitions (219 lines)
- **`installation.go`** - Installation and dependency management (451 lines)
- **`verification.go`** - Tool verification and status reporting (91 lines)
- **`nerdfonts.go`** - Specialized nerd-fonts functionality (459 lines)
- **`uninstall_commands.go`** - Uninstall operations (156 lines)
- **`utils.go`** - Common utilities and helpers

**Benefits**:
- **98% reduction** in main file size (2250 → 56 lines)
- **Clear separation** of concerns and responsibilities
- **Easier maintenance** and feature development
- **Better testing** of individual components

#### Modern Go Practices
- **Deprecated function removal**: All `ioutil.*` functions replaced with modern `os.*` equivalents
- **Structured error handling**: Custom error types with context and suggestions
- **Type safety**: Comprehensive input validation and sanitization
- **Memory efficiency**: Optimized data structures and concurrent operations

#### Enhanced Build Cache System
Complete implementation of the build cache system in `scripts/lib/build/cache.sh`:

```bash
# Cache management functions
cache_binary <tool> <build_type> <version> <binary_path>    # Cache compiled binary
is_cached <tool> <build_type> <version>                     # Check cache status
get_cached_binary <tool> <build_type> <version>             # Retrieve from cache
show_cache_stats                                             # Display usage statistics
clean_old_cache [max_age_days]                              # Cleanup old entries
validate_cache                                               # Integrity verification
```

**Performance benefits**:
- **Faster reinstallations**: Skip compilation for cached builds
- **Automatic cleanup**: Configurable retention policies
- **Integrity validation**: Verify cached binaries before use
- **Space monitoring**: Track cache usage and disk space

### Architecture Migration Strategy

#### Shell vs Go Analysis
Comprehensive analysis of 50+ shell scripts (16,464 lines) identified optimal migration opportunities:

**High Priority for Go Migration**:
- **Configuration Management**: Type-safe config with validation
- **Health Check System**: Structured diagnostics and reporting  
- **Bundle Management**: Complex dependency resolution logic

**Keep as Shell** (Domain Expertise):
- **Individual Tool Scripts**: Build tool orchestration expertise
- **System Integration**: File operations and environment setup
- **Build Modules**: Natural fit for command execution

**Hybrid Approach**:
- **Go Orchestration**: Complex logic, parallel coordination, error handling
- **Shell Execution**: Specific tool builds and system interactions
- **Structured Data Flow**: Type-safe communication between components

#### Future Roadmap
1. **Phase 1**: Migrate configuration and health check systems
2. **Phase 2**: Enhanced installation orchestration with Go
3. **Phase 3**: Rich APIs and structured logging integration

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