#!/bin/bash
#
# Gearbox Quick Start Setup
# Installs essential tools for immediate productivity
# Usage: ./quickstart.sh [profile]
# Profiles: minimal, developer, full

set -euo pipefail

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
PURPLE='\033[0;35m'
NC='\033[0m'

PROFILE="${1:-developer}"

header() {
    echo -e "${PURPLE}$1${NC}"
}

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

install_profile() {
    case "$PROFILE" in
        "minimal")
            header "🚀 Minimal Profile: Essential CLI Tools"
            log "Installing: fd, ripgrep, fzf"
            gearbox install --minimal fd ripgrep fzf
            ;;
        "developer")
            header "💻 Developer Profile: Core Development Tools"
            log "Installing: fd, ripgrep, fzf, jq, nerd-fonts, starship"
            gearbox install fd ripgrep fzf jq
            gearbox install nerd-fonts --fonts="FiraCode,JetBrainsMono,Hack"
            gearbox install starship
            ;;
        "full")
            header "🎯 Full Profile: Complete Terminal Experience"
            log "Installing: All essential tools + media processing"
            gearbox install fd ripgrep fzf jq zoxide bat eza
            gearbox install nerd-fonts --standard
            gearbox install starship delta lazygit bottom
            ;;
        *)
            echo "Usage: $0 [minimal|developer|full]"
            echo ""
            echo "Profiles:"
            echo "  minimal   - fd, ripgrep, fzf (fastest)"
            echo "  developer - Core dev tools + fonts (recommended)"
            echo "  full      - Complete setup (slower but comprehensive)"
            exit 1
            ;;
    esac
}

show_setup_complete() {
    header "🎉 Quick Start Complete!"
    echo
    echo -e "${GREEN}Your terminal is now enhanced!${NC}"
    echo
    echo -e "${YELLOW}Quick verification:${NC}"
    echo "  gearbox status nerd-fonts    # Check fonts"
    echo "  gearbox doctor               # System health"
    echo "  fd --version                 # Test fd"
    echo "  rg --version                 # Test ripgrep"
    echo
    echo -e "${YELLOW}Next steps:${NC}"
    echo "  1. Restart your terminal"
    echo "  2. Try: fd .md | head        # Find markdown files"
    echo "  3. Try: rg TODO              # Search for TODOs"
    echo "  4. Configure your editor to use the new fonts"
    echo
}

main() {
    # Check if gearbox is available
    if ! command -v gearbox >/dev/null 2>&1; then
        echo "❌ gearbox not found in PATH"
        echo "Please run the installer first:"
        echo "  curl -fsSL https://raw.githubusercontent.com/makutaku/gearbox/main/install.sh | bash"
        exit 1
    fi
    
    echo "🚀 Gearbox Quick Start"
    echo "Profile: $PROFILE"
    echo
    
    install_profile
    show_setup_complete
}

main "$@"