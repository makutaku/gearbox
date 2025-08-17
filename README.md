# Essential Tools Installer

Automated installation scripts for essential command-line tools on Debian Linux. Build from source with optimized configurations and proper dependency management.

## 🚀 Quick Start

### One-Command Installation
```bash
./install-tools
```

### Selective Installation
```bash
./install-tools fd ripgrep fzf
```

### Different Build Types
```bash
./install-tools --minimal     # Fast, basic builds
./install-tools --maximum     # Full-featured builds
```

## 📦 Available Tools

| Tool | Description | Language | Key Features |
|------|-------------|----------|--------------|
| **fd** | Fast file finder | Rust | Intuitive syntax, parallel search |
| **ripgrep** | Fast text search | Rust | PCRE2 support, multi-line search |
| **fzf** | Fuzzy finder | Go | Shell integration, preview support |
| **jq** | JSON processor | C | JSONPath queries, streaming |
| **ffmpeg** | Video/audio processing | C/C++ | Comprehensive codec support |
| **7zip** | Compression tool | C/C++ | High compression ratios |

## 🛠 Build Options

Each tool supports multiple build types:
- **Minimal**: Fast builds with essential features
- **Standard**: Balanced performance and features (default)
- **Maximum**: Full-featured builds with all optimizations

## 🔧 Architecture

- **`config.sh`**: Shared configuration and utility functions
- **`install-tools`**: Main wrapper script
- **`scripts/`**: Individual installation scripts for each tool
- **Installation directories**:
  - Source builds: `~/tools/build/`
  - Binaries: `/usr/local/bin/`
  - Cache: `~/tools/cache/`

## 📚 Documentation

- [Quick Start Guide](docs/QUICK_START.md) - Get up and running fast
- [Development Guide](docs/DEVELOPMENT.md) - Architecture and contribution guidelines
- [Contributing](CONTRIBUTING.md) - How to add new tools

## 🛡 Safety Features

- Non-root execution enforcement
- Idempotent installations
- Proper dependency management
- Version conflict resolution
- Existing installation detection
