#!/bin/bash

# All Tools Installation Script
# Installs all development tools in optimal order with shared dependencies
# Usage: ./install-all-tools.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/lib/common.sh" ]]; then
    source "$REPO_DIR/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/lib/" >&2
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

# Available tools (validated list)
declare -A AVAILABLE_TOOLS_MAP=(
    ["ffmpeg"]="1" ["7zip"]="1" ["jq"]="1" ["fd"]="1" ["ripgrep"]="1" ["fzf"]="1" 
    ["imagemagick"]="1" ["yazi"]="1" ["zoxide"]="1" ["fclones"]="1" ["serena"]="1" 
    ["uv"]="1" ["ruff"]="1" ["bat"]="1" ["starship"]="1" ["eza"]="1" ["delta"]="1" 
    ["lazygit"]="1" ["bottom"]="1" ["procs"]="1" ["tokei"]="1" ["hyperfine"]="1" 
    ["gh"]="1" ["dust"]="1" ["sd"]="1" ["tealdeer"]="1" ["choose"]="1" 
    ["difftastic"]="1" ["bandwhich"]="1" ["xsv"]="1"
)

AVAILABLE_TOOLS=("ffmpeg" "7zip" "jq" "fd" "ripgrep" "fzf" "imagemagick" "yazi" "zoxide" "fclones" "serena" "uv" "ruff" "bat" "starship" "eza" "delta" "lazygit" "bottom" "procs" "tokei" "hyperfine" "gh" "dust" "sd" "tealdeer" "choose" "difftastic" "bandwhich" "xsv")
SELECTED_TOOLS=()

# Security functions
validate_tool_name() {
    local tool="$1"
    
    # Check if tool name is in allowed list
    [[ -z "${AVAILABLE_TOOLS_MAP[$tool]:-}" ]] && error "Invalid tool name: '$tool'. Use --help to see available tools."
    
    # Additional security check for tool name format
    [[ "$tool" =~ ^[a-zA-Z0-9-]+$ ]] || error "Invalid characters in tool name: '$tool'"
    
    return 0
}

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
            if [[ -n "${AVAILABLE_TOOLS_MAP[$1]:-}" ]]; then
                validate_tool_name "$1"
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

# Get build type flag for each tool
get_build_flag() {
    local tool=$1
    case $tool in
        ffmpeg)
            case $BUILD_TYPE in
                minimal) echo "-m" ;;
                standard) echo "-g" ;;
                maximum) echo "-x" ;;
            esac
            ;;
        7zip)
            case $BUILD_TYPE in
                minimal) echo "-b" ;;
                standard) echo "-o" ;;
                maximum) echo "-a" ;;
            esac
            ;;
        jq)
            case $BUILD_TYPE in
                minimal) echo "-m" ;;
                standard) echo "-s" ;;
                maximum) echo "-o" ;;  # optimized is closest to maximum
            esac
            ;;
        fd)
            case $BUILD_TYPE in
                minimal) echo "-m" ;;
                standard) echo "-r" ;;
                maximum) echo "-r" ;;  # release is the highest
            esac
            ;;
        ripgrep)
            case $BUILD_TYPE in
                minimal) echo "--no-pcre2" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;  # optimized
            esac
            ;;
        fzf)
            case $BUILD_TYPE in
                minimal) echo "-s" ;;
                standard) echo "-s" ;;
                maximum) echo "-p" ;;  # profiling build
            esac
            ;;
        imagemagick)
            case $BUILD_TYPE in
                minimal) echo "-m" ;;
                standard) echo "-f" ;;
                maximum) echo "-p" ;;
            esac
            ;;
        yazi)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-r" ;;  # release is highest for yazi
            esac
            ;;
        zoxide)
            # zoxide doesn't have build types, just installation modes
            # Return empty string to use default behavior
            echo ""
            ;;
        fclones)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        serena)
            case $BUILD_TYPE in
                minimal) echo "-m" ;;
                standard) echo "-s" ;;
                maximum) echo "-f" ;;
            esac
            ;;
        uv)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        ruff)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        bat)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        starship)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        eza)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        delta)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        lazygit)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        bottom)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        procs)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        tokei)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        hyperfine)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        gh)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        difftastic)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        bandwhich)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        xsv)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
        dust|sd|tealdeer|choose)
            # These tools use simple installation, no build flags needed
            echo ""
            ;;
    esac
}

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

# Add selected tools in optimal order
for tool in "fzf" "ripgrep" "fd" "zoxide" "yazi" "fclones" "uv" "ruff" "bat" "starship" "eza" "delta" "lazygit" "bottom" "procs" "tokei" "difftastic" "bandwhich" "xsv" "hyperfine" "gh" "dust" "sd" "tealdeer" "choose" "serena" "jq" "ffmpeg" "7zip" "imagemagick"; do
    if [[ " ${SELECTED_TOOLS[*]} " =~ " ${tool} " ]]; then
        INSTALLATION_ORDER+=("$tool")
    fi
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
    case $tool in
        ffmpeg)
            if command -v ffmpeg &> /dev/null; then
                log "✓ ffmpeg: $(ffmpeg -version 2>&1 | head -n1)"
            else
                FAILED_TOOLS+=("ffmpeg")
            fi
            ;;
        7zip)
            if command -v 7zz &> /dev/null; then
                log "✓ 7zz: $(7zz 2>&1 | head -n1)"
            else
                FAILED_TOOLS+=("7zip")
            fi
            ;;
        jq)
            if command -v jq &> /dev/null; then
                log "✓ jq: $(jq --version)"
            else
                FAILED_TOOLS+=("jq")
            fi
            ;;
        fd)
            if command -v fd &> /dev/null; then
                log "✓ fd: $(fd --version)"
            else
                FAILED_TOOLS+=("fd")
            fi
            ;;
        ripgrep)
            if command -v rg &> /dev/null; then
                log "✓ rg: $(rg --version | head -n1)"
            else
                FAILED_TOOLS+=("ripgrep")
            fi
            ;;
        fzf)
            if command -v fzf &> /dev/null; then
                log "✓ fzf: $(fzf --version)"
            else
                FAILED_TOOLS+=("fzf")
            fi
            ;;
        imagemagick)
            if command -v magick &> /dev/null; then
                log "✓ imagemagick: $(magick --version | head -n1)"
            else
                FAILED_TOOLS+=("imagemagick")
            fi
            ;;
        yazi)
            if command -v yazi &> /dev/null; then
                log "✓ yazi: $(yazi --version)"
            else
                FAILED_TOOLS+=("yazi")
            fi
            ;;
        zoxide)
            if command -v zoxide &> /dev/null; then
                log "✓ zoxide: $(zoxide --version)"
            else
                FAILED_TOOLS+=("zoxide")
            fi
            ;;
        fclones)
            if command -v fclones &> /dev/null; then
                log "✓ fclones: $(fclones --version | head -n1)"
            else
                FAILED_TOOLS+=("fclones")
            fi
            ;;
        serena)
            if command -v serena &> /dev/null; then
                log "✓ serena: $(serena --version 2>/dev/null | head -n1 || echo 'installed')"
            else
                FAILED_TOOLS+=("serena")
            fi
            ;;
        uv)
            if command -v uv &> /dev/null; then
                log "✓ uv: $(uv --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("uv")
            fi
            ;;
        ruff)
            if command -v ruff &> /dev/null; then
                log "✓ ruff: $(ruff --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("ruff")
            fi
            ;;
        bat)
            if command -v bat &> /dev/null; then
                log "✓ bat: $(bat --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("bat")
            fi
            ;;
        starship)
            if command -v starship &> /dev/null; then
                log "✓ starship: $(starship --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("starship")
            fi
            ;;
        eza)
            if command -v eza &> /dev/null; then
                log "✓ eza: $(eza --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("eza")
            fi
            ;;
        delta)
            if command -v delta &> /dev/null; then
                log "✓ delta: $(delta --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("delta")
            fi
            ;;
        lazygit)
            if command -v lazygit &> /dev/null; then
                log "✓ lazygit: $(lazygit --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("lazygit")
            fi
            ;;
        bottom)
            if command -v btm &> /dev/null; then
                log "✓ bottom: $(btm --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("bottom")
            fi
            ;;
        procs)
            if command -v procs &> /dev/null; then
                log "✓ procs: $(procs --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("procs")
            fi
            ;;
        tokei)
            if command -v tokei &> /dev/null; then
                log "✓ tokei: $(tokei --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("tokei")
            fi
            ;;
        hyperfine)
            if command -v hyperfine &> /dev/null; then
                log "✓ hyperfine: $(hyperfine --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("hyperfine")
            fi
            ;;
        gh)
            if command -v gh &> /dev/null; then
                log "✓ gh: $(gh --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("gh")
            fi
            ;;
        dust)
            if command -v dust &> /dev/null; then
                log "✓ dust: $(dust --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("dust")
            fi
            ;;
        sd)
            if command -v sd &> /dev/null; then
                log "✓ sd: $(sd --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("sd")
            fi
            ;;
        tealdeer)
            if command -v tldr &> /dev/null; then
                log "✓ tealdeer: $(tldr --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("tealdeer")
            fi
            ;;
        choose)
            if command -v choose &> /dev/null; then
                log "✓ choose: $(choose --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("choose")
            fi
            ;;
        difftastic)
            if command -v difft &> /dev/null; then
                log "✓ difftastic: $(difft --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("difftastic")
            fi
            ;;
        bandwhich)
            if command -v bandwhich &> /dev/null; then
                log "✓ bandwhich: $(bandwhich --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("bandwhich")
            fi
            ;;
        xsv)
            if command -v xsv &> /dev/null; then
                log "✓ xsv: $(xsv --version 2>/dev/null | head -n1)"
            else
                FAILED_TOOLS+=("xsv")
            fi
            ;;
    esac
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