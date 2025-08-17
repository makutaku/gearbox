#\!/bin/bash
# tealdeer - Fast tldr pages client
set -e; RED='\033[0;31m'; GREEN='\033[0;32m'; BLUE='\033[0;34m'; NC='\033[0m'
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }; success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }; error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
[[ $EUID -eq 0 ]] && error "Don't run as root"; log "Installing tealdeer - Fast tldr client"
if command -v tldr &> /dev/null && [[ "$1" \!= "--force" ]]; then log "tldr already installed"; exit 0; fi
\! command -v rustc &> /dev/null && { curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y; source ~/.cargo/env; }
cargo install tealdeer
command -v tldr &> /dev/null && { success "tealdeer installed\! Usage: tldr COMMAND"; tldr --update; } || error "Installation failed"
