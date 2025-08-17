#!/bin/bash

# Common Dependencies Installation Script for All Build Tools
# Installs shared dependencies once to avoid redundant installations
# Usage: ./install-common-deps.sh

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
RUST_MIN_VERSION="1.88.0"  # Highest requirement (ripgrep)
GO_VERSION="1.23.4"        # Latest stable

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

log "Installing common dependencies for all build tools..."

# Update package list once
log "Updating package list..."
sudo apt update || error "Failed to update package list"

# Install common build tools
log "Installing common build tools..."
sudo apt install -y \
    build-essential \
    git \
    curl \
    wget \
    make \
    cmake \
    pkg-config \
    autoconf \
    automake \
    libtool \
    || error "Failed to install common build tools"

# Install additional tools used by multiple scripts
log "Installing additional common tools..."
sudo apt install -y \
    nasm \
    yasm \
    bison \
    flex \
    python3-dev \
    || warning "Some additional tools may not be available"

success "Common build tools installed successfully"

# Install/Update Rust to satisfy highest requirement
log "Installing/updating Rust to version >= $RUST_MIN_VERSION..."
if command -v rustc &> /dev/null; then
    current_version=$(rustc --version | cut -d' ' -f2)
    log "Found Rust version: $current_version"
    
    if version_compare $current_version $RUST_MIN_VERSION; then
        log "Rust version is sufficient (>= $RUST_MIN_VERSION)"
    else
        warning "Rust version $current_version is below required $RUST_MIN_VERSION"
        log "Updating Rust..."
        rustup update || error "Failed to update Rust"
        success "Rust updated successfully"
    fi
else
    log "Rust not found, installing..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y || error "Failed to install Rust"
    source ~/.cargo/env || error "Failed to source Rust environment"
    success "Rust installed successfully"
fi

# Ensure Rust is up to date
log "Ensuring Rust toolchain is current..."
rustup update || warning "Failed to update Rust toolchain"

# Add MUSL target for static builds (ripgrep)
log "Adding MUSL target for static builds..."
rustup target add x86_64-unknown-linux-musl || warning "Failed to add MUSL target"

# Install/Update Go
log "Installing/updating Go to version $GO_VERSION..."
if command -v go &> /dev/null; then
    current_version=$(go version | sed 's/go version go\([0-9.]*\).*/\1/')
    log "Found Go version: $current_version"
    
    if version_compare $current_version "1.20"; then
        log "Go version is sufficient (>= 1.20)"
    else
        warning "Go version $current_version is below required 1.20"
        log "Installing newer Go..."
    fi
else
    log "Go not found, installing..."
fi

# Install Go if needed
if ! command -v go &> /dev/null || ! version_compare $(go version | sed 's/go version go\([0-9.]*\).*/\1/') "1.20"; then
    GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://golang.org/dl/${GO_TARBALL}"
    
    cd /tmp
    wget -q "$GO_URL" || error "Failed to download Go"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_TARBALL" || error "Failed to extract Go"
    rm "$GO_TARBALL"
    
    success "Go installed successfully"
fi

# Setup environment once
log "Setting up environment variables..."

# Add Rust to PATH if not already there
if [[ ":$PATH:" != *":$HOME/.cargo/bin:"* ]]; then
    if ! grep -q 'export PATH="$HOME/.cargo/bin:$PATH"' ~/.bashrc 2>/dev/null; then
        echo '# Rust environment' >> ~/.bashrc
        echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
        log "Added Rust to PATH in ~/.bashrc"
    fi
    export PATH="$HOME/.cargo/bin:$PATH"
fi

# Add Go to PATH if not already there
if [[ ":$PATH:" != *":/usr/local/go/bin:"* ]]; then
    if ! grep -q 'export PATH="/usr/local/go/bin:$PATH"' ~/.bashrc 2>/dev/null; then
        echo '# Go environment' >> ~/.bashrc
        echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
        log "Added Go to PATH in ~/.bashrc"
    fi
    export PATH="/usr/local/go/bin:$PATH"
fi

# Verify installations
log "Verifying toolchain installations..."
echo
log "Installed versions:"
if command -v rustc &> /dev/null; then
    log "  Rust: $(rustc --version)"
else
    warning "  Rust: Not found in current session"
fi

if command -v go &> /dev/null; then
    log "  Go: $(go version)"
else
    warning "  Go: Not found in current session"
fi

echo
success "Common dependencies installation completed successfully!"
echo
log "Toolchains installed:"
log "  - Rust >= $RUST_MIN_VERSION (for fd, ripgrep)"
log "  - Go >= 1.20 (for fzf)"
log "  - C/C++ build tools (for ffmpeg, 7zip, jq)"
echo
log "To use the toolchains immediately in this terminal, run:"
log "  source ~/.bashrc"
echo
log "Now you can run individual tool installation scripts with --skip-deps flag"