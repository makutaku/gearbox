# Contributing to Essential Tools Installer

Thank you for your interest in contributing! This guide will help you add new tools and improve the installer.

## Development Setup

### 1. Fork and Clone
```bash
# Fork the repository on GitHub, then:
git clone https://github.com/yourusername/gearbox.git
cd gearbox
```

### 2. Understand the Architecture
Read the [Development Guide](docs/DEVELOPMENT.md) to understand:
- Three-tier script architecture (config → orchestration → individual scripts)
- Shared dependency management strategy
- Build type system (minimal/standard/maximum)

### 3. Set Up Development Environment
```bash
# Test that existing scripts work
./tests/test-runner.sh

# Try a minimal installation
./install-tools --minimal fd
```

## Adding New Tools

### 1. Create Installation Script
Create `scripts/install-newtool.sh` following the established pattern:

**Required Elements:**
- Source `config.sh` for shared functions and configuration
- Support standard command-line flags: `--debug`, `--release`, `--optimized`
- Include `--skip-deps`, `--run-tests`, `--force` options
- Use shared logging functions: `log()`, `error()`, `success()`, `warning()`
- Follow the standard directory structure (`~/tools/build/`, `/usr/local/bin/`)

**Example Template:**
```bash
#!/bin/bash
# Tool Name Installation Script for Debian Linux
# Usage: ./install-newtool.sh [OPTIONS]

set -e

# Source shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
source "$REPO_DIR/config.sh"

# Tool configuration
TOOL_DIR="newtool"
TOOL_REPO="https://github.com/author/newtool.git"
MIN_VERSION="1.0.0"

# Standard options
BUILD_TYPE="release"
MODE="install"
SKIP_DEPS=false
RUN_TESTS=false
FORCE_INSTALL=false

# Help function, argument parsing, dependency installation,
# build configuration, compilation, and installation logic
```

### 2. Update Main Installer
Add your tool to `scripts/install-all-tools.sh`:

```bash
# Add to AVAILABLE_TOOLS array
AVAILABLE_TOOLS=("ffmpeg" "7zip" "jq" "fd" "ripgrep" "fzf" "newtool")

# Add build flag mapping
newtool)
    case $BUILD_TYPE in
        minimal) echo "-d" ;;
        standard) echo "-r" ;;
        maximum) echo "-o" ;;
    esac
    ;;

# Add to installation order (consider dependencies)
```

### 3. Update Documentation
- Add tool entry to main `README.md` tools table
- Include usage examples in `docs/QUICK_START.md`
- Document any special requirements or limitations

## Script Guidelines

### Command-Line Interface Standards
All scripts must support:
```bash
# Build types
-d, --debug           # Debug build with symbols
-r, --release         # Optimized release build (default)
-o, --optimized       # Maximum optimizations

# Common options
--skip-deps          # Skip dependency installation
--run-tests          # Run test suite after building
--force              # Force reinstallation
-h, --help           # Show help message
```

### Error Handling Best Practices
```bash
# Always use error handling
set -e

# Use shared logging functions
log "Starting installation..."
success "Installation completed!"
error "Build failed: $error_message"
warning "Non-critical issue detected"

# Provide context in error messages
error "Failed to compile $TOOL_NAME: missing dependency $DEP_NAME"
```

### Directory Management
```bash
# Use shared configuration
source "$REPO_DIR/config.sh"

# Follow standard paths
TOOL_SOURCE_DIR="$(get_source_dir "$TOOL_NAME")"
CACHE_FILE="$(get_cache_path "tool-download.tar.gz")"

# Clean up on failure
trap 'rm -rf "$TEMP_DIR"' EXIT
```

### Version and Dependency Handling
```bash
# Check existing installations
if command -v "$TOOL_NAME" &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "$TOOL_NAME already installed. Use --force to reinstall."
    exit 0
fi

# Install dependencies conditionally
if [[ "$SKIP_DEPS" != true ]]; then
    install_dependencies
fi

# Verify installation
if ! command -v "$TOOL_NAME" &> /dev/null; then
    error "$TOOL_NAME installation failed - binary not found"
fi
```

## Testing Guidelines

### Basic Testing
```bash
# Run the basic test suite
./tests/test-runner.sh

# Test your specific script
scripts/install-newtool.sh --debug --run-tests

# Test via main installer
./install-tools --minimal newtool --run-tests
```

### Manual Verification
```bash
# Verify installation success
which newtool
newtool --version

# Test in clean environment (Docker recommended)
docker run -it debian:bookworm /bin/bash
# Clone repo and test installation
```

### Integration Testing
```bash
# Test with different build types
scripts/install-newtool.sh --debug
scripts/install-newtool.sh --release  
scripts/install-newtool.sh --optimized

# Test dependency skipping
scripts/install-common-deps.sh
scripts/install-newtool.sh --skip-deps

# Test forced reinstallation
scripts/install-newtool.sh --force
```

## Code Quality Standards

### Script Structure
1. **Header**: Tool description and usage
2. **Configuration**: Tool-specific settings and defaults
3. **Functions**: Helper functions and main logic
4. **Argument parsing**: Handle command-line options
5. **Dependency management**: Install or skip dependencies
6. **Build process**: Configure, compile, install
7. **Verification**: Confirm successful installation

### Documentation Requirements
- Clear usage examples in help text
- Document any special build requirements
- Include troubleshooting notes for common issues
- Update relevant documentation files

### Performance Considerations
- Use optimal build flags for each build type
- Consider parallel builds where appropriate
- Minimize redundant dependency installations
- Clean up temporary files and directories

## Submission Process

### Before Submitting
1. **Test thoroughly**: Run tests on clean systems
2. **Update documentation**: Ensure all docs are current
3. **Follow conventions**: Match existing code style and patterns
4. **Verify integration**: Test with main installer script

### Pull Request Guidelines
```bash
# Create feature branch
git checkout -b add-newtool-support

# Make your changes
# ... implement script and documentation updates ...

# Test your changes
./tests/test-runner.sh
./install-tools newtool --run-tests

# Commit with clear messages
git add .
git commit -m "Add newtool installation script

- Implement scripts/install-newtool.sh with standard options
- Add newtool to main installer with build type support  
- Update documentation with usage examples
- Include test validation for successful installation"

# Push and create pull request
git push origin add-newtool-support
```

### Pull Request Description
Include:
- **Tool description**: What the tool does and why it's useful
- **Implementation details**: Build process, dependencies, special considerations
- **Testing performed**: Platforms tested, build types verified
- **Documentation updates**: Files modified and examples added

## Getting Help

- **Questions**: Open an issue with the "question" label
- **Bugs**: Report issues with detailed reproduction steps
- **Feature requests**: Propose new tools or improvements

Thank you for contributing to make development environment setup easier for everyone!
