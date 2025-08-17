#!/bin/bash

# Starship Installation Script for Debian Linux
# Automated dependency installation, build, and install with smart Nerd Font handling
# Usage: ./install-starship.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
STARSHIP_DIR="starship"
STARSHIP_REPO="https://github.com/starship/starship.git"
RUST_MIN_VERSION="1.85.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="source"   # source, official
INSTALL_FONTS=false
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
SETUP_SHELL=true

# Show help
show_help() {
    cat << EOF
Starship Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native + LTO)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Installation Methods:
  --source             Build from source (default, follows gearbox patterns)
  --official           Use official installer (faster, prebuilt binary)

Font Options:
  --install-fonts      Automatically install FiraCode Nerd Font
  --skip-fonts         Skip Nerd Font detection and installation

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building (source builds only)
  --no-shell           Skip shell integration setup
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: source build with font detection
  $0 --official        # Use official installer (faster)
  $0 --install-fonts   # Force install Nerd Fonts
  $0 -d -c             # Debug build, config only
  $0 -o --run-tests    # Optimized build with tests
  $0 --skip-deps       # Skip dependency installation

About Starship:
  The minimal, blazing-fast, and infinitely customizable prompt for any shell.
  Works with bash, zsh, fish, PowerShell and more. Provides contextual information
  about your git repo, programming language, and system status.

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
        --official)
            INSTALL_METHOD="official"
            shift
            ;;
        --install-fonts)
            INSTALL_FONTS=true
            shift
            ;;
        --skip-fonts)
            INSTALL_FONTS=false
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
        --no-shell)
            SETUP_SHELL=false
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
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 0
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 1
        fi
    done
    return 0
}

# Check and optionally install Nerd Fonts
check_nerd_fonts() {
    log "Checking for Nerd Fonts..."
    
    # Check if fc-list is available
    if ! command -v fc-list &> /dev/null; then
        warning "fc-list not found, installing fontconfig..."
        sudo apt update &> /dev/null
        sudo apt install -y fontconfig &> /dev/null
    fi
    
    # Check for existing Nerd Fonts
    if fc-list | grep -i "nerd\|powerline" &> /dev/null; then
        success "Nerd Fonts detected - full visual experience available"
        return 0
    else
        warning "No Nerd Fonts detected"
        log "Starship will work but some icons may not display correctly"
        echo
        
        if [[ "$INSTALL_FONTS" == true ]] || ask_user "Install FiraCode Nerd Font for optimal experience?"; then
            install_firacode_nerd_font
        else
            log "Continuing with basic configuration"
            log "You can install Nerd Fonts later: https://www.nerdfonts.com/"
            echo
        fi
    fi
}

# Install FiraCode Nerd Font
install_firacode_nerd_font() {
    log "Installing FiraCode Nerd Font..."
    
    local font_dir="$HOME/.local/share/fonts"
    mkdir -p "$font_dir"
    
    # Download FiraCode Nerd Font variants
    local base_url="https://github.com/ryanoasis/nerd-fonts/raw/HEAD/patched-fonts/FiraCode"
    local fonts=(
        "Regular/FiraCodeNerdFont-Regular.ttf"
        "Bold/FiraCodeNerdFont-Bold.ttf"
        "Medium/FiraCodeNerdFont-Medium.ttf"
        "Light/FiraCodeNerdFont-Light.ttf"
    )
    
    for font in "${fonts[@]}"; do
        local font_name=$(basename "$font")
        local font_path="$font_dir/$font_name"
        
        if [[ ! -f "$font_path" ]]; then
            log "Downloading $font_name..."
            if curl -fLo "$font_path" "$base_url/$font" 2>/dev/null; then
                log "✓ Downloaded $font_name"
            else
                warning "Failed to download $font_name, skipping..."
            fi
        else
            log "✓ $font_name already exists"
        fi
    done
    
    # Refresh font cache
    log "Refreshing font cache..."
    fc-cache -fv &> /dev/null
    
    success "FiraCode Nerd Font installed!"
    echo
    warning "Please restart your terminal for font changes to take effect"
    echo
}

# Configure shell integration
configure_shell_integration() {
    if [[ "$SETUP_SHELL" != true ]]; then
        log "Skipping shell integration setup as requested"
        return 0
    fi
    
    log "Configuring shell integration..."
    
    # Detect current shell
    local current_shell=$(basename "$SHELL" 2>/dev/null || echo "bash")
    
    case $current_shell in
        bash)
            local config_file="$HOME/.bashrc"
            local init_line='eval "$(starship init bash)"'
            ;;
        zsh)
            local config_file="$HOME/.zshrc"
            local init_line='eval "$(starship init zsh)"'
            ;;
        fish)
            local config_file="$HOME/.config/fish/config.fish"
            local init_line='starship init fish | source'
            mkdir -p "$(dirname "$config_file")"
            ;;
        *)
            warning "Unsupported shell: $current_shell"
            log "Manual setup required. See: https://starship.rs/guide/#%F0%9F%9A%80-installation"
            return 0
            ;;
    esac
    
    # Check if already configured
    if [[ -f "$config_file" ]] && grep -q "starship init" "$config_file"; then
        log "✓ Starship already configured in $config_file"
        return 0
    fi
    
    # Add initialization line
    log "Adding Starship initialization to $config_file"
    echo "" >> "$config_file"
    echo "# Initialize Starship prompt" >> "$config_file"
    echo "$init_line" >> "$config_file"
    
    success "Shell integration configured for $current_shell"
    log "Restart your shell or run 'source $config_file' to activate"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "======================================="
log "Starship Installation Script"
log "======================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Install method: $INSTALL_METHOD"
log "Install fonts: $INSTALL_FONTS"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
log "Shell integration: $SETUP_SHELL"
echo

# Check if starship is already installed
if command -v starship &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(starship --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    
    # For official installer method, respect existing installation
    if [[ "$INSTALL_METHOD" == "official" ]]; then
        log "starship is already installed (version: $CURRENT_VERSION)"
        log "Use --force to reinstall"
        exit 0
    else
        # For source builds, inform but continue (gearbox philosophy: build latest from source)
        log "Found existing starship installation (version: $CURRENT_VERSION)"
        log "Building latest version from source (gearbox builds from latest main branch)"
    fi
fi

# Official installer method
if [[ "$INSTALL_METHOD" == "official" ]]; then
    log "Using official starship installer..."
    
    if [[ "$MODE" == "config" ]]; then
        log "Config mode: would download and install starship via official installer"
        success "Configuration completed (official installer ready)!"
        exit 0
    fi
    
    log "Downloading and installing starship via official installer..."
    curl -sS https://starship.rs/install.sh | sh -s -- --yes || error "Official starship installation failed"
    
    # Add to PATH if not already there
    export PATH="$HOME/.cargo/bin:$PATH"
    
    # Verify installation
    if command -v starship &> /dev/null; then
        INSTALLED_VERSION=$(starship --version 2>/dev/null | head -n1)
        success "starship installation completed successfully!"
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: Official installer"
        log "Binary location: $(which starship)"
        
        # Handle Nerd Fonts
        check_nerd_fonts
        
        # Configure shell integration
        configure_shell_integration
        
        echo
        log "Basic usage:"
        log "  Your prompt should now be enhanced with contextual information"
        log "  Configuration file: ~/.config/starship.toml (optional)"
        log "  For more information: https://starship.rs/config/"
    else
        error "starship installation verification failed"
    fi
    
    exit 0
fi

# Source build method continues below
log "Using source build method..."

# Install dependencies for source build
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for starship..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl fontconfig
    
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

# Get starship source directory
STARSHIP_SOURCE_DIR="$BUILD_DIR/$STARSHIP_DIR"

# Clone or update repository
if [[ ! -d "$STARSHIP_SOURCE_DIR" ]]; then
    log "Cloning starship repository..."
    git clone "$STARSHIP_REPO" "$STARSHIP_SOURCE_DIR"
else
    log "Updating starship repository..."
    cd "$STARSHIP_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/master
fi

cd "$STARSHIP_SOURCE_DIR"

# Configure build
log "Configuring starship build..."

# Verify we have Cargo.toml
if [[ ! -f "Cargo.toml" ]]; then
    error "Cargo.toml not found. This doesn't appear to be a valid Rust project."
fi

log "starship source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build
log "Building starship..."

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
        log "Building optimized version with target-cpu=native and LTO..."
        RUSTFLAGS="-C target-cpu=native -C lto=fat" cargo build --release
        TARGET_DIR="target/release"
        ;;
esac

# Verify build
if [[ ! -f "$TARGET_DIR/starship" ]]; then
    error "Build failed - starship binary not found in $TARGET_DIR"
fi

success "starship build completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Binary location: $STARSHIP_SOURCE_DIR/$TARGET_DIR/starship"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running starship tests..."
    if cargo test; then
        success "All tests passed!"
    else
        warning "Some tests failed, but continuing with installation"
    fi
fi

# Install
log "Installing starship..."

# Copy binary to /usr/local/bin
sudo cp "$TARGET_DIR/starship" /usr/local/bin/
sudo chmod +x /usr/local/bin/starship

# Verify installation
if command -v starship &> /dev/null; then
    INSTALLED_VERSION=$(starship --version 2>/dev/null | head -n1)
    success "starship installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Source build ($BUILD_TYPE)"
    log "Binary location: /usr/local/bin/starship"
    
    # Handle Nerd Fonts
    echo
    check_nerd_fonts
    
    # Configure shell integration
    echo
    configure_shell_integration
    
    # Show usage information
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
            log "Features: Maximum performance build with LTO and target-cpu optimizations"
            ;;
    esac
    echo
    log "Key features enabled:"
    log "  ✓ Contextual information (git, language, time, etc.)"
    log "  ✓ Support for 50+ shells and platforms"
    log "  ✓ Highly customizable with TOML configuration"
    log "  ✓ Fast and minimal - written in Rust"
    echo
    log "Configuration:"
    log "  Default config: ~/.config/starship.toml (auto-created if needed)"
    log "  Presets: starship preset bracketed-segments -o ~/.config/starship.toml"
    log "  Documentation: https://starship.rs/config/"
    echo
    log "For more information: starship --help"
else
    error "starship installation verification failed - starship not found in PATH"
fi

# Update library cache
sudo ldconfig

success "starship installation completed!"
log "Restart your shell or run 'source ~/.bashrc' (or your shell's config) to activate"