# Essential Tools Installer

Automated installation scripts for essential command-line tools on Debian Linux. Build from source with optimized configurations and proper dependency management.

## Why This Project?

- **Source builds** with optimized configurations for best performance
- **Smart dependency management** - install once, share across tools
- **Multiple build types** - minimal, standard, or maximum feature sets
- **Battle-tested** - idempotent, safe, non-root installations

## Quick Start

```bash
# See what tools are available
gearbox list

# Install specific tools (recommended)
gearbox install fd ripgrep fzf jq

# Install with enhanced terminal experience
gearbox install fd ripgrep fzf nerd-fonts starship

# Install all tools (with confirmation prompt)
gearbox install

# Fast builds for quick setup
gearbox install --minimal fd ripgrep

# Check system health and disk usage
gearbox doctor

# Clean up build artifacts to save space
gearbox doctor cleanup --all --mode standard
```

## Building from Source

The gearbox CLI is built from Go source code. Build instructions:

```bash
# Build the CLI binary
make cli

# Build CLI + all components (orchestrator, tools)
make build

# Development setup (install dependencies)
make dev-setup

# Clean and rebuild
make clean
make build
```

**Requirements:**
- Go 1.22+ 
- Git
- Standard build tools (gcc, make)

**After building:** The `gearbox` binary will be available in the project root.

## Available Tools

| Tool | Description | Key Features |
|------|-------------|--------------|
| **fd** | Fast file finder | Intuitive syntax, parallel search |
| **ripgrep** | Fast text search | PCRE2 support, multi-line search |
| **fzf** | Fuzzy finder | Shell integration, preview support |
| **jq** | JSON processor | JSONPath queries, streaming |
| **zoxide** | Smart cd command | Frecency-based navigation |
| **yazi** | Terminal file manager | Vim-like keys, preview support |
| **fclones** | Duplicate file finder | Efficient deduplication, link replacement |
| **serena** | Coding agent toolkit | Semantic retrieval, IDE-like analysis |
| **uv** | Python package manager | 10-100x faster than pip, unified tooling |
| **ruff** | Python linter & formatter | 10-100x faster than Flake8/Black, 800+ rules |
| **bat** | Enhanced cat with syntax highlighting | Git integration, themes, automatic paging |
| **starship** | Customizable shell prompt | Fast, contextual info, Nerd Font support |
| **nerd-fonts** | Developer fonts with icons | Programming ligatures, file icons, Git symbols |
| **eza** | Modern ls replacement | Git integration, tree view, enhanced colors |
| **delta** | Syntax-highlighting pager | Git diff enhancement, word-level highlighting |
| **lazygit** | Terminal UI for Git | Interactive Git operations, visual interface |
| **bottom** | Cross-platform system monitor | Beautiful terminal resource monitoring |
| **procs** | Modern ps replacement | Enhanced process info, tree view, colors |
| **tokei** | Code statistics tool | Fast line counting for 200+ languages |
| **difftastic** | Structural diff tool | Syntax-aware code change analysis |
| **bandwhich** | Network bandwidth monitor | Terminal bandwidth utilization by process |
| **xsv** | CSV data toolkit | Fast CSV processing and analysis |
| **hyperfine** | Command-line benchmarking | Statistical command execution analysis |
| **gh** | GitHub CLI | Repository management, PRs, issues, workflows |
| **dust** | Better disk usage analyzer | Intuitive directory size visualization |
| **sd** | Find & replace CLI | Intuitive sed alternative for text replacement |
| **tealdeer** | Fast tldr client | Quick command help without full man pages |
| **choose** | Cut/awk alternative | Human-friendly text column selection |
| **ffmpeg** | Video/audio processing | Comprehensive codec support |
| **imagemagick** | Image manipulation | Powerful processing toolkit |
| **7zip** | Compression tool | High compression ratios |

## Disk Space Management

Building 31+ tools from source can consume 8GB+ of disk space. Gearbox provides intelligent cleanup:

```bash
# Show disk usage report
gearbox doctor cleanup

# Clean specific tools (recommended)
gearbox doctor cleanup ruff yazi bottom

# Clean all tools with standard cleanup
gearbox doctor cleanup --all --mode standard

# Aggressive cleanup for maximum space savings
gearbox doctor cleanup --all --mode aggressive --dry-run  # Preview first
gearbox doctor cleanup --all --mode aggressive           # Execute

# Enable automatic cleanup after installations
gearbox doctor cleanup --auto-cleanup
```

**Cleanup Modes:**
- **Minimal**: Remove temp files only (~5% space savings)
- **Standard**: Remove build artifacts, keep source (~50% space savings) 
- **Aggressive**: Remove everything except source code (~90% space savings)

**Smart Detection**: Automatically detects installation patterns and preserves essential files while removing build waste.

## Documentation

### For Users
📖 **[User Guide](docs/USER_GUIDE.md)** - Complete installation guide, usage examples, and troubleshooting

💾 **[Disk Space Management](docs/DISK_SPACE_MANAGEMENT.md)** - Comprehensive cleanup and optimization guide

🔧 **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Solutions for common issues and error diagnostics

### For Developers
🛠 **[Installation Methods](docs/INSTALLATION_METHODS.md)** - Technical details on installation patterns and design decisions

🏗 **[Developer Guide](docs/DEVELOPER_GUIDE.md)** - Architecture, adding tools, and development guidelines

👥 **[Contributing](CONTRIBUTING.md)** - Quick start for contributors

## Key Features

### Installation & Build System
- **Three build types**: Minimal (fast) → Standard (balanced) → Maximum (full-featured)
- **Optimal installation order**: Shared toolchains (Go → Rust → C/C++)
- **Multiple installation patterns**: Cargo install, direct copy, official installers
- **Build cache system**: Faster reinstallations with SHA256 integrity verification
- **Fail-fast approach**: Clear error messages, no hidden fallback failures

### Safety & Reliability  
- **Non-root execution**: Secure user-space installations
- **Existing installation detection**: Smart handling of pre-installed tools
- **Binary name resolution**: Metadata-driven detection (bottom→btm, ripgrep→rg)
- **Installation verification**: Comprehensive health checks and diagnostics

### Disk Space Management
- **Intelligent cleanup**: Three modes from minimal to aggressive cleanup
- **Smart detection**: Preserves installed binaries while removing build waste
- **Automatic cleanup**: Optional post-installation artifact removal
- **Space monitoring**: Real-time disk usage reports and recommendations

### User Experience
- **Shell integration**: Automatic setup for enhanced tools (fzf, starship, zoxide)
- **Progress tracking**: Real-time installation progress and status updates
- **Comprehensive CLI**: Unified interface for installation, health checks, and cleanup
- **Extensive documentation**: Guides for users, troubleshooting, and developers
- **Clean architecture**: Source builds in `~/tools/build/`, binaries in `/usr/local/bin/`

## Requirements

- Debian Linux (or derivatives)
- Internet connection for downloading sources
- `sudo` access for system package installation

---

**Ready to get started?** See the [User Guide](docs/USER_GUIDE.md) for comprehensive installation instructions and examples.