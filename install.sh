#!/bin/bash
#
# Gearbox Easy Installer
# One-liner: curl -fsSL https://raw.githubusercontent.com/makutaku/gearbox/main/install.sh | bash
#
# This script:
# 1. Checks system requirements
# 2. Installs dependencies if needed
# 3. Clones the repository
# 4. Builds the CLI
# 5. Sets up the environment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
GEARBOX_REPO="https://github.com/makutaku/gearbox.git"
INSTALL_DIR="$HOME/gearbox"
BINARY_PATH="$HOME/.local/bin"

# Create .local/bin if it doesn't exist
mkdir -p "$BINARY_PATH"

# Logging functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

header() {
    echo -e "${PURPLE}$1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check system requirements
check_requirements() {
    header "🔍 Checking System Requirements"
    
    local missing_deps=()
    
    # Check OS
    if [[ ! -f /etc/os-release ]]; then
        error "Unsupported operating system. This installer requires a Linux distribution."
        exit 1
    fi
    
    local os_name
    os_name=$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '"')
    log "Operating System: $os_name"
    
    # Check required commands
    local required_commands=("git" "curl" "make")
    for cmd in "${required_commands[@]}"; do
        if ! command_exists "$cmd"; then
            missing_deps+=("$cmd")
        else
            log "✓ $cmd found"
        fi
    done
    
    # Check Go
    if ! command_exists "go"; then
        missing_deps+=("golang")
    else
        local go_version
        go_version=$(go version | cut -d' ' -f3 | sed 's/go//')
        log "✓ Go found: $go_version"
        
        # Check Go version (need 1.22+)
        if ! go version | grep -E "go1\.(2[2-9]|[3-9][0-9])" >/dev/null; then
            warning "Go 1.22+ recommended, found: $go_version"
        fi
    fi
    
    # Check build essentials
    if ! command_exists "gcc"; then
        missing_deps+=("build-essential")
    else
        log "✓ build tools found"
    fi
    
    return ${#missing_deps[@]}
}

# Install missing dependencies
install_dependencies() {
    header "📦 Installing Dependencies"
    
    local missing_deps=()
    
    # Recheck what's missing
    command_exists "git" || missing_deps+=("git")
    command_exists "curl" || missing_deps+=("curl")
    command_exists "make" || missing_deps+=("make")
    command_exists "gcc" || missing_deps+=("build-essential")
    command_exists "go" || missing_deps+=("golang-go")
    
    if [[ ${#missing_deps[@]} -eq 0 ]]; then
        success "All dependencies are already installed"
        return 0
    fi
    
    log "Installing missing packages: ${missing_deps[*]}"
    
    # Detect package manager and install
    if command_exists "apt"; then
        log "Using apt package manager..."
        sudo apt update
        sudo apt install -y "${missing_deps[@]}"
    elif command_exists "yum"; then
        log "Using yum package manager..."
        # Adjust package names for RHEL/CentOS
        local yum_deps=()
        for dep in "${missing_deps[@]}"; do
            case "$dep" in
                "build-essential") yum_deps+=("gcc" "gcc-c++" "make") ;;
                "golang-go") yum_deps+=("golang") ;;
                *) yum_deps+=("$dep") ;;
            esac
        done
        sudo yum install -y "${yum_deps[@]}"
    elif command_exists "dnf"; then
        log "Using dnf package manager..."
        local dnf_deps=()
        for dep in "${missing_deps[@]}"; do
            case "$dep" in
                "build-essential") dnf_deps+=("gcc" "gcc-c++" "make") ;;
                "golang-go") dnf_deps+=("golang") ;;
                *) dnf_deps+=("$dep") ;;
            esac
        done
        sudo dnf install -y "${dnf_deps[@]}"
    else
        error "No supported package manager found (apt, yum, dnf)"
        error "Please install these dependencies manually: ${missing_deps[*]}"
        exit 1
    fi
    
    success "Dependencies installed successfully"
}

# Clone or update repository
setup_repository() {
    header "📂 Setting Up Repository"
    
    if [[ -d "$INSTALL_DIR" ]]; then
        log "Gearbox directory exists, updating..."
        cd "$INSTALL_DIR"
        git pull origin main
    else
        log "Cloning gearbox repository..."
        git clone "$GEARBOX_REPO" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
    fi
    
    success "Repository ready: $INSTALL_DIR"
}

# Build the CLI
build_gearbox() {
    header "🔨 Building Gearbox CLI"
    
    cd "$INSTALL_DIR"
    
    log "Building CLI binary..."
    make cli
    
    if [[ ! -f "gearbox" ]]; then
        error "Build failed: gearbox binary not found"
        exit 1
    fi
    
    log "Installing binary to $BINARY_PATH..."
    cp gearbox "$BINARY_PATH/"
    chmod +x "$BINARY_PATH/gearbox"
    
    success "Gearbox CLI built and installed successfully"
}

# Setup shell integration
setup_shell() {
    header "🐚 Setting Up Shell Integration"
    
    # Add ~/.local/bin to PATH if not already there
    local shell_rc=""
    if [[ -n "${BASH_VERSION:-}" ]] && [[ -f "$HOME/.bashrc" ]]; then
        shell_rc="$HOME/.bashrc"
    elif [[ -n "${ZSH_VERSION:-}" ]] && [[ -f "$HOME/.zshrc" ]]; then
        shell_rc="$HOME/.zshrc"
    elif [[ -f "$HOME/.profile" ]]; then
        shell_rc="$HOME/.profile"
    fi
    
    if [[ -n "$shell_rc" ]]; then
        if ! grep -q "$BINARY_PATH" "$shell_rc"; then
            log "Adding $BINARY_PATH to PATH in $shell_rc"
            echo "" >> "$shell_rc"
            echo "# Added by Gearbox installer" >> "$shell_rc"
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$shell_rc"
            success "PATH updated in $shell_rc"
        else
            log "PATH already includes $BINARY_PATH"
        fi
    else
        warning "Could not detect shell configuration file"
        warning "Please manually add $BINARY_PATH to your PATH"
    fi
}

# Verify installation
verify_installation() {
    header "✅ Verifying Installation"
    
    # Add to PATH for this session
    export PATH="$BINARY_PATH:$PATH"
    
    if command_exists "gearbox"; then
        local version
        version=$(gearbox --version 2>/dev/null || echo "unknown")
        success "Gearbox CLI installed successfully: $version"
        
        log "Testing basic functionality..."
        gearbox list >/dev/null 2>&1 && success "✓ gearbox list works"
        
        return 0
    else
        error "Installation verification failed"
        error "Binary should be at: $BINARY_PATH/gearbox"
        return 1
    fi
}

# Show next steps
show_next_steps() {
    header "🎉 Installation Complete!"
    
    echo
    echo -e "${GREEN}Gearbox is now installed!${NC}"
    echo
    echo -e "${WHITE}Next steps:${NC}"
    echo -e "  ${CYAN}1.${NC} Restart your shell or run: ${YELLOW}source ~/.bashrc${NC}"
    echo -e "  ${CYAN}2.${NC} ${GREEN}NEW!${NC} Launch interactive TUI: ${YELLOW}gearbox tui${NC}"
    echo -e "  ${CYAN}3.${NC} See available tools: ${YELLOW}gearbox list${NC}"
    echo -e "  ${CYAN}4.${NC} Install essential tools: ${YELLOW}gearbox install fd ripgrep fzf${NC}"
    echo -e "  ${CYAN}5.${NC} Setup beautiful terminal: ${YELLOW}gearbox install nerd-fonts starship${NC}"
    echo -e "  ${CYAN}6.${NC} Check system health: ${YELLOW}gearbox doctor${NC}"
    echo
    echo -e "${WHITE}Quick start examples:${NC}"
    echo -e "  ${BLUE}# Core development tools${NC}"
    echo -e "  ${YELLOW}gearbox install fd ripgrep fzf jq${NC}"
    echo
    echo -e "  ${BLUE}# Enhanced terminal experience${NC}"
    echo -e "  ${YELLOW}gearbox install nerd-fonts starship${NC}"
    echo
    echo -e "  ${BLUE}# Check font installation${NC}"
    echo -e "  ${YELLOW}gearbox status nerd-fonts${NC}"
    echo
    echo -e "${WHITE}Documentation:${NC} ${INSTALL_DIR}/docs/USER_GUIDE.md"
    echo -e "${WHITE}Repository:${NC} ${INSTALL_DIR}"
    echo
}

# Main installation flow
main() {
    clear
    
    # ASCII art header
    cat << 'EOF'
   ▄████  ▄▄▄ ▄▄▄     ▄████▄   ▄▄▄▄    ▄▄▄█████▓
  ▓█   ▀ ▒████▄   ▒██▀ ▀█  ▓█████▄  ▓  ██▒ ▓▒
  ▒███   ▒██  ▀█▄ ▒▓█    ▄ ▒██▒ ▄██ ▒ ▓██░ ▒░
  ▒▓█  ▄ ░██▄▄▄▄██▒▓▓▄ ▄██▒▒██░█▀   ░ ▓██▓ ░ 
  ░▒████▒ ▓█   ▓██▒ ▓███▀ ░░▓█  ▀█▓   ▒██▒ ░ 
  ░░ ▒░ ░ ▒▒   ▓▒█░ ░▒ ▒  ░░▒▓███▀▒   ▒ ░░   
   ░ ░  ░  ▒   ▒▒ ░ ░  ▒   ▒░▒   ░      ░    
     ░     ░   ▒  ░       ░  ░    ░    ░      
     ░  ░      ░  ░ ░          ░              
                  ░               ░           
EOF
    
    echo -e "${PURPLE}🚀 Gearbox Easy Installer${NC}"
    echo -e "${CYAN}Essential command-line tools built from source${NC}"
    echo
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        error "Do not run this installer as root!"
        error "It will install to your user directory: $HOME"
        exit 1
    fi
    
    # Main installation steps
    if ! check_requirements; then
        log "Missing dependencies detected, installing..."
        install_dependencies
    fi
    
    setup_repository
    build_gearbox
    setup_shell
    
    if verify_installation; then
        show_next_steps
    else
        error "Installation failed during verification"
        exit 1
    fi
}

# Handle Ctrl+C gracefully
trap 'echo -e "\n${YELLOW}Installation cancelled by user${NC}"; exit 130' INT

# Run main function
main "$@"