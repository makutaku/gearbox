#\!/bin/bash
# choose - Human-friendly cut/awk alternative
set -e; RED='\033[0;31m'; GREEN='\033[0;32m'; BLUE='\033[0;34m'; NC='\033[0m'
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }; success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }; error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
[[ $EUID -eq 0 ]] && error "Don't run as root"; log "Installing choose - Human-friendly cut alternative"
if command -v choose &> /dev/null && [[ "$1" \!= "--force" ]]; then log "choose already installed"; exit 0; fi
\! command -v rustc &> /dev/null && { curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y; source ~/.cargo/env; }
cargo install choose
command -v choose &> /dev/null && success "choose installed\! Usage: echo 'a b c' | choose 1" || error "Installation failed"
