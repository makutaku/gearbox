#!/bin/bash

# lazygit Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-lazygit.sh [OPTIONS]

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
LAZYGIT_DIR="lazygit"
LAZYGIT_REPO="https://github.com/jesseduffield/lazygit.git"
GO_MIN_VERSION="1.24.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="binary"   # source, binary
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
SETUP_CONFIG=true

# Show help
show_help() {
    cat << EOF
lazygit Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types (source builds only):
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with extra flags)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Installation Methods:
  --source             Build from source (requires Go toolchain)
  --binary             Download pre-built binary (default, faster)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --no-config          Skip configuration directory setup
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: binary installation
  $0 --source          # Build from source
  $0 -d -c --source    # Debug build, config only (source)
  $0 -o --run-tests --source  # Optimized build with tests (source)
  $0 --no-config       # Skip config directory setup
  $0 --skip-deps       # Skip dependency installation

About lazygit:
  A simple terminal UI for git commands. Provides an interactive interface
  for complex git operations including staging lines, interactive rebasing,
  cherry-picking, git bisect, worktree management, and commit graph
  visualization. Perfect for developers who prefer terminal interfaces.

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
        --no-config)
            SETUP_CONFIG=false
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

# Setup configuration directory
setup_config_directory() {
    if [[ "$SETUP_CONFIG" != true ]]; then
        log "Skipping configuration directory setup as requested"
        return 0
    fi
    
    log "Setting up configuration directory..."
    
    local config_dir="$HOME/.config/lazygit"
    
    if [[ -d "$config_dir" ]]; then
        log "Configuration directory already exists: $config_dir"
        return 0
    fi
    
    if ask_user "Create default configuration directory ($config_dir)?"; then
        mkdir -p "$config_dir"
        success "Created configuration directory: $config_dir"
        echo
        log "Configuration tips:"
        log "  ✓ Lazygit will create config.yml on first run"
        log "  ✓ Edit $config_dir/config.yml to customize behavior"
        log "  ✓ Repository-specific config: .git/lazygit.yml"
        log "  ✓ See: https://github.com/jesseduffield/lazygit/wiki/Config"
    else
        log "Skipping configuration directory creation"
        log "Lazygit will create ~/.config/lazygit/ on first run"
    fi
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "========================================"
log "lazygit Installation Script"
log "========================================"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
log "Setup config: $SETUP_CONFIG"
echo

# Check if lazygit is already installed
if command -v lazygit &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(lazygit --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    
    # For binary installation method, respect existing installation
    if [[ "$INSTALL_METHOD" == "binary" ]]; then
        log "lazygit is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"
        exit 0
    else
        # For source builds, inform but continue (gearbox philosophy: build latest from source)
        log "Found existing lazygit installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Binary installation method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation method..."
    
    if [[ "$MODE" == "config" ]]; then
        log "Config mode: would download and install lazygit binary"
        success "Configuration completed (binary installation ready)!"
        exit 0
    fi
    
    # Detect architecture
    ARCH=$(uname -m)
    OS="Linux"
    case $ARCH in
        x86_64) ARCH_TAG="x86_64" ;;
        aarch64|arm64) ARCH_TAG="arm64" ;;
        armv6l) ARCH_TAG="armv6" ;;
        i386|i686) ARCH_TAG="386" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest lazygit binary for ${OS}_${ARCH_TAG}..."
    
    # Get latest release info
    RELEASE_URL="https://api.github.com/repos/jesseduffield/lazygit/releases/latest"
    RELEASE_INFO=$(curl -s "$RELEASE_URL")
    VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name"' | cut -d '"' -f 4)
    
    if [[ -z "$VERSION" ]]; then
        error "Could not determine latest version from GitHub API"
    fi
    
    # Construct download URL
    DOWNLOAD_URL="https://github.com/jesseduffield/lazygit/releases/download/${VERSION}/lazygit_${VERSION#v}_${OS}_${ARCH_TAG}.tar.gz"
    
    log "Downloading version $VERSION from: $DOWNLOAD_URL"
    
    # Download and extract
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    if ! curl -fLO "$DOWNLOAD_URL"; then
        error "Failed to download lazygit binary"
    fi
    
    # Extract (assuming .tar.gz format)
    ARCHIVE_NAME=$(basename "$DOWNLOAD_URL")
    tar -xzf "$ARCHIVE_NAME"
    
    # Find lazygit binary
    if [[ ! -f "lazygit" ]]; then
        error "lazygit binary not found in downloaded archive"
    fi
    
    # Install binary
    sudo cp "lazygit" /usr/local/bin/lazygit
    sudo chmod +x /usr/local/bin/lazygit
    
    # Clean up
    cd /
    rm -rf "$TEMP_DIR"
    
    # Verify installation
    if command -v lazygit &> /dev/null; then
        INSTALLED_VERSION=$(lazygit --version 2>/dev/null | head -n1)
        success "lazygit installation completed successfully!"
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: Binary download"
        log "Binary location: /usr/local/bin/lazygit"
        
        # Setup configuration
        echo
        setup_config_directory
        
        echo
        log "Basic usage:"
        log "  lazygit                      # Start lazygit in current directory"
        log "  lazygit -p /path/to/repo     # Start in specific repository"
        log "  lazygit --help               # Show help and options"
        log "  For more information: https://github.com/jesseduffield/lazygit"
    else
        error "lazygit installation verification failed"
    fi
    
    exit 0
fi

# Source build method continues below
log "Using source build method..."

# Install dependencies for source build
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for lazygit..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl
    
    # Check Go installation
    if ! command -v go &> /dev/null; then
        log "Installing Go..."
        # Download and install Go
        GO_VERSION="1.24.1"  # Use a recent stable version
        GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
        curl -fL "https://golang.org/dl/${GO_TARBALL}" -o "/tmp/${GO_TARBALL}"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
        
        # Add Go to PATH
        export PATH="/usr/local/go/bin:$PATH"
        echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
        
        rm "/tmp/${GO_TARBALL}"
    else
        GO_VERSION=$(go version | grep -oP 'go\d+\.\d+\.\d+' | sed 's/go//')
        log "Found Go version: $GO_VERSION"
        
        if ! version_compare "$GO_VERSION" "$GO_MIN_VERSION"; then
            warning "Go version $GO_VERSION is below minimum $GO_MIN_VERSION"
            log "Consider upgrading Go for best compatibility"
        else
            log "Go version is sufficient (>= $GO_MIN_VERSION)"
        fi
    fi
    
    success "Dependencies installation completed!"
else
    log "Skipping dependency installation as requested"
fi

# Ensure we have access to Go tools
if ! command -v go &> /dev/null; then
    if [[ -d /usr/local/go/bin ]]; then
        export PATH="/usr/local/go/bin:$PATH"
    else
        error "Go is not available in PATH"
    fi
fi

# Create build directory
BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Get lazygit source directory
LAZYGIT_SOURCE_DIR="$BUILD_DIR/$LAZYGIT_DIR"

# Clone or update repository
if [[ ! -d "$LAZYGIT_SOURCE_DIR" ]]; then
    log "Cloning lazygit repository..."
    git clone "$LAZYGIT_REPO" "$LAZYGIT_SOURCE_DIR"
else
    log "Updating lazygit repository..."
    cd "$LAZYGIT_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/master
fi

cd "$LAZYGIT_SOURCE_DIR"

# Configure build
log "Configuring lazygit build..."

# Verify we have go.mod
if [[ ! -f "go.mod" ]]; then
    error "go.mod not found. This doesn't appear to be a valid Go project."
fi

# Get version info for build
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

log "lazygit source configured successfully"
log "Version: $VERSION, Commit: ${COMMIT:0:8}, Date: $DATE"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building lazygit..."

# Prepare build flags
LDFLAGS="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.buildSource=gearbox"

case $BUILD_TYPE in
    debug)
        log "Building debug version..."
        go build -ldflags "$LDFLAGS" -o lazygit
        ;;
    release)
        log "Building release version..."
        go build -ldflags "$LDFLAGS -s -w" -o lazygit
        ;;
    optimized)
        log "Building optimized version..."
        go build -ldflags "$LDFLAGS -s -w" -gcflags="-m=2" -o lazygit
        ;;
esac

# Verify build
if [[ ! -f "lazygit" ]]; then
    error "Build failed - lazygit binary not found"
fi

success "lazygit build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $LAZYGIT_SOURCE_DIR/lazygit"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running lazygit tests..."
    if go test ./...; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing lazygit..."

# Copy binary to /usr/local/bin
sudo cp "lazygit" /usr/local/bin/
sudo chmod +x /usr/local/bin/lazygit

# Verify installation
if command -v lazygit &> /dev/null; then
    INSTALLED_VERSION=$(lazygit --version 2>/dev/null | head -n1)
    success "lazygit installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/lazygit"
    
    # Setup configuration
    echo
    setup_config_directory
    
    # Show usage information
    echo
    log "Basic usage examples:"
    log "  lazygit                          # Start in current directory"
    log "  lazygit -p /path/to/repo         # Start in specific repository"
    log "  lazygit --help                   # Show all command-line options"
    echo
    log "Key features:"
    log "  • Interactive staging (space to stage/unstage)"
    log "  • Visual commit graph and branch management"
    log "  • Interactive rebase, cherry-pick, and merge"
    log "  • Diff view with word-level highlighting"
    log "  • Stash management and worktree support"
    log "  • Custom commands and keybinding configuration"
    echo
    log "Navigation tips:"
    log "  hjkl or arrow keys - Navigate panels"
    log "  tab - Switch between panels"
    log "  space - Stage/unstage files or hunks"
    log "  c - Commit staged changes"
    log "  P - Push to remote"
    log "  p - Pull from remote"
    log "  ? - Show keybindings help"
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
            log "Features: Maximum performance build with additional optimizations"
            ;;
    esac
    echo
    log "Configuration: Edit ~/.config/lazygit/config.yml to customize"
    log "Documentation: https://github.com/jesseduffield/lazygit/wiki"
    log "Keybindings: Press '?' inside lazygit for help"
else
    error "lazygit installation verification failed - lazygit not found in PATH"
fi

success "lazygit installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"
