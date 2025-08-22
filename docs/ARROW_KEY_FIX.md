# Arrow Key Navigation Fix

## Issue
Arrow keys were not working in any of the TUI pages.

## Root Cause
The `handleKeyPress` function in `app.go` was intercepting all key presses but only handling specific keys (like 'd', 't', 'b', etc.) for view switching. Arrow keys and other navigation keys were not being passed to the current view for processing.

## Solution
Modified the `handleKeyPress` function to delegate unhandled key presses to the current view:

```go
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Help):
		m.state.CurrentView = ViewHelp
		return m, nil
	case key.Matches(msg, keys.Tab):
		// Cycle through views
		m.nextView()
		return m, nil
	}

	// View-specific keybindings
	switch msg.String() {
	case "d", "D":
		m.state.CurrentView = ViewDashboard
		return m, nil
	// ... other view switching keys ...
	}

	// If the key wasn't handled above, delegate to the current view
	// This allows arrow keys and other navigation to work
	if m.ready {
		return m.updateCurrentView(msg)
	}

	return m, nil
}
```

## Technical Details

1. **Before**: Arrow keys were captured by `handleKeyPress` but not processed or forwarded
2. **After**: Unhandled keys (including arrow keys) are now passed to `updateCurrentView` which delegates them to the appropriate view

## Impact
- Arrow keys now work in all views (Tool Browser, Bundle Explorer, Config, Health Monitor, etc.)
- Both arrow keys and vim-style navigation (hjkl) are supported
- No changes needed to individual view implementations

## Testing
After rebuilding with `make build`, arrow keys should work in:
- Tool Browser: Navigate up/down through tool list
- Bundle Explorer: Navigate through bundles
- Config View: Navigate through configuration items
- Health Monitor: Navigate through health checks
- Install Manager: Navigate through installation queue