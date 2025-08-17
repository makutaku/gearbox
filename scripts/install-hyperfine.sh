#!/bin/bash

# hyperfine Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-hyperfine.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
HYPERFINE_DIR="hyperfine"
HYPERFINE_REPO="https://github.com/sharkdp/hyperfine.git"
RUST_MIN_VERSION="1.76.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="binary"   # source, binary, package
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
SETUP_COMPLETIONS=true

# Show help
show_help() {
    cat << EOF
hyperfine Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types (source builds only):
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Installation Methods:
  --source             Build from source (requires Rust toolchain)
  --binary             Download pre-built binary (default, faster)
  --package            Use apt package manager (fastest, may be outdated)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --no-completions     Skip shell completion setup
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: binary installation
  $0 --package         # Use apt package (may be outdated)
  $0 --source          # Build from source
  $0 -o --source       # Optimized source build
  $0 --skip-deps       # Skip dependency installation

About hyperfine:
  A command-line benchmarking tool written in Rust. Provides statistical
  analysis of command execution times with warmup runs, outlier detection,
  and multiple export formats (CSV, JSON, Markdown). Essential for
  performance testing and comparing command implementations.

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
        --package) INSTALL_METHOD="package"; shift ;;
        --skip-deps) SKIP_DEPS=true; shift ;;
        --run-tests) RUN_TESTS=true; shift ;;
        --no-completions) SETUP_COMPLETIONS=false; shift ;;
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

# Setup shell completions
setup_shell_completions() {
    if [[ "$SETUP_COMPLETIONS" != true ]] || ! command -v hyperfine &> /dev/null; then
        log "Skipping shell completion setup"; return 0
    fi
    
    log "Setting up shell completions..."
    local bash_comp_dir="$HOME/.local/share/bash-completion/completions"
    local zsh_comp_dir="$HOME/.local/share/zsh/site-functions"
    local fish_comp_dir="$HOME/.config/fish/completions"
    
    mkdir -p "$bash_comp_dir" "$zsh_comp_dir" "$fish_comp_dir"
    
    # Generate completions using hyperfine
    hyperfine --generate=bash > "$bash_comp_dir/hyperfine" 2>/dev/null || true
    hyperfine --generate=zsh > "$zsh_comp_dir/_hyperfine" 2>/dev/null || true
    hyperfine --generate=fish > "$fish_comp_dir/hyperfine.fish" 2>/dev/null || true
    
    success "Shell completions installed!"
}

# Check if running as root
[[ $EUID -eq 0 ]] && error "This script should not be run as root for security reasons"

# Header
echo; log "========================================"; log "hyperfine Installation Script"
log "========================================"; log "Build type: $BUILD_TYPE"; log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"; echo

# Check if hyperfine is already installed
if command -v hyperfine &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(hyperfine --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    if [[ "$INSTALL_METHOD" != "source" ]]; then
        log "hyperfine is already installed (version: $CURRENT_VERSION)"; log "Use --force to reinstall"; exit 0
    else
        log "Found existing hyperfine installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Package installation method
if [[ "$INSTALL_METHOD" == "package" ]]; then
    log "Using apt package installation method..."
    [[ "$MODE" == "config" ]] && { log "Config mode: would install via apt"; success "Configuration completed!"; exit 0; }
    
    log "Installing hyperfine via apt..."
    sudo apt update; sudo apt install -y hyperfine
    
    if command -v hyperfine &> /dev/null; then
        success "hyperfine installation completed successfully!"
        log "Installed version: $(hyperfine --version 2>/dev/null | head -n1)"
        log "Installation method: APT package manager"
        setup_shell_completions
        echo; log "Basic usage:"; log "  hyperfine 'sleep 1'              # Benchmark single command"
        log "  hyperfine 'cmd1' 'cmd2'         # Compare two commands"
        log "  hyperfine --help                # Show all options"
    else
        error "hyperfine installation verification failed"
    fi
    exit 0
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    [[ "$MODE" == "config" ]] && { log "Config mode: would download hyperfine binary"; success "Configuration completed!"; exit 0; }
    
    # Detect architecture
    ARCH=$(uname -m); case $ARCH in
        x86_64) ARCH_TAG="x86_64-unknown-linux-musl" ;;
        aarch64|arm64) ARCH_TAG="aarch64-unknown-linux-musl" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest hyperfine binary for $ARCH_TAG..."
    RELEASE_URL="https://api.github.com/repos/sharkdp/hyperfine/releases/latest"
    VERSION=$(curl -s "$RELEASE_URL" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$VERSION" ]] && error "Could not determine latest version from GitHub API"
    
    DOWNLOAD_URL="https://github.com/sharkdp/hyperfine/releases/download/${VERSION}/hyperfine-${VERSION}-${ARCH_TAG}.tar.gz"
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" || error "Failed to download hyperfine binary"
    tar -xzf "$(basename "$DOWNLOAD_URL")" || error "Failed to extract archive"
    
    # Find hyperfine binary (should be in extracted directory)
    HYPERFINE_BIN=$(find . -name "hyperfine" -type f -executable | head -n1)
    [[ -z "$HYPERFINE_BIN" ]] && error "hyperfine binary not found in downloaded archive"
    
    sudo cp "$HYPERFINE_BIN" /usr/local/bin/hyperfine; sudo chmod +x /usr/local/bin/hyperfine
    cd /; rm -rf "$TEMP_DIR"
    
    if command -v hyperfine &> /dev/null; then
        success "hyperfine installation completed successfully!"
        log "Installed version: $(hyperfine --version 2>/dev/null | head -n1)"
        log "Installation method: Binary download"; log "Binary location: /usr/local/bin/hyperfine"
        setup_shell_completions
        echo; log "Basic usage:"; log "  hyperfine 'sleep 1'              # Benchmark single command"
        log "  hyperfine 'cmd1' 'cmd2'         # Compare two commands"
        log "  hyperfine --export-json out.json 'cmd'  # Export results"
    else
        error "hyperfine installation verification failed"
    fi
    exit 0
fi

# Source build method
log "Using source build method..."

if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for hyperfine..."
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
else
    log "Skipping dependency installation as requested"
fi

# Ensure Rust access
if ! command -v rustc &> /dev/null; then
    [[ -f ~/.cargo/env ]] && source ~/.cargo/env || error "Rust is not available"
fi

# Build from source
BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"; mkdir -p "$BUILD_DIR"; cd "$BUILD_DIR"
HYPERFINE_SOURCE_DIR="$BUILD_DIR/$HYPERFINE_DIR"

if [[ ! -d "$HYPERFINE_SOURCE_DIR" ]]; then
    log "Cloning hyperfine repository..."; git clone "$HYPERFINE_REPO" "$HYPERFINE_SOURCE_DIR"
else
    log "Updating hyperfine repository..."; cd "$HYPERFINE_SOURCE_DIR"
    git fetch origin; git reset --hard origin/master
fi

cd "$HYPERFINE_SOURCE_DIR"
[[ ! -f "Cargo.toml" ]] && error "Cargo.toml not found"
log "hyperfine source configured successfully"
[[ "$MODE" == "config" ]] && { success "Configuration completed!"; exit 0; }

log "Building hyperfine..."
case $BUILD_TYPE in
    debug) cargo build; TARGET_DIR="target/debug" ;;
    release) cargo build --release; TARGET_DIR="target/release" ;;
    optimized) RUSTFLAGS="-C target-cpu=native" cargo build --release; TARGET_DIR="target/release" ;;
esac

[[ ! -f "$TARGET_DIR/hyperfine" ]] && error "Build failed - hyperfine binary not found"
success "hyperfine build completed successfully!"
[[ "$MODE" == "build" ]] && { success "Build completed!"; log "Binary location: $HYPERFINE_SOURCE_DIR/$TARGET_DIR/hyperfine"; exit 0; }

[[ "$RUN_TESTS" == true ]] && { log "Running hyperfine tests..."; cargo test || warning "Some tests failed"; }

log "Installing hyperfine..."
sudo cp "$TARGET_DIR/hyperfine" /usr/local/bin/; sudo chmod +x /usr/local/bin/hyperfine

if command -v hyperfine &> /dev/null; then
    success "hyperfine installation completed successfully!"
    log "Installed version: $(hyperfine --version 2>/dev/null | head -n1)"
    log "Installation method: Source build ($BUILD_TYPE)"; log "Binary location: /usr/local/bin/hyperfine"
    setup_shell_completions
    
    echo; log "Usage examples:"; log "  hyperfine 'sleep 1'              # Basic benchmark"
    log "  hyperfine 'cmd1' 'cmd2'         # Compare commands"; log "  hyperfine --warmup 3 'cmd'      # Add warmup runs"
    log "  hyperfine --min-runs 5 'cmd'    # Minimum number of runs"
    log "  hyperfine --export-json out.json 'cmd'  # Export to JSON"
    log "  hyperfine --export-csv out.csv 'cmd'    # Export to CSV"
    echo; log "Advanced features:"; log "  hyperfine --parameter-scan num 1 10 'cmd {num}'  # Parameter sweeps"
    log "  hyperfine --setup 'make' 'cmd'          # Setup command"; log "  hyperfine --cleanup 'cleanup' 'cmd'     # Cleanup command"
    log "  hyperfine --shell=bash 'cmd'            # Specify shell"
    echo; log "Key features:"; log "  ✓ Statistical analysis with confidence intervals"
    log "  ✓ Outlier detection and warmup runs"; log "  ✓ Multiple export formats (JSON, CSV, Markdown)"
    log "  ✓ Cross-platform shell command support"; log "  ✓ Parameter scanning for performance analysis"
    echo; log "For more information: hyperfine --help"
else
    error "hyperfine installation verification failed"
fi

sudo ldconfig; success "hyperfine installation completed!"