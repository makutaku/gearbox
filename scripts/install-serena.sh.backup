#!/bin/bash

# Serena Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-serena.sh [OPTIONS]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERENA_DIR="serena"
SERENA_REPO="https://github.com/oraios/serena.git"
PYTHON_MIN_VERSION="3.11.0"
PYTHON_MAX_VERSION="3.12.0"

# Default options
BUILD_TYPE="standard"    # minimal, standard, full
MODE="install"           # config, build, install
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help
show_help() {
    cat << EOF
Serena Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -m, --minimal         Minimal installation (core features only)
  -s, --standard        Standard installation (default, recommended features)
  -f, --full            Full installation (all optional dependencies)

Modes:
  -c, --config-only     Configure only (prepare build environment)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Default: standard build with install
  $0 -m -c             # Minimal build, config only
  $0 -f --run-tests    # Full build with tests
  $0 --skip-deps       # Skip dependency installation

About Serena:
  Powerful coding agent toolkit providing semantic retrieval and editing
  capabilities. Turns an LLM into a fully-featured agent working directly
  on codebases with IDE-like semantic analysis across programming languages.

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--minimal)
            BUILD_TYPE="minimal"
            shift
            ;;
        -s|--standard)
            BUILD_TYPE="standard"
            shift
            ;;
        -f|--full)
            BUILD_TYPE="full"
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

# Check if version is in range
version_in_range() {
    local version=$1
    local min_version=$2
    local max_version=$3
    
    if version_compare "$version" "$min_version" && ! version_compare "$version" "$max_version"; then
        return 0
    else
        return 1
    fi
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Header
echo
log "==================================="
log "Serena Installation Script"
log "==================================="
log "Build type: $BUILD_TYPE"
log "Mode: $MODE"
log "Skip dependencies: $SKIP_DEPS"
log "Run tests: $RUN_TESTS"
echo

# Check if serena is already installed
if command -v serena &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    CURRENT_VERSION=$(serena --version 2>/dev/null | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
    log "Serena is already installed (version: $CURRENT_VERSION)"
    log "Use --force to reinstall"
    exit 0
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for Serena..."
    
    # Update package list
    log "Updating package list..."
    sudo apt update
    
    # Install basic build dependencies
    log "Installing basic build tools..."
    sudo apt install -y build-essential git curl wget software-properties-common
    
    # Check Python installation
    PYTHON_CMD=""
    for py_cmd in python3.11 python3 python; do
        if command -v "$py_cmd" &> /dev/null; then
            PYTHON_VERSION=$($py_cmd --version 2>&1 | grep -oP '\d+\.\d+\.\d+')
            log "Found Python version: $PYTHON_VERSION ($py_cmd)"
            
            if version_in_range "$PYTHON_VERSION" "$PYTHON_MIN_VERSION" "$PYTHON_MAX_VERSION"; then
                PYTHON_CMD="$py_cmd"
                log "Python version is compatible (>= $PYTHON_MIN_VERSION, < $PYTHON_MAX_VERSION)"
                break
            else
                warning "Python version $PYTHON_VERSION is not in required range ($PYTHON_MIN_VERSION - $PYTHON_MAX_VERSION)"
            fi
        fi
    done
    
    if [[ -z "$PYTHON_CMD" ]]; then
        log "Installing Python 3.11..."
        sudo apt install -y software-properties-common
        sudo add-apt-repository -y ppa:deadsnakes/ppa
        sudo apt update
        sudo apt install -y python3.11 python3.11-dev python3.11-venv python3-pip
        PYTHON_CMD="python3.11"
        
        # Verify installation
        if ! command -v python3.11 &> /dev/null; then
            error "Failed to install Python 3.11"
        fi
        
        PYTHON_VERSION=$(python3.11 --version | grep -oP '\d+\.\d+\.\d+')
        log "Installed Python version: $PYTHON_VERSION"
    fi
    
    # Install uv (modern Python package installer)
    if ! command -v uv &> /dev/null; then
        log "Installing uv (modern Python package installer)..."
        curl -LsSf https://astral.sh/uv/install.sh | sh
        source ~/.cargo/env || export PATH="$HOME/.cargo/bin:$PATH"
        
        # Verify uv installation
        if ! command -v uv &> /dev/null; then
            error "Failed to install uv"
        fi
        
        UV_VERSION=$(uv --version | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
        log "Installed uv version: $UV_VERSION"
    else
        UV_VERSION=$(uv --version | grep -oP '\d+\.\d+\.\d+' || echo "unknown")
        log "Found uv version: $UV_VERSION"
    fi
    
    success "Dependencies installation completed!"
else
    log "Skipping dependency installation as requested"
    
    # Still need to verify critical dependencies
    if ! command -v python3.11 &> /dev/null && ! command -v python3 &> /dev/null; then
        error "Python is not available. Install Python 3.11 or run without --skip-deps"
    fi
    
    if ! command -v uv &> /dev/null; then
        error "uv is not available. Install uv or run without --skip-deps"
    fi
fi

# Ensure we have access to uv
if ! command -v uv &> /dev/null; then
    if [[ -f ~/.cargo/env ]]; then
        source ~/.cargo/env
    else
        export PATH="$HOME/.cargo/bin:$PATH"
    fi
    
    if ! command -v uv &> /dev/null; then
        error "uv is not available in PATH"
    fi
fi

# Create build directory
BUILD_DIR="${BUILD_DIR:-$HOME/tools/build}"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# Get Serena source directory
SERENA_SOURCE_DIR="$BUILD_DIR/$SERENA_DIR"

# Clone or update repository
if [[ ! -d "$SERENA_SOURCE_DIR" ]]; then
    log "Cloning Serena repository..."
    git clone "$SERENA_REPO" "$SERENA_SOURCE_DIR"
else
    log "Updating Serena repository..."
    cd "$SERENA_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$SERENA_SOURCE_DIR"

# Configure build
log "Configuring Serena build..."

# Verify we have pyproject.toml
if [[ ! -f "pyproject.toml" ]]; then
    error "pyproject.toml not found. This doesn't appear to be a valid Serena project."
fi

log "Serena source configured successfully"

if [[ "$MODE" == "config" ]]; then
    success "Configuration completed!"
    exit 0
fi

# Build/Install using uv
log "Installing Serena with uv..."

# Get build options based on build type
UV_INSTALL_OPTIONS=""
case $BUILD_TYPE in
    minimal)
        log "Installing minimal version (core features only)..."
        UV_INSTALL_OPTIONS="--no-dev"
        ;;
    standard)
        log "Installing standard version (recommended features)..."
        UV_INSTALL_OPTIONS=""
        ;;
    full)
        log "Installing full version (all optional dependencies)..."
        UV_INSTALL_OPTIONS="--all-extras"
        ;;
esac

# Create virtual environment for Serena (Python best practice)
log "Creating virtual environment for Serena..."
uv venv --python python3.11 .venv || error "Failed to create virtual environment"

# Install Serena in virtual environment
log "Installing Serena in virtual environment..."
if [[ -n "$UV_INSTALL_OPTIONS" ]]; then
    uv pip install -e . $UV_INSTALL_OPTIONS || error "Serena installation failed"
else
    uv pip install -e . || error "Serena installation failed"
fi

# Create global wrapper script for system-wide access
log "Creating global wrapper script..."
WRAPPER_SCRIPT="/usr/local/bin/serena"
sudo tee "$WRAPPER_SCRIPT" > /dev/null << EOF
#!/bin/bash
# Serena wrapper script - executes serena from virtual environment
exec "$SERENA_SOURCE_DIR/.venv/bin/serena" "\$@"
EOF

sudo chmod +x "$WRAPPER_SCRIPT"

# Verify wrapper works
if [[ -x "$WRAPPER_SCRIPT" ]]; then
    success "Global wrapper script created: $WRAPPER_SCRIPT"
else
    error "Failed to create executable wrapper script"
fi

success "Serena installation completed successfully!"

if [[ "$MODE" == "build" ]]; then
    success "Build completed!"
    log "Serena is installed and ready to use"
    exit 0
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running Serena tests..."
    if uv run python -m pytest tests/ 2>/dev/null || python3 -m pytest tests/ 2>/dev/null; then
        success "All tests passed!"
    else
        warning "Tests not available or failed, but continuing with installation"
    fi
fi

# Verify installation
log "Verifying installation..."

# Ensure uv-installed packages are in PATH
export PATH="$HOME/.local/bin:$PATH"

if command -v serena &> /dev/null; then
    INSTALLED_VERSION=$(serena --version 2>/dev/null | head -n1 || echo "installed")
    success "Serena installation verified!"
    log "Installed version: $INSTALLED_VERSION"
    
    # Show basic usage
    echo
    log "Basic usage examples:"
    log "  serena start-mcp-server              # Start MCP server"
    log "  serena project index                 # Index current project for faster operations"
    log "  serena --help                        # Show help and available commands"
    echo
    log "Configuration:"
    log "  ~/.config/serena/serena_config.yml   # Global configuration"
    log "  .serena/project.yml                  # Project-specific configuration"
    echo
    log "Integration:"
    log "  - Works with Claude Code, Claude Desktop, VSCode extensions"
    log "  - Provides semantic code retrieval and editing capabilities"
    log "  - Supports multiple programming languages via language servers"
    echo
    log "Build type: $BUILD_TYPE"
    case $BUILD_TYPE in
        minimal)
            log "Features: Core semantic analysis and retrieval tools"
            ;;
        standard)
            log "Features: Full semantic toolkit with recommended integrations"
            ;;
        full)
            log "Features: Complete toolkit with all optional dependencies"
            ;;
    esac
else
    error "Serena installation verification failed - serena command not found in PATH"
fi

success "Serena installation completed!"
log "You may need to restart your shell or run 'source ~/.bashrc' to ensure PATH changes are active"