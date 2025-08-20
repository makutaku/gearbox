#!/bin/bash

# 7-Zip Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-7zip.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/lib/common.sh" ]]; then
    source "$REPO_DIR/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/lib/" >&2
    exit 1
fi


# Configuration
SEVENZIP_DIR="7zip"
SEVENZIP_REPO="https://github.com/ip7z/7zip.git"

# Default options
BUILD_TYPE="optimized"  # basic, optimized, asm-optimized
MODE="install"          # config, build, install
SKIP_DEPS=false

# Show help
show_help() {
    cat << EOF
7-Zip Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -b, --basic           Basic build (fastest, no optimizations)
  -o, --optimized       Optimized build (default, good performance)
  -a, --asm-optimized   Assembly optimized build (maximum performance)

Modes:
  -c, --config-only     Configure only (prepare build)
  -B, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  -h, --help           Show this help message

Examples:
  $0                   # Default: optimized build with install
  $0 -b -c             # Basic build, config only
  $0 -a -B             # Assembly optimized, build only
  $0 --skip-deps       # Skip dependency installation

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -b|--basic)
            BUILD_TYPE="basic"
            shift
            ;;
        -o|--optimized)
            BUILD_TYPE="optimized"
            shift
            ;;
        -a|--asm-optimized)
            BUILD_TYPE="asm-optimized"
            shift
            ;;
        -c|--config-only)
            MODE="config"
            shift
            ;;
        -B|--build-only)
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

# Get makefile and build options based on build type
get_build_config() {
    case $BUILD_TYPE in
        basic)
            echo "makefile.gcc"
            ;;
        optimized)
            echo "makefile.gcc"
            ;;
        asm-optimized)
            # Check architecture to determine appropriate makefile
            ARCH=$(uname -m)
            case $ARCH in
                x86_64)
                    echo "../../cmpl_gcc_x64.mak"
                    ;;
                aarch64)
                    echo "../../cmpl_gcc_arm64.mak"
                    ;;
                *)
                    warning "Assembly optimizations not available for $ARCH, falling back to optimized build"
                    echo "makefile.gcc"
                    ;;
            esac
            ;;
    esac
}

# Get build variables based on build type
get_build_vars() {
    case $BUILD_TYPE in
        basic)
            echo ""
            ;;
        optimized)
            echo "IS_X64=1"
            ;;
        asm-optimized)
            echo "IS_X64=1 USE_ASM=1"
            ;;
    esac
}

# Install dependencies based on build type
install_dependencies() {
    if [[ "$SKIP_DEPS" == true ]]; then
        log "Skipping dependency installation as requested"
        return 0
    fi

    # Update package list
    log "Updating package list..."
    sudo apt update || error "Failed to update package list"

    # Install build essentials (always needed)
    log "Installing build tools..."
    sudo apt install -y \
        build-essential \
        gcc \
        g++ \
        make \
        git \
        || error "Failed to install build tools"

    # Install additional tools based on build type
    case $BUILD_TYPE in
        basic|optimized)
            log "Installing basic build dependencies..."
            # No additional tools needed
            ;;
        asm-optimized)
            log "Installing assembly optimization dependencies..."
            # For assembly builds, we might need additional tools
            # Most assembly code in 7-Zip is optional and falls back to C++ if not available
            warning "Assembly optimizations may require specific assemblers - build will fall back to C++ if unavailable"
            ;;
    esac

    success "Dependencies installed successfully"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting 7-Zip $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"

# Handle 7-Zip source code
if [[ -d "$SEVENZIP_DIR" ]]; then
    log "Found existing 7-Zip directory: $SEVENZIP_DIR"
    
    # Check if it's a git repository
    if [[ -d "$SEVENZIP_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$SEVENZIP_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$SEVENZIP_REPO" ]]; then
            log "Repository origin matches expected 7-Zip repository"
            
            if ask_user "Do you want to pull the latest changes from the 7-Zip repository?"; then
                log "Pulling latest changes..."
                git pull origin main || error "Failed to pull latest changes"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh 7-Zip repository?"; then
                cd ..
                rm -rf "$SEVENZIP_DIR"
                log "Cloning 7-Zip repository..."
                git clone "$SEVENZIP_REPO" "$SEVENZIP_DIR" || error "Failed to clone 7-Zip repository"
                cd "$SEVENZIP_DIR"
                success "7-Zip repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh 7-Zip repository?"; then
            rm -rf "$SEVENZIP_DIR"
            log "Cloning 7-Zip repository..."
            git clone "$SEVENZIP_REPO" "$SEVENZIP_DIR" || error "Failed to clone 7-Zip repository"
            success "7-Zip repository cloned successfully"
        else
            error "Cannot proceed without a proper 7-Zip source directory"
        fi
    fi
else
    log "Cloning 7-Zip repository..."
    git clone "$SEVENZIP_REPO" "$SEVENZIP_DIR" || error "Failed to clone 7-Zip repository"
    success "7-Zip repository cloned successfully"
fi

# Change to 7-Zip build directory
BUILD_DIR="$SEVENZIP_DIR/CPP/7zip/Bundles/Alone2"
if [[ ! -d "$BUILD_DIR" ]]; then
    error "Invalid 7-Zip source directory - missing build directory: $BUILD_DIR"
fi

cd "$BUILD_DIR"

# Verify we're in the correct directory
if [[ ! -f "makefile.gcc" ]]; then
    error "Invalid 7-Zip build directory - missing makefile.gcc"
fi

# Install dependencies
install_dependencies

# Clean previous build
log "Cleaning previous build files..."
if [[ -f "_o/7zz" ]]; then
    rm -f _o/7zz || warning "Failed to clean previous build, continuing..."
fi

# Get build configuration
MAKEFILE=$(get_build_config)
BUILD_VARS=$(get_build_vars)

log "Configuring 7-Zip with $BUILD_TYPE settings..."
log "Using makefile: $MAKEFILE"
if [[ -n "$BUILD_VARS" ]]; then
    log "Build variables: $BUILD_VARS"
fi

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    success "Build command would be: make -j$(get_optimal_jobs) -f $MAKEFILE $BUILD_VARS"
    exit 0
fi

# Build 7-Zip
log "Building 7-Zip (this may take a while)..."
CORES=$(get_optimal_jobs)
log "Using $CORES CPU cores for parallel build"

if [[ -n "$BUILD_VARS" ]]; then
    build_with_options make "-j$CORES -f $MAKEFILE" "$BUILD_VARS"
else
    execute_command_safely make -j$CORES -f "$MAKEFILE"
fi

# Verify build output
if [[ ! -f "_o/7zz" ]]; then
    error "Build completed but 7zz executable not found"
fi

success "Build completed successfully"

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install or manually copy 7zz to install."
    log "Build output: $(pwd)/_o/7zz"
    exit 0
fi

# Install 7-Zip
log "Installing 7-Zip..."

# Install the binary
sudo cp _o/7zz /usr/local/bin/ || error "Installation failed"
sudo chmod +x /usr/local/bin/7zz || error "Failed to set executable permissions"

# Verify installation
log "Verifying installation..."
if command -v 7zz &> /dev/null; then
    success "7-Zip installation verified!"
    echo
    log "7-Zip version information:"
    7zz | head -n 3
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        basic)
            log "Features: Basic compression/decompression functionality"
            ;;
        optimized)
            log "Features: Optimized performance with architecture-specific improvements"
            ;;
        asm-optimized)
            log "Features: Maximum performance with assembly optimizations"
            ;;
    esac
    echo
    success "7-Zip installation completed successfully!"
    log "You can now use the '7zz' command for archive operations"
    echo
    log "Usage examples:"
    log "  7zz a archive.7z files/       # Create archive"
    log "  7zz x archive.7z              # Extract archive"
    log "  7zz l archive.7z              # List archive contents"
    echo
    log "Script completed in directory: $(pwd)"
else
    error "7-Zip installation verification failed"
fi
