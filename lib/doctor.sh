#!/bin/bash

# Gearbox Health Check System (gearbox doctor)
# Comprehensive diagnostic tool for validating installation health

# Simple logging functions (fallback)
if ! command -v log &>/dev/null; then
    log() { echo "$1"; }
    error() { echo "ERROR: $1" >&2; }
    warning() { echo "WARNING: $1" >&2; }
    success() { echo "SUCCESS: $1"; }
fi

# Health check categories
declare -A HEALTH_CHECKS=(
    ["system"]="System Requirements"
    ["tools"]="Installed Tools"
    ["environment"]="Environment Variables"
    ["permissions"]="File Permissions"
    ["configuration"]="Configuration"
    ["cache"]="Build Cache"
    ["dependencies"]="System Dependencies"
)

# Global health status
HEALTH_STATUS="healthy"
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNING_CHECKS=0

# Check result tracking
declare -a CHECK_RESULTS=()

# Add check result
add_check_result() {
    local status="$1"
    local category="$2"
    local check_name="$3"
    local message="$4"
    
    CHECK_RESULTS+=("$status|$category|$check_name|$message")
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    case "$status" in
        "PASS")
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
            ;;
        "FAIL")
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
            HEALTH_STATUS="unhealthy"
            ;;
        "WARN")
            WARNING_CHECKS=$((WARNING_CHECKS + 1))
            if [[ "$HEALTH_STATUS" == "healthy" ]]; then
                HEALTH_STATUS="warning"
            fi
            ;;
    esac
}

# System requirements checks
check_system_requirements() {
    log "Checking system requirements..."
    
    # Check operating system
    if [[ -f /etc/os-release ]]; then
        local os_name
        os_name=$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '"')
        if [[ "$os_name" =~ Ubuntu|Debian ]]; then
            add_check_result "PASS" "system" "Operating System" "Supported OS: $os_name"
        else
            add_check_result "WARN" "system" "Operating System" "Untested OS: $os_name (may work)"
        fi
    else
        add_check_result "FAIL" "system" "Operating System" "Cannot determine OS version"
    fi
    
    # Check architecture
    local arch
    arch=$(uname -m)
    if [[ "$arch" =~ x86_64|amd64 ]]; then
        add_check_result "PASS" "system" "Architecture" "Supported architecture: $arch"
    else
        add_check_result "WARN" "system" "Architecture" "Untested architecture: $arch"
    fi
    
    # Check available memory
    local mem_gb
    mem_gb=$(free -g | awk '/^Mem:/{print $2}')
    if [[ $mem_gb -ge 4 ]]; then
        add_check_result "PASS" "system" "Memory" "Sufficient memory: ${mem_gb}GB"
    elif [[ $mem_gb -ge 2 ]]; then
        add_check_result "WARN" "system" "Memory" "Limited memory: ${mem_gb}GB (recommend 4GB+)"
    else
        add_check_result "FAIL" "system" "Memory" "Insufficient memory: ${mem_gb}GB (need 2GB+)"
    fi
    
    # Check available disk space
    local disk_gb
    disk_gb=$(df -BG . | awk 'NR==2{print $4}' | tr -d 'G')
    if [[ $disk_gb -ge 5 ]]; then
        add_check_result "PASS" "system" "Disk Space" "Sufficient space: ${disk_gb}GB available"
    elif [[ $disk_gb -ge 2 ]]; then
        add_check_result "WARN" "system" "Disk Space" "Limited space: ${disk_gb}GB (recommend 5GB+)"
    else
        add_check_result "FAIL" "system" "Disk Space" "Insufficient space: ${disk_gb}GB (need 2GB+)"
    fi
    
    # Check internet connectivity
    if command -v curl &>/dev/null; then
        if curl -s --connect-timeout 5 https://github.com >/dev/null; then
            add_check_result "PASS" "system" "Internet" "Internet connectivity working"
        else
            add_check_result "FAIL" "system" "Internet" "No internet connectivity (needed for downloads)"
        fi
    elif command -v wget &>/dev/null; then
        if wget -q --timeout=5 --spider https://github.com; then
            add_check_result "PASS" "system" "Internet" "Internet connectivity working"
        else
            add_check_result "FAIL" "system" "Internet" "No internet connectivity (needed for downloads)"
        fi
    else
        add_check_result "WARN" "system" "Internet" "Cannot test connectivity (curl/wget missing)"
    fi
}

# Check installed tools
check_installed_tools() {
    log "Checking installed tools..."
    
    local tools=(
        "fd:Fast file finder"
        "rg:ripgrep text search"
        "fzf:Fuzzy finder"
        "jq:JSON processor"
        "yazi:Terminal file manager"
        "zoxide:Smart cd command"
        "bat:Enhanced cat"
        "eza:Modern ls"
        "delta:Git diff pager"
        "starship:Shell prompt"
        "lazygit:Git UI"
        "bottom:System monitor"
        "procs:Process viewer"
        "tokei:Code statistics"
        "hyperfine:Benchmarking"
        "gh:GitHub CLI"
        "dust:Disk usage"
        "sd:Search & replace"
        "tldr:Help pages"
        "choose:Text selection"
        "difft:Structural diff"
        "bandwhich:Network monitor"
        "xsv:CSV toolkit"
        "uv:Python package manager"
        "ruff:Python linter"
        "fclones:Duplicate finder"
        "serena:Coding agent"
        "ffmpeg:Video processing"
        "magick:Image processing"
        "7zz:Compression"
    )
    
    local installed_count=0
    local total_tools=${#tools[@]}
    
    for tool_desc in "${tools[@]}"; do
        IFS=':' read -r tool_name description <<< "$tool_desc"
        
        if command -v "$tool_name" &>/dev/null; then
            local version=""
            case "$tool_name" in
                fd) version=$(fd --version 2>/dev/null | head -1) ;;
                rg) version=$(rg --version 2>/dev/null | head -1) ;;
                fzf) version=$(fzf --version 2>/dev/null) ;;
                jq) version=$(jq --version 2>/dev/null) ;;
                *) version="installed" ;;
            esac
            add_check_result "PASS" "tools" "$tool_name" "$description: $version"
            installed_count=$((installed_count + 1))
        else
            add_check_result "WARN" "tools" "$tool_name" "$description: Not installed"
        fi
    done
    
    # Overall tool installation status
    local coverage_percent=$((installed_count * 100 / total_tools))
    if [[ $coverage_percent -ge 80 ]]; then
        add_check_result "PASS" "tools" "Coverage" "Excellent tool coverage: ${installed_count}/${total_tools} (${coverage_percent}%)"
    elif [[ $coverage_percent -ge 50 ]]; then
        add_check_result "WARN" "tools" "Coverage" "Good tool coverage: ${installed_count}/${total_tools} (${coverage_percent}%)"
    else
        add_check_result "WARN" "tools" "Coverage" "Limited tool coverage: ${installed_count}/${total_tools} (${coverage_percent}%)"
    fi
}

# Check environment variables
check_environment() {
    log "Checking environment variables..."
    
    # Check PATH
    if [[ ":$PATH:" == *":/usr/local/bin:"* ]]; then
        add_check_result "PASS" "environment" "PATH /usr/local/bin" "PATH includes /usr/local/bin"
    else
        add_check_result "WARN" "environment" "PATH /usr/local/bin" "PATH missing /usr/local/bin"
    fi
    
    if [[ ":$PATH:" == *":$HOME/.cargo/bin:"* ]]; then
        add_check_result "PASS" "environment" "PATH cargo" "PATH includes ~/.cargo/bin"
    else
        add_check_result "WARN" "environment" "PATH cargo" "PATH missing ~/.cargo/bin (for Rust tools)"
    fi
    
    # Check shell
    if [[ -n "$SHELL" ]]; then
        add_check_result "PASS" "environment" "Shell" "Shell: $SHELL"
    else
        add_check_result "WARN" "environment" "Shell" "Shell not set"
    fi
    
    # Check editor
    if [[ -n "$EDITOR" ]]; then
        add_check_result "PASS" "environment" "Editor" "Editor: $EDITOR"
    else
        add_check_result "WARN" "environment" "Editor" "EDITOR not set (recommend setting)"
    fi
    
    # Check gearbox configuration variables
    local config_vars=(
        "GEARBOX_DEFAULT_BUILD_TYPE"
        "GEARBOX_MAX_PARALLEL_JOBS"
        "GEARBOX_CACHE_ENABLED"
    )
    
    for var in "${config_vars[@]}"; do
        if [[ -n "${!var}" ]]; then
            add_check_result "PASS" "environment" "$var" "$var: ${!var}"
        else
            add_check_result "WARN" "environment" "$var" "$var: Not set (using default)"
        fi
    done
}

# Check file permissions
check_permissions() {
    log "Checking file permissions..."
    
    # Check if running as root (should not be)
    if [[ $EUID -eq 0 ]]; then
        add_check_result "FAIL" "permissions" "Root User" "Running as root (security risk)"
    else
        add_check_result "PASS" "permissions" "Root User" "Not running as root (good)"
    fi
    
    # Check write permissions to common directories
    local test_dirs=(
        "$HOME:Home directory"
        "/tmp:Temporary directory"
        ".:Current directory"
    )
    
    for dir_desc in "${test_dirs[@]}"; do
        IFS=':' read -r dir_path description <<< "$dir_desc"
        
        if [[ -w "$dir_path" ]]; then
            add_check_result "PASS" "permissions" "Write $dir_path" "$description: Writable"
        else
            add_check_result "FAIL" "permissions" "Write $dir_path" "$description: Not writable"
        fi
    done
    
    # Check sudo access
    if sudo -n true 2>/dev/null; then
        add_check_result "PASS" "permissions" "Sudo" "Sudo access available"
    else
        add_check_result "WARN" "permissions" "Sudo" "Sudo access needed for installation"
    fi
}

# Check configuration
check_configuration() {
    log "Checking configuration..."
    
    local config_file="$HOME/.gearboxrc"
    if [[ -f "$config_file" ]]; then
        if [[ -r "$config_file" ]]; then
            local config_size
            config_size=$(stat -c%s "$config_file" 2>/dev/null || echo "0")
            add_check_result "PASS" "configuration" "Config File" "Configuration exists: $config_file (${config_size} bytes)"
            
            # Validate config file syntax
            if grep -qE '^[A-Z_]+=.+$' "$config_file"; then
                add_check_result "PASS" "configuration" "Config Syntax" "Configuration file has valid syntax"
            else
                add_check_result "WARN" "configuration" "Config Syntax" "Configuration file may have syntax issues"
            fi
        else
            add_check_result "WARN" "configuration" "Config File" "Configuration exists but not readable"
        fi
    else
        add_check_result "WARN" "configuration" "Config File" "No configuration file (using defaults)"
    fi
    
    # Check common library
    local common_lib="./lib/common.sh"
    if [[ -f "$common_lib" ]]; then
        add_check_result "PASS" "configuration" "Common Library" "Common library exists: $common_lib"
    else
        add_check_result "FAIL" "configuration" "Common Library" "Common library missing: $common_lib"
    fi
}

# Check build cache
check_cache() {
    log "Checking build cache..."
    
    local cache_dir="$HOME/tools/cache"
    if [[ -d "$cache_dir" ]]; then
        local cache_size
        cache_size=$(du -sh "$cache_dir" 2>/dev/null | cut -f1)
        local cache_files
        cache_files=$(find "$cache_dir" -type f 2>/dev/null | wc -l)
        add_check_result "PASS" "cache" "Cache Directory" "Cache exists: $cache_dir ($cache_size, $cache_files files)"
        
        # Check for old cache files
        local old_files
        old_files=$(find "$cache_dir" -type f -mtime +7 2>/dev/null | wc -l)
        if [[ $old_files -gt 0 ]]; then
            add_check_result "WARN" "cache" "Cache Cleanup" "$old_files cache files older than 7 days (consider cleanup)"
        else
            add_check_result "PASS" "cache" "Cache Cleanup" "Cache is clean (no files older than 7 days)"
        fi
    else
        add_check_result "WARN" "cache" "Cache Directory" "Cache directory not found (will be created when needed)"
    fi
}

# Check system dependencies
check_dependencies() {
    log "Checking system dependencies..."
    
    local build_deps=(
        "git:Git version control"
        "curl:HTTP client"
        "build-essential:Build tools"
        "pkg-config:Package config"
        "cmake:CMake build system"
        "autoconf:Autotools"
        "automake:Automake"
        "libtool:Libtool"
    )
    
    for dep_desc in "${build_deps[@]}"; do
        IFS=':' read -r dep_name description <<< "$dep_desc"
        
        case "$dep_name" in
            "build-essential")
                if command -v gcc &>/dev/null && command -v make &>/dev/null; then
                    add_check_result "PASS" "dependencies" "$dep_name" "$description: Available"
                else
                    add_check_result "WARN" "dependencies" "$dep_name" "$description: Missing (install with apt)"
                fi
                ;;
            *)
                if command -v "$dep_name" &>/dev/null; then
                    add_check_result "PASS" "dependencies" "$dep_name" "$description: Available"
                else
                    add_check_result "WARN" "dependencies" "$dep_name" "$description: Missing (install with apt)"
                fi
                ;;
        esac
    done
    
    # Check language toolchains
    local toolchains=(
        "rustc:Rust compiler"
        "cargo:Rust package manager"
        "go:Go compiler"
        "node:Node.js"
        "python3:Python 3"
    )
    
    for tool_desc in "${toolchains[@]}"; do
        IFS=':' read -r tool_name description <<< "$tool_desc"
        
        if command -v "$tool_name" &>/dev/null; then
            local version=""
            case "$tool_name" in
                rustc) version=$(rustc --version 2>/dev/null) ;;
                cargo) version=$(cargo --version 2>/dev/null) ;;
                go) version=$(go version 2>/dev/null) ;;
                node) version=$(node --version 2>/dev/null) ;;
                python3) version=$(python3 --version 2>/dev/null) ;;
            esac
            add_check_result "PASS" "dependencies" "$tool_name" "$description: $version"
        else
            add_check_result "WARN" "dependencies" "$tool_name" "$description: Not installed"
        fi
    done
}

# Display results
display_results() {
    echo
    echo "=============================================="
    echo "           GEARBOX HEALTH CHECK REPORT"
    echo "=============================================="
    echo
    
    # Overall status
    case "$HEALTH_STATUS" in
        "healthy")
            success "Overall Status: HEALTHY ✓"
            ;;
        "warning")
            warning "Overall Status: WARNING ⚠"
            ;;
        "unhealthy")
            echo "Overall Status: UNHEALTHY ✗"
            ;;
    esac
    
    echo
    echo "Summary: $PASSED_CHECKS passed, $WARNING_CHECKS warnings, $FAILED_CHECKS failed (total: $TOTAL_CHECKS checks)"
    echo
    
    # Display results by category
    for category in "${!HEALTH_CHECKS[@]}"; do
        local category_name="${HEALTH_CHECKS[$category]}"
        echo "--- $category_name ---"
        
        for result in "${CHECK_RESULTS[@]}"; do
            IFS='|' read -r status result_category check_name message <<< "$result"
            
            if [[ "$result_category" == "$category" ]]; then
                case "$status" in
                    "PASS")
                        echo "  ✓ $check_name: $message"
                        ;;
                    "WARN")
                        echo "  ⚠ $check_name: $message"
                        ;;
                    "FAIL")
                        echo "  ✗ $check_name: $message"
                        ;;
                esac
            fi
        done
        echo
    done
    
    # Recommendations
    if [[ $FAILED_CHECKS -gt 0 || $WARNING_CHECKS -gt 0 ]]; then
        echo "--- Recommendations ---"
        
        if [[ $FAILED_CHECKS -gt 0 ]]; then
            echo "🔧 Critical issues found that may prevent gearbox from working properly."
            echo "   Please address all FAILED checks before proceeding with installations."
        fi
        
        if [[ $WARNING_CHECKS -gt 0 ]]; then
            echo "💡 Some warnings detected. While not critical, addressing these may improve"
            echo "   your experience with gearbox:"
            
            # Specific recommendations based on warnings
            for result in "${CHECK_RESULTS[@]}"; do
                IFS='|' read -r status result_category check_name message <<< "$result"
                
                if [[ "$status" == "WARN" ]]; then
                    case "$check_name" in
                        "PATH"*)
                            echo "   - Add missing directories to your PATH in ~/.bashrc"
                            ;;
                        "Editor")
                            echo "   - Set EDITOR environment variable: export EDITOR=nano"
                            ;;
                        *"coverage"*)
                            echo "   - Install more tools with: gearbox install"
                            ;;
                        *"cache"*)
                            echo "   - Clean old cache files with: gearbox config set CACHE_MAX_AGE_DAYS 3"
                            ;;
                        *"dependency"*)
                            echo "   - Install missing dependencies with: sudo apt update && sudo apt install $check_name"
                            ;;
                    esac
                fi
            done
        fi
        echo
    fi
    
    echo "For more help: gearbox config help"
    echo "For tool installation: gearbox install [tool-name]"
    echo "=============================================="
}

# Main health check function
run_health_check() {
    local categories="${1:-all}"
    
    echo "🔍 Running Gearbox health check..."
    echo
    
    if [[ "$categories" == "all" ]]; then
        check_system_requirements
        check_installed_tools
        check_environment
        check_permissions
        check_configuration
        check_cache
        check_dependencies
    else
        # Run specific categories
        IFS=',' read -ra CATS <<< "$categories"
        for cat in "${CATS[@]}"; do
            case "$cat" in
                "system") check_system_requirements ;;
                "tools") check_installed_tools ;;
                "environment") check_environment ;;
                "permissions") check_permissions ;;
                "configuration") check_configuration ;;
                "cache") check_cache ;;
                "dependencies") check_dependencies ;;
                *) warning "Unknown category: $cat" ;;
            esac
        done
    fi
    
    display_results
    
    # Return appropriate exit code
    case "$HEALTH_STATUS" in
        "healthy") return 0 ;;
        "warning") return 1 ;;
        "unhealthy") return 2 ;;
    esac
}