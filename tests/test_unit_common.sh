#!/bin/bash

# Unit Tests for lib/common.sh functions
# Tests core shared library functionality

# Test setup
setup() {
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    
    # Create test files and directories
    mkdir -p test_data
    echo "test content" > test_data/test_file.txt
    echo "#!/bin/bash\necho 'test script'" > test_data/test_script.sh
    chmod +x test_data/test_script.sh
}

# Test teardown
teardown() {
    rm -rf test_data
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
}

# Test validate_tool_name function
test_validate_tool_name() {
    # Valid tool names
    assert_command_success "validate_tool_name 'fd'" "Valid tool name should pass"
    assert_command_success "validate_tool_name 'ripgrep'" "Valid tool name should pass"
    assert_command_success "validate_tool_name 'fzf'" "Valid tool name should pass"
    
    # Invalid tool names
    assert_command_failure "validate_tool_name 'invalid-tool'" "Invalid tool name should fail"
    assert_command_failure "validate_tool_name ''" "Empty tool name should fail"
    assert_command_failure "validate_tool_name '../etc/passwd'" "Path traversal should fail"
}

# Test validate_file_path function
test_validate_file_path() {
    # Valid paths
    assert_command_success "validate_file_path 'test_data/test_file.txt'" "Valid file path should pass"
    assert_command_success "validate_file_path 'scripts/install-fd.sh'" "Valid script path should pass"
    
    # Invalid paths
    assert_command_failure "validate_file_path '../../../etc/passwd'" "Path traversal should fail"
    assert_command_failure "validate_file_path '/etc/passwd'" "Absolute system path should fail"
    assert_command_failure "validate_file_path 'file with spaces and \$dangerous'" "Dangerous characters should fail"
}

# Test validate_build_type function
test_validate_build_type() {
    # Valid build types
    assert_command_success "validate_build_type 'minimal'" "Valid build type should pass"
    assert_command_success "validate_build_type 'standard'" "Valid build type should pass"
    assert_command_success "validate_build_type 'maximum'" "Valid build type should pass"
    
    # Invalid build types
    assert_command_failure "validate_build_type 'invalid'" "Invalid build type should fail"
    assert_command_failure "validate_build_type ''" "Empty build type should fail"
    assert_command_failure "validate_build_type 'rm -rf /'" "Dangerous build type should fail"
}

# Test get_optimal_jobs function
test_get_optimal_jobs() {
    local jobs
    jobs=$(get_optimal_jobs)
    
    # Should return a positive integer
    assert_command_success "[[ '$jobs' =~ ^[0-9]+$ ]]" "Should return a number"
    assert_command_success "[[ '$jobs' -gt 0 ]]" "Should return positive number"
    assert_command_success "[[ '$jobs' -le 32 ]]" "Should return reasonable number"
}

# Test version comparison function
test_version_compare() {
    # Source the version_compare function from common scripts
    source "$REPO_DIR/scripts/install-fd.sh" 2>/dev/null || true
    
    if declare -f version_compare >/dev/null; then
        # Equal versions
        assert_command_success "version_compare '1.0.0' '1.0.0'" "Equal versions should return true"
        
        # Greater versions
        assert_command_success "version_compare '1.1.0' '1.0.0'" "Greater version should return true"
        assert_command_success "version_compare '2.0.0' '1.9.9'" "Major version should return true"
        
        # Lesser versions
        assert_command_failure "version_compare '1.0.0' '1.1.0'" "Lesser version should return false"
        assert_command_failure "version_compare '1.9.9' '2.0.0'" "Lesser major should return false"
        
        # Edge cases
        assert_command_success "version_compare '1.0' '1.0.0'" "Missing patch should work"
        assert_command_success "version_compare '1.0.1' '1.0'" "Missing patch comparison should work"
    else
        test_skip "test_version_compare" "version_compare function not available"
    fi
}

# Test secure_execute_script function
test_secure_execute_script() {
    # Valid script execution
    assert_command_success "secure_execute_script 'test_data/test_script.sh'" "Valid script should execute"
    
    # Invalid script paths
    assert_command_failure "secure_execute_script '/nonexistent/script.sh'" "Non-existent script should fail"
    assert_command_failure "secure_execute_script 'test_data/test_file.txt'" "Non-executable file should fail"
    
    # Test argument validation
    assert_command_failure "secure_execute_script 'test_data/test_script.sh' '--bad;arg'" "Dangerous args should fail"
    assert_command_failure "secure_execute_script 'test_data/test_script.sh' '\$(rm -rf /)'" "Command injection should fail"
}

# Test validate_url function
test_validate_url() {
    # Valid URLs
    assert_command_success "validate_url 'https://github.com/user/repo.git'" "HTTPS URL should pass"
    assert_command_success "validate_url 'http://example.com/file.tar.gz'" "HTTP URL should pass"
    
    # Invalid URLs
    assert_command_failure "validate_url 'ftp://example.com/file'" "FTP URL should fail"
    assert_command_failure "validate_url 'not-a-url'" "Invalid format should fail"
    assert_command_failure "validate_url 'https://example.com/file with spaces'" "Spaces should fail"
    assert_command_failure "validate_url 'https://example.com/file|rm-rf'" "Dangerous chars should fail"
}

# Test sanitize_filename function
test_sanitize_filename() {
    local result
    
    # Normal filename
    result=$(sanitize_filename "normal-file.txt")
    assert_equals "normal-file.txt" "$result" "Normal filename should remain unchanged"
    
    # Filename with dangerous characters
    result=$(sanitize_filename "bad file\$name&with|chars")
    assert_not_equals "bad file\$name&with|chars" "$result" "Dangerous chars should be sanitized"
    assert_contains "$result" "bad" "Should preserve safe characters"
    
    # Very long filename
    local long_name=$(printf 'a%.0s' {1..200})
    result=$(sanitize_filename "$long_name")
    assert_command_success "[[ ${#result} -le 100 ]]" "Long filename should be truncated"
    
    # Empty filename
    result=$(sanitize_filename "")
    assert_not_equals "" "$result" "Empty filename should get default name"
}

# Test build cache functions
test_cache_functions() {
    local test_tool="test-tool"
    local test_build_type="standard"
    local test_binary="test_data/test_script.sh"
    
    # Initially should not be cached
    assert_command_failure "is_cached '$test_tool' '$test_build_type'" "Tool should not be cached initially"
    
    # Cache the build
    assert_command_success "cache_build '$test_tool' '$test_build_type' '$test_binary'" "Should cache successfully"
    
    # Now should be cached
    assert_command_success "is_cached '$test_tool' '$test_build_type'" "Tool should now be cached"
    
    # Should be able to retrieve cached binary
    local temp_target="test_data/cached_binary"
    assert_command_success "get_cached_binary 'test_script.sh' '$test_build_type' '$temp_target'" "Should retrieve cached binary"
    assert_file_exists "$temp_target" "Cached binary should be copied to target"
}

# Test error handling and logging
test_logging_functions() {
    # Test that logging functions don't crash
    assert_command_success "log 'Test log message'" "Log function should work"
    assert_command_success "warning 'Test warning message'" "Warning function should work"
    assert_command_success "success 'Test success message'" "Success function should work"
    assert_command_success "debug 'Test debug message'" "Debug function should work"
    
    # Test error function (should exit, so test in subshell)
    assert_command_failure "(error 'Test error message')" "Error function should exit with failure"
}

# Test configuration loading
test_configuration_loading() {
    # Test that configuration files can be loaded
    if [[ -f "$REPO_DIR/config.sh" ]]; then
        assert_command_success "source '$REPO_DIR/config.sh'" "Config should load without errors"
    else
        test_skip "test_configuration_loading" "config.sh not found"
    fi
    
    # Test lib/config.sh loading
    if [[ -f "$REPO_DIR/lib/config.sh" ]]; then
        assert_command_success "source '$REPO_DIR/lib/config.sh'" "lib/config.sh should load without errors"
    else
        test_skip "test_configuration_loading" "lib/config.sh not found"
    fi
}

# Test parallel job calculation with different scenarios
test_get_optimal_jobs_scenarios() {
    # Mock different CPU counts and test behavior
    local original_nproc_output
    
    # Test with different simulated CPU counts
    for cpu_count in 1 2 4 8 16 32; do
        # Create a mock nproc command
        echo "#!/bin/bash\necho $cpu_count" > test_data/mock_nproc
        chmod +x test_data/mock_nproc
        
        # Temporarily modify PATH to use our mock
        local old_path="$PATH"
        export PATH="$(pwd)/test_data:$PATH"
        
        local jobs
        jobs=$(get_optimal_jobs)
        
        # Restore PATH
        export PATH="$old_path"
        
        # Verify reasonable job count
        assert_command_success "[[ '$jobs' -gt 0 ]]" "Should return positive jobs for $cpu_count CPUs"
        assert_command_success "[[ '$jobs' -le '$((cpu_count + 2))' ]]" "Should not exceed reasonable limit for $cpu_count CPUs"
    done
    
    rm -f test_data/mock_nproc
}