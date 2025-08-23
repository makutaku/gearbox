#!/bin/bash

# zoxide Installation Script - Gearbox Standard Protocol
# Smart cd command (Rust)
# Usage: ./install-zoxide.sh [OPTIONS]

set -e  # Exit on any error

# Find the script directory and load common library
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")")"

# Source common library for shared functions
if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: common.sh not found in $REPO_DIR/scripts/lib/" >&2
    exit 1
fi

# Script metadata
SCRIPT_VERSION="1.0.0"
TOOL_NAME="zoxide"
TOOL_DESCRIPTION="Smart cd command (Rust)"

# Standard protocol variables
BUILD_TYPE="standard"
MODE="install" 
SKIP_DEPS=false
FORCE=false
RUN_TESTS=false
NO_SHELL=false
DRY_RUN=false
VERBOSE=false
QUIET=false

# Tool-specific configuration
RUST_MIN_VERSION="1.85.0"
INSTALL_FZF=false  # Legacy option

# Standard help display
show_help() {
    cat << EOF
$TOOL_NAME Installation Script - Gearbox Standard Protocol

DESCRIPTION:
    $TOOL_DESCRIPTION

USAGE:
    $0 [OPTIONS]

BUILD TYPES:
    --minimal           Fast build with essential features only
    --standard          Balanced build with reasonable features (default)
    --maximum           Full-featured build with all optimizations

EXECUTION MODES:
    --config-only       Configure only (prepare build environment)
    --build-only        Configure and build (no installation)
    --install           Configure, build, and install (default)

COMMON OPTIONS:
    --skip-deps         Skip dependency installation
    --force             Force reinstallation if already installed
    --run-tests         Run test suite after building
    --no-shell          Skip shell integration setup
    --dry-run           Show what would be done without executing
    --verbose           Enable verbose output
    --quiet             Suppress non-error output
    --help, -h          Show this help message
    --version           Show script version

LEGACY OPTIONS (deprecated):
    --install-fzf       Also install fzf via cargo (use ./install-fzf.sh instead)

EXAMPLES:
    $0                              # Standard build and install
    $0 --minimal --config-only      # Minimal build, config only  
    $0 --maximum --force            # Maximum build, force reinstall
    $0 --dry-run --verbose          # Preview with verbose output

EOF
}

# Standard argument parsing
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --minimal)
                BUILD_TYPE="minimal"
                shift
                ;;
            --standard)
                BUILD_TYPE="standard"
                shift
                ;;
            --maximum)
                BUILD_TYPE="maximum"
                shift
                ;;
            --config-only)
                MODE="config"
                shift
                ;;
            --build-only)
                MODE="build"
                shift
                ;;
            --install)
                MODE="install"
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --run-tests)
                RUN_TESTS=true
                shift
                ;;
            --no-shell)
                NO_SHELL=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --quiet)
                QUIET=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            --version)
                echo "$TOOL_NAME installation script version: $SCRIPT_VERSION"
                exit 0
                ;;
            # Legacy options for backward compatibility
            -c|--config-only-legacy)
                MODE="config"
                shift
                ;;
            -i|--install-legacy)
                MODE="install"
                shift
                ;;
            --install-fzf)
                INSTALL_FZF=true
                warning "Option --install-fzf is deprecated, use ./install-fzf.sh instead"
                shift
                ;;
            -*)
                warning "Unknown option: $1 (ignoring)"
                shift
                ;;
            *)
                error "Unexpected argument: $1"
                ;;
        esac
    done
}

# Parse arguments
parse_args "$@"

# Standard logging functions (override common.sh if needed)
log() {
    [[ $QUIET == false ]] && echo "📦 $TOOL_NAME: $@"
}

log_verbose() {
    [[ $VERBOSE == true ]] && echo "🔍 $TOOL_NAME: $@"
}

log_success() {
    [[ $QUIET == false ]] && success "$@"
}

log_error() {
    error "$@"
}

log_warning() {
    warning "$@"
}

# Dry run helper
execute() {
    if [[ $DRY_RUN == true ]]; then
        log_verbose "DRY RUN: $@"
    else
        log_verbose "Executing: $@"
        "$@"
    fi
}

ask_user() {
    while true; do
        read -p "$1 (y/n): " yn
        case $yn in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

# Version comparison function
version_compare() {
    if [[ $1 == $2 ]]; then
        return 0
    fi
    local IFS=.
    local i ver1=($1) ver2=($2)
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
        ver1[i]=0
    done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 0
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 1
        fi
    done
    return 0
}

# Check and install Rust
install_rust() {
    if command -v rustc &> /dev/null; then
        local current_version=$(rustc --version | cut -d' ' -f2)
        log "Found Rust version: $current_version"
        
        if version_compare $current_version $RUST_MIN_VERSION; then
            log "Rust version is sufficient (>= $RUST_MIN_VERSION)"
            return 0
        else
            warning "Rust version $current_version is below minimum required $RUST_MIN_VERSION"
            if ask_user "Do you want to update Rust?"; then
                log "Updating Rust..."
                rustup update || error "Failed to update Rust"
                success "Rust updated successfully"
            else
                error "Cannot proceed with insufficient Rust version"
            fi
        fi
    else
        log "Rust not found, installing..."
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y || error "Failed to install Rust"
        source ~/.cargo/env || error "Failed to source Rust environment"
        success "Rust installed successfully"
    fi
}

# Install dependencies
install_dependencies() {
    if [[ "$SKIP_DEPS" == true ]]; then
        log "Skipping dependency installation as requested"
        return 0
    fi

    # Update package list
    log "Updating package list..."
    sudo apt update || error "Failed to update package list"

    # Install basic build tools
    log "Installing build tools..."
    sudo apt install -y \
        build-essential \
        git \
        curl \
        || error "Failed to install build tools"

    # Install Rust
    install_rust

    success "Dependencies installed successfully"
}

# Setup shell integrations
setup_shell_integration() {
    if [[ "$SETUP_SHELL" == false ]]; then
        log "Skipping shell integration setup as requested"
        return 0
    fi
    
    log "Setting up shell integrations..."
    
    # Detect available shells
    local shells=()
    if command -v bash &> /dev/null; then
        shells+=("bash")
    fi
    if command -v zsh &> /dev/null; then
        shells+=("zsh")
    fi
    if command -v fish &> /dev/null; then
        shells+=("fish")
    fi
    
    log "Detected shells: ${shells[*]}"
    
    # Setup shell integration for each detected shell
    for shell in "${shells[@]}"; do
        case $shell in
            bash)
                log "Setting up bash integration..."
                if ! grep -q "zoxide init bash" ~/.bashrc 2>/dev/null; then
                    echo '# zoxide integration' >> ~/.bashrc
                    echo 'eval "$(zoxide init bash)"' >> ~/.bashrc
                    log "Added zoxide integration to ~/.bashrc"
                else
                    log "zoxide integration already exists in ~/.bashrc"
                fi
                ;;
            zsh)
                log "Setting up zsh integration..."
                if [[ -f ~/.zshrc ]]; then
                    if ! grep -q "zoxide init zsh" ~/.zshrc 2>/dev/null; then
                        echo '# zoxide integration' >> ~/.zshrc
                        echo 'eval "$(zoxide init zsh)"' >> ~/.zshrc
                        log "Added zoxide integration to ~/.zshrc"
                    else
                        log "zoxide integration already exists in ~/.zshrc"
                    fi
                fi
                ;;
            fish)
                log "Setting up fish integration..."
                local fish_config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish"
                if [[ -d "$fish_config_dir" ]]; then
                    local fish_config="$fish_config_dir/config.fish"
                    if ! grep -q "zoxide init fish" "$fish_config" 2>/dev/null; then
                        echo '# zoxide integration' >> "$fish_config"
                        echo 'zoxide init fish | source' >> "$fish_config"
                        log "Added zoxide integration to $fish_config"
                    else
                        log "zoxide integration already exists in $fish_config"
                    fi
                fi
                ;;
        esac
    done
    
    success "Shell integration setup completed"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

log "Starting zoxide $MODE process for Debian Linux"
log "Mode: $MODE"
log "Shell integration: $SETUP_SHELL"
log "Install fzf: $INSTALL_FZF"

# Install dependencies
install_dependencies

# Ensure cargo is in PATH
if ! command -v cargo &> /dev/null; then
    source ~/.cargo/env || error "Failed to source Rust environment"
fi

# Check if already installed
if command -v zoxide &> /dev/null && [[ "$FORCE_INSTALL" == false ]]; then
    log "zoxide is already installed: $(zoxide --version)"
    if ask_user "Do you want to reinstall zoxide?"; then
        FORCE_INSTALL=true
    else
        log "Skipping installation, proceeding to shell integration"
    fi
fi

log "Configuring zoxide installation..."
log "Using cargo install method for optimal performance"

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --install to proceed."
    success "Install command would be: cargo install zoxide --locked"
    if [[ "$INSTALL_FZF" == true ]]; then
        success "fzf install command would be: cargo install fzf"
    fi
    exit 0
fi

# Install zoxide
log "Installing zoxide via cargo..."

if [[ "$FORCE_INSTALL" == true ]]; then
    cargo install zoxide --locked --force || error "zoxide installation failed"
else
    cargo install zoxide --locked || error "zoxide installation failed"
fi

success "zoxide installed successfully"

# Install fzf if requested
if [[ "$INSTALL_FZF" == true ]]; then
    log "Installing fzf for enhanced directory selection..."
    if command -v fzf &> /dev/null && [[ "$FORCE_INSTALL" == false ]]; then
        log "fzf is already installed: $(fzf --version | head -n1)"
    else
        log "Installing fzf via cargo (fail-fast approach)..."
        if [[ "$FORCE_INSTALL" == true ]]; then
            cargo install fzf --force || error "fzf installation failed - use dedicated './install-fzf.sh' script for better fzf installation options"
        else
            cargo install fzf || error "fzf installation failed - use dedicated './install-fzf.sh' script for better fzf installation options"
        fi
        success "fzf installed successfully"
    fi
fi

# Add cargo bin to PATH if not already there
if [[ ":$PATH:" != *":$HOME/.cargo/bin:"* ]]; then
    log "Adding ~/.cargo/bin to PATH..."
    echo 'export PATH="$HOME/.cargo/bin:$PATH"' >> ~/.bashrc
    export PATH="$HOME/.cargo/bin:$PATH"
    warning "You may need to restart your shell or run 'source ~/.bashrc' for PATH changes to take effect"
fi

# Create system-wide symlink to ensure our version takes precedence
log "Creating system-wide symlink..."
sudo ln -sf "$HOME/.cargo/bin/zoxide" /usr/local/bin/zoxide || warning "Failed to create zoxide symlink"
success "Symlink created for zoxide command"

# Setup shell integrations
setup_shell_integration

# Verify installation
log "Verifying installation..."
# Force PATH update for verification
export PATH="/usr/local/bin:$HOME/.cargo/bin:$PATH"
# Clear bash command hash table to ensure new binary is used
hash -r
if command -v zoxide &> /dev/null; then
    success "zoxide installation verified!"
    echo
    log "zoxide version information:"
    zoxide --version
    echo
    success "zoxide installation completed successfully!"
    log "You can now use zoxide as a smarter 'cd' command"
    echo
    log "Basic usage:"
    log "  z <directory>                # Jump to directory"
    log "  zi                          # Interactive directory selection"
    if command -v fzf &> /dev/null; then
        log "  zi (with fzf)               # Enhanced interactive selection"
    fi
    echo
    if [[ "$SETUP_SHELL" == true ]]; then
        log "Shell integrations installed:"
        log "  'z' command for smart directory jumping"
        log "  'zi' command for interactive selection"
        echo
        log "To activate shell integrations immediately, run:"
        log "  hash -r && source ~/.bashrc"
        log "Or simply open a new terminal window"
    fi
    echo
    log "Installation paths:"
    log "  zoxide: $(which zoxide)"
    if command -v fzf &> /dev/null; then
        log "  fzf: $(which fzf)"
    fi
    echo
    log "Start using zoxide by navigating around, then use 'z <partial-name>' to jump back!"
    log "Script completed successfully"
else
    error "zoxide installation verification failed - try restarting your shell or run 'source ~/.bashrc'"
fi
