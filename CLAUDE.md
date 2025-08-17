# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Essential Tools Installer - a collection of automated installation scripts for essential command-line tools on Debian Linux. The project focuses on building tools from source with various optimization levels and proper dependency management.

## Common Commands

### Installation Commands
- `gearbox install` - Show confirmation prompt and install all 30 tools
- `gearbox install fd ripgrep fzf` - Install only specified tools (recommended approach)
- `gearbox install --minimal fd ripgrep` - Install with minimal/fast builds  
- `gearbox install --maximum ffmpeg` - Install with full-featured builds
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
- `gearbox doctor system` - Check system requirements only
- `gearbox doctor tools` - Check installed tools status
- `gearbox doctor env` - Check environment variables
- `gearbox doctor help` - Show diagnostic help

### Advanced Options
- `gearbox install --skip-common-deps` - Skip common dependency installation
- `gearbox install --run-tests` - Run test suites for tools that support it
- `gearbox install --no-shell` - Skip shell integration setup (fzf)

### Testing
- `./tests/test-runner.sh` - Run basic validation tests for installation scripts

### Individual Tool Installation
Each tool has its own script in `scripts/`:
- `scripts/install-fd.sh -r` - Install fd with release build
- `scripts/install-ripgrep.sh --release` - Install ripgrep with optimized build
- `scripts/install-fzf.sh --standard --no-shell` - Install fzf with standard build, no shell integration

## Architecture

### Directory Structure
- `scripts/` - Individual installation scripts for each tool (30+ scripts)
- `lib/` - Shared library modules:
  * `common.sh` - Core shared functions, logging, and utilities
  * `config.sh` - Configuration management system (~/.gearboxrc)
  * `doctor.sh` - Health check and diagnostic system
- `config.sh` - Legacy configuration (migrated to lib/common.sh)
- `gearbox` - Main CLI script with commands: install, list, config, doctor, help
- `docs/` - Documentation files
- `examples/` - Example usage scripts
- `tests/` - Basic validation tests

### Build System Architecture

**Shared Library System (`lib/`)**:
- `lib/common.sh` - Central shared library for all scripts:
  * Unified logging functions (log, error, warning, success)
  * Build utilities (get_optimal_jobs, parallel execution)
  * Progress indicators and status reporting
  * Build cache system for performance optimization
  * Safe command execution (prevents injection attacks)
  * Cleanup and error handling with traps
- `lib/config.sh` - Configuration management system:
  * User preferences in ~/.gearboxrc (10 configurable settings)
  * Default build types, parallel job limits, caching options
  * Interactive configuration wizard and CLI management
- `lib/doctor.sh` - Health check and diagnostic system:
  * Comprehensive system validation (OS, memory, disk, internet)
  * Installed tool verification and coverage analysis
  * Environment and permission checks

**Main Installation Script (`scripts/install-all-tools.sh`)**:
- Orchestrates tool installation in optimal dependency order for all 30 tools
- Supports three build types: minimal, standard, maximum (configurable via ~/.gearboxrc)
- Handles common dependency installation via `install-common-deps.sh`
- Installation order optimized for shared toolchains: Go tools → Rust tools → C/C++ tools
- Progress indicators for multi-tool installations
- Includes confirmation prompt when installing all tools (30-60 minute process)
- Configuration-aware defaults (respects user preferences)

**Individual Tool Scripts (`scripts/install-*.sh`)**:
- Each tool has a dedicated installation script following consistent patterns
- All scripts migrated to use lib/common.sh (eliminated code duplication)
- Common command-line interface with build type flags (-m, -r, -o, etc.)
- Support for --skip-deps, --run-tests, --force flags
- Build from source with proper dependency validation
- Integrated build cache system for faster reinstallations
- Safe command execution (no eval usage, array-based commands)
- Root prevention checks for security

### Available Tools (30 total)

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
- `test-runner.sh` validates script executability and configuration loading
- Individual tools support `--run-tests` flag for post-build validation  
- Final verification checks that all installed tools are accessible via command line
- Health check system provides ongoing validation: `gearbox doctor`

## Key Implementation Patterns

### Script Structure
All installation scripts follow consistent patterns:
- Source `lib/common.sh` for shared functions, logging, and utilities
- Standardized command-line argument parsing with build type flags
- Root user prevention checks for security
- Build cache integration for performance optimization
- Safe command execution (no eval, array-based commands)
- Parallel build support with optimal CPU utilization
- Progress indicators for user feedback
- Error handling with cleanup traps
- Install dependencies (unless `--skip-deps` specified)
- Clone source to build directories outside script directory
- Configure build based on build type (minimal/standard/maximum, configurable)
- Build and install to `/usr/local/bin/` or `~/.cargo/bin/`
- Cache successful builds for future use
- Optional post-install testing and shell integration

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