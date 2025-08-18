#!/bin/bash

# bat Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-bat.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/lib/common.sh" ]]; then
    source "$REPO_DIR/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/lib/" >&2
    exit 1
fi


# Configuration
BAT_DIR="bat"
BAT_REPO="https://github.com/sharkdp/bat.git"
RUST_MIN_VERSION="1.74.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
bat Installation Script for Debian Linux

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
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: release build with install
  $0 -d -c             # Debug build, config only
  $0 -o --run-tests    # Optimized build with tests
  $0 --skip-deps       # Skip dependency installation

About bat:
  A cat(1) clone with wings - enhanced file viewer with syntax highlighting,
  Git integration, line numbers, and automatic paging. Perfect replacement
  for cat with intelligent features for developers and sysadmins.

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
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Note: Logging functions now provided by lib/common.sh

# Note: Logging functions now provided by lib/common.sh

# Note: Logging functions now provided by lib/common.sh

# Note: Logging functions now provided by lib/common.sh

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
log "==================================="
log "bat Installation Script"
log "==================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
echo

# Check if bat is already installed
if command -v bat &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(bat --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    log "Found existing bat installation (version: $CURRENT_VERSION)"
    log "Building latest version from source (gearbox builds from latest main branch)"
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for bat..."
    
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

# Get bat source directory
BAT_SOURCE_DIR="$BUILD_DIR/$BAT_DIR"

# Clone or update repository
if [[ ! -d "$BAT_SOURCE_DIR" ]]; then
    log "Cloning bat repository..."
    git clone "$BAT_REPO" "$BAT_SOURCE_DIR"
else
    log "Updating bat repository..."
    cd "$BAT_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/master
fi

cd "$BAT_SOURCE_DIR"

# Configure build
log "Configuring bat build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

log "bat source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building bat..."

case $BUILD_TYPE in
    debug)
        log "Building debug version..."
        cargo build --bins
        TARGET_DIR="target/debug"
        ;;
    release)
        log "Building release version..."
        cargo install --path . --locked --root . --quiet
        TARGET_DIR="bin"
        ;;
    optimized)
        log "Building optimized version with target-cpu=native..."
        RUSTFLAGS="-C target-cpu=native" cargo install --path . --locked --root . --quiet
        TARGET_DIR="bin"
        ;;
esac

# Verify build
if [[ ! -f "$TARGET_DIR/bat" ]]; then
    error "Build failed - bat binary not found in $TARGET_DIR"
fi

success "bat build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $BAT_SOURCE_DIR/$TARGET_DIR/bat"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running bat tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing bat..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/bat" /usr/local/bin/
sudo chmod +x /usr/local/bin/bat

# Verify installation
if command -v bat &> /dev/null; then
    INSTALLED_VERSION=$(bat --version 2>/dev/null | head -n1)
    success "bat installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/bat"
    
    # Show basic usage
    echo
    log "Basic usage examples:"
    log "  bat filename.txt                 # View file with syntax highlighting"
    log "  bat src/*.rs                     # View multiple files"
    log "  bat -n filename.py               # Show line numbers"
    log "  bat -A filename.js               # Show non-printable characters"
    log "  bat --style=numbers,grid file.md # Custom styling"
    echo
    log "Advanced features:"
    log "  bat -d                           # Show git diff"
    log "  bat --theme=Dracula file.py      # Use specific theme"
    log "  bat --list-themes                # List available themes"
    log "  bat --generate-config-file       # Create config file"
    echo
    log "Integration with other tools:"
    log "  alias cat='bat --paging=never'   # Replace cat"
    log "  export PAGER='bat'               # Use as pager"
    log "  git log --oneline | bat          # Syntax highlight git output"
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
    log "  ✓ Syntax highlighting for 200+ languages"
    log "  ✓ Git integration (shows modifications)"
    log "  ✓ Automatic paging for long files"
    log "  ✓ Line numbering and grid display"
    log "  ✓ Theme support and customization"
    echo
    log "Configuration: Run 'bat --generate-config-file' to create ~/.config/bat/config"
    log "For more information: bat --help"
else
    error "bat installation verification failed - bat not found in PATH"
fi

# Update library cache
sudo ldconfig

success "bat installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"
