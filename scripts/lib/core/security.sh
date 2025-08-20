#!/bin/bash
#
# @file lib/core/security.sh
# @brief Security functions and safe operations for gearbox
# @description
#   Provides security-focused functions including root prevention,
#   secure downloads, safe command execution, and path validation.
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_SECURITY_LOADED:-}" ]] && return 0
readonly GEARBOX_SECURITY_LOADED=1

# =============================================================================
# SECURITY FUNCTIONS
# =============================================================================

# @function ensure_not_root
# @brief Ensure script is not running as root for security
ensure_not_root() {
    if [[ $EUID -eq 0 ]]; then
        error "This script should not be run as root for security reasons"
    fi
}

# @function check_tool_installed
# @brief Check if tool is already installed and handle force flag
# @param $1 Tool name
# @param $2 Force install flag (default: false)
check_tool_installed() {
    local tool_name="$1"
    local force_install="${2:-false}"
    
    if command -v "$tool_name" &> /dev/null && [[ "$force_install" != "true" ]]; then
        log "$tool_name already installed. Use --force to reinstall."
        exit 0
    fi
}

# @function execute_command_safely
# @brief Safely execute commands without eval, preventing injection attacks
# @param $@ Command and arguments to execute
execute_command_safely() {
    local cmd=("$@")
    
    # Validate that we have at least one argument
    [[ ${#cmd[@]} -eq 0 ]] && error "No command specified"
    
    # Log the command being executed
    debug "Executing command: ${cmd[*]}"
    
    # Execute safely without eval
    "${cmd[@]}"
}

# @function secure_download
# @brief Securely download file with checksum verification
# @param $1 URL to download
# @param $2 Output file path
# @param $3 Expected SHA256 checksum (optional)
# @param $4 Maximum file size in bytes (optional, default 100MB)
secure_download() {
    local url="$1"
    local output_file="$2"
    local expected_checksum="${3:-}"
    local max_size="${4:-104857600}"  # 100MB default
    
    [[ -z "$url" ]] && error "Download URL not specified"
    [[ -z "$output_file" ]] && error "Output file not specified"
    
    # Validate URL format
    [[ "$url" =~ ^https?:// ]] || error "Invalid URL format: $url (must start with http:// or https://)"
    
    # Validate output path
    validate_file_path "$(basename "$output_file")"
    
    log "Securely downloading from $url..."
    
    # Download with size limit and timeout
    curl --fail --silent --show-error --location \
         --max-filesize "$max_size" \
         --max-time 300 \
         --proto '=https,http' \
         --tlsv1.2 \
         --output "$output_file" \
         "$url" || error "Failed to download $url"
    
    # Verify file was actually downloaded
    [[ -f "$output_file" ]] || error "Downloaded file not found: $output_file"
    
    # Checksum verification if provided
    if [[ -n "$expected_checksum" ]]; then
        log "Verifying checksum..."
        local actual_checksum
        actual_checksum=$(sha256sum "$output_file" | cut -d' ' -f1)
        
        if [[ "$actual_checksum" != "$expected_checksum" ]]; then
            rm -f "$output_file"  # Remove corrupted download
            error "Checksum verification failed! Expected: $expected_checksum, Got: $actual_checksum"
        fi
        
        success "Checksum verification passed"
    else
        warning "No checksum provided - download integrity not verified"
    fi
    
    success "Secure download completed: $output_file"
}

# @function secure_download_and_pipe
# @brief Securely download and pipe to command (for install scripts)
# @param $1 URL to download
# @param $2 Command to pipe to
# @param $3 Expected content pattern (optional)
secure_download_and_pipe() {
    local url="$1"
    local pipe_command="$2"
    local expected_pattern="${3:-}"
    
    [[ -z "$url" ]] && error "Download URL not specified"
    [[ -z "$pipe_command" ]] && error "Pipe command not specified"
    
    # Validate URL
    [[ "$url" =~ ^https:// ]] || error "Only HTTPS URLs allowed for piped downloads: $url"
    
    # Create temporary file for inspection
    local temp_file
    temp_file=$(mktemp)
    
    # Download to temp file first
    log "Downloading script from $url for inspection..."
    curl --fail --silent --show-error --location \
         --max-filesize 10485760 \
         --max-time 60 \
         --proto '=https' \
         --tlsv1.2 \
         --output "$temp_file" \
         "$url" || {
        rm -f "$temp_file"
        error "Failed to download $url"
    }
    
    # Basic content validation
    if [[ -n "$expected_pattern" ]]; then
        if ! grep -q "$expected_pattern" "$temp_file"; then
            rm -f "$temp_file"
            error "Downloaded script does not contain expected pattern: $expected_pattern"
        fi
    fi
    
    # Check for suspicious content
    if grep -q -E "(curl.*\|.*sh|wget.*\|.*sh|rm.*-rf.*\$|>\s*/dev/)" "$temp_file"; then
        warning "Downloaded script contains potentially dangerous patterns"
        warning "Please review the script before continuing: $temp_file"
        read -p "Continue anyway? [y/N]: " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            rm -f "$temp_file"
            error "Installation cancelled by user"
        fi
    fi
    
    # Execute the downloaded script
    log "Executing downloaded script..."
    $pipe_command < "$temp_file" || {
        rm -f "$temp_file"
        error "Script execution failed"
    }
    
    # Cleanup
    rm -f "$temp_file"
    success "Script executed successfully"
}

# @function secure_execute_script
# @brief Execute script with security checks
# @param $1 Script path
# @param $@ Additional arguments
secure_execute_script() {
    local script_path="$1"
    shift
    local args=("$@")
    
    # Validate script exists
    [[ -f "$script_path" ]] || error "Script not found: $script_path"
    
    # Validate script is executable
    [[ -x "$script_path" ]] || error "Script is not executable: $script_path"
    
    # Basic security check - no path traversal
    validate_file_path "$script_path"
    
    log "Executing script: $script_path ${args[*]}"
    
    # Execute safely without eval
    "$script_path" "${args[@]}"
}

# @function safe_sudo_copy
# @brief Safely copy file with sudo, validating paths
# @param $1 Source file path
# @param $2 Destination file path
safe_sudo_copy() {
    local src_file="$1"
    local dest_file="$2"
    
    # Validate inputs
    [[ -z "$src_file" ]] && error "Source file not specified"
    [[ -z "$dest_file" ]] && error "Destination file not specified"
    
    # Validate source file exists
    [[ -f "$src_file" ]] || error "Source file does not exist: $src_file"
    
    # Security: Ensure absolute paths
    [[ "$src_file" =~ ^/ ]] || error "Source must be absolute path: $src_file"
    [[ "$dest_file" =~ ^/ ]] || error "Destination must be absolute path: $dest_file"
    
    # Validate paths for security
    validate_file_path "$src_file"
    validate_file_path "$dest_file"
    
    # Ensure destination directory exists
    local dest_dir
    dest_dir=$(dirname "$dest_file")
    if [[ ! -d "$dest_dir" ]]; then
        warning "Destination directory does not exist, creating: $dest_dir"
        sudo mkdir -p "$dest_dir" || error "Failed to create destination directory: $dest_dir"
    fi
    
    # Perform the copy
    log "Copying $src_file to $dest_file (with sudo)"
    if ! sudo cp "$src_file" "$dest_file"; then
        error "Failed to copy $src_file to $dest_file"
    fi
    
    # Set appropriate permissions
    log "Setting permissions on $dest_file"
    
    # Determine if this is a binary (executable) or config file
    if [[ "$dest_file" =~ /usr/local/bin/ ]] || [[ -x "$src_file" ]]; then
        # Executable binary
        sudo chmod 755 "$dest_file" || error "Failed to set permissions on $dest_file"
    else
        # Regular file
        sudo chmod 644 "$dest_file" || error "Failed to set permissions on $dest_file"
    fi
    
    success "Successfully copied and set permissions: $dest_file"
}