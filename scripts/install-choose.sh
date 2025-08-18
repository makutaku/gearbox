#!/bin/bash

# choose Installation Script for Debian Linux
# Human-friendly cut/awk alternative
# Usage: ./install-choose.sh [OPTIONS]

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


# Default options
SKIP_DEPS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
choose Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Options:
  --skip-deps          Skip dependency installation
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

About choose:
  A human-friendly alternative to cut and awk for selecting columns
  from text. Much more intuitive syntax than traditional Unix tools.

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

# Check if running as root
[[ $EUID -eq 0 ]] && error "This script should not be run as root for security reasons"

log "Installing choose - Human-friendly cut alternative"

# Check if choose is already installed
if command -v choose &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "choose already installed. Use --force to reinstall"
    exit 0
fi

# Install Rust if needed
if [[ "$SKIP_DEPS" != true ]] && ! command -v rustc &> /dev/null; then
    log "Installing Rust..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source ~/.cargo/env
fi

# Ensure Rust is available
if ! command -v rustc &> /dev/null; then
    [[ -f ~/.cargo/env ]] && source ~/.cargo/env || error "Rust is not available"
fi

# Install choose
log "Installing choose from crates.io..."
cargo install choose

# Verify installation
if command -v choose &> /dev/null; then
    success "choose installed successfully!"
    log "Installed version: $(choose --version 2>/dev/null | head -n1)"
    echo
    log "Usage examples:"
    log "  echo 'a b c d' | choose 1       # Select second column (0-indexed)"
    log "  echo 'a b c d' | choose 1:3     # Select columns 1-3"
    log "  echo 'a b c d' | choose -1      # Select last column"
    log "  echo 'a:b:c:d' | choose -f ':' 1  # Use custom delimiter"
else
    error "Installation failed"
fi
