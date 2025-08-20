#!/bin/bash
#
# @file lib/common.sh
# @brief Modular shared library for all gearbox installation scripts
# @description
#   New modular entry point that loads core modules and provides lazy loading
#   for optional functionality. Replaces the monolithic common.sh with a
#   clean, maintainable modular system.
#
# @author Essential Tools Installer Team
# @version 2.0.0
# @since 2024-01-01
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_COMMON_LOADED:-}" ]] && return 0
readonly GEARBOX_COMMON_LOADED=1

# =============================================================================
# INITIALIZATION AND ENVIRONMENT
# =============================================================================

# Find the script directory and set global paths
if [[ -z "${GEARBOX_LIB_DIR:-}" ]]; then
    GEARBOX_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    GEARBOX_REPO_DIR="$(dirname "$GEARBOX_LIB_DIR")"
    readonly GEARBOX_LIB_DIR GEARBOX_REPO_DIR
fi

# Track loaded modules to prevent duplicate loading
declare -A GEARBOX_LOADED_MODULES

# =============================================================================
# CORE MODULE LOADING
# =============================================================================

# @function load_module
# @brief Load a gearbox module with validation and duplicate prevention
# @param $1 Module path (relative to lib/ directory)
# @return 0 on success, 1 on failure
load_module() {
    local module_path="$1"
    local full_path="$GEARBOX_LIB_DIR/$module_path"
    
    # Check if already loaded
    if [[ -n "${GEARBOX_LOADED_MODULES[$module_path]:-}" ]]; then
        return 0  # Already loaded
    fi
    
    # Validate module exists
    if [[ ! -f "$full_path" ]]; then
        echo "ERROR: Module not found: $module_path" >&2
        return 1
    fi
    
    # Load the module
    if source "$full_path"; then
        GEARBOX_LOADED_MODULES["$module_path"]=1
        return 0
    else
        echo "ERROR: Failed to load module: $module_path" >&2
        return 1
    fi
}

# Load core modules (always needed)
load_module "core/logging.sh" || exit 1
load_module "core/validation.sh" || exit 1
load_module "core/security.sh" || exit 1
load_module "core/utilities.sh" || exit 1

# Load tracking module
load_module "tracking.sh" || exit 1

# Load configuration management
if [[ -f "$GEARBOX_LIB_DIR/config.sh" ]]; then
    source "$GEARBOX_LIB_DIR/config.sh"
    # Initialize configuration system if function exists
    if declare -f init_config >/dev/null; then
        init_config
    fi
fi

# =============================================================================
# LAZY LOADING FUNCTIONS FOR OPTIONAL MODULES
# =============================================================================

# @function require_build_modules
# @brief Load build-related modules when needed
require_build_modules() {
    load_module "build/dependencies.sh" || return 1
    load_module "build/execution.sh" || return 1
    load_module "build/cache.sh" || return 1
    load_module "build/cleanup.sh" || return 1
}

# @function require_system_modules
# @brief Load system-related modules when needed
require_system_modules() {
    load_module "system/installation.sh" || return 1
    load_module "system/backup.sh" || return 1
    load_module "system/environment.sh" || return 1
}

# =============================================================================
# CONVENIENCE FUNCTIONS FOR COMMON OPERATIONS
# =============================================================================

# @function ensure_build_environment
# @brief Ensure build modules are loaded and environment is ready
ensure_build_environment() {
    require_build_modules || error "Failed to load build modules"
    ensure_not_root
}

# @function ensure_installation_environment
# @brief Ensure installation modules are loaded and environment is ready
ensure_installation_environment() {
    require_system_modules || error "Failed to load system modules"
    ensure_not_root
}

# =============================================================================
# MODULE STATUS AND DEBUGGING
# =============================================================================

# @function list_loaded_modules
# @brief List all currently loaded modules (for debugging)
list_loaded_modules() {
    debug "Loaded modules:"
    for module in "${!GEARBOX_LOADED_MODULES[@]}"; do
        debug "  - $module"
    done
}

# @function get_module_status
# @brief Check if a specific module is loaded
# @param $1 Module path
# @return 0 if loaded, 1 if not loaded
get_module_status() {
    local module_path="$1"
    [[ -n "${GEARBOX_LOADED_MODULES[$module_path]:-}" ]]
}

# =============================================================================
# BACKWARD COMPATIBILITY LAYER
# =============================================================================

# For scripts that expect certain functions to be immediately available,
# we can add auto-loading here. This section can be removed once all
# scripts are migrated to use explicit module loading.

# Auto-load build modules if certain functions are called
# Note: This is a compatibility layer and should be removed in v3.0

# Create wrapper functions that auto-load modules when needed
if ! declare -f install_rust_if_needed >/dev/null; then
    install_rust_if_needed() {
        require_build_modules || return 1
        install_rust_if_needed "$@"
    }
fi

if ! declare -f safe_install_binary >/dev/null; then
    safe_install_binary() {
        require_system_modules || return 1
        safe_install_binary "$@"
    }
fi

# =============================================================================
# INITIALIZATION COMPLETE
# =============================================================================

# Log successful initialization in debug mode
debug "Gearbox modular library initialized (v2.0.0)"
debug "Library directory: $GEARBOX_LIB_DIR"
debug "Repository directory: $GEARBOX_REPO_DIR"