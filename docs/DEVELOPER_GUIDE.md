# Developer Guide

Complete technical guide for understanding the architecture and contributing to the Essential Tools Installer.

## 🎯 Recent Major Improvements (2024)

The project has undergone significant architectural enhancements for better maintainability, testing, and performance:

### ✅ **Comprehensive Test Coverage**
- **450+ test cases** across all Go packages (`pkg/errors`, `pkg/logger`, `pkg/manifest`, `pkg/uninstall`)
- **Benchmarks and edge case coverage** for critical operations
- **Type-safe testing** with structured assertions and validation

### ✅ **Modular Architecture Refactoring** 
- **98% code size reduction**: Split `pkg/orchestrator/main.go` (2250 lines → 56 lines)
- **9 focused modules**: types, config, orchestrator, commands, installation, verification, nerdfonts, uninstall, utils
- **Clear separation of concerns** for easier maintenance and feature development

### ✅ **Enhanced Build Cache System**
- **Complete implementation** of `scripts/lib/build/cache.sh` with metadata and validation
- **Performance optimization** for faster reinstallations and automatic cleanup
- **Integrity validation** and comprehensive cache management functions

### ✅ **Modern Go Practices**
- **Eliminated deprecated functions**: All `ioutil.*` replaced with modern `os.*` equivalents
- **Structured error handling**: Custom error types with context and suggestions
- **Type safety improvements** throughout the codebase

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Development Setup](#development-setup)
3. [Adding New Tools](#adding-new-tools)
4. [Technical Standards](#technical-standards)
5. [Language-Specific Patterns](#language-specific-patterns)
6. [Testing and Validation](#testing-and-validation)
7. [Advanced Topics](#advanced-topics)

## Architecture Overview

### Modern Modular System

The installer follows a modern modular architecture with comprehensive testing:

#### 1. Modular Library System (`scripts/lib/`)
- **Core Modules**: Essential functions (logging, validation, security, utilities)
- **Build Modules**: Build system functions (dependencies, execution, cache, cleanup)  
- **System Modules**: System integration (installation, backup, environment)
- **Configuration**: User preferences and system configuration management
- **Lazy Loading**: Efficient module loading for optimal performance

#### 2. Orchestration Layer (`scripts/installation/common/`)
- `install-all-tools.sh` - Manages installation order for optimal dependency sharing
- `install-common-deps.sh` - Coordinates shared dependency installation
- Handles build type flags and common options
- Provides unified interface for multiple tools

#### 3. Categorized Tool Scripts (`scripts/installation/categories/`)
- **Core Tools** (`core/`): fd, ripgrep, fzf, jq, zoxide
- **Development Tools** (`development/`): gh, lazygit, delta, difftastic, etc.
- **System Tools** (`system/`): bottom, procs, bandwhich, dust, fclones
- **Text Tools** (`text/`): bat, sd, xsv, tealdeer, eza, choose
- **Media Tools** (`media/`): ffmpeg, imagemagick, 7zip
- **UI Tools** (`ui/`): nerd-fonts, starship, yazi
- Each tool follows consistent patterns while leveraging modular infrastructure

#### 4. Go Package Architecture (`pkg/`)
- **Modular Orchestrator**: Split into 9 focused modules (types, config, orchestrator, commands, installation, verification, nerdfonts, uninstall, utils)
- **Thread-Safe Configuration**: ConfigManager with RWMutex synchronization replaces global variables anti-pattern
- **Builder Pattern**: Clean orchestrator construction with validation steps instead of monolithic constructor
- **Dynamic Resource Management**: Intelligent job limits based on CPU count and memory with comprehensive cleanup
- **Comprehensive Testing**: 450+ test cases across all packages with benchmarks
- **Type-Safe Operations**: Structured error handling, manifest tracking, configuration management
- **Modern Practices**: Updated to current Go standards, eliminated deprecated functions (no more `ioutil.*`)
- **Security Enhancements**: Secure debug logging with proper file permissions and location

#### 5. Comprehensive Testing System (`tests/`)
- **Go Test Coverage**: 450+ test cases across `pkg/errors`, `pkg/logger`, `pkg/manifest`, `pkg/uninstall`
- **Shell Function Tests**: Complete coverage (50+ functions across all modules)
- **Integration Tests**: Multi-tool workflow and cross-component validation
- **Performance Benchmarks**: Timing and resource usage analysis with optimization identification
- **Security Tests**: Command injection prevention, privilege escalation protection

### Directory Strategy

```
~/tools/build/              # Source repositories (temporary)
~/tools/cache/              # Downloads and cache files  
/usr/local/bin/             # Installed binaries (system-wide)
~/gearbox/                  # This repository root
├── scripts/                # All shell code
│   ├── lib/               # Modular shared libraries
│   │   ├── core/          # Essential modules (logging, validation, security, utilities)
│   │   ├── build/         # Build system modules
│   │   ├── system/        # System integration modules
│   │   └── *.sh           # Configuration and diagnostics
│   └── installation/      # Installation scripts by category
│       ├── common/        # Shared installation scripts
│       └── categories/    # Tool scripts organized by functionality
├── tests/                 # Comprehensive testing system
├── cmd/                   # Go CLI source code
├── templates/             # Script generation templates  
└── docs/                  # Documentation
```

This modular organization provides clean separation of concerns while maintaining shared infrastructure.

### Dependency Management Philosophy

**Shared Toolchains:** Install programming language toolchains once, use across multiple tools:
- Go toolchain → fzf, lazygit, gh
- Rust toolchain → ripgrep, fd, zoxide, yazi, fclones, bat, starship, eza, delta, bottom, procs, tokei, difftastic, bandwhich, xsv, hyperfine, dust, sd, tealdeer, choose
- Python toolchain → serena, uv, ruff (virtual environment pattern)
- C/C++ toolchain → jq, ffmpeg, imagemagick, 7zip

**Optimal Installation Order:** 
1. Go tools (fzf, lazygit, gh) - installs Go
2. Rust tools (21 tools total) - installs Rust, reuses toolchain
3. Python tools (serena, uv, ruff) - installs Python 3.11 + uv, uses virtual environment pattern
4. C/C++ tools (jq, ffmpeg, imagemagick, 7zip) - independent builds

## TUI Development

The Text User Interface (TUI) provides a rich interactive experience using the Bubble Tea framework.

### TUI Architecture

The TUI follows the Elm Architecture pattern with:
- **Model**: Application state and data
- **View**: Rendering functions for each component
- **Update**: Message handling and state transitions

### Directory Structure

```
cmd/gearbox/tui/
├── app.go              # Main TUI application
├── state.go            # Global state management
├── taskprovider.go     # Task adapter interface
├── styles/
│   └── theme.go        # Consistent theming
├── tasks/
│   └── manager.go      # Background tasks
└── views/
    ├── interfaces.go   # Common interfaces
    ├── dashboard.go    # Dashboard view
    ├── toolbrowser.go  # Tool browser
    ├── bundleexplorer.go # Bundle explorer
    ├── installmanager.go # Install manager
    ├── config.go       # Configuration
    └── health.go       # Health monitor
```

### Adding a New View

1. **Create the view file** in `views/`:
```go
package views

type MyView struct {
    width  int
    height int
    // Add view-specific state
}

func NewMyView() *MyView {
    return &MyView{}
}

func (v *MyView) SetSize(width, height int) {
    v.width = width
    v.height = height
}

func (v *MyView) Update(msg tea.Msg) tea.Cmd {
    // Handle messages
    return nil
}

func (v *MyView) Render() string {
    // Return rendered view
    return "My View"
}
```

2. **Add view type** to `state.go`:
```go
const (
    // ... existing views
    ViewMyView ViewType = iota
)
```

3. **Register in app.go**:
- Add field to Model struct
- Initialize in NewModel()
- Add size update in WindowSizeMsg handler
- Add update delegation in updateCurrentView()
- Add render method

4. **Add navigation**:
- Add keyboard shortcut in handleKeyPress()
- Add to navigation bar tabs
- Update help text

### Styling Guidelines

Use the theme system for consistent styling:
```go
// Use predefined styles
styles.TitleStyle().Render("Title")
styles.SuccessStyle().Render("✓ Success")
styles.ErrorStyle().Render("✗ Error")
styles.SelectedStyle().Render("Selected Item")

// Use theme colors
styles.CurrentTheme.Primary
styles.CurrentTheme.Success
styles.CurrentTheme.Warning
```

### Task Integration

For long-running operations:
```go
// Add task to manager
taskID := m.taskManager.AddTask(tool, buildType)

// Track in install manager
m.installManager.AddTaskID(taskID)

// Task updates flow automatically via channels
```

### Testing TUI Components

While the TUI requires a terminal, you can test individual components:
```go
// Test view logic
view := NewMyView()
view.SetSize(80, 24)
output := view.Render()
// Assert output structure

// Test update logic
cmd := view.Update(tea.KeyMsg{Type: tea.KeyEnter})
// Assert state changes
```

### Common Patterns

**List Navigation**:
```go
func (v *MyView) moveUp() {
    if v.cursor > 0 {
        v.cursor--
    }
}

func (v *MyView) moveDown() {
    if v.cursor < len(v.items)-1 {
        v.cursor++
    }
}
```

**Scrolling Support**:
```go
visibleItems := height - headerSize - footerSize
if v.cursor >= v.scrollOffset+visibleItems {
    v.scrollOffset = v.cursor - visibleItems + 1
}
```

**Search Implementation**:
```go
searchInput := textinput.New()
searchInput.Placeholder = "Search..."

// In Update()
if v.searchActive {
    v.searchInput, cmd = v.searchInput.Update(msg)
    v.applyFilters()
}
```

### Modern TUI Architecture

The TUI (Text User Interface) has been completely redesigned with a modern, maintainable architecture following Bubble Tea best practices:

#### **Production-Ready Architecture**
```
cmd/gearbox/tui/
├── app/                    # Core application model and messages  
│   ├── model.go           # Clean model definition with interfaces
│   └── messages.go        # Structured message types and constructors
├── interfaces/            # Interface definitions for decoupling
│   └── interfaces.go      # ToolManager, HealthChecker, TaskProvider, etc.
├── state/                 # State machine for complex workflows
│   └── machine.go         # Robust stage-based operations with recovery
├── error/                 # Centralized error handling
│   └── handler.go         # Structured error categorization and recovery
├── cache/                 # Intelligent content caching
│   └── content.go         # Data-aware cache with LRU eviction
├── benchmark/             # Performance monitoring tools
│   └── performance.go     # Real-time metrics and benchmarking
├── testing/               # Comprehensive test framework
│   └── framework.go       # teatest-based testing utilities
└── views/                 # Individual view implementations
```

#### **Key Architectural Patterns**
- **🏗️ Dependency Injection**: Factory pattern for clean service management
- **🔌 Interface-Driven Design**: Abstractions for testability and maintainability
- **📨 Message Routing**: Generic routing eliminates tight coupling  
- **⚡ Conditional Compilation**: Debug code eliminated in production builds
- **🚀 Content Caching**: Hash-based invalidation with LRU eviction
- **🔄 State Machine**: Robust workflow orchestration with error recovery

#### **Performance Features**
- **Zero-Latency Startup**: <50ms initialization with lazy data loading
- **Intelligent Caching**: 10-100x performance improvement for repeated renders
- **Memory Efficiency**: Proactive cleanup with leak detection
- **Real-Time Metrics**: Performance monitoring with <1ms overhead

#### **Advanced Testing Infrastructure**
- **teatest Framework**: Official Bubble Tea testing framework integration
- **Stress Testing**: High-frequency input simulation and memory leak detection
- **Performance Benchmarks**: Automated rendering and update performance tests
- **Navigation Testing**: Comprehensive view switching and keyboard shortcut validation

## Development Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/yourusername/gearbox.git
cd gearbox
```

### 2. Understand the Codebase

```bash
# Study the shared configuration
cat config.sh

# Examine the main orchestrator
cat scripts/install-all-tools.sh

# Look at individual tool patterns
cat scripts/install-fd.sh
cat scripts/install-ripgrep.sh
```

### 3. Set Up Development Environment

```bash
# Test that existing scripts work
./tests/test-runner.sh

# Try a minimal installation to understand the flow
gearbox --minimal fd

# Examine the build directory structure
ls -la ~/tools/build/
```

## Adding New Tools

### Step 1: Create Installation Script

Create `scripts/install-newtool.sh` following this complete template:

```bash
#!/bin/bash

# Tool Name Installation Script for Debian Linux
# Automated clone, dependency installation, configuration, build, and install
# Usage: ./install-newtool.sh [OPTIONS]

set -e  # Exit on any error

# Source shared configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
source "$REPO_DIR/config.sh"

# Tool-specific configuration
TOOL_NAME="newtool"
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
Tool Name Installation Script for Debian Linux

Usage: $0 [OPTIONS]

Build Types:
  -d, --debug           Debug build with symbols
  -r, --release         Optimized release build (default)
  -o, --optimized       Maximum optimizations

Modes:
  -c, --config-only     Configure only (prepare build)
  -b, --build-only      Configure and build (no install)
  -i, --install         Configure, build, and install (default)

Options:
  --skip-deps          Skip dependency installation
  --run-tests          Run test suite after building
  --force              Force reinstallation if already installed
  -h, --help           Show this help message

Examples:
  $0                   # Standard installation
  $0 --debug           # Debug build
  $0 --optimized       # Maximum performance
  $0 --skip-deps       # Skip dependency installation
  $0 --run-tests       # Include test validation

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--debug)
            BUILD_TYPE="debug"
            shift
            ;;
        -r|--release)
            BUILD_TYPE="release"
            shift
            ;;
        -o|--optimized)
            BUILD_TYPE="optimized"
            shift
            ;;
        -c|--config-only)
            MODE="config"
            shift
            ;;
        -b|--build-only)
            MODE="build"
            shift
            ;;
        -i|--install)
            MODE="install"
            shift
            ;;
        --skip-deps)
            SKIP_DEPS=true
            shift
            ;;
        --run-tests)
            RUN_TESTS=true
            shift
            ;;
        --force)
            FORCE_INSTALL=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

# Check if tool is already installed
if command -v "$TOOL_NAME" &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "$TOOL_NAME already installed. Use --force to reinstall."
    exit 0
fi

# Install dependencies
if [[ "$SKIP_DEPS" != true ]]; then
    log "Installing dependencies for $TOOL_NAME..."
    # Tool-specific dependency installation logic
fi

# Get source directory
TOOL_SOURCE_DIR="$(get_source_dir "$TOOL_DIR")"

# Clone or update repository
if [[ ! -d "$TOOL_SOURCE_DIR" ]]; then
    log "Cloning $TOOL_NAME repository..."
    git clone "$TOOL_REPO" "$TOOL_SOURCE_DIR"
else
    log "Updating $TOOL_NAME repository..."
    cd "$TOOL_SOURCE_DIR"
    git fetch origin
    git reset --hard origin/main
fi

cd "$TOOL_SOURCE_DIR"

# Configure build based on type
case $BUILD_TYPE in
    debug)
        log "Configuring debug build..."
        # Debug configuration
        ;;
    release)
        log "Configuring release build..."
        # Release configuration
        ;;
    optimized)
        log "Configuring optimized build..."
        # Optimized configuration
        ;;
esac

# Build
if [[ "$MODE" != "config" ]]; then
    log "Building $TOOL_NAME..."
    # Build commands
fi

# Install
if [[ "$MODE" == "install" ]]; then
    log "Installing $TOOL_NAME..."
    # Installation commands
fi

# Run tests
if [[ "$RUN_TESTS" == true ]]; then
    log "Running tests for $TOOL_NAME..."
    # Test commands
fi

# Verify installation
if [[ "$MODE" == "install" ]]; then
    if command -v "$TOOL_NAME" &> /dev/null; then
        success "$TOOL_NAME installation completed successfully!"
        log "Version: $($TOOL_NAME --version)"
    else
        error "$TOOL_NAME installation failed - binary not found in PATH"
    fi
fi
```

### Step 2: Update Main Installer

Add your tool to `scripts/install-all-tools.sh`:

```bash
# Add to AVAILABLE_TOOLS array
AVAILABLE_TOOLS=("ffmpeg" "7zip" "jq" "fd" "ripgrep" "fzf" "imagemagick" "yazi" "zoxide" "fclones" "serena" "uv" "ruff" "bat" "starship" "eza" "delta" "lazygit" "bottom" "procs" "tokei" "hyperfine" "gh" "dust" "sd" "tealdeer" "choose" "difftastic" "bandwhich" "xsv" "newtool")

# Add build flag mapping in get_build_flag function
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

# Add to installation order (consider dependencies)
# Go tools first, then Rust tools, then Python tools, then C/C++ tools
INSTALLATION_ORDER=()
for tool in "fzf" "lazygit" "gh" "ripgrep" "fd" "zoxide" "yazi" "fclones" "bat" "starship" "eza" "delta" "bottom" "procs" "tokei" "difftastic" "bandwhich" "xsv" "hyperfine" "dust" "sd" "tealdeer" "choose" "serena" "uv" "ruff" "jq" "ffmpeg" "imagemagick" "7zip" "newtool"; do
    if [[ " ${SELECTED_TOOLS[*]} " =~ " ${tool} " ]]; then
        INSTALLATION_ORDER+=("$tool")
    fi
done

# Add verification logic
case $tool in
    # ... existing cases ...
    newtool)
        if command -v newtool &> /dev/null; then
            log "✓ newtool: $(newtool --version)"
        else
            FAILED_TOOLS+=("newtool")
        fi
        ;;
esac
```

### Step 3: Update Documentation

1. Add tool to main `README.md` tools table
2. Add comprehensive usage examples to `docs/USER_GUIDE.md`
3. Document any special requirements or limitations

## Technical Standards

### Script Structure

All scripts must follow this structure:

1. **Header**: Shebang, description, usage
2. **Error handling**: `set -e` and error traps
3. **Configuration**: Source `config.sh`, define variables
4. **Functions**: Help, utility functions
5. **Argument parsing**: Standard flag handling
6. **Main logic**: Dependencies → configure → build → install → verify

### Command-Line Interface Standards

All scripts must support these options:

```bash
# Build types
-d, --debug           # Debug build with symbols
-r, --release         # Optimized release build (default)
-o, --optimized       # Maximum optimizations

# Modes
-c, --config-only     # Configure only
-b, --build-only      # Build only
-i, --install         # Full installation (default)

# Common options
--skip-deps          # Skip dependency installation
--run-tests          # Run test suite after building
--force              # Force reinstallation
-h, --help           # Show help message
```

### Error Handling Best Practices

```bash
# Always enable error handling
set -e

# Use shared logging functions
log "Starting installation..."
success "Installation completed!"
error "Build failed: specific error message"
warning "Non-critical issue detected"

# Provide context in error messages
error "Failed to compile $TOOL_NAME: missing dependency $DEP_NAME"

# Clean up on failure
trap 'cleanup_on_error' EXIT

cleanup_on_error() {
    if [[ $? -ne 0 ]]; then
        warning "Installation failed, cleaning up..."
        rm -rf "$TEMP_DIR"
    fi
}
```

### Directory Management

```bash
# Use shared configuration functions
source "$REPO_DIR/config.sh"

# Get paths using utility functions
TOOL_SOURCE_DIR="$(get_source_dir "$TOOL_NAME")"
CACHE_FILE="$(get_cache_path "tool-download.tar.gz")"

# Ensure directories exist
mkdir -p "$BUILD_DIR" "$CACHE_DIR"

# Always return to known directory
cd "$TOOL_SOURCE_DIR" || error "Failed to access $TOOL_SOURCE_DIR"
```

### Version and Dependency Handling

```bash
# Check existing installations
if command -v "$TOOL_NAME" &> /dev/null && [[ "$FORCE_INSTALL" != true ]]; then
    log "$TOOL_NAME already installed. Use --force to reinstall."
    exit 0
fi

# Version comparison example
check_version() {
    local current_version required_version
    current_version="$(tool --version | grep -oP '\d+\.\d+\.\d+')"
    required_version="$MIN_VERSION"
    
    if version_compare "$current_version" "$required_version"; then
        log "Version $current_version meets requirement $required_version"
    else
        error "Version $current_version is below required $required_version"
    fi
}

# Install dependencies conditionally
if [[ "$SKIP_DEPS" != true ]]; then
    install_dependencies
fi

# Verify installation success
verify_installation() {
    if ! command -v "$TOOL_NAME" &> /dev/null; then
        error "$TOOL_NAME installation failed - binary not found"
    fi
    
    log "Verifying $TOOL_NAME installation..."
    if ! "$TOOL_NAME" --version &> /dev/null; then
        error "$TOOL_NAME binary exists but doesn't work properly"
    fi
    
    success "$TOOL_NAME verified successfully"
}
```

## Language-Specific Patterns

### Python Tools Pattern

Python tools require special handling due to virtual environment best practices and externally managed Python environments (like Debian). Use this pattern for Python-based tools:

#### Virtual Environment + Global Wrapper Pattern

This pattern provides the best balance of Python best practices and gearbox consistency:

```bash
# Create virtual environment for tool isolation
log "Creating virtual environment for $TOOL_NAME..."
uv venv --python python3.11 .venv || error "Failed to create virtual environment"

# Install tool in virtual environment
log "Installing $TOOL_NAME in virtual environment..."
if [[ -n "$UV_INSTALL_OPTIONS" ]]; then
    uv pip install -e . $UV_INSTALL_OPTIONS || error "$TOOL_NAME installation failed"
else
    uv pip install -e . || error "$TOOL_NAME installation failed"
fi

# Create global wrapper script for system-wide access
log "Creating global wrapper script..."
WRAPPER_SCRIPT="/usr/local/bin/$TOOL_NAME"
sudo tee "$WRAPPER_SCRIPT" > /dev/null << EOF
#!/bin/bash
# $TOOL_NAME wrapper script - executes from virtual environment
exec "$TOOL_SOURCE_DIR/.venv/bin/$TOOL_NAME" "\$@"
EOF

sudo chmod +x "$WRAPPER_SCRIPT"

# Verify wrapper works
if [[ -x "$WRAPPER_SCRIPT" ]]; then
    success "Global wrapper script created: $WRAPPER_SCRIPT"
else
    error "Failed to create executable wrapper script"
fi
```

#### Python Dependency Management

```bash
# Check Python version compatibility
PYTHON_MIN_VERSION="3.11.0"
PYTHON_MAX_VERSION="3.12.0"

# Version range checking function
version_in_range() {
    local version=$1
    local min_version=$2
    local max_version=$3
    
    if version_compare "$version" "$min_version" && ! version_compare "$version" "$max_version"; then
        return 0
    else
        return 1
    fi
}

# Find compatible Python
PYTHON_CMD=""
for py_cmd in python3.11 python3 python; do
    if command -v "$py_cmd" &> /dev/null; then
        PYTHON_VERSION=$($py_cmd --version 2>&1 | grep -oP '\d+\.\d+\.\d+')
        if version_in_range "$PYTHON_VERSION" "$PYTHON_MIN_VERSION" "$PYTHON_MAX_VERSION"; then
            PYTHON_CMD="$py_cmd"
            break
        fi
    fi
done

# Install uv if needed (modern Python package installer)
if ! command -v uv &> /dev/null; then
    log "Installing uv (modern Python package installer)..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    source ~/.cargo/env || export PATH="$HOME/.cargo/bin:$PATH"
fi
```

#### Build Type Mapping for Python Tools

```bash
# Map gearbox build types to Python-specific options
get_python_build_options() {
    case $BUILD_TYPE in
        minimal)
            echo "--no-dev"  # Core features only
            ;;
        standard)
            echo ""          # Default recommended features
            ;;
        full)
            echo "--all-extras"  # All optional dependencies
            ;;
    esac
}
```

#### Why This Pattern?

1. **Python Best Practices**: Uses virtual environments for proper isolation
2. **System Compatibility**: Respects externally managed Python (Debian policy)
3. **User Experience**: Maintains global command availability like other gearbox tools
4. **Dependency Safety**: Prevents conflicts with system Python packages
5. **Maintainability**: Virtual environment stays with source code as "build artifact"

#### Alternative Approaches Considered

- **`--system` installation**: Blocked by externally managed Python
- **`--user` installation**: Complex PATH management, not globally available
- **Global virtual environment**: Breaks isolation principle
- **Per-tool system directories**: Too complex for the gearbox model

### Rust Tools Pattern

For reference, Rust tools follow this simpler pattern since Rust doesn't have system conflicts:

```bash
# Install/update Rust if needed
if ! command -v cargo &> /dev/null; then
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
    source ~/.cargo/env
fi

# Build with cargo
case $BUILD_TYPE in
    debug)
        cargo build
        ;;
    release)
        cargo build --release
        ;;
    optimized)
        RUSTFLAGS="-C target-cpu=native -C lto=fat" cargo build --release
        ;;
esac

# Install directly to system
sudo cp target/release/$TOOL_NAME /usr/local/bin/
```

### Go Tools Pattern

Go tools use direct installation:

```bash
# Install Go if needed
if ! command -v go &> /dev/null; then
    # Install Go via system packages or download
fi

# Build and install
go build -o $TOOL_NAME
sudo cp $TOOL_NAME /usr/local/bin/
```

## Testing and Validation

### Basic Testing

```bash
# Run the basic test suite
./tests/test-runner.sh

# Test individual script
scripts/install-newtool.sh --debug --run-tests

# Test via main installer
gearbox --minimal newtool --run-tests
```

### Manual Verification

```bash
# Verify installation success
which newtool
newtool --version

# Test functionality
newtool --help
newtool basic-command

# Check in clean environment
docker run -it debian:bookworm bash
# Install and test in container
```

### Integration Testing

```bash
# Test all build types
scripts/install-newtool.sh --debug
scripts/install-newtool.sh --release  
scripts/install-newtool.sh --optimized

# Test dependency coordination
scripts/install-common-deps.sh
scripts/install-newtool.sh --skip-deps

# Test error conditions
scripts/install-newtool.sh --force
# Test with missing dependencies
# Test with network issues
```

### Test Coverage Areas

1. **Installation scenarios**: All build types, with/without deps
2. **Error handling**: Network failures, missing dependencies, disk space
3. **Idempotency**: Multiple runs should be safe
4. **Clean environments**: Fresh systems, containers
5. **Integration**: Tool works with main installer
6. **Verification**: Binary works correctly after installation

## Advanced Topics

### Performance Optimization

**Build Parallelization:**
```bash
# Use available CPU cores
CORES=$(nproc)
make -j"$CORES"

# Rust builds
cargo build --jobs="$CORES"
```

**Cache Management:**
```bash
# Use cache for repeated builds
CACHE_FILE="$(get_cache_path "source-$VERSION.tar.gz")"
if [[ ! -f "$CACHE_FILE" ]]; then
    download_source "$CACHE_FILE"
fi
```

**Dependency Optimization:**
```bash
# Skip redundant installations
if [[ "$SKIP_DEPS" == true ]] || command -v rustc &> /dev/null; then
    log "Skipping Rust installation"
else
    install_rust
fi
```

### Security Considerations

**Source Verification:**
```bash
# Verify Git repository authenticity
git verify-commit HEAD

# Check checksums for downloads
echo "$EXPECTED_SHA256 $DOWNLOAD_FILE" | sha256sum -c
```

**Safe Installation:**
```bash
# Never run as root
if [[ $EUID -eq 0 ]]; then
   error "This script should not be run as root for security reasons"
fi

# Validate paths to prevent directory traversal
validate_path() {
    local path="$1"
    if [[ "$path" =~ \.\./|^/ ]]; then
        error "Invalid path: $path"
    fi
}
```

### Debugging and Troubleshooting

**Debug Mode:**
```bash
# Enable verbose output
if [[ "$DEBUG" == true ]]; then
    set -x
fi

# Detailed logging
debug() {
    if [[ "$DEBUG" == true ]]; then
        echo "[DEBUG] $*" >&2
    fi
}
```

**Error Investigation:**
```bash
# Preserve build artifacts on failure
cleanup_on_error() {
    if [[ $? -ne 0 ]] && [[ "$PRESERVE_ON_ERROR" == true ]]; then
        warning "Build failed, preserving artifacts in $BUILD_DIR"
    else
        rm -rf "$BUILD_DIR"
    fi
}
```

### Contributing Guidelines

1. **Follow existing patterns**: Study current scripts before creating new ones
2. **Test thoroughly**: All build types, clean environments, error conditions
3. **Document comprehensively**: Update all relevant documentation files
4. **Use standard interfaces**: Command-line flags, error handling, logging
5. **Consider performance**: Parallel builds, caching, dependency optimization
6. **Maintain security**: No root execution, path validation, source verification

### Submission Process

```bash
# Create feature branch
git checkout -b add-newtool-support

# Implement changes
# ... create script, update installer, update docs ...

# Test thoroughly
./tests/test-runner.sh
gearbox newtool --run-tests

# Commit with descriptive message
git add .
git commit -m "Add newtool installation script

- Implement scripts/install-newtool.sh with standard options
- Add newtool to main installer with build type support  
- Update USER_GUIDE.md with usage examples
- Include test validation for successful installation"

# Push and create pull request
git push origin add-newtool-support
```

For user-focused documentation, see the [User Guide](USER_GUIDE.md).