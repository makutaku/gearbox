#!/bin/bash

# All Tools Installation Script for Debian Linux
# Installs all development tools in optimal order with shared dependencies
# Usage: ./install-all-tools.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/scripts/lib/" >&2
    exit 1
fi

# Source configuration helpers
if [[ -f "$REPO_DIR/scripts/lib/config-helpers.sh" ]]; then
    source "$REPO_DIR/scripts/lib/config-helpers.sh"
else
    echo "ERROR: config-helpers.sh not found in $REPO_DIR/scripts/lib/" >&2
    exit 1
fi

# Note: Logging functions now provided by lib/common.sh

# Default options (can be overridden by configuration)
SKIP_COMMON_DEPS=false
BUILD_TYPE="${GEARBOX_DEFAULT_BUILD_TYPE:-standard}"  # minimal, standard, maximum
RUN_TESTS="${GEARBOX_SKIP_TESTS_BY_DEFAULT:-false}"
SETUP_SHELL="${GEARBOX_SHELL_INTEGRATION:-true}"

# Invert the skip tests logic (config is SKIP_TESTS_BY_DEFAULT, we use RUN_TESTS)
if [[ "${GEARBOX_SKIP_TESTS_BY_DEFAULT:-false}" == "true" ]]; then
    RUN_TESTS=false
else
    RUN_TESTS=false  # Default to false unless explicitly requested
fi

# Load available tools from configuration
AVAILABLE_TOOLS=($(jq -r '.tools[].name' "$REPO_DIR/config/tools.json" | sort))
SELECTED_TOOLS=()

# Security functions - now use config helper
# validate_tool_name function is provided by config-helpers.sh

validate_build_type() {
    local build_type="$1"
    
    case "$build_type" in
        minimal|standard|maximum)
            return 0
            ;;
        *)
            error "Invalid build type: '$build_type'. Valid options: minimal, standard, maximum"
            ;;
    esac
}

# Show help
show_help() {
    cat << EOF
All Tools Installation Script

Usage: $0 [OPTIONS] [TOOLS...]

Options:
  --skip-common-deps   Skip common dependency installation
  --minimal           Use minimal build types where available
  --maximum           Use maximum feature build types where available
  --run-tests         Run test suites for tools that support it
  --no-shell          Skip shell integration setup (fzf)
  -h, --help          Show this help message

Tools (install all if none specified):
  fd                  Fast file finder
  ripgrep             Fast text search
  fzf                 Fuzzy finder
  jq                  JSON processor
  zoxide              Smart cd command
  yazi                Terminal file manager
  fclones             Duplicate file finder
  serena              Coding agent toolkit
  uv                  Python package manager
  ruff                Python linter and formatter
  bat                 Enhanced cat with syntax highlighting
  starship            Customizable shell prompt
  eza                 Modern ls replacement with Git integration
  delta               Syntax-highlighting pager for Git and diff output
  lazygit             Terminal UI for Git operations
  bottom              Cross-platform system monitor
  procs               Modern ps replacement with tree view
  tokei               Code statistics and line counting tool
  difftastic          Structural diff tool for code analysis
  bandwhich           Network bandwidth monitor by process
  xsv                 Fast CSV data toolkit
  hyperfine           Command-line benchmarking tool
  gh                  GitHub CLI for repository management
  dust                Better disk usage analyzer
  sd                  Intuitive find & replace CLI
  tealdeer            Fast tldr pages client
  choose              Human-friendly cut/awk alternative
  ffmpeg              Video/audio processing suite
  imagemagick         Image manipulation toolkit
  7zip                Compression tool

Examples:
  $0                              # Install all tools with standard builds
  $0 --minimal fd ripgrep fzf     # Install only specified tools with minimal builds
  $0 --maximum --run-tests        # Install all tools with maximum features and tests
  $0 --skip-common-deps           # Skip common dependency installation

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-common-deps)
            SKIP_COMMON_DEPS=true
            shift
            ;;
        --minimal)
            BUILD_TYPE="minimal"
            validate_build_type "$BUILD_TYPE"
            shift
            ;;
        --maximum)
            BUILD_TYPE="maximum"
            validate_build_type "$BUILD_TYPE"
            shift
            ;;
        --run-tests)
            RUN_TESTS=true
            shift
            ;;
        --no-shell)
            SETUP_SHELL=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            # Check if it's a valid tool name
            if validate_tool_name "$1" 2>/dev/null; then
                SELECTED_TOOLS+=("$1")
                shift
            else
                error "Unknown option or tool: '$1'. Use --help to see available options and tools."
            fi
            ;;
    esac
done

# If no tools selected, install all
if [[ ${#SELECTED_TOOLS[@]} -eq 0 ]]; then
    SELECTED_TOOLS=("${AVAILABLE_TOOLS[@]}")
fi

# Logging function
# Note: Logging functions provided by lib/common.sh above

# Get build flag function is now provided by config-helpers.sh
# This replaces 188 lines of nested case statements with data-driven configuration

# Check if script exists
check_script() {
    local tool=$1
    local script="$SCRIPT_DIR/install-${tool}.sh"
    if [[ ! -f "$script" ]]; then
        error "Installation script not found: $script"
    fi
    if [[ ! -x "$script" ]]; then
        error "Installation script not executable: $script"
    fi
}

# Install a single tool
install_tool() {
    local tool=$1
    
    # Validate tool name for security
    validate_tool_name "$tool"
    
    # Construct script path securely
    local script="$SCRIPT_DIR/install-${tool}.sh"
    
    # Validate script exists and is executable
    [[ -f "$script" ]] || error "Installation script not found: $script"
    [[ -x "$script" ]] || error "Installation script not executable: $script"
    
    local build_flag=$(get_build_flag "$tool")
    local extra_flags=""
    
    # Add common flags
    if [[ "$SKIP_COMMON_DEPS" == true ]] || [[ "$tool" != "${SELECTED_TOOLS[0]}" ]]; then
        extra_flags="$extra_flags --skip-deps"
    fi
    
    if [[ "$RUN_TESTS" == true ]]; then
        extra_flags="$extra_flags --run-tests"
    fi
    
    # Tool-specific flags
    case $tool in
        fzf)
            if [[ "$SETUP_SHELL" == false ]]; then
                extra_flags="$extra_flags --no-shell"
            fi
            ;;
    esac
    
    log "Installing $tool with flags: $build_flag $extra_flags"
    
    # Prepare arguments array for secure execution
    local args=()
    [[ -n "$build_flag" ]] && args+=("$build_flag")
    
    # Parse extra_flags safely into array
    if [[ -n "$extra_flags" ]]; then
        # Split extra_flags by spaces, but handle quoted arguments properly
        local flag
        for flag in $extra_flags; do
            args+=("$flag")
        done
    fi
    
    # Run the installation securely without eval
    log "Executing: $script ${args[*]}"
    "$script" "${args[@]}" || error "Failed to install $tool"
    
    success "$tool installation completed"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "=================================="
log "All Tools Installation Script"
log "=================================="
log "Build type: $BUILD_TYPE"
log "Selected tools: ${SELECTED_TOOLS[*]}"
log "Run tests: $RUN_TESTS"
log "Shell integration: $SETUP_SHELL"
log "Skip common deps: $SKIP_COMMON_DEPS"
echo

# Verify all scripts exist
log "Verifying installation scripts..."
for tool in "${SELECTED_TOOLS[@]}"; do
    check_script "$tool"
done
success "All installation scripts found and executable"

# Install common dependencies first (unless skipped)
if [[ "$SKIP_COMMON_DEPS" == false ]]; then
    log "Installing common dependencies..."
    if [[ -f "$SCRIPT_DIR/install-common-deps.sh" ]] && [[ -x "$SCRIPT_DIR/install-common-deps.sh" ]]; then
        "$SCRIPT_DIR/install-common-deps.sh" || error "Failed to install common dependencies"
        success "Common dependencies installed"
        echo
        log "Sourcing environment for subsequent installations..."
        source ~/.bashrc || warning "Failed to source ~/.bashrc"
    else
        warning "Common dependencies script not found, will install dependencies individually"
    fi
fi

# Install tools in optimal order
# Optimal order: Go tools first, then Rust tools (using same toolchain), then C/C++ tools

# Define installation order for optimal dependency handling
INSTALLATION_ORDER=()

# Get installation order from configuration (grouped by language for efficiency)
# Go tools first, then Rust tools, then C/C++ tools
for language in "go" "rust" "python" "c"; do
    while IFS= read -r tool; do
        if [[ " ${SELECTED_TOOLS[*]} " =~ " ${tool} " ]]; then
            INSTALLATION_ORDER+=("$tool")
        fi
    done < <(jq -r ".tools[] | select(.language == \"$language\") | .name" "$REPO_DIR/config/tools.json")
done

log "Installation order: ${INSTALLATION_ORDER[*]}"
echo

# Install each tool with progress tracking
echo
log "Installing ${#INSTALLATION_ORDER[@]} tools..."
echo

current_tool=0
for tool in "${INSTALLATION_ORDER[@]}"; do
    current_tool=$((current_tool + 1))
    
    echo
    show_progress "$current_tool" "${#INSTALLATION_ORDER[@]}" "Installing $tool..."
    echo
    log_step "$current_tool" "${#INSTALLATION_ORDER[@]}" "Starting $tool installation"
    
    install_tool "$tool"
    
    success "$tool installation completed successfully!"
done

echo
show_progress "${#INSTALLATION_ORDER[@]}" "${#INSTALLATION_ORDER[@]}" "All tools installed successfully!"

# Final setup
echo
log "================================================"
log "Finalizing installation..."
log "================================================"

# Update library cache once at the end
log "Updating system library cache..."
sudo ldconfig || warning "Failed to update library cache"

# Clear command hash table
log "Clearing command hash table..."
hash -r

# Final verification
echo
log "Verifying installations..."
FAILED_TOOLS=()

for tool in "${SELECTED_TOOLS[@]}"; do
    # Get binary name and test command from configuration
    binary_name=$(jq -r ".tools[] | select(.name == \"$tool\") | .binary_name" "$REPO_DIR/config/tools.json")
    test_command=$(jq -r ".tools[] | select(.name == \"$tool\") | .test_command" "$REPO_DIR/config/tools.json")
    
    if [[ -n "$binary_name" && "$binary_name" != "null" ]]; then
        if command -v "$binary_name" &> /dev/null; then
            if [[ -n "$test_command" && "$test_command" != "null" ]]; then
                version_info=$($test_command 2>/dev/null | head -n1 || echo "installed")
                log "✓ $tool: $version_info"
            else
                log "✓ $tool: installed"
            fi
        else
            FAILED_TOOLS+=("$tool")
        fi
    else
        # Fallback for tools without binary_name configured
        FAILED_TOOLS+=("$tool")
    fi
done

echo
if [[ ${#FAILED_TOOLS[@]} -eq 0 ]]; then
    success "All selected tools installed and verified successfully!"
    echo
    log "🎉 Installation Summary:"
    log "  - Build type: $BUILD_TYPE"
    log "  - Tools installed: ${SELECTED_TOOLS[*]}"
    log "  - All tools are ready to use!"
    echo
    log "💡 Pro tip: Open a new terminal or run 'source ~/.bashrc' to ensure"
    log "   all PATH changes and shell integrations are active."
    echo
    log "🚀 Happy coding!"
else
    warning "Some tools failed verification: ${FAILED_TOOLS[*]}"
    log "You may need to restart your shell or run 'source ~/.bashrc'"
    exit 1
fi