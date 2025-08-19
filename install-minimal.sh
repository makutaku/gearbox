#!/bin/bash
#
# Gearbox Minimal Installer
# For systems with dependencies already installed
# One-liner: curl -fsSL https://raw.githubusercontent.com/makutaku/gearbox/main/install-minimal.sh | bash

set -euo pipefail

GEARBOX_REPO="https://github.com/makutaku/gearbox.git"
INSTALL_DIR="$HOME/gearbox"
BINARY_PATH="$HOME/.local/bin"

echo "🚀 Gearbox Minimal Installer"
echo "=============================="

# Create directories
mkdir -p "$BINARY_PATH"

# Clone or update
if [[ -d "$INSTALL_DIR" ]]; then
    echo "Updating existing installation..."
    cd "$INSTALL_DIR" && git pull
else
    echo "Cloning gearbox..."
    git clone "$GEARBOX_REPO" "$INSTALL_DIR"
fi

# Build and install
cd "$INSTALL_DIR"
echo "Building CLI..."
make cli

echo "Installing to $BINARY_PATH..."
cp gearbox "$BINARY_PATH/"
chmod +x "$BINARY_PATH/gearbox"

# Add to PATH if needed
if ! echo "$PATH" | grep -q "$BINARY_PATH"; then
    echo "Adding to PATH..."
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
fi

echo "✅ Installation complete!"
echo ""
echo "Next steps:"
echo "  source ~/.bashrc      # Reload shell"
echo "  gearbox list          # See available tools"
echo "  gearbox install fd    # Install your first tool"