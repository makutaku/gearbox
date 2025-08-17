# Essential Tools Installer

Automated installation scripts for essential command-line tools on Debian Linux. Build from source with optimized configurations and proper dependency management.

## Why This Project?

- **Source builds** with optimized configurations for best performance
- **Smart dependency management** - install once, share across tools
- **Multiple build types** - minimal, standard, or maximum feature sets
- **Battle-tested** - idempotent, safe, non-root installations

## Quick Start

```bash
# See what tools are available
gearbox list

# Install specific tools (recommended)
gearbox install fd ripgrep fzf jq

# Install all tools (with confirmation prompt)
gearbox install

# Fast builds for quick setup
gearbox install --minimal fd ripgrep
```

## Available Tools

| Tool | Description | Key Features |
|------|-------------|--------------|
| **fd** | Fast file finder | Intuitive syntax, parallel search |
| **ripgrep** | Fast text search | PCRE2 support, multi-line search |
| **fzf** | Fuzzy finder | Shell integration, preview support |
| **jq** | JSON processor | JSONPath queries, streaming |
| **zoxide** | Smart cd command | Frecency-based navigation |
| **yazi** | Terminal file manager | Vim-like keys, preview support |
| **fclones** | Duplicate file finder | Efficient deduplication, link replacement |
| **serena** | Coding agent toolkit | Semantic retrieval, IDE-like analysis |
| **uv** | Python package manager | 10-100x faster than pip, unified tooling |
| **ruff** | Python linter & formatter | 10-100x faster than Flake8/Black, 800+ rules |
| **bat** | Enhanced cat with syntax highlighting | Git integration, themes, automatic paging |
| **starship** | Customizable shell prompt | Fast, contextual info, Nerd Font support |
| **eza** | Modern ls replacement | Git integration, tree view, enhanced colors |
| **delta** | Syntax-highlighting pager | Git diff enhancement, word-level highlighting |
| **lazygit** | Terminal UI for Git | Interactive Git operations, visual interface |
| **bottom** | Cross-platform system monitor | Beautiful terminal resource monitoring |
| **procs** | Modern ps replacement | Enhanced process info, tree view, colors |
| **tokei** | Code statistics tool | Fast line counting for 200+ languages |
| **hyperfine** | Command-line benchmarking | Statistical command execution analysis |
| **ffmpeg** | Video/audio processing | Comprehensive codec support |
| **imagemagick** | Image manipulation | Powerful processing toolkit |
| **7zip** | Compression tool | High compression ratios |

## Documentation

### For Users
📖 **[User Guide](docs/USER_GUIDE.md)** - Complete installation guide, usage examples, and troubleshooting

### For Contributors  
🛠 **[Developer Guide](docs/DEVELOPER_GUIDE.md)** - Architecture, adding tools, and development guidelines

👥 **[Contributing](CONTRIBUTING.md)** - Quick start for contributors

## Key Features

- **Three build types**: Minimal (fast) → Standard (balanced) → Maximum (full-featured)
- **Optimal installation order**: Shared toolchains (Go → Rust → C/C++)
- **Safety first**: Non-root execution, existing installation detection
- **Shell integration**: Automatic setup for fzf key bindings
- **Clean architecture**: Source builds in `~/tools/build/`, binaries in `/usr/local/bin/`

## Requirements

- Debian Linux (or derivatives)
- Internet connection for downloading sources
- `sudo` access for system package installation

---

**Ready to get started?** See the [User Guide](docs/USER_GUIDE.md) for comprehensive installation instructions and examples.