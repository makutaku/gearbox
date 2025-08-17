# Development Tools Installation Scripts

Automated installation scripts for essential development tools on Debian Linux.

## 🚀 Quick Start

### Install Everything (Recommended)
```bash
# Install all tools with optimal dependency management
./install-all-tools.sh
```

### Install Common Dependencies First
```bash
# Install shared dependencies once (Rust, Go, build tools)
./install-common-deps.sh

# Then install individual tools with --skip-deps
./install-fd.sh --skip-deps
./install-ripgrep.sh --skip-deps
./install-fzf.sh --skip-deps
```

## 📦 Available Tools

| Tool | Language | Description | Script |
|------|----------|-------------|---------|
| **FFmpeg** | C/C++ | Video/audio processing suite | `install-ffmpeg.sh` |
| **7-Zip** | C/C++ | High-ratio compression tool | `install-7zip.sh` |
| **jq** | C | JSON processor and query tool | `install-jq.sh` |
| **fd** | Rust | Fast alternative to `find` | `install-fd.sh` |
| **ripgrep** | Rust | Fast text search tool | `install-ripgrep.sh` |
| **fzf** | Go | Fuzzy finder for files/commands | `install-fzf.sh` |

## 🛠 Build Types

### FFmpeg
- `-m, --minimal`: Basic codecs only
- `-g, --general`: Standard codec support (default)
- `-x, --maximum`: All available codecs

### 7-Zip
- `-b, --basic`: Basic build
- `-o, --optimized`: Optimized build (default)
- `-a, --asm-optimized`: Assembly optimized

### jq
- `-m, --minimal`: Minimal features
- `-s, --standard`: Standard build (default)
- `-o, --optimized`: Performance optimized

### fd & ripgrep
- `-d, --debug`: Debug build
- `-r, --release`: Release build (default)
- `-o, --optimized`: CPU optimized

### fzf
- `-s, --standard`: Standard build (default)
- `-p, --profiling`: Profiling build

## 🎯 Usage Examples

### Individual Tools
```bash
# Standard installations
./install-fd.sh                    # Fast file finder
./install-ripgrep.sh               # Text search with PCRE2
./install-fzf.sh                   # Fuzzy finder with shell integration

# Optimized builds
./install-ffmpeg.sh --maximum      # All codecs
./install-ripgrep.sh --optimized   # CPU-optimized
./install-7zip.sh --asm-optimized  # Assembly optimized

# With testing
./install-jq.sh --run-tests        # Run test suite
./install-fzf.sh --run-tests       # Test Go and Ruby integration
```

### Coordinated Installation
```bash
# Install specific tools with minimal builds
./install-all-tools.sh --minimal fd ripgrep fzf

# Install all tools with maximum features and testing
./install-all-tools.sh --maximum --run-tests

# Install without shell integration
./install-all-tools.sh --no-shell
```

## 🔧 Dependencies Handled

### Programming Languages
- **Rust ≥ 1.88.0** (for fd, ripgrep)
- **Go ≥ 1.20** (for fzf)  
- **C/C++ toolchain** (for ffmpeg, 7zip, jq)

### System Packages
- `build-essential`, `git`, `curl`, `wget`, `make`
- `autoconf`, `automake`, `libtool` (for autotools projects)
- `nasm`, `yasm` (for assembly optimizations)
- Codec libraries (for ffmpeg)

## 🎨 Shell Integration

### fzf Automatic Setup
- **Bash**: `eval "$(fzf --bash)"`
- **Zsh**: `source <(fzf --zsh)`
- **Fish**: `fzf --fish | source`

### Key Bindings Added
- **CTRL-T**: File selection
- **CTRL-R**: Command history search  
- **ALT-C**: Directory navigation

## 📁 Installation Paths

| Tool | Binary Location | Additional |
|------|----------------|------------|
| ffmpeg | `/usr/local/bin/` | ffprobe, ffplay |
| 7zz | `/usr/local/bin/` | - |
| jq | `/usr/local/bin/` | libraries in `/usr/local/lib/` |
| fd | `~/.cargo/bin/` → `/usr/local/bin/` | symlink as `fdfind` too |
| rg | `~/.cargo/bin/` → `/usr/local/bin/` | - |
| fzf | `/usr/local/bin/` | shell integration files |

## 🛡 Safety Features

- ✅ Rust version conflict resolution (unified to 1.88.0)
- ✅ Idempotent git repository handling
- ✅ Existing installation detection
- ✅ Library path management (`ldconfig`)
- ✅ Command hash clearing (`hash -r`)
- ✅ Non-root execution enforcement

## 🔄 Installation Order

Optimal order handled automatically by `install-all-tools.sh`:

1. **fzf** (installs Go)
2. **ripgrep** (installs Rust 1.88.0+)
3. **fd** (uses existing Rust)
4. **jq** (independent)
5. **ffmpeg** (independent)
6. **7zip** (independent)

## 🆘 Troubleshooting

### Command not found after installation
```bash
# Clear hash table and reload shell
hash -r && source ~/.bashrc
```

### Wrong version showing
```bash
# Check which binary is being used
which command_name
type command_name

# Clear hash and try again
hash -r
command_name --version
```

### Rust version conflicts
```bash
# Update to latest Rust (handles all tools)
rustup update
```

## 📝 Notes

- All scripts support `--skip-deps`, `--run-tests`, `--force` flags
- Scripts are idempotent and can be run multiple times safely
- PATH modifications are added to `~/.bashrc`
- Library cache is updated automatically (`ldconfig`)
- Open a new terminal or `source ~/.bashrc` after installation

---

Built with ❤️ for efficient development environment setup.