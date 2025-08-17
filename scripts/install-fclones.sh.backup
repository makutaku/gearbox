#!/bin/bash

# fclones Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-fclones.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FCLONES_DIR="fclones"
FCLONES_REPO="https://github.com/pkolaczk/fclones.git"
RUST_MIN_VERSION="1.70.0"

# Default options
BUILD_TYPE="release"   # debug, release, optimized
MODE="install"         # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
fclones Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with LTO and target-cpu=native)

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

About fclones:
  Efficient duplicate file finder that helps identify groups of identical files,
  remove redundant copies, and replace duplicates with links for deduplication.

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
log "=================================="
log "fclones Installation Script"
log "=================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
echo

# Check if fclones is already installed
if command -v fclones &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(fclones --version | head -n1 | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    log "fclones is already installed (version: $CURRENT_VERSION)"
    log "Use --force to reinstall"
    exit 0
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for fclones..."
    
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

# Get fclones source directory
FCLONES_SOURCE_DIR="$BUILD_DIR/$FCLONES_DIR"

# Clone or update repository
if [[ ! -d "$FCLONES_SOURCE_DIR" ]]; then
    log "Cloning fclones repository..."
    git clone "$FCLONES_REPO" "$FCLONES_SOURCE_DIR"
else
    log "Updating fclones repository..."
    cd "$FCLONES_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$FCLONES_SOURCE_DIR"

# Configure build
log "Configuring fclones build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

log "fclones source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building fclones..."

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
        log "Building optimized version with LTO and native CPU..."
        RUSTFLAGS="-C target-cpu=native -C lto=fat" cargo build --release
        TARGET_DIR="target/release"
        ;;
esac

# Verify build
if [[ ! -f "$TARGET_DIR/fclones" ]]; then
    error "Build failed - fclones binary not found in $TARGET_DIR"
fi

success "fclones build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $FCLONES_SOURCE_DIR/$TARGET_DIR/fclones"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running fclones tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing fclones..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/fclones" /usr/local/bin/
sudo chmod +x /usr/local/bin/fclones

# Verify installation
if command -v fclones &> /dev/null; then
    INSTALLED_VERSION=$(fclones --version | head -n1)
    success "fclones installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Binary location: /usr/local/bin/fclones"
    
    # Show basic usage
    echo
    log "Basic usage examples:"
    log "  fclones group .                    # Find duplicate files in current directory"
    log "  fclones group --depth 1 .         # Find duplicates with limited depth"
    log "  fclones group --cache .            # Enable caching for faster subsequent runs"
    log "  fclones remove --dry-run .         # Preview duplicate removal (safe)"
    log "  fclones link .                     # Replace duplicates with hard links"
    echo
    log "For more information: fclones --help"
else
    error "Installation failed - fclones not found in PATH"
fi

# Update library cache
sudo ldconfig

success "fclones installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"