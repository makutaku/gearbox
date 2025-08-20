#!/bin/bash
#
# @file install-dust.sh
# @brief Installation script for dust - Better disk usage analyzer
# @description
#   Installs dust, a more intuitive version of du written in Rust.
#   Supports both binary and source installation methods.
#
# @usage ./install-dust.sh [OPTIONS]
# @param --source      Build from source instead of downloading binary
# @param --force       Force reinstallation if already installed
REPO_DIR="$(dirname "$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")")"
#

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

# Configuration loaded from lib/common.sh

# Define simplified shared functions
# Note: Logging functions now provided by lib/common.sh

# Note: Logging functions now provided by lib/common.sh

# Note: Logging functions now provided by lib/common.sh

# Note: Logging functions now provided by lib/common.sh

check_existing_installation() {
    local tool_name="$1"
    local force_install="${2:-false}"
    
    if command -v "$tool_name" &> /dev/null && [[ "$force_install" != "true" ]]; then
        local version_info=""
        if "$tool_name" --version &> /dev/null; then
            version_info=" ($("$tool_name" --version 2>/dev/null | head -1))"
        fi
        log "$tool_name already installed$version_info. Use --force to reinstall."
        exit 0
    fi
}

create_temp_dir() {
    TEMP_DIR=$(mktemp -d)
    echo "$TEMP_DIR"
}

safe_sudo_copy() {
    local src_file="$1"
    local dest_file="$2"
    
    [[ -f "$src_file" ]] || error "Source file does not exist: $src_file"
    
    sudo cp "$src_file" "$dest_file" || error "Failed to copy $src_file to $dest_file"
    sudo chmod +x "$dest_file" || error "Failed to make $dest_file executable"
    
    success "Copied $src_file to $dest_file"
}

setup_error_cleanup() {
    trap 'cleanup_on_error $?' EXIT INT TERM
}

cleanup_on_error() {
    local exit_code="$1"
    if [[ $exit_code -ne 0 ]]; then
        warning "Operation failed with exit code $exit_code, performing cleanup..."
        [[ -n "${TEMP_DIR:-}" && -d "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"
    fi
}

# Tool-specific configuration
readonly TOOL_NAME="dust"
readonly TOOL_REPO="https://github.com/bootandy/dust.git"
readonly GITHUB_REPO="bootandy/dust"
readonly BINARY_NAME="dust"
readonly PACKAGE_NAME="du-dust"  # Cargo package name

# Default options
INSTALL_METHOD="binary"
FORCE_INSTALL=false

# =============================================================================
# COMMAND LINE ARGUMENT PARSING
# =============================================================================

show_help() {
    cat << EOF
dust Installation Script - Better disk usage analyzer

Usage: $0 [OPTIONS]

Options:
  --source             Build from source using Rust/Cargo
  --binary             Download pre-built binary (default)
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Install using pre-built binary
  $0 --source          # Build from source
  $0 --force           # Force reinstall

About dust:
  A more intuitive version of du written in Rust. Provides a colorful
  and interactive view of disk usage with directory sizes.

EOF
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --source)
            INSTALL_METHOD="source"
            shift
            ;;
        --binary)
            INSTALL_METHOD="binary"
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
            error "Unknown option: $1. Use -h for help."
            ;;
    esac
done

# =============================================================================
# MAIN INSTALLATION LOGIC
# =============================================================================

main() {
    # Set up error cleanup
    setup_error_cleanup
    
    # Security check
    [[ $EUID -eq 0 ]] && error "Don't run as root for security reasons"
    
    log "Installing $TOOL_NAME - Better disk usage analyzer"
    
    # Check existing installation
    check_existing_installation "$BINARY_NAME" "$FORCE_INSTALL"
    
    # Install based on method
    case "$INSTALL_METHOD" in
        binary)
            install_from_binary
            ;;
        source)
            install_from_source
            ;;
        *)
            error "Unknown installation method: $INSTALL_METHOD"
            ;;
    esac
    
    # Verify installation
    verify_installation
}

# =============================================================================
# INSTALLATION METHODS
# =============================================================================

install_from_binary() {
    log "Installing $TOOL_NAME from pre-built binary..."
    
    # Detect architecture
    local arch=$(uname -m)
    local arch_tag
    case $arch in
        x86_64)
            arch_tag="x86_64-unknown-linux-musl"
            ;;
        *)
            error "Unsupported architecture: $arch. Use --source to build from source."
            ;;
    esac
    
    # Get latest version
    log "Fetching latest release information..."
    local version
    version=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$version" ]] && error "Failed to fetch latest version"
    
    # Download and extract securely
    local download_url="https://github.com/$GITHUB_REPO/releases/download/${version}/dust-${version}-${arch_tag}.tar.gz"
    local filename="dust-${version}-${arch_tag}.tar.gz"
    
    local temp_dir
    temp_dir=$(create_temp_dir)
    cd "$temp_dir"
    
    # Use secure download function
    local temp_file="$temp_dir/$filename"
    secure_download "$download_url" "$temp_file"
    
    # Extract archive
    tar -xzf "$temp_file" || error "Failed to extract archive"
    
    # Find and install binary
    local dust_bin
    dust_bin=$(find . -name "$BINARY_NAME" -type f -executable | head -1)
    [[ -z "$dust_bin" ]] && error "$TOOL_NAME binary not found in archive"
    
    # Install binary
    local install_path="/usr/local/bin/$BINARY_NAME"
    safe_sudo_copy "$dust_bin" "$install_path"
    
    success "$TOOL_NAME $version installed successfully from binary"
}

install_from_source() {
    log "Installing $TOOL_NAME from source..."
    
    # Install Rust if needed
    if ! command -v rustc &> /dev/null; then
        log "Installing Rust..."
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source ~/.cargo/env
    fi
    
    # Install using cargo
    log "Building $TOOL_NAME using cargo..."
    cargo install "$PACKAGE_NAME" || error "Failed to build $TOOL_NAME from source"
    
    # Copy from cargo bin to system bin if needed
    local cargo_bin="$HOME/.cargo/bin/$BINARY_NAME"
    local system_bin="/usr/local/bin/$BINARY_NAME"
    
    if [[ -f "$cargo_bin" && "$cargo_bin" != "$system_bin" ]]; then
        safe_sudo_copy "$cargo_bin" "$system_bin"
    fi
    
    success "$TOOL_NAME installed successfully from source"
}

verify_installation() {
    log "Verifying $TOOL_NAME installation..."
    
    if command -v "$BINARY_NAME" &> /dev/null; then
        local version_info
        version_info=$("$BINARY_NAME" --version 2>/dev/null || echo "version unknown")
        success "$TOOL_NAME verified successfully: $version_info"
        
        log "Usage examples:"
        log "  dust              # Show disk usage for current directory"
        log "  dust -r           # Reverse sort (largest last)"
        log "  dust -n 20        # Show top 20 entries"
        log "  dust -d 3         # Limit depth to 3 levels"
        log "  dust /path        # Analyze specific directory"
    else
        error "$TOOL_NAME installation failed - binary not found in PATH"
    fi
}

# =============================================================================
# SCRIPT EXECUTION
# =============================================================================

main "$@"
