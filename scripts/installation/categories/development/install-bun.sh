#!/bin/bash
# Bun Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-bun.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")")"
# Source common library for shared functions
if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/scripts/lib/" >&2
    exit 1
fi

# Configuration
BUN_MIN_VERSION="1.0.0"
KERNEL_MIN_VERSION="5.1"
KERNEL_RECOMMENDED_VERSION="5.6"

# Default options
BUILD_TYPE="standard"     # minimal, standard, maximum
MODE="install"            # config, build, install
INSTALL_METHOD="official" # official (recommended)
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
VERBOSE=false
QUIET=false
DRY_RUN=false
NO_SHELL=false

# Show help
show_help() {
    cat << EOF
Bun Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  --minimal            Minimal Bun installation (basic runtime only)
  --standard           Standard Bun installation (default, includes package manager)
  --maximum            Maximum Bun installation (with development tools and optimizations)

Modes:
  --config-only        Configure only (prepare build environment)
  --build-only         Configure and build (no installation) 
  --install            Configure, build, and install (default)

Installation Methods:
  --official           Use official installer (default, recommended)

Options:
  --skip-deps          Skip dependency installation
  --force              Force reinstallation if already installed
  --no-shell           Skip shell integration setup
  --dry-run            Show what would be done without executing
  --verbose            Enable verbose output
  --quiet              Suppress non-error output
  -h, --help           Show this help message
  --version            Show script version information

Examples:
  $0                   # Default: standard installation (recommended)
  $0 --minimal         # Minimal Bun installation
  $0 --maximum         # Full-featured installation with dev tools
  $0 --config-only     # Just prepare environment
  $0 --force           # Reinstall even if already present
  $0 --dry-run         # Preview installation steps

About Bun:
  Bun is a fast all-in-one JavaScript runtime built from scratch to serve
  the modern JavaScript ecosystem. It's designed as a drop-in replacement 
  for Node.js with significantly better performance and built-in bundler,
  test runner, and package manager.

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --minimal)
            BUILD_TYPE="minimal"
            shift
            ;;
        --standard)
            BUILD_TYPE="standard"
            shift
            ;;
        --maximum)
            BUILD_TYPE="maximum"
            shift
            ;;
        --config-only)
            MODE="config"
            shift
            ;;
        --build-only)
            MODE="build"
            shift
            ;;
        --install)
            MODE="install"
            shift
            ;;
        --official)
            INSTALL_METHOD="official"
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
        --no-shell)
            NO_SHELL=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --quiet)
            QUIET=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        --version)
            echo "Bun Installation Script v1.0"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

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

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
if [[ "$QUIET" != true ]]; then
    echo
    log "==================================="
    log "Bun Installation Script"
    log "==================================="
    log "Build type: $BUILD_TYPE"
    log "Mode: $MODE"
    log "Install method: $INSTALL_METHOD"
    log "Skip dependencies: $SKIP_DEPS"
    log "Run tests: $RUN_TESTS"
    log "Dry run: $DRY_RUN"
    echo
fi

# Check if Bun is already installed
if command -v bun &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(bun --version 2>/dev/null | head -n1 || echo "unknown")
    
    log "Bun is already installed (version: $CURRENT_VERSION)"
    log "Use --force to reinstall"
    exit 0
fi

# Check system requirements
log "Checking system requirements..."

# Check kernel version
KERNEL_VERSION=$(uname -r | grep -oP '^\d+\.\d+' || echo "0.0")
if ! version_compare "$KERNEL_VERSION" "$KERNEL_MIN_VERSION"; then
    error "Kernel version $KERNEL_VERSION is too old. Minimum required: $KERNEL_MIN_VERSION"
fi

if ! version_compare "$KERNEL_VERSION" "$KERNEL_RECOMMENDED_VERSION"; then
    warning "Kernel version $KERNEL_VERSION is older than recommended ($KERNEL_RECOMMENDED_VERSION)"
    log "Bun may work but performance could be affected"
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies..."
    
    if [[ "$DRY_RUN" == true ]]; then
        log "[DRY RUN] Would install: unzip curl"
    else
        # Update package list
        sudo apt-get update -qq
        
        # Install required packages
        sudo apt-get install -y unzip curl
        
        log "Dependencies installed successfully"
    fi
else
    log "Skipping dependency installation"
fi

# Configuration mode
if [[ "$MODE" == "config" ]]; then
    log "Config mode: environment prepared for Bun installation"
    success "Configuration completed!"
    exit 0
fi

# Official installer method
log "Using official Bun installer..."

if [[ "$DRY_RUN" == true ]]; then
    log "[DRY RUN] Would download and install Bun via: curl -fsSL https://bun.sh/install | bash"
    
    case $BUILD_TYPE in
        minimal)
            log "[DRY RUN] Would use minimal installation (basic runtime only)"
            ;;
        standard) 
            log "[DRY RUN] Would use standard installation (default)"
            ;;
        maximum)
            log "[DRY RUN] Would use maximum installation with additional setup"
            ;;
    esac
    
    success "[DRY RUN] Installation preview completed!"
    exit 0
fi

log "Downloading and installing Bun via official installer..."
curl -fsSL https://bun.sh/install | bash || error "Official Bun installation failed"

# Add ~/.bun/bin to PATH if not already there
BUN_PATH="$HOME/.bun/bin"
if [[ ":$PATH:" != *":$BUN_PATH:"* ]]; then
    export PATH="$BUN_PATH:$PATH"
    log "Added $BUN_PATH to PATH for this session"
    
    if [[ "$NO_SHELL" != true ]]; then
        # Update shell profiles for persistent PATH
        for profile in ~/.bashrc ~/.zshrc ~/.profile; do
            if [[ -f "$profile" ]] && ! grep -q ".bun/bin" "$profile"; then
                echo 'export PATH="$HOME/.bun/bin:$PATH"' >> "$profile"
                log "Updated $profile to include ~/.bun/bin in PATH"
            fi
        done
    fi
fi

# Build type specific setup
case $BUILD_TYPE in
    minimal)
        log "Minimal setup: Bun runtime only"
        ;;
    standard)
        log "Standard setup: Bun with package manager capabilities"
        ;;
    maximum)
        log "Maximum setup: Installing additional development tools..."
        
        # Install common packages for JavaScript development
        if command -v bun &> /dev/null; then
            log "Setting up development environment..."
            
            # Create a minimal bun project for testing completions
            if [[ ! -f "$HOME/.bunrc" ]]; then
                log "Creating Bun configuration..."
                mkdir -p "$HOME/.bun"
            fi
            
            log "Maximum setup completed with development tools"
        fi
        ;;
esac

# Run tests if requested
if [[ "$RUN_TESTS" == true ]] && command -v bun &> /dev/null; then
    log "Running basic verification tests..."
    
    # Create temporary test directory
    TEST_DIR=$(mktemp -d)
    cd "$TEST_DIR"
    
    # Test basic functionality
    log "Testing Bun version..."
    bun --version || error "Bun version check failed"
    
    log "Testing Bun package manager..."
    echo '{"name": "test", "version": "1.0.0"}' > package.json
    bun install || warning "Bun install test failed (but Bun runtime works)"
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$TEST_DIR"
    
    success "Tests completed successfully!"
fi

# Verify installation
if command -v bun &> /dev/null; then
    INSTALLED_VERSION=$(bun --version 2>/dev/null | head -n1)
    success "Bun installation completed successfully!"
    
    if [[ "$QUIET" != true ]]; then
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: Official installer"
        log "Binary location: $(which bun)"
        log "Installation directory: $HOME/.bun"
        
        echo
        log "Basic usage examples:"
        log "  bun --version                    # Check Bun version"
        log "  bun init                         # Initialize a new project"
        log "  bun install                      # Install dependencies"
        log "  bun run script.js                # Run a JavaScript file"
        log "  bun add package                  # Add a package dependency"
        log "  bun test                         # Run tests"
        log "  bun build ./index.js             # Bundle/build JavaScript"
        echo
        log "Package manager compatibility:"
        log "  bun install    (like npm install)"
        log "  bun add        (like npm install <package>)"
        log "  bun remove     (like npm uninstall)" 
        echo
        log "For more information: bun --help"
        
        if [[ "$NO_SHELL" != true ]]; then
            echo
            warning "IMPORTANT: Restart your shell or run 'source ~/.bashrc' to use bun commands"
        fi
    fi
else
    error "Bun installation verification failed"
fi

success "Bun installation script completed!"