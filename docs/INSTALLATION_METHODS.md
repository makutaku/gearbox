# Installation Method Decision Matrix

This document explains the rationale behind installation method choices for each tool category in gearbox.

## Installation Patterns

### 1. `cargo_install` Pattern (Rust Tools)
**Method**: `cargo install` → `~/.cargo/bin/tool` → symlink at `/usr/local/bin/tool`

**Used by**: fd, ripgrep, bandwhich, dust, choose, zoxide, tealdeer, sd

**Rationale**: 
- Leverages Cargo's built-in installation system
- Automatic dependency resolution
- Easy updates via `cargo install --force`
- Creates consistent symlink structure for system-wide access

**Detection**: Binary in `~/.cargo/bin/` AND `/usr/local/bin/` is symlink

### 2. `direct_copy` Pattern (Build + Copy)
**Method**: Build locally → `sudo cp target/release/tool /usr/local/bin/`

**Used by**: 
- **Rust tools**: uv (source builds), difftastic, procs, bat, fclones, tokei, xsv, hyperfine, yazi, delta, starship, bottom, eza, ruff
- **Go tools**: lazygit, gh, fzf  
- **C/C++ tools**: 7zip, jq, imagemagick, ffmpeg

**Rationale**:
- **Rust tools**: Complex build requirements not compatible with `cargo install`
- **Go tools**: No equivalent to `cargo install` in Go ecosystem
- **C/C++ tools**: Traditional make/cmake build systems
- **Multiple binaries**: Tools like yazi that install multiple executables
- **Custom optimizations**: Target-specific builds (e.g., `-C target-cpu=native`)

**Detection**: Binary exists directly in `/usr/local/bin/` (not a symlink)

### 3. `official_installer` Pattern
**Method**: Vendor-provided installer → `~/.local/bin/tool`

**Used by**: uv (default method)

**Rationale**:
- Faster installation (prebuilt binaries)
- Vendor-maintained and tested
- Automatic updates via vendor mechanisms
- Follows XDG standards (`~/.local/bin/`)

**Detection**: Binary exists in `~/.local/bin/tool`

### 4. Language-Specific Package Managers
**Method**: Native package manager installation

**Used by**: serena (pip install in virtual environment)

**Rationale**:
- Leverages language ecosystem best practices
- Automatic dependency resolution within language
- Virtual environment isolation for Python tools

## Decision Criteria

### Choose `cargo_install` when:
- ✅ Tool is published to crates.io
- ✅ Standard Rust project structure
- ✅ No complex external dependencies
- ✅ Single binary output
- ✅ No custom build requirements

### Choose `direct_copy` when:
- ✅ Tool requires custom build flags
- ✅ Multiple binaries need installation
- ✅ Complex dependency requirements
- ✅ Language lacks standard installation system (Go, C/C++)
- ✅ Need maximum performance optimizations

### Choose `official_installer` when:
- ✅ Vendor provides reliable installer
- ✅ Prebuilt binaries available
- ✅ Installer handles system integration
- ✅ Performance/speed is priority over customization

### Choose `language_package_manager` when:
- ✅ Tool is distributed via language ecosystem (PyPI, npm, etc.)
- ✅ Complex language-specific dependencies
- ✅ Virtual environment isolation beneficial

## Anti-Patterns to Avoid

### ❌ Fallback Methods
```bash
# WRONG: Hides real errors
cargo install tool || sudo cp target/release/tool /usr/local/bin/
```

### ❌ Error Suppression
```bash
# WRONG: Masks installation failures  
pip install package || warning "Installation failed, but continuing"
```

### ✅ Fail-Fast Approach
```bash
# CORRECT: Clear error with helpful guidance
cargo install tool || error "Installation failed - check network connectivity and Rust toolchain"
```

## Method Consistency Rules

1. **One method per tool**: Each tool should use exactly one installation method
2. **Explicit method selection**: Method choice should be explicit, not fallback-based
3. **Clear error messages**: Failures should provide actionable guidance
4. **Documentation**: Method choice should be documented and justified

## Migration Guidelines

When changing installation methods:

1. **Update cleanup detection**: Ensure cleanup system recognizes the new pattern
2. **Test installation**: Verify the new method works reliably
3. **Update documentation**: Document the change and rationale
4. **Consider existing users**: Provide migration path if needed