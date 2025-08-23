#!/bin/bash

# just Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-just.sh [OPTIONS]

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
JUST_DIR="just"
JUST_REPO="https://github.com/casey/just.git"
RUST_MIN_VERSION="1.70.0"

# Default options
BUILD_TYPE="release"      # debug, release, optimized
MODE="install"            # config, build, install
INSTALL_METHOD="source"   # source, binary (default to source for reliability)
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
just Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build (unoptimized, with debug symbols)
  -r, --release         Release build (default, optimized)
  -o, --optimized       Optimized build (release with target-cpu=native)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --minimal             Minimal build (maps to -d debug)
  --standard            Standard build (maps to -r release, default)  
  --maximum             Maximum build (maps to -o optimized)
  --skip-deps           Skip dependency installation
  --run-tests           Run test suites after build
  --force               Force reinstallation if already installed
  -h, --help            Show this help message

Examples:
  $0                    # Standard installation (release build)
  $0 --minimal          # Minimal debug build
  $0 --maximum          # Optimized build with native CPU features
  $0 --config-only      # Prepare build environment only
  $0 --skip-deps --force # Skip deps and force reinstall

This script installs 'just', a modern command runner written in Rust.
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            # Build types
            -d|--debug)
                BUILD_TYPE="debug"
                shift ;;
            -r|--release)
                BUILD_TYPE="release"
                shift ;;
            -o|--optimized)
                BUILD_TYPE="optimized"
                shift ;;
            
            # Gearbox Standard Protocol v1.0
            --minimal)
                BUILD_TYPE="debug"
                shift ;;
            --standard)
                BUILD_TYPE="release"
                shift ;;
            --maximum)
                BUILD_TYPE="optimized"
                shift ;;
            
            # Modes
            -c|--config-only)
                MODE="config"
                shift ;;
            -b|--build-only)
                MODE="build"
                shift ;;
            -i|--install)
                MODE="install"
                shift ;;
            
            # Options
            --skip-deps)
                SKIP_DEPS=true
                shift ;;
            --run-tests)
                RUN_TESTS=true
                shift ;;
            --force)
                FORCE_INSTALL=true
                shift ;;
            -h|--help)
                show_help
                exit 0 ;;
            *)
                log "WARNING" "Unknown option: $1 (ignoring)"
                shift ;;
        esac
    done
}

# Check if just is already installed
check_existing_installation() {
    if command -v just >/dev/null 2>&1 && [[ "$FORCE_INSTALL" != "true" ]]; then
        EXISTING_VERSION=$(just --version 2>/dev/null | head -n1 | awk '{print $2}' || echo "unknown")
        log "SUCCESS" "just is already installed (version: $EXISTING_VERSION)"
        log "INFO" "Use --force to reinstall"
        exit 0
    fi
}

# Install dependencies
install_dependencies() {
    if [[ "$SKIP_DEPS" == "true" ]]; then
        log "INFO" "Skipping dependency installation"
        return 0
    fi
    
    log "INFO" "Installing dependencies for just..."
    
    # Load and check for Rust installation
    load_module "build/dependencies"
    
    # Ensure Rust is installed with minimum version
    if ! ensure_rust_version "$RUST_MIN_VERSION"; then
        error "Failed to install or verify Rust $RUST_MIN_VERSION+"
    fi
    
    success "Dependencies installed successfully"
}

# Configure build
configure_build() {
    log "INFO" "Configuring just build..."
    
    # Create build directory and navigate
    BUILD_DIR="$HOME/tools/build/$JUST_DIR"
    mkdir -p "$(dirname "$BUILD_DIR")"
    
    # Clone or update repository
    if [[ -d "$BUILD_DIR" ]]; then
        if [[ "$FORCE_INSTALL" == "true" ]]; then
            log "INFO" "Removing existing build directory"
            rm -rf "$BUILD_DIR"
        else
            log "INFO" "Updating existing repository"
            cd "$BUILD_DIR"
            git pull --quiet
            return 0
        fi
    fi
    
    log "INFO" "Cloning just repository..."
    git clone --depth 1 "$JUST_REPO" "$BUILD_DIR"
    
    success "Build configured successfully"
}

# Build just
build_just() {
    log "INFO" "Building just with $BUILD_TYPE build..."
    
    BUILD_DIR="$HOME/tools/build/$JUST_DIR"
    cd "$BUILD_DIR"
    
    # Get optimal number of jobs (function is in core/utilities.sh)
    JOBS=$(get_optimal_jobs)
    
    # Configure build flags based on build type
    case "$BUILD_TYPE" in
        debug)
            log "INFO" "Building with debug flags (fast compilation)"
            cargo build --jobs "$JOBS"
            BINARY_PATH="target/debug/just"
            ;;
        release)
            log "INFO" "Building with release flags (optimized)"
            cargo build --release --jobs "$JOBS"
            BINARY_PATH="target/release/just"
            ;;
        optimized)
            log "INFO" "Building with optimized flags (maximum performance)"
            export RUSTFLAGS="-C target-cpu=native"
            cargo build --release --jobs "$JOBS"
            BINARY_PATH="target/release/just"
            ;;
        *)
            error "Unknown build type: $BUILD_TYPE"
            ;;
    esac
    
    # Verify binary was created
    if [[ ! -f "$BINARY_PATH" ]]; then
        error "Build failed: binary not found at $BINARY_PATH"
    fi
    
    success "just built successfully"
}

# Run tests
run_tests() {
    if [[ "$RUN_TESTS" != "true" ]]; then
        return 0
    fi
    
    log "INFO" "Running just test suite..."
    
    BUILD_DIR="$HOME/tools/build/$JUST_DIR"
    cd "$BUILD_DIR"
    
    # Run tests
    if cargo test --quiet; then
        success "All tests passed"
    else
        log "WARNING" "Some tests failed, but continuing installation"
    fi
}

# Install just
install_just() {
    log "INFO" "Installing just to /usr/local/bin..."
    
    BUILD_DIR="$HOME/tools/build/$JUST_DIR"
    cd "$BUILD_DIR"
    
    # Copy binary
    sudo cp "$BINARY_PATH" /usr/local/bin/just
    sudo chmod +x /usr/local/bin/just
    
    # Verify installation
    if command -v just >/dev/null 2>&1; then
        INSTALLED_VERSION=$(just --version 2>/dev/null | head -n1 | awk '{print $2}' || echo "unknown")
        success "just installed successfully (version: $INSTALLED_VERSION)"
        
        # Basic functionality test
        log "INFO" "Testing just functionality..."
        if just --help >/dev/null 2>&1; then
            success "just is working correctly"
        else
            log "WARNING" "just installation may have issues - please test manually"
        fi
    else
        error "Installation verification failed - just not found in PATH"
    fi
}

# Main execution
main() {
    # Parse arguments
    parse_args "$@"
    
    log "INFO" "Starting just installation..."
    log "INFO" "Build type: $BUILD_TYPE, Mode: $MODE"
    
    # Check if already installed (unless forcing)
    check_existing_installation
    
    # Install dependencies
    install_dependencies
    
    # Configure build
    configure_build
    
    # Stop here if config-only mode
    if [[ "$MODE" == "config" ]]; then
        success "Configuration completed successfully"
        exit 0
    fi
    
    # Build
    build_just
    
    # Run tests if requested
    run_tests
    
    # Stop here if build-only mode
    if [[ "$MODE" == "build" ]]; then
        success "Build completed successfully"
        success "Binary available at: $BUILD_DIR/$BINARY_PATH"
        exit 0
    fi
    
    # Install
    install_just
    
    success "just installation completed successfully!"
    log "INFO" "Run 'just --help' to get started"
}

# Execute main function
main "$@"