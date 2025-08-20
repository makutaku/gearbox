#!/bin/bash
#
# @file lib/core/validation.sh
# @brief Input validation and security functions for gearbox
# @description
#   Provides input validation, path sanitization, version comparison,
#   and security checks to prevent common vulnerabilities.
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_VALIDATION_LOADED:-}" ]] && return 0
readonly GEARBOX_VALIDATION_LOADED=1

# =============================================================================
# VERSION COMPARISON UTILITIES
# =============================================================================

# @function version_compare
# @brief Compare two semantic version strings
# @description
#   Compares version strings in the format X.Y.Z where X, Y, Z are integers.
#   Returns 0 if first version >= second version, 1 otherwise.
#
# @param $1 First version string (e.g., "1.2.3")
# @param $2 Second version string (e.g., "1.1.0")
# @return 0 if $1 >= $2, 1 if $1 < $2
#
# @example
#   if version_compare "1.2.3" "1.1.0"; then
#       echo "Version 1.2.3 is newer or equal to 1.1.0"
#   fi
version_compare() {
    local ver1="$1"
    local ver2="$2"
    
    # Handle identical versions
    [[ "$ver1" == "$ver2" ]] && return 0
    
    # Split versions into arrays
    local IFS='.'
    local i ver1_parts=($ver1) ver2_parts=($ver2)
    
    # Pad shorter version with zeros
    local max_parts=$((${#ver1_parts[@]} > ${#ver2_parts[@]} ? ${#ver1_parts[@]} : ${#ver2_parts[@]}))
    
    for ((i=${#ver1_parts[@]}; i<max_parts; i++)); do
        ver1_parts[i]=0
    done
    
    for ((i=${#ver2_parts[@]}; i<max_parts; i++)); do
        ver2_parts[i]=0
    done
    
    # Compare each part
    for ((i=0; i<max_parts; i++)); do
        local part1="${ver1_parts[i]:-0}"
        local part2="${ver2_parts[i]:-0}"
        
        # Force numeric comparison with 10# prefix
        if ((10#$part1 > 10#$part2)); then
            return 0  # ver1 > ver2
        elif ((10#$part1 < 10#$part2)); then
            return 1  # ver1 < ver2
        fi
        # Continue if parts are equal
    done
    
    return 0  # Versions are equal
}

# =============================================================================
# INPUT VALIDATION FUNCTIONS
# =============================================================================

# @function validate_build_type
# @brief Validate build type parameter
# @param $1 Build type to validate
# @return 0 if valid, exits with error if invalid
validate_build_type() {
    local build_type="$1"
    case "$build_type" in
        minimal|standard|maximum|debug|release|optimized)
            return 0
            ;;
        *)
            error "Invalid build type: '$build_type'. Valid options: minimal, standard, maximum, debug, release, optimized"
            ;;
    esac
}

# @function validate_tool_name
# @brief Validate tool name for security
# @param $1 Tool name to validate
# @return 0 if valid, exits with error if invalid
validate_tool_name() {
    local tool_name="$1"
    
    # Check for empty or null
    [[ -z "$tool_name" ]] && error "Tool name cannot be empty"
    
    # Check for valid characters (alphanumeric, hyphen, underscore)
    [[ "$tool_name" =~ ^[a-zA-Z0-9_-]+$ ]] || error "Invalid tool name: '$tool_name'. Only alphanumeric, hyphen, and underscore allowed."
    
    # Check length (reasonable bounds)
    local len=${#tool_name}
    [[ $len -lt 2 || $len -gt 50 ]] && error "Tool name length must be between 2 and 50 characters"
    
    return 0
}

# @function validate_file_path
# @brief Validate file path for security (prevent path traversal)
# @param $1 File path to validate
# @return 0 if valid, exits with error if invalid
validate_file_path() {
    local file_path="$1"
    
    # Check for path traversal attempts (but allow absolute paths)
    [[ "$file_path" =~ \.\./ ]] && error "Invalid file path: '$file_path' (path traversal detected)"
    
    # Check for null bytes (disabled temporarily)
    # [[ "$file_path" == *$'\0'* ]] && error "Invalid file path: contains null byte"
    
    return 0
}

# @function validate_url
# @brief Validate URL format and protocol
# @param $1 URL to validate
# @return 0 if valid, 1 if invalid
validate_url() {
    local url="$1"
    
    # Basic URL validation - must start with http:// or https://
    if [[ "$url" =~ ^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$ ]]; then
        return 0
    else
        return 1
    fi
}

# @function sanitize_filename
# @brief Sanitize filename for safe filesystem operations
# @param $1 Filename to sanitize
# @return Sanitized filename on stdout
sanitize_filename() {
    local filename="$1"
    
    # Remove path separators and dangerous characters
    filename="${filename//[\/\\:*?\"<>|]/_}"
    
    # Remove leading/trailing dots and spaces
    filename="${filename#"${filename%%[![:space:].]}"}"
    filename="${filename%"${filename##*[![:space:].])}"}"
    
    # Ensure filename is not empty
    [[ -z "$filename" ]] && filename="sanitized_file"
    
    echo "$filename"
}

# @function ensure_directory
# @brief Safely create directory with validation
# @param $1 Directory path to create
ensure_directory() {
    local dir_path="$1"
    
    # Validate path
    validate_file_path "$dir_path"
    
    # Create directory if it doesn't exist
    if [[ ! -d "$dir_path" ]]; then
        mkdir -p "$dir_path" || error "Failed to create directory: $dir_path"
    fi
}

# @function create_temp_dir
# @brief Create temporary directory safely
create_temp_dir() {
    local temp_dir
    temp_dir=$(mktemp -d) || error "Failed to create temporary directory"
    echo "$temp_dir"
}

# @function create_temp_file
# @brief Create temporary file safely
# @param $1 Optional file prefix
create_temp_file() {
    local prefix="${1:-gearbox}"
    local temp_file
    temp_file=$(mktemp "${TMPDIR:-/tmp}/${prefix}.XXXXXX") || error "Failed to create temporary file"
    echo "$temp_file"
}