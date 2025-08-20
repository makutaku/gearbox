#!/bin/bash

# Test script for bundle functionality

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GEARBOX_DIR="$(dirname "$SCRIPT_DIR")"
GEARBOX_BIN="$GEARBOX_DIR/build/gearbox"

# Function to print test result
print_result() {
    local test_name="$1"
    local result="$2"
    
    if [ "$result" = "pass" ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        ((FAILED++))
    fi
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Test 1: Check if gearbox binary exists
test_gearbox_exists() {
    if [ -f "$GEARBOX_BIN" ]; then
        print_result "Gearbox binary exists" "pass"
    else
        print_result "Gearbox binary exists" "fail"
        echo "  Error: Gearbox binary not found at $GEARBOX_BIN"
        echo "  Please run 'make build' first"
        exit 1
    fi
}

# Test 2: List bundles
test_list_bundles() {
    local output
    output=$("$GEARBOX_BIN" list bundles 2>&1) || true
    
    if echo "$output" | grep -q "Available Bundles"; then
        print_result "List bundles command works" "pass"
    else
        print_result "List bundles command works" "fail"
        echo "  Output: $output"
    fi
    
    # Check for specific bundles
    for bundle in "essential" "developer" "minimal"; do
        if echo "$output" | grep -q "$bundle"; then
            print_result "  - Bundle '$bundle' is listed" "pass"
        else
            print_result "  - Bundle '$bundle' is listed" "fail"
        fi
    done
}

# Test 3: Show bundle details
test_show_bundle() {
    local bundles=("essential" "developer" "minimal")
    
    for bundle in "${bundles[@]}"; do
        local output
        output=$("$GEARBOX_BIN" show bundle "$bundle" 2>&1) || true
        
        if echo "$output" | grep -q "Bundle: $bundle"; then
            print_result "Show bundle '$bundle' works" "pass"
        else
            print_result "Show bundle '$bundle' works" "fail"
            echo "  Output: $output"
        fi
        
        # Check for tools listing
        if echo "$output" | grep -q "Tools:"; then
            print_result "  - Shows tools for '$bundle'" "pass"
        else
            print_result "  - Shows tools for '$bundle'" "fail"
        fi
    done
}

# Test 4: Show non-existent bundle
test_show_nonexistent_bundle() {
    local output
    output=$("$GEARBOX_BIN" show bundle nonexistent 2>&1) || true
    
    if echo "$output" | grep -qi "not found\|error"; then
        print_result "Error handling for non-existent bundle" "pass"
    else
        print_result "Error handling for non-existent bundle" "fail"
        echo "  Expected error message, got: $output"
    fi
}

# Test 5: Install with --bundle flag (dry run)
test_install_bundle_flag() {
    local output
    output=$("$GEARBOX_BIN" install --bundle minimal --dry-run 2>&1) || true
    
    if echo "$output" | grep -q "tools\|Tools\|Installing"; then
        print_result "Install --bundle flag recognized" "pass"
    else
        print_result "Install --bundle flag recognized" "fail"
        echo "  Output: $output"
    fi
}

# Test 6: Test bundle expansion in dry-run
test_bundle_expansion() {
    local output
    output=$("$GEARBOX_BIN" install essential --dry-run 2>&1) || true
    
    # Check if it recognizes 'essential' as a bundle
    if echo "$output" | grep -qi "bundle\|expanded"; then
        print_result "Bundle name expansion works" "pass"
    else
        print_result "Bundle name expansion works" "fail"
        echo "  Output: $output"
    fi
}

# Main test execution
echo "Running Gearbox Bundle Tests"
echo "============================"
echo

# Run tests
test_gearbox_exists
test_list_bundles
test_show_bundle
test_show_nonexistent_bundle
test_install_bundle_flag
test_bundle_expansion

echo
echo "============================"
echo "Test Summary:"
echo -e "  Passed: ${GREEN}$PASSED${NC}"
echo -e "  Failed: ${RED}$FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi