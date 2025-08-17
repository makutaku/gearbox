#!/bin/bash

# Configuration for all installation scripts
# This file contains shared settings and paths

# Build directory (where source repositories are cloned)
export BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"

# Cache directory (for downloads and temporary files)
export CACHE_DIR="${CACHE_DIR:-$HOME/tools/cache}"

# Installation prefix (where binaries are installed)
export INSTALL_PREFIX="${INSTALL_PREFIX:-/usr/local}"

# Ensure directories exist
mkdir -p "$BUILD_DIR" "$CACHE_DIR"

# Function to get source directory path
get_source_dir() {
    local tool_name="$1"
    echo "$BUILD_DIR/$tool_name"
}

# Function to get cache file path
get_cache_path() {
    local filename="$1"
    echo "$CACHE_DIR/$filename"
}

# Color definitions (shared across scripts)
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export BLUE='\033[0;34m'
export NC='\033[0m'

# Shared logging functions
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