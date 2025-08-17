# Developer Guide

Complete technical guide for understanding the architecture and contributing to the Essential Tools Installer.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Development Setup](#development-setup)
3. [Adding New Tools](#adding-new-tools)
4. [Technical Standards](#technical-standards)
5. [Testing and Validation](#testing-and-validation)
6. [Advanced Topics](#advanced-topics)

## Architecture Overview

### Three-Tier System

The installer follows a clean three-tier architecture:

#### 1. Configuration Layer (`config.sh`)
- Defines build directories, cache paths, install prefixes
- Provides shared logging functions and color definitions
- Contains utility functions for path and version management
- Single source of configuration truth

#### 2. Orchestration Layer (`scripts/install-all-tools.sh`)
- Manages installation order for optimal dependency sharing
- Handles build type flags and common options
- Coordinates common dependency installation
- Provides unified interface for multiple tools

#### 3. Individual Tool Scripts (`scripts/install-*.sh`)
- Each tool has a dedicated script following consistent patterns
- Supports standard command-line interface
- Handles tool-specific build configurations
- Maintains independence while leveraging shared infrastructure

### Directory Strategy

```
~/tools/build/     # Source repositories (temporary)
~/tools/cache/     # Downloads and cache files
/usr/local/bin/    # Installed binaries (system-wide)
~/gearbox/         # This repository (scripts)
```

This separation keeps the scripts repository clean while organizing build artifacts logically.

### Dependency Management Philosophy

**Shared Toolchains:** Install programming language toolchains once, use across multiple tools:
- Go toolchain → fzf
- Rust toolchain → ripgrep, fd
- C/C++ toolchain → jq, ffmpeg, 7zip

**Optimal Installation Order:** 
1. Go tools (fzf) - installs Go
2. Rust tools (ripgrep, fd) - installs Rust, reuses toolchain
3. C/C++ tools (jq, ffmpeg, 7zip) - independent builds

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/yourusername/gearbox.git
cd gearbox
```

### 2. Understand the Codebase

```bash
# Study the shared configuration
cat config.sh

# Examine the main orchestrator
cat scripts/install-all-tools.sh

# Look at individual tool patterns
cat scripts/install-fd.sh
cat scripts/install-ripgrep.sh
```

### 3. Set Up Development Environment

```bash
# Test that existing scripts work
./tests/test-runner.sh

# Try a minimal installation to understand the flow
gearbox --minimal fd

# Examine the build directory structure
ls -la ~/tools/build/
```

## Adding New Tools

### Step 1: Create Installation Script

Create `scripts/install-newtool.sh` following this complete template:

```bash
#!/bin/bash

# Tool Name Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-newtool.sh [OPTIONS]

set -e  # Exit on any error

# Source shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
source "$REPO_DIR/config.sh"

# Tool-specific configuration
TOOL_NAME="newtool"
TOOL_DIR="newtool"
TOOL_REPO="https://github.com/author/newtool.git"
MIN_VERSION="1.0.0"

# Default options
BUILD_TYPE="release"
MODE="install"
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help function
show_help() {
    cat << EOF
Tool Name Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build with symbols
  -r, --release         Optimized release build (default)
  -o, --optimized       Maximum optimizations

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Standard installation
  $0 --debug           # Debug build
  $0 --optimized       # Maximum performance
  $0 --skip-deps       # Skip dependency installation
  $0 --run-tests       # Include test validation

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--debug)
            BUILD_TYPE="debug"
            shift
            ;;
        -r|--release)
            BUILD_TYPE="release"
            shift
            ;;
        -o|--optimized)
            BUILD_TYPE="optimized"
            shift
            ;;
        -c|--config-only)
            MODE="config"
            shift
            ;;
        -b|--build-only)
            MODE="build"
            shift
            ;;
        -i|--install)
            MODE="install"
            shift
            ;;
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --run-tests)
            RUN_TESTS=true
            shift
            ;;
        --force)
            FORCE_INSTALL=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Check if tool is already installed
if command -v "$TOOL_NAME" &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "$TOOL_NAME already installed. Use --force to reinstall."
    exit 0
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for $TOOL_NAME..."
    # Tool-specific dependency installation logic
fi

# Get source directory
TOOL_SOURCE_DIR="$(get_source_dir "$TOOL_DIR")"

# Clone or update repository
if [[ ! -d "$TOOL_SOURCE_DIR" ]]; then
    log "Cloning $TOOL_NAME repository..."
    git clone "$TOOL_REPO" "$TOOL_SOURCE_DIR"
else
    log "Updating $TOOL_NAME repository..."
    cd "$TOOL_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$TOOL_SOURCE_DIR"

# Configure build based on type
case $BUILD_TYPE in
    debug)
        log "Configuring debug build..."
        # Debug configuration
        ;;
    release)
        log "Configuring release build..."
        # Release configuration
        ;;
    optimized)
        log "Configuring optimized build..."
        # Optimized configuration
        ;;
esac

# Build
if [[ "$MODE" != "config" ]]; then
    log "Building $TOOL_NAME..."
    # Build commands
fi

# Install
if [[ "$MODE" == "install" ]]; then
    log "Installing $TOOL_NAME..."
    # Installation commands
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running tests for $TOOL_NAME..."
    # Test commands
fi

# Verify installation
if [[ "$MODE" == "install" ]]; then
    if command -v "$TOOL_NAME" &> /dev/null; then
        success "$TOOL_NAME installation completed successfully!"
        log "Version: $($TOOL_NAME --version)"
    else
        error "$TOOL_NAME installation failed - binary not found in PATH"
    fi
fi
```

### Step 2: Update Main Installer

Add your tool to `scripts/install-all-tools.sh`:

```bash
# Add to AVAILABLE_TOOLS array
AVAILABLE_TOOLS=("ffmpeg" "7zip" "jq" "fd" "ripgrep" "fzf" "newtool")

# Add build flag mapping in get_build_flag function
get_build_flag() {
    local tool=$1
    case $tool in
        # ... existing cases ...
        newtool)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
    esac
}

# Add to installation order (consider dependencies)
INSTALLATION_ORDER=()
for tool in "fzf" "ripgrep" "fd" "jq" "ffmpeg" "7zip" "newtool"; do
    if [[ " ${SELECTED_TOOLS[*]} " =~ " ${tool} " ]]; then
        INSTALLATION_ORDER+=("$tool")
    fi
done

# Add verification logic
case $tool in
    # ... existing cases ...
    newtool)
        if command -v newtool &> /dev/null; then
            log "✓ newtool: $(newtool --version)"
        else
            FAILED_TOOLS+=("newtool")
        fi
        ;;
esac
```

### Step 3: Update Documentation

1. Add tool to main `README.md` tools table
2. Add comprehensive usage examples to `docs/USER_GUIDE.md`
3. Document any special requirements or limitations

## Technical Standards

### Script Structure

All scripts must follow this structure:

1. **Header**: Shebang, description, usage
2. **Error handling**: `set -e` and error traps
3. **Configuration**: Source `config.sh`, define variables
4. **Functions**: Help, utility functions
5. **Argument parsing**: Standard flag handling
6. **Main logic**: Dependencies → configure → build → install → verify

### Command-Line Interface Standards

All scripts must support these options:

```bash
# Build types
-d, --debug           # Debug build with symbols
-r, --release         # Optimized release build (default)
-o, --optimized       # Maximum optimizations

# Modes
-c, --config-only     # Configure only
-b, --build-only      # Build only
-i, --install         # Full installation (default)

# Common options
--skip-deps          # Skip dependency installation
--run-tests          # Run test suite after building
--force              # Force reinstallation
-h, --help           # Show help message
```

### Error Handling Best Practices

```bash
# Always enable error handling
set -e

# Use shared logging functions
log "Starting installation..."
success "Installation completed!"
error "Build failed: specific error message"
warning "Non-critical issue detected"

# Provide context in error messages
error "Failed to compile $TOOL_NAME: missing dependency $DEP_NAME"

# Clean up on failure
trap 'cleanup_on_error' EXIT

cleanup_on_error() {
    if [[ $? -ne 0 ]]; then
        warning "Installation failed, cleaning up..."
        rm -rf "$TEMP_DIR"
    fi
}
```

### Directory Management

```bash
# Use shared configuration functions
source "$REPO_DIR/config.sh"

# Get paths using utility functions
TOOL_SOURCE_DIR="$(get_source_dir "$TOOL_NAME")"
CACHE_FILE="$(get_cache_path "tool-download.tar.gz")"

# Ensure directories exist
mkdir -p "$BUILD_DIR" "$CACHE_DIR"

# Always return to known directory
cd "$TOOL_SOURCE_DIR" || error "Failed to access $TOOL_SOURCE_DIR"
```

### Version and Dependency Handling

```bash
# Check existing installations
if command -v "$TOOL_NAME" &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "$TOOL_NAME already installed. Use --force to reinstall."
    exit 0
fi

# Version comparison example
check_version() {
    local current_version required_version
    current_version="$(tool --version | grep -oP '\d+\.\d+\.\d+')"
    required_version="$MIN_VERSION"
    
    if version_compare "$current_version" "$required_version"; then
        log "Version $current_version meets requirement $required_version"
    else
        error "Version $current_version is below required $required_version"
    fi
}

# Install dependencies conditionally
if [[ "$SKIP_DEPS" != true ]]; then
    install_dependencies
fi

# Verify installation success
verify_installation() {
    if ! command -v "$TOOL_NAME" &> /dev/null; then
        error "$TOOL_NAME installation failed - binary not found"
    fi
    
    log "Verifying $TOOL_NAME installation..."
    if ! "$TOOL_NAME" --version &> /dev/null; then
        error "$TOOL_NAME binary exists but doesn't work properly"
    fi
    
    success "$TOOL_NAME verified successfully"
}
```

## Testing and Validation

### Basic Testing

```bash
# Run the basic test suite
./tests/test-runner.sh

# Test individual script
scripts/install-newtool.sh --debug --run-tests

# Test via main installer
gearbox --minimal newtool --run-tests
```

### Manual Verification

```bash
# Verify installation success
which newtool
newtool --version

# Test functionality
newtool --help
newtool basic-command

# Check in clean environment
docker run -it debian:bookworm bash
# Install and test in container
```

### Integration Testing

```bash
# Test all build types
scripts/install-newtool.sh --debug
scripts/install-newtool.sh --release  
scripts/install-newtool.sh --optimized

# Test dependency coordination
scripts/install-common-deps.sh
scripts/install-newtool.sh --skip-deps

# Test error conditions
scripts/install-newtool.sh --force
# Test with missing dependencies
# Test with network issues
```

### Test Coverage Areas

1. **Installation scenarios**: All build types, with/without deps
2. **Error handling**: Network failures, missing dependencies, disk space
3. **Idempotency**: Multiple runs should be safe
4. **Clean environments**: Fresh systems, containers
5. **Integration**: Tool works with main installer
6. **Verification**: Binary works correctly after installation

## Advanced Topics

### Performance Optimization

**Build Parallelization:**
```bash
# Use available CPU cores
CORES=$(nproc)
make -j"$CORES"

# Rust builds
cargo build --jobs="$CORES"
```

**Cache Management:**
```bash
# Use cache for repeated builds
CACHE_FILE="$(get_cache_path "source-$VERSION.tar.gz")"
if [[ ! -f "$CACHE_FILE" ]]; then
    download_source "$CACHE_FILE"
fi
```

**Dependency Optimization:**
```bash
# Skip redundant installations
if [[ "$SKIP_DEPS" == true ]] || command -v rustc &> /dev/null; then
    log "Skipping Rust installation"
else
    install_rust
fi
```

### Security Considerations

**Source Verification:**
```bash
# Verify Git repository authenticity
git verify-commit HEAD

# Check checksums for downloads
echo "$EXPECTED_SHA256 $DOWNLOAD_FILE" | sha256sum -c
```

**Safe Installation:**
```bash
# Never run as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Validate paths to prevent directory traversal
validate_path() {
    local path="$1"
    if [[ "$path" =~ \.\./|^/ ]]; then
        error "Invalid path: $path"
    fi
}
```

### Debugging and Troubleshooting

**Debug Mode:**
```bash
# Enable verbose output
if [[ "$DEBUG" == true ]]; then
    set -x
fi

# Detailed logging
debug() {
    if [[ "$DEBUG" == true ]]; then
        echo "[DEBUG] $*" >&2
    fi
}
```

**Error Investigation:**
```bash
# Preserve build artifacts on failure
cleanup_on_error() {
    if [[ $? -ne 0 ]] && [[ "$PRESERVE_ON_ERROR" == true ]]; then
        warning "Build failed, preserving artifacts in $BUILD_DIR"
    else
        rm -rf "$BUILD_DIR"
    fi
}
```

### Contributing Guidelines

1. **Follow existing patterns**: Study current scripts before creating new ones
2. **Test thoroughly**: All build types, clean environments, error conditions
3. **Document comprehensively**: Update all relevant documentation files
4. **Use standard interfaces**: Command-line flags, error handling, logging
5. **Consider performance**: Parallel builds, caching, dependency optimization
6. **Maintain security**: No root execution, path validation, source verification

### Submission Process

```bash
# Create feature branch
git checkout -b add-newtool-support

# Implement changes
# ... create script, update installer, update docs ...

# Test thoroughly
./tests/test-runner.sh
gearbox newtool --run-tests

# Commit with descriptive message
git add .
git commit -m "Add newtool installation script

- Implement scripts/install-newtool.sh with standard options
- Add newtool to main installer with build type support  
- Update USER_GUIDE.md with usage examples
- Include test validation for successful installation"

# Push and create pull request
git push origin add-newtool-support
```

For user-focused documentation, see the [User Guide](USER_GUIDE.md).