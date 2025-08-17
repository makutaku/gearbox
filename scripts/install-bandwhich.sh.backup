#!/bin/bash

# bandwhich Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-bandwhich.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BANDWHICH_DIR="bandwhich"
BANDWHICH_REPO="https://github.com/imsnif/bandwhich.git"
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
bandwhich Installation Script for Debian Linux

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

About bandwhich:
  A terminal bandwidth utilization tool that displays network activity
  by process. Shows which processes are using bandwidth and how much.
  Essential for network troubleshooting and monitoring.

NOTE: bandwhich requires root privileges to monitor network interfaces.
Use with: sudo bandwhich

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
echo; log "========================================"; log "bandwhich Installation Script"
log "========================================"; log "Build type: $BUILD_TYPE"; log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"; echo

# Check if bandwhich is already installed
if command -v bandwhich &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(bandwhich --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "bandwhich is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"; exit 0
    else
        log "Found existing bandwhich installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    
    # Detect architecture
    ARCH=$(uname -m); case $ARCH in
        x86_64) ARCH_TAG="x86_64-unknown-linux-musl" ;;
        aarch64|arm64) ARCH_TAG="aarch64-unknown-linux-musl" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest bandwhich binary for $ARCH_TAG..."
    RELEASE_URL="https://api.github.com/repos/imsnif/bandwhich/releases/latest"
    VERSION=$(curl -s "$RELEASE_URL" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$VERSION" ]] && error "Could not determine latest version from GitHub API"
    
    DOWNLOAD_URL="https://github.com/imsnif/bandwhich/releases/download/${VERSION}/bandwhich-${VERSION}-${ARCH_TAG}.tar.gz"
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" || error "Failed to download bandwhich binary"
    tar -xzf "$(basename "$DOWNLOAD_URL")" || error "Failed to extract archive"
    [[ ! -f "bandwhich" ]] && error "bandwhich binary not found in downloaded archive"
    
    sudo cp "bandwhich" /usr/local/bin/bandwhich; sudo chmod +x /usr/local/bin/bandwhich
    cd /; rm -rf "$TEMP_DIR"
    
    if command -v bandwhich &> /dev/null; then
        success "bandwhich installation completed successfully!"
        log "Installed version: $(bandwhich --version 2>/dev/null | head -n1)"
        log "Installation method: Binary download"; log "Binary location: /usr/local/bin/bandwhich"
        echo; log "Usage examples:"; log "  sudo bandwhich                  # Monitor all interfaces"
        log "  sudo bandwhich -i eth0          # Monitor specific interface"
        log "  sudo bandwhich --raw            # Raw output mode"
        echo; warning "bandwhich requires root privileges to access network interfaces"
        log "Always run with: sudo bandwhich"
    else
        error "bandwhich installation verification failed"
    fi
    exit 0
fi

# Source build method
log "Using source build method..."

if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for bandwhich..."
    sudo apt update; sudo apt install -y build-essential git curl pkg-config libpcap-dev
    
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
BANDWHICH_SOURCE_DIR="$BUILD_DIR/$BANDWHICH_DIR"

if [[ ! -d "$BANDWHICH_SOURCE_DIR" ]]; then
    log "Cloning bandwhich repository..."; git clone "$BANDWHICH_REPO" "$BANDWHICH_SOURCE_DIR"
else
    log "Updating bandwhich repository..."; cd "$BANDWHICH_SOURCE_DIR"
    git fetch origin; git reset --hard origin/main
fi

cd "$BANDWHICH_SOURCE_DIR"
[[ ! -f "Cargo.toml" ]] && error "Cargo.toml not found"
log "bandwhich source configured successfully"

log "Building bandwhich..."
case $BUILD_TYPE in
    debug) cargo build; TARGET_DIR="target/debug" ;;
    release) cargo build --release; TARGET_DIR="target/release" ;;
    optimized) RUSTFLAGS="-C target-cpu=native" cargo build --release; TARGET_DIR="target/release" ;;
esac

[[ ! -f "$TARGET_DIR/bandwhich" ]] && error "Build failed - bandwhich binary not found"
success "bandwhich build completed successfully!"

[[ "$RUN_TESTS" == true ]] && { log "Running bandwhich tests..."; cargo test || warning "Some tests failed"; }

log "Installing bandwhich..."
sudo cp "$TARGET_DIR/bandwhich" /usr/local/bin/; sudo chmod +x /usr/local/bin/bandwhich

if command -v bandwhich &> /dev/null; then
    success "bandwhich installation completed successfully!"
    log "Installed version: $(bandwhich --version 2>/dev/null | head -n1)"
    log "Installation method: Source build ($BUILD_TYPE)"; log "Binary location: /usr/local/bin/bandwhich"
    
    echo; log "Usage examples:"; log "  sudo bandwhich                  # Monitor all interfaces"
    log "  sudo bandwhich -i eth0          # Monitor specific interface"
    log "  sudo bandwhich --raw            # Raw output mode"
    log "  sudo bandwhich --no-resolve     # Don't resolve hostnames"
    echo; log "Key features:"; log "  ✓ Real-time network bandwidth monitoring by process"
    log "  ✓ Per-interface and per-connection statistics"; log "  ✓ DNS resolution of connections"
    log "  ✓ Filterable by interface or process"; echo
    warning "bandwhich requires root privileges to access network interfaces"
    log "Always run with: sudo bandwhich"; log "For more information: bandwhich --help"
else
    error "bandwhich installation verification failed"
fi

sudo ldconfig; success "bandwhich installation completed!"