#!/bin/bash

# fzf Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-fzf.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
FZF_DIR="fzf"
FZF_REPO="https://github.com/junegunn/fzf.git"
GO_MIN_VERSION="1.20"

# Default options
BUILD_TYPE="standard"  # standard, profiling
MODE="install"         # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false
SETUP_SHELL=true

# Show help
show_help() {
    cat << EOF
fzf Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -s, --standard        Standard build (default, optimized)
  -p, --profiling       Profiling build (with pprof support)

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --force              Force reinstallation if already installed
  --no-shell           Skip shell integration setup
  -h, --help           Show this help message

Examples:
  $0                   # Default: standard build with shell integration
  $0 -p -c             # Profiling build, config only
  $0 --no-shell        # Standard build without shell integration
  $0 --run-tests       # Standard build with tests

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--standard)
            BUILD_TYPE="standard"
            shift
            ;;
        -p|--profiling)
            BUILD_TYPE="profiling"
            shift
            ;;
        -c|--config-only)
            MODE="config"
            shift
            ;;
        -b|--build-only)
            MODE="build"
            shift
            ;;
        -i|--install)
            MODE="install"
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
            SETUP_SHELL=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

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

# Check and install Go
install_go() {
    if command -v go &> /dev/null; then
        local current_version=$(go version | sed 's/go version go\([0-9.]*\).*/\1/')
        log "Found Go version: $current_version"
        
        if version_compare $current_version $GO_MIN_VERSION; then
            log "Go version is sufficient (>= $GO_MIN_VERSION)"
            return 0
        else
            warning "Go version $current_version is below minimum required $GO_MIN_VERSION"
        fi
    else
        log "Go not found, installing..."
    fi
    
    # Install Go from official releases
    log "Installing Go $GO_MIN_VERSION..."
    local GO_VERSION="1.23.4"  # Latest stable
    local GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    local GO_URL="https://golang.org/dl/${GO_TARBALL}"
    
    # Download and install Go
    cd /tmp
    wget -q "$GO_URL" || error "Failed to download Go"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$GO_TARBALL" || error "Failed to extract Go"
    rm "$GO_TARBALL"
    
    # Add Go to PATH
    if [[ ":$PATH:" != *":/usr/local/go/bin:"* ]]; then
        export PATH="/usr/local/go/bin:$PATH"
        echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
    fi
    
    success "Go installed successfully"
}

# Get make command based on build type
get_make_command() {
    case $BUILD_TYPE in
        standard)
            echo "make install"
            ;;
        profiling)
            echo "TAGS=pprof make clean install"
            ;;
    esac
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
        wget \
        curl \
        make \
        || error "Failed to install build tools"

    # Install Go
    install_go

    # Install Ruby for integration tests (if requested)
    if [[ "$RUN_TESTS" == true ]]; then
        log "Installing Ruby for integration tests..."
        sudo apt install -y \
            ruby \
            ruby-dev \
            bundler \
            tmux \
            zsh \
            fish \
            || warning "Some test dependencies may not be available"
    fi

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
                if ! grep -q "fzf --bash" ~/.bashrc 2>/dev/null; then
                    echo '# fzf integration' >> ~/.bashrc
                    echo 'eval "$(fzf --bash)"' >> ~/.bashrc
                    log "Added fzf integration to ~/.bashrc"
                else
                    log "fzf integration already exists in ~/.bashrc"
                fi
                ;;
            zsh)
                log "Setting up zsh integration..."
                if [[ -f ~/.zshrc ]]; then
                    if ! grep -q "fzf --zsh" ~/.zshrc 2>/dev/null; then
                        echo '# fzf integration' >> ~/.zshrc
                        echo 'source <(fzf --zsh)' >> ~/.zshrc
                        log "Added fzf integration to ~/.zshrc"
                    else
                        log "fzf integration already exists in ~/.zshrc"
                    fi
                fi
                ;;
            fish)
                log "Setting up fish integration..."
                local fish_config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/fish"
                if [[ -d "$fish_config_dir" ]]; then
                    local fish_config="$fish_config_dir/config.fish"
                    if ! grep -q "fzf --fish" "$fish_config" 2>/dev/null; then
                        echo '# fzf integration' >> "$fish_config"
                        echo 'fzf --fish | source' >> "$fish_config"
                        log "Added fzf integration to $fish_config"
                    else
                        log "fzf integration already exists in $fish_config"
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

log "Starting fzf $MODE process for Debian Linux"
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Shell integration: $SETUP_SHELL"

# Handle fzf source code
if [[ -d "$FZF_DIR" ]]; then
    log "Found existing fzf directory: $FZF_DIR"
    
    # Check if it's a git repository
    if [[ -d "$FZF_DIR/.git" ]]; then
        log "Existing directory is a git repository"
        
        # Check if it's the correct repository
        cd "$FZF_DIR"
        CURRENT_ORIGIN=$(git remote get-url origin 2>/dev/null || echo "")
        
        if [[ "$CURRENT_ORIGIN" == "$FZF_REPO" ]]; then
            log "Repository origin matches expected fzf repository"
            
            if ask_user "Do you want to pull the latest changes from the fzf repository?"; then
                log "Pulling latest changes..."
                git pull origin master || error "Failed to pull latest changes"
                success "Repository updated successfully"
            else
                log "Using existing code without updates"
            fi
        else
            warning "Existing git repository has different origin: $CURRENT_ORIGIN"
            if ask_user "Do you want to remove this directory and clone fresh fzf repository?"; then
                cd ..
                rm -rf "$FZF_DIR"
                log "Cloning fzf repository..."
                git clone "$FZF_REPO" "$FZF_DIR" || error "Failed to clone fzf repository"
                cd "$FZF_DIR"
                success "fzf repository cloned successfully"
            else
                log "Continuing with existing repository"
            fi
        fi
        cd ..
    else
        warning "Directory exists but is not a git repository"
        if ask_user "Do you want to remove this directory and clone fresh fzf repository?"; then
            rm -rf "$FZF_DIR"
            log "Cloning fzf repository..."
            git clone "$FZF_REPO" "$FZF_DIR" || error "Failed to clone fzf repository"
            success "fzf repository cloned successfully"
        else
            error "Cannot proceed without a proper fzf source directory"
        fi
    fi
else
    log "Cloning fzf repository..."
    git clone "$FZF_REPO" "$FZF_DIR" || error "Failed to clone fzf repository"
    success "fzf repository cloned successfully"
fi

# Change to fzf directory
cd "$FZF_DIR"

# Verify we're in the correct directory
if [[ ! -f "Makefile" ]]; then
    error "Invalid fzf source directory - missing Makefile"
fi

# Install dependencies
install_dependencies

# Ensure Go is in PATH
if ! command -v go &> /dev/null; then
    export PATH="/usr/local/go/bin:$PATH"
    if ! command -v go &> /dev/null; then
        error "Failed to find Go in PATH"
    fi
fi

# Clean previous build
log "Cleaning previous build files..."
make clean || warning "Failed to clean previous build, continuing..."

# Get build configuration
MAKE_COMMAND=$(get_make_command)

log "Configuring fzf with $BUILD_TYPE settings..."
log "Make command: $MAKE_COMMAND"

success "Configuration completed successfully"

# Exit if config-only mode
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed. Run with --build-only or --install to continue."
    success "Build command would be: $MAKE_COMMAND"
    exit 0
fi

# Build fzf
log "Building fzf (this may take a while)..."

eval "$MAKE_COMMAND" || error "Build failed"

# Verify build output
if [[ ! -f "bin/fzf" ]]; then
    error "Build completed but fzf executable not found in bin/"
fi

success "Build completed successfully"

# Run tests if requested
if [[ "$RUN_TESTS" == true ]]; then
    log "Running test suite..."
    make test || warning "Some unit tests failed, but continuing"
    if command -v ruby &> /dev/null; then
        make itest || warning "Some integration tests failed, but continuing"
    else
        warning "Ruby not available, skipping integration tests"
    fi
    success "Test suite completed"
fi

# Exit if build-only mode
if [[ "$MODE" == "build" ]]; then
    success "Build completed. Run with --install to install."
    log "Build output: $(pwd)/bin/fzf"
    exit 0
fi

# Install fzf
log "Installing fzf..."

# Install the binary to system location
sudo cp bin/fzf /usr/local/bin/ || error "Installation failed"
sudo chmod +x /usr/local/bin/fzf || error "Failed to set executable permissions"

# Setup shell integrations
setup_shell_integration

# Verify installation
log "Verifying installation..."
# Force PATH update for verification
export PATH="/usr/local/bin:$PATH"
# Clear bash command hash table to ensure new binary is used
hash -r
if command -v fzf &> /dev/null; then
    success "fzf installation verified!"
    echo
    log "fzf version information:"
    fzf --version
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        standard)
            log "Features: Standard optimized build"
            ;;
        profiling)
            log "Features: Profiling build with pprof support"
            ;;
    esac
    echo
    success "fzf installation completed successfully!"
    log "You can now use the 'fzf' command for fuzzy finding"
    echo
    log "Usage examples:"
    log "  fzf                          # Interactive fuzzy finder"
    log "  find . -type f | fzf         # Find files"
    log "  history | fzf                # Search command history"
    log "  ps aux | fzf                 # Search processes"
    echo
    if [[ "$SETUP_SHELL" == true ]]; then
        log "Shell integrations installed:"
        log "  CTRL-T: File selection"
        log "  CTRL-R: Command history search"
        log "  ALT-C: Directory navigation"
        echo
        log "To activate shell integrations immediately, run:"
        log "  hash -r && source ~/.bashrc"
        log "Or simply open a new terminal window"
    fi
    echo
    log "Installation paths:"
    log "  fzf: $(which fzf)"
    echo
    log "Script completed in directory: $(pwd)"
else
    error "fzf installation verification failed - try restarting your shell or run 'source ~/.bashrc'"
fi