#!/bin/bash
# Basic test runner for installation scripts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

echo "Testing installation scripts..."

# Test that all scripts are executable
for script in "$REPO_DIR/scripts"/install-*.sh; do
    if [[ -x "$script" ]]; then
        echo "✓ $(basename "$script") is executable"
    else
        echo "✗ $(basename "$script") is not executable"
        exit 1
    fi
done

# Test configuration loading
if source "$REPO_DIR/config.sh"; then
    echo "✓ config.sh loads successfully"
else
    echo "✗ config.sh failed to load"
    exit 1
fi

echo "All tests passed!"
