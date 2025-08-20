#!/bin/bash
#
# @file lib/config.sh
# @brief Simple Configuration Management for Gearbox
# @description Handles ~/.gearboxrc configuration file and user preferences

# Simple logging functions
log() { echo "$1"; }
error() { echo "ERROR: $1" >&2; exit 1; }
warning() { echo "WARNING: $1" >&2; }
success() { echo "SUCCESS: $1"; }

# Configuration file location
GEARBOX_CONFIG_FILE="$HOME/.gearboxrc"

# Default configuration values
declare -A DEFAULT_CONFIG=(
    ["DEFAULT_BUILD_TYPE"]="standard"
    ["MAX_PARALLEL_JOBS"]="auto"
    ["CACHE_ENABLED"]="true"
    ["CACHE_MAX_AGE_DAYS"]="7"
    ["AUTO_UPDATE_REPOS"]="true"
    ["INSTALL_MISSING_DEPS"]="true"
    ["SKIP_TESTS_BY_DEFAULT"]="false"
    ["VERBOSE_OUTPUT"]="false"
    ["SHELL_INTEGRATION"]="true"
    ["BACKUP_BEFORE_INSTALL"]="true"
)

# Load configuration from file
load_config() {
    local config_file="${1:-$GEARBOX_CONFIG_FILE}"
    
    # Initialize with defaults
    declare -gA GEARBOX_CONFIG
    for key in "${!DEFAULT_CONFIG[@]}"; do
        GEARBOX_CONFIG["$key"]="${DEFAULT_CONFIG[$key]}"
    done
    
    # Load user configuration if it exists
    if [[ -f "$config_file" ]]; then
        while IFS='=' read -r key value; do
            # Skip empty lines and comments
            [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
            
            # Remove quotes from value
            value="${value%\"}"
            value="${value#\"}"
            value="${value%\'}"
            value="${value#\'}"
            
            # Set the configuration value
            if [[ -n "${DEFAULT_CONFIG[$key]:-}" ]]; then
                GEARBOX_CONFIG["$key"]="$value"
            fi
        done < <(grep -v '^[[:space:]]*#' "$config_file" | grep '=')
    fi
    
    return 0
}

# Get configuration value
get_config() {
    local key="$1"
    local default_value="$2"
    
    if [[ -n "${GEARBOX_CONFIG[$key]:-}" ]]; then
        echo "${GEARBOX_CONFIG[$key]}"
    elif [[ -n "$default_value" ]]; then
        echo "$default_value"
    else
        echo "${DEFAULT_CONFIG[$key]:-}"
    fi
}

# Set configuration value
set_config() {
    local key="$1"
    local value="$2"
    
    if [[ -z "${DEFAULT_CONFIG[$key]:-}" ]]; then
        error "Unknown configuration key: $key"
        return 1
    fi
    
    GEARBOX_CONFIG["$key"]="$value"
    return 0
}

# Save configuration to file
save_config() {
    local config_file="${1:-$GEARBOX_CONFIG_FILE}"
    
    echo "Saving configuration to $config_file"
    
    # Create backup if file exists
    if [[ -f "$config_file" ]]; then
        cp "$config_file" "${config_file}.backup-$(date +%Y%m%d-%H%M%S)" 2>/dev/null || true
    fi
    
    # Create directory if it doesn't exist
    local config_dir
    config_dir=$(dirname "$config_file")
    [[ ! -d "$config_dir" ]] && mkdir -p "$config_dir"
    
    # Generate configuration file
    cat > "$config_file" << EOF
# Gearbox Configuration File
# Generated on $(date)

EOF
    
    # Write configuration values
    for key in "${!DEFAULT_CONFIG[@]}"; do
        local value="${GEARBOX_CONFIG[$key]:-${DEFAULT_CONFIG[$key]}}"
        echo "$key=\"$value\"" >> "$config_file"
    done
    
    echo "Configuration saved to $config_file"
    return 0
}

# Show current configuration
show_config() {
    echo
    echo "Current Gearbox Configuration:"
    echo
    
    for key in "${!DEFAULT_CONFIG[@]}"; do
        local value="${GEARBOX_CONFIG[$key]:-${DEFAULT_CONFIG[$key]}}"
        local is_default=""
        
        if [[ "$value" == "${DEFAULT_CONFIG[$key]}" ]]; then
            is_default=" (default)"
        fi
        
        printf "  %-25s = %-10s %s\n" "$key" "$value" "$is_default"
    done
    
    echo
    echo "Configuration file: $GEARBOX_CONFIG_FILE"
    if [[ -f "$GEARBOX_CONFIG_FILE" ]]; then
        echo "File exists: $(ls -la "$GEARBOX_CONFIG_FILE" | awk '{print $5 " bytes, modified " $6 " " $7 " " $8}')"
    else
        echo "File does not exist (using defaults)"
    fi
}

# Apply configuration to environment
apply_config() {
    export GEARBOX_DEFAULT_BUILD_TYPE="$(get_config DEFAULT_BUILD_TYPE)"
    export GEARBOX_MAX_PARALLEL_JOBS="$(get_config MAX_PARALLEL_JOBS)"
    export GEARBOX_CACHE_ENABLED="$(get_config CACHE_ENABLED)"
    export GEARBOX_AUTO_UPDATE_REPOS="$(get_config AUTO_UPDATE_REPOS)"
    export GEARBOX_INSTALL_MISSING_DEPS="$(get_config INSTALL_MISSING_DEPS)"
    export GEARBOX_SKIP_TESTS_BY_DEFAULT="$(get_config SKIP_TESTS_BY_DEFAULT)"
    export GEARBOX_SHELL_INTEGRATION="$(get_config SHELL_INTEGRATION)"
    export GEARBOX_BACKUP_BEFORE_INSTALL="$(get_config BACKUP_BEFORE_INSTALL)"
}

# Initialize configuration system
init_config() {
    load_config
    apply_config
}