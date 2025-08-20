#!/bin/bash
#
# @file lib/system/installation.sh
# @brief System installation and binary management
# @description
#   Handles binary installation, system integration,
#   and post-installation setup for tools.
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_SYSTEM_INSTALLATION_LOADED:-}" ]] && return 0
readonly GEARBOX_SYSTEM_INSTALLATION_LOADED=1

# @function safe_install_binary
# @brief Safely install binary to system location
# @param $1 Source binary path
# @param $2 Binary name (optional, defaults to basename of source)
safe_install_binary() {
    local src_binary="$1"
    local binary_name="${2:-$(basename "$src_binary")}"
    local dest_path="/usr/local/bin/$binary_name"
    
    [[ -z "$src_binary" ]] && error "Source binary not specified"
    [[ ! -f "$src_binary" ]] && error "Source binary does not exist: $src_binary"
    [[ ! -x "$src_binary" ]] && error "Source binary is not executable: $src_binary"
    
    log "Installing binary: $binary_name -> $dest_path"
    safe_sudo_copy "$src_binary" "$dest_path"
    
    # Verify installation
    if command -v "$binary_name" &> /dev/null; then
        success "Binary installed successfully: $binary_name"
    else
        error "Binary installation failed - not found in PATH: $binary_name"
    fi
}

log "System installation module loaded"