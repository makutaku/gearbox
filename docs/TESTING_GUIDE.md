# Testing Guide

Comprehensive guide to gearbox's testing infrastructure, covering all test types and usage patterns.

## Overview

Gearbox includes a comprehensive multi-layered testing system designed to ensure reliability, security, and performance across all components:

### 🎯 **Recent Testing Enhancements (2024)**
- **450+ Go Test Cases**: Comprehensive coverage across all Go packages with benchmarks and edge cases
- **Modular Test Architecture**: Type-safe testing with structured assertions and validation
- **Enhanced Coverage**: All critical Go packages now have extensive test suites

### 📋 **Complete Test Coverage**
- **Go Package Testing**: 450+ test cases across `pkg/errors`, `pkg/logger`, `pkg/manifest`, `pkg/uninstall`
- **Shell Function Testing**: 50+ functions tested across modular library system
- **Security Testing**: Protection against injection attacks, path traversal, privilege escalation
- **Performance Benchmarking**: Function timing, memory usage, parallel execution analysis
- **Integration Testing**: Multi-tool workflows and CLI-to-script delegation
- **Error Resilience**: Graceful failure handling, cleanup, and rollback functionality

## Test Suite Overview

### Quick Test Commands

```bash
# Run all tests (Go + Shell)
make test

# Run Go tests only
go test ./... -v                          # All Go packages with verbose output
go test ./pkg/errors/... -v               # Test error handling package
go test ./pkg/logger/... -v               # Test structured logging package  
go test ./pkg/manifest/... -v             # Test manifest management package
go test ./pkg/uninstall/... -v            # Test uninstall functionality package
go test ./pkg/orchestrator/... -v         # Test orchestrator package

# Run Go tests with benchmarks
go test ./... -bench=. -benchmem          # Run all benchmarks with memory stats

# Run basic script validation
./tests/test-runner.sh

# Run specific shell test suites
./tests/test_core_functions.sh           # Essential function validation (14 functions)
./tests/test_unit_comprehensive.sh       # Complete unit tests (50+ functions) 
./tests/test_workflow_integration.sh     # Multi-tool workflow testing
./tests/test_performance_benchmarks.sh   # Performance analysis
./tests/test_error_handling.sh          # Security & resilience testing
```

## Test Types and Coverage

### 0. Go Package Testing (450+ Test Cases)

Comprehensive type-safe testing across all Go packages:

#### **pkg/orchestrator Package Testing**
- **ConfigManager tests** covering thread-safe configuration access with RWMutex synchronization
- **Builder pattern tests** for clean orchestrator construction with validation steps  
- **Resource management tests** including dynamic job limits and cleanup mechanisms
- **Architecture pattern validation** ensuring proper elimination of global variables anti-pattern
- **Integration tests** with updated configMgr.GetConfig() pattern throughout codebase

#### **pkg/errors Package Testing**
- **25+ test functions** covering all error types, context methods, and suggestion generation
- **Error type validation**: Installation, configuration, validation, uninstall errors
- **Context handling**: WithContext, WithDetails, GetSuggestion methods
- **Edge cases**: Nil errors, empty contexts, malformed suggestions

#### **pkg/logger Package Testing**  
- **Structured logging tests** with JSON validation and level handling
- **Configuration testing**: Multiple output formats, time formats, log levels
- **Thread safety**: Concurrent logging operations and global logger management
- **Output validation**: JSON structure validation, message formatting

#### **pkg/manifest Package Testing**
- **Complete CRUD operations**: Create, read, update, delete manifest entries
- **Atomic writes**: Transaction safety and backup/restore functionality
- **Concurrent access**: Multi-threaded access patterns and race condition testing
- **Data integrity**: JSON serialization, validation, and schema compliance

#### **pkg/uninstall Package Testing**
- **Removal execution**: Dry-run support, multiple removal methods (cargo, go, source build)
- **Dependency analysis**: Safety checks, pre-existing tool protection
- **Error handling**: Graceful failure modes and recovery procedures
- **Space calculation**: Accurate disk space measurement and reporting

#### **Benchmark Testing**
- **Performance benchmarks** for critical operations
- **Memory allocation tracking** with `-benchmem` flag
- **Optimization identification** for bottlenecks and resource usage

**Usage:**
```bash
# Run all Go tests with coverage
go test ./... -v -cover

# Run specific package tests
go test ./pkg/errors/... -v
go test ./pkg/manifest/... -cover

# Run with benchmarks and memory profiling
go test ./... -bench=. -benchmem -v
```

## Shell Test Types and Coverage

### 1. Core Function Testing (`test_core_functions.sh`)

Quick validation of essential functions across all modules:

**Logging Functions:**
- `log()`, `success()`, `warning()`, `debug()` 
- Real-time progress indicators and status reporting

**Validation Functions:**
- `validate_tool_name()` - Tool name sanitization and security
- `validate_build_type()` - Build type validation (minimal/standard/maximum)
- `validate_url()` - URL format and protocol validation
- `version_compare()` - Semantic version comparison

**Utility Functions:**
- `get_optimal_jobs()` - CPU-aware parallel job calculation
- `sanitize_filename()` - Safe filename creation with length limits
- `human_readable_size()` - Disk space formatting

**Security Functions:**
- `ensure_not_root()` - Root user prevention
- `execute_command_safely()` - Command injection protection

**Usage:**
```bash
# Quick test (30 seconds)
./tests/test_core_functions.sh

# Expected output:
# 🧪 Core Function Tests
# =====================
# Testing log... ✓ PASS
# Testing success... ✓ PASS
# ...
# 📊 Results: 14 passed, 0 failed
# 🎉 All core functions working!
```

### 2. Comprehensive Unit Testing (`test_unit_comprehensive.sh`)

Detailed testing of all shell functions across the modular library system:

**Modules Tested:**
- **Logging Module** (`scripts/lib/core/logging.sh`): 9 functions
- **Validation Module** (`scripts/lib/core/validation.sh`): 9 functions  
- **Utilities Module** (`scripts/lib/core/utilities.sh`): 7 functions
- **Security Module** (`scripts/lib/core/security.sh`): 7 functions
- **Build Modules** (`scripts/lib/build/*`): Cache and execution functions
- **System Modules** (`scripts/lib/system/*`): Installation and backup functions

**Key Test Areas:**
- Input validation and sanitization
- Error handling and edge cases
- Security protection (injection, traversal)
- Performance characteristics
- Module loading and integration

**Usage:**
```bash
# Comprehensive testing (2-3 minutes)
./tests/test_unit_comprehensive.sh
```

### 3. Workflow Integration Testing (`test_workflow_integration.sh`)

Tests realistic multi-tool installation workflows:

**Workflow Scenarios:**
- **Sequential Installation**: Installing tools one after another
- **Parallel Installation**: Concurrent tool installations with safety
- **Dependency Chains**: Complex dependency resolution workflows
- **Build Type Consistency**: Same build types across different tools
- **Error Recovery**: Graceful handling of installation failures
- **CLI Integration**: CLI-to-script delegation workflows

**Bundle Simulation:**
- Essential bundle installation (fd, ripgrep, fzf)
- Development bundle setup (gh, lazygit, delta)
- Complete development environment workflow

**Usage:**
```bash
# Integration testing (1-2 minutes)
./tests/test_workflow_integration.sh

# Creates mock installations for testing without side effects
```

### 4. Performance Benchmarking (`test_performance_benchmarks.sh`)

Analyzes performance characteristics and identifies optimization opportunities:

**Performance Metrics:**
- **Function Call Performance**: Timing for validation, utilities, security functions
- **Script Startup Time**: Library loading and initialization performance
- **Memory Usage**: Peak memory consumption during operations
- **Parallel vs Sequential**: Speedup analysis for concurrent operations
- **Cache Performance**: Build cache lookup and storage efficiency

**Benchmark Categories:**
```bash
# Function call benchmarks
validate_tool_name (1000 calls): X.XXXs
validate_url (100 calls): X.XXXs
sanitize_filename (500 calls): X.XXXs

# Script operation benchmarks  
script startup (20 iterations): X.XXXs
workload: light/medium/heavy: X.XXXs each

# Resource usage
peak memory usage: XXXKB
cache lookup performance: X.XXXs
```

**Usage:**
```bash
# Performance analysis (30-60 seconds)
./tests/test_performance_benchmarks.sh

# Outputs timing analysis and optimization recommendations
```

### 5. Error Handling & Security Testing (`test_error_handling.sh`)

Comprehensive security and resilience validation:

**Security Testing:**
- **Command Injection Protection**: Prevents `tool; rm -rf /` attacks
- **Path Traversal Prevention**: Blocks `../../../etc/passwd` access
- **Privilege Escalation**: Root user detection and prevention
- **Input Sanitization**: Buffer overflow and malicious input handling
- **File System Protection**: System file access prevention

**Error Resilience:**
- **Graceful Failure Handling**: Clean error messages and proper exit codes
- **Cleanup After Interruption**: Temporary file cleanup on script failure
- **Rollback Functionality**: File backup and restore capabilities
- **Partial Installation Recovery**: Detection and cleanup of incomplete installs
- **High Load Error Handling**: Error handling under concurrent stress

**Edge Cases:**
- Empty and whitespace-only inputs
- Very long inputs (buffer overflow protection)
- Malformed URLs and invalid parameters
- Missing libraries and corrupted files
- Network connectivity failures

**Usage:**
```bash
# Security and resilience testing (1-2 minutes)
./tests/test_error_handling.sh

# Tests 25+ error scenarios and security protections
```

## Test Framework Architecture

### Framework Components

**Test Framework** (`tests/framework/test-framework.sh`):
- Assertion functions: `assert_equals`, `assert_contains`, `assert_file_exists`
- Command validation: `assert_command_success`, `assert_command_failure`
- Test lifecycle: `test_start`, `test_pass`, `test_fail`, `test_skip`
- Result reporting: `print_test_summary` with pass/fail/skip counts

**Test Utilities:**
- Mock script generation for testing workflows
- Temporary environment setup and cleanup
- Progress tracking and timing analysis
- Benchmark result collection and analysis

### Writing New Tests

**Basic Test Structure:**
```bash
#!/bin/bash
# Source framework and common library
source "tests/framework/test-framework.sh"
source "scripts/lib/common.sh"

# Test setup
setup() {
    export GEARBOX_TEST_MODE=true
    mkdir -p test_data
}

teardown() {
    rm -rf test_data
    unset GEARBOX_TEST_MODE
}

# Individual test function
test_new_functionality() {
    test_start "new functionality description"
    
    # Test assertions
    assert_command_success "your_function 'valid_input'" "Should handle valid input"
    assert_command_failure "your_function 'invalid_input'" "Should reject invalid input"
    
    test_pass "New functionality works correctly"
}

# Main execution
main() {
    setup
    test_new_functionality
    teardown
    print_test_summary
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
```

**Best Practices:**
- Use descriptive test names and assertion messages
- Test both success and failure cases
- Include edge case testing (empty inputs, long inputs, special characters)
- Use mock objects to avoid side effects
- Clean up test artifacts in teardown functions

## CI/CD Integration

### Automated Testing

**Makefile Integration:**
```bash
# Run all tests
make test

# Individual test suites  
make test-core       # Core function tests
make test-unit       # Unit tests
make test-integration # Integration tests
make test-performance # Performance benchmarks
make test-security   # Security tests
```

**GitHub Actions Workflow:**
```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Test Suite
        run: |
          make test
          ./tests/test_core_functions.sh
          ./tests/test_workflow_integration.sh
```

### Test Coverage Requirements

**Minimum Coverage:**
- All new shell functions must have unit tests
- All new CLI commands must have integration tests
- Security-sensitive functions require edge case testing
- Performance-critical paths need benchmark coverage

**Quality Gates:**
- All core function tests must pass
- No security test failures allowed
- Performance regression testing for critical paths
- Integration tests must pass for multi-tool workflows

## Debugging and Troubleshooting

### Test Debugging

**Verbose Test Output:**
```bash
# Run tests with detailed output
./tests/test_core_functions.sh --verbose

# Debug specific test failures
./tests/test_unit_comprehensive.sh 2>&1 | grep -A 5 -B 5 "FAIL"
```

**Common Test Issues:**

1. **Function Not Found**: 
   - Ensure `scripts/lib/common.sh` is sourced correctly
   - Check if function is in the right module (core, build, system)
   - Verify module loading with `declare -f function_name`

2. **Permission Errors**:
   - Tests should not require root privileges
   - Use mock operations instead of actual system modifications
   - Check file permissions in test setup

3. **Race Conditions**:
   - Use proper synchronization in parallel tests
   - Add small delays between concurrent operations
   - Clean up shared resources properly

**Performance Test Debugging:**
```bash
# Check if bc is available for floating point math
command -v bc || echo "Install bc for precise timing"

# Monitor resource usage during tests
top -p $(pgrep -f test_performance) -d 1

# Profile memory usage
valgrind --tool=massif ./tests/test_performance_benchmarks.sh
```

## Test Maintenance

### Updating Tests

When modifying gearbox functionality:

1. **Update Unit Tests**: Modify corresponding test functions for changed behavior
2. **Add New Test Cases**: Cover new functionality with appropriate test scenarios  
3. **Update Integration Tests**: Ensure workflow tests cover new tool interactions
4. **Performance Baselines**: Update expected performance characteristics
5. **Security Review**: Add security tests for new input validation requirements

### Test Performance

**Monitoring Test Suite Performance:**
- Core function tests: < 30 seconds
- Unit tests: < 3 minutes  
- Integration tests: < 2 minutes
- Performance benchmarks: < 1 minute
- Security tests: < 2 minutes

**Total Test Suite**: < 10 minutes for complete coverage

### Regular Maintenance

**Weekly:**
- Run full test suite on clean environment
- Check for any flaky or intermittent test failures
- Review performance benchmark trends

**Monthly:**
- Update security test scenarios for new threat vectors
- Review and optimize slow-running tests
- Update mock scripts to match real script changes

**Before Releases:**
- Run complete test suite on multiple environments
- Verify all security protections are working
- Confirm performance benchmarks meet expectations
- Test with different system configurations

## Conclusion

The gearbox testing infrastructure provides comprehensive coverage across security, performance, and functionality. This multi-layered approach ensures the system remains reliable, secure, and performant as it evolves.

For questions or test-related issues, refer to the test framework documentation or examine existing test patterns in the test suite files.