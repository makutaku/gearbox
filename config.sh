#!/bin/bash
#
# @file config.sh
# @brief Central configuration file for the Essential Tools Installer
# @description
#   This file contains all shared configuration settings, paths, and
#   environment variables used across the gearbox installation system.
#   It serves as the single source of truth for all configurable values.
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_CONFIG_LOADED:-}" ]] && return 0
readonly GEARBOX_CONFIG_LOADED=1

# =============================================================================
# CORE DIRECTORY CONFIGURATION
# =============================================================================

# Build directory (where source repositories are cloned)
export BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"

# Cache directory (for downloads and temporary files)
export CACHE_DIR="${CACHE_DIR:-$HOME/tools/cache}"

# Installation prefix (where binaries are installed)
export INSTALL_PREFIX="${INSTALL_PREFIX:-/usr/local}"

# Binary installation directory
export BIN_DIR="${BIN_DIR:-$INSTALL_PREFIX/bin}"

# =============================================================================
# TOOL VERSION REQUIREMENTS
# =============================================================================

# Minimum required versions for language toolchains
export RUST_MIN_VERSION="${RUST_MIN_VERSION:-1.88.0}"
export GO_MIN_VERSION="${GO_MIN_VERSION:-1.23.4}"
export PYTHON_MIN_VERSION="${PYTHON_MIN_VERSION:-3.11.0}"

# =============================================================================
# BUILD CONFIGURATION
# =============================================================================

# Default build settings
export DEFAULT_BUILD_TYPE="${DEFAULT_BUILD_TYPE:-standard}"
export DEFAULT_PARALLEL_JOBS="${DEFAULT_PARALLEL_JOBS:-$(nproc)}"

# Build timeout (in seconds)
export BUILD_TIMEOUT="${BUILD_TIMEOUT:-3600}"  # 1 hour

# =============================================================================
# COLOR DEFINITIONS
# =============================================================================

export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export BLUE='\033[0;34m'
export PURPLE='\033[0;35m'
export CYAN='\033[0;36m'
export WHITE='\033[1;37m'
export NC='\033[0m'  # No Color

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# @function get_source_dir
# @brief Get the full path to a tool's source directory
# @param $1 Tool name
# @return Full path to tool's source directory
get_source_dir() {
    local tool_name="$1"
    [[ -z "$tool_name" ]] && { echo "Error: Tool name required" >&2; return 1; }
    echo "$BUILD_DIR/$tool_name"
}

# @function get_cache_path
# @brief Get the full path to a cache file
# @param $1 Cache filename
# @return Full path to cache file
get_cache_path() {
    local filename="$1"
    [[ -z "$filename" ]] && { echo "Error: Filename required" >&2; return 1; }
    echo "$CACHE_DIR/$filename"
}

# @function get_binary_path
# @brief Get the full path where a binary should be installed
# @param $1 Binary name
# @return Full path to binary installation location
get_binary_path() {
    local binary_name="$1"
    [[ -z "$binary_name" ]] && { echo "Error: Binary name required" >&2; return 1; }
    echo "$BIN_DIR/$binary_name"
}

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================

# Ensure required directories exist
mkdir -p "$BUILD_DIR" "$CACHE_DIR" || {
    echo "Error: Failed to create required directories" >&2
    exit 1
}

# Set up PATH to include installation directory
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    export PATH="$BIN_DIR:$PATH"
fi

# Set up environment variables for builds
export MAKEFLAGS="-j${DEFAULT_PARALLEL_JOBS}"

# =============================================================================
# DEPRECATED FUNCTIONS WARNING
# =============================================================================

# Legacy logging functions - DEPRECATED
# These are kept for backward compatibility but should not be used.
# Scripts should source lib/common.sh instead for proper shared functionality.

log() {
    echo "WARNING: Using deprecated log() from config.sh. Source lib/common.sh instead." >&2
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo "WARNING: Using deprecated error() from config.sh. Source lib/common.sh instead." >&2
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

success() {
    echo "WARNING: Using deprecated success() from config.sh. Source lib/common.sh instead." >&2
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo "WARNING: Using deprecated warning() from config.sh. Source lib/common.sh instead." >&2
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# =============================================================================
# INITIALIZATION COMPLETE
# =============================================================================

# Export configuration loaded flag
export GEARBOX_CONFIG_LOADED