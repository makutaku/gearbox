# Quick Start Guide

## Installation Methods

### 🚀 One-Command Installation (Recommended)
Install all tools with optimal dependency management:
```bash
./install-tools
```

### 🎯 Selective Installation
Install only specific tools:
```bash
./install-tools fd ripgrep fzf
./install-tools jq ffmpeg
./install-tools 7zip
```

### ⚡ Build Type Options
```bash
./install-tools --minimal     # Fast, basic builds
./install-tools --maximum     # Full-featured builds with all optimizations
./install-tools --run-tests   # Include test suites for validation
```

### 🔧 Advanced Options
```bash
# Skip dependency installation (if already installed)
./install-tools --skip-common-deps fd ripgrep

# Install without shell integration
./install-tools --no-shell fzf

# Combine options
./install-tools --minimal --run-tests fd ripgrep jq
```

## Individual Tool Installation

### Basic Usage
Each tool can be installed individually:
```bash
# Standard installations
scripts/install-fd.sh           # Fast file finder
scripts/install-ripgrep.sh      # Text search with PCRE2
scripts/install-fzf.sh          # Fuzzy finder with shell integration
scripts/install-jq.sh           # JSON processor
scripts/install-ffmpeg.sh       # Video/audio processing
scripts/install-7zip.sh         # Compression tool
```

### Build Type Examples
```bash
# Minimal builds (fastest)
scripts/install-fd.sh --minimal
scripts/install-ripgrep.sh --no-pcre2
scripts/install-ffmpeg.sh --minimal

# Optimized builds (best performance)
scripts/install-fd.sh --release
scripts/install-ripgrep.sh --optimized
scripts/install-ffmpeg.sh --maximum

# Debug builds (for development)
scripts/install-jq.sh --debug
scripts/install-fzf.sh --debug
```

### Advanced Individual Options
```bash
# Skip dependencies (useful after first tool)
scripts/install-fd.sh --skip-deps --release

# Run tests after installation
scripts/install-jq.sh --run-tests

# Force reinstallation
scripts/install-ripgrep.sh --force --optimized

# No shell integration for fzf
scripts/install-fzf.sh --no-shell
```

## Common Workflow Examples

### Development Environment Setup
```bash
# Essential development tools
./install-tools fd ripgrep fzf jq

# With testing validation
./install-tools --run-tests fd ripgrep fzf jq
```

### Media Processing Setup
```bash
# Media tools with maximum features
./install-tools --maximum ffmpeg 7zip

# Just ffmpeg with specific codec support
scripts/install-ffmpeg.sh --maximum --run-tests
```

### Minimal Installation
```bash
# Fast installation for CI/containers
./install-tools --minimal --skip-common-deps fd ripgrep
```

## Verification and Troubleshooting

### Verify Installation
```bash
# Check if tools are available
which fd rg fzf jq ffmpeg 7zz

# Check versions
fd --version
rg --version
fzf --version
jq --version
ffmpeg -version
7zz
```

### Refresh Shell Environment
After installation, refresh your shell:
```bash
# Clear command hash and reload shell configuration
hash -r && source ~/.bashrc

# Or open a new terminal session
```

### Common Issues
```bash
# Command not found after installation
hash -r
source ~/.bashrc

# Check which binary is being used
type fd
which ripgrep

# Verify PATH includes /usr/local/bin
echo $PATH | tr ':' '\n' | grep local
```

## Directory Structure
- **Source repositories**: `~/tools/build/` (temporary build files)
- **Scripts repository**: `~/gearbox/` (this repository)
- **Binaries installed**: `/usr/local/bin/` (added to PATH)
- **Cache directory**: `~/tools/cache/` (downloads and temporary files)

## Shell Integration Features

### fzf Integration (Automatic)
After installing fzf, these key bindings are available:
- **CTRL-T**: File selection in current directory
- **CTRL-R**: Command history search
- **ALT-C**: Directory navigation

### Manual Shell Setup
If shell integration doesn't work automatically:
```bash
# Bash
echo 'eval "$(fzf --bash)"' >> ~/.bashrc

# Zsh  
echo 'source <(fzf --zsh)' >> ~/.zshrc

# Fish
echo 'fzf --fish | source' >> ~/.config/fish/config.fish
```

## Next Steps

After installation:
1. Open a new terminal or run `source ~/.bashrc`
2. Try the installed tools:
   ```bash
   fd . --type f                    # Find files
   rg "pattern" --type rust         # Search in Rust files
   echo '{"name": "test"}' | jq .   # Process JSON
   fzf                              # Interactive fuzzy finder
   ```
3. Read tool-specific documentation for advanced usage
4. Customize shell aliases and functions as needed
