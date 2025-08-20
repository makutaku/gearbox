#!/bin/bash
# Generated shell helpers for gearbox configuration
# Generated from: config/tools.json

# Configuration manager path
CONFIG_MANAGER="/home/rosantos/gearbox/bin/config-manager"
CONFIG_FILE="config/tools.json"

# Get build flag for a tool and build type
get_build_flag() {
    local tool="$1"
    local build_type="${2:-standard}"
    
    "$CONFIG_MANAGER" --config="$CONFIG_FILE" build-flag "$tool" --build-type="$build_type" 2>/dev/null
}

# List all available tools
list_tools() {
    "$CONFIG_MANAGER" --config="$CONFIG_FILE" list "$@"
}

# Validate configuration
validate_config() {
    "$CONFIG_MANAGER" --config="$CONFIG_FILE" validate
}

# Get tool info as JSON (requires jq)
get_tool_info() {
    local tool="$1"
    [[ -f "$CONFIG_FILE" ]] && jq -r ".tools[] | select(.name == \"$tool\")" "$CONFIG_FILE"
}

# Get all tools in category
get_tools_in_category() {
    local category="$1"
    [[ -f "$CONFIG_FILE" ]] && jq -r ".tools[] | select(.category == \"$category\") | .name" "$CONFIG_FILE"
}

# Validate tool name (generated from config)
validate_tool_name() {
    local tool="$1"
    case "$tool" in
        7zip|bandwhich|bat|bottom|choose|delta|difftastic|dust|eza|fclones|fd|ffmpeg|fzf|gh|hyperfine|imagemagick|jq|lazygit|procs|ripgrep|ruff|sd|serena|starship|tealdeer|tokei|uv|xsv|yazi|zoxide)
            return 0
            ;;
        *)
            echo "ERROR: Invalid tool name: '$tool'. Use 'gearbox list' to see available tools." >&2
            return 1
            ;;
    esac
}

