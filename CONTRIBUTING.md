# Contributing to Dev Tools Installer

## Development Setup
1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Test your changes
6. Submit a pull request

## Script Guidelines
- Follow the existing pattern
- Support `--skip-deps`, `--run-tests`, `--force` flags
- Use the shared configuration from `config.sh`
- Keep source builds in `$BUILD_DIR`
- Update documentation

## Testing
Run the test suite before submitting:
```bash
./tests/test-runner.sh
```
