#!/bin/bash

# sd Installation Script for Debian Linux
# Intuitive find & replace CLI (sed alternative)
# Usage: ./install-sd.sh [OPTIONS]

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
RUST_MIN_VERSION="1.70.0"

# Default options
SKIP_DEPS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
sd Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Options:
  --skip-deps          Skip dependency installation
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

About sd:
  An intuitive find & replace CLI (sed alternative). Provides a more
  user-friendly interface for text replacement operations with better
  syntax and safer defaults than traditional sed commands.

Examples:
  sd 'old_text' 'new_text' file.txt    # Replace in file
  sd 'pattern' 'replacement' **/*.rs   # Replace in multiple files

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-deps) SKIP_DEPS=true; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) shift ;;
    esac
done

# Note: Logging functions now provided by lib/common.sh

# Version comparison function
version_compare() {
    if [[ $1 == $2 ]]; then return 0; fi
    local IFS=.; local i ver1=($1) ver2=($2)
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do ver1[i]=0; done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then ver2[i]=0; fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then return 0; fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then return 1; fi
    done; return 0
}

# Check if running as root
[[ $EUID -eq 0 ]] && error "This script should not be run as root for security reasons"

# Header
echo
log "======================================"
log "sd Installation Script"
log "======================================"
echo

# Check if sd is already installed
if command -v sd &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(sd --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    log "sd is already installed (version: $CURRENT_VERSION)"
    log "Building latest version from source (gearbox builds from latest main branch)"
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for sd..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl
    
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

# Install sd using cargo
log "Installing sd from crates.io..."
cargo install sd

# Verify installation
if command -v sd &> /dev/null; then
    INSTALLED_VERSION=$(sd --version 2>/dev/null | head -n1)
    success "sd installation completed successfully!"
    log "Installed version: $INSTALLED_VERSION"
    log "Installation method: Cargo install from crates.io"
    log "Binary location: $(which sd)"
    
    # Show usage information
    echo
    log "Basic usage examples:"
    log "  sd 'old_text' 'new_text' file.txt       # Replace in single file"
    log "  sd 'pattern' 'replacement' **/*.rs      # Replace in multiple files"
    log "  sd 'foo' 'bar' < input.txt              # Replace from stdin"
    log "  sd -p '\\d+' 'NUMBER' file.txt           # Use regex patterns"
    echo
    log "Advanced features:"
    log "  sd --preview 'old' 'new' file.txt       # Preview changes"
    log "  sd --string-mode 'literal' 'new' file   # Literal string mode"
    log "  sd -s 'case' 'CASE' file.txt            # Case-sensitive mode"
    echo
    log "Key advantages over sed:"
    log "  ✓ Intuitive syntax without cryptic flags"
    log "  ✓ Safe defaults (no accidental global replacement)"
    log "  ✓ Unicode support"
    log "  ✓ Better error messages"
    log "  ✓ Preview mode to see changes before applying"
    echo
    log "For more information: sd --help"
else
    error "sd installation verification failed - sd not found in PATH"
fi

success "sd installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"
