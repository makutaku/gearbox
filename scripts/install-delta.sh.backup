#!/bin/bash

# delta Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-delta.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DELTA_DIR="delta"
DELTA_REPO="https://github.com/dandavison/delta.git"
RUST_MIN_VERSION="1.59.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="source"   # source, binary
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
SETUP_GIT=true

# Show help
show_help() {
    cat << EOF
delta Installation Script for Debian Linux

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
  --source             Build from source (default, follows gearbox patterns)
  --binary             Download pre-built binary (faster)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --no-git             Skip Git integration setup
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: source build with Git integration
  $0 --binary          # Use pre-built binary (faster)
  $0 -d -c             # Debug build, config only
  $0 -o --run-tests    # Optimized build with tests
  $0 --no-git          # Skip Git configuration
  $0 --skip-deps       # Skip dependency installation

About delta:
  A syntax-highlighting pager for Git, diff, grep, and blame output.
  Provides word-level diff highlighting, side-by-side views, and enhanced
  visual formatting for better code review experience. Integrates seamlessly
  with Git commands and supports themes from bat.

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
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --run-tests)
            RUN_TESTS=true
            shift
            ;;
        --no-git)
            SETUP_GIT=false
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

# Configure Git integration
configure_git_integration() {
    if [[ "$SETUP_GIT" != true ]]; then
        log "Skipping Git integration setup as requested"
        return 0
    fi
    
    log "Configuring Git integration..."
    
    # Check if user wants to configure Git
    if ! ask_user "Configure Git to use delta as the default pager?"; then
        log "Skipping Git configuration"
        return 0
    fi
    
    log "Setting up Git configuration for delta..."
    
    # Set delta as core pager
    git config --global core.pager delta
    
    # Set delta for interactive diff filter
    git config --global interactive.diffFilter 'delta --color-only'
    
    # Enable navigation in delta
    git config --global delta.navigate true
    
    # Set improved merge conflict style
    git config --global merge.conflictStyle zdiff3
    
    # Detect terminal theme
    if [[ "${TERM}" == *"light"* ]] || [[ "${COLORFGBG}" == *"15;0"* ]]; then
        git config --global delta.light true
    else
        git config --global delta.dark true
    fi
    
    success "Git integration configured!"
    echo
    log "Git configuration applied:"
    log "  ✓ core.pager = delta"
    log "  ✓ interactive.diffFilter = delta --color-only"
    log "  ✓ delta.navigate = true"
    log "  ✓ merge.conflictStyle = zdiff3"
    echo
    log "You can customize delta further by editing ~/.gitconfig"
    log "See: https://dandavison.github.io/delta/configuration.html"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "======================================"
log "delta Installation Script"
log "======================================"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
log "Git integration: $SETUP_GIT"
echo

# Check if delta is already installed
if command -v delta &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(delta --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    
    # For binary installation method, respect existing installation
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "delta is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"
        exit 0
    else
        # For source builds, inform but continue (gearbox philosophy: build latest from source)
        log "Found existing delta installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    
    if [[ "$MODE" == "config" ]]; then
        log "Config mode: would download and install delta binary"
        success "Configuration completed (binary installation ready)!"
        exit 0
    fi
    
    # Detect architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH_TAG="x86_64-unknown-linux-gnu" ;;
        aarch64|arm64) ARCH_TAG="aarch64-unknown-linux-gnu" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest delta binary for $ARCH_TAG..."
    
    # Get latest release info
    RELEASE_URL="https://api.github.com/repos/dandavison/delta/releases/latest"
    DOWNLOAD_URL=$(curl -s "$RELEASE_URL" | grep "browser_download_url.*${ARCH_TAG}" | cut -d '"' -f 4)
    
    if [[ -z "$DOWNLOAD_URL" ]]; then
        error "Could not find binary download URL for architecture: $ARCH_TAG"
    fi
    
    # Download and extract
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    log "Downloading from: $DOWNLOAD_URL"
    curl -fLO "$DOWNLOAD_URL"
    
    # Extract (assuming .tar.gz format)
    ARCHIVE_NAME=$(basename "$DOWNLOAD_URL")
    tar -xzf "$ARCHIVE_NAME"
    
    # Find delta binary
    DELTA_BIN=$(find . -name "delta" -type f -executable | head -n1)
    if [[ -z "$DELTA_BIN" ]]; then
        error "delta binary not found in downloaded archive"
    fi
    
    # Install binary
    sudo cp "$DELTA_BIN" /usr/local/bin/delta
    sudo chmod +x /usr/local/bin/delta
    
    # Clean up
    cd /
    rm -rf "$TEMP_DIR"
    
    # Verify installation
    if command -v delta &> /dev/null; then
        INSTALLED_VERSION=$(delta --version 2>/dev/null | head -n1)
        success "delta installation completed successfully!"
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: Binary download"
        log "Binary location: /usr/local/bin/delta"
        
        # Configure Git integration
        echo
        configure_git_integration
        
        echo
        log "Basic usage:"
        log "  git diff | delta              # Use delta as diff pager"
        log "  git show | delta              # View commit with delta"
        log "  delta file1.txt file2.txt     # Direct file comparison"
        log "  For more information: delta --help"
    else
        error "delta installation verification failed"
    fi
    
    exit 0
fi

# Source build method continues below
log "Using source build method..."

# Install dependencies for source build
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for delta..."
    
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

# Get delta source directory
DELTA_SOURCE_DIR="$BUILD_DIR/$DELTA_DIR"

# Clone or update repository
if [[ ! -d "$DELTA_SOURCE_DIR" ]]; then
    log "Cloning delta repository..."
    git clone "$DELTA_REPO" "$DELTA_SOURCE_DIR"
else
    log "Updating delta repository..."
    cd "$DELTA_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$DELTA_SOURCE_DIR"

# Configure build
log "Configuring delta build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

log "delta source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building delta..."

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
if [[ ! -f "$TARGET_DIR/delta" ]]; then
    error "Build failed - delta binary not found in $TARGET_DIR"
fi

success "delta build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $DELTA_SOURCE_DIR/$TARGET_DIR/delta"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running delta tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing delta..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/delta" /usr/local/bin/
sudo chmod +x /usr/local/bin/delta

# Verify installation
if command -v delta &> /dev/null; then
    INSTALLED_VERSION=$(delta --version 2>/dev/null | head -n1)
    success "delta installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/delta"
    
    # Configure Git integration
    echo
    configure_git_integration
    
    # Show usage information
    echo
    log "Basic usage examples:"
    log "  git diff                         # Use delta automatically with Git"
    log "  git show HEAD                    # View commit with delta"
    log "  git log -p                       # View commit history with delta"
    log "  delta file1.txt file2.txt        # Direct file comparison"
    log "  diff file1.txt file2.txt | delta # Pipe diff output to delta"
    echo
    log "Advanced features:"
    log "  delta --side-by-side             # Side-by-side view"
    log "  delta --line-numbers             # Show line numbers"
    log "  delta --navigate                 # Enable navigation (n/N keys)"
    log "  delta --dark/--light             # Force theme"
    echo
    log "Integration with other tools:"
    log "  git -c core.pager=delta diff     # Use delta for one command"
    log "  export PAGER='delta'             # Use as system pager"
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
    log "  ✓ Syntax highlighting for diffs and Git output"
    log "  ✓ Word-level diff highlighting"
    log "  ✓ Side-by-side view with line wrapping"
    log "  ✓ Enhanced merge conflict display"
    log "  ✓ Hyperlink support for commit hashes"
    log "  ✓ Integration with bat themes"
    echo
    log "Configuration: Edit ~/.gitconfig or use 'git config --global delta.*'"
    log "Documentation: https://dandavison.github.io/delta/"
    log "For more information: delta --help"
else
    error "delta installation verification failed - delta not found in PATH"
fi

# Update library cache
sudo ldconfig

success "delta installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"