# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **Interactive TUI (Text User Interface)** - A comprehensive visual interface for gearbox operations
  - Dashboard view with system statistics and recommendations
  - Tool Browser with real-time search and multi-selection
  - Bundle Explorer with hierarchical bundle organization
  - Install Manager with real-time progress tracking
  - Configuration view for interactive settings management
  - Health Monitor for system diagnostics
  - Full keyboard navigation and help system
  - Built with Bubble Tea framework for smooth, responsive UI
  - Launch with `gearbox tui` command
- **gopls (Go Language Server)** - Added to go-dev and intermediate bundles for improved Go development experience

### Enhanced
- Documentation updated to include TUI usage and development guidelines
- Added troubleshooting section for common TUI issues

### Fixed
- **Architecture Improvements**: Comprehensive refactoring to eliminate anti-patterns and enhance code quality
  - **Eliminated global variables**: Removed `globalConfig` and `bundleConfigs` anti-pattern
  - **Added ConfigManager**: Thread-safe configuration management with proper synchronization
  - **Implemented Builder pattern**: Replaced monolithic 89-line NewOrchestrator() constructor
  - **Added comprehensive cleanup**: Resource management with proper cleanup mechanisms
  - **Dynamic resource calculation**: Intelligent job limits based on system capabilities

### Security
- **Secure debug logging**: Moved from world-readable `/tmp/gearbox-debug.log` to user-only `~/.cache/gearbox/tui-debug.log` with 0600 permissions
- **Fixed debug code in production**: Replaced fmt.Printf debug statements with proper structured logging
- **Information disclosure prevention**: Secure debug file locations and access controls
- **Async tool detection**: Resolved inconsistent behavior in TUI Tools view

### Technical
- Added Bubble Tea dependencies (bubbletea, bubbles, lipgloss)
- Implemented Model-View-Update architecture for reactive UI
- Created modular view system with consistent theming
- Integrated background task management with channel-based updates
- Added TaskProvider interface for decoupling views from implementation
- **ConfigManager with RWMutex** for thread-safe configuration access
- **Builder pattern** for orchestrator construction with validation
- **Dynamic worker pool sizing** based on CPU count and memory
- **Modern Go practices**: Eliminated deprecated ioutil.* functions
- **Structured error handling**: Added context and suggestions to error messages

### Post-Refactoring Cleanup
- **Eliminated legacy wrapper**: Removed `NewOrchestrator()` function, replaced with direct Builder pattern usage
- **Fixed duplicate bundle loading**: Integrated real JSON parsing into Builder, removed placeholder implementation
- **Consistent debug logging**: Fixed inconsistent security between TUI debug files
- **Code consolidation**: Removed redundant functions and TODO placeholders from refactoring artifacts

## [Previous versions]

(Previous changelog entries would go here)