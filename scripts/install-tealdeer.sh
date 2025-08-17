#!/bin/bash

# tealdeer Installation Script for Debian Linux
# Fast tldr pages client
# Usage: ./install-tealdeer.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
SKIP_DEPS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
tealdeer Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Options:
  --skip-deps          Skip dependency installation
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

About tealdeer:
  A fast implementation of tldr in Rust. Provides quick command help
  without needing to read full man pages. Much faster than the Python
  version typically found in Debian repositories.

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

# Logging functions
log() { echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

# Check if running as root
[[ $EUID -eq 0 ]] && error "This script should not be run as root for security reasons"

log "Installing tealdeer - Fast tldr client"

# Check if tldr is already installed
if command -v tldr &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "tldr already installed. Use --force to reinstall"
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

# Install tealdeer
log "Installing tealdeer from crates.io..."
cargo install tealdeer

# Verify installation and update cache
if command -v tldr &> /dev/null; then
    success "tealdeer installed successfully!"
    log "Installed version: $(tldr --version 2>/dev/null | head -n1)"
    log "Updating tldr cache..."
    tldr --update
    echo
    log "Usage examples:"
    log "  tldr tar                         # Quick tar command help"
    log "  tldr --list                      # List all available pages"
    log "  tldr --update                    # Update the cache"
    log "  tldr --random                    # Show a random page"
else
    error "Installation failed"
fi