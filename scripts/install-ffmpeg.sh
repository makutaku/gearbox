#!/bin/bash

# FFmpeg Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-ffmpeg.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FFMPEG_DIR="ffmpeg"
FFMPEG_REPO="https://git.ffmpeg.org/ffmpeg.git"

# Default options
BUILD_TYPE="general"  # minimal, general, maximum
MODE="install"        # config, build, install
SKIP_DEPS=false

# Show help
show_help() {
    cat << EOF
FFmpeg Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -m, --minimal         Minimal build (fastest, basic codecs only)
  -g, --general         General purpose build (default, good codec support)
  -x, --maximum         Maximum features build (all available codecs)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  -h, --help           Show this help message

Examples:
  $0                   # Default: general purpose build with install
  $0 -m -c             # Minimal build, config only
  $0 -x -b             # Maximum features, build only
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
        -g|--general)
            BUILD_TYPE="general"
            shift
            ;;
        -x|--maximum)
            BUILD_TYPE="maximum"
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

# Get configuration based on build type
get_configure_options() {
    local base_options="--prefix=/usr/local"
    
    case $BUILD_TYPE in
        minimal)
            echo "$base_options --enable-shared --enable-small --disable-debug --disable-doc"
            ;;
        general)
            echo "$base_options --enable-gpl --enable-version3 --enable-shared --enable-static --enable-small --enable-libx264 --enable-libx265 --enable-libvpx --enable-libmp3lame --enable-libopus --enable-libvorbis --enable-libass --enable-libfreetype --enable-libfontconfig --enable-libfribidi --enable-openssl"
            ;;
        maximum)
            echo "$base_options --enable-gpl --enable-version3 --enable-nonfree --enable-shared --enable-static --enable-libx264 --enable-libx265 --enable-libvpx --enable-libfdk-aac --enable-libmp3lame --enable-libopus --enable-libvorbis --enable-libtheora --enable-libass --enable-libfreetype --enable-libfontconfig --enable-libfribidi --enable-libharfbuzz --enable-openssl --enable-libxml2"
            ;;
    esac
}

# Get dependencies based on build type
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
        nasm \
        yasm \
        pkg-config \
        cmake \
        git \
        || error "Failed to install build tools"

    # Install codec libraries based on build type
    case $BUILD_TYPE in
        minimal)
            log "Installing minimal dependencies..."
            # No additional codec libraries for minimal build
            ;;
        general)
            log "Installing general purpose codec libraries..."
            sudo apt install -y \
                libass-dev \
                libx264-dev \
                libx265-dev \
                libvpx-dev \
                libmp3lame-dev \
                libopus-dev \
                libvorbis-dev \
                libfreetype6-dev \
                libfontconfig1-dev \
                libfribidi-dev \
                libssl-dev \
                libbz2-dev \
                libz-dev \
                || error "Failed to install codec libraries"
            ;;
        maximum)
            log "Installing maximum feature codec libraries..."
            sudo apt install -y \
                libass-dev \
                libx264-dev \
                libx265-dev \
                libvpx-dev \
                libfdk-aac-dev \
                libmp3lame-dev \
                libopus-dev \
                libvorbis-dev \
                libtheora-dev \
                libfreetype6-dev \
                libfontconfig1-dev \
                libfribidi-dev \
                libharfbuzz-dev \
                libxml2-dev \
                libssl-dev \
                libbz2-dev \
                libz-dev \
                || warning "Some maximum feature libraries may not be available"
            ;;
    esac

    success "Dependencies installed successfully"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting FFmpeg $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"

# Handle FFmpeg source code
if [[ -d "$FFMPEG_DIR" ]]; then
    log "Found existing FFmpeg directory: $FFMPEG_DIR"
    
    # Check if it's a git repository
    if [[ -d "$FFMPEG_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$FFMPEG_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$FFMPEG_REPO" ]]; then
            log "Repository origin matches expected FFmpeg repository"
            
            if ask_user "Do you want to pull the latest changes from the FFmpeg repository?"; then
                log "Pulling latest changes..."
                git pull origin master || error "Failed to pull latest changes"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh FFmpeg repository?"; then
                cd ..
                rm -rf "$FFMPEG_DIR"
                log "Cloning FFmpeg repository..."
                git clone "$FFMPEG_REPO" "$FFMPEG_DIR" || error "Failed to clone FFmpeg repository"
                cd "$FFMPEG_DIR"
                success "FFmpeg repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh FFmpeg repository?"; then
            rm -rf "$FFMPEG_DIR"
            log "Cloning FFmpeg repository..."
            git clone "$FFMPEG_REPO" "$FFMPEG_DIR" || error "Failed to clone FFmpeg repository"
            success "FFmpeg repository cloned successfully"
        else
            error "Cannot proceed without a proper FFmpeg source directory"
        fi
    fi
else
    log "Cloning FFmpeg repository..."
    git clone "$FFMPEG_REPO" "$FFMPEG_DIR" || error "Failed to clone FFmpeg repository"
    success "FFmpeg repository cloned successfully"
fi

# Change to FFmpeg directory
cd "$FFMPEG_DIR"

# Verify we're in the correct directory
if [[ ! -f "configure" ]] || [[ ! -f "INSTALL.md" ]]; then
    error "Invalid FFmpeg source directory - missing required files"
fi

# Install dependencies
install_dependencies

# Clean previous build
log "Cleaning previous build files..."
if [[ -f "Makefile" ]]; then
    make clean || warning "Failed to clean, continuing..."
fi

# Configure FFmpeg
log "Configuring FFmpeg with $BUILD_TYPE settings..."
CONFIGURE_OPTIONS=$(get_configure_options)
log "Configure options: $CONFIGURE_OPTIONS"

eval "./configure $CONFIGURE_OPTIONS" || error "Configuration failed"

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    exit 0
fi

# Build FFmpeg
log "Building FFmpeg (this may take a while)..."
NPROC=$(nproc)
log "Using $NPROC CPU cores for parallel build"

make -j$NPROC || error "Build failed"

success "Build completed successfully"

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install or 'sudo make install' to install."
    exit 0
fi

# Install FFmpeg
log "Installing FFmpeg..."
sudo make install || error "Installation failed"

# Update library cache
log "Updating library cache..."
echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/ffmpeg.conf > /dev/null
sudo ldconfig || error "Failed to update library cache"

# Verify installation
log "Verifying installation..."
if command -v ffmpeg &> /dev/null; then
    success "FFmpeg installation verified!"
    echo
    log "FFmpeg version information:"
    ffmpeg -version | head -n 1
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        minimal)
            log "Available: Basic codecs and formats"
            ;;
        general)
            log "Available codecs include: H.264, H.265, VP8/VP9, MP3, Opus, Vorbis, and more"
            ;;
        maximum)
            log "Available: All supported codecs including FDK-AAC, Theora, and advanced features"
            ;;
    esac
    echo
    success "FFmpeg installation completed successfully!"
    log "You can now use 'ffmpeg', 'ffprobe', and 'ffplay' commands"
    echo
    log "Script completed in directory: $(pwd)"
else
    error "FFmpeg installation verification failed"
fi