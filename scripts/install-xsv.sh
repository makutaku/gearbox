#!/bin/bash

# xsv Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-xsv.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
XSV_DIR="xsv"
XSV_REPO="https://github.com/BurntSushi/xsv.git"
RUST_MIN_VERSION="1.70.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="binary"   # source, binary
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
xsv Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native)

Installation Methods:
  --source             Build from source (requires Rust toolchain)
  --binary             Download pre-built binary (default, faster)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

About xsv:
  A fast CSV command line toolkit written in Rust. Provides indexing,
  slicing, analysis, splitting, and joining of CSV data. Essential for
  data analysis workflows with powerful filtering and transformation
  capabilities.

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--debug) BUILD_TYPE="debug"; shift ;;
        -r|--release) BUILD_TYPE="release"; shift ;;
        -o|--optimized) BUILD_TYPE="optimized"; shift ;;
        --source) INSTALL_METHOD="source"; shift ;;
        --binary) INSTALL_METHOD="binary"; shift ;;
        --skip-deps) SKIP_DEPS=true; shift ;;
        --run-tests) RUN_TESTS=true; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) echo "Unknown option: $1"; show_help; exit 1 ;;
    esac
done

# Logging functions
log() { echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# Version comparison
version_compare() {
    if [[ $1 == $2 ]]; then return 0; fi
    local IFS=.; local i ver1=($1) ver2=($2)
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do ver1[i]=0; done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then ver2[i]=0; fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then return 0; fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then return 1; fi
    done; return 0
}

# Check if running as root
[[ $EUID -eq 0 ]] && error "This script should not be run as root for security reasons"

# Header
echo; log "===================================="; log "xsv Installation Script"
log "===================================="; log "Build type: $BUILD_TYPE"; log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"; echo

# Check if xsv is already installed
if command -v xsv &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(xsv --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "xsv is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"; exit 0
    else
        log "Found existing xsv installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    
    # Detect architecture
    ARCH=$(uname -m); case $ARCH in
        x86_64) ARCH_TAG="x86_64-unknown-linux-musl" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest xsv binary for $ARCH_TAG..."
    RELEASE_URL="https://api.github.com/repos/BurntSushi/xsv/releases/latest"
    VERSION=$(curl -s "$RELEASE_URL" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$VERSION" ]] && error "Could not determine latest version from GitHub API"
    
    DOWNLOAD_URL="https://github.com/BurntSushi/xsv/releases/download/${VERSION}/xsv-${VERSION}-${ARCH_TAG}.tar.gz"
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" || error "Failed to download xsv binary"
    tar -xzf "$(basename "$DOWNLOAD_URL")" || error "Failed to extract archive"
    [[ ! -f "xsv" ]] && error "xsv binary not found in downloaded archive"
    
    sudo cp "xsv" /usr/local/bin/xsv; sudo chmod +x /usr/local/bin/xsv
    cd /; rm -rf "$TEMP_DIR"
    
    if command -v xsv &> /dev/null; then
        success "xsv installation completed successfully!"
        log "Installed version: $(xsv --version 2>/dev/null | head -n1)"
        log "Installation method: Binary download"; log "Binary location: /usr/local/bin/xsv"
        echo; log "Basic usage:"; log "  xsv headers data.csv            # Show column headers"
        log "  xsv count data.csv              # Count rows"
        log "  xsv select name,age data.csv    # Select specific columns"
        log "  xsv stats data.csv              # Statistical summary"
        echo; log "Key features:"; log "  ✓ Fast CSV indexing and searching"
        log "  ✓ Powerful filtering and joining"; log "  ✓ Statistical analysis of data"
        log "  ✓ Memory-efficient large file processing"
        echo; log "For more information: xsv --help"
    else
        error "xsv installation verification failed"
    fi
    exit 0
fi

# Source build method
log "Using source build method..."

if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for xsv..."
    sudo apt update; sudo apt install -y build-essential git curl
    
    if ! command -v rustc &> /dev/null; then
        log "Installing Rust..."
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y; source ~/.cargo/env
    else
        RUST_VERSION=$(rustc --version | grep -oP '\d+\.\d+\.\d+')
        log "Found Rust version: $RUST_VERSION"
        if ! version_compare "$RUST_VERSION" "$RUST_MIN_VERSION"; then
            log "Updating Rust..."; rustup update
        fi
    fi
    rustup update stable; rustup default stable; success "Dependencies installation completed!"
fi

# Ensure Rust access
if ! command -v rustc &> /dev/null; then
    [[ -f ~/.cargo/env ]] && source ~/.cargo/env || error "Rust is not available"
fi

# Build from source
BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"; mkdir -p "$BUILD_DIR"; cd "$BUILD_DIR"
XSV_SOURCE_DIR="$BUILD_DIR/$XSV_DIR"

if [[ ! -d "$XSV_SOURCE_DIR" ]]; then
    log "Cloning xsv repository..."; git clone "$XSV_REPO" "$XSV_SOURCE_DIR"
else
    log "Updating xsv repository..."; cd "$XSV_SOURCE_DIR"
    git fetch origin; git reset --hard origin/master
fi

cd "$XSV_SOURCE_DIR"
[[ ! -f "Cargo.toml" ]] && error "Cargo.toml not found"
log "xsv source configured successfully"

log "Building xsv..."
case $BUILD_TYPE in
    debug) cargo build; TARGET_DIR="target/debug" ;;
    release) cargo build --release; TARGET_DIR="target/release" ;;
    optimized) RUSTFLAGS="-C target-cpu=native" cargo build --release; TARGET_DIR="target/release" ;;
esac

[[ ! -f "$TARGET_DIR/xsv" ]] && error "Build failed - xsv binary not found"
success "xsv build completed successfully!"

[[ "$RUN_TESTS" == true ]] && { log "Running xsv tests..."; cargo test || warning "Some tests failed"; }

log "Installing xsv..."
sudo cp "$TARGET_DIR/xsv" /usr/local/bin/; sudo chmod +x /usr/local/bin/xsv

if command -v xsv &> /dev/null; then
    success "xsv installation completed successfully!"
    log "Installed version: $(xsv --version 2>/dev/null | head -n1)"
    log "Installation method: Source build ($BUILD_TYPE)"; log "Binary location: /usr/local/bin/xsv"
    
    echo; log "Essential commands:"; log "  xsv headers data.csv            # Show column headers"
    log "  xsv count data.csv              # Count rows"
    log "  xsv select name,age data.csv    # Select specific columns"
    log "  xsv search 'pattern' data.csv   # Search for pattern"
    log "  xsv stats data.csv              # Statistical summary"
    log "  xsv sort -s age data.csv        # Sort by column"
    echo; log "Advanced operations:"; log "  xsv join id data1.csv id data2.csv  # Join CSV files"
    log "  xsv frequency -s status data.csv    # Frequency count"
    log "  xsv sample 100 data.csv         # Random sample"
    log "  xsv slice -s 10 -e 20 data.csv  # Extract rows 10-20"
    echo; log "Key features:"; log "  ✓ Fast CSV indexing for large files"
    log "  ✓ Memory-efficient streaming operations"; log "  ✓ Powerful filtering and transformation"
    log "  ✓ Statistical analysis and frequency counting"; log "  ✓ Joining and splitting CSV files"
    echo; log "For complete documentation: xsv --help"
else
    error "xsv installation verification failed"
fi

sudo ldconfig; success "xsv installation completed!"