#!/bin/bash

# ripgrep Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-ripgrep.sh [OPTIONS]

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
RIPGREP_DIR="ripgrep"
RIPGREP_REPO="https://github.com/BurntSushi/ripgrep.git"
RUST_MIN_VERSION="1.88.0"

# Default options
BUILD_TYPE="release"   # debug, release, static, optimized
MODE="install"         # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
ENABLE_PCRE2=true

# Show help
show_help() {
    cat << EOF
ripgrep Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized with PCRE2)
  -s, --static          Static build (self-contained MUSL binary)
  -o, --optimized       CPU-optimized build (native target CPU)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --force              Force reinstallation if already installed
  --no-pcre2           Build without PCRE2 support
  -h, --help           Show this help message

Examples:
  $0                   # Default: release build with PCRE2 and install
  $0 -d -c             # Debug build, config only
  $0 -s --no-pcre2     # Static build without PCRE2
  $0 -o --run-tests    # CPU-optimized build with tests

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
        -s|--static)
            BUILD_TYPE="static"
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
        --no-pcre2)
            ENABLE_PCRE2=false
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
    local options="--release --locked"
    
    case $BUILD_TYPE in
        debug)
            options=""
            ;;
        release|optimized)
            options="--release --locked"
            ;;
        static)
            options="--release --locked --target x86_64-unknown-linux-musl"
            ;;
    esac
    
    # Add PCRE2 feature if enabled
    if [[ "$ENABLE_PCRE2" == true ]]; then
        options="$options --features pcre2"
    fi
    
    echo "$options"
}

# Get cargo install options
get_cargo_install_options() {
    local install_opts="--path . --locked"
    
    if [[ "$FORCE_INSTALL" == true ]]; then
        install_opts="$install_opts --force"
    fi
    
    # Add PCRE2 feature if enabled
    if [[ "$ENABLE_PCRE2" == true ]]; then
        install_opts="$install_opts --features pcre2"
    fi
    
    echo "$install_opts"
}

# Get environment variables for build
get_build_env() {
    case $BUILD_TYPE in
        optimized)
            echo "RUSTFLAGS=\"-C target-cpu=native\""
            ;;
        static)
            if [[ "$ENABLE_PCRE2" == true ]]; then
                echo "PCRE2_SYS_STATIC=1"
            fi
            ;;
        *)
            echo ""
            ;;
    esac
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

    # Install PCRE2 development libraries if enabled
    if [[ "$ENABLE_PCRE2" == true ]]; then
        log "Installing PCRE2 development libraries..."
        sudo apt install -y \
            libpcre2-dev \
            libpcre2-posix2 \
            || warning "PCRE2 libraries not available, will build from source"
    fi

    # Install MUSL tools for static builds
    if [[ "$BUILD_TYPE" == "static" ]]; then
        log "Installing MUSL tools for static build..."
        sudo apt install -y \
            musl-tools \
            musl-dev \
            || warning "MUSL tools not available, static build may fail"
    fi

    # Install Rust
    install_rust

    # Add MUSL target for static builds
    if [[ "$BUILD_TYPE" == "static" ]]; then
        log "Adding MUSL target for static builds..."
        rustup target add x86_64-unknown-linux-musl || warning "Failed to add MUSL target"
    fi

    success "Dependencies installed successfully"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting ripgrep $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "PCRE2 support: $ENABLE_PCRE2"

# Handle ripgrep source code
if [[ -d "$RIPGREP_DIR" ]]; then
    log "Found existing ripgrep directory: $RIPGREP_DIR"
    
    # Check if it's a git repository
    if [[ -d "$RIPGREP_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$RIPGREP_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$RIPGREP_REPO" ]]; then
            log "Repository origin matches expected ripgrep repository"
            
            if ask_user "Do you want to pull the latest changes from the ripgrep repository?"; then
                log "Pulling latest changes..."
                git pull origin master || error "Failed to pull latest changes"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh ripgrep repository?"; then
                cd ..
                rm -rf "$RIPGREP_DIR"
                log "Cloning ripgrep repository..."
                git clone "$RIPGREP_REPO" "$RIPGREP_DIR" || error "Failed to clone ripgrep repository"
                cd "$RIPGREP_DIR"
                success "ripgrep repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh ripgrep repository?"; then
            rm -rf "$RIPGREP_DIR"
            log "Cloning ripgrep repository..."
            git clone "$RIPGREP_REPO" "$RIPGREP_DIR" || error "Failed to clone ripgrep repository"
            success "ripgrep repository cloned successfully"
        else
            error "Cannot proceed without a proper ripgrep source directory"
        fi
    fi
else
    log "Cloning ripgrep repository..."
    git clone "$RIPGREP_REPO" "$RIPGREP_DIR" || error "Failed to clone ripgrep repository"
    success "ripgrep repository cloned successfully"
fi

# Change to ripgrep directory
cd "$RIPGREP_DIR"

# Verify we're in the correct directory
if [[ ! -f "Cargo.toml" ]]; then
    error "Invalid ripgrep source directory - missing Cargo.toml"
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
BUILD_ENV=$(get_build_env)

log "Configuring ripgrep with $BUILD_TYPE settings..."
log "Cargo build options: $CARGO_BUILD_OPTIONS"
log "Cargo install options: $CARGO_INSTALL_OPTIONS"
if [[ -n "$BUILD_ENV" ]]; then
    log "Build environment: $BUILD_ENV"
fi

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    if [[ -n "$BUILD_ENV" ]]; then
        success "Build command would be: $BUILD_ENV cargo build $CARGO_BUILD_OPTIONS"
    else
        success "Build command would be: cargo build $CARGO_BUILD_OPTIONS"
    fi
    exit 0
fi

# Build ripgrep
log "Building ripgrep (this may take a while)..."

if [[ -n "$BUILD_ENV" ]]; then
    if [[ -n "$CARGO_BUILD_OPTIONS" ]]; then
        env $BUILD_ENV build_with_options cargo "$CARGO_BUILD_OPTIONS"
    else
        env $BUILD_ENV execute_command_safely cargo build
    fi
else
    if [[ -n "$CARGO_BUILD_OPTIONS" ]]; then
        build_with_options cargo "$CARGO_BUILD_OPTIONS"
    else
        execute_command_safely cargo build
    fi
fi

# Verify build output
BUILD_DIR="target"
if [[ "$BUILD_TYPE" == "debug" ]]; then
    BUILD_DIR="$BUILD_DIR/debug"
elif [[ "$BUILD_TYPE" == "static" ]]; then
    BUILD_DIR="$BUILD_DIR/x86_64-unknown-linux-musl/release"
else
    BUILD_DIR="$BUILD_DIR/release"
fi

if [[ ! -f "$BUILD_DIR/rg" ]]; then
    error "Build completed but rg executable not found in $BUILD_DIR"
fi

success "Build completed successfully"

# Run tests if requested
if [[ "$RUN_TESTS" == true ]]; then
    log "Running test suite..."
    if [[ "$ENABLE_PCRE2" == true ]]; then
        cargo test --verbose --workspace --features pcre2 || warning "Some tests failed, but continuing"
    else
        cargo test --verbose --workspace || warning "Some tests failed, but continuing"
    fi
    success "Test suite completed"
fi

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install or use 'cargo install --path .' to install."
    log "Build output: $(pwd)/$BUILD_DIR/rg"
    exit 0
fi

# Install ripgrep
log "Installing ripgrep..."

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

# Create system-wide symlink to ensure our version takes precedence
log "Creating system-wide symlink..."
sudo ln -sf "$HOME/.cargo/bin/rg" /usr/local/bin/rg || warning "Failed to create rg symlink"
success "Symlink created for rg command"

# Verify installation
log "Verifying installation..."
# Force PATH update for verification
export PATH="/usr/local/bin:$HOME/.cargo/bin:$PATH"
# Clear bash command hash table to ensure new symlinks are used
hash -r
if command -v rg &> /dev/null; then
    success "ripgrep installation verified!"
    echo
    log "ripgrep version information:"
    rg --version
    echo
    log "Build type: $BUILD_TYPE"
    log "PCRE2 support: $ENABLE_PCRE2"
    case $BUILD_TYPE in
        debug)
            log "Features: Debug build with symbols for development"
            ;;
        release)
            log "Features: Optimized release build with standard features"
            ;;
        static)
            log "Features: Self-contained static binary (MUSL)"
            ;;
        optimized)
            log "Features: CPU-optimized build for maximum performance"
            ;;
    esac
    echo
    success "ripgrep installation completed successfully!"
    log "You can now use the 'rg' command for fast text searching"
    echo
    log "Usage examples:"
    log "  rg pattern                    # Search for pattern in current directory"
    log "  rg -i pattern                 # Case-insensitive search"
    log "  rg -t py pattern              # Search only in Python files"
    log "  rg -C 3 pattern               # Show 3 lines of context"
    if [[ "$ENABLE_PCRE2" == true ]]; then
        log "  rg -P '(?<=foo)bar'           # PCRE2 lookbehind (use -P flag)"
    fi
    echo
    log "Installation paths:"
    log "  rg: $(which rg)"
    echo
    log "To use immediately in this terminal, run:"
    log "  hash -r && source ~/.bashrc"
    log "Or simply open a new terminal window"
    log "Script completed in directory: $(pwd)"
else
    error "ripgrep installation verification failed - try restarting your shell or run 'source ~/.bashrc'"
fi