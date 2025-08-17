# User Guide

Complete guide for installing and using essential command-line tools.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation Options](#installation-options)
3. [Individual Tools](#individual-tools)
4. [Common Workflows](#common-workflows)
5. [Shell Integration](#shell-integration)
6. [Troubleshooting](#troubleshooting)
7. [Reference](#reference)

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

This will show a confirmation prompt before installing all 6 tools, since the process takes 30-60 minutes. The installer builds tools in optimal order, sharing dependencies efficiently.

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
- Rust ≥ 1.88.0 (for fd, ripgrep)
- Go ≥ 1.23.4 (for fzf)
- C/C++ toolchain (for ffmpeg, 7zip, jq)

**System Packages:**
- `build-essential`, `git`, `curl`, `wget`, `make`
- `autoconf`, `automake`, `libtool`
- `nasm`, `yasm` (for assembly optimizations)
- Codec libraries (for ffmpeg)

## Individual Tools

### fd - Fast File Finder

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

### ripgrep - Fast Text Search

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

### fzf - Fuzzy Finder

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

### jq - JSON Processor

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

### ffmpeg - Video/Audio Processing

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

### 7zip - Compression Tool

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

## Common Workflows

### Development Environment Setup

Essential tools for development:

```bash
# Core development tools
gearbox install fd ripgrep fzf jq

# With testing validation
gearbox install --run-tests fd ripgrep fzf jq

# Minimal for CI/containers
gearbox install --minimal fd ripgrep
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
| ffmpeg | `ffmpeg` | Also `ffprobe`, `ffplay` |
| 7zip | `7zz` | - |

### Build Order (Automatic)

Tools are installed in this order for optimal dependency sharing:

1. **fzf** (installs Go toolchain)
2. **ripgrep** (installs Rust 1.88.0+)
3. **fd** (uses existing Rust)
4. **jq** (independent C build)
5. **ffmpeg** (independent C/C++ build)
6. **7zip** (independent C/C++ build)

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