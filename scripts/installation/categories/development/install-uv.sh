#!/bin/bash

# uv Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-uv.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")")"
REPO_DIR="$(dirname "$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")")"
# Source common library for shared functions
if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/scripts/lib/" >&2
    exit 1
fi


# Configuration
UV_DIR="uv"
UV_REPO="https://github.com/astral-sh/uv.git"
RUST_MIN_VERSION="1.70.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="official" # official (recommended), source
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
uv Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Installation Methods:
  --official           Use official installer (default, faster, prebuilt binary)
  --source             Build from source (follows gearbox patterns, slower)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: official installer (recommended)
  $0 --source          # Build from source (for development/customization)
  $0 --source -d -c    # Source build: debug mode, config only
  $0 --source -o --run-tests  # Source build: optimized with tests
  $0 --skip-deps       # Skip dependency installation

About uv:
  An extremely fast Python package and project manager, written in Rust.
  Provides a unified interface for Python package management, virtual
  environments, and project dependency handling - 10-100x faster than pip.

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
        --official)
            INSTALL_METHOD="official"
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
log "uv Installation Script"
log "==================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
echo

# Check if uv is already installed
if command -v uv &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(uv --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    
    # For official installer method, respect existing installation
    if [[ "$INSTALL_METHOD" == "official" ]]; then
        log "uv is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"
        exit 0
    else
        # For source builds, inform but continue (gearbox philosophy: build latest from source)
        log "Found existing uv installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Official installer method
if [[ "$INSTALL_METHOD" == "official" ]]; then
    log "Using official uv installer..."
    
    if [[ "$MODE" == "config" ]]; then
        log "Config mode: would download and install uv via official installer"
        success "Configuration completed (official installer ready)!"
        exit 0
    fi
    
    log "Downloading and installing uv via official installer..."
    curl -LsSf https://astral.sh/uv/install.sh | sh || error "Official uv installation failed"
    
    # Add ~/.local/bin to PATH if not already there (official installer uses ~/.local/bin)
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        export PATH="$HOME/.local/bin:$PATH"
        log "Added ~/.local/bin to PATH for this session"
        
        # Update shell profiles for persistent PATH
        for profile in ~/.bashrc ~/.zshrc ~/.profile; do
            if [[ -f "$profile" ]] && ! grep -q ".local/bin" "$profile"; then
                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$profile"
                log "Updated $profile to include ~/.local/bin in PATH"
            fi
        done
    fi
    
    # Verify installation
    if command -v uv &> /dev/null; then
        INSTALLED_VERSION=$(uv --version 2>/dev/null | head -n1)
        success "uv installation completed successfully!"
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: Official installer"
        log "Binary location: $(which uv)"
        
        echo
        log "Basic usage examples:"
        log "  uv init                          # Initialize a new project"
        log "  uv add requests                  # Add a dependency"
        log "  uv run python script.py         # Run a script in project environment"
        log "  uv pip install package          # pip-compatible interface"
        log "  uv python install 3.12          # Install Python 3.12"
        log "  uv venv                          # Create virtual environment"
        echo
        log "For more information: uv --help"
    else
        error "uv installation verification failed"
    fi
    
    exit 0
fi

# Source build method continues below
log "Using source build method..."

# Install dependencies for source build
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for uv..."
    
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

# Get uv source directory
UV_SOURCE_DIR="$BUILD_DIR/$UV_DIR"

# Clone or update repository
if [[ ! -d "$UV_SOURCE_DIR" ]]; then
    log "Cloning uv repository..."
    git clone "$UV_REPO" "$UV_SOURCE_DIR"
else
    log "Updating uv repository..."
    cd "$UV_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$UV_SOURCE_DIR"

# Configure build
log "Configuring uv build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

log "uv source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building uv..."

case $BUILD_TYPE in
    debug)
        log "Building debug version..."
        cargo build
        TARGET_DIR="target/debug"
        ;;
    release)
        log "Building release version..."
        cargo build --release
        TARGET_DIR="target/release"
        ;;
    optimized)
        log "Building optimized version with target-cpu=native..."
        RUSTFLAGS="-C target-cpu=native" cargo build --release
        TARGET_DIR="target/release"
        ;;
esac

# Verify build
if [[ ! -f "$TARGET_DIR/uv" ]]; then
    error "Build failed - uv binary not found in $TARGET_DIR"
fi

success "uv build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $UV_SOURCE_DIR/$TARGET_DIR/uv"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running uv tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing uv..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/uv" /usr/local/bin/
sudo chmod +x /usr/local/bin/uv

# Verify installation
if command -v uv &> /dev/null; then
    INSTALLED_VERSION=$(uv --version 2>/dev/null | head -n1)
    success "uv installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/uv"
    
    # Show basic usage
    echo
    log "Basic usage examples:"
    log "  uv init                          # Initialize a new project"
    log "  uv add requests                  # Add a dependency"
    log "  uv run python script.py         # Run a script in project environment"
    log "  uv pip install package          # pip-compatible interface"
    log "  uv python install 3.12          # Install Python 3.12"
    log "  uv venv                          # Create virtual environment"
    echo
    log "Advanced usage:"
    log "  uv self update                  # Update uv (if installed via official installer)"
    log "  uv tool install ruff            # Install Python tools globally"
    log "  uvx ruff check .                 # Run tools without installing (alias)"
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
    log "For more information: uv --help"
else
    error "uv installation verification failed - uv not found in PATH"
fi

# Update library cache
sudo ldconfig

success "uv installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"
