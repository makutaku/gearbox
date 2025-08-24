#!/bin/bash
# claude-monitor Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-claude-monitor.sh [OPTIONS]

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

# Configuration
CLAUDE_MONITOR_REPO="https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor.git"
PYTHON_MIN_VERSION="3.9.0"

# Default options
BUILD_TYPE="standard"     # minimal, standard, maximum
MODE="install"            # config, build, install
INSTALL_METHOD="uv"       # uv (recommended), pip, source
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
VERBOSE=false
QUIET=false
DRY_RUN=false
NO_SHELL=false

# Show help
show_help() {
    cat << EOF
claude-monitor Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  --minimal            Minimal claude-monitor installation (basic features)
  --standard           Standard claude-monitor installation (default, recommended)
  --maximum            Maximum claude-monitor installation (with development tools)

Modes:
  --config-only        Configure only (prepare build environment)
  --build-only         Configure and build (no installation)
  --install            Configure, build, and install (default)

Installation Methods:
  --uv                 Use uv tool install (default, recommended)
  --pip                Use pip install (global or user-local)
  --source             Build from source (development)

Options:
  --skip-deps          Skip dependency installation
  --force              Force reinstallation if already installed
  --run-tests          Run test suite after building (source builds only)
  --no-shell           Skip shell integration setup
  --dry-run            Show what would be done without executing
  --verbose            Enable verbose output
  --quiet              Suppress non-error output
  -h, --help           Show this help message
  --version            Show script version information

Examples:
  $0                   # Default: uv tool install (recommended)
  $0 --pip             # Use pip for installation
  $0 --source          # Build from source (for development)
  $0 --minimal --uv    # Minimal uv installation
  $0 --maximum --source# Full source build with dev tools
  $0 --dry-run         # Preview installation steps

About claude-monitor:
  A real-time terminal monitoring tool for tracking Claude AI token usage,
  providing advanced analytics and intelligent predictions about session limits.
  Features ML-based usage predictions, rich color-coded UI, and support for
  multiple Claude subscription plans with configurable display options.

EOF
}

# Parse command line arguments
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
        --uv)
            INSTALL_METHOD="uv"
            shift
            ;;
        --pip)
            INSTALL_METHOD="pip"
            shift
            ;;
        --source)
            INSTALL_METHOD="source"
            shift
            ;;
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --run-tests)
            RUN_TESTS=true
            shift
            ;;
        --force)
            FORCE_INSTALL=true
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
        -h|--help)
            show_help
            exit 0
            ;;
        --version)
            echo "claude-monitor Installation Script v1.0"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

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

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
if [[ "$QUIET" != true ]]; then
    echo
    log "==================================="
    log "claude-monitor Installation Script"
    log "==================================="
    log "Build type: $BUILD_TYPE"
    log "Mode: $MODE"
    log "Install method: $INSTALL_METHOD"
    log "Skip dependencies: $SKIP_DEPS"
    log "Run tests: $RUN_TESTS"
    log "Dry run: $DRY_RUN"
    echo
fi

# Check if claude-monitor is already installed
if command -v claude-monitor &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(claude-monitor --version 2>/dev/null | head -n1 || echo "unknown")
    
    log "claude-monitor is already installed (version: $CURRENT_VERSION)"
    log "Use --force to reinstall"
    exit 0
fi

# Check Python version
if command -v python3 &> /dev/null; then
    PYTHON_VERSION=$(python3 --version 2>&1 | grep -oP 'Python \K\d+\.\d+\.\d+' || echo "0.0.0")
    if ! version_compare "$PYTHON_VERSION" "$PYTHON_MIN_VERSION"; then
        error "Python version $PYTHON_VERSION is too old. Minimum required: $PYTHON_MIN_VERSION"
    fi
    log "Python version $PYTHON_VERSION is sufficient"
else
    error "Python 3 is required but not found"
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies..."
    
    if [[ "$DRY_RUN" == true ]]; then
        log "[DRY RUN] Would install: python3 python3-pip python3-venv"
        if [[ "$INSTALL_METHOD" == "uv" ]]; then
            log "[DRY RUN] Would install uv package manager"
        fi
        if [[ "$INSTALL_METHOD" == "source" ]]; then
            log "[DRY RUN] Would install: git build-essential"
        fi
    else
        # Update package list
        sudo apt-get update -qq
        
        # Install required packages
        sudo apt-get install -y python3 python3-pip python3-venv
        
        # Install uv if using uv method
        if [[ "$INSTALL_METHOD" == "uv" ]]; then
            if ! command -v uv &> /dev/null; then
                log "Installing uv package manager..."
                curl -LsSf https://astral.sh/uv/install.sh | sh || error "Failed to install uv"
                export PATH="$HOME/.local/bin:$PATH"
                
                # Update shell profiles for persistent uv access
                if [[ "$NO_SHELL" != true ]]; then
                    for profile in ~/.bashrc ~/.zshrc ~/.profile; do
                        if [[ -f "$profile" ]] && ! grep -q ".local/bin" "$profile"; then
                            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$profile"
                            log "Updated $profile to include ~/.local/bin in PATH"
                        fi
                    done
                fi
            else
                log "uv is already installed"
            fi
        fi
        
        # Install additional dependencies for source builds
        if [[ "$INSTALL_METHOD" == "source" ]]; then
            sudo apt-get install -y git build-essential
        fi
        
        log "Dependencies installed successfully"
    fi
else
    log "Skipping dependency installation"
fi

# Configuration mode
if [[ "$MODE" == "config" ]]; then
    log "Config mode: environment prepared for claude-monitor installation"
    success "Configuration completed!"
    exit 0
fi

# Check installation method availability
if [[ "$INSTALL_METHOD" == "uv" ]] && ! command -v uv &> /dev/null; then
    warning "uv not available, falling back to pip"
    INSTALL_METHOD="pip"
fi

# Installation based on method
case $INSTALL_METHOD in
    uv)
        log "Installing claude-monitor via uv..."
        
        if [[ "$DRY_RUN" == true ]]; then
            log "[DRY RUN] Would run: uv tool install claude-monitor"
        else
            log "Installing claude-monitor globally via uv tool..."
            uv tool install claude-monitor || error "Failed to install claude-monitor via uv"
            
            # Ensure uv tools are in PATH
            UV_TOOLS_PATH="$HOME/.local/bin"
            if [[ ":$PATH:" != *":$UV_TOOLS_PATH:"* ]]; then
                export PATH="$UV_TOOLS_PATH:$PATH"
                log "Added $UV_TOOLS_PATH to PATH for this session"
            fi
        fi
        ;;
    
    pip)
        log "Installing claude-monitor via pip..."
        
        if [[ "$DRY_RUN" == true ]]; then
            log "[DRY RUN] Would run: pip install claude-monitor (with user fallback)"
        else
            log "Installing claude-monitor globally via pip..."
            # Try global install first, fallback to user install if permissions fail
            if ! pip install claude-monitor 2>/dev/null; then
                log "Global pip install failed, trying user-local installation..."
                pip install --user claude-monitor || error "Failed to install claude-monitor via pip"
                
                # Add ~/.local/bin to PATH if not already there
                USER_LOCAL_PATH="$HOME/.local/bin"
                if [[ ":$PATH:" != *":$USER_LOCAL_PATH:"* ]]; then
                    export PATH="$USER_LOCAL_PATH:$PATH"
                    log "Added $USER_LOCAL_PATH to PATH for this session"
                    
                    if [[ "$NO_SHELL" != true ]]; then
                        # Update shell profiles for persistent PATH
                        for profile in ~/.bashrc ~/.zshrc ~/.profile; do
                            if [[ -f "$profile" ]] && ! grep -q ".local/bin" "$profile"; then
                                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$profile"
                                log "Updated $profile to include ~/.local/bin in PATH"
                            fi
                        done
                    fi
                fi
            fi
        fi
        ;;
    
    source)
        log "Building claude-monitor from source..."
        
        if [[ "$DRY_RUN" == true ]]; then
            log "[DRY RUN] Would clone $CLAUDE_MONITOR_REPO"
            log "[DRY RUN] Would build Python package from source"
            log "[DRY RUN] Would install via uv tool or pip"
        else
            # Create temporary directory for source build
            BUILD_DIR=$(mktemp -d)
            cd "$BUILD_DIR"
            
            log "Cloning claude-monitor repository..."
            git clone "$CLAUDE_MONITOR_REPO" claude-monitor || error "Failed to clone claude-monitor repository"
            cd claude-monitor
            
            # Run tests if requested
            if [[ "$RUN_TESTS" == true ]]; then
                log "Running test suite..."
                if [[ -f "requirements-dev.txt" ]]; then
                    pip install --user -r requirements-dev.txt || warning "Failed to install dev dependencies"
                fi
                python -m pytest || warning "Tests failed but continuing with installation"
            fi
            
            # Install from source using preferred method
            log "Installing claude-monitor from source..."
            if command -v uv &> /dev/null; then
                log "Using uv for source installation..."
                uv tool install . || error "Failed to install claude-monitor from source via uv"
            else
                log "Using pip for source installation..."
                pip install --user . || error "Failed to install claude-monitor from source via pip"
                
                # Add ~/.local/bin to PATH if not already there
                USER_LOCAL_PATH="$HOME/.local/bin"
                if [[ ":$PATH:" != *":$USER_LOCAL_PATH:"* ]]; then
                    export PATH="$USER_LOCAL_PATH:$PATH"
                    log "Added $USER_LOCAL_PATH to PATH for this session"
                    
                    if [[ "$NO_SHELL" != true ]]; then
                        # Update shell profiles for persistent PATH
                        for profile in ~/.bashrc ~/.zshrc ~/.profile; do
                            if [[ -f "$profile" ]] && ! grep -q ".local/bin" "$profile"; then
                                echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$profile"
                                log "Updated $profile to include ~/.local/bin in PATH"
                            fi
                        done
                    fi
                fi
            fi
            
            # Cleanup
            cd - > /dev/null
            rm -rf "$BUILD_DIR"
            log "Source build cleanup completed"
        fi
        ;;
esac

# Build type specific setup
if [[ "$BUILD_TYPE" == "maximum" ]] && [[ "$DRY_RUN" != true ]]; then
    log "Maximum setup: Configuring additional features..."
    
    # Create configuration directory if it doesn't exist
    CLAUDE_MONITOR_CONFIG_DIR="$HOME/.config/claude-monitor"
    if [[ ! -d "$CLAUDE_MONITOR_CONFIG_DIR" ]]; then
        mkdir -p "$CLAUDE_MONITOR_CONFIG_DIR"
        log "Created claude-monitor configuration directory: $CLAUDE_MONITOR_CONFIG_DIR"
    fi
    
    # Create example configuration if none exists
    CONFIG_FILE="$CLAUDE_MONITOR_CONFIG_DIR/config.json"
    if [[ ! -f "$CONFIG_FILE" ]]; then
        cat > "$CONFIG_FILE" << 'EOF'
{
  "plan": "pro",
  "theme": "auto",
  "refresh_rate": 5,
  "timezone": "auto",
  "logging_level": "info",
  "display_mode": "realtime"
}
EOF
        log "Created example configuration file: $CONFIG_FILE"
    fi
    
    log "Maximum setup completed with configuration support"
fi

# Verify installation
INSTALLED=false
INSTALL_LOCATION=""

if [[ "$DRY_RUN" == true ]]; then
    # For dry run, simulate successful installation
    INSTALLED=true
    if [[ "$INSTALL_METHOD" == "uv" ]]; then
        INSTALL_LOCATION="$HOME/.local/bin/claude-monitor"
    else
        INSTALL_LOCATION="$HOME/.local/bin/claude-monitor"
    fi
elif command -v claude-monitor &> /dev/null; then
    INSTALLED=true
    INSTALL_LOCATION="$(which claude-monitor)"
fi

if [[ "$INSTALLED" == true ]]; then
    if [[ "$DRY_RUN" == true ]]; then
        INSTALLED_VERSION="(simulated)"
    else
        INSTALLED_VERSION=$(claude-monitor --version 2>/dev/null | head -n1 || echo "unknown")
    fi
    
    success "claude-monitor installation completed successfully!"
    
    if [[ "$QUIET" != true ]]; then
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: $INSTALL_METHOD"
        log "Binary location: $INSTALL_LOCATION"
        
        echo
        log "Basic usage examples:"
        log "  claude-monitor                      # Real-time token usage monitor"
        log "  claude-monitor --plan pro           # Monitor with Pro plan limits"
        log "  claude-monitor --plan max5          # Monitor with Claude 3.5 Sonnet limits"
        log "  claude-monitor --theme dark         # Use dark theme"
        log "  claude-monitor --refresh-rate 3     # Update every 3 seconds"
        log "  claude-monitor --help               # Show all available options"
        echo
        log "Key features:"
        log "  • Real-time token consumption tracking"
        log "  • Machine learning-based usage predictions"
        log "  • Rich, color-coded terminal UI"
        log "  • Multiple Claude subscription plan support"
        log "  • Intelligent auto-detection of usage patterns"
        log "  • P90 percentile usage analysis"
        log "  • Configurable views (realtime, daily, monthly)"
        echo
        log "Supported plans: pro, max5, max20, custom"
        log "Configuration directory: ~/.config/claude-monitor/"
        
        if [[ "$NO_SHELL" != true ]]; then
            echo
            warning "IMPORTANT: Restart your shell or run 'source ~/.bashrc' to use claude-monitor commands"
        fi
    fi
else
    error "claude-monitor installation verification failed"
fi

success "claude-monitor installation script completed!"