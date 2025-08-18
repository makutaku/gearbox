#!/bin/bash
#
# @file lib/common.sh
# @brief Shared library for all gearbox installation scripts
# @description
#   Central repository for all shared functions, utilities, and patterns
#   used across the Essential Tools Installer. This eliminates code
#   duplication and ensures consistency across all installation scripts.
#
# @author Essential Tools Installer Team
# @version 1.0.0
# @since 2024-01-01
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_COMMON_LOADED:-}" ]] && return 0
readonly GEARBOX_COMMON_LOADED=1

# =============================================================================
# CONFIGURATION AND INITIALIZATION
# =============================================================================

# Find the script directory and load configuration
if [[ -z "${GEARBOX_LIB_DIR:-}" ]]; then
    GEARBOX_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    GEARBOX_REPO_DIR="$(dirname "$GEARBOX_LIB_DIR")"
    readonly GEARBOX_LIB_DIR GEARBOX_REPO_DIR
fi

# Configuration now handled by lib/config.sh

# Load configuration management library
if [[ -f "$GEARBOX_LIB_DIR/config.sh" ]]; then
    source "$GEARBOX_LIB_DIR/config.sh"
    # Initialize configuration system
    init_config
fi

# =============================================================================
# LOGGING AND OUTPUT FUNCTIONS
# =============================================================================

# @function log
# @brief Standard informational logging with timestamp
# @param $1 Message to log
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

# @function error
# @brief Error logging that exits the script
# @param $1 Error message
# @exit 1 Always exits with error code 1
error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

# @function success
# @brief Success message logging
# @param $1 Success message
success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# @function warning
# @brief Warning message logging (non-fatal)
# @param $1 Warning message
warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# @function debug
# @brief Debug logging (only shown if DEBUG=true)
# @param $1 Debug message
debug() {
    if [[ "${DEBUG:-false}" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1" >&2
    fi
}

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

# =============================================================================
# DEPENDENCY MANAGEMENT FUNCTIONS
# =============================================================================

# @function install_rust_if_needed
# @brief Install or update Rust to meet minimum version requirement
# @param $1 Minimum required Rust version (e.g., "1.88.0")
install_rust_if_needed() {
    local min_version="$1"
    [[ -z "$min_version" ]] && error "Minimum Rust version not specified"
    
    if command -v rustc &> /dev/null; then
        local current_version=$(rustc --version | grep -oP '\d+\.\d+\.\d+' | head -1)
        log "Found Rust version: $current_version"
        
        if version_compare "$current_version" "$min_version"; then
            log "Rust version is sufficient (>= $min_version)"
            return 0
        else
            warning "Rust version $current_version is below required $min_version"
        fi
    else
        log "Rust not found, installing..."
    fi
    
    # Install/update Rust
    log "Installing Rust $min_version or later..."
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    
    # Source environment
    [[ -f ~/.cargo/env ]] && source ~/.cargo/env
    
    # Verify installation
    if command -v rustc &> /dev/null; then
        local installed_version=$(rustc --version | grep -oP '\d+\.\d+\.\d+' | head -1)
        success "Rust $installed_version installed successfully"
    else
        error "Rust installation failed"
    fi
}

# @function install_go_if_needed
# @brief Install Go to meet minimum version requirement
# @param $1 Minimum required Go version (e.g., "1.23.4")
install_go_if_needed() {
    local min_version="$1"
    [[ -z "$min_version" ]] && error "Minimum Go version not specified"
    
    if command -v go &> /dev/null; then
        local current_version=$(go version | grep -oP '\d+\.\d+\.\d+' | head -1)
        log "Found Go version: $current_version"
        
        if version_compare "$current_version" "$min_version"; then
            log "Go version is sufficient (>= $min_version)"
            return 0
        else
            warning "Go version $current_version is below required $min_version"
        fi
    else
        log "Go not found, installing..."
    fi
    
    # Install Go via system package manager
    log "Installing Go $min_version or later..."
    sudo apt update
    sudo apt install -y golang-go
    
    # Verify installation
    if command -v go &> /dev/null; then
        local installed_version=$(go version | grep -oP '\d+\.\d+\.\d+' | head -1)
        success "Go $installed_version installed successfully"
    else
        error "Go installation failed"
    fi
}

# @function install_python_if_needed
# @brief Install Python to meet minimum version requirement
# @param $1 Minimum required Python version (e.g., "3.11.0")
install_python_if_needed() {
    local min_version="$1"
    [[ -z "$min_version" ]] && error "Minimum Python version not specified"
    
    local python_cmd=""
    
    # Try different Python commands
    for cmd in python3.11 python3 python; do
        if command -v "$cmd" &> /dev/null; then
            local current_version=$($cmd --version 2>&1 | grep -oP '\d+\.\d+\.\d+')
            if version_compare "$current_version" "$min_version"; then
                log "Found suitable Python: $cmd version $current_version"
                python_cmd="$cmd"
                break
            fi
        fi
    done
    
    if [[ -z "$python_cmd" ]]; then
        log "Installing Python $min_version or later..."
        sudo apt update
        sudo apt install -y python3 python3-pip python3-venv
        
        # Verify installation
        if command -v python3 &> /dev/null; then
            local installed_version=$(python3 --version 2>&1 | grep -oP '\d+\.\d+\.\d+')
            success "Python $installed_version installed successfully"
        else
            error "Python installation failed"
        fi
    fi
}

# @function install_build_tools
# @brief Install essential build tools and dependencies
install_build_tools() {
    log "Installing essential build tools..."
    
    sudo apt update
    sudo apt install -y \
        build-essential \
        git \
        curl \
        wget \
        make \
        autoconf \
        automake \
        libtool \
        pkg-config \
        ca-certificates
    
    success "Essential build tools installed"
}

# =============================================================================
# REPOSITORY MANAGEMENT FUNCTIONS
# =============================================================================

# @function clone_or_update_repo
# @brief Clone repository or update if it exists
# @param $1 Repository URL
# @param $2 Target directory
# @param $3 Branch name (optional, defaults to main)
clone_or_update_repo() {
    local repo_url="$1"
    local target_dir="$2"
    local branch="${3:-main}"
    
    [[ -z "$repo_url" ]] && error "Repository URL not specified"
    [[ -z "$target_dir" ]] && error "Target directory not specified"
    
    validate_file_path "$target_dir"
    
    if [[ -d "$target_dir" ]]; then
        log "Updating existing repository in $target_dir..."
        cd "$target_dir"
        git fetch origin
        git reset --hard "origin/$branch"
    else
        log "Cloning repository from $repo_url..."
        git clone --branch "$branch" "$repo_url" "$target_dir"
        cd "$target_dir"
    fi
    
    success "Repository ready in $target_dir"
}

# =============================================================================
# INSTALLATION UTILITIES
# =============================================================================

# @function check_existing_installation
# @brief Check if tool is already installed
# @param $1 Tool binary name
# @param $2 Force reinstall flag (true/false)
# @return 0 to continue installation, exits if already installed
check_existing_installation() {
    local tool_name="$1"
    local force_install="${2:-false}"
    
    validate_tool_name "$tool_name"
    
    if command -v "$tool_name" &> /dev/null && [[ "$force_install" != "true" ]]; then
        local version_info=""
        if "$tool_name" --version &> /dev/null; then
            version_info=" ($("$tool_name" --version 2>/dev/null | head -1))"
        fi
        log "$tool_name already installed$version_info. Use --force to reinstall."
        exit 0
    fi
}

# @function run_with_timeout
# @brief Run command with timeout
# @param $1 Timeout in seconds
# @param $@ Command and arguments to run
run_with_timeout() {
    local timeout_seconds="$1"
    shift
    
    timeout "$timeout_seconds" "$@"
}

# @function get_optimal_jobs
# @brief Calculate optimal number of parallel jobs for builds with adaptive memory monitoring
# @return Prints optimal job count
get_optimal_jobs() {
    # Check if user has configured a specific job limit
    local config_jobs="${GEARBOX_MAX_PARALLEL_JOBS:-auto}"
    
    if [[ "$config_jobs" =~ ^[1-9][0-9]*$ ]]; then
        # User specified a specific number
        echo "$config_jobs"
        return 0
    fi
    
    # Auto-calculate based on system resources with enhanced monitoring
    local cpu_cores
    cpu_cores=$(nproc)
    
    # Enhanced memory calculation with load monitoring
    local total_memory_gb available_memory_gb current_load
    total_memory_gb=$(free -g | awk '/^Mem:/{print $2}')
    available_memory_gb=$(free -g | awk '/^Mem:/{print $7}')
    
    # Get current system load (1-minute average)
    current_load=$(uptime | grep -o '[0-9][0-9]*\.[0-9][0-9]*' | head -1 2>/dev/null || echo "0.0")
    current_load_int=${current_load%.*}
    
    # Memory-based job calculation with safety margins
    local memory_per_job_gb=1
    if [[ $total_memory_gb -gt 16 ]]; then
        memory_per_job_gb=2  # Use more memory per job on high-memory systems
    elif [[ $total_memory_gb -lt 4 ]]; then
        memory_per_job_gb=1  # Conservative on low-memory systems
    fi
    
    local memory_limited_jobs=$((available_memory_gb / memory_per_job_gb))
    
    # Load-based adjustment: reduce parallelism if system is already busy
    local load_factor=1
    if [[ ${current_load_int:-0} -gt $((cpu_cores * 2)) ]]; then
        load_factor=2  # High load: halve parallelism
        log "High system load detected (${current_load}), reducing parallelism"
    elif [[ ${current_load_int:-0} -gt $cpu_cores ]]; then
        load_factor=1  # Moderate load: slight reduction
        memory_limited_jobs=$((memory_limited_jobs * 3 / 4))
    fi
    
    # Calculate optimal jobs considering CPU, memory, and load
    local cpu_limited_jobs=$((cpu_cores / load_factor))
    local optimal_jobs=$((cpu_limited_jobs < memory_limited_jobs ? cpu_limited_jobs : memory_limited_jobs))
    
    # Enforce minimum and maximum bounds
    optimal_jobs=$((optimal_jobs > 0 ? optimal_jobs : 1))
    optimal_jobs=$((optimal_jobs < 8 ? optimal_jobs : 8))  # Cap at 8 for stability
    
    log "System resources: ${cpu_cores} CPUs, ${available_memory_gb}GB available memory, load ${current_load}"
    log "Optimal parallel jobs: $optimal_jobs"
    
    echo "$optimal_jobs"
}

# =============================================================================
# ERROR HANDLING AND CLEANUP
# =============================================================================

# @function setup_error_cleanup
# @brief Set up error handling and cleanup on script exit
# @param $1 Optional cleanup function name
setup_error_cleanup() {
    local cleanup_function="${1:-cleanup_on_error}"
    
    # Set up trap for cleanup
    trap "$cleanup_function \$?" EXIT INT TERM
}

# @function cleanup_on_error
# @brief Default cleanup function for errors
# @param $1 Exit code
cleanup_on_error() {
    local exit_code="$1"
    
    if [[ $exit_code -ne 0 ]]; then
        warning "Operation failed with exit code $exit_code, performing cleanup..."
        
        # Execute rollback actions first
        execute_rollback
        
        # Clean up any temporary directories
        [[ -n "${TEMP_DIR:-}" && -d "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"
        
        # Additional cleanup can be added here
        debug "Cleanup completed"
    else
        # Success - clear rollback actions
        clear_rollback_actions
    fi
}

# =============================================================================
# ROLLBACK AND RECOVERY FUNCTIONS
# =============================================================================

# Global rollback state tracking
declare -a ROLLBACK_ACTIONS=()
declare -g ROLLBACK_ENABLED=true

# @function add_rollback_action
# @brief Add an action to the rollback stack
# @param $1 Rollback command to execute
# @description
#   Adds a command to the rollback stack that will be executed in reverse order
#   if rollback is triggered. Commands should be safe to execute multiple times.
#
# @example
#   add_rollback_action "rm -f /tmp/myfile"
#   add_rollback_action "systemctl stop myservice"
add_rollback_action() {
    local action="$1"
    [[ -z "$action" ]] && error "Rollback action cannot be empty"
    
    if [[ "$ROLLBACK_ENABLED" == "true" ]]; then
        ROLLBACK_ACTIONS+=("$action")
        debug "Added rollback action: $action"
    fi
}

# @function execute_rollback
# @brief Execute all rollback actions in reverse order
# @description
#   Executes all registered rollback actions in LIFO order (last added first).
#   Each action is executed with error handling to ensure all rollbacks run.
execute_rollback() {
    if [[ ${#ROLLBACK_ACTIONS[@]} -eq 0 ]]; then
        debug "No rollback actions to execute"
        return 0
    fi
    
    warning "Executing rollback actions..."
    
    # Execute rollback actions in reverse order
    for ((i=${#ROLLBACK_ACTIONS[@]}-1; i>=0; i--)); do
        local action="${ROLLBACK_ACTIONS[i]}"
        debug "Executing rollback: $action"
        
        # Execute with error handling - don't fail if rollback fails
        if eval "$action" 2>/dev/null; then
            debug "Rollback action succeeded: $action"
        else
            warning "Rollback action failed: $action"
        fi
    done
    
    success "Rollback completed"
    ROLLBACK_ACTIONS=()  # Clear the actions
}

# @function clear_rollback_actions
# @brief Clear all rollback actions without executing them
# @description
#   Clears the rollback stack without executing actions. Use this when
#   installation succeeds and rollback is no longer needed.
clear_rollback_actions() {
    debug "Clearing ${#ROLLBACK_ACTIONS[@]} rollback actions"
    ROLLBACK_ACTIONS=()
}

# @function disable_rollback
# @brief Disable rollback action recording
# @description
#   Disables the recording of new rollback actions. Existing actions
#   are preserved but no new ones will be added.
disable_rollback() {
    ROLLBACK_ENABLED=false
    debug "Rollback action recording disabled"
}

# @function enable_rollback
# @brief Enable rollback action recording
enable_rollback() {
    ROLLBACK_ENABLED=true
    debug "Rollback action recording enabled"
}

# @function backup_file
# @brief Create a backup of a file before modification
# @param $1 File path to backup
# @param $2 Optional backup suffix (default: .backup)
# @return 0 on success, 1 on failure
backup_file() {
    local file_path="$1"
    local backup_suffix="${2:-.backup}"
    
    [[ -z "$file_path" ]] && error "File path not specified"
    [[ ! -f "$file_path" ]] && return 1  # File doesn't exist, no backup needed
    
    local backup_path="${file_path}${backup_suffix}"
    
    if cp "$file_path" "$backup_path" 2>/dev/null; then
        debug "Created backup: $backup_path"
        add_rollback_action "restore_file_backup '$file_path' '$backup_suffix'"
        return 0
    else
        warning "Failed to create backup of $file_path"
        return 1
    fi
}

# @function restore_file_backup
# @brief Restore a file from its backup
# @param $1 Original file path
# @param $2 Optional backup suffix (default: .backup)
restore_file_backup() {
    local file_path="$1"
    local backup_suffix="${2:-.backup}"
    local backup_path="${file_path}${backup_suffix}"
    
    if [[ -f "$backup_path" ]]; then
        mv "$backup_path" "$file_path" 2>/dev/null
        debug "Restored file from backup: $file_path"
    else
        debug "No backup found for: $file_path"
    fi
}

# @function safe_install_binary
# @brief Safely install a binary with automatic rollback
# @param $1 Source binary path
# @param $2 Target installation path
safe_install_binary() {
    local source_path="$1"
    local target_path="$2"
    
    [[ -z "$source_path" ]] && error "Source path not specified"
    [[ -z "$target_path" ]] && error "Target path not specified"
    [[ ! -f "$source_path" ]] && error "Source binary not found: $source_path"
    
    # Backup existing binary if it exists
    if [[ -f "$target_path" ]]; then
        backup_file "$target_path"
    else
        # If file doesn't exist, add rollback to remove it
        add_rollback_action "rm -f '$target_path'"
    fi
    
    # Install the binary
    if install -m 755 "$source_path" "$target_path"; then
        success "Installed binary: $target_path"
        return 0
    else
        error "Failed to install binary: $target_path"
    fi
}

# =============================================================================
# SECURITY FUNCTIONS
# =============================================================================

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
    
    # Execute the piped command safely
    log "Executing: $pipe_command"
    # Split command into array for safe execution
    read -ra cmd_array <<< "$pipe_command"
    cat "$temp_file" | "${cmd_array[@]}" || {
        rm -f "$temp_file"
        error "Piped command failed: $pipe_command"
    }
    
    # Clean up
    rm -f "$temp_file"
    success "Secure download and pipe completed"
}

# @function secure_execute_script
# @brief Safely execute installation script with validation
# @param $1 Script path
# @param $@ Additional arguments
secure_execute_script() {
    local script_path="$1"
    shift
    local args=("$@")
    
    [[ -z "$script_path" ]] && error "Script path not specified"
    
    # Validate script exists and is executable
    [[ -f "$script_path" ]] || error "Script not found: $script_path"
    [[ -x "$script_path" ]] || error "Script not executable: $script_path"
    
    # Validate script path for security
    validate_file_path "$(basename "$script_path")"
    
    # Validate all arguments
    for arg in "${args[@]}"; do
        # Check for command injection attempts
        case "$arg" in
            *\;*|*\|*|*\&*|*\$*|*\`*) error "Invalid argument detected: $arg" ;;
        esac
        
        # Validate flag format
        if [[ "$arg" =~ ^-- ]]; then
            [[ "$arg" =~ ^--[a-zA-Z0-9-]+$ ]] || error "Invalid flag format: $arg"
        elif [[ "$arg" =~ ^- ]]; then
            [[ "$arg" =~ ^-[a-zA-Z0-9]+$ ]] || error "Invalid flag format: $arg"
        fi
    done
    
    log "Executing script: $script_path ${args[*]}"
    
    # Execute safely without eval
    "$script_path" "${args[@]}" || error "Script execution failed: $script_path"
    
    success "Script executed successfully: $script_path"
}

# @function validate_url
# @brief Validate URL for security
# @param $1 URL to validate
validate_url() {
    local url="$1"
    
    [[ -z "$url" ]] && error "URL not specified"
    
    # Check URL format
    [[ "$url" =~ ^https?:// ]] || error "Invalid URL format: $url"
    
    # Check for suspicious patterns
    case "$url" in
        *[[:space:]]*|*\|*|*\&*|*\;*) error "Invalid characters in URL: $url" ;;
    esac
    
    # Limit URL length
    [[ ${#url} -gt 500 ]] && error "URL too long: $url"
    
    return 0
}

# @function sanitize_filename
# @brief Sanitize filename for security
# @param $1 Filename to sanitize
# @return Sanitized filename
sanitize_filename() {
    local filename="$1"
    
    [[ -z "$filename" ]] && error "Filename not specified"
    
    # Remove dangerous characters and replace with underscores
    filename=$(echo "$filename" | tr -cd '[:alnum:]._-' | tr ' ' '_')
    
    # Limit length
    [[ ${#filename} -gt 100 ]] && filename="${filename:0:100}"
    
    # Ensure not empty after sanitization
    [[ -z "$filename" ]] && filename="sanitized_file"
    
    echo "$filename"
}

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# @function create_temp_dir
# @brief Create a temporary directory and set TEMP_DIR variable
create_temp_dir() {
    TEMP_DIR=$(mktemp -d)
    debug "Created temporary directory: $TEMP_DIR"
    echo "$TEMP_DIR"
}

# @function ensure_directory
# @brief Ensure directory exists, create if it doesn't
# @param $1 Directory path
ensure_directory() {
    local dir_path="$1"
    [[ -z "$dir_path" ]] && error "Directory path not specified"
    
    validate_file_path "$dir_path"
    
    if [[ ! -d "$dir_path" ]]; then
        mkdir -p "$dir_path" || error "Failed to create directory: $dir_path"
        debug "Created directory: $dir_path"
    fi
}

# @function safe_sudo_copy
# @brief Safely copy file with sudo, validating paths
# @param $1 Source file
# @param $2 Destination file
safe_sudo_copy() {
    local src_file="$1"
    local dest_file="$2"
    
    [[ -z "$src_file" ]] && error "Source file not specified"
    [[ -z "$dest_file" ]] && error "Destination file not specified"
    
    # Validate source file exists and is readable
    [[ -f "$src_file" ]] || error "Source file does not exist: $src_file"
    [[ -r "$src_file" ]] || error "Source file not readable: $src_file"
    
    # Validate file paths for security
    validate_file_path "$(basename "$src_file")"
    validate_file_path "$(basename "$dest_file")"
    
    # Ensure source file is not a symlink to prevent attacks
    [[ -L "$src_file" ]] && error "Source file is a symbolic link, refusing to copy: $src_file"
    
    # Validate destination directory exists and is secure
    local dest_dir
    dest_dir="$(dirname "$dest_file")"
    [[ -d "$dest_dir" ]] || error "Destination directory does not exist: $dest_dir"
    
    # Restrict destination to safe directories
    case "$dest_dir" in
        /usr/local/bin|/opt/*/bin|"$HOME"/.local/bin)
            # Allow these safe installation directories
            ;;
        *)
            error "Unsafe destination directory: $dest_dir. Only /usr/local/bin, /opt/*/bin, and ~/.local/bin are allowed."
            ;;
    esac
    
    # Get file size for verification
    local src_size
    src_size=$(stat -c%s "$src_file" 2>/dev/null || echo "0")
    
    log "Copying $src_file to $dest_file (size: $src_size bytes)"
    
    # Perform copy with error checking
    if ! sudo cp "$src_file" "$dest_file"; then
        error "Failed to copy $src_file to $dest_file"
    fi
    
    # Verify copy was successful
    [[ -f "$dest_file" ]] || error "Copy failed - destination file not found: $dest_file"
    
    local dest_size
    dest_size=$(stat -c%s "$dest_file" 2>/dev/null || echo "0")
    
    [[ "$src_size" == "$dest_size" ]] || error "Copy verification failed - size mismatch (src: $src_size, dest: $dest_size)"
    
    # Set secure permissions
    sudo chmod 755 "$dest_file" || error "Failed to set permissions on $dest_file"
    
    # Verify ownership is root (for system directories)
    if [[ "$dest_dir" == "/usr/local/bin" ]]; then
        local owner
        owner=$(stat -c%U "$dest_file" 2>/dev/null || echo "unknown")
        [[ "$owner" == "root" ]] || warning "Destination file owner is not root: $owner"
    fi
    
    success "Securely copied $src_file to $dest_file"
}

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

# =============================================================================
# SAFE COMMAND EXECUTION
# =============================================================================

# @function execute_command_safely
# @brief Safely execute commands without eval, preventing injection attacks
# @param $@ Command and arguments to execute
execute_command_safely() {
    local cmd=("$@")
    
    # Validate that we have at least one argument
    [[ ${#cmd[@]} -eq 0 ]] && error "No command specified"
    
    # Log the command being executed
    debug "Executing command: ${cmd[*]}"
    
    # Execute the command array safely
    "${cmd[@]}" || {
        local exit_code=$?
        error "Command failed (exit code $exit_code): ${cmd[*]}"
    }
}

# @function build_with_options
# @brief Safely build with dynamic options for cargo, make, etc.
# @param $1 Build command (cargo, make, etc.)
# @param $2 Build options string (may be empty)
# @param $@ Additional fixed arguments
build_with_options() {
    local build_cmd="$1"
    local options="$2"
    shift 2
    
    local cmd=("$build_cmd")
    
    # Add options if provided (split by spaces)
    if [[ -n "$options" ]]; then
        IFS=' ' read -ra option_array <<< "$options"
        cmd+=("${option_array[@]}")
    fi
    
    # Add remaining arguments
    cmd+=("$@")
    
    execute_command_safely "${cmd[@]}"
}

# @function configure_with_options
# @brief Safely run configure scripts with dynamic options
# @param $1 Configure script path
# @param $2 Configure options string (may be empty)
configure_with_options() {
    local configure_script="$1"
    local options="$2"
    
    local cmd=("$configure_script")
    
    # Add options if provided (split by spaces)
    if [[ -n "$options" ]]; then
        IFS=' ' read -ra option_array <<< "$options"
        cmd+=("${option_array[@]}")
    fi
    
    execute_command_safely "${cmd[@]}"
}

# =============================================================================
# CLEANUP AND ERROR HANDLING
# =============================================================================

# Global variables for cleanup tracking
declare -a CLEANUP_DIRS=()
declare -a CLEANUP_FILES=()
declare -a CLEANUP_COMMANDS=()

# @function cleanup_on_exit
# @brief Main cleanup function called on script exit
cleanup_on_exit() {
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        warning "Script exited with error code $exit_code, performing cleanup..."
    else
        debug "Script completed successfully, performing cleanup..."
    fi
    
    # Execute cleanup commands safely
    for cmd in "${CLEANUP_COMMANDS[@]}"; do
        debug "Executing cleanup command: $cmd"
        # Split command into array for safe execution
        read -ra cmd_array <<< "$cmd"
        "${cmd_array[@]}" || warning "Cleanup command failed: $cmd"
    done
    
    # Remove temporary files
    for file in "${CLEANUP_FILES[@]}"; do
        if [[ -f "$file" ]]; then
            debug "Removing temporary file: $file"
            rm -f "$file" || warning "Failed to remove file: $file"
        fi
    done
    
    # Remove temporary directories
    for dir in "${CLEANUP_DIRS[@]}"; do
        if [[ -d "$dir" ]]; then
            debug "Removing temporary directory: $dir"
            rm -rf "$dir" || warning "Failed to remove directory: $dir"
        fi
    done
    
    # Reset arrays
    CLEANUP_DIRS=()
    CLEANUP_FILES=()
    CLEANUP_COMMANDS=()
}

# @function register_cleanup_dir
# @brief Register a directory for cleanup on exit
# @param $1 Directory path to clean up
register_cleanup_dir() {
    local dir="$1"
    CLEANUP_DIRS+=("$dir")
    debug "Registered directory for cleanup: $dir"
}

# @function register_cleanup_file
# @brief Register a file for cleanup on exit
# @param $1 File path to clean up
register_cleanup_file() {
    local file="$1"
    CLEANUP_FILES+=("$file")
    debug "Registered file for cleanup: $file"
}

# @function register_cleanup_command
# @brief Register a command to run during cleanup
# @param $1 Command to execute during cleanup
register_cleanup_command() {
    local cmd="$1"
    CLEANUP_COMMANDS+=("$cmd")
    debug "Registered cleanup command: $cmd"
}

# @function setup_cleanup_trap
# @brief Set up the cleanup trap for the current script
setup_cleanup_trap() {
    trap cleanup_on_exit EXIT INT TERM
    debug "Cleanup trap set up successfully"
}

# @function create_temp_dir
# @brief Create a temporary directory and register for cleanup
# @param $1 Optional prefix for directory name
# @return Prints the temporary directory path
create_temp_dir() {
    local prefix="${1:-gearbox}"
    local temp_dir
    temp_dir=$(mktemp -d -t "${prefix}.XXXXXX")
    register_cleanup_dir "$temp_dir"
    echo "$temp_dir"
}

# @function create_temp_file
# @brief Create a temporary file and register for cleanup
# @param $1 Optional prefix for file name
# @return Prints the temporary file path
create_temp_file() {
    local prefix="${1:-gearbox}"
    local temp_file
    temp_file=$(mktemp -t "${prefix}.XXXXXX")
    register_cleanup_file "$temp_file"
    echo "$temp_file"
}

# =============================================================================
# PROGRESS INDICATORS
# =============================================================================

# @function show_progress
# @brief Show progress indicator during long operations
# @param $1 Current step number
# @param $2 Total steps
# @param $3 Description of current step
show_progress() {
    local current="$1"
    local total="$2"
    local description="$3"
    
    local percentage=$((current * 100 / total))
    local completed=$((current * 50 / total))
    local remaining=$((50 - completed))
    
    # Build progress bar
    local bar=""
    for ((i=0; i<completed; i++)); do bar+="█"; done
    for ((i=0; i<remaining; i++)); do bar+="░"; done
    
    printf "\r${BLUE}[%3d%%]${NC} ${bar} ${YELLOW}(%d/%d)${NC} %s" \
           "$percentage" "$current" "$total" "$description"
    
    # Add newline when complete
    if [[ $current -eq $total ]]; then
        echo
    fi
}

# @function start_spinner
# @brief Start a spinner for indefinite operations
# @param $1 Message to display with spinner
start_spinner() {
    local message="$1"
    local spinner_chars="⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
    local i=0
    
    while true; do
        printf "\r${BLUE}${spinner_chars:$i:1}${NC} %s" "$message"
        sleep 0.1
        i=$(( (i + 1) % ${#spinner_chars} ))
    done
}

# @function stop_spinner
# @brief Stop the spinner and clear the line
stop_spinner() {
    printf "\r\033[K"  # Clear the line
}

# @function log_step
# @brief Log a step with step number for tracking progress
# @param $1 Step number
# @param $2 Total steps
# @param $3 Step description
log_step() {
    local step="$1"
    local total="$2"
    local description="$3"
    
    log "${YELLOW}[Step $step/$total]${NC} $description"
}

# =============================================================================
# BUILD CACHE SYSTEM
# =============================================================================

# @function get_cache_key
# @brief Generate a cache key for a tool build
# @param $1 Tool name
# @param $2 Build type
# @param $3 Version/commit hash
get_cache_key() {
    local tool_name="$1"
    local build_type="$2"
    local version="$3"
    
    echo "${tool_name}-${build_type}-${version}"
}

# @function is_cached
# @brief Check if a build is already cached with integrity verification
# @param $1 Tool name
# @param $2 Build type
# @param $3 Version/commit hash
is_cached() {
    local tool_name="$1"
    local build_type="$2"
    local version="$3"
    local cache_key
    cache_key=$(get_cache_key "$tool_name" "$build_type" "$version")
    local cache_path="$CACHE_DIR/builds/$cache_key"
    
    # Basic existence check
    if [[ ! -d "$cache_path" || ! -f "$cache_path/.build_complete" ]]; then
        return 1
    fi
    
    # Integrity check: verify binary exists and is executable
    local binary_name="${tool_name}"
    local cached_binary="$cache_path/bin/$binary_name"
    
    if [[ ! -f "$cached_binary" || ! -x "$cached_binary" ]]; then
        warning "Cache integrity check failed for $tool_name: binary missing or not executable"
        # Clean up corrupted cache entry
        rm -rf "$cache_path" 2>/dev/null || true
        return 1
    fi
    
    # Age check: reject cache entries older than configured limit (default 7 days)
    local max_age_days="${GEARBOX_CACHE_MAX_AGE_DAYS:-7}"
    if [[ -n "$max_age_days" && "$max_age_days" -gt 0 ]]; then
        local cache_age_seconds
        cache_age_seconds=$(( $(date +%s) - $(stat -c %Y "$cache_path/.build_complete" 2>/dev/null || echo 0) ))
        local max_age_seconds=$((max_age_days * 86400))
        
        if [[ $cache_age_seconds -gt $max_age_seconds ]]; then
            log "Cache entry for $tool_name is older than $max_age_days days, invalidating"
            rm -rf "$cache_path" 2>/dev/null || true
            return 1
        fi
    fi
    
    return 0
}

# @function get_cached_binary
# @brief Retrieve cached binary if available
# @param $1 Tool name
# @param $2 Build type
# @param $3 Version/commit hash
# @param $4 Binary name (optional, defaults to tool name)
get_cached_binary() {
    local tool_name="$1"
    local build_type="$2"
    local version="$3"
    local binary_name="${4:-$tool_name}"
    
    local cache_key
    cache_key=$(get_cache_key "$tool_name" "$build_type" "$version")
    local cache_path="$CACHE_DIR/builds/$cache_key"
    
    if is_cached "$tool_name" "$build_type" "$version"; then
        local cached_binary="$cache_path/bin/$binary_name"
        if [[ -f "$cached_binary" ]]; then
            log "Using cached binary for $tool_name"
            sudo cp "$cached_binary" "/usr/local/bin/"
            return 0
        fi
    fi
    
    return 1
}

# @function cache_build
# @brief Cache a successful build (non-fatal if binary not found)
# @param $1 Tool name
# @param $2 Build type
# @param $3 Version/commit hash
# @param $4 Path to built binary
cache_build() {
    local tool_name="$1"
    local build_type="$2"
    local version="$3"
    local binary_path="$4"
    
    # Validate parameters
    if [[ -z "$tool_name" || -z "$build_type" || -z "$version" || -z "$binary_path" ]]; then
        warning "Cache build skipped: missing parameters (tool=$tool_name, type=$build_type, version=$version, path=$binary_path)"
        return 0  # Non-fatal
    fi
    
    local cache_key
    cache_key=$(get_cache_key "$tool_name" "$build_type" "$version")
    local cache_path="$CACHE_DIR/builds/$cache_key"
    
    # Create cache directory structure (non-fatal if it fails)
    if ! ensure_directory "$cache_path/bin" 2>/dev/null; then
        warning "Failed to create cache directory, skipping cache for $tool_name"
        return 0  # Non-fatal
    fi
    
    # Copy binary to cache with integrity verification
    if [[ -f "$binary_path" ]]; then
        if cp "$binary_path" "$cache_path/bin/" 2>/dev/null; then
            local cached_binary="$cache_path/bin/$(basename "$binary_path")"
            
            # Generate checksum for integrity verification
            local checksum
            if command -v sha256sum >/dev/null 2>&1; then
                checksum=$(sha256sum "$cached_binary" 2>/dev/null | cut -d' ' -f1)
            elif command -v shasum >/dev/null 2>&1; then
                checksum=$(shasum -a 256 "$cached_binary" 2>/dev/null | cut -d' ' -f1)
            else
                checksum="unavailable"
            fi
            
            # Store build metadata with integrity information
            {
                echo "timestamp=$(date '+%Y-%m-%d %H:%M:%S')"
                echo "tool=$tool_name"
                echo "build_type=$build_type"
                echo "version=$version"
                echo "binary_size=$(stat -c%s "$cached_binary" 2>/dev/null || echo 0)"
                echo "checksum=$checksum"
                echo "gearbox_version=${GEARBOX_VERSION:-unknown}"
            } > "$cache_path/.build_complete"
            
            log "Cached build for $tool_name ($build_type, $version) with checksum: ${checksum:0:12}..."
        else
            warning "Failed to copy binary to cache: $binary_path"
        fi
    else
        warning "Binary not found for caching: $binary_path"
    fi
    
    # Always return success to prevent script failure
    return 0
}

# @function get_tool_version
# @brief Get version/commit hash for a tool from its repository
# @param $1 Repository directory path
get_tool_version() {
    local repo_dir="$1"
    
    if [[ -d "$repo_dir/.git" ]]; then
        cd "$repo_dir" && git rev-parse --short HEAD
    else
        echo "unknown"
    fi
}

# @function verify_cached_binary
# @brief Verify cached binary integrity with comprehensive checks
# @param $1 Tool name
# @param $2 Build type
# @param $3 Version/commit hash
verify_cached_binary() {
    local tool_name="$1"
    local build_type="$2"
    local version="$3"
    local cache_key
    cache_key=$(get_cache_key "$tool_name" "$build_type" "$version")
    local cache_path="$CACHE_DIR/builds/$cache_key"
    local cached_binary="$cache_path/bin/$tool_name"
    
    # Basic existence and executable check
    if [[ ! -f "$cached_binary" || ! -x "$cached_binary" ]]; then
        return 1
    fi
    
    # Verify checksum if available
    if [[ -f "$cache_path/.build_complete" ]]; then
        local stored_checksum
        stored_checksum=$(grep "^checksum=" "$cache_path/.build_complete" | cut -d'=' -f2)
        
        if [[ "$stored_checksum" != "unavailable" && -n "$stored_checksum" ]]; then
            local current_checksum
            if command -v sha256sum >/dev/null 2>&1; then
                current_checksum=$(sha256sum "$cached_binary" 2>/dev/null | cut -d' ' -f1)
            elif command -v shasum >/dev/null 2>&1; then
                current_checksum=$(shasum -a 256 "$cached_binary" 2>/dev/null | cut -d' ' -f1)
            else
                return 0  # Skip checksum verification if tools unavailable
            fi
            
            if [[ "$stored_checksum" != "$current_checksum" ]]; then
                warning "Checksum mismatch for cached $tool_name: expected $stored_checksum, got $current_checksum"
                return 1
            fi
        fi
    fi
    
    # Quick functional test: check if binary responds to --version or --help
    if timeout 5 "$cached_binary" --version >/dev/null 2>&1 || 
       timeout 5 "$cached_binary" --help >/dev/null 2>&1 || 
       timeout 5 "$cached_binary" -V >/dev/null 2>&1; then
        return 0
    else
        warning "Cached binary for $tool_name failed functional test"
        return 1
    fi
}

# @function cache_warmup
# @brief Pre-warm cache with commonly used tools
# @param $1 Optional space-separated list of tools to warm up
cache_warmup() {
    local tools_to_warm="${1:-fd ripgrep fzf bat}"
    local warmed_count=0
    local total_count=0
    
    log "Starting cache warmup for commonly used tools..."
    
    for tool in $tools_to_warm; do
        total_count=$((total_count + 1))
        
        # Check if already cached
        if is_cached "$tool" "standard" "latest"; then
            log "✓ $tool already cached"
            warmed_count=$((warmed_count + 1))
        else
            log "○ $tool not cached - will be built on first use"
        fi
    done
    
    log "Cache warmup complete: $warmed_count/$total_count tools cached"
    
    # Show cache statistics
    show_cache_stats
}

# @function show_cache_stats
# @brief Display cache usage statistics
show_cache_stats() {
    local cache_builds_dir="$CACHE_DIR/builds"
    
    if [[ ! -d "$cache_builds_dir" ]]; then
        log "Cache directory not found"
        return 0
    fi
    
    local total_entries cached_size
    total_entries=$(find "$cache_builds_dir" -name ".build_complete" 2>/dev/null | wc -l)
    cached_size=$(du -sh "$cache_builds_dir" 2>/dev/null | cut -f1 || echo "unknown")
    
    log "Cache statistics:"
    log "  - Total cached builds: $total_entries"
    log "  - Cache size: $cached_size"
    log "  - Cache location: $cache_builds_dir"
    
    # Show recently used tools
    if [[ $total_entries -gt 0 ]]; then
        log "  - Recent builds:"
        find "$cache_builds_dir" -name ".build_complete" -exec grep "^tool=" {} \; 2>/dev/null | \
            cut -d'=' -f2 | sort | uniq -c | sort -nr | head -5 | \
            while read count tool; do
                log "    $tool ($count versions)"
            done
    fi
}

# @function clean_old_cache
# @brief Clean cache entries older than specified days
# @param $1 Maximum age in days (default: 30)
clean_old_cache() {
    local max_age_days="${1:-30}"
    local cache_builds_dir="$CACHE_DIR/builds"
    
    if [[ -d "$cache_builds_dir" ]]; then
        log "Cleaning cache entries older than $max_age_days days..."
        # Use a more robust approach to avoid exit code issues
        local old_files
        old_files=$(find "$cache_builds_dir" -type d -name "*-*-*" -mtime +$max_age_days 2>/dev/null || true)
        if [[ -n "$old_files" ]]; then
            echo "$old_files" | while read -r dir; do
                [[ -d "$dir" ]] && rm -rf "$dir" 2>/dev/null || true
            done
        fi
    fi
    return 0
}

# =============================================================================
# DISK SPACE OPTIMIZATION AND BUILD ARTIFACT CLEANUP
# =============================================================================

# @function cleanup_build_artifacts
# @brief Clean build artifacts while preserving installed binaries and essential files
# @param $1 Tool name
# @param $2 Optional: cleanup mode (standard|aggressive|minimal) - default: standard
cleanup_build_artifacts() {
    local tool_name="$1"
    local cleanup_mode="${2:-standard}"
    local tool_build_dir="$BUILD_DIR"
    
    if [[ -z "$tool_name" ]]; then
        warning "Tool name required for build artifact cleanup"
        return 1
    fi
    
    # Validate cleanup mode
    case "$cleanup_mode" in
        minimal|standard|aggressive)
            ;;
        *)
            warning "Invalid cleanup mode '$cleanup_mode'. Using 'standard'"
            cleanup_mode="standard"
            ;;
    esac
    
    log "Starting $cleanup_mode cleanup of build artifacts for $tool_name..."
    
    # Find potential build directories for this tool
    local build_dirs=(
        "$tool_build_dir/$tool_name"
        "$tool_build_dir/${tool_name}-*"
    )
    
    local cleaned_size=0
    local preserved_count=0
    
    for pattern in "${build_dirs[@]}"; do
        # Use shell expansion to find matching directories
        for build_dir in $(ls -d $pattern 2>/dev/null || true); do
            if [[ -d "$build_dir" ]]; then
                local dir_size_before
                dir_size_before=$(du -sb "$build_dir" 2>/dev/null | cut -f1 || echo 0)
                
                case "$cleanup_mode" in
                    minimal)
                        # Only clean obvious temporary files
                        cleanup_minimal_artifacts "$build_dir" "$tool_name"
                        ;;
                    standard)
                        # Clean intermediate build files but preserve source and key files
                        cleanup_standard_artifacts "$build_dir" "$tool_name"
                        ;;
                    aggressive)
                        # Remove everything except preserved source files if configured
                        cleanup_aggressive_artifacts "$build_dir" "$tool_name"
                        ;;
                esac
                
                # Calculate space saved
                local dir_size_after
                dir_size_after=$(du -sb "$build_dir" 2>/dev/null | cut -f1 || echo 0)
                local space_saved=$((dir_size_before - dir_size_after))
                cleaned_size=$((cleaned_size + space_saved))
                
                if [[ $space_saved -gt 0 ]]; then
                    log "Cleaned $(human_readable_size $space_saved) from $(basename "$build_dir")"
                else
                    preserved_count=$((preserved_count + 1))
                fi
            fi
        done
    done
    
    if [[ $cleaned_size -gt 0 ]]; then
        success "Build cleanup complete: freed $(human_readable_size $cleaned_size) for $tool_name"
    else
        log "No build artifacts found to clean for $tool_name"
    fi
    
    return 0
}

# @function cleanup_minimal_artifacts
# @brief Minimal cleanup - only obvious temporary files
# @param $1 Build directory path
# @param $2 Tool name
cleanup_minimal_artifacts() {
    local build_dir="$1"
    local tool_name="$2"
    
    # Only remove clearly temporary files and directories
    local temp_patterns=(
        "*.tmp"
        "*.temp"
        ".tmp*"
        "tmp"
        "temp"
        "*.log"
        "*.pid"
        "core.*"
        "*.core"
    )
    
    for pattern in "${temp_patterns[@]}"; do
        find "$build_dir" -name "$pattern" -type f -delete 2>/dev/null || true
        find "$build_dir" -name "$pattern" -type d -exec rm -rf {} + 2>/dev/null || true
    done
}

# @function cleanup_standard_artifacts
# @brief Standard cleanup - remove intermediate build files but preserve source
# @param $1 Build directory path  
# @param $2 Tool name
cleanup_standard_artifacts() {
    local build_dir="$1"
    local tool_name="$2"
    
    # Start with minimal cleanup
    cleanup_minimal_artifacts "$build_dir" "$tool_name"
    
    # Remove common build artifacts while preserving source and important files
    
    # Rust cleanup: remove target directory build artifacts
    # Handle two installation patterns:
    # 1. cargo install -> ~/.cargo/bin/tool + symlink /usr/local/bin/tool
    # 2. Direct copy -> /usr/local/bin/tool (e.g., uv)
    if [[ -d "$build_dir/target" ]]; then
        # Check if tool is properly installed using either pattern
        local cargo_binary="$HOME/.cargo/bin/$tool_name"
        local system_binary="/usr/local/bin/$tool_name"
        local is_properly_installed="false"
        local installation_type=""
        
        # Pattern 1: cargo install (symlink from /usr/local/bin to ~/.cargo/bin)
        if [[ -f "$cargo_binary" && -L "$system_binary" ]]; then
            is_properly_installed="true"
            installation_type="cargo_install"
        # Pattern 2: direct copy (binary directly in /usr/local/bin)
        elif [[ -f "$system_binary" && ! -L "$system_binary" ]]; then
            is_properly_installed="true"
            installation_type="direct_copy"
        fi
        
        # Only clean if the tool is properly installed
        if [[ "$is_properly_installed" == "true" ]]; then
            log "Tool $tool_name is properly installed ($installation_type), cleaning build artifacts from target/"
            
            # Clean all build artifacts - target/release binaries are no longer needed
            # since the final binary is in /usr/local/bin/ (either via symlink or direct copy)
            rm -rf "$build_dir/target/release/build" 2>/dev/null || true
            rm -rf "$build_dir/target/release/deps" 2>/dev/null || true 
            rm -rf "$build_dir/target/release/incremental" 2>/dev/null || true
            rm -rf "$build_dir/target/debug" 2>/dev/null || true
            find "$build_dir/target" -name "*.rlib" -delete 2>/dev/null || true
            find "$build_dir/target" -name "*.rmeta" -delete 2>/dev/null || true
            rm -f "$build_dir/target/.rustc_info.json" 2>/dev/null || true
            rm -f "$build_dir/target/CACHEDIR.TAG" 2>/dev/null || true
            
            # Remove the intermediate binary since it's no longer used
            rm -f "$build_dir/target/release/$tool_name" 2>/dev/null || true
        else
            warning "Tool $tool_name not properly installed, preserving build artifacts"
            log "Expected: $system_binary (either direct binary or symlink to $cargo_binary)"
        fi
    fi
    
    # Go cleanup
    find "$build_dir" -name "*.o" -type f -delete 2>/dev/null || true
    find "$build_dir" -name "*.a" -type f -delete 2>/dev/null || true
    rm -rf "$build_dir/pkg" 2>/dev/null || true
    
    # C/C++ cleanup
    find "$build_dir" -name "*.o" -type f -delete 2>/dev/null || true
    find "$build_dir" -name "*.lo" -type f -delete 2>/dev/null || true
    find "$build_dir" -name "*.la" -type f -delete 2>/dev/null || true
    rm -rf "$build_dir/.libs" 2>/dev/null || true
    rm -rf "$build_dir/autom4te.cache" 2>/dev/null || true
    rm -f "$build_dir/config.log" "$build_dir/config.status" 2>/dev/null || true
    
    # Python cleanup
    find "$build_dir" -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true
    find "$build_dir" -name "*.pyc" -type f -delete 2>/dev/null || true
    find "$build_dir" -name "*.pyo" -type f -delete 2>/dev/null || true
    rm -rf "$build_dir/build/lib" "$build_dir/build/temp" 2>/dev/null || true
    rm -rf "$build_dir/dist" 2>/dev/null || true
    find "$build_dir" -name "*.egg-info" -type d -exec rm -rf {} + 2>/dev/null || true
    
    # Git cleanup: remove large pack files but keep repository for version tracking
    rm -rf "$build_dir/.git/objects/pack" 2>/dev/null || true
    
    # General cleanup
    rm -rf "$build_dir/node_modules" 2>/dev/null || true
}

# @function cleanup_aggressive_artifacts
# @brief Aggressive cleanup - remove everything except source if configured to preserve
# @param $1 Build directory path
# @param $2 Tool name  
cleanup_aggressive_artifacts() {
    local build_dir="$1"
    local tool_name="$2"
    
    # Check configuration for source preservation
    local preserve_source="${GEARBOX_PRESERVE_SOURCE:-true}"
    
    if [[ "$preserve_source" == "false" ]]; then
        # Remove entire build directory
        log "Aggressive cleanup: removing entire build directory for $tool_name"
        rm -rf "$build_dir" 2>/dev/null || true
        return 0
    fi
    
    # Create temporary directory to hold preserved files
    local temp_preserve_dir
    temp_preserve_dir=$(mktemp -d)
    
    # Copy essential source files and configuration  
    find "$build_dir" -name "*.rs" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "*.go" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "*.c" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "*.cpp" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "*.h" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "Cargo.toml" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "go.mod" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "Makefile" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "README*" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    find "$build_dir" -name "LICENSE*" -type f -exec cp --parents {} "$temp_preserve_dir/" \; 2>/dev/null || true
    
    # Remove original and recreate with preserved files
    rm -rf "$build_dir" 2>/dev/null || true
    mkdir -p "$build_dir"
    
    # Restore preserved files  
    if [[ -d "$temp_preserve_dir$build_dir" ]]; then
        cp -r "$temp_preserve_dir$build_dir"/* "$build_dir/" 2>/dev/null || true
    fi
    
    rm -rf "$temp_preserve_dir"
    
    log "Aggressive cleanup: preserved only source files for $tool_name"
}

# @function human_readable_size
# @brief Convert bytes to human readable format
# @param $1 Size in bytes
human_readable_size() {
    local bytes="$1"
    local units=("B" "KB" "MB" "GB" "TB")
    local unit=0
    
    while [[ $bytes -ge 1024 && $unit -lt 4 ]]; do
        bytes=$((bytes / 1024))
        unit=$((unit + 1))
    done
    
    echo "${bytes}${units[$unit]}"
}

# @function show_disk_usage
# @brief Display disk usage information for build and cache directories
show_disk_usage() {
    log "Disk Usage Report for Gearbox"
    log "============================="
    
    # Build directory usage
    if [[ -d "$BUILD_DIR" ]]; then
        local build_size
        build_size=$(du -sh "$BUILD_DIR" 2>/dev/null | cut -f1 || echo "unknown")
        local build_count
        build_count=$(find "$BUILD_DIR" -maxdepth 1 -type d | wc -l)
        build_count=$((build_count - 1))  # Subtract 1 for the directory itself
        
        log "Build Directory ($BUILD_DIR):"
        log "  - Total size: $build_size"
        log "  - Tool directories: $build_count"
        
        # Show largest build directories
        if [[ $build_count -gt 0 ]]; then
            log "  - Largest builds:"
            du -sh "$BUILD_DIR"/*/ 2>/dev/null | sort -hr | head -5 | while read size dir; do
                log "    $(basename "$dir"): $size"
            done
        fi
    else
        log "Build Directory: not found"
    fi
    
    echo
    
    # Cache directory usage (from existing function)
    show_cache_stats
    
    echo
    
    # Cleanup recommendations
    local total_build_size_bytes
    total_build_size_bytes=$(du -sb "$BUILD_DIR" 2>/dev/null | cut -f1 || echo 0)
    
    if [[ $total_build_size_bytes -gt $((1024 * 1024 * 1024)) ]]; then  # > 1GB
        log "🧹 Cleanup Recommendations:"
        log "  - Build directory is large (>1GB)"
        log "  - Run cleanup_build_artifacts <tool_name> to clean specific tools"
        log "  - Consider setting GEARBOX_AUTO_CLEANUP=true for automatic cleanup"
        log "  - Set GEARBOX_CLEANUP_MODE=aggressive for maximum space savings"
    fi
}

# @function auto_cleanup_after_install
# @brief Automatically cleanup build artifacts after successful installation
# @param $1 Tool name
# @param $2 Installation success (true/false)
auto_cleanup_after_install() {
    local tool_name="$1"
    local install_success="${2:-false}"
    
    # Only cleanup if installation was successful
    if [[ "$install_success" != "true" ]]; then
        debug "Skipping auto-cleanup for $tool_name: installation not successful"
        return 0
    fi
    
    # Check if auto-cleanup is enabled
    local auto_cleanup="${GEARBOX_AUTO_CLEANUP:-false}"
    local cleanup_mode="${GEARBOX_CLEANUP_MODE:-standard}"
    
    if [[ "$auto_cleanup" == "true" ]]; then
        log "Auto-cleanup enabled: cleaning build artifacts for $tool_name"
        cleanup_build_artifacts "$tool_name" "$cleanup_mode"
    else
        debug "Auto-cleanup disabled for $tool_name. Set GEARBOX_AUTO_CLEANUP=true to enable"
    fi
}

# =============================================================================
# DIRECTORY CONFIGURATION
# =============================================================================

# Global directory paths
export BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"
export CACHE_DIR="${CACHE_DIR:-$HOME/tools/cache}"

# =============================================================================
# INITIALIZATION
# =============================================================================

# Ensure required directories exist
ensure_directory "$BUILD_DIR"
ensure_directory "$CACHE_DIR"
ensure_directory "$CACHE_DIR/builds"

# Clean old cache entries on startup (older than 30 days) - disabled for now to prevent startup issues
# if [[ "${GEARBOX_SKIP_CACHE_CLEANUP:-}" != "true" ]]; then
#     clean_old_cache 30 2>/dev/null || true
# fi

# Export important variables for use by calling scripts
export GEARBOX_COMMON_LOADED GEARBOX_LIB_DIR GEARBOX_REPO_DIR

# Automatically set up cleanup traps for any script that sources this library
setup_cleanup_trap

debug "Common library loaded successfully with cleanup traps enabled"