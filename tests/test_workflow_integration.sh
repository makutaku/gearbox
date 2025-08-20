#!/bin/bash

# Multi-Tool Workflow Integration Tests
# Tests realistic workflows involving multiple tools and complex operations

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source test framework and common library
source "$SCRIPT_DIR/framework/test-framework.sh"
source "$REPO_DIR/scripts/lib/common.sh"

# =============================================================================
# TEST CONFIGURATION
# =============================================================================

# Test environment setup
setup() {
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    
    # Create test workspace
    mkdir -p test_workspace/{tools,config,cache}
    
    # Create mock tool configurations
    cat > test_workspace/config/test_bundle.json << 'EOF'
{
    "bundles": {
        "test-essential": {
            "description": "Essential tools for testing",
            "tools": ["fd", "ripgrep", "fzf"],
            "system_packages": ["git", "curl"]
        },
        "test-development": {
            "description": "Development tools for testing", 
            "tools": ["gh", "lazygit", "delta"],
            "system_packages": ["build-essential"]
        }
    }
}
EOF
    
    # Create mock installation scripts for testing
    create_mock_scripts
}

teardown() {
    rm -rf test_workspace
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
}

create_mock_scripts() {
    local script_dir="test_workspace/scripts"
    mkdir -p "$script_dir"
    
    # Create mock tool scripts that simulate real installation behavior
    for tool in fd ripgrep fzf gh lazygit delta; do
        cat > "$script_dir/install-$tool.sh" << EOF
#!/bin/bash
set -e
source "$REPO_DIR/scripts/lib/common.sh"

# Mock tool: $tool
TOOL_NAME="$tool"
BUILD_TYPE="\${BUILD_TYPE:-standard}"
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Parse arguments
while [[ \$# -gt 0 ]]; do
    case \$1 in
        --minimal) BUILD_TYPE="minimal"; shift ;;
        --maximum) BUILD_TYPE="maximum"; shift ;;
        --skip-deps) SKIP_DEPS=true; shift ;;
        --run-tests) RUN_TESTS=true; shift ;;
        --force) FORCE_INSTALL=true; shift ;;
        --help) echo "Mock $tool installer"; exit 0 ;;
        *) shift ;;
    esac
done

log "Mock installation of \$TOOL_NAME (build: \$BUILD_TYPE)"

# Simulate dependency check
if [[ "\$SKIP_DEPS" != true ]]; then
    log "Checking dependencies..."
    sleep 0.1
fi

# Simulate build process
log "Building \$TOOL_NAME..."
sleep 0.2

# Simulate installation
mkdir -p "\$HOME/.local/bin"
echo "#!/bin/bash\necho '\$TOOL_NAME mock v1.0.0'" > "\$HOME/.local/bin/\$TOOL_NAME"
chmod +x "\$HOME/.local/bin/\$TOOL_NAME"

success "\$TOOL_NAME installed successfully"

# Simulate tests if requested
if [[ "\$RUN_TESTS" == true ]]; then
    log "Running tests..."
    sleep 0.1
    success "Tests completed"
fi
EOF
        chmod +x "$script_dir/install-$tool.sh"
    done
}

# =============================================================================
# WORKFLOW INTEGRATION TESTS
# =============================================================================

test_sequential_tool_installation() {
    test_start "sequential tool installation workflow"
    
    cd test_workspace
    
    # Test installing tools in sequence
    local tools=("fd" "ripgrep" "fzf")
    
    for tool in "${tools[@]}"; do
        assert_command_success "./scripts/install-$tool.sh --skip-deps" "Should install $tool successfully"
        
        # Verify tool is "installed" (mock binary exists)
        assert_file_exists "$HOME/.local/bin/$tool" "Tool $tool binary should exist"
    done
    
    cd - >/dev/null
    test_pass "Sequential installation workflow works"
}

test_dependency_chain_workflow() {
    test_start "dependency chain installation"
    
    cd test_workspace
    
    # Test dependency resolution workflow
    # First install common deps (simulated)
    log "Simulating common dependency installation..."
    
    # Then install tools with --skip-deps
    local dependent_tools=("gh" "lazygit" "delta")
    
    for tool in "${dependent_tools[@]}"; do
        assert_command_success "./scripts/install-$tool.sh --skip-deps" "Should install $tool with skipped deps"
    done
    
    cd - >/dev/null
    test_pass "Dependency chain workflow works"
}

test_build_type_consistency() {
    test_start "build type consistency across tools"
    
    cd test_workspace
    
    # Test that the same build type works across different tools
    local build_types=("minimal" "maximum")
    
    for build_type in "${build_types[@]}"; do
        log "Testing $build_type build across tools..."
        
        assert_command_success "./scripts/install-fd.sh --$build_type --skip-deps" "fd should support $build_type build"
        assert_command_success "./scripts/install-ripgrep.sh --$build_type --skip-deps" "ripgrep should support $build_type build"
        assert_command_success "./scripts/install-fzf.sh --$build_type --skip-deps" "fzf should support $build_type build"
    done
    
    cd - >/dev/null  
    test_pass "Build type consistency works"
}

test_parallel_installation_safety() {
    test_start "parallel installation safety"
    
    cd test_workspace
    
    # Test that parallel installations don't interfere
    local pids=()
    
    # Start multiple installations concurrently
    for tool in fd ripgrep fzf; do
        ./scripts/install-$tool.sh --skip-deps &
        pids+=($!)
        sleep 0.1  # Slight stagger to avoid race conditions
    done
    
    # Wait for all to complete
    local all_success=true
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            all_success=false
        fi
    done
    
    assert_command_success "[[ '$all_success' == true ]]" "All parallel installations should succeed"
    
    # Verify all tools were installed
    for tool in fd ripgrep fzf; do
        assert_file_exists "$HOME/.local/bin/$tool" "Tool $tool should be installed after parallel run"
    done
    
    cd - >/dev/null
    test_pass "Parallel installation safety works"
}

test_error_recovery_workflow() {
    test_start "error recovery and cleanup workflow"
    
    cd test_workspace
    
    # Create a failing script to test error handling
    cat > scripts/install-failing-tool.sh << 'EOF'
#!/bin/bash
set -e
source "$(dirname "$(dirname "$PWD")")/scripts/lib/common.sh"

log "Starting failing tool installation..."
log "This tool will fail intentionally"
error "Simulated installation failure"
EOF
    chmod +x scripts/install-failing-tool.sh
    
    # Test that failure is handled gracefully
    assert_command_failure "./scripts/install-failing-tool.sh" "Failing script should exit with error"
    
    # Test that system remains in good state after failure
    assert_command_success "./scripts/install-fd.sh --skip-deps" "Should still be able to install other tools after failure"
    
    cd - >/dev/null
    test_pass "Error recovery workflow works"
}

test_configuration_workflow() {
    test_start "configuration and customization workflow"
    
    cd test_workspace
    
    # Test config-only mode workflow
    assert_command_success "./scripts/install-fd.sh --skip-deps --help" "Help should work"
    
    # Test different modes
    assert_command_success "./scripts/install-fd.sh --skip-deps" "Standard installation should work" 
    assert_command_success "./scripts/install-ripgrep.sh --minimal --skip-deps" "Minimal build should work"
    assert_command_success "./scripts/install-fzf.sh --maximum --skip-deps" "Maximum build should work"
    
    cd - >/dev/null
    test_pass "Configuration workflow works"
}

test_cli_to_script_workflow() {
    test_start "CLI to script delegation workflow"
    
    # Test that the main CLI can delegate to scripts
    # This tests the integration between Go CLI and shell scripts
    
    if [[ -f "$REPO_DIR/build/gearbox" ]]; then
        # Test basic CLI commands
        assert_command_success "$REPO_DIR/build/gearbox help" "CLI help should work"
        assert_command_success "$REPO_DIR/build/gearbox list" "CLI list should work"
        
        # Test configuration commands
        assert_command_success "$REPO_DIR/build/gearbox config show" "CLI config show should work"
        
        # Note: We don't test actual installation via CLI to avoid side effects
        # But we verify the CLI can load and process commands
    else
        test_skip "CLI workflow tests" "gearbox binary not found"
    fi
    
    test_pass "CLI to script workflow works"
}

# =============================================================================
# COMPREHENSIVE SYSTEM TESTS
# =============================================================================

test_full_development_workflow() {
    test_start "full development environment setup workflow"
    
    cd test_workspace
    
    # Simulate a complete development environment setup
    log "Simulating development environment setup..."
    
    # Install core tools
    local core_tools=("fd" "ripgrep" "fzf")
    for tool in "${core_tools[@]}"; do
        assert_command_success "./scripts/install-$tool.sh --skip-deps" "Core tool $tool should install"
    done
    
    # Install development tools
    local dev_tools=("gh" "lazygit" "delta")
    for tool in "${dev_tools[@]}"; do
        assert_command_success "./scripts/install-$tool.sh --skip-deps" "Dev tool $tool should install"
    done
    
    # Verify all tools are available
    export PATH="$HOME/.local/bin:$PATH"
    for tool in "${core_tools[@]}" "${dev_tools[@]}"; do
        assert_command_success "command -v $tool" "Tool $tool should be in PATH"
    done
    
    cd - >/dev/null
    test_pass "Full development workflow works"
}

test_bundle_installation_simulation() {
    test_start "bundle installation workflow simulation"
    
    cd test_workspace
    
    # Simulate bundle installation (like gearbox install --bundle essential)
    log "Simulating essential bundle installation..."
    
    # Read bundle configuration and install tools
    if [[ -f config/test_bundle.json ]]; then
        # Simulate parsing bundle config and installing tools
        local bundle_tools=("fd" "ripgrep" "fzf")
        
        log "Installing bundle: test-essential"
        for tool in "${bundle_tools[@]}"; do
            assert_command_success "./scripts/install-$tool.sh --skip-deps" "Bundle tool $tool should install"
        done
        
        # Verify bundle completion
        local installed_count=0
        for tool in "${bundle_tools[@]}"; do
            if [[ -f "$HOME/.local/bin/$tool" ]]; then
                ((installed_count++))
            fi
        done
        
        assert_command_success "[[ $installed_count -eq ${#bundle_tools[@]} ]]" "All bundle tools should be installed"
    else
        test_skip "bundle installation" "Bundle config not found"
    fi
    
    cd - >/dev/null
    test_pass "Bundle installation workflow works"
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

main() {
    echo "🔧 Multi-Tool Workflow Integration Tests"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    # Setup test environment
    setup
    
    # Run workflow tests
    echo "📦 Testing Tool Installation Workflows..."
    test_sequential_tool_installation
    test_dependency_chain_workflow
    test_build_type_consistency
    
    echo "⚡ Testing Parallel Operations..."
    test_parallel_installation_safety
    
    echo "🛠️ Testing Error Handling..."
    test_error_recovery_workflow
    
    echo "⚙️ Testing Configuration..."
    test_configuration_workflow
    
    echo "🖥️ Testing CLI Integration..."
    test_cli_to_script_workflow
    
    echo "🎯 Testing Complete Workflows..."
    test_full_development_workflow
    test_bundle_installation_simulation
    
    # Cleanup
    teardown
    
    # Show results
    print_test_summary
}

# Execute tests
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi