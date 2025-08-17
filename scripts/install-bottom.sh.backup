#!/bin/bash

# bottom Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-bottom.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BOTTOM_DIR="bottom"
BOTTOM_REPO="https://github.com/ClementTsang/bottom.git"
RUST_MIN_VERSION="1.81.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
ENABLE_BATTERY=true
ENABLE_GPU=true
ENABLE_ZFS=true

# Show help
show_help() {
    cat << EOF
bottom Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Feature Options:
  --no-battery          Disable battery monitoring support
  --no-gpu              Disable GPU monitoring support
  --no-zfs              Disable ZFS filesystem support
  --minimal             Disable all optional features (battery, GPU, ZFS)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: release build with all features
  $0 -d -c             # Debug build, config only
  $0 -o --run-tests    # Optimized build with tests
  $0 --minimal         # Minimal build without optional features
  $0 --no-gpu          # Build without GPU monitoring
  $0 --skip-deps       # Skip dependency installation

About bottom:
  A cross-platform graphical process/system monitor for the terminal.
  Provides beautiful visualizations of CPU, memory, network, disk usage,
  temperatures, and running processes. Features customizable layouts,
  process tree view, and interactive process management.

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
        --no-battery)
            ENABLE_BATTERY=false
            shift
            ;;
        --no-gpu)
            ENABLE_GPU=false
            shift
            ;;
        --no-zfs)
            ENABLE_ZFS=false
            shift
            ;;
        --minimal)
            ENABLE_BATTERY=false
            ENABLE_GPU=false
            ENABLE_ZFS=false
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

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "======================================="
log "bottom Installation Script"
log "======================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
log "Features: Battery=$ENABLE_BATTERY, GPU=$ENABLE_GPU, ZFS=$ENABLE_ZFS"
echo

# Check if bottom is already installed (binary name is 'btm')
if command -v btm &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(btm --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    log "Found existing bottom installation (version: $CURRENT_VERSION)"
    log "Building latest version from source (gearbox builds from latest main branch)"
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for bottom..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl pkg-config
    
    # Install system monitoring dependencies
    log "Installing system monitoring libraries..."
    sudo apt install -y libssl-dev
    
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

# Get bottom source directory
BOTTOM_SOURCE_DIR="$BUILD_DIR/$BOTTOM_DIR"

# Clone or update repository
if [[ ! -d "$BOTTOM_SOURCE_DIR" ]]; then
    log "Cloning bottom repository..."
    git clone "$BOTTOM_REPO" "$BOTTOM_SOURCE_DIR"
else
    log "Updating bottom repository..."
    cd "$BOTTOM_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/master
fi

cd "$BOTTOM_SOURCE_DIR"

# Configure build
log "Configuring bottom build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

# Prepare build features
BUILD_FEATURES=""
if [[ "$ENABLE_BATTERY" == true ]] || [[ "$ENABLE_GPU" == true ]] || [[ "$ENABLE_ZFS" == true ]]; then
    FEATURE_LIST=()
    if [[ "$ENABLE_BATTERY" == true ]]; then
        FEATURE_LIST+=("battery")
    fi
    if [[ "$ENABLE_GPU" == true ]]; then
        FEATURE_LIST+=("gpu")
    fi
    if [[ "$ENABLE_ZFS" == true ]]; then
        FEATURE_LIST+=("zfs")
    fi
    
    if [[ ${#FEATURE_LIST[@]} -gt 0 ]]; then
        # Use --no-default-features and specify only desired features
        BUILD_FEATURES="--no-default-features --features $(IFS=,; echo "${FEATURE_LIST[*]}")"
    fi
else
    # Minimal build - no optional features
    BUILD_FEATURES="--no-default-features"
fi

log "bottom source configured successfully"
log "Build features: ${BUILD_FEATURES:-default}"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building bottom..."

case $BUILD_TYPE in
    debug)
        log "Building debug version..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build $BUILD_FEATURES
        else
            cargo build
        fi
        TARGET_DIR="target/debug"
        ;;
    release)
        log "Building release version..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            cargo build --release $BUILD_FEATURES
        else
            cargo build --release
        fi
        TARGET_DIR="target/release"
        ;;
    optimized)
        log "Building optimized version with target-cpu=native..."
        if [[ -n "$BUILD_FEATURES" ]]; then
            RUSTFLAGS="-C target-cpu=native" cargo build --release $BUILD_FEATURES
        else
            RUSTFLAGS="-C target-cpu=native" cargo build --release
        fi
        TARGET_DIR="target/release"
        ;;
esac

# Verify build (binary is named 'btm')
if [[ ! -f "$TARGET_DIR/btm" ]]; then
    error "Build failed - btm binary not found in $TARGET_DIR"
fi

success "bottom build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $BOTTOM_SOURCE_DIR/$TARGET_DIR/btm"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running bottom tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing bottom..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/btm" /usr/local/bin/
sudo chmod +x /usr/local/bin/btm

# Verify installation
if command -v btm &> /dev/null; then
    INSTALLED_VERSION=$(btm --version 2>/dev/null | head -n1)
    success "bottom installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/btm"
    
    # Show usage information
    echo
    log "Basic usage examples:"
    log "  btm                              # Start bottom with default view"
    log "  btm --basic                      # htop-like basic mode"
    log "  btm --tree                       # Show processes in tree mode"
    log "  btm --expanded                   # Start with expanded view"
    log "  btm --time_delta 500             # 500ms refresh rate"
    echo
    log "Advanced options:"
    log "  btm --battery                    # Show battery widget"
    log "  btm --celsius                    # Temperature in Celsius"
    log "  btm --fahrenheit                 # Temperature in Fahrenheit"
    log "  btm --rate 1000                  # 1-second refresh rate"
    log "  btm --group                      # Group processes by name"
    echo
    log "Interactive controls (while running):"
    log "  ?/F1     Show help and keybindings"
    log "  q        Quit bottom"
    log "  /        Search processes"
    log "  dd       Kill selected process"
    log "  Tab      Switch between widgets"
    log "  +/-      Zoom time graphs in/out"
    log "  t        Toggle tree mode"
    log "  s        Sort processes"
    echo
    log "Configuration:"
    log "  btm --generate_config            # Generate default config file"
    log "  Config location: ~/.config/bottom/bottom.toml"
    log "  Themes and layouts can be customized"
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
    log "Enabled features:"
    if [[ "$ENABLE_BATTERY" == true ]]; then
        log "  ✓ Battery monitoring"
    else
        log "  ✗ Battery monitoring disabled"
    fi
    if [[ "$ENABLE_GPU" == true ]]; then
        log "  ✓ GPU monitoring"
    else
        log "  ✗ GPU monitoring disabled"
    fi
    if [[ "$ENABLE_ZFS" == true ]]; then
        log "  ✓ ZFS filesystem support"
    else
        log "  ✗ ZFS filesystem support disabled"
    fi
    echo
    log "Key system monitoring features:"
    log "  ✓ CPU usage with per-core display"
    log "  ✓ Memory and swap monitoring"
    log "  ✓ Network I/O tracking"
    log "  ✓ Disk capacity and I/O"
    log "  ✓ Process management with tree view"
    log "  ✓ Temperature sensors (if available)"
    echo
    log "Tips:"
    log "  • Use arrow keys or hjkl for navigation"
    log "  • Press '?' for complete keybinding help"
    log "  • Customize with ~/.config/bottom/bottom.toml"
    log "  • Use --help for all command-line options"
    log "For more information: https://github.com/ClementTsang/bottom"
else
    error "bottom installation verification failed - btm not found in PATH"
fi

# Update library cache
sudo ldconfig

success "bottom installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"