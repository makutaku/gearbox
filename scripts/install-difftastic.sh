#!/bin/bash

# difftastic Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-difftastic.sh [OPTIONS]

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
DIFFTASTIC_DIR="difftastic"
DIFFTASTIC_REPO="https://github.com/Wilfred/difftastic.git"
RUST_MIN_VERSION="1.74.1"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="binary"   # source, binary
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
SETUP_GIT=true

# Show help
show_help() {
    cat << EOF
difftastic Installation Script for Debian Linux

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

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --no-git             Skip Git integration setup
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: binary installation with Git integration
  $0 --source          # Build from source
  $0 -d -c --source    # Debug build, config only (source)
  $0 -o --source       # Optimized build (source)
  $0 --no-git          # Skip Git configuration
  $0 --skip-deps       # Skip dependency installation

About difftastic:
  A structural diff tool that compares files based on syntax rather than
  line-by-line. Understands code structure for 30+ programming languages
  using tree-sitter parsing. Provides more meaningful diffs by showing
  changes in context of code structure, nesting, and language semantics.

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
        --no-git) SETUP_GIT=false; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) echo "Unknown option: $1"; show_help; exit 1 ;;
    esac
done

# Note: Logging functions now provided by lib/common.sh

ask_user() {
    while true; do
        read -p "$1 (y/n): " yn
        case $yn in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

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

# Configure Git integration
configure_git_integration() {
    if [[ "$SETUP_GIT" != true ]]; then
        log "Skipping Git integration setup as requested"
        return 0
    fi
    
    log "Configuring Git integration..."
    
    if ! ask_user "Configure Git to use difftastic for diffs?"; then
        log "Skipping Git configuration"
        return 0
    fi
    
    log "Setting up Git configuration for difftastic..."
    
    # Set difftastic as external diff tool
    git config --global diff.external difft
    
    # Create useful aliases
    git config --global alias.dlog '-c diff.external=difft log --ext-diff'
    git config --global alias.dshow '-c diff.external=difft show --ext-diff'
    git config --global alias.ddiff '-c diff.external=difft diff'
    
    success "Git integration configured!"
    echo
    log "Git configuration applied:"
    log "  ✓ diff.external = difft"
    log "  ✓ alias.dlog = '-c diff.external=difft log --ext-diff'"
    log "  ✓ alias.dshow = '-c diff.external=difft show --ext-diff'"
    log "  ✓ alias.ddiff = '-c diff.external=difft diff'"
    echo
    log "Usage examples:"
    log "  git ddiff                        # Use difftastic for git diff"
    log "  git dlog -p                      # Use difftastic for git log"
    log "  git dshow HEAD                   # Use difftastic for git show"
    log "  git -c diff.external=difft diff # One-time usage"
}

# Check if running as root
[[ $EUID -eq 0 ]] && error "This script should not be run as root for security reasons"

# Header
echo; log "========================================"; log "difftastic Installation Script"
log "========================================"; log "Build type: $BUILD_TYPE"; log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"; log "Git integration: $SETUP_GIT"; echo

# Check if difftastic is already installed (binary name is 'difft')
if command -v difft &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(difft --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "difftastic is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"; exit 0
    else
        log "Found existing difftastic installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    [[ "$MODE" == "config" ]] && { log "Config mode: would download difftastic binary"; success "Configuration completed!"; exit 0; }
    
    # Detect architecture
    ARCH=$(uname -m); case $ARCH in
        x86_64) ARCH_TAG="x86_64-unknown-linux-gnu" ;;
        aarch64|arm64) ARCH_TAG="aarch64-unknown-linux-gnu" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest difftastic binary for $ARCH_TAG..."
    RELEASE_URL="https://api.github.com/repos/Wilfred/difftastic/releases/latest"
    VERSION=$(curl -s "$RELEASE_URL" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$VERSION" ]] && error "Could not determine latest version from GitHub API"
    
    DOWNLOAD_URL="https://github.com/Wilfred/difftastic/releases/download/${VERSION}/difft-${ARCH_TAG}.tar.gz"
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" || error "Failed to download difftastic binary"
    tar -xzf "$(basename "$DOWNLOAD_URL")" || error "Failed to extract archive"
    [[ ! -f "difft" ]] && error "difft binary not found in downloaded archive"
    
    sudo cp "difft" /usr/local/bin/difft; sudo chmod +x /usr/local/bin/difft
    cd /; rm -rf "$TEMP_DIR"
    
    if command -v difft &> /dev/null; then
        success "difftastic installation completed successfully!"
        log "Installed version: $(difft --version 2>/dev/null | head -n1)"
        log "Installation method: Binary download"; log "Binary location: /usr/local/bin/difft"
        
        echo; configure_git_integration
        
        echo; log "Basic usage:"; log "  difft file1.py file2.py      # Compare two files"
        log "  difft --language=auto f1 f2   # Auto-detect language"
        log "  git ddiff                     # Use with Git (if configured)"
        log "  For more information: https://difftastic.wilfred.me.uk/"
    else
        error "difftastic installation verification failed"
    fi
    exit 0
fi

# Source build method
log "Using source build method..."

if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for difftastic..."
    sudo apt update; sudo apt install -y build-essential git curl pkg-config
    
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
DIFFTASTIC_SOURCE_DIR="$BUILD_DIR/$DIFFTASTIC_DIR"

if [[ ! -d "$DIFFTASTIC_SOURCE_DIR" ]]; then
    log "Cloning difftastic repository..."; git clone "$DIFFTASTIC_REPO" "$DIFFTASTIC_SOURCE_DIR"
else
    log "Updating difftastic repository..."; cd "$DIFFTASTIC_SOURCE_DIR"
    git fetch origin; git reset --hard origin/master
fi

cd "$DIFFTASTIC_SOURCE_DIR"
[[ ! -f "Cargo.toml" ]] && error "Cargo.toml not found"
log "difftastic source configured successfully"
[[ "$MODE" == "config" ]] && { success "Configuration completed!"; exit 0; }

log "Building difftastic..."
case $BUILD_TYPE in
    debug) cargo build; TARGET_DIR="target/debug" ;;
    release) cargo build --release; TARGET_DIR="target/release" ;;
    optimized) RUSTFLAGS="-C target-cpu=native" cargo build --release; TARGET_DIR="target/release" ;;
esac

[[ ! -f "$TARGET_DIR/difft" ]] && error "Build failed - difft binary not found"
success "difftastic build completed successfully!"
[[ "$MODE" == "build" ]] && { success "Build completed!"; log "Binary location: $DIFFTASTIC_SOURCE_DIR/$TARGET_DIR/difft"; exit 0; }

[[ "$RUN_TESTS" == true ]] && { log "Running difftastic tests..."; cargo test || warning "Some tests failed"; }

log "Installing difftastic..."
sudo cp "$TARGET_DIR/difft" /usr/local/bin/; sudo chmod +x /usr/local/bin/difft

if command -v difft &> /dev/null; then
    success "difftastic installation completed successfully!"
    log "Installed version: $(difft --version 2>/dev/null | head -n1)"
    log "Installation method: Source build ($BUILD_TYPE)"; log "Binary location: /usr/local/bin/difft"
    
    echo; configure_git_integration
    
    echo; log "Usage examples:"; log "  difft file1.py file2.py          # Compare two files"
    log "  difft old.js new.js              # Structural diff of JavaScript"
    log "  difft --language=rust a.rs b.rs  # Force language detection"
    log "  difft --side-by-side a.c b.c     # Side-by-side view"
    echo; log "Supported languages:"; log "  • 30+ languages including Rust, Python, JavaScript, C/C++, Java"
    log "  • Tree-sitter based parsing for accurate syntax understanding"
    log "  • Automatic language detection from file extensions"
    echo; log "Key features:"; log "  ✓ Syntax-aware structural diffing"
    log "  ✓ Intelligent alignment of code changes"; log "  ✓ Side-by-side display with context"
    log "  ✓ Git integration for enhanced workflow"; log "  ✓ Fallback to line-oriented diff when needed"
    echo; log "Environment variables:"; log "  DFT_BACKGROUND=light     # Adjust colors for light terminals"
    log "  DFT_PARSE_ERROR_LIMIT=N  # Control parse error tolerance"
    echo; log "For more information: difft --help or https://difftastic.wilfred.me.uk/"
else
    error "difftastic installation verification failed"
fi

sudo ldconfig; success "difftastic installation completed!"
