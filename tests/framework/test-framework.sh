#!/bin/bash

# Gearbox Testing Framework
# Comprehensive testing system for shell functions, integrations, and templates

set -e

# Framework configuration
FRAMEWORK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="$(dirname "$FRAMEWORK_DIR")"
REPO_DIR="$(dirname "$TESTS_DIR")"
TEMP_TEST_DIR=""

# Test results tracking
declare -a PASSED_TESTS=()
declare -a FAILED_TESTS=()
declare -a SKIPPED_TESTS=()
TOTAL_TESTS=0
START_TIME=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Source the common library for testing
if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: Cannot load scripts/lib/common.sh for testing" >&2
    exit 1
fi

# Test framework functions
test_start() {
    local test_name="$1"
    echo -e "${BLUE}[TEST]${NC} Starting: $test_name"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

test_pass() {
    local test_name="$1"
    echo -e "${GREEN}[PASS]${NC} $test_name"
    PASSED_TESTS+=("$test_name")
}

test_fail() {
    local test_name="$1"
    local error_msg="$2"
    echo -e "${RED}[FAIL]${NC} $test_name: $error_msg"
    FAILED_TESTS+=("$test_name")
}

test_skip() {
    local test_name="$1"
    local reason="$2"
    echo -e "${YELLOW}[SKIP]${NC} $test_name: $reason"
    SKIPPED_TESTS+=("$test_name")
}

# Assert functions
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Values should be equal}"
    
    if [[ "$expected" == "$actual" ]]; then
        return 0
    else
        echo "Expected: '$expected', Got: '$actual' - $message"
        return 1
    fi
}

assert_not_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Values should not be equal}"
    
    if [[ "$expected" != "$actual" ]]; then
        return 0
    else
        echo "Expected: not '$expected', Got: '$actual' - $message"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-String should contain substring}"
    
    if [[ "$haystack" == *"$needle"* ]]; then
        return 0
    else
        echo "Expected '$haystack' to contain '$needle' - $message"
        return 1
    fi
}

assert_file_exists() {
    local file="$1"
    local message="${2:-File should exist}"
    
    if [[ -f "$file" ]]; then
        return 0
    else
        echo "File '$file' does not exist - $message"
        return 1
    fi
}

assert_command_success() {
    local command="$1"
    local message="${2:-Command should succeed}"
    
    if eval "$command" >/dev/null 2>&1; then
        return 0
    else
        echo "Command '$command' failed - $message"
        return 1
    fi
}

assert_command_failure() {
    local command="$1"
    local message="${2:-Command should fail}"
    
    if ! eval "$command" >/dev/null 2>&1; then
        return 0
    else
        echo "Command '$command' succeeded but should have failed - $message"
        return 1
    fi
}

# Test environment setup
setup_test_env() {
    TEMP_TEST_DIR=$(mktemp -d)
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    cd "$TEMP_TEST_DIR"
}

teardown_test_env() {
    cd "$REPO_DIR"
    if [[ -n "$TEMP_TEST_DIR" && -d "$TEMP_TEST_DIR" ]]; then
        rm -rf "$TEMP_TEST_DIR"
    fi
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
}

# Test discovery and execution
discover_tests() {
    local test_dir="$1"
    find "$test_dir" -name "test_*.sh" -type f -executable | sort
}

run_test_suite() {
    local test_suite="$1"
    local suite_name=$(basename "$test_suite" .sh)
    
    echo -e "\n${BLUE}=== Running Test Suite: $suite_name ===${NC}"
    
    # Source the test suite
    if ! source "$test_suite"; then
        test_fail "$suite_name" "Failed to source test suite"
        return 1
    fi
    
    # Run setup if it exists
    if declare -f setup >/dev/null; then
        setup_test_env
        setup || {
            test_fail "$suite_name" "Setup failed"
            teardown_test_env
            return 1
        }
    fi
    
    # Run all test functions
    local test_functions
    test_functions=$(declare -F | grep "declare -f test_" | cut -d' ' -f3)
    
    for test_func in $test_functions; do
        test_start "$suite_name::$test_func"
        
        if $test_func; then
            test_pass "$suite_name::$test_func"
        else
            test_fail "$suite_name::$test_func" "Test function failed"
        fi
    done
    
    # Run teardown if it exists
    if declare -f teardown >/dev/null; then
        teardown || {
            test_fail "$suite_name" "Teardown failed"
        }
        teardown_test_env
    fi
}

# Test execution engine
run_all_tests() {
    START_TIME=$(date +%s)
    echo -e "${BLUE}🧪 Gearbox Testing Framework${NC}"
    echo -e "${BLUE}==============================${NC}\n"
    
    # Discover all test suites
    local test_suites
    test_suites=$(discover_tests "$TESTS_DIR")
    
    if [[ -z "$test_suites" ]]; then
        echo "No test suites found in $TESTS_DIR"
        return 0
    fi
    
    # Run each test suite
    for suite in $test_suites; do
        run_test_suite "$suite"
    done
    
    # Print summary
    print_test_summary
}

run_specific_tests() {
    local pattern="$1"
    START_TIME=$(date +%s)
    echo -e "${BLUE}🧪 Running tests matching: $pattern${NC}\n"
    
    local test_suites
    test_suites=$(discover_tests "$TESTS_DIR" | grep "$pattern")
    
    for suite in $test_suites; do
        run_test_suite "$suite"
    done
    
    print_test_summary
}

print_test_summary() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    
    echo -e "\n${BLUE}=== Test Summary ===${NC}"
    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: ${#PASSED_TESTS[@]}${NC}"
    echo -e "${RED}Failed: ${#FAILED_TESTS[@]}${NC}"
    echo -e "${YELLOW}Skipped: ${#SKIPPED_TESTS[@]}${NC}"
    echo -e "Duration: ${duration}s"
    
    if [[ ${#FAILED_TESTS[@]} -gt 0 ]]; then
        echo -e "\n${RED}Failed Tests:${NC}"
        for test in "${FAILED_TESTS[@]}"; do
            echo -e "  - $test"
        done
        return 1
    else
        echo -e "\n${GREEN}🎉 All tests passed!${NC}"
        return 0
    fi
}

# CLI interface
show_help() {
    cat << EOF
Gearbox Testing Framework

Usage: $0 [OPTIONS] [PATTERN]

Options:
  --help, -h          Show this help message
  --verbose, -v       Enable verbose output
  --unit              Run only unit tests
  --integration       Run only integration tests
  --template          Run only template tests
  --performance       Run only performance tests

Examples:
  $0                  # Run all tests
  $0 unit             # Run tests matching 'unit'
  $0 --integration    # Run integration tests only
  $0 test_validation  # Run specific test

EOF
}

# Main execution
main() {
    local verbose=false
    local test_type=""
    local pattern=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            --verbose|-v)
                verbose=true
                shift
                ;;
            --unit)
                test_type="unit"
                shift
                ;;
            --integration)
                test_type="integration"
                shift
                ;;
            --template)
                test_type="template"
                shift
                ;;
            --performance)
                test_type="performance"
                shift
                ;;
            *)
                pattern="$1"
                shift
                ;;
        esac
    done
    
    # Set verbose mode
    if [[ "$verbose" == true ]]; then
        set -x
    fi
    
    # Determine what to run
    if [[ -n "$test_type" ]]; then
        pattern="test_${test_type}"
    fi
    
    if [[ -n "$pattern" ]]; then
        run_specific_tests "$pattern"
    else
        run_all_tests
    fi
}

# Execute if called directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi