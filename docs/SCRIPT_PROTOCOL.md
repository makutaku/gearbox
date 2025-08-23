# Gearbox Installation Script Protocol

## Overview

This document defines the standardized protocol that all Gearbox installation scripts must support to ensure consistent behavior and seamless orchestration.

## Standard Interface

### Required Arguments Support

All installation scripts MUST support these standardized flags and either implement them or gracefully ignore them:

#### Build Types (mutually exclusive)
```bash
--minimal        # Fast build with essential features only
--standard       # Balanced build with reasonable features (default)
--maximum        # Full-featured build with all optimizations
```

#### Execution Modes (mutually exclusive) 
```bash
--config-only    # Configure only (prepare build environment)
--build-only     # Configure and build (no installation)
--install        # Configure, build, and install (default)
```

#### Common Options
```bash
--skip-deps      # Skip dependency installation
--force          # Force reinstallation if already installed
--run-tests      # Run test suite after building (if applicable)
--no-shell       # Skip shell integration setup
--dry-run        # Show what would be done without executing
--help           # Show usage information
--version        # Show script version information
```

#### Advanced Options (tool-specific)
```bash
--verbose        # Enable verbose output
--quiet          # Suppress non-error output
--no-cache       # Disable build cache usage
--clean          # Clean build artifacts before building
```

### Standard Behavior Rules

#### 1. Graceful Degradation
- Scripts MUST accept all standard flags without error
- If a flag is not applicable, it should be silently ignored
- Unknown flags should produce a warning but not fail

#### 2. Build Type Handling
- If build types are not meaningful for a tool, all build flags should be accepted and ignored
- Scripts should default to `--standard` behavior if no build flag is specified
- Build flags should map to tool-specific optimizations where applicable

#### 3. Error Handling
- Exit code 0: Success
- Exit code 1: Installation failure  
- Exit code 2: Configuration error
- Exit code 3: Dependency failure
- All errors should output meaningful messages to stderr

#### 4. Output Standards
- Progress indicators should be consistent across scripts
- Success/failure should be clearly indicated
- Verbose output should be controlled by `--verbose`
- Quiet mode should suppress all non-error output

### Example Implementation Template

```bash
#!/bin/bash
# Standard Gearbox Installation Script Template

set -e

# Parse arguments with standard protocol
BUILD_TYPE="standard"  
MODE="install"
SKIP_DEPS=false
FORCE=false
RUN_TESTS=false
NO_SHELL=false
DRY_RUN=false
VERBOSE=false
QUIET=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --minimal)     BUILD_TYPE="minimal"; shift ;;
        --standard)    BUILD_TYPE="standard"; shift ;;
        --maximum)     BUILD_TYPE="maximum"; shift ;;
        --config-only) MODE="config"; shift ;;
        --build-only)  MODE="build"; shift ;;
        --install)     MODE="install"; shift ;;
        --skip-deps)   SKIP_DEPS=true; shift ;;
        --force)       FORCE=true; shift ;;
        --run-tests)   RUN_TESTS=true; shift ;;
        --no-shell)    NO_SHELL=true; shift ;;
        --dry-run)     DRY_RUN=true; shift ;;
        --verbose)     VERBOSE=true; shift ;;
        --quiet)       QUIET=true; shift ;;
        --help)        show_help; exit 0 ;;
        --version)     echo "Script version: 1.0"; exit 0 ;;
        -*)            echo "Warning: unknown option $1" >&2; shift ;;
        *)             echo "Error: unexpected argument $1" >&2; exit 1 ;;
    esac
done

# Standard function implementations
log() { [[ $QUIET == false ]] && echo "$@"; }
log_verbose() { [[ $VERBOSE == true ]] && echo "$@"; }
log_error() { echo "ERROR: $@" >&2; }

# Tool-specific implementation follows...
```

### Migration Strategy

#### Phase 1: Update Orchestrator
- Modify orchestrator to use standardized flags
- Remove tool-specific build flags from `tools.json`
- Add backward compatibility for existing scripts

#### Phase 2: Script Template
- Create standardized script template
- Update shared library functions to support protocol

#### Phase 3: Script Migration
- Update all existing scripts to follow protocol
- Validate consistency across all tools
- Update tests to verify protocol compliance

### Validation

Each script should be tested with:
```bash
# Required flag combinations
script.sh --help
script.sh --minimal --config-only --skip-deps
script.sh --standard --install --force
script.sh --maximum --build-only --run-tests
script.sh --dry-run --verbose
script.sh --unknown-flag  # Should warn, not fail
```

### Benefits

1. **Consistent UX**: Users get same interface across all tools
2. **Reliable Orchestration**: Orchestrator can safely pass flags to any script
3. **Future-Proof**: New features can be added without breaking existing scripts
4. **Easier Maintenance**: Standardized patterns reduce complexity
5. **Better Testing**: Uniform interface enables comprehensive validation