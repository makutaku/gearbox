# Troubleshooting Guide

This guide helps resolve common issues when using gearbox for tool installation and management.

## Installation Issues

### Error: "Tool not properly installed"

**Symptom**: Cleanup system reports tool as not properly installed despite successful installation.

**Diagnosis**:
```bash
# Check if binary exists
which <tool_name>

# Check binary name mapping
cat config/tools.json | grep -A 5 '"name": "<tool_name>"'

# Test cleanup detection
source lib/common.sh
get_binary_name_for_tool "<tool_name>"
```

**Solutions**:

1. **Binary name mismatch**: Tool installs with different binary name
   ```bash
   # Example: bottom installs as 'btm'
   which btm  # Should show /usr/local/bin/btm
   
   # Verify metadata is correct
   grep -A 10 '"name": "bottom"' config/tools.json
   ```

2. **Installation pattern not recognized**: 
   ```bash
   # Check installation location
   ls -la /usr/local/bin/<binary_name>
   ls -la ~/.cargo/bin/<binary_name>
   ls -la ~/.local/bin/<binary_name>
   ```

3. **Metadata missing**: Add tool to `config/tools.json`
   ```json
   {
     "name": "tool_name",
     "binary_name": "actual_binary_name",
     ...
   }
   ```

### Error: "Failed to upgrade pip tools"

**Symptom**: Python tool installation fails during pip upgrade.

**Diagnosis**:
```bash
# Check Python environment
python3 --version
pip --version

# Test network connectivity
curl -I https://pypi.org/
```

**Solutions**:

1. **Network connectivity**: Check firewall/proxy settings
2. **Python environment**: Ensure Python 3.11+ is available
3. **Virtual environment**: Recreate if corrupted
   ```bash
   rm -rf tool-venv
   python3 -m venv tool-venv
   ```

### Error: "cargo install fzf failed"

**Symptom**: Zoxide installation fails when trying to install fzf.

**Diagnosis**:
```bash
# Check if fzf is published to crates.io
cargo search fzf

# Check Rust toolchain
rustc --version
cargo --version
```

**Solutions**:

1. **Use dedicated script**: Install fzf separately
   ```bash
   ./install-zoxide.sh  # Skip fzf option
   ./install-fzf.sh     # Dedicated fzf installation
   ```

2. **Disable fzf integration**:
   ```bash
   ./install-zoxide.sh --no-shell
   ```

## Build Issues

### Error: "Rust version insufficient"

**Symptom**: Build fails with Rust version error.

**Diagnosis**:
```bash
rustc --version
rustup show
```

**Solutions**:

1. **Update Rust**:
   ```bash
   rustup update stable
   rustup default stable
   ```

2. **Install specific version**:
   ```bash
   rustup install 1.88.0
   rustup default 1.88.0
   ```

### Error: "Build artifacts not found"

**Symptom**: Installation completes but binary not found in expected location.

**Diagnosis**:
```bash
# Check build directory
ls -la ~/tools/build/<tool_name>/target/*/

# Check for build errors
./install-<tool>.sh --build-only
```

**Solutions**:

1. **Check build type**: Different build types use different directories
   ```bash
   # Debug builds
   ls target/debug/
   
   # Release builds  
   ls target/release/
   ```

2. **Verify build completion**:
   ```bash
   ./install-<tool>.sh --build-only
   echo $?  # Should be 0 for success
   ```

## Cleanup Issues

### Error: "Summary shows 0 B freed but shell shows MB freed"

**Symptom**: Cleanup works but summary reporting is incorrect.

**Diagnosis**:
```bash
# Check CLI version
./gearbox --version

# Test shell cleanup directly
source lib/common.sh
cleanup_build_artifacts "tool_name" "standard"
```

**Solutions**:

1. **Rebuild CLI**: Ensure latest version with parsing fixes
   ```bash
   make cli
   ```

2. **Check output format**: Ensure shell output matches expected patterns
   ```bash
   # Expected formats:
   # "Cleaned 123MB from tool"
   # "freed 123MB"
   ```

### Error: "Permission denied during cleanup"

**Symptom**: Cleanup fails with permission errors.

**Diagnosis**:
```bash
# Check directory permissions
ls -la ~/tools/build/
ls -la ~/tools/build/<tool_name>/

# Check ownership
stat ~/tools/build/<tool_name>/
```

**Solutions**:

1. **Fix ownership**:
   ```bash
   sudo chown -R $USER:$USER ~/tools/build/
   ```

2. **Fix permissions**:
   ```bash
   chmod -R u+w ~/tools/build/<tool_name>/
   ```

## Configuration Issues

### Error: "tools.json not found"

**Symptom**: Binary name detection fails due to missing configuration.

**Diagnosis**:
```bash
# Check config file location
ls -la config/tools.json
ls -la $REPO_DIR/config/tools.json

# Check current directory
pwd
```

**Solutions**:

1. **Run from correct directory**:
   ```bash
   cd /path/to/gearbox
   ./gearbox doctor cleanup tool_name
   ```

2. **Set REPO_DIR explicitly**:
   ```bash
   export REPO_DIR=/path/to/gearbox
   ```

### Error: "jq command not found"

**Symptom**: Binary name detection falls back to grep/sed parsing.

**This is not an error** - the system gracefully falls back. However, for best performance:

**Solution**:
```bash
# Install jq for better JSON parsing
sudo apt install jq
```

## Performance Issues

### Slow Build Times

**Symptom**: Builds take much longer than expected.

**Diagnosis**:
```bash
# Check system resources
htop
df -h
free -h

# Check parallel job settings
echo $MAX_PARALLEL_JOBS
```

**Solutions**:

1. **Optimize parallel jobs**:
   ```bash
   export MAX_PARALLEL_JOBS=4  # Adjust based on CPU/memory
   ```

2. **Use faster build types**:
   ```bash
   ./install-tool.sh -m  # Minimal/debug build
   ```

3. **Enable build cache**:
   ```bash
   export GEARBOX_CACHE_ENABLED=true
   ```

### High Disk Usage

**Symptom**: Build directories consume excessive disk space.

**Diagnosis**:
```bash
# Check disk usage
gearbox doctor cleanup

# Check largest directories
du -sh ~/tools/build/* | sort -hr | head -10
```

**Solutions**:

1. **Regular cleanup**:
   ```bash
   gearbox doctor cleanup --all --mode standard
   ```

2. **Enable auto-cleanup**:
   ```bash
   export GEARBOX_AUTO_CLEANUP=true
   ```

3. **Aggressive cleanup for unused tools**:
   ```bash
   gearbox doctor cleanup --mode aggressive unused_tool1 unused_tool2
   ```

## Network Issues

### Error: "Failed to clone repository"

**Symptom**: Git clone operations fail.

**Diagnosis**:
```bash
# Test connectivity
ping github.com
curl -I https://github.com/

# Check git configuration
git config --list | grep -E "(http|proxy)"
```

**Solutions**:

1. **Proxy configuration**:
   ```bash
   git config --global http.proxy http://proxy:port
   git config --global https.proxy https://proxy:port
   ```

2. **SSH vs HTTPS**:
   ```bash
   # Use HTTPS instead of SSH
   git config --global url."https://github.com/".insteadOf git@github.com:
   ```

3. **DNS issues**:
   ```bash
   # Try alternative DNS
   echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
   ```

## Environment Issues

### Error: "Command not found after installation"

**Symptom**: Tool installs successfully but command not available.

**Diagnosis**:
```bash
# Check PATH
echo $PATH

# Check binary location
find /usr/local/bin ~/.cargo/bin ~/.local/bin -name "<binary_name>" 2>/dev/null

# Check symlinks
ls -la /usr/local/bin/<binary_name>
```

**Solutions**:

1. **Update PATH**:
   ```bash
   export PATH="/usr/local/bin:$HOME/.cargo/bin:$HOME/.local/bin:$PATH"
   
   # Make permanent
   echo 'export PATH="/usr/local/bin:$HOME/.cargo/bin:$HOME/.local/bin:$PATH"' >> ~/.bashrc
   ```

2. **Refresh shell**:
   ```bash
   hash -r
   source ~/.bashrc
   ```

3. **Create missing symlink**:
   ```bash
   sudo ln -sf ~/.cargo/bin/<binary> /usr/local/bin/<binary>
   ```

## Getting Help

### Diagnostic Information

When reporting issues, include:

```bash
# System information
uname -a
lsb_release -a

# Tool versions
rustc --version
go version
python3 --version

# Gearbox environment
./gearbox --version
echo $REPO_DIR
ls -la config/tools.json

# Specific tool information
which <tool_name>
<tool_name> --version

# Recent logs
tail -20 ~/.local/share/gearbox/logs/install.log  # if exists
```

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Enable debug logging
export GEARBOX_DEBUG=true

# Verbose installation
./install-tool.sh --verbose

# Verbose cleanup
gearbox doctor cleanup --verbose tool_name
```

### Community Support

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Check latest docs for updates
- **Configuration Examples**: Review example configurations

This troubleshooting guide covers the most common issues. For complex problems, enable debug mode and gather diagnostic information before seeking help.