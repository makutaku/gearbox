#!/bin/bash

# procs Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-procs.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROCS_DIR="procs"
PROCS_REPO="https://github.com/dalance/procs.git"
RUST_MIN_VERSION="1.74.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="binary"   # source, binary
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
ENABLE_DOCKER=true
SETUP_COMPLETIONS=true

# Show help
show_help() {
    cat << EOF
procs Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types (source builds only):
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with musl static linking)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Installation Methods:
  --source             Build from source (requires Rust toolchain)
  --binary             Download pre-built binary (default, faster)

Feature Options:
  --no-docker          Disable Docker container support
  --no-completions     Skip shell completion setup

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: binary installation with all features
  $0 --source          # Build from source
  $0 -d -c --source    # Debug build, config only (source)
  $0 -o --source       # Optimized musl build (source)
  $0 --no-docker       # Build without Docker support
  $0 --skip-deps       # Skip dependency installation

About procs:
  A modern replacement for the ps command written in Rust. Provides
  colored, human-readable process information with additional details
  like TCP/UDP ports, read/write throughput, Docker container names,
  tree view, and enhanced memory information. Perfect for system
  administrators and developers who want better process visibility.

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
        --source)
            INSTALL_METHOD="source"
            shift
            ;;
        --binary)
            INSTALL_METHOD="binary"
            shift
            ;;
        --no-docker)
            ENABLE_DOCKER=false
            shift
            ;;
        --no-completions)
            SETUP_COMPLETIONS=false
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

# Setup shell completions
setup_shell_completions() {
    if [[ "$SETUP_COMPLETIONS" != true ]]; then
        log "Skipping shell completion setup as requested"
        return 0
    fi
    
    log "Setting up shell completions..."
    
    # Create completion directories
    local bash_comp_dir="$HOME/.local/share/bash-completion/completions"
    local zsh_comp_dir="$HOME/.local/share/zsh/site-functions"
    local fish_comp_dir="$HOME/.config/fish/completions"
    
    mkdir -p "$bash_comp_dir" "$zsh_comp_dir" "$fish_comp_dir"
    
    # Generate completions using procs
    if command -v procs &> /dev/null; then
        log "Generating shell completions..."
        
        # Bash completion
        procs --completion bash > "$bash_comp_dir/procs" 2>/dev/null || true
        
        # Zsh completion
        procs --completion zsh > "$zsh_comp_dir/_procs" 2>/dev/null || true
        
        # Fish completion
        procs --completion fish > "$fish_comp_dir/procs.fish" 2>/dev/null || true
        
        success "Shell completions installed!"
        log "Completions installed for bash, zsh, and fish"
    else
        warning "Could not generate completions - procs not found in PATH"
    fi
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "======================================"
log "procs Installation Script"
log "======================================"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
log "Docker support: $ENABLE_DOCKER"
log "Shell completions: $SETUP_COMPLETIONS"
echo

# Check if procs is already installed
if command -v procs &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(procs --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    
    # For binary installation method, respect existing installation
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "procs is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"
        exit 0
    else
        # For source builds, inform but continue (gearbox philosophy: build latest from source)
        log "Found existing procs installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    
    if [[ "$MODE" == "config" ]]; then
        log "Config mode: would download and install procs binary"
        success "Configuration completed (binary installation ready)!"
        exit 0
    fi
    
    # Detect architecture
    ARCH=$(uname -m)
    OS="unknown-linux-musl"  # Use musl for better compatibility
    case $ARCH in
        x86_64) ARCH_TAG="x86_64" ;;
        aarch64|arm64) ARCH_TAG="aarch64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest procs binary for ${ARCH_TAG}-${OS}..."
    
    # Get latest release info
    RELEASE_URL="https://api.github.com/repos/dalance/procs/releases/latest"
    RELEASE_INFO=$(curl -s "$RELEASE_URL")
    VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name"' | cut -d '"' -f 4)
    
    if [[ -z "$VERSION" ]]; then
        error "Could not determine latest version from GitHub API"
    fi
    
    # Construct download URL
    DOWNLOAD_URL="https://github.com/dalance/procs/releases/download/${VERSION}/procs-${VERSION}-${ARCH_TAG}-${OS}.zip"
    
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    # Download and extract
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    if ! curl -fLO "$DOWNLOAD_URL"; then
        error "Failed to download procs binary"
    fi
    
    # Extract (assuming .zip format)
    ARCHIVE_NAME=$(basename "$DOWNLOAD_URL")
    unzip -q "$ARCHIVE_NAME"
    
    # Find procs binary
    if [[ ! -f "procs" ]]; then
        error "procs binary not found in downloaded archive"
    fi
    
    # Install binary
    sudo cp "procs" /usr/local/bin/procs
    sudo chmod +x /usr/local/bin/procs
    
    # Clean up
    cd /
    rm -rf "$TEMP_DIR"
    
    # Verify installation
    if command -v procs &> /dev/null; then
        INSTALLED_VERSION=$(procs --version 2>/dev/null | head -n1)
        success "procs installation completed successfully!"
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: Binary download"
        log "Binary location: /usr/local/bin/procs"
        
        # Setup shell completions
        echo
        setup_shell_completions
        
        echo
        log "Basic usage:"
        log "  procs                        # Show all processes"
        log "  procs --tree                 # Show process tree"
        log "  procs firefox                # Search for processes"
        log "  procs --help                 # Show all options"
        log "  For more information: https://github.com/dalance/procs"
    else
        error "procs installation verification failed"
    fi
    
    exit 0
fi

# Source build method continues below
log "Using source build method..."

# Install dependencies for source build
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for procs..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl pkg-config libssl-dev
    
    # For musl builds, install musl development tools
    if [[ "$BUILD_TYPE" == "optimized" ]]; then
        log "Installing musl development tools for static linking..."
        sudo apt install -y musl-tools musl-dev
    fi
    
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
    
    # Add musl target for optimized builds
    if [[ "$BUILD_TYPE" == "optimized" ]]; then
        log "Adding musl target for static linking..."
        rustup target add x86_64-unknown-linux-musl
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

# Get procs source directory
PROCS_SOURCE_DIR="$BUILD_DIR/$PROCS_DIR"

# Clone or update repository
if [[ ! -d "$PROCS_SOURCE_DIR" ]]; then
    log "Cloning procs repository..."
    git clone "$PROCS_REPO" "$PROCS_SOURCE_DIR"
else
    log "Updating procs repository..."
    cd "$PROCS_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/master
fi

cd "$PROCS_SOURCE_DIR"

# Configure build
log "Configuring procs build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

# Prepare build features
BUILD_FEATURES=""
if [[ "$ENABLE_DOCKER" != true ]]; then
    BUILD_FEATURES="--no-default-features"
fi

log "procs source configured successfully"
log "Build features: ${BUILD_FEATURES:-default (docker support enabled)}"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building procs..."

case $BUILD_TYPE in
    debug)
        log "Building debug version..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build $BUILD_FEATURES
        else
            cargo build
        fi
        TARGET_DIR="target/debug"
        ;;
    release)
        log "Building release version..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build --release $BUILD_FEATURES
        else
            cargo build --release
        fi
        TARGET_DIR="target/release"
        ;;
    optimized)
        log "Building optimized version with musl static linking..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build --release --target x86_64-unknown-linux-musl $BUILD_FEATURES
        else
            cargo build --release --target x86_64-unknown-linux-musl
        fi
        TARGET_DIR="target/x86_64-unknown-linux-musl/release"
        ;;
esac

# Verify build
if [[ ! -f "$TARGET_DIR/procs" ]]; then
    error "Build failed - procs binary not found in $TARGET_DIR"
fi

success "procs build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $PROCS_SOURCE_DIR/$TARGET_DIR/procs"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running procs tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing procs..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/procs" /usr/local/bin/
sudo chmod +x /usr/local/bin/procs

# Verify installation
if command -v procs &> /dev/null; then
    INSTALLED_VERSION=$(procs --version 2>/dev/null | head -n1)
    success "procs installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/procs"
    
    # Setup shell completions
    echo
    setup_shell_completions
    
    # Show usage information
    echo
    log "Basic usage examples:"
    log "  procs                            # Show all processes"
    log "  procs --tree                     # Show process tree"
    log "  procs --watch                    # Watch mode (like top)"
    log "  procs firefox                    # Search for processes by name"
    log "  procs --and firefox --and tab    # Multiple keyword search"
    log "  procs --or firefox --or chrome   # OR search"
    echo
    log "Advanced options:"
    log "  procs --sortd cpu                # Sort by CPU usage (descending)"
    log "  procs --sorta mem                # Sort by memory usage (ascending)"
    log "  procs --filter 'cpu > 50'       # Filter by CPU usage"
    log "  procs --thread                   # Show threads"
    log "  procs --color=always             # Force colored output"
    echo
    log "Information columns:"
    log "  • PID, PPID, user, state, CPU%, memory, start time"
    log "  • TCP/UDP ports, read/write throughput"
    if [[ "$ENABLE_DOCKER" == true ]]; then
        log "  • Docker container names and IDs"
    fi
    log "  • Command line arguments and environment"
    echo
    log "Integration tips:"
    log "  alias ps='procs'                 # Replace ps with procs"
    log "  procs | less                     # Use with pager"
    log "  watch procs                      # Monitor continuously"
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
            log "Features: Static musl build for maximum portability"
            ;;
    esac
    echo
    log "Key features enabled:"
    log "  ✓ Colored, human-readable process information"
    log "  ✓ Tree view and search capabilities"
    log "  ✓ TCP/UDP port information"
    log "  ✓ Read/write throughput monitoring"
    if [[ "$ENABLE_DOCKER" == true ]]; then
        log "  ✓ Docker container support"
    else
        log "  ✗ Docker container support disabled"
    fi
    log "  ✓ Multi-column keyword search"
    log "  ✓ Watch mode for real-time monitoring"
    echo
    log "Configuration: Create ~/.config/procs/config.toml for custom settings"
    log "Completions: Available for bash, zsh, and fish shells"
    log "For more information: procs --help"
else
    error "procs installation verification failed - procs not found in PATH"
fi

# Update library cache
sudo ldconfig

success "procs installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"