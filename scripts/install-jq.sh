#!/bin/bash

# jq Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-jq.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
JQ_DIR="jq"
JQ_REPO="https://github.com/jqlang/jq.git"

# Default options
BUILD_TYPE="standard"  # minimal, standard, static, optimized
MODE="install"         # config, build, install
SKIP_DEPS=false
RUN_TESTS=false

# Show help
show_help() {
    cat << EOF
jq Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -m, --minimal         Minimal build (fastest, from tarball if available)
  -s, --standard        Standard build (default, good functionality)
  -S, --static          Static build (self-contained executable)
  -o, --optimized       Optimized build (maximum performance)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  -h, --help           Show this help message

Examples:
  $0                   # Default: standard build with install
  $0 -m -c             # Minimal build, config only
  $0 -S -b             # Static build, build only
  $0 --skip-deps --run-tests  # Skip deps, run tests

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--minimal)
            BUILD_TYPE="minimal"
            shift
            ;;
        -s|--standard)
            BUILD_TYPE="standard"
            shift
            ;;
        -S|--static)
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

# Get configuration options based on build type
get_configure_options() {
    local base_options="--prefix=/usr/local --with-oniguruma=builtin"
    
    case $BUILD_TYPE in
        minimal)
            echo "$base_options --disable-maintainer-mode --disable-docs"
            ;;
        standard)
            echo "$base_options"
            ;;
        static)
            echo "$base_options --enable-all-static"
            ;;
        optimized)
            echo "$base_options --disable-docs"
            ;;
    esac
}

# Get make options based on build type
get_make_options() {
    case $BUILD_TYPE in
        minimal|standard|optimized)
            echo ""
            ;;
        static)
            echo "LDFLAGS=-all-static"
            ;;
    esac
}

# Get CFLAGS based on build type
get_cflags() {
    case $BUILD_TYPE in
        minimal|standard|static)
            echo ""
            ;;
        optimized)
            echo "CFLAGS=\"-O2 -pthread -fstack-protector-all\""
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

    # Install autotools and build essentials (always needed)
    log "Installing build tools and autotools..."
    sudo apt install -y \
        autoconf \
        automake \
        build-essential \
        libtool \
        git \
        || error "Failed to install build tools"

    # Install additional tools for building from git
    log "Installing additional build dependencies..."
    sudo apt install -y \
        bison \
        flex \
        libonig-dev \
        python3-dev \
        || warning "Some optional dependencies may not be available"

    success "Dependencies installed successfully"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting jq $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"

# Handle jq source code
if [[ -d "$JQ_DIR" ]]; then
    log "Found existing jq directory: $JQ_DIR"
    
    # Check if it's a git repository
    if [[ -d "$JQ_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$JQ_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$JQ_REPO" ]]; then
            log "Repository origin matches expected jq repository"
            
            if ask_user "Do you want to pull the latest changes from the jq repository?"; then
                log "Pulling latest changes..."
                git pull origin master || error "Failed to pull latest changes"
                log "Updating submodules..."
                git submodule update --init --recursive || error "Failed to update submodules"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh jq repository?"; then
                cd ..
                rm -rf "$JQ_DIR"
                log "Cloning jq repository..."
                git clone --recursive "$JQ_REPO" "$JQ_DIR" || error "Failed to clone jq repository"
                cd "$JQ_DIR"
                success "jq repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh jq repository?"; then
            rm -rf "$JQ_DIR"
            log "Cloning jq repository..."
            git clone --recursive "$JQ_REPO" "$JQ_DIR" || error "Failed to clone jq repository"
            success "jq repository cloned successfully"
        else
            error "Cannot proceed without a proper jq source directory"
        fi
    fi
else
    log "Cloning jq repository..."
    git clone --recursive "$JQ_REPO" "$JQ_DIR" || error "Failed to clone jq repository"
    success "jq repository cloned successfully"
fi

# Change to jq directory
cd "$JQ_DIR"

# Verify we're in the correct directory
if [[ ! -f "configure.ac" ]] && [[ ! -f "configure" ]]; then
    error "Invalid jq source directory - missing configure.ac or configure"
fi

# Install dependencies
install_dependencies

# Generate configure script if building from git
if [[ -f "configure.ac" ]] && [[ ! -f "configure" ]]; then
    log "Generating configure script from git repository..."
    autoreconf -i || error "Failed to generate configure script"
    success "Configure script generated successfully"
fi

# Clean previous build
log "Cleaning previous build files..."
if [[ -f "Makefile" ]]; then
    make clean || warning "Failed to clean previous build, continuing..."
fi

# Get build configuration
CONFIGURE_OPTIONS=$(get_configure_options)
MAKE_OPTIONS=$(get_make_options)
CFLAGS_OPTIONS=$(get_cflags)

log "Configuring jq with $BUILD_TYPE settings..."
log "Configure options: $CONFIGURE_OPTIONS"
if [[ -n "$CFLAGS_OPTIONS" ]]; then
    log "CFLAGS: $CFLAGS_OPTIONS"
fi
if [[ -n "$MAKE_OPTIONS" ]]; then
    log "Make options: $MAKE_OPTIONS"
fi

# Run configure with appropriate options
if [[ -n "$CFLAGS_OPTIONS" ]]; then
    eval "$CFLAGS_OPTIONS ./configure $CONFIGURE_OPTIONS" || error "Configuration failed"
else
    eval "./configure $CONFIGURE_OPTIONS" || error "Configuration failed"
fi

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    if [[ -n "$MAKE_OPTIONS" ]]; then
        success "Build command would be: make -j$(nproc) $MAKE_OPTIONS"
    else
        success "Build command would be: make -j$(nproc)"
    fi
    exit 0
fi

# Build jq
log "Building jq (this may take a while)..."
NPROC=$(nproc)
log "Using $NPROC CPU cores for parallel build"

if [[ -n "$MAKE_OPTIONS" ]]; then
    eval "make -j$NPROC $MAKE_OPTIONS" || error "Build failed"
else
    make -j$NPROC || error "Build failed"
fi

# Verify build output
if [[ ! -f "jq" ]]; then
    error "Build completed but jq executable not found"
fi

success "Build completed successfully"

# Run tests if requested
if [[ "$RUN_TESTS" == true ]]; then
    log "Running test suite..."
    make check || warning "Some tests failed, but continuing with installation"
    success "Test suite completed"
fi

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install or 'sudo make install' to install."
    log "Build output: $(pwd)/jq"
    exit 0
fi

# Install jq
log "Installing jq..."
sudo make install || error "Installation failed"

# Update library cache
log "Updating library cache..."
echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/jq.conf > /dev/null
sudo ldconfig || error "Failed to update library cache"

# Verify installation
log "Verifying installation..."
if command -v jq &> /dev/null; then
    success "jq installation verified!"
    echo
    log "jq version information:"
    jq --version
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        minimal)
            log "Features: Basic JSON processing functionality"
            ;;
        standard)
            log "Features: Full JSON processing with all standard features"
            ;;
        static)
            log "Features: Self-contained static executable with all features"
            ;;
        optimized)
            log "Features: Performance-optimized build with all features"
            ;;
    esac
    echo
    success "jq installation completed successfully!"
    log "You can now use the 'jq' command for JSON processing"
    echo
    log "Usage examples:"
    log "  echo '{\"name\":\"John\"}' | jq '.name'     # Extract field"
    log "  cat data.json | jq '.[] | select(.age > 30)'  # Filter array"
    log "  jq -r '.items[].name' file.json              # Raw string output"
    echo
    log "Script completed in directory: $(pwd)"
else
    error "jq installation verification failed"
fi