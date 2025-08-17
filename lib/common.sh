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

# Source main configuration file
if [[ -f "$GEARBOX_REPO_DIR/config.sh" ]]; then
    source "$GEARBOX_REPO_DIR/config.sh"
else
    echo "ERROR: config.sh not found in $GEARBOX_REPO_DIR" >&2
    exit 1
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
    [[ "${DEBUG:-false}" == "true" ]] && echo -e "${BLUE}[DEBUG]${NC} $1" >&2
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
# @brief Calculate optimal number of parallel jobs for builds
# @return Prints optimal job count
get_optimal_jobs() {
    local cpu_cores
    cpu_cores=$(nproc)
    
    # Limit based on available memory (assume 1GB per job for safety)
    local available_memory_gb
    available_memory_gb=$(($(free -g | awk '/^Mem:/{print $7}') + 1))
    
    local memory_limited_jobs=$((available_memory_gb))
    
    # Return the smaller of CPU cores and memory-limited jobs, minimum of 1
    local optimal_jobs=$((cpu_cores < memory_limited_jobs ? cpu_cores : memory_limited_jobs))
    echo $((optimal_jobs > 0 ? optimal_jobs : 1))
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
        
        # Clean up any temporary directories
        [[ -n "${TEMP_DIR:-}" && -d "$TEMP_DIR" ]] && rm -rf "$TEMP_DIR"
        
        # Additional cleanup can be added here
        debug "Cleanup completed"
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
    
    # Execute the piped command
    log "Executing: $pipe_command"
    cat "$temp_file" | eval "$pipe_command" || {
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
        [[ "$arg" =~ [;\|&\$\`] ]] && error "Invalid argument detected: $arg"
        
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
    [[ "$url" =~ [[:space:]\|&\;] ]] && error "Invalid characters in URL: $url"
    
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
# INITIALIZATION
# =============================================================================

# Ensure required directories exist
ensure_directory "$BUILD_DIR"
ensure_directory "$CACHE_DIR"

# Export important variables for use by calling scripts
export GEARBOX_COMMON_LOADED GEARBOX_LIB_DIR GEARBOX_REPO_DIR

debug "Common library loaded successfully"