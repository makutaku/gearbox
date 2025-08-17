#!/bin/bash

# gh Installation Script for Debian Linux
# Usage: ./install-gh.sh [OPTIONS]

set -e

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'

# Configuration
GH_DIR="cli"; GH_REPO="https://github.com/cli/cli.git"; GO_MIN_VERSION="1.21.0"

# Defaults
BUILD_TYPE="release"; MODE="install"; INSTALL_METHOD="source"
SKIP_DEPS=false; RUN_TESTS=false; FORCE_INSTALL=false; SETUP_COMPLETIONS=true

show_help() {
    cat << EOF
gh Installation Script for Debian Linux

Build Types: -d/--debug, -r/--release (default), -o/--optimized
Modes: -c/--config-only, -b/--build-only, -i/--install (default)
Methods: --source (Go build, default), --binary, --package (apt)
Options: --skip-deps, --run-tests, --no-completions, --force, -h/--help

About gh: GitHub CLI for repository management, PRs, issues, and workflows.
Essential for modern GitHub-based development workflows.
EOF
}

# Parse args
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--debug) BUILD_TYPE="debug"; shift ;;
        -r|--release) BUILD_TYPE="release"; shift ;;
        -o|--optimized) BUILD_TYPE="optimized"; shift ;;
        -c|--config-only) MODE="config"; shift ;;
        -b|--build-only) MODE="build"; shift ;;
        -i|--install) MODE="install"; shift ;;
        --source) INSTALL_METHOD="source"; shift ;;
        --binary) INSTALL_METHOD="binary"; shift ;;
        --package) INSTALL_METHOD="package"; shift ;;
        --skip-deps) SKIP_DEPS=true; shift ;;
        --run-tests) RUN_TESTS=true; shift ;;
        --no-completions) SETUP_COMPLETIONS=false; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        -h|--help) show_help; exit 0 ;;
        *) echo "Unknown option: $1"; show_help; exit 1 ;;
    esac
done

# Functions
log() { echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

version_compare() {
    if [[ $1 == $2 ]]; then return 0; fi
    local IFS=.; local i ver1=($1) ver2=($2)
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do ver1[i]=0; done
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then ver2[i]=0; fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then return 0; fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then return 1; fi
    done; return 0
}

setup_completions() {
    [[ "$SETUP_COMPLETIONS" != true ]] || ! command -v gh &> /dev/null && return
    log "Setting up shell completions..."
    mkdir -p "$HOME/.local/share/bash-completion/completions" "$HOME/.local/share/zsh/site-functions" "$HOME/.config/fish/completions"
    gh completion -s bash > "$HOME/.local/share/bash-completion/completions/gh" 2>/dev/null || true
    gh completion -s zsh > "$HOME/.local/share/zsh/site-functions/_gh" 2>/dev/null || true  
    gh completion -s fish > "$HOME/.config/fish/completions/gh.fish" 2>/dev/null || true
    success "Shell completions installed!"
}

[[ $EUID -eq 0 ]] && error "Don't run as root"

echo; log "========================================"; log "gh Installation Script"
log "Build: $BUILD_TYPE, Mode: $MODE, Method: $INSTALL_METHOD"; echo

# Check existing
if command -v gh &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(gh --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' | head -1 || echo "unknown")
    if [[ "$INSTALL_METHOD" != "source" ]]; then
        log "gh already installed (version: $CURRENT_VERSION). Use --force to reinstall"; exit 0
    else
        log "Found gh (version: $CURRENT_VERSION). Building latest from source"
    fi
fi

# Package method
if [[ "$INSTALL_METHOD" == "package" ]]; then
    log "Using apt package installation..."
    [[ "$MODE" == "config" ]] && { log "Config: would install via apt"; success "Config done!"; exit 0; }
    
    # Add GitHub CLI repository
    log "Adding GitHub CLI repository..."
    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
    sudo chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
    
    sudo apt update; sudo apt install -y gh
    
    if command -v gh &> /dev/null; then
        success "gh installed via apt!"; log "Version: $(gh --version | head -1)"
        setup_completions; echo; log "Usage: gh repo clone owner/repo, gh pr create, gh auth login"
    else
        error "gh installation failed"
    fi
    exit 0
fi

# Binary method
if [[ "$INSTALL_METHOD" == "binary" ]]; then
    log "Using binary installation..."
    [[ "$MODE" == "config" ]] && { log "Config: would download gh binary"; success "Config done!"; exit 0; }
    
    ARCH=$(uname -m); case $ARCH in
        x86_64) ARCH_TAG="linux_amd64" ;;
        aarch64|arm64) ARCH_TAG="linux_arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac
    
    log "Downloading latest gh for $ARCH_TAG..."
    VERSION=$(curl -s "https://api.github.com/repos/cli/cli/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)
    [[ -z "$VERSION" ]] && error "Could not get version"
    
    DOWNLOAD_URL="https://github.com/cli/cli/releases/download/${VERSION}/gh_${VERSION#v}_${ARCH_TAG}.tar.gz"
    log "Downloading: $DOWNLOAD_URL"
    
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" || error "Download failed"
    tar -xzf "$(basename "$DOWNLOAD_URL")" || error "Extract failed"
    
    GH_BIN=$(find . -name "gh" -type f -executable | head -1)
    [[ -z "$GH_BIN" ]] && error "gh binary not found"
    
    sudo cp "$GH_BIN" /usr/local/bin/gh; sudo chmod +x /usr/local/bin/gh
    cd /; rm -rf "$TEMP_DIR"
    
    if command -v gh &> /dev/null; then
        success "gh installed!"; log "Version: $(gh --version | head -1)"; log "Location: /usr/local/bin/gh"
        setup_completions; echo; log "Usage: gh auth login, gh repo clone owner/repo, gh pr create"
    else
        error "Installation verification failed"
    fi
    exit 0
fi

# Source method
log "Using source build..."

if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies..."
    sudo apt update; sudo apt install -y build-essential git curl
    
    if ! command -v go &> /dev/null; then
        log "Installing Go..."
        GO_VERSION="1.22.1"
        curl -fL "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "/tmp/go.tar.gz"
        sudo rm -rf /usr/local/go; sudo tar -C /usr/local -xzf "/tmp/go.tar.gz"
        export PATH="/usr/local/go/bin:$PATH"; echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
        rm "/tmp/go.tar.gz"
    else
        GO_VERSION=$(go version | grep -oP 'go\d+\.\d+\.\d+' | sed 's/go//')
        log "Found Go: $GO_VERSION"
        if ! version_compare "$GO_VERSION" "$GO_MIN_VERSION"; then
            warning "Go $GO_VERSION < $GO_MIN_VERSION. Consider upgrading"
        fi
    fi
    success "Dependencies done!"
fi

[[ ! -d /usr/local/go/bin ]] || export PATH="/usr/local/go/bin:$PATH"
! command -v go &> /dev/null && error "Go not available"

BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"; mkdir -p "$BUILD_DIR"; cd "$BUILD_DIR"
GH_SOURCE_DIR="$BUILD_DIR/$GH_DIR"

if [[ ! -d "$GH_SOURCE_DIR" ]]; then
    log "Cloning gh repository..."; git clone "$GH_REPO" "$GH_SOURCE_DIR"
else
    log "Updating gh repository..."; cd "$GH_SOURCE_DIR"; git fetch origin; git reset --hard origin/trunk
fi

cd "$GH_SOURCE_DIR"
[[ ! -f "go.mod" ]] && error "go.mod not found"
log "gh source configured"
[[ "$MODE" == "config" ]] && { success "Config done!"; exit 0; }

log "Building gh..."
mkdir -p bin
case $BUILD_TYPE in
    debug) 
        go build -o bin/gh ./cmd/gh ;;
    release) 
        go build -ldflags "-s -w" -o bin/gh ./cmd/gh ;;
    optimized) 
        CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/gh ./cmd/gh ;;
esac

GH_BINARY="bin/gh"
[[ ! -f "$GH_BINARY" ]] && error "Build failed - gh binary not found at $GH_BINARY"
success "gh build completed!"
[[ "$MODE" == "build" ]] && { success "Build done!"; log "Binary: $PWD/$GH_BINARY"; exit 0; }

[[ "$RUN_TESTS" == true ]] && { log "Running tests..."; go test ./... || warning "Some tests failed"; }

log "Installing gh..."
sudo cp "$GH_BINARY" /usr/local/bin/; sudo chmod +x /usr/local/bin/gh

if command -v gh &> /dev/null; then
    success "gh installation completed!"; log "Version: $(gh --version | head -1)"
    log "Method: Source build ($BUILD_TYPE)"; log "Location: /usr/local/bin/gh"
    setup_completions
    
    echo; log "Usage:"; log "  gh auth login                # Authenticate with GitHub"
    log "  gh repo clone owner/repo     # Clone repository"; log "  gh pr create                 # Create pull request"
    log "  gh issue list                # List issues"; log "  gh release create            # Create release"
    echo; log "Key features:"; log "  ✓ Pull request management"; log "  ✓ Issue tracking"
    log "  ✓ Repository operations"; log "  ✓ GitHub Actions workflow management"
    log "  ✓ Release and gist management"; echo; log "Setup: Run 'gh auth login' first"
else
    error "Installation verification failed"
fi

sudo ldconfig; success "gh installation completed!"