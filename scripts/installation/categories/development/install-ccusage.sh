#!/bin/bash
# ccusage Installation Script for Debian Linux
# Automated dependency installation, build, and install
# Usage: ./install-ccusage.sh [OPTIONS]

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
CCUSAGE_REPO="https://github.com/ryoppippi/ccusage.git"
CCUSAGE_MIN_VERSION="1.0.0"

# Default options
BUILD_TYPE="standard"     # minimal, standard, maximum
MODE="install"            # config, build, install
INSTALL_METHOD="npm"      # npm (recommended), bun, source
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
ccusage Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  --minimal            Minimal ccusage installation (npm install only)
  --standard           Standard ccusage installation (default, global npm install)
  --maximum            Maximum ccusage installation (source build with development tools)

Modes:
  --config-only        Configure only (prepare build environment)
  --build-only         Configure and build (no installation)
  --install            Configure, build, and install (default)

Installation Methods:
  --npm                Use npm global install (default, recommended)
  --bun                Use bun global install (if Bun is available)
  --source             Build from source (TypeScript compilation)

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
  $0                   # Default: npm global installation (recommended)
  $0 --bun             # Use Bun for installation (faster)
  $0 --source          # Build from source (for development)
  $0 --minimal --npm   # Minimal npm installation
  $0 --maximum --source# Full source build with dev tools
  $0 --dry-run         # Preview installation steps

About ccusage:
  A CLI tool for analyzing Claude Code usage from local JSONL files.
  Provides incredibly fast and informative token usage analysis, cost
  tracking, and live monitoring dashboards for Claude Code projects.

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
        --npm)
            INSTALL_METHOD="npm"
            shift
            ;;
        --bun)
            INSTALL_METHOD="bun"
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
            echo "ccusage Installation Script v1.0"
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
    log "ccusage Installation Script"
    log "==================================="
    log "Build type: $BUILD_TYPE"
    log "Mode: $MODE"
    log "Install method: $INSTALL_METHOD"
    log "Skip dependencies: $SKIP_DEPS"
    log "Run tests: $RUN_TESTS"
    log "Dry run: $DRY_RUN"
    echo
fi

# Check if ccusage is already installed
if command -v ccusage &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(ccusage --version 2>/dev/null | head -n1 || echo "unknown")
    
    log "ccusage is already installed (version: $CURRENT_VERSION)"
    log "Use --force to reinstall"
    exit 0
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies..."
    
    if [[ "$DRY_RUN" == true ]]; then
        log "[DRY RUN] Would install: nodejs npm"
        if [[ "$INSTALL_METHOD" == "bun" ]]; then
            log "[DRY RUN] Would ensure Bun is available for installation"
        fi
        if [[ "$INSTALL_METHOD" == "source" ]]; then
            log "[DRY RUN] Would install: git build-essential"
        fi
    else
        # Update package list
        sudo apt-get update -qq
        
        # Install required packages
        sudo apt-get install -y nodejs npm
        
        # Check for Bun if using bun method
        if [[ "$INSTALL_METHOD" == "bun" ]] && ! command -v bun &> /dev/null; then
            warning "Bun not found but --bun specified. Falling back to npm installation."
            INSTALL_METHOD="npm"
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
    log "Config mode: environment prepared for ccusage installation"
    success "Configuration completed!"
    exit 0
fi

# Check installation method availability
if [[ "$INSTALL_METHOD" == "bun" ]] && ! command -v bun &> /dev/null; then
    warning "Bun not available, falling back to npm"
    INSTALL_METHOD="npm"
fi

# Installation based on method
case $INSTALL_METHOD in
    npm)
        log "Installing ccusage via npm..."
        
        if [[ "$DRY_RUN" == true ]]; then
            case $BUILD_TYPE in
                minimal)
                    log "[DRY RUN] Would run: npx ccusage@latest --version (no installation)"
                    ;;
                standard|maximum)
                    log "[DRY RUN] Would run: npm install -g ccusage"
                    ;;
            esac
        else
            case $BUILD_TYPE in
                minimal)
                    log "Minimal installation: Using npx without global install"
                    # Test that npx works
                    npx ccusage@latest --version || error "Failed to test ccusage via npx"
                    log "ccusage can be run via 'npx ccusage@latest'"
                    ;;
                standard|maximum)
                    log "Installing ccusage globally via npm..."
                    # Try global install first, fallback to user-local if permissions fail
                    if ! npm install -g ccusage 2>/dev/null; then
                        log "Global npm install failed, trying user-local installation..."
                        # Configure npm to use user-local prefix
                        mkdir -p ~/.npm-global
                        npm config set prefix '~/.npm-global'
                        npm install -g ccusage || error "Failed to install ccusage via npm"
                        
                        # Add ~/.npm-global/bin to PATH if not already there
                        NPM_GLOBAL_PATH="$HOME/.npm-global/bin"
                        if [[ ":$PATH:" != *":$NPM_GLOBAL_PATH:"* ]]; then
                            export PATH="$NPM_GLOBAL_PATH:$PATH"
                            log "Added $NPM_GLOBAL_PATH to PATH for this session"
                            
                            if [[ "$NO_SHELL" != true ]]; then
                                # Update shell profiles for persistent PATH
                                for profile in ~/.bashrc ~/.zshrc ~/.profile; do
                                    if [[ -f "$profile" ]] && ! grep -q ".npm-global/bin" "$profile"; then
                                        echo 'export PATH="$HOME/.npm-global/bin:$PATH"' >> "$profile"
                                        log "Updated $profile to include ~/.npm-global/bin in PATH"
                                    fi
                                done
                            fi
                        fi
                    fi
                    ;;
            esac
        fi
        ;;
    
    bun)
        log "Installing ccusage via Bun..."
        
        if [[ "$DRY_RUN" == true ]]; then
            case $BUILD_TYPE in
                minimal)
                    log "[DRY RUN] Would run: bunx ccusage --version (no installation)"
                    ;;
                standard|maximum)
                    log "[DRY RUN] Would run: bun install -g ccusage"
                    ;;
            esac
        else
            case $BUILD_TYPE in
                minimal)
                    log "Minimal installation: Using bunx without global install"
                    # Test that bunx works
                    bunx ccusage --version || error "Failed to test ccusage via bunx"
                    log "ccusage can be run via 'bunx ccusage'"
                    ;;
                standard|maximum)
                    log "Installing ccusage globally via Bun..."
                    # Try global install first, fallback if permissions fail
                    if ! bun install -g ccusage 2>/dev/null; then
                        log "Global bun install failed, trying alternative installation..."
                        # Fallback to npm user-local installation
                        mkdir -p ~/.npm-global
                        npm config set prefix '~/.npm-global'
                        npm install -g ccusage || error "Failed to install ccusage via fallback npm"
                        
                        # Add ~/.npm-global/bin to PATH if not already there
                        NPM_GLOBAL_PATH="$HOME/.npm-global/bin"
                        if [[ ":$PATH:" != *":$NPM_GLOBAL_PATH:"* ]]; then
                            export PATH="$NPM_GLOBAL_PATH:$PATH"
                            log "Added $NPM_GLOBAL_PATH to PATH for this session"
                            
                            if [[ "$NO_SHELL" != true ]]; then
                                # Update shell profiles for persistent PATH
                                for profile in ~/.bashrc ~/.zshrc ~/.profile; do
                                    if [[ -f "$profile" ]] && ! grep -q ".npm-global/bin" "$profile"; then
                                        echo 'export PATH="$HOME/.npm-global/bin:$PATH"' >> "$profile"
                                        log "Updated $profile to include ~/.npm-global/bin in PATH"
                                    fi
                                done
                            fi
                        fi
                    fi
                    ;;
            esac
        fi
        ;;
    
    source)
        log "Building ccusage from source..."
        
        if [[ "$DRY_RUN" == true ]]; then
            log "[DRY RUN] Would clone $CCUSAGE_REPO"
            log "[DRY RUN] Would build TypeScript source"
            log "[DRY RUN] Would install globally from source"
        else
            # Create temporary directory for source build
            BUILD_DIR=$(mktemp -d)
            cd "$BUILD_DIR"
            
            log "Cloning ccusage repository..."
            git clone "$CCUSAGE_REPO" ccusage || error "Failed to clone ccusage repository"
            cd ccusage
            
            # Install dependencies
            if command -v bun &> /dev/null; then
                log "Using Bun for dependency installation..."
                bun install || error "Failed to install dependencies with bun"
            else
                log "Using npm for dependency installation..."
                npm install || error "Failed to install dependencies with npm"
            fi
            
            # Run tests if requested
            if [[ "$RUN_TESTS" == true ]]; then
                log "Running test suite..."
                if command -v bun &> /dev/null; then
                    bun run test || warning "Tests failed but continuing with installation"
                else
                    npm run test || warning "Tests failed but continuing with installation"
                fi
            fi
            
            # Build the project
            log "Building ccusage from source..."
            if command -v bun &> /dev/null; then
                bun run build || error "Failed to build ccusage"
            else
                npm run build || error "Failed to build ccusage"
            fi
            
            # Install globally
            log "Installing ccusage globally from source..."
            npm install -g . || error "Failed to install ccusage globally"
            
            # Cleanup
            cd - > /dev/null
            rm -rf "$BUILD_DIR"
            log "Source build cleanup completed"
        fi
        ;;
esac

# Build type specific setup for maximum installation
if [[ "$BUILD_TYPE" == "maximum" ]] && [[ "$DRY_RUN" != true ]]; then
    log "Maximum setup: Configuring additional features..."
    
    # Create configuration directory if it doesn't exist
    CCUSAGE_CONFIG_DIR="$HOME/.ccusage"
    if [[ ! -d "$CCUSAGE_CONFIG_DIR" ]]; then
        mkdir -p "$CCUSAGE_CONFIG_DIR"
        log "Created ccusage configuration directory: $CCUSAGE_CONFIG_DIR"
    fi
    
    # Create example configuration if none exists
    CONFIG_FILE="$CCUSAGE_CONFIG_DIR/config.json"
    if [[ ! -f "$CONFIG_FILE" ]]; then
        cat > "$CONFIG_FILE" << 'EOF'
{
  "timezone": "UTC",
  "locale": "en-US",
  "defaultCommand": "daily",
  "outputFormat": "table"
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
    case $BUILD_TYPE in
        minimal)
            if [[ "$INSTALL_METHOD" == "npm" ]]; then
                INSTALL_LOCATION="npx ccusage@latest"
            elif [[ "$INSTALL_METHOD" == "bun" ]]; then
                INSTALL_LOCATION="bunx ccusage"
            fi
            ;;
        *)
            INSTALL_LOCATION="/usr/local/bin/ccusage"
            ;;
    esac
elif command -v ccusage &> /dev/null; then
    INSTALLED=true
    INSTALL_LOCATION="$(which ccusage)"
elif [[ "$BUILD_TYPE" == "minimal" ]]; then
    if [[ "$INSTALL_METHOD" == "npm" ]] && command -v npx &> /dev/null; then
        if npx ccusage@latest --version &> /dev/null; then
            INSTALLED=true
            INSTALL_LOCATION="npx ccusage@latest"
        fi
    elif [[ "$INSTALL_METHOD" == "bun" ]] && command -v bunx &> /dev/null; then
        if bunx ccusage --version &> /dev/null; then
            INSTALLED=true
            INSTALL_LOCATION="bunx ccusage"
        fi
    fi
fi

if [[ "$INSTALLED" == true ]]; then
    if [[ "$BUILD_TYPE" == "minimal" ]] && [[ "$INSTALL_LOCATION" == *"npx"* || "$INSTALL_LOCATION" == *"bunx"* ]]; then
        INSTALLED_VERSION="via $INSTALL_LOCATION"
    else
        INSTALLED_VERSION=$(ccusage --version 2>/dev/null | head -n1 || echo "unknown")
    fi
    
    success "ccusage installation completed successfully!"
    
    if [[ "$QUIET" != true ]]; then
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: $INSTALL_METHOD"
        log "Binary/command: $INSTALL_LOCATION"
        
        echo
        log "Basic usage examples:"
        if [[ "$BUILD_TYPE" == "minimal" ]]; then
            if [[ "$INSTALL_METHOD" == "npm" ]]; then
                log "  npx ccusage@latest              # Daily token usage report"
                log "  npx ccusage@latest daily        # Daily usage (explicit)"
                log "  npx ccusage@latest monthly      # Monthly aggregated report"
                log "  npx ccusage@latest blocks --live# Real-time monitoring"
            elif [[ "$INSTALL_METHOD" == "bun" ]]; then
                log "  bunx ccusage                    # Daily token usage report"
                log "  bunx ccusage daily              # Daily usage (explicit)"
                log "  bunx ccusage monthly            # Monthly aggregated report"
                log "  bunx ccusage blocks --live      # Real-time monitoring"
            fi
        else
            log "  ccusage                         # Daily token usage report"
            log "  ccusage daily                   # Daily usage (explicit)"
            log "  ccusage monthly                 # Monthly aggregated report"
            log "  ccusage blocks --live           # Real-time monitoring dashboard"
            log "  ccusage --json                  # JSON output format"
            log "  ccusage --help                  # Show all available options"
        fi
        echo
        log "Features:"
        log "  • Daily and monthly token usage analysis"
        log "  • Session and 5-hour block tracking"
        log "  • Live monitoring dashboard"
        log "  • Model-specific cost breakdown"
        log "  • Date range filtering and timezone support"
        log "  • Compact display modes and JSON output"
        echo
        log "Note: ccusage analyzes Claude Code usage from local JSONL files"
        log "Make sure your Claude Code project has logging enabled for analysis"
    fi
else
    error "ccusage installation verification failed"
fi

success "ccusage installation script completed!"