#!/bin/bash

# gopls Installation Script for Debian Linux
# Official Go language server installation
# Usage: ./install-gopls.sh [OPTIONS]

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
GOPLS_MODULE="golang.org/x/tools/gopls"
GO_MIN_VERSION="1.19.0"

# Default options - gopls uses go install, so simplified options
BUILD_TYPE="standard"  # minimal, standard, maximum (affects go build flags)
MODE="install"         # config, build, install
SKIP_DEPS=false
FORCE_INSTALL=false
DRY_RUN=false
VERBOSE=false
QUIET=false

# Show help
show_help() {
    cat << EOF
gopls Installation Script for Debian Linux

gopls (Go please) is the official language server for Go, providing IDE features
like code completion, navigation, diagnostics, and refactoring to LSP-compatible editors.

Usage: $0 [OPTIONS]

Build Types:
  -m, --minimal         Minimal build flags
  -s, --standard        Standard build (default)  
  -o, --maximum         Optimized build with additional flags

Modes:
  --config-only         Configure only (prepare environment)
  --build-only          Configure and build (no installation)
  --install             Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --force              Force reinstallation if already installed
  --dry-run            Show what would be done without executing
  --verbose            Enable verbose output
  --quiet              Suppress non-error output
  --help               Show this help message

Examples:
  $0                   # Standard installation
  $0 --maximum         # Optimized build
  $0 --force           # Force reinstall even if present
  $0 --dry-run         # Preview installation steps
  $0 --skip-deps       # Skip Go installation check

About gopls:
  • Official Go language server (LSP implementation)
  • Provides IDE features: completion, navigation, diagnostics, refactoring
  • Works with VS Code, Vim, Emacs, and other LSP-compatible editors
  • Supports Go modules and GOPATH modes
  • Developed and maintained by the Go team
  • Installation via: go install golang.org/x/tools/gopls@latest

Editor Integration:
  • VS Code: Go extension automatically manages gopls
  • Vim/Neovim: Configure with vim-lsp, coc.nvim, or built-in LSP
  • Emacs: Use lsp-mode with gopls configuration
  • Other editors: Configure LSP client to use 'gopls' command

EOF
}

# Parse command line arguments using Gearbox Standard Protocol v1.0
while [[ $# -gt 0 ]]; do
    case $1 in
        # Build types
        -m|--minimal)
            BUILD_TYPE="minimal"
            shift
            ;;
        -s|--standard)
            BUILD_TYPE="standard"
            shift
            ;;
        -o|--maximum)
            BUILD_TYPE="maximum"
            shift
            ;;
        # Execution modes
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
        # Common options
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --force)
            FORCE_INSTALL=true
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
        --help)
            show_help
            exit 0
            ;;
        # Ignore unsupported flags gracefully (Gearbox Standard Protocol)
        --run-tests|--no-shell|--version)
            if [[ "$VERBOSE" == true ]]; then
                warning "Ignoring unsupported flag: $1 (not applicable to gopls)"
            fi
            shift
            ;;
        *)
            warning "Unknown option: $1 (ignoring)"
            shift
            ;;
    esac
done

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
if [[ "$QUIET" != true ]]; then
    echo
    log "========================================"
    log "gopls Installation Script"
    log "========================================"
    log "Build type: $BUILD_TYPE"
    log "Mode: $MODE"
    log "Skip dependencies: $SKIP_DEPS"
    log "Force install: $FORCE_INSTALL"
    log "Dry run: $DRY_RUN"
    echo
fi

# Check if gopls is already installed
if command -v gopls &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(gopls version 2>/dev/null | head -n1 || echo "unknown")
    
    if [[ "$QUIET" != true ]]; then
        log "gopls is already installed: $CURRENT_VERSION"
        log "Use --force to reinstall with latest version"
    fi
    
    if [[ "$DRY_RUN" != true ]]; then
        exit 0
    fi
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    if [[ "$VERBOSE" == true ]] || [[ "$QUIET" != true ]]; then
        log "Checking Go installation..."
    fi
    
    if [[ "$DRY_RUN" == true ]]; then
        log "[DRY RUN] Would check and install Go if needed"
    else
        # Check if Go is installed
        if ! command -v go &> /dev/null; then
            log "Go is not installed. Installing Go..."
            
            # Update package list
            sudo apt update
            
            # Install Go using apt (will install latest available version)
            sudo apt install -y golang-go
            
            # Verify installation
            if ! command -v go &> /dev/null; then
                error "Go installation failed"
            fi
        fi
        
        # Check Go version
        GO_VERSION=$(go version | grep -oP 'go\d+\.\d+\.\d+' | sed 's/go//')
        if [[ "$VERBOSE" == true ]] || [[ "$QUIET" != true ]]; then
            log "Found Go version: $GO_VERSION"
        fi
        
        # Version check (simplified comparison)
        if [[ "$GO_VERSION" < "$GO_MIN_VERSION" ]]; then
            warning "Go version $GO_VERSION is below recommended minimum $GO_MIN_VERSION"
            log "gopls may work but consider upgrading Go for best compatibility"
        fi
    fi
    
    if [[ "$VERBOSE" == true ]] || [[ "$QUIET" != true ]]; then
        success "Go environment ready"
    fi
else
    if [[ "$VERBOSE" == true ]]; then
        log "Skipping dependency installation as requested"
    fi
fi

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    if [[ "$QUIET" != true ]]; then
        log "Ready to install gopls using: go install $GOPLS_MODULE@latest"
    fi
    exit 0
fi

# Build/Install gopls
if [[ "$VERBOSE" == true ]] || [[ "$QUIET" != true ]]; then
    log "Installing gopls using go install..."
fi

if [[ "$DRY_RUN" == true ]]; then
    log "[DRY RUN] Would run: go install $GOPLS_MODULE@latest"
    log "[DRY RUN] Build flags would be applied based on build type: $BUILD_TYPE"
    success "[DRY RUN] Installation preview completed"
    exit 0
fi

# Set build flags based on build type
case $BUILD_TYPE in
    minimal)
        # Minimal: fastest build, basic optimizations
        export GOFLAGS="-buildmode=default"
        if [[ "$VERBOSE" == true ]]; then
            log "Using minimal build flags for faster compilation"
        fi
        ;;
    standard)
        # Standard: balanced build
        export GOFLAGS="-trimpath"
        if [[ "$VERBOSE" == true ]]; then
            log "Using standard build flags"
        fi
        ;;
    maximum)
        # Maximum: optimized build with trimmed paths and stripping
        export GOFLAGS="-trimpath -ldflags=-s -ldflags=-w"
        if [[ "$VERBOSE" == true ]]; then
            log "Using maximum optimization build flags"
        fi
        ;;
esac

# Install gopls
if ! go install "$GOPLS_MODULE@latest"; then
    error "Failed to install gopls"
fi

# Ensure GOPATH/bin is in PATH
GOPATH_BIN="$(go env GOPATH)/bin"
if [[ ":$PATH:" != *":$GOPATH_BIN:"* ]]; then
    if [[ "$VERBOSE" == true ]] || [[ "$QUIET" != true ]]; then
        log "Adding Go binary directory to PATH..."
    fi
    
    # Add to current session
    export PATH="$GOPATH_BIN:$PATH"
    
    # Add to shell profile for persistence
    if [[ -f "$HOME/.bashrc" ]]; then
        if ! grep -q "export PATH.*$(go env GOPATH)/bin" "$HOME/.bashrc"; then
            echo "export PATH=\"$(go env GOPATH)/bin:\$PATH\"" >> "$HOME/.bashrc"
            if [[ "$VERBOSE" == true ]]; then
                log "Added Go binary path to ~/.bashrc"
            fi
        fi
    fi
    
    # Also try .profile as fallback
    if [[ -f "$HOME/.profile" ]]; then
        if ! grep -q "export PATH.*$(go env GOPATH)/bin" "$HOME/.profile"; then
            echo "export PATH=\"$(go env GOPATH)/bin:\$PATH\"" >> "$HOME/.profile"
        fi
    fi
fi

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    if [[ "$QUIET" != true ]]; then
        log "gopls built and ready at: $GOPATH_BIN/gopls"
    fi
    exit 0
fi

# Verify installation
if command -v gopls &> /dev/null; then
    INSTALLED_VERSION=$(gopls version 2>/dev/null | head -n1)
    
    if [[ "$QUIET" != true ]]; then
        success "gopls installation completed successfully!"
        log "Installed version: $INSTALLED_VERSION"
        log "Installation method: go install (Go modules)"
        log "Binary location: $GOPATH_BIN/gopls"
        log "Build type: $BUILD_TYPE"
        
        echo
        log "Editor Integration:"
        log "• VS Code: Install the Go extension (gopls auto-configured)"
        log "• Vim/Neovim: Configure LSP client to use 'gopls' command"
        log "• Emacs: Use lsp-mode with (lsp-register-client gopls)"
        log "• Other editors: Point LSP client to 'gopls' binary"
        echo
        log "Basic verification:"
        log "  gopls version              # Show version information"
        log "  gopls help                 # Show available commands"
        log "  gopls check /path/to/file  # Check Go file syntax"
        echo
        log "LSP Features:"
        log "• Code completion and IntelliSense"
        log "• Go to definition/references/implementation"
        log "• Symbol search and workspace navigation"  
        log "• Real-time error detection and diagnostics"
        log "• Code actions (organize imports, extract functions)"
        log "• Refactoring support (rename symbols, etc.)"
        log "• Hover information and documentation"
        echo
        log "Configuration:"
        log "• gopls uses Go toolchain configuration (go env)"
        log "• Project-specific settings via .vscode/settings.json (VS Code)"
        log "• Editor-specific LSP client configuration"
        log "• Documentation: https://pkg.go.dev/golang.org/x/tools/gopls"
    fi
else
    error "gopls installation verification failed - gopls not found in PATH"
fi

if [[ "$QUIET" != true ]]; then
    success "gopls installation completed!"
    echo
    log "Next steps:"
    log "1. Configure your editor to use gopls as the Go language server"
    log "2. Open a Go project to test code completion and navigation"
    log "3. Check editor-specific gopls setup documentation"
    echo
    log "You may need to restart your shell or run 'source ~/.bashrc'"
    log "to ensure PATH changes are active in new terminal sessions."
fi