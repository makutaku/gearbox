#!/bin/bash

# ImageMagick Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-imagemagick.sh [OPTIONS]

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
IMAGEMAGICK_DIR="ImageMagick"
IMAGEMAGICK_REPO="https://github.com/ImageMagick/ImageMagick.git"

# Default options
BUILD_TYPE="full-featured"  # minimal, full-featured, professional
MODE="install"              # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
ImageMagick Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -m, --minimal         Minimal build (Q8, basic formats, fastest)
  -f, --full-featured   Full-featured build (default, Q16, comprehensive formats)
  -p, --professional    Professional build (Q16 HDRI, all features, highest quality)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (may take a while)
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: full-featured build with install
  $0 -m -c             # Minimal build, config only
  $0 -p --run-tests    # Professional build with tests
  $0 --skip-deps       # Skip dependency installation

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--minimal)
            BUILD_TYPE="minimal"
            shift
            ;;
        -f|--full-featured)
            BUILD_TYPE="full-featured"
            shift
            ;;
        -p|--professional)
            BUILD_TYPE="professional"
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

# Get configuration options based on build type
get_configure_options() {
    local base_options="--prefix=/usr/local --enable-shared --disable-static --with-modules --enable-openmp --disable-dependency-tracking"
    
    case $BUILD_TYPE in
        minimal)
            echo "$base_options --with-quantum-depth=8 --disable-hdri --without-magick-plus-plus --disable-docs"
            ;;
        full-featured)
            echo "$base_options --with-quantum-depth=16 --disable-hdri"
            ;;
        professional)
            echo "$base_options --with-quantum-depth=16 --enable-hdri --enable-64bit-channel-masks"
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
        pkg-config \
        autoconf \
        automake \
        libtool \
        git \
        wget \
        || error "Failed to install build tools"

    # Install core delegate libraries (always needed)
    log "Installing core image format libraries..."
    sudo apt install -y \
        libjpeg-dev \
        libpng-dev \
        libtiff-dev \
        zlib1g-dev \
        libfreetype6-dev \
        libfontconfig1-dev \
        || error "Failed to install core libraries"

    # Install additional libraries based on build type
    case $BUILD_TYPE in
        minimal)
            log "Installing minimal additional dependencies..."
            # Core libraries already installed
            ;;
        full-featured|professional)
            log "Installing comprehensive image format libraries..."
            sudo apt install -y \
                libwebp-dev \
                libopenjp2-7-dev \
                librsvg2-dev \
                libxml2-dev \
                liblcms2-dev \
                libde265-dev \
                libheif-dev \
                libopenexr-dev \
                libraw-dev \
                libdjvulibre-dev \
                || warning "Some optional libraries may not be available"
            ;;
    esac

    # Install testing dependencies if requested
    if [[ "$RUN_TESTS" == true ]]; then
        log "Installing testing dependencies..."
        sudo apt install -y \
            ghostscript \
            fonts-dejavu-core \
            || warning "Some test dependencies may not be available"
    fi

    success "Dependencies installed successfully"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting ImageMagick $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"

# Handle ImageMagick source code
if [[ -d "$IMAGEMAGICK_DIR" ]]; then
    log "Found existing ImageMagick directory: $IMAGEMAGICK_DIR"
    
    # Check if it's a git repository
    if [[ -d "$IMAGEMAGICK_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$IMAGEMAGICK_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$IMAGEMAGICK_REPO" ]]; then
            log "Repository origin matches expected ImageMagick repository"
            
            if ask_user "Do you want to pull the latest changes from the ImageMagick repository?"; then
                log "Pulling latest changes..."
                git pull origin main || error "Failed to pull latest changes"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh ImageMagick repository?"; then
                cd ..
                rm -rf "$IMAGEMAGICK_DIR"
                log "Cloning ImageMagick repository..."
                git clone "$IMAGEMAGICK_REPO" "$IMAGEMAGICK_DIR" || error "Failed to clone ImageMagick repository"
                cd "$IMAGEMAGICK_DIR"
                success "ImageMagick repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh ImageMagick repository?"; then
            rm -rf "$IMAGEMAGICK_DIR"
            log "Cloning ImageMagick repository..."
            git clone "$IMAGEMAGICK_REPO" "$IMAGEMAGICK_DIR" || error "Failed to clone ImageMagick repository"
            success "ImageMagick repository cloned successfully"
        else
            error "Cannot proceed without a proper ImageMagick source directory"
        fi
    fi
else
    log "Cloning ImageMagick repository..."
    git clone "$IMAGEMAGICK_REPO" "$IMAGEMAGICK_DIR" || error "Failed to clone ImageMagick repository"
    success "ImageMagick repository cloned successfully"
fi

# Change to ImageMagick directory
cd "$IMAGEMAGICK_DIR"

# Verify we're in the correct directory
if [[ ! -f "configure.ac" ]] && [[ ! -f "configure" ]]; then
    error "Invalid ImageMagick source directory - missing configure.ac or configure"
fi

# Install dependencies
install_dependencies

# Generate configure script if building from git
if [[ -f "configure.ac" ]] && [[ ! -f "configure" ]]; then
    log "Generating configure script from git repository..."
    autoreconf -fiv || error "Failed to generate configure script"
    success "Configure script generated successfully"
fi

# Clean previous build
log "Cleaning previous build files..."
if [[ -f "Makefile" ]]; then
    make distclean || warning "Failed to clean previous build, continuing..."
fi

# Get build configuration
CONFIGURE_OPTIONS=$(get_configure_options)

log "Configuring ImageMagick with $BUILD_TYPE settings..."
log "Configure options: $CONFIGURE_OPTIONS"

# Run configure
configure_with_options "./configure" "$CONFIGURE_OPTIONS"

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    success "Build command would be: make -j$(get_optimal_jobs)"
    exit 0
fi

# Build ImageMagick
log "Building ImageMagick (this may take a while)..."
CORES=$(get_optimal_jobs)
log "Using $CORES CPU cores for parallel build"

make -j$CORES || error "Build failed"

# Verify build output
if [[ ! -f "utilities/magick" ]] && [[ ! -f "utilities/.libs/magick" ]]; then
    error "Build completed but magick executable not found"
fi

success "Build completed successfully"

# Run tests if requested
if [[ "$RUN_TESTS" == true ]]; then
    log "Running test suite (this may take a very long time)..."
    make check || warning "Some tests failed, but continuing with installation"
    success "Test suite completed"
fi

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install or 'sudo make install' to install."
    log "Build output in: $(pwd)/utilities/"
    exit 0
fi

# Install ImageMagick
log "Installing ImageMagick..."
sudo make install || error "Installation failed"

# Update library cache
log "Updating library cache..."
echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/imagemagick.conf > /dev/null
sudo ldconfig || error "Failed to update library cache"

# Verify installation
log "Verifying installation..."
if command -v magick &> /dev/null; then
    success "ImageMagick installation verified!"
    echo
    log "ImageMagick version information:"
    magick -version | head -n 3
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        minimal)
            log "Features: Q8 build optimized for speed and memory efficiency"
            ;;
        full-featured)
            log "Features: Q16 build with comprehensive format support"
            ;;
        professional)
            log "Features: Q16 HDRI build with maximum quality and precision"
            ;;
    esac
    echo
    log "Supported delegate libraries:"
    magick identify -list delegate | head -n 10
    echo
    log "Supported image formats:"
    magick identify -list format | wc -l | xargs echo "Total formats supported:"
    echo
    success "ImageMagick installation completed successfully!"
    log "You can now use ImageMagick commands: magick, identify, convert, etc."
    echo
    log "Usage examples:"
    log "  magick input.jpg -resize 50% output.jpg    # Resize image"
    log "  magick identify image.jpg                   # Get image info"
    log "  magick convert input.png output.jpg        # Convert format"
    log "  magick input.jpg -quality 85 output.jpg    # Adjust quality"
    echo
    log "Installation paths:"
    log "  magick: $(which magick)"
    log "  identify: $(which identify 2>/dev/null || echo 'not found')"
    echo
    log "Configuration files: /usr/local/etc/ImageMagick-7/"
    log "Script completed in directory: $(pwd)"
else
    error "ImageMagick installation verification failed"
fi
