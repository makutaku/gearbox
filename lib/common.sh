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
    
    # Validate source file exists
    [[ -f "$src_file" ]] || error "Source file does not exist: $src_file"
    
    # Validate paths
    validate_file_path "$(basename "$src_file")"
    validate_file_path "$(basename "$dest_file")"
    
    # Perform copy
    sudo cp "$src_file" "$dest_file" || error "Failed to copy $src_file to $dest_file"
    sudo chmod +x "$dest_file" || error "Failed to make $dest_file executable"
    
    success "Copied $src_file to $dest_file"
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