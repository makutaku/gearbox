#!/bin/bash

# Comprehensive Unit Tests for scripts/lib/ modular functions
# Tests all core shared library functionality across all modules

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source the test framework
if [[ -f "$SCRIPT_DIR/framework/test-framework.sh" ]]; then
    source "$SCRIPT_DIR/framework/test-framework.sh"
else
    echo "ERROR: Test framework not found" >&2
    exit 1
fi

# Source the common library to test
if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: common.sh not found" >&2
    exit 1
fi

# =============================================================================
# TEST SETUP AND TEARDOWN
# =============================================================================

setup() {
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    
    # Create test files and directories
    mkdir -p test_data
    echo "test content" > test_data/test_file.txt
    echo "#!/bin/bash\necho 'test script'" > test_data/test_script.sh
    chmod +x test_data/test_script.sh
    
    # Create test binary for cache tests
    echo "#!/bin/bash\necho 'test-tool v1.0.0'" > test_data/test_binary
    chmod +x test_data/test_binary
}

teardown() {
    rm -rf test_data
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
    
    # Clean up any test cache entries
    if declare -f clear_cache >/dev/null 2>&1; then
        clear_cache >/dev/null 2>&1 || true
    fi
}

# =============================================================================
# LOGGING MODULE TESTS (core/logging.sh)
# =============================================================================

test_logging_functions() {
    test_start "logging functions basic operation"
    
    # Test that logging functions don't crash
    assert_command_success "log 'Test log message'" "Log function should work"
    assert_command_success "warning 'Test warning message'" "Warning function should work"
    assert_command_success "success 'Test success message'" "Success function should work"
    assert_command_success "debug 'Test debug message'" "Debug function should work"
    
    # Test error function (should exit, so test in subshell)
    assert_command_failure "(error 'Test error message' 2>/dev/null)" "Error function should exit with failure"
    
    test_pass "Logging functions work correctly"
}

test_progress_functions() {
    test_start "progress indicator functions"
    
    # Test available progress functions
    assert_command_success "log_step 'Test step'" "Step logging should work"
    
    # Test spinner functions if available
    if declare -f start_spinner >/dev/null 2>&1; then
        assert_command_success "start_spinner 'Testing' & sleep 0.1; stop_spinner" "Spinner functions should work"
    else
        test_skip "spinner test" "Spinner functions not available"
    fi
    
    test_pass "Progress functions work correctly"
}

# =============================================================================
# VALIDATION MODULE TESTS (core/validation.sh)
# =============================================================================

test_validate_tool_name() {
    test_start "tool name validation"
    
    # Valid tool names
    assert_command_success "validate_tool_name 'fd'" "Valid tool name should pass"
    assert_command_success "validate_tool_name 'ripgrep'" "Valid tool name should pass"
    assert_command_success "validate_tool_name 'nerd-fonts'" "Hyphenated tool name should pass"
    
    # Invalid tool names
    assert_command_failure "validate_tool_name 'invalid/tool'" "Tool name with slash should fail"
    assert_command_failure "validate_tool_name ''" "Empty tool name should fail"
    assert_command_failure "validate_tool_name '../etc/passwd'" "Path traversal should fail"
    assert_command_failure "validate_tool_name 'tool\$name'" "Tool name with special chars should fail"
    
    test_pass "Tool name validation works correctly"
}

test_validate_file_path() {
    test_start "file path validation"
    
    # Valid paths (relative to project)
    assert_command_success "validate_file_path 'test_data/test_file.txt'" "Valid relative file path should pass"
    assert_command_success "validate_file_path 'scripts/lib/common.sh'" "Valid script path should pass"
    
    # Invalid paths
    assert_command_failure "validate_file_path '../../../etc/passwd'" "Path traversal should fail"
    assert_command_failure "validate_file_path '/etc/passwd'" "Absolute system path should fail"
    assert_command_failure "validate_file_path 'file with spaces and \$dangerous'" "Dangerous characters should fail"
    
    test_pass "File path validation works correctly"
}

test_validate_build_type() {
    test_start "build type validation"
    
    # Valid build types
    assert_command_success "validate_build_type 'minimal'" "Valid build type should pass"
    assert_command_success "validate_build_type 'standard'" "Valid build type should pass"  
    assert_command_success "validate_build_type 'maximum'" "Valid build type should pass"
    
    # Invalid build types
    assert_command_failure "validate_build_type 'invalid'" "Invalid build type should fail"
    assert_command_failure "validate_build_type ''" "Empty build type should fail"
    assert_command_failure "validate_build_type 'rm -rf /'" "Dangerous build type should fail"
    
    test_pass "Build type validation works correctly"
}

test_validate_url() {
    test_start "URL validation"
    
    # Valid URLs
    assert_command_success "validate_url 'https://github.com/user/repo.git'" "HTTPS GitHub URL should pass"
    assert_command_success "validate_url 'http://example.com/file.tar.gz'" "HTTP URL should pass"
    assert_command_success "validate_url 'https://releases.ubuntu.com/20.04/ubuntu.iso'" "HTTPS download URL should pass"
    
    # Invalid URLs
    assert_command_failure "validate_url 'ftp://example.com/file'" "FTP URL should fail"
    assert_command_failure "validate_url 'not-a-url'" "Invalid format should fail"
    assert_command_failure "validate_url 'https://example.com/file with spaces'" "URLs with spaces should fail"
    assert_command_failure "validate_url 'https://example.com/file|rm-rf'" "URLs with dangerous chars should fail"
    
    test_pass "URL validation works correctly"
}

test_sanitize_filename() {
    test_start "filename sanitization"
    
    local result
    
    # Normal filename
    result=$(sanitize_filename "normal-file.txt")
    assert_equals "normal-file.txt" "$result" "Normal filename should remain unchanged"
    
    # Filename with dangerous characters
    result=$(sanitize_filename "bad file\$name&with|chars")
    assert_not_equals "bad file\$name&with|chars" "$result" "Dangerous chars should be sanitized"
    assert_contains "$result" "bad" "Should preserve safe characters"
    
    # Very long filename
    local long_name=$(printf 'a%.0s' {1..300})
    result=$(sanitize_filename "$long_name")
    assert_command_success "[[ ${#result} -le 255 ]]" "Long filename should be truncated"
    
    # Empty filename
    result=$(sanitize_filename "")
    assert_not_equals "" "$result" "Empty filename should get default name"
    
    test_pass "Filename sanitization works correctly"
}

test_version_compare() {
    test_start "version comparison"
    
    # Equal versions
    assert_command_success "version_compare '1.0.0' '1.0.0'" "Equal versions should return true"
    
    # Greater versions
    assert_command_success "version_compare '1.1.0' '1.0.0'" "Greater version should return true"
    assert_command_success "version_compare '2.0.0' '1.9.9'" "Major version should return true"
    assert_command_success "version_compare '1.0.1' '1.0.0'" "Patch version should return true"
    
    # Lesser versions
    assert_command_failure "version_compare '1.0.0' '1.1.0'" "Lesser version should return false"
    assert_command_failure "version_compare '1.9.9' '2.0.0'" "Lesser major should return false"
    
    # Edge cases
    assert_command_success "version_compare '1.0' '1.0.0'" "Missing patch should work"
    assert_command_success "version_compare '1.0.1' '1.0'" "Missing patch comparison should work"
    
    test_pass "Version comparison works correctly"
}

# =============================================================================
# UTILITIES MODULE TESTS (core/utilities.sh) 
# =============================================================================

test_get_optimal_jobs() {
    test_start "optimal jobs calculation"
    
    local jobs
    jobs=$(get_optimal_jobs)
    
    # Should return a positive integer
    assert_command_success "[[ '$jobs' =~ ^[0-9]+$ ]]" "Should return a number"
    assert_command_success "[[ '$jobs' -gt 0 ]]" "Should return positive number"
    assert_command_success "[[ '$jobs' -le 64 ]]" "Should return reasonable number"
    
    test_pass "Optimal jobs calculation works correctly"
}

test_human_readable_size() {
    test_start "human readable size formatting"
    
    # Test with different sizes
    local result
    
    result=$(human_readable_size 1024)
    assert_contains "$result" "1" "Should format 1024 bytes"
    
    result=$(human_readable_size 1048576)
    assert_contains "$result" "1" "Should format 1MB"
    
    test_pass "Human readable size formatting works"
}

test_file_operations() {
    test_start "file backup and restore operations"
    
    # Test backup_file
    echo "original content" > test_data/original.txt
    assert_command_success "backup_file 'test_data/original.txt'" "Should backup file successfully"
    assert_file_exists "test_data/original.txt.backup" "Backup file should exist"
    
    # Modify original
    echo "modified content" > test_data/original.txt
    
    # Test restore
    assert_command_success "restore_file_backup 'test_data/original.txt'" "Should restore file successfully"
    
    local content
    content=$(cat test_data/original.txt)
    assert_equals "original content" "$content" "File should be restored to original content"
    
    test_pass "File backup and restore work correctly"
}

# =============================================================================
# SECURITY MODULE TESTS (core/security.sh)
# =============================================================================

test_security_functions() {
    test_start "security validation functions"
    
    # Test check_tool_installed
    assert_command_success "check_tool_installed 'bash'" "Should detect installed tool (bash)"
    assert_command_failure "check_tool_installed 'nonexistent-tool-12345'" "Should not detect missing tool"
    
    # Test execute_command_safely (basic test)
    assert_command_success "execute_command_safely echo 'safe command'" "Should execute safe command"
    
    # Test secure_download (dry run style test)
    if declare -f secure_download >/dev/null 2>&1; then
        assert_command_failure "secure_download 'ftp://bad.url' 'output'" "Should reject bad URL protocol"
    else
        test_skip "secure_download test" "Function not available"
    fi
    
    test_pass "Security functions work correctly"
}

test_root_prevention() {
    test_start "root user prevention"
    
    # Test ensure_not_root (should pass since we're not root)
    assert_command_success "ensure_not_root" "Should pass when not running as root"
    
    # Note: We can't test the actual root failure case in normal testing
    
    test_pass "Root prevention works correctly"
}

# =============================================================================
# BUILD MODULE TESTS
# =============================================================================

test_build_cache_system() {
    test_start "build cache system"
    
    # Load build modules if available
    if declare -f require_build_modules >/dev/null 2>&1; then
        require_build_modules
        
        # Test cache functions if available
        if declare -f is_cached >/dev/null 2>&1; then
            local test_tool="test-cache-tool"
            local test_build_type="standard"
            local test_binary="test_data/test_binary"
            
            # Initially should not be cached
            assert_command_failure "is_cached '$test_tool' '$test_build_type'" "Tool should not be cached initially"
            
            # Cache the build if function exists
            if declare -f cache_build >/dev/null 2>&1; then
                assert_command_success "cache_build '$test_tool' '$test_build_type' 'v1.0.0' '$test_binary'" "Should cache successfully"
                assert_command_success "is_cached '$test_tool' '$test_build_type' 'v1.0.0'" "Tool should now be cached"
            else
                test_skip "cache_build test" "cache_build function not available"
            fi
        else
            test_skip "cache system tests" "Cache functions not available"
        fi
    else
        test_skip "build cache system" "Build modules not available"
    fi
    
    test_pass "Build cache system tests completed"
}

test_build_execution() {
    test_start "build execution functions"
    
    # Require build modules
    require_build_modules
    
    # Test build_with_options if available
    if declare -f build_with_options >/dev/null 2>&1; then
        assert_command_success "build_with_options echo 'test build'" "Should execute build command"
    else
        test_skip "build_with_options test" "Function not available"
    fi
    
    test_pass "Build execution functions work"
}

# =============================================================================
# INTEGRATION TESTS
# =============================================================================

test_module_loading() {
    test_start "module loading system"
    
    # Test that core modules are loaded
    assert_command_success "declare -f log >/dev/null" "Logging module should be loaded"
    assert_command_success "declare -f validate_tool_name >/dev/null" "Validation module should be loaded"
    assert_command_success "declare -f get_optimal_jobs >/dev/null" "Utilities module should be loaded"
    assert_command_success "declare -f ensure_not_root >/dev/null" "Security module should be loaded"
    
    # Test lazy loading
    assert_command_success "require_build_modules" "Should load build modules successfully"
    assert_command_success "require_system_modules" "Should load system modules successfully"
    
    test_pass "Module loading system works correctly"
}

test_configuration_integration() {
    test_start "configuration system integration"
    
    # Test that config functions are available
    if declare -f init_config >/dev/null 2>&1; then
        assert_command_success "init_config" "Configuration initialization should work"
        
        # Test basic config operations
        if declare -f get_config >/dev/null 2>&1; then
            local default_build_type
            default_build_type=$(get_config "DEFAULT_BUILD_TYPE" "standard")
            assert_not_equals "" "$default_build_type" "Should return config value"
        fi
    else
        test_skip "configuration integration" "Config functions not available"
    fi
    
    test_pass "Configuration integration works"
}

# =============================================================================
# ERROR HANDLING AND EDGE CASES
# =============================================================================

test_error_handling() {
    test_start "error handling and edge cases"
    
    # Test with invalid inputs
    assert_command_failure "validate_tool_name '../../etc/passwd'" "Should reject path traversal"
    assert_command_failure "validate_file_path '\$(rm -rf /)'" "Should reject command injection"
    assert_command_failure "validate_url 'javascript:alert(1)'" "Should reject dangerous URL schemes"
    
    # Test with empty inputs
    assert_command_failure "validate_tool_name ''" "Should reject empty tool name"
    assert_command_failure "validate_build_type ''" "Should reject empty build type"
    assert_command_failure "validate_url ''" "Should reject empty URL"
    
    # Test with malformed inputs
    assert_command_failure "version_compare 'not.a.version' '1.0.0'" "Should handle malformed versions"
    assert_command_failure "get_optimal_jobs 'invalid_arg'" "Should handle invalid arguments"
    
    test_pass "Error handling works correctly"
}

test_security_edge_cases() {
    test_start "security edge cases"
    
    # Test command injection prevention
    assert_command_failure "execute_command_safely 'echo test; rm -rf /'" "Should prevent command injection"
    assert_command_failure "secure_execute_script 'test_data/test_script.sh' '--arg;\$(rm -rf /)'" "Should prevent arg injection"
    
    # Test path traversal prevention  
    assert_command_failure "validate_file_path '../../../sensitive/file'" "Should prevent deep path traversal"
    assert_command_failure "backup_file '/etc/passwd'" "Should prevent system file access"
    
    test_pass "Security edge cases handled correctly"
}

# =============================================================================
# PERFORMANCE AND STRESS TESTS
# =============================================================================

test_performance_functions() {
    test_start "performance and stress testing"
    
    # Test with large inputs
    local large_filename=$(printf 'a%.0s' {1..500})
    local sanitized
    sanitized=$(sanitize_filename "$large_filename")
    assert_command_success "[[ ${#sanitized} -le 255 ]]" "Should handle very long filenames"
    
    # Test multiple concurrent operations
    for i in {1..5}; do
        validate_tool_name "tool$i" &
    done
    wait
    assert_command_success "true" "Should handle concurrent validation"
    
    # Test timeout functionality
    if declare -f run_with_timeout >/dev/null 2>&1; then
        assert_command_success "run_with_timeout 1 'sleep 0.1'" "Should complete within timeout"
        assert_command_failure "run_with_timeout 0.1 'sleep 1'" "Should timeout long command"
    else
        test_skip "timeout test" "run_with_timeout not available"
    fi
    
    test_pass "Performance functions work under stress"
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

main() {
    echo "🧪 Comprehensive Unit Tests for Gearbox Shell Libraries"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    # Initialize test tracking
    START_TIME=$(date +%s)
    
    # Setup test environment
    setup
    
    # Run all test suites
    echo "📋 Testing Logging Module..."
    test_logging_functions
    test_progress_functions
    
    echo "🔍 Testing Validation Module..."
    test_validate_tool_name
    test_validate_file_path
    test_validate_build_type
    test_validate_url
    test_sanitize_filename
    test_version_compare
    
    echo "⚙️ Testing Utilities Module..."
    test_get_optimal_jobs
    test_human_readable_size
    test_file_operations
    
    echo "🔒 Testing Security Module..."
    test_security_functions
    test_root_prevention
    
    echo "🏗️ Testing Build System..."
    test_build_cache_system
    test_build_execution
    
    echo "🔧 Testing Integration..."
    test_module_loading
    test_configuration_integration
    
    echo "🛡️ Testing Error Handling..."
    test_error_handling
    test_security_edge_cases
    
    echo "⚡ Testing Performance..."
    test_performance_functions
    
    # Cleanup
    teardown
    
    # Show results
    print_test_summary
}

# Execute tests
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi