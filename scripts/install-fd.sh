#!/bin/bash

# fd Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-fd.sh [OPTIONS]

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
FD_DIR="fd"
FD_REPO="https://github.com/sharkdp/fd.git"
RUST_MIN_VERSION="1.88.0"

# Default options
BUILD_TYPE="release"   # debug, release, minimal
MODE="install"         # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
fd Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -m, --minimal         Minimal build (no optional features)

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
  $0 -m -b             # Minimal build, build only
  $0 --skip-deps --run-tests  # Skip deps, run tests

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
        -m|--minimal)
            BUILD_TYPE="minimal"
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
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 0
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 1
        fi
    done
    return 0
}

# Check and install Rust
install_rust() {
    if command -v rustc &> /dev/null; then
        local current_version=$(rustc --version | cut -d' ' -f2)
        log "Found Rust version: $current_version"
        
        if version_compare $current_version $RUST_MIN_VERSION; then
            log "Rust version is sufficient (>= $RUST_MIN_VERSION)"
            return 0
        else
            warning "Rust version $current_version is below minimum required $RUST_MIN_VERSION"
            if ask_user "Do you want to update Rust?"; then
                log "Updating Rust..."
                rustup update || error "Failed to update Rust"
                success "Rust updated successfully"
            else
                error "Cannot proceed with insufficient Rust version"
            fi
        fi
    else
        log "Rust not found, installing..."
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y || error "Failed to install Rust"
        source ~/.cargo/env || error "Failed to source Rust environment"
        success "Rust installed successfully"
    fi
}

# Get cargo build options based on build type
get_cargo_build_options() {
    case $BUILD_TYPE in
        debug)
            echo ""
            ;;
        release)
            echo "--release --locked"
            ;;
        minimal)
            echo "--release --locked --no-default-features"
            ;;
    esac
}

# Get cargo install options
get_cargo_install_options() {
    local install_opts="--path . --locked"
    
    if [[ "$FORCE_INSTALL" == true ]]; then
        install_opts="$install_opts --force"
    fi
    
    case $BUILD_TYPE in
        minimal)
            install_opts="$install_opts --no-default-features"
            ;;
    esac
    
    echo "$install_opts"
}

# Install dependencies
install_dependencies() {
    if [[ "$SKIP_DEPS" == true ]]; then
        log "Skipping dependency installation as requested"
        return 0
    fi

    # Update package list
    log "Updating package list..."
    sudo apt update || error "Failed to update package list"

    # Install basic build tools
    log "Installing build tools..."
    sudo apt install -y \
        build-essential \
        git \
        curl \
        || error "Failed to install build tools"

    # Install Rust
    install_rust

    success "Dependencies installed successfully"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting fd $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"

# Handle fd source code
if [[ -d "$FD_DIR" ]]; then
    log "Found existing fd directory: $FD_DIR"
    
    # Check if it's a git repository
    if [[ -d "$FD_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$FD_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$FD_REPO" ]]; then
            log "Repository origin matches expected fd repository"
            
            if ask_user "Do you want to pull the latest changes from the fd repository?"; then
                log "Pulling latest changes..."
                git pull origin master || error "Failed to pull latest changes"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh fd repository?"; then
                cd ..
                rm -rf "$FD_DIR"
                log "Cloning fd repository..."
                git clone "$FD_REPO" "$FD_DIR" || error "Failed to clone fd repository"
                cd "$FD_DIR"
                success "fd repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh fd repository?"; then
            rm -rf "$FD_DIR"
            log "Cloning fd repository..."
            git clone "$FD_REPO" "$FD_DIR" || error "Failed to clone fd repository"
            success "fd repository cloned successfully"
        else
            error "Cannot proceed without a proper fd source directory"
        fi
    fi
else
    log "Cloning fd repository..."
    git clone "$FD_REPO" "$FD_DIR" || error "Failed to clone fd repository"
    success "fd repository cloned successfully"
fi

# Change to fd directory
cd "$FD_DIR"

# Verify we're in the correct directory
if [[ ! -f "Cargo.toml" ]]; then
    error "Invalid fd source directory - missing Cargo.toml"
fi

# Install dependencies
install_dependencies

# Ensure cargo is in PATH
if ! command -v cargo &> /dev/null; then
    source ~/.cargo/env || error "Failed to source Rust environment"
fi

# Clean previous build
log "Cleaning previous build files..."
cargo clean || warning "Failed to clean previous build, continuing..."

# Get build configuration
CARGO_BUILD_OPTIONS=$(get_cargo_build_options)
CARGO_INSTALL_OPTIONS=$(get_cargo_install_options)

log "Configuring fd with $BUILD_TYPE settings..."
log "Cargo build options: $CARGO_BUILD_OPTIONS"
log "Cargo install options: $CARGO_INSTALL_OPTIONS"

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    if [[ -n "$CARGO_BUILD_OPTIONS" ]]; then
        success "Build command would be: cargo build $CARGO_BUILD_OPTIONS"
    else
        success "Build command would be: cargo build"
    fi
    exit 0
fi

# Build fd
log "Building fd (this may take a while)..."

if [[ -n "$CARGO_BUILD_OPTIONS" ]]; then
    build_with_options cargo "$CARGO_BUILD_OPTIONS"
else
    execute_command_safely cargo build
fi

# Verify build output
BUILD_DIR="target"
if [[ "$BUILD_TYPE" == "debug" ]]; then
    BUILD_DIR="$BUILD_DIR/debug"
else
    BUILD_DIR="$BUILD_DIR/release"
fi

if [[ ! -f "$BUILD_DIR/fd" ]]; then
    error "Build completed but fd executable not found in $BUILD_DIR"
fi

success "Build completed successfully"

# Run tests if requested
if [[ "$RUN_TESTS" == true ]]; then
    log "Running test suite..."
    cargo test || warning "Some tests failed, but continuing"
    success "Test suite completed"
fi

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install or use 'cargo install --path .' to install."
    log "Build output: $(pwd)/$BUILD_DIR/fd"
    exit 0
fi

# Install fd
log "Installing fd..."

if [[ -n "$CARGO_INSTALL_OPTIONS" ]]; then
    build_with_options cargo install "$CARGO_INSTALL_OPTIONS"
else
    execute_command_safely cargo install --path . --locked
fi

# Add cargo bin to PATH if not already there
if [[ ":$PATH:" != *":$HOME/.cargo/bin:"* ]]; then
    log "Adding ~/.cargo/bin to PATH..."
    echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
    export PATH="$HOME/.cargo/bin:$PATH"
    warning "You may need to restart your shell or run 'source ~/.bashrc' for PATH changes to take effect"
fi

# Create system-wide symlinks to ensure our version takes precedence
log "Creating system-wide symlinks..."
sudo ln -sf "$HOME/.cargo/bin/fd" /usr/local/bin/fd || warning "Failed to create fd symlink"
sudo ln -sf "$HOME/.cargo/bin/fd" /usr/local/bin/fdfind || warning "Failed to create fdfind symlink"
success "Symlinks created for both fd and fdfind commands"

# Verify installation
log "Verifying installation..."
# Force PATH update for verification
export PATH="/usr/local/bin:$HOME/.cargo/bin:$PATH"
hash -r
if command -v fd &> /dev/null; then
    success "fd installation verified!"
    echo
    log "fd version information:"
    fd --version
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        debug)
            log "Features: Debug build with symbols for development"
            ;;
        release)
            log "Features: Optimized release build with all default features"
            ;;
        minimal)
            log "Features: Minimal build without optional features"
            ;;
    esac
    echo
    success "fd installation completed successfully!"
    log "You can now use the 'fd' command for fast file searching"
    echo
    log "Usage examples:"
    log "  fd pattern                    # Find files matching pattern"
    log "  fd -t f pattern              # Find only files (not directories)"
    log "  fd -t d pattern              # Find only directories"
    log "  fd -e txt                    # Find files with .txt extension"
    log "  fd pattern /path/to/search   # Search in specific directory"
    echo
    log "Installation paths:"
    log "  fd: $(which fd)"
    log "  fdfind: $(which fdfind)"
    echo
    log "Verifying both commands point to the built version:"
    log "  fd version: $(fd --version 2>/dev/null || echo 'not found')"
    log "  fdfind version: $(fdfind --version 2>/dev/null || echo 'not found')"
    log "Script completed in directory: $(pwd)"
else
    error "fd installation verification failed - try restarting your shell or run 'source ~/.bashrc'"
fi