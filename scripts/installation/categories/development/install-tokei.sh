#!/bin/bash

# tokei Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-tokei.sh [OPTIONS]

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
TOKEI_DIR="tokei"
TOKEI_REPO="https://github.com/XAMPPRocky/tokei.git"
RUST_MIN_VERSION="1.70.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="source"   # source, binary (default to source for reliability)
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
tokei Installation Script for Debian Linux

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
  --source             Build from source (default, most reliable)
  --binary             Download pre-built binary (faster, may not be available)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

About tokei:
  A program that displays statistics about your code. Counts lines of code,
  comments, and blank lines in over 200 programming languages. Provides
  detailed breakdown by language, file, and supports output in multiple
  formats including JSON, YAML, and CBOR.

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--debug) BUILD_TYPE="debug"; shift ;;
        -r|--release) BUILD_TYPE="release"; shift ;;
        -o|--optimized) BUILD_TYPE="optimized"; shift ;;
        -c|--config-only) MODE="config"; shift ;;
        -b|--build-only) MODE="build"; shift ;;
        -i|--install) MODE="install"; shift ;;
        --source) INSTALL_METHOD="source"; shift ;;
        --binary) INSTALL_METHOD="binary"; shift ;;
        --skip-deps) SKIP_DEPS=true; shift ;;
        --run-tests) RUN_TESTS=true; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) echo "Unknown option: $1"; show_help; exit 1 ;;
    esac
done

# Note: Logging functions now provided by lib/common.sh

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
echo; log "======================================"; log "tokei Installation Script"
log "======================================"; log "Build type: $BUILD_TYPE"; log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"; echo

# Check if tokei is already installed
if command -v tokei &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(tokei --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "tokei is already installed (version: $CURRENT_VERSION)"; log "Use --force to reinstall"; exit 0
    else
        log "Found existing tokei installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    [[ "$MODE" == "config" ]] && { log "Config mode: would download tokei binary"; success "Configuration completed!"; exit 0; }
    
    # Detect architecture
    ARCH=$(uname -m); case $ARCH in
        x86_64) ARCH_TAG="x86_64-unknown-linux-musl" ;;
        aarch64|arm64) ARCH_TAG="aarch64-unknown-linux-musl" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest tokei binary for $ARCH_TAG..."
    RELEASE_URL="https://api.github.com/repos/XAMPPRocky/tokei/releases/latest"
    VERSION=$(curl -s "$RELEASE_URL" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$VERSION" ]] && error "Could not determine latest version from GitHub API"
    
    DOWNLOAD_URL="https://github.com/XAMPPRocky/tokei/releases/download/${VERSION}/tokei-${ARCH_TAG}.tar.gz"
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" || error "Failed to download tokei binary"
    tar -xzf "$(basename "$DOWNLOAD_URL")" || error "Failed to extract archive"
    [[ ! -f "tokei" ]] && error "tokei binary not found in downloaded archive"
    
    sudo cp "tokei" /usr/local/bin/tokei; sudo chmod +x /usr/local/bin/tokei
    cd /; rm -rf "$TEMP_DIR"
    
    if command -v tokei &> /dev/null; then
        success "tokei installation completed successfully!"
        log "Installed version: $(tokei --version 2>/dev/null | head -n1)"
        log "Installation method: Binary download"; log "Binary location: /usr/local/bin/tokei"
        echo; log "Basic usage:"; log "  tokei                    # Count lines in current directory"
        log "  tokei --languages        # List supported languages"
        log "  tokei --output json      # JSON output format"
    else
        error "tokei installation verification failed"
    fi
    exit 0
fi

# Source build method
log "Using source build method..."

if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for tokei..."
    install_build_tools
    install_rust_if_needed "$RUST_MIN_VERSION"
    success "Dependencies installation completed!"
else
    log "Skipping dependency installation as requested"
fi

# Ensure Rust access
if ! command -v rustc &> /dev/null; then
    [[ -f ~/.cargo/env ]] && source ~/.cargo/env || error "Rust is not available"
fi

# Build from source
ensure_directory "$BUILD_DIR"
cd "$BUILD_DIR"
TOKEI_SOURCE_DIR="$BUILD_DIR/$TOKEI_DIR"

clone_or_update_repo "$TOKEI_REPO" "$TOKEI_SOURCE_DIR" "master"
cd "$TOKEI_SOURCE_DIR"
[[ ! -f "Cargo.toml" ]] && error "Cargo.toml not found"
log "tokei source configured successfully"
[[ "$MODE" == "config" ]] && { success "Configuration completed!"; exit 0; }

log "Building tokei..."
case $BUILD_TYPE in
    debug) 
        execute_command_safely cargo build
        TARGET_DIR="target/debug" 
        ;;
    release) 
        execute_command_safely cargo build --release
        TARGET_DIR="target/release" 
        ;;
    optimized) 
        env RUSTFLAGS="-C target-cpu=native" execute_command_safely cargo build --release
        TARGET_DIR="target/release" 
        ;;
esac

[[ ! -f "$TARGET_DIR/tokei" ]] && error "Build failed - tokei binary not found"
success "tokei build completed successfully!"
[[ "$MODE" == "build" ]] && { success "Build completed!"; log "Binary location: $TOKEI_SOURCE_DIR/$TARGET_DIR/tokei"; exit 0; }

[[ "$RUN_TESTS" == true ]] && { log "Running tokei tests..."; cargo test || warning "Some tests failed"; }

log "Installing tokei..."
# Try system-wide installation first, fall back to user-local if sudo fails
if sudo cp "$TARGET_DIR/tokei" "/usr/local/bin/tokei" && sudo chmod 755 "/usr/local/bin/tokei" 2>/dev/null; then
    INSTALL_PATH="/usr/local/bin/tokei"
    log "Installed tokei to system directory: $INSTALL_PATH"
else
    warning "System-wide installation failed, installing to user directory"
    ensure_directory "$HOME/.local/bin"
    cp "$TARGET_DIR/tokei" "$HOME/.local/bin/tokei"
    chmod +x "$HOME/.local/bin/tokei"
    INSTALL_PATH="$HOME/.local/bin/tokei"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        log "Adding ~/.local/bin to PATH..."
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
        export PATH="$HOME/.local/bin:$PATH"
        warning "You may need to restart your shell or run 'source ~/.bashrc' for PATH changes to take effect"
    fi
    log "Installed tokei to user directory: $INSTALL_PATH"
fi

if command -v tokei &> /dev/null; then
    success "tokei installation completed successfully!"
    log "Installed version: $(tokei --version 2>/dev/null | head -n1)"
    log "Installation method: Source build ($BUILD_TYPE)"; log "Binary location: $INSTALL_PATH"
    echo; log "Usage examples:"; log "  tokei                    # Count lines in current directory"
    log "  tokei src/               # Count lines in specific directory"
    log "  tokei --languages        # List supported languages"
    log "  tokei --output json      # JSON output"; log "  tokei --files            # Show per-file statistics"
    log "  tokei --exclude '*.md'   # Exclude markdown files"
    echo; log "Key features:"; log "  ✓ 200+ programming languages supported"
    log "  ✓ Lines of code, comments, and blank lines"; log "  ✓ Multiple output formats (JSON, YAML, CBOR)"
    log "  ✓ Fast parallel processing"; log "  ✓ Git ignore support"
    echo; log "For more information: tokei --help"
else
    error "tokei installation verification failed"
fi

# Update library cache if possible (non-critical)
sudo ldconfig 2>/dev/null || true
success "tokei installation completed!"
