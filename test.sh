#!/bin/bash

# Simple test runner for local development
# Provides quick validation of core functionality

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 Gearbox Quick Test Runner${NC}"
echo "================================"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -n "Testing $test_name... "
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

echo -e "\n${BLUE}📋 Basic Functionality Tests${NC}"
echo "----------------------------"

# Test basic script execution
run_test "gearbox help" "./gearbox help"
run_test "gearbox list" "./gearbox list"
run_test "lib/config.sh loading" "source lib/config.sh"
run_test "lib/common.sh loading" "source lib/common.sh"

# Test script executability
echo -e "\n${BLUE}🔧 Script Validation${NC}"
echo "-------------------"

for script in scripts/install-*.sh; do
    if [[ -f "$script" ]]; then
        script_name=$(basename "$script")
        run_test "$script_name executable" "[[ -x '$script' ]]"
        run_test "$script_name syntax" "bash -n '$script'"
        if [[ "$script_name" == "install-common-deps.sh" ]]; then
            run_test "$script_name help" "grep -q 'Usage:' '$script'"
        else
            run_test "$script_name help" "timeout 5 '$script' --help"
        fi
    fi
done

# Test Go tools if available
if command -v go >/dev/null 2>&1; then
    echo -e "\n${BLUE}🚀 Go Tools Tests${NC}"
    echo "-----------------"
    
    if [[ -d "tools/orchestrator" ]]; then
        run_test "orchestrator build" "cd tools/orchestrator && go build -o ../../bin/orchestrator"
        if [[ -f "bin/orchestrator" ]]; then
            run_test "orchestrator help" "./bin/orchestrator --help"
            run_test "orchestrator list" "./bin/orchestrator list"
        fi
    fi
    
    if [[ -d "tools/script-generator" ]]; then
        run_test "script-generator build" "cd tools/script-generator && go build -o ../../bin/script-generator"
        if [[ -f "bin/script-generator" ]]; then
            run_test "script-generator help" "./bin/script-generator --help"
            run_test "script-generator list" "./bin/script-generator list"
        fi
    fi
    
    if [[ -d "tools/config-manager" ]]; then
        run_test "config-manager build" "cd tools/config-manager && go build -o ../../bin/config-manager"
        if [[ -f "bin/config-manager" ]]; then
            run_test "config-manager help" "./bin/config-manager --help"
        fi
    fi
fi

# Test configuration
echo -e "\n${BLUE}⚙️  Configuration Tests${NC}"
echo "----------------------"

run_test "tools.json syntax" "python3 -m json.tool config/tools.json"
run_test "gearbox config show" "./gearbox config show"

# Test template generation if available
if [[ -f "bin/script-generator" ]]; then
    echo -e "\n${BLUE}📝 Template Tests${NC}"
    echo "----------------"
    
    run_test "template dry-run fd" "./bin/script-generator generate --dry-run --validate=false fd"
    run_test "template dry-run ripgrep" "./bin/script-generator generate --dry-run --validate=false ripgrep"
    run_test "template dry-run fzf" "./bin/script-generator generate --dry-run --validate=false fzf"
fi

# Security checks
echo -e "\n${BLUE}🔒 Security Checks${NC}"
echo "-----------------"

run_test "no dangerous eval" "! find scripts -name '*.sh' -exec grep -l 'eval.*\$[A-Z_]' {} \;"
run_test "no dangerous rm" "! find scripts -name '*.sh' -exec grep -l 'rm.*-rf.*/' {} \;"
run_test "no password mods" "! find scripts -name '*.sh' -exec grep -l 'sudo.*passwd' {} \;"

# Performance checks
if [[ -f "bin/orchestrator" ]]; then
    echo -e "\n${BLUE}⚡ Performance Checks${NC}"
    echo "-------------------"
    
    run_test "orchestrator startup" "timeout 5 ./bin/orchestrator list"
    
    if [[ -f "bin/script-generator" ]]; then
        run_test "script generation speed" "timeout 10 ./bin/script-generator generate --dry-run fd ripgrep fzf"
    fi
fi

# Summary
echo -e "\n${BLUE}📊 Test Summary${NC}"
echo "==============="
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ $FAILED_TESTS test(s) failed${NC}"
    exit 1
fi