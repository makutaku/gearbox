#!/bin/bash

# Template Validation Tests
# Tests template generation, validation, and output quality

# Test setup
setup() {
    export GEARBOX_TEST_MODE=true
    
    # Create test environment
    mkdir -p test_env/{templates,scripts,config,generated}
    
    # Copy templates and configuration
    if [[ -d "$REPO_DIR/templates" ]]; then
        cp -r "$REPO_DIR/templates/"* test_env/templates/
    fi
    
    if [[ -f "$REPO_DIR/config/tools.json" ]]; then
        cp -r "$REPO_DIR/config" test_env/
    fi
    
    # Copy script generator if available
    if [[ -f "$REPO_DIR/bin/script-generator" ]]; then
        mkdir -p test_env/bin
        cp "$REPO_DIR/bin/script-generator" test_env/bin/
    fi
}

# Test teardown
teardown() {
    rm -rf test_env
    unset GEARBOX_TEST_MODE
}

# Test template syntax validation
test_template_syntax() {
    cd test_env
    
    # Check that all templates exist
    for template in rust.sh.tmpl go.sh.tmpl python.sh.tmpl c.sh.tmpl base.sh.tmpl; do
        assert_file_exists "templates/$template" "Template $template should exist"
    done
    
    # Basic syntax check - templates should not have obvious syntax errors
    for template in templates/*.tmpl; do
        local template_name=$(basename "$template")
        
        # Check for balanced braces
        local open_braces=$(grep -o '{{' "$template" | wc -l)
        local close_braces=$(grep -o '}}' "$template" | wc -l)
        assert_equals "$open_braces" "$close_braces" "Template $template_name should have balanced braces"
        
        # Check for proper template syntax
        assert_command_failure "grep -q '{{ *[^}]*[^}] *}}' '$template'" "Template $template_name should not have malformed expressions"
    done
    
    cd - >/dev/null
}

# Test template variable usage
test_template_variables() {
    cd test_env
    
    # Check for consistent variable usage across templates
    local common_vars=("Tool.Name" "Tool.Description" "Tool.BinaryName" "Tool.Repository")
    
    for template in templates/*.tmpl; do
        local template_name=$(basename "$template")
        
        for var in "${common_vars[@]}"; do
            if grep -q "{{.*$var.*}}" "$template"; then
                # If variable is used, check it's used correctly
                assert_command_failure "grep -q '{{[^}]*$var[^}]*|[^}]*}}' '$template'" "Variable $var should not have pipes in $template_name"
            fi
        done
    done
    
    cd - >/dev/null
}

# Test script generation with valid tools
test_script_generation() {
    cd test_env
    
    if [[ ! -f bin/script-generator ]] || [[ ! -f config/tools.json ]]; then
        test_skip "test_script_generation" "Script generator or config not available"
        return
    fi
    
    # Test generating scripts for each language
    local test_tools=("fd" "fzf" "serena" "jq")
    
    for tool in "${test_tools[@]}"; do
        # Test dry-run generation
        assert_command_success "./bin/script-generator generate --dry-run --validate=false '$tool'" "Dry-run generation should work for $tool"
        
        # Test actual generation
        assert_command_success "./bin/script-generator generate --force --validate=false -o generated '$tool'" "Script generation should work for $tool"
        
        # Check generated script exists
        assert_file_exists "generated/install-$tool.sh" "Generated script should exist for $tool"
        
        # Check generated script is executable
        assert_command_success "[[ -x 'generated/install-$tool.sh' ]]" "Generated script should be executable for $tool"
    done
    
    cd - >/dev/null
}

# Test generated script quality
test_generated_script_quality() {
    cd test_env
    
    if [[ ! -f bin/script-generator ]]; then
        test_skip "test_generated_script_quality" "Script generator not available"
        return
    fi
    
    # Generate a test script
    ./bin/script-generator generate --force --validate=false -o generated fd >/dev/null 2>&1 || true
    
    if [[ -f generated/install-fd.sh ]]; then
        local script="generated/install-fd.sh"
        
        # Check for required elements
        assert_contains "$(cat "$script")" "#!/bin/bash" "Generated script should have shebang"
        assert_contains "$(cat "$script")" "set -e" "Generated script should have error handling"
        assert_contains "$(cat "$script")" "show_help()" "Generated script should have help function"
        
        # Check for security practices
        assert_command_failure "grep -q 'eval' '$script'" "Generated script should not use eval"
        assert_contains "$(cat "$script")" "EUID" "Generated script should check for root user"
        
        # Check for consistent variable usage
        assert_command_failure "grep -q '\$[A-Z_]*[a-z]' '$script'" "Generated script should use consistent variable naming"
        
        # Check for proper argument parsing
        assert_contains "$(cat "$script")" "while.*\[\[ \$# -gt 0 \]\]" "Generated script should have argument parsing loop"
        
        # Test script help functionality
        assert_command_success "./$script --help" "Generated script help should work"
    else
        test_fail "test_generated_script_quality" "Could not generate test script"
    fi
    
    cd - >/dev/null
}

# Test template consistency across languages
test_template_consistency() {
    cd test_env
    
    # Check that all language templates have similar structure
    local required_sections=("show_help" "parse.*arguments" "install.*dependencies" "Build.*TYPE")
    
    for template in templates/rust.sh.tmpl templates/go.sh.tmpl templates/python.sh.tmpl templates/c.sh.tmpl; do
        if [[ -f "$template" ]]; then
            local template_name=$(basename "$template")
            
            for section in "${required_sections[@]}"; do
                assert_command_success "grep -qi '$section' '$template'" "Template $template_name should have $section section"
            done
        fi
    done
    
    cd - >/dev/null
}

# Test template configuration integration
test_configuration_integration() {
    cd test_env
    
    if [[ ! -f config/tools.json ]]; then
        test_skip "test_configuration_integration" "tools.json not available"
        return
    fi
    
    # Validate JSON structure
    assert_command_success "python3 -m json.tool config/tools.json >/dev/null" "tools.json should be valid JSON"
    
    # Check for required fields in tool configurations
    local required_fields=("name" "description" "language" "repository" "binary_name")
    
    for field in "${required_fields[@]}"; do
        assert_command_success "grep -q '\"$field\"' config/tools.json" "tools.json should contain $field field"
    done
    
    # Check for build_types consistency
    assert_command_success "grep -q '\"build_types\"' config/tools.json" "tools.json should contain build_types"
    
    cd - >/dev/null
}

# Test template security
test_template_security() {
    cd test_env
    
    # Check templates for potential security issues
    for template in templates/*.tmpl; do
        local template_name=$(basename "$template")
        
        # Check for dangerous commands
        assert_command_failure "grep -qi 'rm.*-rf.*/' '$template'" "Template $template_name should not contain dangerous rm commands"
        assert_command_failure "grep -qi 'eval.*\$' '$template'" "Template $template_name should not use eval with variables"
        assert_command_failure "grep -qi 'sudo.*passwd' '$template'" "Template $template_name should not modify passwords"
        
        # Check for proper variable escaping in shell commands
        assert_command_failure "grep -q '\$[A-Z_]*[^}].*|' '$template'" "Template $template_name should not have unescaped variables in pipes"
    done
    
    cd - >/dev/null
}

# Test build type flag consistency
test_build_type_flags() {
    cd test_env
    
    if [[ ! -f config/tools.json ]] || [[ ! -f bin/script-generator ]]; then
        test_skip "test_build_type_flags" "Required files not available"
        return
    fi
    
    # Generate scripts for a few tools and check flag consistency
    local test_tools=("fd" "ripgrep")
    
    for tool in "${test_tools[@]}"; do
        ./bin/script-generator generate --force --validate=false -o generated "$tool" >/dev/null 2>&1 || continue
        
        local script="generated/install-$tool.sh"
        if [[ -f "$script" ]]; then
            # Check that all build type flags are unique
            local flags=$(grep -o '\-[a-z].*) *$' "$script" | sed 's/)//' | sort)
            local unique_flags=$(echo "$flags" | sort -u)
            
            local flag_count=$(echo "$flags" | wc -l)
            local unique_count=$(echo "$unique_flags" | wc -l)
            
            assert_equals "$flag_count" "$unique_count" "All flags should be unique in $tool script"
        fi
    done
    
    cd - >/dev/null
}

# Test template performance
test_template_performance() {
    cd test_env
    
    if [[ ! -f bin/script-generator ]]; then
        test_skip "test_template_performance" "Script generator not available"
        return
    fi
    
    # Time script generation
    local start_time=$(date +%s%N)
    
    # Generate multiple scripts
    for tool in fd ripgrep fzf jq; do
        ./bin/script-generator generate --force --validate=false -o generated "$tool" >/dev/null 2>&1 || true
    done
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    # Generation should be reasonably fast (under 5 seconds for 4 tools)
    assert_command_success "[[ $duration -lt 5000 ]]" "Script generation should be fast (${duration}ms for 4 tools)"
    
    cd - >/dev/null
}

# Test template error handling
test_template_error_handling() {
    cd test_env
    
    if [[ ! -f bin/script-generator ]]; then
        test_skip "test_template_error_handling" "Script generator not available"
        return
    fi
    
    # Test with invalid tool name
    assert_command_failure "./bin/script-generator generate 'nonexistent-tool'" "Should fail with invalid tool"
    
    # Test with missing template directory
    mv templates templates.backup
    assert_command_failure "./bin/script-generator generate fd" "Should fail with missing templates"
    mv templates.backup templates
    
    # Test with invalid configuration
    if [[ -f config/tools.json ]]; then
        cp config/tools.json config/tools.json.backup
        echo "{invalid json}" > config/tools.json
        assert_command_failure "./bin/script-generator generate fd" "Should fail with invalid JSON"
        mv config/tools.json.backup config/tools.json
    fi
    
    cd - >/dev/null
}