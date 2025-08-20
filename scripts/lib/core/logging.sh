#!/bin/bash
#
# @file lib/core/logging.sh
# @brief Logging and output functions for gearbox
# @description
#   Centralized logging system with consistent formatting and color support.
#   Provides standard logging levels: debug, info, warning, error, success.
#

# Prevent multiple inclusion
[[ -n "${GEARBOX_LOGGING_LOADED:-}" ]] && return 0
readonly GEARBOX_LOGGING_LOADED=1

# Color definitions
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

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

# @function log_step
# @brief Log a step with progress indicator
# @param $1 Step number
# @param $2 Total steps
# @param $3 Step description
log_step() {
    local step="$1"
    local total="$2"
    local description="$3"
    
    log "Step [$step/$total]: $description"
}

# @function show_progress
# @brief Show progress for long-running operations
# @param $1 Current progress (0-100)
# @param $2 Operation description
show_progress() {
    local progress="$1"
    local description="${2:-Progress}"
    local bar_length=50
    local filled_length=$((progress * bar_length / 100))
    
    # Create progress bar
    local bar=""
    for ((i=0; i<filled_length; i++)); do
        bar+="█"
    done
    for ((i=filled_length; i<bar_length; i++)); do
        bar+="░"
    done
    
    printf "\r${BLUE}[PROGRESS]${NC} %s [%s] %d%%" "$description" "$bar" "$progress"
    
    if [[ $progress -eq 100 ]]; then
        echo  # New line when complete
    fi
}

# @function start_spinner
# @brief Start a spinner for background operations
# @param $1 Message to display
start_spinner() {
    local message="$1"
    local spinner_chars="⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
    local delay=0.1
    
    # Start spinner in background
    (
        local i=0
        while true; do
            printf "\r${BLUE}[${spinner_chars:$i:1}]${NC} %s" "$message"
            i=$(((i + 1) % ${#spinner_chars}))
            sleep $delay
        done
    ) &
    
    # Store spinner PID
    SPINNER_PID=$!
}

# @function stop_spinner
# @brief Stop the spinner
stop_spinner() {
    if [[ -n "${SPINNER_PID:-}" ]]; then
        kill "$SPINNER_PID" 2>/dev/null || true
        wait "$SPINNER_PID" 2>/dev/null || true
        unset SPINNER_PID
        printf "\r\033[K"  # Clear line
    fi
}