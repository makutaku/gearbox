#!/bin/bash
# dust Installation Script - Better disk usage analyzer
set -e
RED='\033[0;31m'; GREEN='\033[0;32m'; BLUE='\033[0;34m'; NC='\033[0m'
BUILD_TYPE="release"; INSTALL_METHOD="binary"; FORCE_INSTALL=false
log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
    case $1 in
        --source) INSTALL_METHOD="source"; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        -h|--help) echo "dust - Better disk usage analyzer. Options: --source, --force"; exit 0 ;;
        *) shift ;;
    esac
done

[[ $EUID -eq 0 ]] && error "Don't run as root"
log "Installing dust - Better disk usage analyzer"

if command -v dust &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "dust already installed. Use --force to reinstall"; exit 0
fi

if [[ "$INSTALL_METHOD" == "binary" ]]; then
    ARCH=$(uname -m); case $ARCH in x86_64) ARCH_TAG="x86_64-unknown-linux-musl" ;; *) error "Unsupported arch" ;; esac
    VERSION=$(curl -s "https://api.github.com/repos/bootandy/dust/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)
    DOWNLOAD_URL="https://github.com/bootandy/dust/releases/download/${VERSION}/dust-${VERSION}-${ARCH_TAG}.tar.gz"
    TEMP_DIR=$(mktemp -d); cd "$TEMP_DIR"
    curl -fLO "$DOWNLOAD_URL" && tar -xzf "$(basename "$DOWNLOAD_URL")"
    DUST_BIN=$(find . -name "dust" -type f -executable | head -1)
    [[ -z "$DUST_BIN" ]] && error "dust binary not found"
    sudo cp "$DUST_BIN" /usr/local/bin/dust; sudo chmod +x /usr/local/bin/dust
    cd /; rm -rf "$TEMP_DIR"
else
    # Source build
    ! command -v rustc &> /dev/null && { curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y; source ~/.cargo/env; }
    cargo install du-dust
fi

if command -v dust &> /dev/null; then
    success "dust installed successfully!"
    log "Usage: dust, dust -r (reverse), dust -n 20 (top 20), dust -d 3 (depth 3)"
else
    error "Installation failed"
fi