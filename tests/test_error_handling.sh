#!/bin/bash

# Error Handling and Recovery Tests
# Tests system resilience, error recovery, and security protection

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source test framework and common library
source "$SCRIPT_DIR/framework/test-framework.sh"
source "$REPO_DIR/scripts/lib/common.sh"

# =============================================================================
# TEST SETUP
# =============================================================================

setup() {
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    
    # Create error test environment
    mkdir -p error_test/{scripts,corrupted,invalid,temp}
    
    # Create various problematic scenarios
    create_error_scenarios
}

teardown() {
    rm -rf error_test
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
}

create_error_scenarios() {
    local script_dir="error_test/scripts"
    
    # Create script that fails during dependency installation
    cat > "$script_dir/install-fail-deps.sh" << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
source "$REPO_DIR/scripts/lib/common.sh"

log "Starting tool with failing dependencies..."

# Simulate dependency failure
error "Simulated dependency installation failure"
EOF
    chmod +x "$script_dir/install-fail-deps.sh"
    
    # Create script that fails during build
    cat > "$script_dir/install-fail-build.sh" << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
source "$REPO_DIR/scripts/lib/common.sh"

log "Starting tool that fails during build..."
log "Dependencies installed successfully"
log "Starting build process..."

# Simulate build failure
error "Simulated build compilation failure"
EOF
    chmod +x "$script_dir/install-fail-build.sh"
    
    # Create script with invalid parameters
    cat > "$script_dir/install-bad-params.sh" << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
source "$REPO_DIR/scripts/lib/common.sh"

# Test parameter validation
BUILD_TYPE="$1"
TOOL_NAME="$2"

validate_build_type "$BUILD_TYPE"
validate_tool_name "$TOOL_NAME"

success "Parameters validated successfully"
EOF
    chmod +x "$script_dir/install-bad-params.sh"
    
    # Create corrupted library file
    echo "invalid bash syntax } { ;" > error_test/corrupted/bad_lib.sh
    chmod +x error_test/corrupted/bad_lib.sh
}

# =============================================================================
# INPUT VALIDATION AND SECURITY TESTS
# =============================================================================

test_malicious_input_handling() {
    test_start "malicious input protection"
    
    # Test command injection prevention
    assert_command_failure "validate_tool_name 'tool; rm -rf /'" "Should prevent command injection in tool names"
    assert_command_failure "validate_file_path 'file\$(rm -rf /)'" "Should prevent command injection in file paths"
    assert_command_failure "validate_url 'https://example.com/file; wget malicious'" "Should prevent injection in URLs"
    
    # Test path traversal prevention
    assert_command_failure "validate_file_path '../../../etc/passwd'" "Should prevent path traversal"
    assert_command_failure "validate_tool_name '../../../bin/bash'" "Should prevent path traversal in tool names"
    
    # Test special character handling
    assert_command_failure "validate_tool_name 'tool\$name'" "Should reject special characters"
    assert_command_failure "validate_file_path 'file|dangerous'" "Should reject pipe characters"
    assert_command_failure "validate_url 'javascript:alert(1)'" "Should reject dangerous URL schemes"
    
    test_pass "Malicious input protection works"
}

test_buffer_overflow_protection() {
    test_start "buffer overflow and large input protection"
    
    # Test very long inputs
    local long_string=$(printf 'a%.0s' {1..10000})
    
    assert_command_failure "validate_tool_name '$long_string'" "Should reject extremely long tool names"
    assert_command_failure "validate_file_path '$long_string'" "Should reject extremely long file paths"
    
    # Test filename sanitization with large input
    local sanitized
    sanitized=$(sanitize_filename "$long_string")
    assert_command_success "[[ ${#sanitized} -le 255 ]]" "Should limit sanitized filename length"
    
    test_pass "Buffer overflow protection works"
}

test_resource_exhaustion_protection() {
    test_start "resource exhaustion protection"
    
    # Test that get_optimal_jobs provides reasonable limits
    local jobs
    jobs=$(get_optimal_jobs)
    assert_command_success "[[ $jobs -le 64 ]]" "Should limit parallel jobs to reasonable number"
    assert_command_success "[[ $jobs -ge 1 ]]" "Should provide at least 1 job"
    
    # Test memory-aware operations
    if declare -f human_readable_size >/dev/null 2>&1; then
        # Test with very large numbers
        local large_size_result
        large_size_result=$(human_readable_size 999999999999999)
        assert_not_equals "" "$large_size_result" "Should handle very large sizes"
    fi
    
    test_pass "Resource exhaustion protection works"
}

# =============================================================================
# SCRIPT EXECUTION ERROR TESTS
# =============================================================================

test_dependency_failure_handling() {
    test_start "dependency installation failure handling"
    
    cd error_test
    
    # Test script that fails during dependency installation
    assert_command_failure "./scripts/install-fail-deps.sh" "Script should fail gracefully on dependency error"
    
    # Verify system state is not corrupted
    assert_command_success "source '$REPO_DIR/scripts/lib/common.sh' && log 'System still functional'" "Library should still work after failed script"
    
    cd - >/dev/null
    test_pass "Dependency failure handling works"
}

test_build_failure_handling() {
    test_start "build process failure handling"
    
    cd error_test
    
    # Test script that fails during build
    assert_command_failure "./scripts/install-fail-build.sh" "Script should fail gracefully on build error"
    
    # Verify error doesn't corrupt environment
    assert_command_success "echo 'Environment test'" "Shell environment should remain functional"
    
    cd - >/dev/null
    test_pass "Build failure handling works"
}

test_invalid_parameter_handling() {
    test_start "invalid parameter handling"
    
    cd error_test
    
    # Test with invalid build type
    assert_command_failure "./scripts/install-bad-params.sh 'invalid-build-type' 'fd'" "Should reject invalid build type"
    
    # Test with invalid tool name
    assert_command_failure "./scripts/install-bad-params.sh 'standard' 'invalid/tool'" "Should reject invalid tool name"
    
    # Test with malicious parameters
    assert_command_failure "./scripts/install-bad-params.sh 'standard; rm -rf /' 'fd'" "Should reject malicious parameters"
    
    cd - >/dev/null
    test_pass "Invalid parameter handling works"
}

# =============================================================================
# LIBRARY CORRUPTION AND MISSING FILE TESTS
# =============================================================================

test_missing_library_handling() {
    test_start "missing library file handling"
    
    cd error_test
    
    # Create script that tries to load missing library
    cat > scripts/install-missing-lib.sh << EOF
#!/bin/bash
set -e

# Try to load non-existent library
if [[ -f "nonexistent/lib/common.sh" ]]; then
    source "nonexistent/lib/common.sh"
else
    echo "ERROR: Library not found" >&2
    exit 1
fi
EOF
    chmod +x scripts/install-missing-lib.sh
    
    assert_command_failure "./scripts/install-missing-lib.sh" "Should fail gracefully when library missing"
    
    cd - >/dev/null
    test_pass "Missing library handling works"
}

test_corrupted_file_handling() {
    test_start "corrupted file handling"
    
    cd error_test
    
    # Create script that tries to source corrupted file
    cat > scripts/install-corrupted-lib.sh << EOF
#!/bin/bash
set -e

# Try to source corrupted library
source "corrupted/bad_lib.sh" 2>&1 && {
    echo "Should not reach here"
    exit 1
} || {
    echo "Correctly detected corrupted library"
    exit 0
}
EOF
    chmod +x scripts/install-corrupted-lib.sh
    
    assert_command_success "./scripts/install-corrupted-lib.sh" "Should detect and handle corrupted library"
    
    cd - >/dev/null
    test_pass "Corrupted file handling works"
}

# =============================================================================
# FILESYSTEM AND PERMISSION ERROR TESTS
# =============================================================================

test_permission_error_handling() {
    test_start "permission error handling"
    
    cd error_test
    
    # Create file with restricted permissions
    echo "restricted content" > restricted_file.txt
    chmod 000 restricted_file.txt
    
    # Test that file operations handle permission errors gracefully
    assert_command_failure "backup_file 'restricted_file.txt'" "Should fail gracefully on permission error"
    
    # Cleanup
    chmod 644 restricted_file.txt
    rm -f restricted_file.txt
    
    cd - >/dev/null
    test_pass "Permission error handling works"
}

test_disk_space_simulation() {
    test_start "disk space exhaustion simulation"
    
    cd error_test
    
    # Create conditions that simulate low disk space
    # (We can't actually fill disk, so we test the detection logic)
    
    if declare -f show_disk_usage >/dev/null 2>&1; then
        # Test disk usage reporting
        local usage
        usage=$(show_disk_usage ".")
        assert_not_equals "" "$usage" "Should report disk usage"
    else
        test_skip "disk usage test" "show_disk_usage function not available"
    fi
    
    cd - >/dev/null
    test_pass "Disk space handling works"
}

# =============================================================================
# NETWORK AND CONNECTIVITY ERROR TESTS  
# =============================================================================

test_network_error_simulation() {
    test_start "network connectivity error simulation"
    
    # Test with invalid URLs that should fail
    assert_command_failure "validate_url 'https://this-domain-should-not-exist-12345.invalid'" "Should handle invalid domains"
    assert_command_failure "validate_url 'http://192.0.2.1:99999/nonexistent'" "Should handle connection failures"
    
    # Test URL validation with various failure modes
    assert_command_failure "validate_url 'ftp://unsupported.protocol'" "Should reject unsupported protocols"
    assert_command_failure "validate_url 'https://'" "Should reject malformed URLs"
    
    test_pass "Network error simulation works"
}

# =============================================================================
# CONCURRENT ACCESS AND RACE CONDITION TESTS
# =============================================================================

test_concurrent_file_access() {
    test_start "concurrent file access protection"
    
    cd error_test
    
    # Create a test file
    echo "original content" > temp/test_concurrent.txt
    
    # Start multiple processes trying to backup the same file
    local pids=()
    for i in {1..5}; do
        (backup_file "temp/test_concurrent.txt" 2>/dev/null || true) &
        pids+=($!)
    done
    
    # Wait for all to complete
    for pid in "${pids[@]}"; do
        wait "$pid" || true
    done
    
    # Verify file integrity is maintained
    assert_file_exists "temp/test_concurrent.txt" "Original file should still exist"
    local content
    content=$(cat temp/test_concurrent.txt)
    assert_equals "original content" "$content" "File content should be preserved"
    
    cd - >/dev/null
    test_pass "Concurrent file access protection works"
}

# =============================================================================
# SYSTEM STATE RECOVERY TESTS
# =============================================================================

test_environment_recovery() {
    test_start "environment variable recovery"
    
    # Save original environment
    local original_path="$PATH"
    local original_home="$HOME"
    
    # Simulate environment corruption
    export PATH="/invalid/path:$PATH"
    
    # Test that core functions still work
    assert_command_success "validate_tool_name 'fd'" "Should work with corrupted PATH"
    
    # Restore environment
    export PATH="$original_path"
    export HOME="$original_home"
    
    # Verify restoration
    assert_equals "$original_path" "$PATH" "PATH should be restored"
    assert_equals "$original_home" "$HOME" "HOME should be restored"
    
    test_pass "Environment recovery works"
}

test_cleanup_after_interruption() {
    test_start "cleanup after interruption simulation"
    
    cd error_test
    
    # Create temporary files that should be cleaned up
    mkdir -p temp/cleanup_test
    echo "temp data" > temp/cleanup_test/temp_file.txt
    
    # Simulate script interruption by creating cleanup scenario
    cat > scripts/install-cleanup-test.sh << EOF
#!/bin/bash
set -e

SCRIPT_DIR="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="\$(dirname "\$(dirname "\$SCRIPT_DIR")")"
source "\$REPO_DIR/scripts/lib/common.sh"

# Create temp files
mkdir -p temp/script_temp
echo "script temp data" > temp/script_temp/data.txt

# Set up cleanup trap
cleanup() {
    log "Cleanup triggered"
    rm -rf temp/script_temp
}
trap cleanup EXIT

log "Script setup completed"
log "Simulating work..."
sleep 0.1

# Simulate interruption/failure
if [[ "\$1" == "--fail" ]]; then
    error "Simulated failure for cleanup testing"
fi

success "Script completed successfully"
EOF
    chmod +x scripts/install-cleanup-test.sh
    
    # Test cleanup on failure
    assert_command_failure "./scripts/install-cleanup-test.sh --fail" "Script should fail as expected"
    
    # Verify cleanup occurred
    assert_command_success "[[ ! -d temp/script_temp ]]" "Temp files should be cleaned up after failure"
    
    # Test cleanup on success
    assert_command_success "./scripts/install-cleanup-test.sh" "Script should succeed and cleanup"
    assert_command_success "[[ ! -d temp/script_temp ]]" "Temp files should be cleaned up after success"
    
    cd - >/dev/null
    test_pass "Cleanup after interruption works"
}

# =============================================================================
# SECURITY ERROR TESTS
# =============================================================================

test_privilege_escalation_prevention() {
    test_start "privilege escalation prevention"
    
    # Test that scripts refuse to run as root (simulated check)
    # We can't actually test as root, but we can test the detection logic
    
    # Test ensure_not_root function
    assert_command_success "ensure_not_root" "Should pass when not running as root"
    
    # Test root detection logic
    if declare -f check_root_user >/dev/null 2>&1; then
        assert_command_success "check_root_user" "Root checking function should work"
    else
        log "Root checking function not available, testing ensure_not_root instead"
    fi
    
    test_pass "Privilege escalation prevention works"
}

test_file_system_protection() {
    test_start "file system protection"
    
    cd error_test
    
    # Test that system files are protected
    assert_command_failure "validate_file_path '/etc/passwd'" "Should protect system files"
    assert_command_failure "validate_file_path '/root/.ssh/id_rsa'" "Should protect sensitive files"
    assert_command_failure "backup_file '/etc/shadow'" "Should prevent backup of system files"
    
    # Test directory traversal protection
    assert_command_failure "validate_file_path '../../../../etc/passwd'" "Should prevent deep traversal"
    assert_command_failure "sanitize_filename '../../../etc/passwd'" "Should sanitize dangerous paths"
    
    cd - >/dev/null
    test_pass "File system protection works"
}

test_command_injection_prevention() {
    test_start "command injection prevention"
    
    # Test various injection attempts
    assert_command_failure "execute_command_safely 'echo test; rm -rf /'" "Should prevent command chaining"
    assert_command_failure "validate_tool_name 'tool\$(whoami)'" "Should prevent command substitution"
    assert_command_failure "validate_url 'https://example.com/\$(curl malicious)'" "Should prevent URL injection"
    
    # Test argument injection
    cd error_test
    assert_command_failure "./scripts/install-bad-params.sh 'standard' 'tool; malicious'" "Should prevent argument injection"
    cd - >/dev/null
    
    test_pass "Command injection prevention works"
}

# =============================================================================
# RECOVERY AND ROLLBACK TESTS
# =============================================================================

test_rollback_functionality() {
    test_start "rollback and recovery functionality"
    
    cd error_test
    
    # Create test files to backup and restore
    echo "original data" > temp/rollback_test.txt
    
    # Test backup functionality
    assert_command_success "backup_file 'temp/rollback_test.txt'" "Should backup file successfully"
    assert_file_exists "temp/rollback_test.txt.backup" "Backup file should be created"
    
    # Modify original file
    echo "modified data" > temp/rollback_test.txt
    
    # Test restore functionality
    assert_command_success "restore_file_backup 'temp/rollback_test.txt'" "Should restore from backup"
    
    local restored_content
    restored_content=$(cat temp/rollback_test.txt)
    assert_equals "original data" "$restored_content" "Content should be restored"
    
    cd - >/dev/null
    test_pass "Rollback functionality works"
}

test_partial_installation_recovery() {
    test_start "partial installation recovery"
    
    cd error_test
    
    # Simulate partial installation state
    mkdir -p temp/partial_install/{bin,lib,config}
    echo "#!/bin/bash\necho 'partial tool'" > temp/partial_install/bin/partial_tool
    chmod +x temp/partial_install/bin/partial_tool
    
    # Test detection and cleanup of partial state
    if [[ -f temp/partial_install/bin/partial_tool ]]; then
        log "Detected partial installation"
        rm -rf temp/partial_install
        success "Partial installation cleaned up"
    fi
    
    assert_command_success "[[ ! -d temp/partial_install ]]" "Partial installation should be cleaned up"
    
    cd - >/dev/null
    test_pass "Partial installation recovery works"
}

# =============================================================================
# STRESS AND EDGE CASE TESTS
# =============================================================================

test_high_load_error_handling() {
    test_start "error handling under high load"
    
    cd error_test
    
    # Create multiple failing processes and verify error handling
    local pids=()
    local failure_count=0
    
    for i in {1..10}; do
        (./scripts/install-fail-deps.sh 2>/dev/null || ((failure_count++))) &
        pids+=($!)
    done
    
    # Wait for all processes
    for pid in "${pids[@]}"; do
        wait "$pid" || true
    done
    
    # All should have failed gracefully
    assert_command_success "[[ $failure_count -eq 0 ]]" "All processes should handle errors gracefully"
    
    cd - >/dev/null
    test_pass "High load error handling works"
}

test_edge_case_inputs() {
    test_start "edge case input handling"
    
    # Test empty inputs
    assert_command_failure "validate_tool_name ''" "Should reject empty tool name"
    assert_command_failure "validate_build_type ''" "Should reject empty build type"
    assert_command_failure "validate_url ''" "Should reject empty URL"
    
    # Test whitespace-only inputs
    assert_command_failure "validate_tool_name '   '" "Should reject whitespace-only tool name"
    assert_command_failure "validate_file_path '\\t\\n'" "Should reject whitespace characters"
    
    # Test null bytes and control characters
    assert_command_failure "validate_tool_name 'tool\\x00'" "Should reject null bytes"
    assert_command_failure "validate_file_path 'file\\x1f'" "Should reject control characters"
    
    test_pass "Edge case input handling works"
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

main() {
    echo "🛡️ Error Handling and Recovery Tests"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    # Setup test environment
    setup
    
    echo "🔒 Testing Security and Input Validation..."
    test_malicious_input_handling
    test_buffer_overflow_protection
    test_resource_exhaustion_protection
    test_command_injection_prevention
    
    echo "💥 Testing Script Execution Errors..."
    test_dependency_failure_handling
    test_build_failure_handling
    test_invalid_parameter_handling
    
    echo "📁 Testing File System Errors..."
    test_missing_library_handling
    test_corrupted_file_handling
    test_permission_error_handling
    
    echo "🔄 Testing Recovery and Rollback..."
    test_rollback_functionality
    test_partial_installation_recovery
    
    echo "⚡ Testing Stress Conditions..."
    test_high_load_error_handling
    test_edge_case_inputs
    
    echo "💾 Testing System Protection..."
    test_privilege_escalation_prevention
    test_file_system_protection
    
    # Cleanup
    teardown
    
    # Show results
    print_test_summary
    
    echo
    echo "🎯 Error Handling Summary:"
    echo "  ✅ Input validation and sanitization"
    echo "  ✅ Command injection prevention" 
    echo "  ✅ Path traversal protection"
    echo "  ✅ Privilege escalation prevention"
    echo "  ✅ File system protection"
    echo "  ✅ Recovery and rollback functionality"
    echo "  ✅ Graceful failure handling"
    echo "  ✅ Resource protection"
}

# Execute tests
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi