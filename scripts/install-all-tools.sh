#!/bin/bash

# All Tools Installation Script
# Installs all development tools in optimal order with shared dependencies
# Usage: ./install-all-tools.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
SKIP_COMMON_DEPS=false
BUILD_TYPE="standard"  # minimal, standard, maximum
RUN_TESTS=false
SETUP_SHELL=true

# Available tools
AVAILABLE_TOOLS=("ffmpeg" "7zip" "jq" "fd" "ripgrep" "fzf" "imagemagick" "yazi" "zoxide" "fclones" "serena" "uv" "ruff" "bat" "starship" "eza" "delta" "lazygit" "bottom" "procs" "tokei")
SELECTED_TOOLS=()

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
            shift
            ;;
        --maximum)
            BUILD_TYPE="maximum"
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
        ffmpeg|7zip|jq|fd|ripgrep|fzf|imagemagick|yazi|zoxide|fclones|serena|uv|ruff|bat|starship|eza|delta|lazygit|bottom|procs|tokei)
            SELECTED_TOOLS+=("$1")
            shift
            ;;
        *)
            echo "Unknown option or tool: $1"
            show_help
            exit 1
            ;;
    esac
done

# If no tools selected, install all
if [[ ${#SELECTED_TOOLS[@]} -eq 0 ]]; then
    SELECTED_TOOLS=("${AVAILABLE_TOOLS[@]}")
fi

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

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
    local script="$SCRIPT_DIR/install-${tool}.sh"
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
    
    # Run the installation
    if [[ -n "$build_flag" ]]; then
        eval "$script $build_flag $extra_flags" || error "Failed to install $tool"
    else
        eval "$script $extra_flags" || error "Failed to install $tool"
    fi
    
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
for tool in "fzf" "ripgrep" "fd" "zoxide" "yazi" "fclones" "uv" "ruff" "bat" "starship" "eza" "delta" "lazygit" "bottom" "procs" "tokei" "serena" "jq" "ffmpeg" "7zip" "imagemagick"; do
    if [[ " ${SELECTED_TOOLS[*]} " =~ " ${tool} " ]]; then
        INSTALLATION_ORDER+=("$tool")
    fi
done

log "Installation order: ${INSTALLATION_ORDER[*]}"
echo

# Install each tool
for tool in "${INSTALLATION_ORDER[@]}"; do
    echo
    log "================================================"
    log "Installing $tool..."
    log "================================================"
    
    install_tool "$tool"
    
    echo
    success "$tool installation completed successfully!"
done

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