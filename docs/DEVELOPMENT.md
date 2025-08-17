# Development Guide

## Repository Structure
- `scripts/` - All installation scripts
- `docs/` - Documentation
- `examples/` - Usage examples
- `tests/` - Test scripts
- `config.sh` - Shared configuration

## Adding New Tools
1. Create `scripts/install-newtool.sh`
2. Follow the existing script pattern
3. Add to `install-all-tools.sh`
4. Update documentation

## Build Directory
All source repositories are cloned to `~/tools/build/` by default.
This keeps the scripts repository clean and separate.
