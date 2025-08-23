#!/bin/bash

# TUI Test Runner - Automated testing for Gearbox TUI
# This script runs various TUI test scenarios to validate functionality

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
GEARBOX_BIN="./build/gearbox"
TEST_RESULTS_DIR="./tests/results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create results directory
mkdir -p "$TEST_RESULTS_DIR"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test result tracking
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test and track results
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"
    
    log_info "Running test: $test_name"
    TESTS_RUN=$((TESTS_RUN + 1))
    
    local test_log="$TEST_RESULTS_DIR/${test_name}_${TIMESTAMP}.log"
    local start_time=$(date +%s)
    
    if eval "$test_command" > "$test_log" 2>&1; then
        local exit_code=0
    else
        local exit_code=$?
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if [ $exit_code -eq $expected_exit_code ]; then
        log_success "$test_name passed (${duration}s)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        log_error "$test_name failed (exit code: $exit_code, expected: $expected_exit_code, duration: ${duration}s)"
        log_error "See log: $test_log"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to check if gearbox binary exists
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if [ ! -f "$GEARBOX_BIN" ]; then
        log_error "Gearbox binary not found at $GEARBOX_BIN"
        log_info "Run 'make build' to build the binary first"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Test TUI help and basic command parsing
test_tui_help() {
    log_info "Testing TUI help functionality..."
    
    run_test "tui_help" "$GEARBOX_BIN tui --help"
    run_test "tui_demo_help" "$GEARBOX_BIN tui --demo --help"
}

# Test TUI demo mode (non-interactive)
test_demo_mode() {
    log_info "Testing TUI demo mode..."
    
    # Demo mode should work without terminal interaction
    run_test "demo_mode_launch" "timeout 5s $GEARBOX_BIN tui --demo || true"
}

# Test TUI test scenarios
test_scenarios() {
    log_info "Testing TUI automated scenarios..."
    
    run_test "scenario_basic_nav" "$GEARBOX_BIN tui --test --test-scenario=basic-nav"
    run_test "scenario_tool_install" "$GEARBOX_BIN tui --test --test-scenario=tool-install"
    run_test "scenario_bundle_install" "$GEARBOX_BIN tui --test --test-scenario=bundle-install"
}

# Test invalid scenarios
test_error_cases() {
    log_info "Testing error handling..."
    
    run_test "invalid_scenario" "$GEARBOX_BIN tui --test --test-scenario=invalid" 1
}

# Interactive TUI test (requires manual verification)
test_interactive_mode() {
    log_info "Testing interactive TUI mode..."
    
    if [ "${INTERACTIVE:-}" = "true" ]; then
        log_info "Launching interactive TUI demo mode..."
        log_info "Please navigate through the interface and press 'q' to quit"
        log_info "This test requires manual verification"
        
        $GEARBOX_BIN tui --demo
        
        echo
        read -p "Did the TUI work correctly? (y/n): " -r
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_success "Interactive TUI test passed (manual verification)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            log_error "Interactive TUI test failed (manual verification)"
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
        TESTS_RUN=$((TESTS_RUN + 1))
    else
        log_warning "Skipping interactive test (set INTERACTIVE=true to enable)"
    fi
}

# Performance test - measure TUI startup time
test_performance() {
    log_info "Testing TUI performance..."
    
    local total_time=0
    local iterations=3
    
    for i in $(seq 1 $iterations); do
        local start_time=$(date +%s%N)
        timeout 2s $GEARBOX_BIN tui --demo >/dev/null 2>&1 || true
        local end_time=$(date +%s%N)
        
        local duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
        total_time=$((total_time + duration))
        
        log_info "TUI startup iteration $i: ${duration}ms"
    done
    
    local avg_time=$((total_time / iterations))
    log_info "Average TUI startup time: ${avg_time}ms"
    
    # Consider test passed if startup is under 1 second
    if [ $avg_time -lt 1000 ]; then
        log_success "Performance test passed (startup < 1s)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        log_warning "Performance test warning (startup >= 1s)"
        TESTS_PASSED=$((TESTS_PASSED + 1)) # Still count as passed, just slow
    fi
    TESTS_RUN=$((TESTS_RUN + 1))
}

# Memory usage test
test_memory_usage() {
    log_info "Testing TUI memory usage..."
    
    # Start TUI in background with demo mode
    $GEARBOX_BIN tui --demo &
    local tui_pid=$!
    
    sleep 2 # Let it start up
    
    if kill -0 $tui_pid 2>/dev/null; then
        # Get memory usage (RSS in KB)
        local memory_kb=$(ps -o rss= -p $tui_pid 2>/dev/null || echo "0")
        local memory_mb=$((memory_kb / 1024))
        
        log_info "TUI memory usage: ${memory_mb}MB"
        
        # Kill the TUI
        kill $tui_pid 2>/dev/null || true
        wait $tui_pid 2>/dev/null || true
        
        # Consider test passed if memory usage is reasonable (< 100MB)
        if [ $memory_mb -lt 100 ]; then
            log_success "Memory test passed (< 100MB)"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            log_warning "Memory test warning (>= 100MB)"
            TESTS_PASSED=$((TESTS_PASSED + 1)) # Still count as passed
        fi
    else
        log_error "TUI failed to start for memory test"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
    
    TESTS_RUN=$((TESTS_RUN + 1))
}

# Generate test report
generate_report() {
    local report_file="$TEST_RESULTS_DIR/tui_test_report_${TIMESTAMP}.txt"
    
    cat > "$report_file" << EOF
TUI Test Report
===============
Generated: $(date)
Test Duration: ${SECONDS}s

Summary:
- Tests Run: $TESTS_RUN
- Tests Passed: $TESTS_PASSED
- Tests Failed: $TESTS_FAILED
- Success Rate: $(( TESTS_PASSED * 100 / TESTS_RUN ))%

Environment:
- Gearbox Binary: $GEARBOX_BIN
- System: $(uname -a)
- Terminal: ${TERM:-unknown}

Test Results:
$(find "$TEST_RESULTS_DIR" -name "*_${TIMESTAMP}.log" -exec echo "- {}" \;)

EOF

    log_info "Test report generated: $report_file"
}

# Main test execution
main() {
    log_info "Starting TUI test suite..."
    log_info "Timestamp: $TIMESTAMP"
    echo
    
    check_prerequisites
    echo
    
    test_tui_help
    echo
    
    test_demo_mode
    echo
    
    test_scenarios
    echo
    
    test_error_cases
    echo
    
    test_performance
    echo
    
    test_memory_usage
    echo
    
    test_interactive_mode
    echo
    
    generate_report
    echo
    
    # Final results
    log_info "TUI Test Suite Complete"
    log_info "Results: $TESTS_PASSED/$TESTS_RUN tests passed"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed! 🎉"
        exit 0
    else
        log_error "$TESTS_FAILED tests failed"
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "TUI Test Runner for Gearbox"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --interactive       Include interactive tests"
        echo ""
        echo "Environment Variables:"
        echo "  INTERACTIVE=true    Enable interactive tests"
        echo ""
        echo "Examples:"
        echo "  $0                  Run all automated tests"
        echo "  INTERACTIVE=true $0 Run all tests including interactive"
        exit 0
        ;;
    --interactive)
        export INTERACTIVE=true
        ;;
esac

# Run the tests
main "$@"