#!/bin/bash
# Basic test runner for installation scripts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

echo "Testing installation scripts..."

# Test that all scripts are executable
# Check categorized scripts
find "$REPO_DIR/scripts/installation/categories" -name "install-*.sh" -type f | while read -r script; do
    if [[ -x "$script" ]]; then
        echo "✓ $(basename "$script") is executable"
    else
        echo "✗ $(basename "$script") is not executable"
        exit 1
    fi
done

# Check common scripts
find "$REPO_DIR/scripts/installation/common" -name "install-*.sh" -type f | while read -r script; do
    if [[ -x "$script" ]]; then
        echo "✓ $(basename "$script") is executable"
    else
        echo "✗ $(basename "$script") is not executable"
        exit 1
    fi
done

# Test configuration loading
if source "$REPO_DIR/scripts/lib/config.sh"; then
    echo "✓ scripts/lib/config.sh loads successfully"
else
    echo "✗ scripts/lib/config.sh failed to load"
    exit 1
fi

echo "All tests passed!"
