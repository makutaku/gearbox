#!/bin/bash

# Performance Benchmarking Tests for Gearbox Build Processes
# Measures performance characteristics and identifies optimization opportunities

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"

# Source test framework and common library
source "$SCRIPT_DIR/framework/test-framework.sh"
source "$REPO_DIR/scripts/lib/common.sh"

# =============================================================================
# BENCHMARK CONFIGURATION
# =============================================================================

# Benchmark results storage
declare -A BENCHMARK_RESULTS
BENCHMARK_START_TIME=""
BENCHMARK_COUNT=0

# Timing utilities
benchmark_start() {
    BENCHMARK_START_TIME=$(date +%s.%N)
}

benchmark_end() {
    local end_time=$(date +%s.%N)
    if command -v bc >/dev/null 2>&1; then
        local duration=$(echo "$end_time - $BENCHMARK_START_TIME" | bc -l)
    else
        # Fallback to integer seconds for systems without bc
        local start_int=${BENCHMARK_START_TIME%%.*}
        local end_int=${end_time%%.*}
        local duration=$((end_int - start_int))
    fi
    echo "$duration"
}

record_benchmark() {
    local test_name="$1"
    local duration="$2"
    BENCHMARK_RESULTS["$test_name"]="$duration"
    ((BENCHMARK_COUNT++))
    
    printf "⏱️  %-40s %8.3fs\\n" "$test_name" "$duration"
}

# Test environment setup
setup() {
    export GEARBOX_TEST_MODE=true
    export GEARBOX_SKIP_CACHE_CLEANUP=true
    
    # Create performance test workspace
    mkdir -p perf_test/{scripts,cache,temp}
    
    # Create representative test scripts for benchmarking
    create_benchmark_scripts
}

teardown() {
    rm -rf perf_test
    unset GEARBOX_TEST_MODE
    unset GEARBOX_SKIP_CACHE_CLEANUP
}

create_benchmark_scripts() {
    local script_dir="perf_test/scripts"
    
    # Create lightweight script for performance testing
    cat > "$script_dir/install-perf-test.sh" << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
source "$REPO_DIR/scripts/lib/common.sh"

# Simulate different workloads
WORKLOAD="${1:-light}"
BUILD_TYPE="${2:-standard}"

case "$WORKLOAD" in
    light)
        # Quick operations
        for i in {1..10}; do
            validate_tool_name "tool$i" >/dev/null
        done
        ;;
    medium)
        # Medium complexity operations
        for i in {1..100}; do
            sanitize_filename "test-file-$i.txt" >/dev/null
            version_compare "1.$i.0" "1.0.0" >/dev/null || true
        done
        ;;
    heavy)
        # Heavy operations (simulated build)
        for i in {1..1000}; do
            get_optimal_jobs >/dev/null
        done
        ;;
esac

success "Performance test completed (workload: $WORKLOAD, build: $BUILD_TYPE)"
EOF
    chmod +x "$script_dir/install-perf-test.sh"
    
    # Create script for dependency installation benchmarks
    cat > "$script_dir/install-deps-perf.sh" << 'EOF'
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
source "$REPO_DIR/scripts/lib/common.sh"

# Simulate dependency checking and validation
log "Checking system dependencies..."

# Simulate checking various tools
tools=("gcc" "make" "git" "curl" "pkg-config" "autoconf" "cmake")
for tool in "${tools[@]}"; do
    check_tool_installed "$tool" true >/dev/null 2>&1 || true
done

# Simulate network checks
log "Simulating network connectivity checks..."
for url in "https://github.com" "https://releases.ubuntu.com" "https://golang.org"; do
    validate_url "$url" >/dev/null
done

success "Dependency check simulation completed"
EOF
    chmod +x "$script_dir/install-deps-perf.sh"
}

# =============================================================================
# CORE FUNCTION PERFORMANCE TESTS
# =============================================================================

test_validation_performance() {
    test_start "validation function performance"
    
    # Benchmark tool name validation
    benchmark_start
    for i in {1..1000}; do
        validate_tool_name "tool$i" >/dev/null
    done
    local duration=$(benchmark_end)
    record_benchmark "validate_tool_name (1000 calls)" "$duration"
    
    # Benchmark URL validation
    benchmark_start
    for i in {1..100}; do
        validate_url "https://github.com/user$i/repo.git" >/dev/null
    done
    duration=$(benchmark_end)
    record_benchmark "validate_url (100 calls)" "$duration"
    
    # Benchmark filename sanitization
    benchmark_start
    for i in {1..500}; do
        sanitize_filename "test-file-$i-with-special-chars\$&.txt" >/dev/null
    done
    duration=$(benchmark_end)
    record_benchmark "sanitize_filename (500 calls)" "$duration"
    
    test_pass "Validation performance benchmarked"
}

test_utility_performance() {
    test_start "utility function performance"
    
    # Benchmark optimal jobs calculation
    benchmark_start
    for i in {1..100}; do
        get_optimal_jobs >/dev/null
    done
    local duration=$(benchmark_end)
    record_benchmark "get_optimal_jobs (100 calls)" "$duration"
    
    # Benchmark version comparison
    benchmark_start
    for i in {1..200}; do
        version_compare "1.$i.0" "1.0.0" >/dev/null || true
    done
    duration=$(benchmark_end)
    record_benchmark "version_compare (200 calls)" "$duration"
    
    # Benchmark human readable size formatting
    benchmark_start
    for size in 1024 1048576 1073741824 2048 4096 8192; do
        for i in {1..50}; do
            human_readable_size $((size * i)) >/dev/null
        done
    done
    duration=$(benchmark_end)
    record_benchmark "human_readable_size (300 calls)" "$duration"
    
    test_pass "Utility performance benchmarked"
}

test_script_startup_performance() {
    test_start "script startup and loading performance"
    
    cd perf_test
    
    # Benchmark script startup time (library loading)
    benchmark_start
    for i in {1..20}; do
        ./scripts/install-perf-test.sh light >/dev/null
    done
    local duration=$(benchmark_end)
    record_benchmark "script startup (20 iterations)" "$duration"
    
    # Benchmark different workloads
    for workload in light medium heavy; do
        benchmark_start
        ./scripts/install-perf-test.sh "$workload" >/dev/null
        duration=$(benchmark_end)
        record_benchmark "workload: $workload" "$duration"
    done
    
    cd - >/dev/null
    test_pass "Script startup performance benchmarked"
}

test_dependency_check_performance() {
    test_start "dependency checking performance"
    
    cd perf_test
    
    # Benchmark dependency checking workflow
    benchmark_start
    ./scripts/install-deps-perf.sh >/dev/null
    local duration=$(benchmark_end)
    record_benchmark "dependency check simulation" "$duration"
    
    cd - >/dev/null
    test_pass "Dependency check performance benchmarked"
}

test_parallel_performance() {
    test_start "parallel operation performance"
    
    cd perf_test
    
    # Benchmark sequential vs parallel execution
    benchmark_start
    for i in {1..5}; do
        ./scripts/install-perf-test.sh light >/dev/null
    done
    local sequential_duration=$(benchmark_end)
    record_benchmark "sequential execution (5 scripts)" "$sequential_duration"
    
    # Parallel execution
    benchmark_start
    local pids=()
    for i in {1..5}; do
        ./scripts/install-perf-test.sh light >/dev/null &
        pids+=($!)
    done
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    local parallel_duration=$(benchmark_end)
    record_benchmark "parallel execution (5 scripts)" "$parallel_duration"
    
    # Calculate speedup
    if command -v bc >/dev/null 2>&1; then
        local speedup=$(echo "scale=2; $sequential_duration / $parallel_duration" | bc -l)
        log "Parallel speedup: ${speedup}x"
    fi
    
    cd - >/dev/null
    test_pass "Parallel performance benchmarked"
}

# =============================================================================
# MEMORY AND RESOURCE USAGE TESTS
# =============================================================================

test_memory_usage() {
    test_start "memory usage analysis"
    
    # Monitor memory usage during script execution
    if command -v ps >/dev/null 2>&1; then
        cd perf_test
        
        # Start a background script and monitor its memory
        ./scripts/install-perf-test.sh heavy &
        local script_pid=$!
        
        # Sample memory usage
        local max_memory=0
        while kill -0 "$script_pid" 2>/dev/null; do
            local current_memory=$(ps -o rss= -p "$script_pid" 2>/dev/null || echo "0")
            if [[ $current_memory -gt $max_memory ]]; then
                max_memory=$current_memory
            fi
            sleep 0.1
        done
        wait "$script_pid"
        
        log "Peak memory usage: ${max_memory}KB"
        record_benchmark "peak memory usage (KB)" "$max_memory"
        
        cd - >/dev/null
    else
        test_skip "memory usage test" "ps command not available"
    fi
    
    test_pass "Memory usage analysis completed"
}

test_cache_performance() {
    test_start "cache system performance"
    
    # Load build modules if available
    if declare -f require_build_modules >/dev/null 2>&1; then
        require_build_modules
        
        if declare -f is_cached >/dev/null 2>&1; then
            # Benchmark cache operations
            benchmark_start
            for i in {1..100}; do
                is_cached "test-tool-$i" "standard" "v1.0.0" >/dev/null 2>&1 || true
            done
            local duration=$(benchmark_end)
            record_benchmark "cache lookup (100 calls)" "$duration"
        else
            test_skip "cache performance" "Cache functions not available"
        fi
    else
        test_skip "cache performance" "Build modules not available"
    fi
    
    test_pass "Cache performance benchmarked"
}

# =============================================================================
# PERFORMANCE ANALYSIS AND REPORTING
# =============================================================================

analyze_performance() {
    echo
    echo "📊 Performance Analysis Results"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    if [[ ${#BENCHMARK_RESULTS[@]} -gt 0 ]]; then
        echo "Benchmark Summary:"
        for test_name in "${!BENCHMARK_RESULTS[@]}"; do
            local duration="${BENCHMARK_RESULTS[$test_name]}"
            printf "  %-40s %8.3fs\\n" "$test_name" "$duration"
        done
        
        echo
        echo "Performance Recommendations:"
        
        # Analyze results and provide recommendations
        local validation_time="${BENCHMARK_RESULTS["validate_tool_name (1000 calls)"]:-0}"
        if command -v bc >/dev/null 2>&1 && [[ $(echo "$validation_time > 1.0" | bc -l) -eq 1 ]]; then
            echo "  ⚠️  Tool name validation may be slow for batch operations"
        fi
        
        local startup_time="${BENCHMARK_RESULTS["script startup (20 iterations)"]:-0}"
        if command -v bc >/dev/null 2>&1 && [[ $(echo "$startup_time > 5.0" | bc -l) -eq 1 ]]; then
            echo "  ⚠️  Script startup time could be optimized"
        else
            echo "  ✅ Script startup performance is acceptable"
        fi
        
        if [[ -n "${BENCHMARK_RESULTS["parallel execution (5 scripts)"]:-}" && -n "${BENCHMARK_RESULTS["sequential execution (5 scripts)"]:-}" ]]; then
            echo "  ✅ Parallel execution benchmarked successfully"
        fi
        
        echo "  📈 Total benchmark operations: $BENCHMARK_COUNT"
    else
        echo "No benchmark results collected"
    fi
}

# =============================================================================
# MAIN BENCHMARK EXECUTION
# =============================================================================

main() {
    echo "⚡ Gearbox Performance Benchmarks"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo
    
    # Check if bc is available for floating point math
    if ! command -v bc >/dev/null 2>&1; then
        warning "bc not available - some timing calculations will be limited"
    fi
    
    # Setup test environment
    setup
    
    echo "🔍 Benchmarking Core Functions..."
    test_validation_performance
    test_utility_performance
    
    echo "🚀 Benchmarking Script Operations..."
    test_script_startup_performance
    test_dependency_check_performance
    
    echo "⚡ Benchmarking Concurrent Operations..."
    test_parallel_performance
    
    echo "💾 Benchmarking Resource Usage..."
    test_memory_usage
    test_cache_performance
    
    # Cleanup
    teardown
    
    # Show performance analysis
    analyze_performance
    
    # Show test results
    print_test_summary
}

# Execute benchmarks
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi