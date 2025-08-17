#!/bin/bash

# eza Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-eza.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
EZA_DIR="eza"
EZA_REPO="https://github.com/eza-community/eza.git"
RUST_MIN_VERSION="1.82.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
DISABLE_GIT=false

# Show help
show_help() {
    cat << EOF
eza Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --disable-git        Build without Git integration
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: release build with install
  $0 -d -c             # Debug build, config only
  $0 -o --run-tests    # Optimized build with tests
  $0 --disable-git     # Build without Git integration
  $0 --skip-deps       # Skip dependency installation

About eza:
  A modern, enhanced replacement for the traditional 'ls' command.
  Features colorful output, Git integration, tree view, extended attributes
  support, and intelligent file type recognition. Perfect for developers
  who want more informative directory listings.

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
        --disable-git)
            DISABLE_GIT=true
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
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Version comparison function
version_compare() {
    if [[ $1 == $2 ]]; then
        return 0
    fi
    local IFS=.
    local i ver1=($1) ver2=($2)
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then
            ver2[i]=0
        fi
    done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 0
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 1
        fi
    done
    return 0
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "===================================="
log "eza Installation Script"
log "===================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
log "Git integration: $([ $DISABLE_GIT = true ] && echo 'disabled' || echo 'enabled')"
echo

# Check if eza is already installed
if command -v eza &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(eza --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    log "Found existing eza installation (version: $CURRENT_VERSION)"
    log "Building latest version from source (gearbox builds from latest main branch)"
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for eza..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl
    
    # Check Rust installation
    if ! command -v rustc &> /dev/null; then
        log "Installing Rust..."
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source ~/.cargo/env
    else
        RUST_VERSION=$(rustc --version | grep -oP '\d+\.\d+\.\d+')
        log "Found Rust version: $RUST_VERSION"
        
        if ! version_compare "$RUST_VERSION" "$RUST_MIN_VERSION"; then
            log "Updating Rust to meet minimum version requirement ($RUST_MIN_VERSION)..."
            rustup update
        else
            log "Rust version is sufficient (>= $RUST_MIN_VERSION)"
        fi
    fi
    
    # Ensure we have the latest stable Rust
    log "Ensuring Rust toolchain is current..."
    rustup update stable
    rustup default stable
    
    success "Dependencies installation completed!"
else
    log "Skipping dependency installation as requested"
fi

# Ensure we have access to Rust tools
if ! command -v rustc &> /dev/null; then
    if [[ -f ~/.cargo/env ]]; then
        source ~/.cargo/env
    else
        error "Rust is not available and cargo environment not found"
    fi
fi

# Create build directory
BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Get eza source directory
EZA_SOURCE_DIR="$BUILD_DIR/$EZA_DIR"

# Clone or update repository
if [[ ! -d "$EZA_SOURCE_DIR" ]]; then
    log "Cloning eza repository..."
    git clone "$EZA_REPO" "$EZA_SOURCE_DIR"
else
    log "Updating eza repository..."
    cd "$EZA_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$EZA_SOURCE_DIR"

# Configure build
log "Configuring eza build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

log "eza source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building eza..."

# Set build features
BUILD_FEATURES=""
if [[ "$DISABLE_GIT" != true ]]; then
    BUILD_FEATURES="--features git"
fi

case $BUILD_TYPE in
    debug)
        log "Building debug version..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build $BUILD_FEATURES
        else
            cargo build --no-default-features
        fi
        TARGET_DIR="target/debug"
        ;;
    release)
        log "Building release version..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build --release $BUILD_FEATURES
        else
            cargo build --release --no-default-features
        fi
        TARGET_DIR="target/release"
        ;;
    optimized)
        log "Building optimized version with target-cpu=native..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            RUSTFLAGS="-C target-cpu=native" cargo build --release $BUILD_FEATURES
        else
            RUSTFLAGS="-C target-cpu=native" cargo build --release --no-default-features
        fi
        TARGET_DIR="target/release"
        ;;
esac

# Verify build
if [[ ! -f "$TARGET_DIR/eza" ]]; then
    error "Build failed - eza binary not found in $TARGET_DIR"
fi

success "eza build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $EZA_SOURCE_DIR/$TARGET_DIR/eza"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running eza tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing eza..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/eza" /usr/local/bin/
sudo chmod +x /usr/local/bin/eza

# Verify installation
if command -v eza &> /dev/null; then
    INSTALLED_VERSION=$(eza --version 2>/dev/null | head -n1)
    success "eza installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/eza"
    
    # Show basic usage
    echo
    log "Basic usage examples:"
    log "  eza                              # List files in current directory"
    log "  eza -l                           # Long format listing"
    log "  eza -la                          # Long format with hidden files"
    log "  eza --tree                       # Tree view of directory structure"
    log "  eza --tree --level=2             # Tree view with depth limit"
    log "  eza --git                        # Show git status for files"
    echo
    log "Advanced features:"
    log "  eza -lh --sort=size              # Sort by file size"
    log "  eza -lh --sort=modified          # Sort by modification time"
    log "  eza --long --header              # Show column headers"
    log "  eza --long --no-time             # Hide time information"
    log "  eza --grid --across              # Grid layout"
    echo
    log "Integration tips:"
    log "  alias ls='eza'                   # Replace ls with eza"
    log "  alias ll='eza -l'                # Long listing"
    log "  alias la='eza -la'               # All files, long format"
    log "  alias tree='eza --tree'          # Tree view"
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        debug)
            log "Features: Debug build with symbols for development"
            ;;
        release)
            log "Features: Optimized release build for daily use"
            ;;
        optimized)
            log "Features: Maximum performance build with target-cpu optimizations"
            ;;
    esac
    echo
    log "Key features enabled:"
    log "  ✓ Colorful file listings with intelligent type detection"
    if [[ "$DISABLE_GIT" != true ]]; then
        log "  ✓ Git integration (shows repository status)"
    else
        log "  ✗ Git integration disabled"
    fi
    log "  ✓ Tree view for directory structure"
    log "  ✓ Extended attributes and metadata display"
    log "  ✓ Human-readable file sizes and dates"
    log "  ✓ Symlink and mount point detection"
    echo
    log "Configuration: Create ~/.config/eza/theme.yml for custom themes"
    log "Environment: Set EZA_COLORS for custom color schemes"
    log "For more information: eza --help"
else
    error "eza installation verification failed - eza not found in PATH"
fi

# Update library cache
sudo ldconfig

success "eza installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"