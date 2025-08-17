# Development Guide

## Repository Structure
- `scripts/` - Individual installation scripts for each tool
- `docs/` - Documentation files
- `examples/` - Usage example scripts
- `tests/` - Test scripts and validation
- `config.sh` - Shared configuration and utility functions
- `install-tools` - Main wrapper script

## Installation Script Architecture

### Three-Tier System
1. **Configuration Layer** (`config.sh`)
   - Defines build directories, cache paths, install prefixes
   - Provides shared logging functions and color definitions
   - Contains utility functions for path and version management

2. **Orchestration Layer** (`scripts/install-all-tools.sh`)
   - Manages installation order for optimal dependency sharing
   - Handles build type flags and common options
   - Coordinates common dependency installation

3. **Individual Scripts** (`scripts/install-*.sh`)
   - Each tool has a dedicated script following consistent patterns
   - Supports standard command-line interface
   - Handles tool-specific build configurations

### Build Directory Strategy
- **Source repositories**: `~/tools/build/` (keeps scripts directory clean)
- **Cache directory**: `~/tools/cache/` (downloads and temporary files)
- **Install prefix**: `/usr/local/` (system-wide binaries)

## Adding New Tools

### 1. Create Installation Script
Create `scripts/install-newtool.sh` following this template:

```bash
#!/bin/bash
# Tool Name Installation Script for Debian Linux
# Usage: ./install-newtool.sh [OPTIONS]

set -e  # Exit on any error

# Source shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
source "$REPO_DIR/config.sh"

# Tool-specific configuration
TOOL_DIR="newtool"
TOOL_REPO="https://github.com/author/newtool.git"
MIN_VERSION="1.0.0"

# Default options
BUILD_TYPE="release"
MODE="install"
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Show help function
show_help() {
    cat << EOF
Tool Name Installation Script

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build
  -r, --release         Release build (default)
  -o, --optimized       Optimized build

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite
  --force              Force reinstallation
  -h, --help           Show this help
EOF
}

# Argument parsing, dependency installation, build, and install logic
```

### 2. Follow Standard Patterns

**Command-line Interface**:
- Support `-d/--debug`, `-r/--release`, `-o/--optimized` build types
- Include `--skip-deps`, `--run-tests`, `--force` flags
- Provide `-h/--help` with clear usage information

**Error Handling**:
- Use `set -e` for immediate exit on errors
- Provide clear error messages with context
- Use shared logging functions from `config.sh`

**Version Management**:
- Check for existing installations before building
- Support `--force` to override existing installations
- Verify installation success with version checks

### 3. Update Installation Order
Add the new tool to `scripts/install-all-tools.sh`:

```bash
# In AVAILABLE_TOOLS array
AVAILABLE_TOOLS=("ffmpeg" "7zip" "jq" "fd" "ripgrep" "fzf" "newtool")

# In get_build_flag function
get_build_flag() {
    local tool=$1
    case $tool in
        # ... existing cases ...
        newtool)
            case $BUILD_TYPE in
                minimal) echo "-d" ;;
                standard) echo "-r" ;;
                maximum) echo "-o" ;;
            esac
            ;;
    esac
}

# In INSTALLATION_ORDER (place according to dependencies)
INSTALLATION_ORDER=()
for tool in "fzf" "ripgrep" "fd" "jq" "ffmpeg" "7zip" "newtool"; do
    if [[ " ${SELECTED_TOOLS[*]} " =~ " ${tool} " ]]; then
        INSTALLATION_ORDER+=("$tool")
    fi
done
```

### 4. Update Documentation
- Add tool to main README.md tools table
- Update QUICK_START.md with examples
- Add any special configuration notes

## Dependency Management

### Shared Dependencies
Common dependencies are installed via `scripts/install-common-deps.sh`:
- **Programming languages**: Rust (≥1.88.0), Go (≥1.23.4), C/C++ toolchain
- **Build tools**: autotools, cmake, make, pkg-config
- **System libraries**: Based on tool requirements

### Installation Order Strategy
Tools are installed in optimal order to maximize shared dependency usage:
1. **Go tools** (fzf) - installs Go toolchain
2. **Rust tools** (ripgrep, fd) - installs Rust, uses shared toolchain
3. **C/C++ tools** (jq, ffmpeg, 7zip) - independent builds

### Dependency Skipping
Individual scripts support `--skip-deps` to avoid redundant installations:
- First tool installs all dependencies
- Subsequent tools skip dependency installation
- Manual coordination possible for complex scenarios

## Testing Strategy

### Test Runner
`tests/test-runner.sh` provides basic validation:
- Verifies script executability
- Tests configuration loading
- Can be extended for integration tests

### Individual Tool Testing
Scripts support `--run-tests` flag:
- Runs tool-specific test suites when available
- Validates installation success
- Checks version compatibility

### Manual Testing
```bash
# Test individual script
./scripts/install-newtool.sh --debug --run-tests

# Test via main installer
./install-tools --minimal newtool --run-tests

# Verify installation
which newtool
newtool --version
```

## Code Quality Guidelines

### Script Standards
- Use consistent shebang: `#!/bin/bash`
- Enable error handling: `set -e`
- Source shared configuration
- Follow existing naming conventions
- Include comprehensive help text

### Error Handling
- Use shared logging functions: `log()`, `error()`, `success()`, `warning()`
- Provide context in error messages
- Clean up on failure when appropriate
- Validate prerequisites before building

### Documentation
- Update all relevant documentation files
- Include usage examples
- Document any special requirements or limitations
- Keep installation instructions current

## Build System Design

### Three Build Types
- **Minimal**: Fast builds, essential features only
- **Standard**: Balanced performance and features (default)
- **Maximum**: Full-featured builds with all optimizations

### Configuration Management
- Shared configuration prevents duplication
- Environment variables allow customization
- Consistent directory structure across tools
- Proper PATH and library management
