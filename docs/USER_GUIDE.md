# User Guide

Complete guide for installing and using essential command-line tools.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Building the CLI](#building-the-cli)
3. [Installation Options](#installation-options)
4. [Individual Tools](#individual-tools)
5. [Common Workflows](#common-workflows)
6. [Shell Integration](#shell-integration)
7. [Troubleshooting](#troubleshooting)
8. [Reference](#reference)

## Getting Started

### Discover Available Tools

See what tools are available before installing:

```bash
gearbox list
```

This shows all available tools with descriptions and usage examples.

### Recommended: Install Specific Tools

Install just the tools you need:

```bash
# Essential development tools
gearbox install fd ripgrep fzf jq

# Media processing tools  
gearbox install ffmpeg 7zip
```

### Install All Tools (With Confirmation)

Install all tools with optimal dependency management:

```bash
gearbox install
```

This will show a confirmation prompt before installing all 31 tools, since the process takes 30-60 minutes. The installer builds tools in optimal order, sharing dependencies efficiently.

### Selective Installation

Install only the tools you need:

```bash
# Essential development tools
gearbox install fd ripgrep fzf jq

# Media processing tools
gearbox install ffmpeg 7zip

# Just file tools
gearbox install fd ripgrep fzf
```

### Verify Installation

After installation, verify tools are available:

```bash
# Check installation
which fd rg fzf jq ffmpeg 7zz

# Check versions
fd --version && rg --version && fzf --version
```

If commands aren't found, refresh your shell:

```bash
hash -r && source ~/.bashrc
```

## Building the CLI

If you need to build the gearbox CLI from source (e.g., after git clone):

```bash
# Quick CLI build
make cli

# Build everything (CLI + orchestrator + tools)
make build

# Development setup (install Go dependencies)
make dev-setup

# Clean rebuild
make clean && make build
```

**Requirements:**
- Go 1.22+
- Git  
- Standard build tools (gcc, make, curl)

**Result:** Creates `./gearbox` binary in project root.

**First-time setup:**
```bash
git clone <repository>
cd gearbox
make build         # Build all components
./gearbox list     # Verify CLI works
```

## Installation Options

### Build Types

Each tool supports three build types:

- **Minimal** (`--minimal`): Fast builds with essential features only
- **Standard** (default): Balanced performance and features
- **Maximum** (`--maximum`): Full-featured builds with all optimizations

```bash
# Fast installation for quick setup
gearbox install --minimal

# Maximum features for production use
gearbox install --maximum fd ripgrep ffmpeg

# Standard build (default)
gearbox install fd ripgrep fzf
```

### Advanced Installation Options

```bash
# Skip dependency installation (if already installed)
gearbox install --skip-common-deps fd ripgrep

# Include test suites for validation
gearbox install --run-tests fd ripgrep jq

# Install without shell integration
gearbox install --no-shell fzf

# Combine options
gearbox install --minimal --run-tests fd ripgrep
```

### Dependencies Handled Automatically

The installer manages these dependencies:

**Programming Languages:**
- Rust ≥ 1.88.0 (for most tools: fd, ripgrep, bat, eza, zoxide, yazi, fclones, etc.)
- Go ≥ 1.23.4 (for fzf, lazygit, gh)
- Python ≥ 3.11 (for serena, uv, ruff)
- C/C++ toolchain (for ffmpeg, 7zip, jq, imagemagick)

**System Packages:**
- `build-essential`, `git`, `curl`, `wget`, `make`
- `autoconf`, `automake`, `libtool`
- `nasm`, `yasm` (for assembly optimizations)
- Codec libraries (for ffmpeg)

## Individual Tools

The installer provides 31 essential tools organized by category. Here are the most commonly used tools with installation and usage examples:

### Core Development Tools

#### fd - Fast File Finder

Modern replacement for `find` with intuitive syntax.

```bash
# Install with different build types
scripts/install-fd.sh              # Standard
scripts/install-fd.sh --minimal    # Faster build
scripts/install-fd.sh --release    # Optimized

# Usage examples
fd pattern                     # Find files matching pattern
fd -t f pattern               # Files only
fd -t d pattern               # Directories only
fd -e txt                     # Files with .txt extension
fd pattern ~/Documents        # Search in specific directory
```

**Build Types:**
- `--minimal`: Basic functionality, fast build
- `--release`: Optimized performance (default)

#### ripgrep - Fast Text Search

High-performance text search with regex support.

```bash
# Install with different options
scripts/install-ripgrep.sh                 # Standard with PCRE2
scripts/install-ripgrep.sh --no-pcre2      # Minimal build
scripts/install-ripgrep.sh --optimized     # CPU-optimized

# Usage examples
rg "pattern"                   # Search in current directory
rg "pattern" --type rust       # Search only Rust files
rg -i "pattern"               # Case-insensitive search
rg -A 3 -B 3 "pattern"        # Context lines
rg "pattern" path/            # Search in specific path
```

**Build Types:**
- `--no-pcre2`: Minimal build without PCRE2 regex support
- `--release`: Standard build with PCRE2 (default)
- `--optimized`: CPU-optimized for maximum performance

#### fzf - Fuzzy Finder

Interactive fuzzy finder for files, commands, and more.

```bash
# Install with shell integration
scripts/install-fzf.sh              # Standard with shell setup
scripts/install-fzf.sh --no-shell   # Without shell integration
scripts/install-fzf.sh --profiling  # Profiling build

# Usage examples
fzf                           # Interactive file selection
ls | fzf                      # Filter any list
vim $(fzf)                    # Open file selected with fzf
git log --oneline | fzf       # Interactive git log browser
```

**Build Types:**
- `--standard`: Standard build (default)
- `--profiling`: Profiling-enabled build

**Automatic Key Bindings:**
- **Ctrl+T**: File selection
- **Ctrl+R**: Command history search
- **Alt+C**: Directory navigation

#### jq - JSON Processor

Command-line JSON processor with powerful query capabilities.

```bash
# Install with different optimizations
scripts/install-jq.sh              # Standard
scripts/install-jq.sh --minimal    # Minimal features
scripts/install-jq.sh --optimized  # Performance optimized

# Usage examples
echo '{"name": "test"}' | jq .              # Pretty-print JSON
jq '.name' file.json                        # Extract field
jq '.[] | select(.status == "active")'      # Filter arrays
jq '.items | length'                        # Array length
```

**Build Types:**
- `--minimal`: Basic functionality
- `--standard`: Standard build (default)
- `--optimized`: Performance optimized

### Navigation & File Management

#### zoxide - Smart cd Command

Smarter directory navigation with frecency (frequency + recency).

```bash
# Install and usage
scripts/install-zoxide.sh
z Documents                   # Jump to Documents directory
z proj                        # Jump to most recent project directory
zi                           # Interactive selection with fzf
```

#### yazi - Terminal File Manager

Fast terminal file manager with vim-like keybindings and preview.

```bash
# Install and usage
scripts/install-yazi.sh
yazi                         # Open file manager
yazi /path/to/directory      # Open specific directory
```

#### bat - Enhanced cat

Cat clone with syntax highlighting, Git integration, and themes.

```bash
# Install and usage
scripts/install-bat.sh
bat file.py                  # View file with syntax highlighting
bat -n file.txt              # Show line numbers
bat --theme=GitHub file.md   # Use specific theme
```

#### eza - Modern ls

Enhanced file listings with Git integration and tree view.

```bash
# Install and usage
scripts/install-eza.sh
eza                          # Enhanced ls
eza -la                      # Long format with hidden files
eza --tree                   # Tree view
eza --git                    # Show Git status
```

### Fonts & Terminal Enhancement

#### nerd-fonts - Developer Fonts with Icons

Patched fonts with programming ligatures, file icons, and Git symbols that enhance terminal and editor experience.

```bash
# Quick start - install essential fonts
gearbox install nerd-fonts --minimal

# Standard collection (recommended)
gearbox install nerd-fonts

# Complete collection with all fonts
gearbox install nerd-fonts --maximum

# Install specific fonts
gearbox install nerd-fonts --fonts="FiraCode"
gearbox install nerd-fonts --fonts="FiraCode,JetBrainsMono"

# Interactive selection with previews
gearbox install nerd-fonts --interactive

# Install and auto-configure applications
gearbox install nerd-fonts --configure-apps
```

**Font Collections:**
- **Minimal** (3 fonts, ~30MB): FiraCode, JetBrains Mono, Hack
- **Standard** (12 fonts, ~120MB): Adds Source Code Pro, Inconsolata, Cascadia Code, Ubuntu Mono, DejaVu Sans Mono, Roboto Mono, Space Mono, Iosevka, Geist Mono
- **Maximum** (64 fonts, ~640MB): Complete collection including all available Nerd Fonts - modern fonts like Monaspace, CommitMono, MartianMono; classic fonts like IBM Plex Mono, Victor Mono; specialized fonts like OpenDyslexic, BigBlueTerminal

**Key Features:**
- **Programming Ligatures**: Enhanced display of `=>`, `!=`, `>=`, `<=`, `&&`, `||`, `->`, `<-`
- **File Icons**: Folder, code, image, archive icons in terminal file listings  
- **Git Symbols**: Branch, commit, merge indicators for Git tools
- **Terminal Enhancement**: Perfect with starship prompt, eza file listings, bat syntax highlighting

**Automatic Configuration** (with `--configure-apps`):
- **VS Code**: Sets FiraCode as editor font with ligatures enabled
- **Kitty Terminal**: Configures JetBrains Mono Nerd Font
- **Alacritty Terminal**: Sets up font family configuration
- **Cross-tool Integration**: Optimizes display for starship, eza, bat, lazygit

**Preview and Selection:**
```bash
# Preview fonts before installing
gearbox install nerd-fonts --preview --fonts="FiraCode"

# Interactive selection with live previews
gearbox install nerd-fonts --interactive
# Navigation: ↑/↓ arrows, SPACE to select, 'p' to preview, ENTER to confirm
```

**Font Management:**
```bash
# View installed fonts by family
gearbox status nerd-fonts

# Basic health check with recommendations
gearbox doctor nerd-fonts

# Detailed diagnostics with verbose output
gearbox doctor nerd-fonts --verbose

# View all font variants (system command)
fc-list | grep -i nerd
```

**Status Output Example:**
```
✅ Found 375 Nerd Fonts from 13 font families
📁 Location: ~/.local/share/fonts (1.3 GB)

📋 Font Families:
 1. CaskaydiaCove (36 variants)
 2. FiraCode (18 variants)  
 3. JetBrainsMono (48 variants)
 4. Iosevka (81 variants)
 ...
```

**Font Examples in Action:**
- **Code Editors**: Enhanced ligatures improve code readability
- **Terminal Prompts**: Starship displays beautiful icons and Git status
- **File Managers**: Yazi and eza show intuitive file type icons
- **Git Tools**: Lazygit displays clear branch and status symbols

### Development Tools

#### uv - Python Package Manager

Extremely fast Python package and project manager (10-100x faster than pip).

```bash
# Install and usage
scripts/install-uv.sh
uv pip install package       # Fast package installation
uv venv                      # Create virtual environment
uv run script.py             # Run Python with automatic dependency management
```

#### ruff - Python Linter & Formatter

10-100x faster than Flake8/Black with 800+ lint rules.

```bash
# Install and usage
scripts/install-ruff.sh
ruff check .                 # Lint current directory
ruff format .                # Format code
ruff check --fix .           # Auto-fix issues
```

#### starship - Shell Prompt

Fast, minimal prompt with contextual information.

```bash
# Install and usage
scripts/install-starship.sh
# Automatically configured for bash/zsh/fish
```

#### lazygit - Terminal UI for Git

Interactive Git operations with visual interface.

```bash
# Install and usage
scripts/install-lazygit.sh
lazygit                      # Launch interactive Git UI
```

#### gh - GitHub CLI

Repository management, PRs, issues, and workflows.

```bash
# Install and usage
scripts/install-gh.sh
gh repo list                 # List repositories
gh pr create                 # Create pull request
gh issue list                # List issues
```

### System Monitoring

#### bottom - System Monitor

Beautiful terminal-based system resource monitoring.

```bash
# Install and usage
scripts/install-bottom.sh
btm                          # Launch system monitor
btm --basic                  # Basic mode
```

#### procs - Modern ps

Enhanced process information with tree view and colors.

```bash
# Install and usage
scripts/install-procs.sh
procs                        # Enhanced process list
procs firefox                # Filter by process name
procs --tree                 # Tree view
```

#### bandwhich - Network Monitor

Terminal bandwidth utilization by process.

```bash
# Install and usage
scripts/install-bandwhich.sh
sudo bandwhich              # Network monitoring (requires sudo)
```

### Media Processing

#### ffmpeg - Video/Audio Processing

Comprehensive media processing suite.

```bash
# Install with different codec support
scripts/install-ffmpeg.sh              # General codecs
scripts/install-ffmpeg.sh --minimal    # Basic codecs only
scripts/install-ffmpeg.sh --maximum    # All available codecs

# Usage examples
ffmpeg -i input.mp4 output.avi              # Convert format
ffmpeg -i input.mp4 -vf scale=720:480 out   # Resize video
ffmpeg -i input.mp4 -an output.mp4          # Remove audio
ffmpeg -i input.mp4 -ss 00:01:00 -t 30 out  # Extract 30s clip
```

**Build Types:**
- `--minimal`: Basic codecs for common formats
- `--general`: Standard codec support (default)
- `--maximum`: All available codecs and features

#### imagemagick - Image Manipulation

Powerful image processing and manipulation toolkit.

```bash
# Install and usage
scripts/install-imagemagick.sh
convert input.jpg -resize 50% output.jpg    # Resize image
identify image.jpg                          # Get image info
montage *.jpg grid.jpg                      # Create image grid
```

#### 7zip - Compression Tool

High-ratio compression tool.

```bash
# Install with different optimizations
scripts/install-7zip.sh                  # Optimized
scripts/install-7zip.sh --basic          # Basic build
scripts/install-7zip.sh --asm-optimized  # Assembly optimized

# Usage examples
7zz a archive.7z files/              # Create archive
7zz x archive.7z                     # Extract archive
7zz l archive.7z                     # List contents
7zz a -mx9 archive.7z files/         # Maximum compression
```

**Build Types:**
- `--basic`: Basic functionality
- `--optimized`: Performance optimized (default)
- `--asm-optimized`: Assembly optimizations for best performance

### All Available Tools

For a complete list of all 31 tools with descriptions:

```bash
gearbox list
```

**Core Development:** fd, ripgrep, fzf, jq  
**Navigation:** zoxide, yazi, fclones, bat, eza, dust  
**Fonts & Enhancement:** nerd-fonts  
**Development:** serena, uv, ruff, starship, delta, lazygit, gh, difftastic  
**System Monitoring:** bottom, procs, bandwhich  
**Text Processing:** sd, xsv, choose, tealdeer  
**Analysis:** tokei, hyperfine  
**Media:** ffmpeg, imagemagick, 7zip

## Common Workflows

### Development Environment Setup

Essential tools for development:

```bash
# Core development tools
gearbox install fd ripgrep fzf jq

# With enhanced terminal experience
gearbox install fd ripgrep fzf jq nerd-fonts starship

# With testing validation
gearbox install --run-tests fd ripgrep fzf jq

# Minimal for CI/containers
gearbox install --minimal fd ripgrep
```

### Checking Installation Status

Verify what's installed and monitor health:

```bash
# Check all tool installation status
gearbox status

# Check specific tools
gearbox status fd ripgrep fzf

# View installed Nerd Fonts by family
gearbox status nerd-fonts

# Health check with diagnostics
gearbox doctor nerd-fonts

# Detailed font analysis
gearbox doctor nerd-fonts --verbose
```

### Media Processing Setup

Tools for media work:

```bash
# Media tools with full features
gearbox install --maximum ffmpeg 7zip

# Just ffmpeg with specific features
scripts/install-ffmpeg.sh --maximum --run-tests
```

### Optimized Performance Setup

Maximum performance configuration:

```bash
# Install optimized versions of all tools
scripts/install-fd.sh --release
scripts/install-ripgrep.sh --optimized
scripts/install-ffmpeg.sh --maximum
scripts/install-7zip.sh --asm-optimized
```

### Quick Setup for New Machines

Fast setup with essential tools:

```bash
# One command for essential tools
gearbox install --minimal fd ripgrep fzf jq
```

## Shell Integration

### fzf Integration (Automatic)

After installing fzf, these key bindings are automatically set up:

- **Ctrl+T**: File selection in current directory
- **Ctrl+R**: Command history search
- **Alt+C**: Directory navigation

### Manual Shell Setup

If automatic setup doesn't work:

```bash
# Bash
echo 'eval "$(fzf --bash)"' >> ~/.bashrc

# Zsh
echo 'source <(fzf --zsh)' >> ~/.zshrc

# Fish
echo 'fzf --fish | source' >> ~/.config/fish/config.fish
```

### Useful Aliases

Add these to your shell configuration:

```bash
# Enhanced file operations
alias ff='fd --type f | fzf --preview "head -n 20 {}"'
alias fcd='cd $(fd --type d | fzf)'

# Enhanced search
alias rgi='rg -i'
alias rgf='rg --files | fzf'

# JSON processing
alias pretty='jq .'
alias json-keys='jq "keys"'
```

## Troubleshooting

### Command Not Found After Installation

```bash
# Refresh shell environment
hash -r && source ~/.bashrc

# Or open a new terminal session
```

### Check Installation Paths

```bash
# Verify PATH includes installation directory
echo $PATH | tr ':' '\n' | grep -E "(local|cargo)"

# Check specific tool locations
which fd rg fzf jq ffmpeg 7zz
type fd
```

### Wrong Version Showing

```bash
# Check which binary is being used
type fd
which ripgrep

# Clear command hash and try again
hash -r
fd --version
```

### Rust Version Conflicts

```bash
# Update to latest Rust (handles all tools)
rustup update

# Check Rust version
rustc --version
```

### Build Failures

```bash
# Check dependencies
./scripts/install-common-deps.sh

# Force reinstall specific tool
scripts/install-fd.sh --force

# Check system packages
sudo apt update && sudo apt install build-essential
```

### Permission Issues

```bash
# Ensure not running as root
whoami  # Should not be 'root'

# Check directory permissions
ls -la ~/tools/
```

## Reference

### Installation Paths

| Component | Location | Description |
|-----------|----------|-------------|
| Source builds | `~/tools/build/` | Temporary build files |
| Binaries | `/usr/local/bin/` | Installed executables |
| Cache | `~/tools/cache/` | Downloaded files |
| Scripts | Current directory | This repository |

### Tool Binary Names

| Tool | Binary | Additional |
|------|--------|------------|
| fd | `fd` | Also available as `fdfind` |
| ripgrep | `rg` | - |
| fzf | `fzf` | - |
| jq | `jq` | - |
| zoxide | `z`, `zi` | Also `zoxide` |
| yazi | `yazi` | - |
| fclones | `fclones` | - |
| serena | `serena` | - |
| uv | `uv` | - |
| ruff | `ruff` | - |
| bat | `bat` | - |
| starship | `starship` | - |
| eza | `eza` | - |
| delta | `delta` | - |
| lazygit | `lazygit` | - |
| bottom | `btm` | - |
| procs | `procs` | - |
| tokei | `tokei` | - |
| difftastic | `difft` | - |
| bandwhich | `bandwhich` | Requires sudo |
| xsv | `xsv` | - |
| hyperfine | `hyperfine` | - |
| gh | `gh` | - |
| dust | `dust` | - |
| sd | `sd` | - |
| tealdeer | `tldr` | - |
| choose | `choose` | - |
| ffmpeg | `ffmpeg` | Also `ffprobe`, `ffplay` |
| imagemagick | `convert`, `identify` | Many utilities |
| 7zip | `7zz` | - |

### Build Order (Automatic)

Tools are installed in this order for optimal dependency sharing:

**Go Tools (installs Go toolchain):**
1. **fzf**, **lazygit**, **gh**

**Rust Tools (installs Rust 1.88.0+, reuses toolchain):**
2. **ripgrep**, **fd**, **zoxide**, **yazi**, **fclones**, **bat**, **starship**, **eza**, **delta**, **bottom**, **procs**, **tokei**, **difftastic**, **bandwhich**, **xsv**, **hyperfine**, **dust**, **sd**, **tealdeer**, **choose**

**Python Tools (installs Python 3.11+ and uv):**
3. **serena**, **uv**, **ruff**

**C/C++ Tools (independent builds):**
4. **jq**, **ffmpeg**, **imagemagick**, **7zip**

### Safety Features

- ✅ Non-root execution enforcement
- ✅ Existing installation detection with `--force` override
- ✅ Idempotent operations (safe to run multiple times)
- ✅ Rust version conflict resolution
- ✅ Automatic library path management (`ldconfig`)
- ✅ Command hash clearing (`hash -r`)

### Next Steps

After installation:

1. **Test the tools**: Try the basic usage examples above
2. **Customize your setup**: Add aliases and shell functions
3. **Explore advanced features**: Read tool-specific documentation
4. **Share your experience**: Consider contributing improvements

For technical details about the project architecture, see the [Developer Guide](DEVELOPER_GUIDE.md).