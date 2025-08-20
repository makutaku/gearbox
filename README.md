# Essential Tools Installer

Automated installation scripts for essential command-line tools on Debian Linux. Build from source with optimized configurations, proper dependency management, and convenient tool bundles for quick setup.

## Why This Project?

- **Source builds** with optimized configurations for best performance
- **Smart dependency management** - install once, share across tools
- **Tool bundles** - curated collections for different workflows (developer, data-science, etc.)
- **Multiple build types** - minimal, standard, or maximum feature sets
- **Battle-tested** - idempotent, safe, non-root installations

## Quick Start

### ⚡ One-Line Install (Recommended)

```bash
# Full installer (handles dependencies, builds CLI)
curl -fsSL https://raw.githubusercontent.com/makutaku/gearbox/main/install.sh | bash

# Minimal installer (if you have Go/Git already)
curl -fsSL https://raw.githubusercontent.com/makutaku/gearbox/main/install-minimal.sh | bash
```

### 🚀 Quick Setup Profiles

```bash
# After installation, choose a profile:
./quickstart.sh minimal     # Essential tools (fd, ripgrep, fzf)
./quickstart.sh developer   # Core dev tools + fonts (recommended)
./quickstart.sh full        # Complete terminal experience
```

### 🔧 Manual Usage

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

### 📦 Tool Bundles

Install curated collections of tools with a single command:

```bash
# Install essential tools bundle (core terminal productivity tools)
gearbox install --bundle essential

# 🎯 Foundation Tier - Start your journey here!
gearbox install --bundle beginner      # New developers start here
gearbox install --bundle intermediate  # Ready for git workflows and development tools
gearbox install --bundle advanced      # High-performance development with debugging tools

# 🏗️ Domain Tier - Choose your role
gearbox install --bundle frontend-dev       # HTML/CSS/JS + design tools
gearbox install --bundle backend-dev        # APIs + databases + servers
gearbox install --bundle data-scientist     # Data science + ML environment
gearbox install --bundle devops-engineer    # Infrastructure + monitoring + deployment
gearbox install --bundle security-researcher # Penetration testing + vulnerability scanning
gearbox install --bundle mobile-dev         # Cross-platform mobile development
gearbox install --bundle game-dev          # Game development + graphics tools

# 🚀 Language Ecosystems - Choose your language
gearbox install --bundle python-ecosystem  # Complete Python development environment
gearbox install --bundle nodejs-ecosystem  # Complete Node.js development environment  
gearbox install --bundle rust-ecosystem    # Complete Rust development environment
gearbox install --bundle go-ecosystem      # Complete Go development environment

# ⚙️ Workflow Tools - Add specialized capabilities
gearbox install --bundle debugging-tools   # Profilers + memory analyzers + network debugging
gearbox install --bundle deployment-tools  # CI/CD + containers + cloud deployment  
gearbox install --bundle code-review-tools # Cross-language linting + formatting + analysis

# 🤖 AI & Modern Development
gearbox install --bundle ai-coding-agent   # AI-powered coding assistance (serena + aider)
gearbox install --bundle modern-dev        # Modern workflow tools (mise + just + aider)
gearbox install --bundle media            # Media processing tools

# List all available bundles
gearbox list bundles

# Show what's in a bundle (including system packages)
gearbox show bundle web-dev
```

**Available Bundles:**

**🎯 Foundation Tier (User Journey):**
- `beginner` - Perfect starting point for new developers (essential tools + beautiful terminal)
- `intermediate` - Productive developer environment with git workflow and development tools
- `advanced` - High-performance development environment with debugging and performance tools

**🏗️ Domain Tier (Role-Based):**
- `frontend-dev` - Complete frontend development (HTML/CSS/JS + design tools)
- `backend-dev` - Complete backend development (APIs + databases + servers)
- `mobile-dev` - Cross-platform mobile development environment
- `data-scientist` - Data science and machine learning development environment
- `devops-engineer` - Infrastructure, monitoring, and deployment toolkit
- `security-researcher` - Security analysis, penetration testing, and vulnerability scanning toolkit
- `game-dev` - Game development environment with graphics and engine tools

**🚀 Language Ecosystem Tier:**
- `python-ecosystem` - Python runtime + pipx, black, flake8, mypy, poetry, pytest, jupyter + essential tools
- `nodejs-ecosystem` - Node.js runtime + TypeScript, ESLint, Angular/Vue/React CLIs, jest + essential tools
- `go-ecosystem` - Go compiler + gopls, golangci-lint, air, staticcheck, delve + essential tools
- `rust-ecosystem` - Rust compiler + rustfmt, clippy, rust-analyzer, cargo tools + essential tools  
- `java-ecosystem` - Java 17 + Maven, Gradle + essential tools
- `ruby-ecosystem` - Ruby runtime + Rails, RSpec, RuboCop, Solargraph + essential tools
- `cpp-ecosystem` - GCC/Clang + CMake, Ninja, GDB, Valgrind, Conan, vcpkg + essential tools

**⚙️ Workflow Tier:**
- `debugging-tools` - Profilers, memory analyzers, and network debugging tools
- `deployment-tools` - CI/CD, containers, and cloud deployment tools
- `code-review-tools` - Code linting, formatting, and analysis tools (cross-language)

**🤖 AI & Modern Development:**
- `ai-coding-agent` - AI-powered coding assistance (serena + aider) 
- `modern-dev` - Modern workflow tools (mise + just + aider)
- `media` - Media processing tools (ffmpeg, imagemagick, 7zip)

**🔧 Legacy Core Bundles:**
- `minimal` - Bare essentials (fd, ripgrep, fzf)
- `essential` - Modern terminal essentials everyone should have

**🎯 User Journey Architecture Design:**

The bundle system follows a **User Journey Architecture** that matches how developers actually work:

**Foundation Tier** - Progressive skill levels:
- **Beginner**: New developers start with essential tools + beautiful terminal
- **Intermediate**: Add git workflows and development productivity tools  
- **Advanced**: Add performance tools, debugging, and system analysis

**Domain Tier** - Role-based environments:
- Choose bundles based on your primary role (frontend dev, backend dev, data scientist, etc.)
- Each domain includes tools and packages specific to that workflow
- Inherits from appropriate foundation level

**Language Tier** - Complete language environments:
- Each includes runtime + language-specific tools + testing frameworks
- **Testing tools are included** in language ecosystems (not separate workflow) since they're language-specific

**Workflow Tier** - Cross-language capabilities:
- Add specialized workflows like debugging, deployment, code review
- Work across multiple programming languages

**Key Benefits:**
- **Clear user journey** from beginner → advanced
- **No tool duplication** - each tool appears once in logical place
- **Role-based approach** - install what you actually need for your job
- **Composable** - mix and match tiers as needed

**Example User Journey:**
```bash
# New developer starting out
gearbox install --bundle beginner

# Ready for more productivity, choose your domain
gearbox install --bundle frontend-dev  # or backend-dev, data-scientist, etc.

# Add language-specific tools  
gearbox install --bundle python-ecosystem

# Add specialized workflows as needed
gearbox install --bundle deployment-tools
```

# Check system health and disk usage
gearbox doctor

# Clean up build artifacts to save space
gearbox doctor cleanup --all --mode standard
```

## Building from Source

The gearbox CLI is built from Go source code with comprehensive testing. Build instructions:

```bash
# Build the CLI binary
make cli

# Build CLI + all components (orchestrator, tools)
make build

# Development setup (install dependencies)
make dev-setup

# Run comprehensive test suite
make test

# Clean and rebuild
make clean
make build
```

**Requirements:**
- Go 1.22+ 
- Git
- Standard build tools (gcc, make)

**After building:** The `gearbox` binary will be available in the project root.

### Testing & Quality Assurance

Gearbox includes comprehensive testing infrastructure:

```bash
# Run all tests (Go + Shell)
make test

# Run specific test suites
./tests/test_core_functions.sh           # Core function validation
./tests/test_workflow_integration.sh     # Multi-tool workflows  
./tests/test_performance_benchmarks.sh   # Performance analysis
./tests/test_error_handling.sh          # Security & resilience

# Run basic script validation
./tests/test-runner.sh
```

**Test Coverage:**
- 🔒 **Security Testing**: Command injection, path traversal, privilege escalation prevention
- ⚡ **Performance Benchmarking**: Function timing, memory usage, parallel execution analysis
- 🛡️ **Error Resilience**: Graceful failure handling, cleanup, rollback functionality
- 🔧 **Integration Testing**: Multi-tool workflows, dependency chains, CLI delegation
- ✅ **50+ Functions Tested**: Comprehensive validation across all modules

## Installation Features

The one-line installer provides:

- 🔍 **Smart Detection** - Automatically detects your OS and package manager
- 📦 **Dependency Management** - Installs Go, Git, build tools as needed  
- 🛡️ **Safety Checks** - Won't run as root, validates everything works
- 🎨 **Beautiful Output** - Color-coded progress with clear next steps
- ⚡ **Fast Setup** - Complete installation in under 2 minutes
- 🔄 **Update Support** - Re-run to update existing installations

**Supported Systems:**
- Ubuntu/Debian (apt)
- RHEL/CentOS (yum) 
- Fedora (dnf)
- Any Linux with Go 1.22+

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
| **serena** | AI coding agent toolkit (MCP server) | Semantic code analysis, multi-language AI assistance |
| **mise** | Multi-language version manager | Replaces nvm/pyenv/rbenv, unified tooling |
| **just** | Modern command runner | Simpler Make alternative, intuitive syntax |
| **aider** | AI pair programming assistant | Terminal-based AI coding, git integration |
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

### Modern Architecture
- **Modular Design**: Separated core functions (logging, validation, security, utilities)
- **Organized Structure**: Installation scripts categorized by functionality (core, development, system, text, media, ui)  
- **Lazy Loading**: Efficient module loading for optimal performance
- **Comprehensive Testing**: 50+ functions tested with security, performance, and integration validation

## Requirements

- Debian Linux (or derivatives)
- Internet connection for downloading sources
- `sudo` access for system package installation

---

**Ready to get started?** See the [User Guide](docs/USER_GUIDE.md) for comprehensive installation instructions and examples.