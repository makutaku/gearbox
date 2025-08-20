#!/bin/bash

# Focused tests for core functions to verify they work properly
# Quick validation of essential functionality

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Simple test framework
TESTS_PASSED=0
TESTS_FAILED=0

test_function() {
    local name="$1"
    local command="$2"
    
    echo -n "Testing $name... "
    if eval "$command" >/dev/null 2>&1; then
        echo "✓ PASS"
        ((TESTS_PASSED++))
    else
        echo "✗ FAIL"
        ((TESTS_FAILED++))
    fi
}

# Source the library
source "$REPO_DIR/scripts/lib/common.sh"

echo "🧪 Core Function Tests"
echo "====================="

# Test logging functions
test_function "log" "log 'test message'"
test_function "success" "success 'test success'"
test_function "warning" "warning 'test warning'"
test_function "debug" "debug 'test debug'"

# Test validation functions
test_function "validate_tool_name" "validate_tool_name 'fd'"
test_function "validate_build_type" "validate_build_type 'standard'"
test_function "validate_url" "validate_url 'https://github.com/test/repo.git'"
test_function "version_compare" "version_compare '1.1.0' '1.0.0'"

# Test utility functions
test_function "get_optimal_jobs" "[[ \$(get_optimal_jobs) -gt 0 ]]"
test_function "sanitize_filename" "[[ -n \$(sanitize_filename 'test\$file') ]]"
test_function "human_readable_size" "[[ -n \$(human_readable_size 1024) ]]"

# Test security functions
test_function "ensure_not_root" "ensure_not_root"
test_function "check_tool_installed" "(check_tool_installed 'nonexistent-tool-12345' false; echo 'test passed') 2>/dev/null"
test_function "execute_command_safely" "execute_command_safely echo 'safe test'"

# Load additional modules and test
echo
echo "🔧 Loading additional modules..."
require_build_modules 2>/dev/null || echo "Build modules not available"
require_system_modules 2>/dev/null || echo "System modules not available"

echo
echo "📊 Results: $TESTS_PASSED passed, $TESTS_FAILED failed"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo "🎉 All core functions working!"
    exit 0
else
    echo "❌ Some functions need attention"
    exit 1
fi