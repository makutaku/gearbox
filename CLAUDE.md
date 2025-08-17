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
- `scripts/` - Individual installation scripts for each tool
- `config.sh` - Shared configuration and utility functions
- `gearbox` - Main CLI script that delegates to `scripts/install-all-tools.sh`
- `docs/` - Documentation files
- `examples/` - Example usage scripts
- `tests/` - Basic validation tests

### Build System Architecture

**Configuration Layer (`config.sh`)**:
- Defines build directories (`~/tools/build/`), cache (`~/tools/cache/`), and install prefix (`/usr/local`)
- Provides shared logging functions and color definitions
- Contains utility functions for path management

**Main Installation Script (`scripts/install-all-tools.sh`)**:
- Orchestrates tool installation in optimal dependency order for all 30 tools
- Supports three build types: minimal, standard, maximum
- Handles common dependency installation via `install-common-deps.sh`
- Installation order optimized for shared toolchains: Go tools → Rust tools → C/C++ tools
- Includes confirmation prompt when installing all tools (30-60 minute process)

**Individual Tool Scripts (`scripts/install-*.sh`)**:
- Each tool has a dedicated installation script following consistent patterns
- Common command-line interface with build type flags (-m, -r, -o, etc.)
- Support for --skip-deps, --run-tests, --force flags
- Build from source with proper dependency validation

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
- **standard**: Balanced builds with reasonable features (default)
- **maximum**: Full-featured builds with all optimizations

### Testing Strategy
- `test-runner.sh` validates script executability and configuration loading
- Individual tools support `--run-tests` flag for post-build validation  
- Final verification checks that all installed tools are accessible via command line

## Key Implementation Patterns

### Script Structure
All installation scripts follow consistent patterns:
- Source `config.sh` for shared configuration and logging functions
- Parse command-line arguments with standardized flags
- Check for existing installations to avoid duplicates
- Install dependencies (unless `--skip-deps` specified)
- Clone source to `~/tools/build/[tool-name]/`
- Configure build based on build type (minimal/standard/maximum)
- Build and install to `/usr/local/bin/`
- Optional post-install testing and shell integration

### CLI Interface Design
- **Safety first**: No command auto-executes without user awareness
- **Confirmation prompts**: Installing all tools shows impact and requires confirmation  
- **Specific over general**: `gearbox install fd ripgrep` recommended over `gearbox install`
- **Clear help**: `gearbox list` shows all tools, `gearbox help` shows usage

### Build System Integration
- **Shared toolchains**: Common dependencies installed once (Rust 1.88.0+, Go 1.23.4+)
- **Optimized order**: Go tools first, then Rust tools, then C/C++ tools
- **Multiple build types**: Tools support different optimization levels via standardized flags
- **Clean separation**: Source builds in `~/tools/build/`, final binaries in `/usr/local/bin/`