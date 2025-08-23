# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with this Essential Tools Installer repository.

## Project Overview

This is an automated installer for essential command-line tools on Debian Linux, focusing on building tools from source with optimized configurations and proper dependency management. The project provides both CLI and TUI interfaces for tool installation and management.

## Quick Start

After building the project:
```bash
make build                                    # Build all components
./build/gearbox --help                       # CLI help
./build/gearbox list                         # Show available tools  
./build/gearbox install fd ripgrep fzf      # Install specific tools (recommended)
./build/gearbox tui                          # Launch interactive TUI
```

## Essential Commands

### Core Operations
- `./build/gearbox list` - Show available tools and bundles
- `./build/gearbox install <tools...>` - Install specific tools (recommended approach)
- `./build/gearbox install --bundle <name>` - Install bundle (e.g., beginner, python-dev)
- `./build/gearbox doctor` - Run comprehensive health checks
- `./build/gearbox config show` - Display configuration settings
- `./build/gearbox tui` - Launch interactive TUI interface

### Installation Examples
```bash
# Specific tools (recommended)
./build/gearbox install fd ripgrep fzf

# Build type variants
./build/gearbox install --minimal fd ripgrep     # Fast builds
./build/gearbox install --maximum ffmpeg         # Full-featured builds

# Bundle installations
./build/gearbox install --bundle beginner        # Essential tools + terminal
./build/gearbox install --bundle python-dev      # Python ecosystem
./build/gearbox install --bundle rust-dev        # Rust ecosystem

# Font management
./build/gearbox install nerd-fonts               # Standard collection (8 fonts)
./build/gearbox install nerd-fonts --minimal     # Essential fonts (3 fonts)
./build/gearbox install nerd-fonts --interactive # Interactive selection
```

### Bundle Tiers
- **Foundation**: `beginner`, `intermediate`, `advanced` - Progressive developer environments  
- **Language**: `python-dev`, `nodejs-dev`, `rust-dev`, `go-dev`, `java-dev`, etc.
- **Domain**: `fullstack-dev`, `devops-dev`, `security-dev`, `data-dev`, etc.
- **Workflow**: `debugging-tools`, `deployment-tools`, `code-review-tools`
- **Infrastructure**: `docker`, `cloud-tools`, `ai-tools`, `editors`, `media-tools`

### Configuration & Diagnostics
```bash
./build/gearbox config wizard                    # Interactive setup
./build/gearbox config set DEFAULT_BUILD_TYPE maximum
./build/gearbox doctor                           # Comprehensive health checks
./build/gearbox doctor system                    # System requirements check
./build/gearbox doctor nerd-fonts               # Advanced font diagnostics (cache, terminal, VS Code, starship)
./build/gearbox doctor zoxide                    # Navigation tool diagnostics (database, shell integration, performance)
./build/gearbox doctor zoxide --verbose          # Detailed analysis with database contents and optimization
./build/gearbox status nerd-fonts               # Detailed tool status
```

### Build System
```bash
make build                                       # Build all components
make cli                                         # Build CLI only
make test                                        # Run all tests
make clean                                       # Clean build artifacts
make install                                     # System-wide installation
```

## Architecture

### Directory Structure
```
├── cmd/                     # Go command entry points
│   ├── gearbox/            # Main CLI (Cobra framework)
│   │   ├── commands/       # CLI command implementations
│   │   └── tui/           # Complete TUI implementation (Bubble Tea)
│   ├── orchestrator/       # Advanced installation orchestrator
│   ├── script-generator/   # Template-based script generator
│   └── config-manager/     # Configuration management tool
├── pkg/                     # Go packages (shared, reusable code)
│   ├── orchestrator/       # Installation orchestration logic (modular)
│   ├── manifest/          # Installation tracking and state management
│   ├── uninstall/         # Safe removal engine with dependency analysis
│   ├── errors/            # Structured error handling with context
│   ├── logger/            # Centralized structured logging
│   └── validation/        # Input validation utilities
├── scripts/                # Shell implementation
│   ├── lib/               # Modular shared library system
│   │   ├── common.sh      # Entry point with lazy loading
│   │   ├── core/          # Essential modules (logging, validation, security)
│   │   ├── build/         # Build system modules (cache, dependencies)
│   │   └── system/        # System integration modules
│   └── installation/      # Tool installation scripts by category
│       └── categories/    # core/, development/, system/, text/, media/, ui/
├── config/                 # Configuration files
│   ├── tools.json         # Tool definitions with metadata and build types
│   └── bundles.json       # Bundle definitions and relationships
├── build/                  # Compiled binaries (Go best practice)
├── templates/              # Script generation templates
├── docs/                   # Comprehensive documentation library
└── tests/                  # Comprehensive test suite
```

### Available Tools (42 total)

**Core Development (9 tools)**:
- fd, ripgrep, fzf, jq, zoxide - Essential command-line tools
- bat, eza, dust, yazi - Enhanced file operations

**Development & Git (10 tools)**:  
- serena, uv, ruff - Python ecosystem
- starship, nerd-fonts - Terminal enhancement
- delta, lazygit, gh, difftastic, hyperfine - Git workflow and analysis

**System & Text (14 tools)**:
- bottom, procs, bandwhich - System monitoring
- sd, xsv, choose, tealdeer, tokei - Text processing and analysis
- fclones - Duplicate file management

**Media Processing (3 tools)**:
- ffmpeg, imagemagick, 7zip - Media and archive handling

### Build System Features

**Modular Library System**: 
- `scripts/lib/common.sh` - Entry point with lazy loading
- Core modules: logging, validation, security, utilities
- Build modules: dependencies, execution, cache, cleanup
- System modules: installation, backup, environment

**Build Types**: All tools support minimal/standard/maximum builds
- **minimal**: Fast builds, basic features
- **standard**: Balanced builds (default, configurable via ~/.gearboxrc)
- **maximum**: Full-featured builds with all optimizations

**Performance Optimizations**:
- Build cache system with automatic cleanup
- Parallel builds with memory-aware job limits  
- Progress indicators for multi-tool installations
- Optimal dependency installation order

**Security & Safety**:
- Command injection prevention (no eval usage)
- Root user prevention checks
- Input validation and sanitization
- Safe array-based command execution

### TUI Interface

**Bubble Tea Framework**: Complete interactive interface with 6 main views
- **Dashboard**: System overview, installed tools, smart recommendations
- **Tool Browser**: Search, filter, multi-select tools with real-time preview
- **Bundle Explorer**: Hierarchical bundle display with installation progress
- **Install Manager**: Real-time installation progress with output streaming
- **Configuration**: Interactive settings editing with validation
- **Health Monitor**: Comprehensive system health checks and diagnostics

**Navigation**: Tab switching, arrow keys, search (`/`), help (`?`), quit (`q`)

### Advanced Features

**Nerd Fonts Management**:
- Three collections: minimal (3 fonts), standard (8 fonts), maximum (15+ fonts)
- Interactive selection with live character previews
- Auto-configuration for VS Code, Kitty, Alacritty
- Cross-tool intelligence (suggests starship, detects terminal types)
- Comprehensive health checks and diagnostics

**Configuration Management**:
- User preferences in `~/.gearboxrc` (10 configurable settings)
- Interactive configuration wizard
- CLI configuration management
- Build type defaults, parallel job limits, caching options

**Health Check System**:
- System requirements validation (OS, memory, disk, internet)
- Tool installation verification and coverage analysis
- Environment and permission checks
- Actionable recommendations for issues

## Testing Strategy

**Comprehensive Coverage**:
- **Go Tests**: 450+ test cases across all packages with benchmarks
- **Shell Tests**: Multi-layered framework with 50+ function coverage
- **Integration Tests**: Multi-tool workflow and TUI testing
- **Security Tests**: Command injection, path traversal prevention
- **Performance Tests**: Function timing, memory usage analysis

**Test Execution**:
```bash
go test ./... -v                              # All Go tests
make test                                     # All tests (Go + shell)
./tests/test-runner.sh                       # Basic validation
./tests/framework/test-framework.sh          # Comprehensive framework
./tests/tui-test-runner.sh                   # TUI integration tests
```

## Key Implementation Patterns

### Standardized Script Protocol

All installation scripts follow the **Gearbox Standard Protocol v1.0** for consistent behavior:

**Required Interface**: All scripts MUST support these standardized flags:
```bash
# Build Types (mutually exclusive)
--minimal        # Fast build with essential features only
--standard       # Balanced build with reasonable features (default)
--maximum        # Full-featured build with all optimizations

# Execution Modes (mutually exclusive) 
--config-only    # Configure only (prepare build environment)
--build-only     # Configure and build (no installation)
--install        # Configure, build, and install (default)

# Common Options
--skip-deps      # Skip dependency installation
--force          # Force reinstallation if already installed
--run-tests      # Run test suite after building (if applicable)
--no-shell       # Skip shell integration setup
--dry-run        # Show what would be done without executing
--verbose        # Enable verbose output
--quiet          # Suppress non-error output
--help           # Show usage information
--version        # Show script version information
```

**Graceful Degradation**: Scripts accept all standard flags and either implement them or silently ignore them. Unknown flags produce warnings but don't fail.

**Orchestrator Integration**: The orchestrator uses standardized flags instead of tool-specific build flags, eliminating configuration mismatches like the previous zoxide `-s` flag issue.

### Script Structure
All installation scripts follow consistent patterns:
- Source modular `scripts/lib/common.sh` with automatic core module loading
- Implement Gearbox Standard Protocol v1.0 for argument parsing
- Build cache integration for performance optimization
- Safe command execution with comprehensive security validation
- Progress indicators and error handling with cleanup traps

### CLI Interface Design
- **Safety first**: No auto-execution without user awareness
- **Specific over general**: Recommend specific tool lists over "install all"
- **Configuration-aware**: Respects user preferences from ~/.gearboxrc
- **Comprehensive help**: Clear documentation and command examples

### Go Architecture
- **Cobra framework**: Type-safe commands with structured error handling
- **Modular orchestrator**: Split from monolithic 2250-line file to focused modules
- **Error handling**: Structured errors with context and suggestions
- **Modern practices**: Eliminated deprecated `ioutil.*` functions

## Development Workflow

### Adding New Tools
1. Define metadata in `config/tools.json` (name, repository, build types, dependencies)
2. Generate script: `./build/gearbox generate <tool-name>`
3. Test with `--dry-run` and validation tests
4. Tool automatically appears in `./build/gearbox list`

### Modifying CLI Commands  
1. Edit files in `cmd/gearbox/commands/`
2. Build and test: `make cli && ./build/gearbox --help`
3. Run tests: `make test`
4. Update help text and examples

### Architecture Migration Strategy
- **Go**: Complex logic, parallel coordination, structured error handling
- **Shell**: Tool builds, system integration, domain expertise
- **Hybrid**: Type-safe orchestration with shell execution for builds

## Documentation Structure

**Quick Reference**:
- `README.md` - Project overview and quick start
- `docs/USER_GUIDE.md` - Complete end-user manual
- `docs/DEVELOPER_GUIDE.md` - Developer setup and contribution guidelines
- `docs/TESTING_GUIDE.md` - Testing framework documentation
- `docs/TROUBLESHOOTING.md` - Common issues and solutions
- `docs/TUI_GUIDE.md` - Text User Interface documentation

**Usage Guidance**:
- **New Users**: README.md → docs/USER_GUIDE.md
- **Developers**: docs/DEVELOPER_GUIDE.md → CLAUDE.md → docs/TESTING_GUIDE.md  
- **Issues**: docs/TROUBLESHOOTING.md first
- **TUI Help**: docs/TUI_GUIDE.md for interface documentation