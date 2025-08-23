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

After cloning the repository:

```bash
# Build the project
make build

# See what tools are available
./build/gearbox list

# Install specific tools (recommended)
./build/gearbox install fd ripgrep fzf jq

# Install with enhanced terminal experience
./build/gearbox install fd ripgrep fzf nerd-fonts starship

# Install all tools (with confirmation prompt)
./build/gearbox install

# Fast builds for quick setup
./build/gearbox install --minimal fd ripgrep

# For convenience, install gearbox system-wide
make install
# Then use: gearbox list, gearbox install, etc.

# Launch the interactive TUI (NEW!)
gearbox tui

### 📦 Tool Bundles

Install curated collections of tools with a single command:

```bash
# Install essential tools bundle (core terminal productivity tools)
./build/gearbox install --bundle essential

# 🎯 Foundation Tier - Start your journey here!
./build/gearbox install --bundle system-foundation  # Essential base system packages
./build/gearbox install --bundle beginner           # New developers start here
./build/gearbox install --bundle intermediate       # Ready for git workflows and development tools
./build/gearbox install --bundle advanced           # High-performance development with debugging tools

# 🔄 Multi-Language Development Environment 
./build/gearbox install --bundle polyglot-dev               # Python + Node.js + Docker + Cloud + Editors

# 🏗️ Domain Tier - Choose your role
./build/gearbox install --bundle fullstack-dev      # Complete web development (frontend + backend + databases)
./build/gearbox install --bundle mobile-dev         # Cross-platform mobile development
./build/gearbox install --bundle data-dev           # Data science + ML environment
./build/gearbox install --bundle devops-dev         # Infrastructure + monitoring + deployment + container tools
./build/gearbox install --bundle security-dev       # Penetration testing + vulnerability scanning + container security
./build/gearbox install --bundle game-dev           # Game development + graphics tools

# 🚀 Language Ecosystems - Choose your language
./build/gearbox install --bundle python-dev         # Complete Python development environment
./build/gearbox install --bundle nodejs-dev         # Complete Node.js development environment  
./build/gearbox install --bundle rust-dev           # Complete Rust development environment
./build/gearbox install --bundle go-dev             # Complete Go development environment
./build/gearbox install --bundle java-dev           # Complete Java development environment
./build/gearbox install --bundle ruby-dev           # Complete Ruby development environment
./build/gearbox install --bundle cpp-dev            # Complete C/C++ development environment

# ⚙️ Workflow Tools - Add specialized capabilities
./build/gearbox install --bundle debugging-tools    # Profilers + memory analyzers + network debugging
./build/gearbox install --bundle deployment-tools   # CI/CD + containers + cloud deployment  
./build/gearbox install --bundle code-review-tools  # Cross-language linting + formatting + analysis
./build/gearbox install --bundle ai-tools           # AI-powered coding assistance (serena + aider)

# 🏗️ Infrastructure Tools - System components
./build/gearbox install --bundle docker-official    # Docker CE from official repository (recommended)
./build/gearbox install --bundle docker-enhanced    # Complete Docker ecosystem with analysis tools
./build/gearbox install --bundle docker-rootless    # Maximum security with rootless mode
./build/gearbox install --bundle cloud-tools        # AWS CLI and cloud platform tools
./build/gearbox install --bundle editors            # Neovim and modern text editors
./build/gearbox install --bundle database-tools     # Database clients and management tools
./build/gearbox install --bundle network-tools      # Network monitoring and diagnostics
./build/gearbox install --bundle media-tools        # Media processing tools

# List all available bundles
gearbox list bundles

# Show what's in a bundle (including system packages)
gearbox show bundle fullstack-dev
```

**Available Bundles (32 total, organized by tier):**

**🎯 Foundation Tier (4 bundles) - User Journey:**
- `system-foundation` - Essential base system packages for development environments
- `beginner` - Perfect starting point for new developers (essential tools + beautiful terminal)
- `intermediate` - Productive developer environment with git workflow and development tools
- `advanced` - High-performance development environment with debugging and performance tools

**🏗️ Domain Tier (7 bundles) - Choose Your Role:**
- `polyglot-dev` - Multi-language development environment (Python + Node.js + Docker + Cloud + Editors)
- `fullstack-dev` - Complete web development (frontend + backend + databases)
- `mobile-dev` - Cross-platform mobile development environment
- `data-dev` - Data science and machine learning development environment
- `devops-dev` - Infrastructure, monitoring, deployment + modern container tools
- `security-dev` - Security analysis, penetration testing + container security
- `game-dev` - Game development environment with graphics and engine tools

**🚀 Language Tier (7 bundles) - Complete Ecosystems:**
- `python-dev` - Python runtime + uv, ruff, black, mypy, poetry, pytest, ipython + essential tools
- `nodejs-dev` - Node.js runtime + TypeScript, ESLint, yarn, pnpm, jest + essential tools
- `go-dev` - Go compiler + gopls, golangci-lint, air, staticcheck, delve + essential tools
- `rust-dev` - Rust compiler + rustfmt, clippy, rust-analyzer, cargo tools + essential tools  
- `java-dev` - Java 17 + Maven, Gradle + essential tools
- `ruby-dev` - Ruby runtime + Rails, RSpec, RuboCop, Solargraph + essential tools
- `cpp-dev` - GCC/Clang + CMake, Ninja, GDB, Valgrind, Conan, vcpkg + essential tools

**⚙️ Workflow Tier (4 bundles) - Specialized Capabilities:**
- `debugging-tools` - Profilers, memory analyzers, and network debugging tools
- `deployment-tools` - CI/CD, containers, and cloud deployment tools with security
- `code-review-tools` - Code linting, formatting, and analysis tools (cross-language)
- `ai-tools` - AI-powered coding assistance (serena + aider + mise + just)

**🏗️ Infrastructure Tier (7 bundles) - System Components:**
- `docker-official` - Docker CE from official repository (2024 best practice)
- `docker-enhanced` - Complete Docker ecosystem with analysis and security tools
- `docker-rootless` - Maximum security with rootless Docker mode
- `cloud-tools` - AWS CLI v2 and cloud platform tools
- `editors` - Neovim and modern text editors
- `database-tools` - Database clients and management tools
- `network-tools` - Network monitoring and diagnostics toolkit
- `media-tools` - Media processing tools (ffmpeg, imagemagick, 7zip)

**🔧 Legacy Tier (2 bundles) - Simple Essentials:**
- `minimal` - Bare essentials (fd, ripgrep, fzf)
- `essential` - Modern terminal essentials everyone should have

**🎯 User Journey Architecture Design:**

The bundle system follows a **5-Tier Architecture** that matches how developers actually work:

**Foundation Tier** - Progressive skill levels:
- **system-foundation**: Base system packages for any development
- **beginner**: New developers start with essential tools + beautiful terminal
- **intermediate**: Add git workflows and development productivity tools  
- **advanced**: Add performance tools, debugging, and system analysis

**Domain Tier** - Role-based environments:
- Choose bundles based on your primary role (fullstack, data, devops, security, etc.)
- Each domain includes tools and packages specific to that workflow
- Inherits from appropriate foundation level for complete environments

**Language Tier** - Complete language ecosystems:
- Each includes runtime + language-specific tools + testing frameworks
- **Professional tooling included** (linting, formatting, testing, debugging)
- Self-contained ecosystems for productive development

**Workflow Tier** - Cross-language capabilities:
- Add specialized workflows like debugging, deployment, code review, AI assistance
- Work across multiple programming languages and domains

**Infrastructure Tier** - System components:
- Docker variants, cloud tools, editors, database tools, network tools
- Mix and match based on your infrastructure needs
- Can be added to any development environment

**Key Benefits:**
- **Clear progression path** from beginner → advanced → specialized
- **No tool duplication** - each tool appears once in its logical tier
- **Role-based approach** - install what you actually need for your job
- **Highly composable** - mix and match tiers for custom environments
- **32 focused bundles** instead of 44+ overlapping ones

## 🐳 Docker Installation Migration (2024 Update)

**IMPORTANT: Gearbox Docker support has been modernized to follow 2024 best practices.**

### **What Changed:**
- ❌ **Old**: `docker.io` package (Ubuntu repository, outdated)
- ✅ **New**: `docker-ce` from official Docker repository (latest, secure)
- ❌ **Old**: `docker-compose` v1 (deprecated)
- ✅ **New**: `docker compose` v2 (built into Docker CE)
- ❌ **Old**: Manual sudo required
- ✅ **New**: Proper user permissions (docker group)

### **Migration Guide:**

**If you previously used docker bundles:**
```bash
# Remove old Docker installation
sudo apt-get purge docker docker.io docker-compose

# Install modern Docker
./build/gearbox install --bundle docker-official

# Or for complete ecosystem
./build/gearbox install --bundle docker-enhanced

# For maximum security
./build/gearbox install --bundle docker-rootless
```

**Key Differences:**
- Uses `docker compose` (space) instead of `docker-compose` (hyphen)
- No sudo required after proper installation
- Latest security updates from Docker's official repository
- Built-in BuildX and Compose v2 support

**Example User Journey:**
```bash
# New developer starting out
./build/gearbox install --bundle beginner

# Ready for more productivity, choose your domain
./build/gearbox install --bundle fullstack-dev  # or mobile-dev, data-dev, devops-dev, etc.

# Add language-specific tools  
./build/gearbox install --bundle python-dev

# Add specialized workflows as needed
./build/gearbox install --bundle deployment-tools
```

# Check system health and disk usage
gearbox doctor

# Tool-specific diagnostics
gearbox doctor nerd-fonts         # Font installation and availability
gearbox doctor zoxide             # Navigation database and shell integration
gearbox doctor zoxide --verbose   # Detailed database contents and performance

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

## 🎨 Interactive TUI (Text User Interface)

Gearbox now includes a comprehensive TUI for an intuitive, visual experience:

```bash
# Launch the TUI
gearbox tui
```

### TUI Features

**📊 Dashboard View**
- System overview with installation statistics
- Recent activity tracking
- Smart recommendations based on your setup
- Quick actions for common tasks

**🔍 Tool Browser**  
- Search and filter 42+ available tools
- Real-time search across names, descriptions, and languages
- Multi-selection with Space key
- Category filtering (Core, Development, System, etc.)
- Side-by-side tool preview with details

**📦 Bundle Explorer**
- Hierarchical view of 32 curated bundles
- Organized by tiers (Foundation, Domain, Language, Workflow, Infrastructure)
- Expandable details showing included tools
- Installation status tracking
- One-click bundle installation

**🚀 Install Manager**
- Real-time installation progress
- Concurrent task management
- Live output streaming
- Progress bars with stage information
- Cancel/retry capabilities

**⚙️ Configuration**
- Interactive settings management
- Edit build types, parallel jobs, cache settings
- Type-safe configuration with validation
- Reset to defaults option

**🏥 Health Monitor**
- Comprehensive system health checks
- Tool installation coverage analysis
- Toolchain verification (Rust, Go, etc.)
- Smart suggestions for issues
- Auto-refresh capability

### TUI Navigation

- **Tab**: Switch between views
- **↑/↓**: Navigate lists
- **Enter**: Select/Confirm
- **Space**: Toggle selection
- **/**: Search (in Tool Browser)
- **?**: Help screen
- **q**: Quit

### Quick View Shortcuts

- **D**: Dashboard
- **T**: Tool Browser
- **B**: Bundle Explorer  
- **I**: Install Manager
- **C**: Configuration
- **H**: Health Monitor

The TUI complements the CLI, providing a visual interface for complex operations while maintaining all CLI functionality.

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
| **dive** | Docker image layer analyzer | Optimize image sizes, inspect layers |
| **trivy** | Container vulnerability scanner | Security scanning for containers and filesystems |
| **podman** | Docker alternative container engine | Rootless, daemonless container management |
| **lazydocker** | Docker TUI management tool | Terminal interface for Docker operations |
| **hadolint** | Dockerfile linter | Best practices and security for Dockerfiles |
| **ctop** | Container monitoring TUI | Real-time container resource monitoring |
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

Building 42+ tools from source can consume 8GB+ of disk space. Gearbox provides intelligent cleanup:

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

🧪 **[Testing Guide](docs/TESTING_GUIDE.md)** - Testing framework, validation procedures, and quality assurance

👥 **[Contributing](CONTRIBUTING.md)** - Quick start for contributors

### Technical Documentation
📋 **[CLAUDE.md](CLAUDE.md)** - Comprehensive technical documentation for Claude Code integration

🔄 **[Configuration Migration](docs/CONFIGURATION_MIGRATION.md)** - Configuration system and user preferences

🚀 **[Go Migration Plan](docs/GO_MIGRATION_PLAN.md)** - Strategic shell-to-Go migration roadmap

**🎯 Recent Improvements (2024)**: The project has undergone significant enhancements including comprehensive test coverage (450+ test cases), modular architecture refactoring (98% code size reduction), enhanced build cache system, and modern Go practices implementation. See [CLAUDE.md](CLAUDE.md) for detailed technical improvements.

## Key Features

### Installation & Build System
- **Three build types**: Minimal (fast) → Standard (balanced) → Maximum (full-featured)
- **Optimal installation order**: Shared toolchains (Go → Rust → C/C++)
- **Multiple installation patterns**: Cargo install, direct copy, official installers
- **Enhanced build cache system**: Complete caching with metadata, validation, statistics, and cleanup
- **Comprehensive testing**: 450+ test cases covering all Go packages with benchmarks
- **Modular architecture**: Refactored from monolithic to focused, maintainable modules
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
- **Interactive TUI**: Full-featured text interface with dashboard, browser, and real-time monitoring
- **Shell integration**: Automatic setup for enhanced tools (fzf, starship, zoxide)
- **Progress tracking**: Real-time installation progress and status updates
- **Comprehensive CLI**: Unified interface for installation, health checks, and cleanup
- **Extensive documentation**: Guides for users, troubleshooting, and developers
- **Clean architecture**: Source builds in `~/tools/build/`, binaries in `/usr/local/bin/`

### Modern Architecture
- **Modular Design**: Clean separation of concerns with focused Go packages and shell modules
- **Comprehensive Test Coverage**: 450+ test cases across all Go packages with benchmarks and edge case coverage
- **Refactored Codebase**: Orchestrator split from 2250 lines to focused modules (98% size reduction)
- **Enhanced Build Cache**: Complete caching system with metadata, validation, and automatic cleanup
- **Type-Safe Operations**: Structured error handling, manifest tracking, and configuration management
- **Modern Go Practices**: Updated to current standards, eliminated deprecated functions
- **Organized Structure**: Installation scripts categorized by functionality (core, development, system, text, media, ui)  
- **Lazy Loading**: Efficient module loading for optimal performance
- **Security-First**: Command injection prevention, root user protection, safe execution patterns

## Requirements

- Debian Linux (or derivatives)
- Internet connection for downloading sources
- `sudo` access for system package installation

---

**Ready to get started?** See the [User Guide](docs/USER_GUIDE.md) for comprehensive installation instructions and examples.