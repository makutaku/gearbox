#!/bin/bash

# Integration Tests for Tool Installation Process
# Tests end-to-end tool installation workflows

# Test setup
setup() {
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    
    # Create isolated test environment
    mkdir -p test_env/{scripts,config,lib,tools}
    
    # Copy essential files
    cp -r "$REPO_DIR/scripts/lib" test_env/
    cp -r "$REPO_DIR/config" test_env/
    cp "$REPO_DIR/build/gearbox" test_env/
    
    # Create minimal test scripts for faster execution
    create_test_scripts
}

# Test teardown
teardown() {
    rm -rf test_env
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
}

# Create minimal test installation scripts
create_test_scripts() {
    cat > test_env/scripts/install-test-tool.sh << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

if [[ -f "$REPO_DIR/scripts/lib/common.sh" ]]; then
    source "$REPO_DIR/scripts/lib/common.sh"
else
    echo "ERROR: common.sh not found" >&2
    exit 1
fi

# Minimal test tool configuration
TOOL_NAME="test-tool"
BUILD_TYPE="${BUILD_TYPE:-standard}"
MODE="${MODE:-install}"
SKIP_DEPS=false

# Parse basic arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --config-only)
            MODE="config"
            shift
            ;;
        --build-only)
            MODE="build"
            shift
            ;;
        --help)
            echo "Test tool installation script"
            echo "Usage: $0 [--skip-deps] [--config-only] [--build-only]"
            exit 0
            ;;
        *)
            shift
            ;;
    esac
done

log "Starting $TOOL_NAME installation (mode: $MODE)"

# Simulate dependency installation
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies..."
    sleep 0.1  # Simulate work
    success "Dependencies installed"
fi

# Exit if config-only
if [[ "$MODE" == "config" ]]; then
    success "Configuration completed"
    exit 0
fi

# Simulate build process
log "Building $TOOL_NAME..."
sleep 0.2  # Simulate build time
success "Build completed"

# Exit if build-only
if [[ "$MODE" == "build" ]]; then
    success "Build completed"
    exit 0
fi

# Simulate installation
log "Installing $TOOL_NAME..."
mkdir -p "$HOME/.local/bin"
echo '#!/bin/bash\necho "test-tool v1.0.0"' > "$HOME/.local/bin/test-tool"
chmod +x "$HOME/.local/bin/test-tool"
success "Installation completed"

# Verify installation
if command -v test-tool >/dev/null 2>&1; then
    success "$TOOL_NAME installation verified"
else
    export PATH="$HOME/.local/bin:$PATH"
    if command -v test-tool >/dev/null 2>&1; then
        success "$TOOL_NAME installation verified (PATH updated)"
    else
        error "$TOOL_NAME installation verification failed"
    fi
fi
EOF

    chmod +x test_env/scripts/install-test-tool.sh
}

# Test basic script execution
test_script_execution() {
    cd test_env
    
    # Test help flag
    assert_command_success "./scripts/install-test-tool.sh --help" "Help flag should work"
    
    # Test config-only mode
    assert_command_success "./scripts/install-test-tool.sh --config-only" "Config-only mode should work"
    
    # Test build-only mode
    assert_command_success "./scripts/install-test-tool.sh --build-only" "Build-only mode should work"
    
    cd - >/dev/null
}

# Test full installation process
test_full_installation() {
    cd test_env
    
    # Test full installation
    assert_command_success "./scripts/install-test-tool.sh" "Full installation should succeed"
    
    # Verify tool is available
    export PATH="$HOME/.local/bin:$PATH"
    assert_command_success "command -v test-tool" "Installed tool should be in PATH"
    
    # Test tool execution
    local output
    output=$(test-tool 2>/dev/null || echo "failed")
    assert_contains "$output" "test-tool v1.0.0" "Tool should return version"
    
    cd - >/dev/null
}

# Test dependency handling
test_dependency_handling() {
    cd test_env
    
    # Test with dependency skipping
    assert_command_success "./scripts/install-test-tool.sh --skip-deps" "Should handle dependency skipping"
    
    cd - >/dev/null
}

# Test error conditions
test_error_conditions() {
    cd test_env
    
    # Test with missing library
    mv scripts/lib/common.sh scripts/lib/common.sh.backup
    assert_command_failure "./scripts/install-test-tool.sh" "Should fail when common.sh missing"
    mv scripts/lib/common.sh.backup scripts/lib/common.sh
    
    cd - >/dev/null
}

# Test orchestrator integration
test_orchestrator_integration() {
    cd test_env
    
    # Test if orchestrator binary exists
    if [[ -f "$REPO_DIR/bin/orchestrator" ]]; then
        # Copy orchestrator for testing
        mkdir -p bin
        cp "$REPO_DIR/bin/orchestrator" bin/
        
        # Test orchestrator list command
        assert_command_success "./bin/orchestrator list" "Orchestrator list should work"
        
        # Test orchestrator help
        assert_command_success "./bin/orchestrator --help" "Orchestrator help should work"
    else
        test_skip "test_orchestrator_integration" "Orchestrator binary not found"
    fi
    
    cd - >/dev/null
}

# Test script generator integration
test_script_generator_integration() {
    cd test_env
    
    # Test if script generator exists
    if [[ -f "$REPO_DIR/bin/script-generator" ]]; then
        # Copy necessary files
        mkdir -p bin templates
        cp "$REPO_DIR/bin/script-generator" bin/
        cp -r "$REPO_DIR/templates/"* templates/
        
        # Test script generator list
        assert_command_success "./bin/script-generator list" "Script generator list should work"
        
        # Test dry-run generation
        assert_command_success "./bin/script-generator generate --dry-run fd" "Dry-run generation should work"
    else
        test_skip "test_script_generator_integration" "Script generator binary not found"
    fi
    
    cd - >/dev/null
}

# Test main gearbox CLI
test_gearbox_cli() {
    cd test_env
    
    # Test help command
    assert_command_success "./gearbox help" "Gearbox help should work"
    
    # Test list command
    assert_command_success "./gearbox list" "Gearbox list should work"
    
    # Test config command
    if [[ -f scripts/lib/config.sh ]]; then
        assert_command_success "./gearbox config show" "Config show should work"
    else
        test_skip "test_gearbox_cli::config" "Config system not available"
    fi
    
    cd - >/dev/null
}

# Test concurrent installation simulation
test_concurrent_installs() {
    cd test_env
    
    # Create multiple test scripts
    for i in {1..3}; do
        cp scripts/install-test-tool.sh scripts/install-test-tool-$i.sh
        sed -i "s/test-tool/test-tool-$i/g" scripts/install-test-tool-$i.sh
    done
    
    # Run multiple installations concurrently
    local pids=()
    for i in {1..3}; do
        ./scripts/install-test-tool-$i.sh &
        pids+=($!)
    done
    
    # Wait for all to complete
    local all_success=true
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            all_success=false
        fi
    done
    
    if [[ "$all_success" == true ]]; then
        assert_command_success "true" "Concurrent installations should succeed"
    else
        assert_command_success "false" "Concurrent installations failed"
    fi
    
    cd - >/dev/null
}

# Test configuration system integration
test_configuration_system() {
    cd test_env
    
    # Test JSON configuration loading
    if [[ -f config/tools.json ]]; then
        # Verify JSON is valid
        assert_command_success "python3 -m json.tool config/tools.json >/dev/null" "tools.json should be valid JSON"
        
        # Test config manager if available
        if [[ -f "$REPO_DIR/bin/config-manager" ]]; then
            mkdir -p bin
            cp "$REPO_DIR/bin/config-manager" bin/
            assert_command_success "./bin/config-manager list" "Config manager should work"
        fi
    else
        test_skip "test_configuration_system" "tools.json not found"
    fi
    
    cd - >/dev/null
}

# Test build cache functionality
test_build_cache() {
    cd test_env
    
    # Test cache directory creation
    local cache_dir="$HOME/.cache/gearbox"
    mkdir -p "$cache_dir"
    
    # Test cache functions from common.sh
    export PATH="$PATH:$PWD/scripts"
    
    # Install a tool to populate cache
    ./scripts/install-test-tool.sh
    
    # Verify cache operations work (basic functionality)
    assert_command_success "source scripts/lib/common.sh && cache_build 'test-tool' 'standard' '$HOME/.local/bin/test-tool'" "Cache build should work"
    assert_command_success "source scripts/lib/common.sh && is_cached 'test-tool' 'standard'" "Cache check should work"
    
    cd - >/dev/null
}

# Test health check system
test_health_checks() {
    cd test_env
    
    # Test doctor functionality if available
    if [[ -f scripts/lib/doctor.sh ]]; then
        # Source doctor functions and test basic health checks
        assert_command_success "source scripts/lib/doctor.sh && check_system_requirements" "System requirements check should work"
        assert_command_success "source scripts/lib/doctor.sh && check_disk_space" "Disk space check should work"
    else
        test_skip "test_health_checks" "Doctor system not available"
    fi
    
    cd - >/dev/null
}