#!/bin/bash
#
# @file lib/core/utilities.sh
# @brief Core utility functions for gearbox
# @description
#   Basic utility functions including parallel job calculation,
#   temporary file management, and common helper functions.
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_UTILITIES_LOADED:-}" ]] && return 0
readonly GEARBOX_UTILITIES_LOADED=1

# =============================================================================
# UTILITY FUNCTIONS
# =============================================================================

# @function get_optimal_jobs
# @brief Calculate optimal number of parallel jobs based on system resources
# @return Number of jobs to use
get_optimal_jobs() {
    local cpu_cores
    local available_memory_gb
    local max_jobs
    local configured_jobs
    
    # Get CPU core count
    cpu_cores=$(nproc 2>/dev/null || echo 1)
    
    # Get available memory in GB
    if command -v free >/dev/null 2>&1; then
        available_memory_gb=$(free -g | awk '/^Mem:/{print $7}')
        [[ -z "$available_memory_gb" || "$available_memory_gb" -eq 0 ]] && available_memory_gb=1
    else
        available_memory_gb=4  # Reasonable default
    fi
    
    # Calculate max jobs based on memory (assume ~1GB per job for Rust builds)
    # For lower memory systems, be more conservative
    if [[ $available_memory_gb -le 2 ]]; then
        max_jobs=1
    elif [[ $available_memory_gb -le 4 ]]; then
        max_jobs=2
    elif [[ $available_memory_gb -le 8 ]]; then
        max_jobs=4
    else
        max_jobs=$cpu_cores
    fi
    
    # Don't use more jobs than CPU cores
    [[ $max_jobs -gt $cpu_cores ]] && max_jobs=$cpu_cores
    
    # Check for user configuration
    configured_jobs=$(get_config "MAX_PARALLEL_JOBS" "auto")
    if [[ "$configured_jobs" != "auto" && "$configured_jobs" =~ ^[0-9]+$ ]]; then
        if [[ $configured_jobs -le $max_jobs ]]; then
            max_jobs=$configured_jobs
        else
            warning "Configured MAX_PARALLEL_JOBS ($configured_jobs) exceeds safe limit ($max_jobs), using $max_jobs"
        fi
    fi
    
    # Ensure at least 1 job
    [[ $max_jobs -lt 1 ]] && max_jobs=1
    
    debug "Optimal parallel jobs: $max_jobs (CPU cores: $cpu_cores, Available memory: ${available_memory_gb}GB)"
    echo "$max_jobs"
}

# @function run_with_timeout
# @brief Run command with timeout
# @param $1 Timeout in seconds
# @param $@ Command to run
run_with_timeout() {
    local timeout_seconds="$1"
    shift
    timeout "$timeout_seconds" "$@"
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
# @brief Show disk usage for a directory
# @param $1 Directory path
show_disk_usage() {
    local dir_path="$1"
    
    if [[ -d "$dir_path" ]]; then
        local size_bytes
        size_bytes=$(du -sb "$dir_path" 2>/dev/null | cut -f1)
        if [[ -n "$size_bytes" ]]; then
            local readable_size
            readable_size=$(human_readable_size "$size_bytes")
            log "Directory size: $dir_path = $readable_size"
        else
            warning "Could not determine size for: $dir_path"
        fi
    else
        warning "Directory does not exist: $dir_path"
    fi
}

# @function backup_file
# @brief Create backup of file before modification
# @param $1 File path to backup
# @return Path to backup file
backup_file() {
    local file_path="$1"
    local backup_path
    
    [[ -z "$file_path" ]] && error "File path not specified for backup"
    [[ ! -f "$file_path" ]] && error "File does not exist: $file_path"
    
    backup_path="${file_path}.backup-$(date +%Y%m%d-%H%M%S)"
    
    log "Creating backup: $file_path -> $backup_path"
    cp "$file_path" "$backup_path" || error "Failed to create backup"
    
    echo "$backup_path"
}

# @function restore_file_backup
# @brief Restore file from backup
# @param $1 Backup file path
# @param $2 Original file path (optional, derived from backup name)
restore_file_backup() {
    local backup_path="$1"
    local original_path="${2:-}"
    
    [[ -z "$backup_path" ]] && error "Backup path not specified"
    [[ ! -f "$backup_path" ]] && error "Backup file does not exist: $backup_path"
    
    # Derive original path if not provided
    if [[ -z "$original_path" ]]; then
        original_path="${backup_path%.backup-*}"
    fi
    
    log "Restoring backup: $backup_path -> $original_path"
    cp "$backup_path" "$original_path" || error "Failed to restore backup"
    
    success "Backup restored successfully"
}

# @function clone_or_update_repo
# @brief Clone repository or update if it already exists
# @param $1 Repository URL
# @param $2 Local directory path
# @param $3 Branch name (optional, defaults to main/master)
clone_or_update_repo() {
    local repo_url="$1"
    local local_dir="$2"
    local branch="${3:-}"
    
    [[ -z "$repo_url" ]] && error "Repository URL not specified"
    [[ -z "$local_dir" ]] && error "Local directory not specified"
    
    # Validate URL
    validate_url "$repo_url" || error "Invalid repository URL: $repo_url"
    
    if [[ -d "$local_dir/.git" ]]; then
        log "Repository exists, updating: $local_dir"
        cd "$local_dir" || error "Failed to enter directory: $local_dir"
        
        # Check if we can update (working directory clean)
        if git diff-index --quiet HEAD --; then
            git fetch origin || warning "Failed to fetch from origin"
            
            # Determine current branch
            local current_branch
            current_branch=$(git branch --show-current)
            
            # Use specified branch or current branch
            local target_branch="${branch:-$current_branch}"
            
            if [[ "$current_branch" != "$target_branch" ]]; then
                log "Switching to branch: $target_branch"
                git checkout "$target_branch" || warning "Failed to switch to branch: $target_branch"
            fi
            
            git pull origin "$target_branch" || warning "Failed to pull updates"
        else
            warning "Repository has local changes, skipping update: $local_dir"
        fi
        
        cd - >/dev/null
    else
        log "Cloning repository: $repo_url -> $local_dir"
        
        # Create parent directory if needed
        local parent_dir
        parent_dir=$(dirname "$local_dir")
        ensure_directory "$parent_dir"
        
        # Clone with optional branch
        local clone_cmd=("git" "clone")
        if [[ -n "$branch" ]]; then
            clone_cmd+=("--branch" "$branch")
        fi
        clone_cmd+=("$repo_url" "$local_dir")
        
        "${clone_cmd[@]}" || error "Failed to clone repository: $repo_url"
        
        success "Repository cloned successfully: $local_dir"
    fi
}