#!/bin/bash

# TUI Demo Script - Interactive demonstration of Gearbox TUI functionality
# This script provides guided examples of TUI usage for testing and demonstration

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

GEARBOX_BIN="./build/gearbox"

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_demo() {
    echo -e "${CYAN}🎬 $1${NC}"
}

# Function to wait for user input
wait_for_user() {
    echo
    read -p "Press Enter to continue..." -r
    echo
}

# Function to check prerequisites
check_prerequisites() {
    echo "🧪 Gearbox TUI Testing & Demo Suite"
    echo "===================================="
    echo

    if [ ! -f "$GEARBOX_BIN" ]; then
        log_error "Gearbox binary not found at $GEARBOX_BIN"
        log_info "Please run 'make build' first to build the binary"
        exit 1
    fi

    log_success "Gearbox binary found at $GEARBOX_BIN"
    echo
}

# Demo 1: Basic TUI Help
demo_help() {
    log_demo "Demo 1: TUI Help and Command Line Interface"
    echo
    log_info "First, let's see the TUI help to understand available options:"
    echo

    echo "Command: $GEARBOX_BIN tui --help"
    echo
    $GEARBOX_BIN tui --help
    echo

    log_success "Demo 1 completed - TUI help displayed"
    wait_for_user
}

# Demo 2: Automated Test Scenarios
demo_test_scenarios() {
    log_demo "Demo 2: Automated Test Scenarios"
    echo
    log_info "Gearbox TUI includes automated test scenarios that simulate user interactions"
    log_info "These are useful for CI/CD testing and validating TUI functionality"
    echo

    scenarios=("basic-nav" "tool-install" "bundle-install")

    for scenario in "${scenarios[@]}"; do
        log_info "Running test scenario: $scenario"
        echo "Command: $GEARBOX_BIN tui --test --test-scenario=$scenario"
        echo

        if $GEARBOX_BIN tui --test --test-scenario="$scenario"; then
            log_success "Scenario '$scenario' completed successfully"
        else
            log_error "Scenario '$scenario' failed"
        fi
        echo
    done

    log_success "Demo 2 completed - All automated scenarios tested"
    wait_for_user
}

# Demo 3: Interactive Demo Mode
demo_interactive() {
    log_demo "Demo 3: Interactive TUI Demo Mode"
    echo
    log_info "Now let's launch the TUI in demo mode for interactive exploration"
    log_info "Demo mode uses mock data, so you can safely explore without affecting your system"
    echo

    log_info "🎮 TUI Controls:"
    echo "  • Tab / Arrow Keys: Navigate between views"
    echo "  • ↑/↓ or j/k: Navigate within lists"
    echo "  • Enter/Space: Select items"
    echo "  • Letters (T/B/I/C/H/D): Jump to specific views"
    echo "  • ?: Help screen"
    echo "  • q: Quit"
    echo

    log_info "📋 Views to explore:"
    echo "  • Dashboard (D): System overview and quick actions"
    echo "  • Tools (T): Browse and select individual tools"
    echo "  • Bundles (B): Explore curated tool collections"
    echo "  • Install Manager (I): Monitor installation progress"
    echo "  • Configuration (C): Adjust Gearbox settings"
    echo "  • Health Monitor (H): System health checks"
    echo

    log_warning "Try the following test flow:"
    echo "  1. Navigate to Tools (press T)"
    echo "  2. Scroll through tools (↑/↓ arrows)"
    echo "  3. Select a tool (press Space)"
    echo "  4. Install selected tools (press i)"
    echo "  5. Switch to Install Manager to see progress"
    echo "  6. Try bundle installation in Bundles view"
    echo

    read -p "Ready to launch interactive demo? (y/n): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Launching TUI demo mode..."
        echo "Command: $GEARBOX_BIN tui --demo"
        echo

        $GEARBOX_BIN tui --demo

        echo
        log_success "Demo 3 completed - Interactive exploration finished"
    else
        log_info "Skipping interactive demo"
    fi

    wait_for_user
}

# Demo 4: Performance Testing
demo_performance() {
    log_demo "Demo 4: Performance Testing"
    echo
    log_info "Let's test TUI performance with multiple rapid launches"
    echo

    log_info "Testing TUI startup time (5 iterations)..."
    total_time=0

    for i in {1..5}; do
        log_info "Iteration $i/5..."
        
        start_time=$(date +%s%N)
        timeout 3s $GEARBOX_BIN tui --demo >/dev/null 2>&1 || true
        end_time=$(date +%s%N)
        
        duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
        total_time=$((total_time + duration))
        
        echo "  Startup time: ${duration}ms"
    done

    avg_time=$((total_time / 5))
    log_success "Average startup time: ${avg_time}ms"

    if [ $avg_time -lt 1000 ]; then
        log_success "Performance: Excellent (< 1 second)"
    elif [ $avg_time -lt 2000 ]; then
        log_warning "Performance: Good (< 2 seconds)"
    else
        log_warning "Performance: Slow (>= 2 seconds)"
    fi

    echo
    log_success "Demo 4 completed - Performance testing finished"
    wait_for_user
}

# Demo 5: Error Handling
demo_error_handling() {
    log_demo "Demo 5: Error Handling & Edge Cases"
    echo
    log_info "Testing how the TUI handles various error conditions"
    echo

    log_info "Testing invalid test scenario..."
    echo "Command: $GEARBOX_BIN tui --test --test-scenario=invalid"
    echo

    if $GEARBOX_BIN tui --test --test-scenario=invalid 2>&1; then
        log_warning "Expected this to fail, but it succeeded"
    else
        log_success "Error handling works correctly - invalid scenario rejected"
    fi

    echo
    log_success "Demo 5 completed - Error handling verified"
    wait_for_user
}

# Demo 6: Integration with Regular Commands
demo_integration() {
    log_demo "Demo 6: Integration with Regular Gearbox Commands"
    echo
    log_info "The TUI complements the regular CLI commands"
    log_info "Let's see how they work together"
    echo

    log_info "Checking available commands:"
    echo "Command: $GEARBOX_BIN --help"
    echo
    $GEARBOX_BIN --help | head -20
    echo "... (truncated)"
    echo

    log_info "You can use regular commands alongside the TUI:"
    echo "  • gearbox list           - List available tools"
    echo "  • gearbox list bundles   - List available bundles"
    echo "  • gearbox doctor         - Run health checks"
    echo "  • gearbox config show    - Show configuration"
    echo "  • gearbox tui            - Launch TUI (normal mode)"
    echo "  • gearbox tui --demo     - Launch TUI (demo mode)"
    echo

    log_success "Demo 6 completed - CLI integration explained"
    wait_for_user
}

# Demo Summary
show_summary() {
    echo "🎯 TUI Testing & Demo Summary"
    echo "============================="
    echo

    log_success "All demos completed successfully!"
    echo

    log_info "🚀 Next Steps:"
    echo "  • Run automated tests: ./tests/tui-test-runner.sh"
    echo "  • Run Go integration tests: go test ./tests/"
    echo "  • Use TUI in production: gearbox tui"
    echo "  • Explore with demo mode: gearbox tui --demo"
    echo

    log_info "📖 For more information:"
    echo "  • Documentation: docs/USER_GUIDE.md"
    echo "  • TUI Help: gearbox tui --help"
    echo "  • General Help: gearbox --help"
    echo

    log_info "🐛 Found issues? Please report them!"
    echo
}

# Main execution
main() {
    check_prerequisites
    
    echo "This demo will guide you through testing Gearbox TUI functionality"
    echo "Each demo section can be skipped if desired"
    echo
    wait_for_user

    demo_help
    demo_test_scenarios
    demo_interactive
    demo_performance
    demo_error_handling
    demo_integration
    show_summary
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "TUI Testing & Demo Script for Gearbox"
        echo ""
        echo "Usage: $0 [demo-name]"
        echo ""
        echo "Available demos:"
        echo "  help         - Show TUI help"
        echo "  scenarios    - Run automated test scenarios"
        echo "  interactive  - Launch interactive demo"
        echo "  performance  - Test performance"
        echo "  errors       - Test error handling"
        echo "  integration  - Show CLI integration"
        echo ""
        echo "Examples:"
        echo "  $0              Run all demos"
        echo "  $0 interactive  Run only interactive demo"
        exit 0
        ;;
    help)
        check_prerequisites
        demo_help
        ;;
    scenarios)
        check_prerequisites
        demo_test_scenarios
        ;;
    interactive)
        check_prerequisites
        demo_interactive
        ;;
    performance)
        check_prerequisites
        demo_performance
        ;;
    errors)
        check_prerequisites
        demo_error_handling
        ;;
    integration)
        check_prerequisites
        demo_integration
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown demo: $1"
        echo "Run '$0 --help' for available options"
        exit 1
        ;;
esac