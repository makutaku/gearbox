# Contributing to Essential Tools Installer

Thank you for your interest in contributing! This project helps developers set up essential command-line tools efficiently, and we welcome contributions of all kinds.

## Ways to Contribute

- **Add new tools** - Expand the collection with additional essential tools
- **Improve existing scripts** - Enhance build processes, add features, fix bugs  
- **Improve documentation** - Help make the project more accessible
- **Report issues** - Help us identify problems and improvement opportunities
- **Share feedback** - Tell us about your experience using the installer

## Quick Start for Contributors

### 1. Get Set Up

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/yourusername/gearbox.git
cd gearbox

# Test the current setup
./tests/test-runner.sh
gearbox install --minimal fd  # Try a simple installation
```

### 2. Understand the Project

- **For users**: Read the [User Guide](docs/USER_GUIDE.md) to understand what we're building
- **For developers**: Read the [Developer Guide](docs/DEVELOPER_GUIDE.md) for architecture and technical details

### 3. Make Your Contribution

**Adding a new tool?** The [Developer Guide](docs/DEVELOPER_GUIDE.md) has a complete step-by-step process with templates and examples.

**Other improvements?** Check existing patterns in the codebase and follow the same conventions.

### 4. Test Your Changes

```bash
# Basic validation
./tests/test-runner.sh

# Test your specific changes
gearbox install your-tool --run-tests

# Test in a clean environment (recommended)
docker run -it debian:bookworm bash
```

### 5. Submit Your Contribution

```bash
# Create a feature branch
git checkout -b your-feature-name

# Make your changes and commit
git add .
git commit -m "Brief description of changes"

# Push and create a pull request
git push origin your-feature-name
```

## What We Look For

- **Follows existing patterns** - Look at current scripts for examples
- **Well tested** - Works on clean systems and different scenarios
- **Good documentation** - Help others understand your contribution
- **User-focused** - Makes the installation experience better

## Getting Help

- **Questions about contributing?** Open an issue with the "question" label
- **Need help with the architecture?** Check the [Developer Guide](docs/DEVELOPER_GUIDE.md)
- **Found a bug?** Open an issue with detailed reproduction steps
- **Have an idea?** Open an issue to discuss it before implementing

## Community Guidelines

- **Be respectful** - We welcome contributors of all experience levels
- **Be collaborative** - Help others and ask for help when needed
- **Be patient** - Reviews and discussions take time
- **Be constructive** - Focus on making the project better

## Recognition

Contributors are recognized in several ways:
- Listed in project contributors
- Mentioned in release notes for significant contributions
- Credited in commit messages and pull requests

## Technical Details

For complete technical information including:
- Project architecture and design principles
- Step-by-step guide for adding new tools
- Coding standards and testing requirements
- Build system details and advanced topics

**See the [Developer Guide](docs/DEVELOPER_GUIDE.md)**

---

We appreciate your interest in making development environment setup easier for everyone! 🚀