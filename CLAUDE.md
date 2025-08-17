# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Essential Tools Installer - a collection of automated installation scripts for essential command-line tools on Debian Linux. The project focuses on building tools from source with various optimization levels and proper dependency management.

## Common Commands

### Installation Commands
- `gearbox` - Install all tools with standard builds
- `gearbox --minimal` - Install all tools with minimal/fast builds  
- `gearbox --maximum` - Install all tools with full-featured builds
- `gearbox fd ripgrep fzf` - Install only specified tools
- `gearbox --skip-common-deps` - Skip common dependency installation
- `gearbox --run-tests` - Run test suites for tools that support it

### Testing
- `./tests/test-runner.sh` - Run basic validation tests for installation scripts

### Individual Tool Installation
Each tool has its own script in `scripts/`:
- `scripts/install-fd.sh -r` - Install fd with release build
- `scripts/install-ripgrep.sh -o` - Install ripgrep with optimized build
- `scripts/install-fzf.sh -s --no-shell` - Install fzf with standard build, no shell integration

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
- Orchestrates tool installation in optimal dependency order
- Supports three build types: minimal, standard, maximum
- Handles common dependency installation via `install-common-deps.sh`
- Installation order optimized for shared toolchains: Go tools → Rust tools → C/C++ tools

**Individual Tool Scripts (`scripts/install-*.sh`)**:
- Each tool has a dedicated installation script following consistent patterns
- Common command-line interface with build type flags (-m, -r, -o, etc.)
- Support for --skip-deps, --run-tests, --force flags
- Build from source with proper dependency validation

### Available Tools
- **ffmpeg** - Video/audio processing (build flags: -m minimal, -g standard, -x maximum)
- **7zip** - Compression tool (build flags: -b basic, -o optimized, -a all-features)
- **jq** - JSON processor (build flags: -m minimal, -s standard, -o optimized)
- **fd** - Fast file finder (build flags: -m minimal, -r release)
- **ripgrep** - Fast text search (build flags: --no-pcre2 minimal, -r release, -o optimized)
- **fzf** - Fuzzy finder (build flags: -s standard, -p profiling)

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