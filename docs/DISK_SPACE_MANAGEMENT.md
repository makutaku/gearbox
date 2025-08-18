# Disk Space Management Guide

This guide explains how to effectively manage disk space when building tools from source with gearbox.

## The Challenge

When building 30+ tools from source, build artifacts accumulate rapidly:
- **Large build directories**: Rust projects can use 100-500MB+ per tool
- **Multiple build types**: Debug, release, and optimized builds
- **Preserved source code**: Git repositories with full history
- **Total impact**: 8GB+ of build artifacts is common

## Disk Space Optimization System

Gearbox provides a comprehensive cleanup system that safely removes build artifacts while preserving installed binaries.

### Quick Start

```bash
# Show disk usage report
gearbox doctor cleanup

# Clean specific tools
gearbox doctor cleanup bottom ripgrep fzf

# Clean all tools
gearbox doctor cleanup --all

# Preview cleanup (dry run)
gearbox doctor cleanup --all --dry-run

# Aggressive cleanup for maximum space savings
gearbox doctor cleanup --all --mode aggressive
```

## Cleanup Modes

### 1. Minimal Mode (`--mode minimal`)
- **Target**: Temporary files, logs, cache files
- **Space savings**: ~5% reduction
- **Safety**: Maximum - preserves all build artifacts
- **Use case**: Quick cleanup of obviously safe files

```bash
gearbox doctor cleanup --mode minimal ripgrep
```

### 2. Standard Mode (default)
- **Target**: Intermediate build artifacts, keep source
- **Space savings**: ~50% reduction  
- **Safety**: High - preserves source code and essential files
- **Use case**: Regular cleanup after installations

```bash
gearbox doctor cleanup ripgrep
# or explicitly:
gearbox doctor cleanup --mode standard ripgrep
```

### 3. Aggressive Mode (`--mode aggressive`)
- **Target**: Everything except source files (if preserved)
- **Space savings**: ~90% reduction
- **Safety**: Moderate - requires rebuild for future changes
- **Use case**: Maximum space recovery

```bash
gearbox doctor cleanup --mode aggressive --all
```

## Installation Pattern Detection

The cleanup system intelligently detects how tools are installed to avoid removing essential files:

### Pattern 1: Cargo Install (Rust tools)
```
~/.cargo/bin/tool  →  /usr/local/bin/tool (symlink)
```
**Tools**: fd, ripgrep, bandwhich, dust, choose, zoxide, tealdeer, sd

### Pattern 2: Direct Copy
```
/usr/local/bin/tool (direct binary)
```
**Tools**: bottom (btm), difftastic (difft), yazi, delta, starship, etc.

### Pattern 3: Official Installer  
```
~/.local/bin/tool
```
**Tools**: uv (when using official installer)

## Binary Name Resolution

The system automatically handles tools with different binary names using metadata:

| Tool | Binary | Detection Method |
|------|--------|------------------|
| bottom | btm | Metadata-driven |
| ripgrep | rg | Metadata-driven |
| difftastic | difft | Metadata-driven |
| tealdeer | tldr | Metadata-driven |

This eliminates hardcoded mappings and supports new tools automatically.

## Build Cache System

Gearbox includes an intelligent build cache to speed up reinstallations:

### How It Works
1. **Successful builds** are automatically cached by tool and build type
2. **Cache validation** uses SHA256 checksums for integrity
3. **Smart retrieval** reuses cached binaries when possible
4. **Cleanup preservation** - cache is preserved during cleanup

### Cache Management
```bash
# View cache statistics
gearbox doctor cleanup  # Shows cache info in report

# Cache is automatically managed
# Location: ~/tools/cache/builds/
```

## Automated Cleanup

### Enable Auto-Cleanup
```bash
# Enable automatic cleanup after installations
gearbox doctor cleanup --auto-cleanup

# Set default cleanup mode
export GEARBOX_AUTO_CLEANUP=true
export GEARBOX_CLEANUP_MODE=standard
```

### Configuration Options
```bash
# In ~/.bashrc or ~/.zshrc:
export GEARBOX_AUTO_CLEANUP=true        # Enable auto-cleanup
export GEARBOX_CLEANUP_MODE=standard    # Set default mode
export GEARBOX_PRESERVE_SOURCE=true     # Keep source in aggressive mode
```

## Disk Usage Monitoring

### Regular Health Checks
```bash
# Comprehensive disk usage report
gearbox doctor cleanup

# Example output:
# Build Directory (/home/user/tools/build):
#   - Total size: 8.1G
#   - Tool directories: 21
#   - Largest builds:
#     ruff: 1.4G
#     yazi: 1.1G
#     bottom: 455M
```

### Cleanup Recommendations
The system provides intelligent recommendations:
- **Large directories** (>1GB): Suggests cleanup
- **Unused tools**: Identifies candidates for aggressive cleanup
- **Cache efficiency**: Shows cache hit rates and storage

## Best Practices

### 1. Regular Maintenance
```bash
# Monthly cleanup routine
gearbox doctor cleanup --all --mode standard
```

### 2. Strategic Tool Management
```bash
# Clean large tools you don't frequently modify
gearbox doctor cleanup --mode aggressive ruff yazi

# Keep development tools in standard mode
gearbox doctor cleanup --mode standard fd ripgrep
```

### 3. Pre-Installation Planning
```bash
# Check space before installing many tools
gearbox doctor cleanup

# Enable auto-cleanup for new installations
export GEARBOX_AUTO_CLEANUP=true
```

### 4. Safe Cleanup Workflow
```bash
# Always dry-run first for aggressive cleanup
gearbox doctor cleanup --all --mode aggressive --dry-run

# Review the output, then execute
gearbox doctor cleanup --all --mode aggressive
```

## Troubleshooting

### Tool Not Properly Installed
If cleanup reports a tool as "not properly installed":

1. **Check the binary name** in `config/tools.json`
2. **Verify installation** with `which <binary_name>`
3. **Check installation pattern** - may be in unexpected location

```bash
# Example: bottom installs as 'btm'
which btm  # Should show /usr/local/bin/btm
```

### Cache Issues
If cache operations fail:

1. **Check permissions** on `~/tools/cache/`
2. **Verify disk space** for cache storage
3. **Clear corrupted cache** if needed:
   ```bash
   rm -rf ~/tools/cache/builds/corrupted-tool-*
   ```

### Summary Reporting Issues
If cleanup reports incorrect space freed:

1. **Check shell output** for actual freed space
2. **Verify parsing** - should match "Cleaned XMB" format
3. **Report pattern mismatches** for improvement

## Performance Impact

### Build Times
- **Standard cleanup**: No impact on build times
- **Aggressive cleanup**: Requires full rebuild for source changes
- **Cache system**: Significantly reduces reinstallation time

### Disk I/O
- **Cleanup operations**: Brief I/O spike during deletion
- **Cache operations**: Minimal ongoing I/O overhead
- **Monitoring**: Negligible performance impact

## Security Considerations

### Safe Deletion
- **Preserved binaries**: Installed tools remain functional
- **Source protection**: Git repositories preserved by default
- **Permission checks**: Cleanup respects file permissions

### Access Control
- **User-space only**: No system-wide cleanup operations
- **Explicit targeting**: Only specified tools are cleaned
- **Confirmation prompts**: Aggressive operations show warnings

This comprehensive disk space management system ensures you can build from source efficiently while maintaining a clean development environment.