# Gearbox TUI (Text User Interface) Guide

## Overview

Gearbox TUI provides a rich, interactive terminal interface for discovering, installing, and managing development tools. Built with the modern Bubble Tea framework, it offers an intuitive and beautiful experience that complements the traditional CLI.

## Features

### 🎯 Key Features

- **Visual Tool Browsing**: Explore 42+ tools with search and filtering
- **Bundle Management**: Discover and install curated tool collections
- **Real-time Progress**: Watch installations with live output
- **System Overview**: Monitor tool installations and system health
- **Keyboard Navigation**: Efficient navigation with vim-style keys
- **Beautiful Design**: Modern UI with consistent theming

### 📊 Views

1. **Dashboard**: System overview, recent activity, and recommendations
2. **Tool Browser**: Browse and search all available tools
3. **Bundle Explorer**: Explore curated bundles with dependency visualization
4. **Install Manager**: Manage installation queue and monitor progress
5. **Configuration**: Interactive settings management
6. **Health Monitor**: System diagnostics and troubleshooting
7. **Help**: Comprehensive keyboard shortcuts and documentation

## Getting Started

### Launch the TUI

```bash
# From the project directory
./build/gearbox tui

# Or if installed system-wide
gearbox tui
```

### Navigation

#### Global Keys
- `Tab` - Cycle through views
- `q` or `Ctrl+C` - Quit
- `?` - Show help
- `/` - Search (context-sensitive)
- `Esc` - Go back/Cancel

#### Quick Navigation
- `D` - Dashboard
- `T` - Tool Browser
- `B` - Bundle Explorer
- `I` - Install Manager
- `C` - Configuration
- `H` - Health Monitor

#### List Navigation
- `↑/k` - Move up
- `↓/j` - Move down
- `←/h` - Move left
- `→/l` - Move right
- `Enter` - Select/Confirm
- `Space` - Toggle selection

## User Workflows

### 1. First-Time User Experience

1. Launch TUI: `gearbox tui`
2. Dashboard shows system status and recommendations
3. Press `B` to explore bundles
4. Navigate to "beginner" bundle
5. Press `Enter` to view bundle details
6. Press `i` to install the bundle
7. Watch real-time progress in Install Manager

### 2. Installing Individual Tools

1. Press `T` for Tool Browser
2. Press `/` to search
3. Type tool name (e.g., "ripgrep")
4. Navigate to tool with arrow keys
5. Press `Space` to select multiple tools
6. Press `i` to add to install queue
7. Press `I` to view Install Manager
8. Confirm and start installation

### 3. Power User Batch Installation

1. Press `T` for Tool Browser
2. Use `/` to filter by category or language
3. Use `Space` to multi-select tools
4. Press `i` to queue installations
5. Press `I` to manage queue
6. Adjust build types if needed
7. Start batch installation

## Dashboard Features

### System Status
- **Tools Installed**: Shows X/42 tools installed
- **Bundles Active**: Counts fully installed bundles
- **Disk Usage**: Estimates total disk usage
- **Health Status**: Quick system health indicator

### Quick Actions
- `[i]` Install Tools - Jump to Tool Browser
- `[b]` Browse Bundles - Open Bundle Explorer
- `[c]` Configuration - Manage settings
- `[h]` Health Check - Run diagnostics

### Recent Activity
Displays recent tool installations with timestamps:
- Shows last 5 installations
- Time-relative display (e.g., "2 hours ago")
- Success/failure indicators

### Smart Recommendations
Intelligent suggestions based on your setup:
- Cross-tool dependencies (e.g., starship ↔ nerd-fonts)
- Beginner bundle for new users
- Complementary tool suggestions

## Tool Browser Features

### Search & Filter
- Real-time search across tool names and descriptions
- Filter by category (Core, Development, System, etc.)
- Filter by language (Rust, Go, C/C++, Python)
- Filter by installation status

### Tool Preview
When selecting a tool, see:
- Full description
- Key features
- Build type options (minimal/standard/maximum)
- Dependencies
- Estimated size
- Installation status

### Batch Operations
- Select multiple tools with `Space`
- Install selected tools with `i`
- View details with `Tab`
- Clear selection with `Esc`

## Bundle Explorer Features

### Bundle Hierarchy
Organized by tiers:
- **Foundation**: Beginner → Intermediate → Advanced
- **Domain**: Fullstack, Mobile, Data Science, DevOps, etc.
- **Language**: Python, Node.js, Rust, Go ecosystems
- **Workflow**: Debugging, Deployment, Code Review tools
- **Infrastructure**: Docker, Cloud, Database tools

### Bundle Details
For each bundle, view:
- Complete tool list
- Installation progress
- System package requirements
- Bundle dependencies
- Total size estimate

### Smart Installation
- Dependency resolution
- Optimal installation order
- Conflict detection
- Progress tracking

## Configuration Panel

### Settings Management
- Default build type (minimal/standard/maximum)
- Parallel job limits
- Cache management
- Shell integration options
- Theme selection

### Interactive Configuration
- Visual setting editors
- Real-time validation
- Reset to defaults option
- Export/Import settings

## Tips & Best Practices

### Performance Tips
1. Use search (`/`) to quickly find tools
2. Multi-select with `Space` for batch operations
3. Use keyboard shortcuts for faster navigation
4. Keep Install Manager open during long installations

### Installation Strategy
1. Start with a foundation bundle (beginner/intermediate)
2. Add domain-specific bundles based on your work
3. Install individual tools as needed
4. Use `--minimal` for faster builds when testing

### Troubleshooting
1. If TUI seems frozen, check Install Manager for active tasks
2. Use `H` to access Health Monitor for diagnostics
3. Press `?` anytime for context-sensitive help
4. Exit and use CLI for debugging if needed

## Advanced Features

### Keyboard Customization
Future versions will support:
- Custom key bindings
- Vim/Emacs mode selection
- Macro recording

### Theme System
- Default theme (balanced colors)
- Dark theme (high contrast)
- Light theme (for bright terminals)
- Custom theme creation

### Integration Features
- Shell completion updates
- Git hooks for tracking
- Export installation scripts
- Bundle sharing

## Technical Details

### Architecture
- Built with Bubble Tea (Elm architecture)
- Lipgloss for styling
- Concurrent task management
- Memory-efficient rendering

### Requirements
- Terminal with Unicode support
- 256-color terminal recommended
- Minimum 80x24 terminal size
- Interactive terminal (no pipe/redirect)

## Future Enhancements

### Planned Features
1. **Tool Browser**
   - Screenshot previews
   - Dependency graphs
   - Size calculations
   
2. **Installation Manager**
   - Pause/Resume support
   - Installation scheduling
   - Bandwidth limiting
   
3. **Bundle Creator**
   - Custom bundle builder
   - Bundle sharing hub
   - Template system

4. **Advanced Search**
   - Fuzzy search
   - Regular expressions
   - Saved searches

## Troubleshooting

### Common Issues

**TUI won't start**
- Ensure terminal is interactive: `gearbox tui`
- Check terminal capabilities: `echo $TERM`
- Try different terminal emulator

**Display issues**
- Ensure Unicode support: `locale`
- Check terminal size: `tput cols && tput lines`
- Try different theme: Future feature

**Navigation problems**
- Check for key binding conflicts
- Try using letter keys instead of arrows
- Disable terminal mouse mode if active

### Debug Mode
Future versions will support:
```bash
gearbox tui --debug
```

## Contributing

The TUI is under active development. To contribute:

1. Check `cmd/gearbox/tui/` for source code
2. Follow Bubble Tea patterns
3. Test with various terminal emulators
4. Submit PRs with screenshots

## Summary

Gearbox TUI transforms tool management into an enjoyable, efficient experience. Whether you're a beginner exploring available tools or a power user managing complex installations, the TUI provides the perfect balance of functionality and usability.

For traditional CLI usage, run `gearbox --help` or see the main documentation.