#!/bin/bash
#
# Installation tracking functions for gearbox uninstallation feature
# Provides functions to track tool installations in manifest

# Source required dependencies
if [[ "${BASH_SOURCE[0]}" != "${0}" ]]; then
    # Being sourced, load dependencies
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    source "$script_dir/core/logging.sh"
    source "$script_dir/core/utilities.sh"
fi

# Global tracking variables
GEARBOX_MANIFEST_DIR="$HOME/.gearbox"
GEARBOX_MANIFEST_FILE="$GEARBOX_MANIFEST_DIR/manifest.json"

# Track tool installation
# Usage: track_installation TOOL_NAME METHOD VERSION [OPTIONS...]
track_installation() {
    local tool_name="$1"
    local method="$2" 
    local version="$3"
    shift 3
    
    # Parse optional parameters
    local binary_paths=""
    local build_dir=""
    local source_repo=""
    local dependencies=""
    local installed_by_bundle=""
    local user_requested="true"
    local config_files=""
    local system_packages=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --binary-paths)
                binary_paths="$2"
                shift 2
                ;;
            --build-dir)
                build_dir="$2"
                shift 2
                ;;
            --source-repo)
                source_repo="$2"
                shift 2
                ;;
            --dependencies)
                dependencies="$2"
                shift 2
                ;;
            --installed-by-bundle)
                installed_by_bundle="$2"
                user_requested="false"
                shift 2
                ;;
            --config-files)
                config_files="$2"
                shift 2
                ;;
            --system-packages)
                system_packages="$2"
                shift 2
                ;;
            --not-user-requested)
                user_requested="false"
                shift
                ;;
            *)
                log_warning "Unknown tracking option: $1"
                shift
                ;;
        esac
    done
    
    # Auto-detect binary paths if not provided
    if [[ -z "$binary_paths" ]]; then
        binary_paths=$(detect_binary_paths "$tool_name")
    fi
    
    # Create tracking config
    local tracking_args=(
        "$tool_name"
        "$method"
        "$version"
    )
    
    if [[ -n "$binary_paths" ]]; then
        tracking_args+=(--binary-paths "$binary_paths")
    fi
    
    if [[ -n "$build_dir" ]]; then
        tracking_args+=(--build-dir "$build_dir")
    fi
    
    if [[ -n "$source_repo" ]]; then
        tracking_args+=(--source-repo "$source_repo")
    fi
    
    if [[ -n "$dependencies" ]]; then
        tracking_args+=(--dependencies "$dependencies")
    fi
    
    if [[ -n "$installed_by_bundle" ]]; then
        tracking_args+=(--installed-by-bundle "$installed_by_bundle")
    fi
    
    if [[ -n "$config_files" ]]; then
        tracking_args+=(--config-files "$config_files")
    fi
    
    if [[ -n "$system_packages" ]]; then
        tracking_args+=(--system-packages "$system_packages")
    fi
    
    if [[ "$user_requested" == "false" ]]; then
        tracking_args+=(--not-user-requested)
    fi
    
    # Call Go-based tracker
    if ! "$GEARBOX_BIN/orchestrator" track-installation "${tracking_args[@]}"; then
        log_error "Failed to track installation of $tool_name"
        return 1
    fi
    
    log_debug "Tracked installation: $tool_name ($method)"
    return 0
}

# Track bundle installation
# Usage: track_bundle BUNDLE_NAME TOOLS... [--not-user-requested]
track_bundle() {
    local bundle_name="$1"
    shift
    
    local user_requested="true"
    local tools=()
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --not-user-requested)
                user_requested="false"
                shift
                ;;
            *)
                tools+=("$1")
                shift
                ;;
        esac
    done
    
    # Convert tools array to comma-separated string
    local tools_str
    tools_str=$(IFS=','; echo "${tools[*]}")
    
    # Track bundle
    local args=(
        "$bundle_name"
        "$tools_str"
    )
    
    if [[ "$user_requested" == "false" ]]; then
        args+=(--not-user-requested)
    fi
    
    if ! "$GEARBOX_BIN/orchestrator" track-bundle "${args[@]}"; then
        log_error "Failed to track bundle installation: $bundle_name"
        return 1
    fi
    
    log_debug "Tracked bundle installation: $bundle_name"
    return 0
}

# Check if tool is already tracked
# Usage: is_tool_tracked TOOL_NAME
is_tool_tracked() {
    local tool_name="$1"
    
    "$GEARBOX_BIN/orchestrator" is-tracked "$tool_name" >/dev/null 2>&1
}

# Detect pre-existing tool installation
# Usage: detect_preexisting TOOL_NAME [BINARY_NAME]
detect_preexisting() {
    local tool_name="$1"
    local binary_name="${2:-$tool_name}"
    
    # Check if already tracked
    if is_tool_tracked "$tool_name"; then
        return 1 # Already tracked, not pre-existing
    fi
    
    # Check if binary exists
    if command -v "$binary_name" >/dev/null 2>&1; then
        local binary_path
        binary_path=$(command -v "$binary_name")
        
        # Get version if possible
        local version="unknown"
        if command -v "$binary_name" >/dev/null 2>&1; then
            # Try common version flags
            for flag in "--version" "-V" "-v" "version"; do
                if version_output=$("$binary_name" "$flag" 2>/dev/null | head -1); then
                    version="$version_output"
                    break
                fi
            done
        fi
        
        # Track as pre-existing
        if "$GEARBOX_BIN/orchestrator" track-preexisting "$tool_name" "$binary_path" "$version"; then
            log_info "Detected pre-existing installation: $tool_name ($binary_path)"
            return 0
        fi
    fi
    
    return 1 # Not found or failed to track
}

# Auto-detect binary paths for a tool
# Usage: detect_binary_paths TOOL_NAME
detect_binary_paths() {
    local tool_name="$1"
    local paths=()
    
    # Check main tool name
    if command -v "$tool_name" >/dev/null 2>&1; then
        paths+=($(command -v "$tool_name"))
    fi
    
    # Check common aliases based on tool metadata
    case "$tool_name" in
        "ripgrep")
            if command -v "rg" >/dev/null 2>&1; then
                paths+=($(command -v "rg"))
            fi
            ;;
        "bottom")
            if command -v "btm" >/dev/null 2>&1; then
                paths+=($(command -v "btm"))
            fi
            ;;
        "difftastic")
            if command -v "difft" >/dev/null 2>&1; then
                paths+=($(command -v "difft"))
            fi
            ;;
        "tealdeer")
            if command -v "tldr" >/dev/null 2>&1; then
                paths+=($(command -v "tldr"))
            fi
            ;;
        "7zip")
            if command -v "7zz" >/dev/null 2>&1; then
                paths+=($(command -v "7zz"))
            fi
            ;;
        "imagemagick")
            for cmd in "convert" "identify" "mogrify" "montage"; do
                if command -v "$cmd" >/dev/null 2>&1; then
                    paths+=($(command -v "$cmd"))
                fi
            done
            ;;
        "ffmpeg")
            for cmd in "ffmpeg" "ffprobe" "ffplay"; do
                if command -v "$cmd" >/dev/null 2>&1; then
                    paths+=($(command -v "$cmd"))
                fi
            done
            ;;
    esac
    
    # Return paths as comma-separated string
    if [[ ${#paths[@]} -gt 0 ]]; then
        IFS=','
        echo "${paths[*]}"
        IFS=' '
    fi
}

# Get tool installation method based on binary path
# Usage: detect_installation_method BINARY_PATH
detect_installation_method() {
    local binary_path="$1"
    
    case "$binary_path" in
        */usr/local/bin/*)
            echo "source_build"
            ;;
        */.cargo/bin/*)
            echo "cargo_install"
            ;;
        */go/bin/*|*/.local/go/bin/*)
            echo "go_install"
            ;;
        */usr/bin/*|*/bin/*)
            echo "system_package"
            ;;
        */.local/bin/*)
            echo "pipx"
            ;;
        */node_modules/.bin/*)
            echo "npm_global"
            ;;
        *)
            echo "manual_download"
            ;;
    esac
}

# Track installation wrapper for existing scripts
# This function provides a simple interface for existing installation scripts
# Usage: track_tool_installation TOOL_NAME [OPTIONS...]
track_tool_installation() {
    local tool_name="$1"
    shift
    
    # Auto-detect installation method and version
    local binary_paths
    binary_paths=$(detect_binary_paths "$tool_name")
    
    if [[ -z "$binary_paths" ]]; then
        log_warning "Could not detect binary paths for $tool_name"
        return 1
    fi
    
    # Use first binary path to detect method
    local first_binary
    first_binary=$(echo "$binary_paths" | cut -d',' -f1)
    
    local method
    method=$(detect_installation_method "$first_binary")
    
    # Try to get version
    local version="unknown"
    if [[ -x "$first_binary" ]]; then
        for flag in "--version" "-V" "-v" "version"; do
            if version_output=$("$first_binary" "$flag" 2>/dev/null | head -1); then
                version="$version_output"
                break
            fi
        done
    fi
    
    # Track installation
    track_installation "$tool_name" "$method" "$version" \
        --binary-paths "$binary_paths" \
        "$@"
}

# Initialize tracking system
# Usage: init_tracking
init_tracking() {
    # Ensure manifest directory exists
    mkdir -p "$GEARBOX_MANIFEST_DIR"
    
    # Initialize manifest if it doesn't exist
    if [[ ! -f "$GEARBOX_MANIFEST_FILE" ]]; then
        "$GEARBOX_BIN/orchestrator" init-manifest
    fi
}

# Check if tracking is enabled
# Usage: is_tracking_enabled
is_tracking_enabled() {
    [[ "${GEARBOX_DISABLE_TRACKING:-false}" != "true" ]]
}

# Wrapper to conditionally track installations
# Usage: maybe_track_installation TOOL_NAME METHOD VERSION [OPTIONS...]
maybe_track_installation() {
    if is_tracking_enabled; then
        track_installation "$@"
    else
        log_debug "Installation tracking disabled"
    fi
}

# Get gearbox binary path
GEARBOX_BIN="${GEARBOX_BIN:-$(dirname "$(dirname "${BASH_SOURCE[0]}")")/../bin}"

# Validate that orchestrator binary exists
if [[ ! -x "$GEARBOX_BIN/orchestrator" ]]; then
    log_warning "Orchestrator binary not found at $GEARBOX_BIN/orchestrator"
    log_warning "Installation tracking will be disabled"
    GEARBOX_DISABLE_TRACKING=true
fi