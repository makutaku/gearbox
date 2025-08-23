package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// QuitRequestedMsg indicates that quit was requested
type QuitRequestedMsg struct{}

// NavigationHandler manages view navigation and keyboard shortcuts
type NavigationHandler struct {
	keyBindings map[string]ViewType
}

// NewNavigationHandler creates a new navigation handler with default keybindings
func NewNavigationHandler() *NavigationHandler {
	keyBindings := map[string]ViewType{
		"D":   ViewDashboard,
		"T":   ViewToolBrowser,
		"B":   ViewBundleExplorer,
		"M":   ViewMonitor,
		"C":   ViewConfig,
		"H":   ViewHealth,
		"?":   ViewHelp,
		"tab": ViewNext,
		"shift+tab": ViewPrevious,
	}
	
	return &NavigationHandler{
		keyBindings: keyBindings,
	}
}

// HandleKeyPress processes keyboard navigation and returns new view and commands
func (n *NavigationHandler) HandleKeyPress(msg tea.KeyMsg, currentView ViewType) (ViewType, tea.Cmd, bool) {
	// Check for quit keys first
	if key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))) {
		return currentView, func() tea.Msg { return QuitRequestedMsg{} }, true
	}
	
	// Handle tab and arrow navigation
	switch msg.Type {
	case tea.KeyTab:
		nextView := n.getNextView(currentView)
		debugLog("Navigation: Tab pressed, switching from %v to %v", currentView, nextView)
		return nextView, nil, true
		
	case tea.KeyShiftTab:
		prevView := n.getPreviousView(currentView)
		debugLog("Navigation: Shift+Tab pressed, switching from %v to %v", currentView, prevView)
		return prevView, nil, true
		
	case tea.KeyLeft:
		// Don't navigate away from Help view with arrows
		if currentView == ViewHelp {
			return currentView, nil, false
		}
		prevView := n.getPreviousView(currentView)
		debugLog("Navigation: Left arrow pressed, switching from %v to %v", currentView, prevView)
		return prevView, nil, true
		
	case tea.KeyRight:
		// Don't navigate away from Help view with arrows
		if currentView == ViewHelp {
			return currentView, nil, false
		}
		nextView := n.getNextView(currentView)
		debugLog("Navigation: Right arrow pressed, switching from %v to %v", currentView, nextView)
		return nextView, nil, true
	}
	
	// Handle single key shortcuts
	keyStr := msg.String()
	if newView, exists := n.keyBindings[keyStr]; exists {
		if newView != ViewNext && newView != ViewPrevious {
			debugLog("Navigation: Key '%s' pressed, switching from %v to %v", keyStr, currentView, newView)
			return newView, nil, true
		}
	}
	
	// Key not handled by navigation
	return currentView, nil, false
}

// getNextView returns the next view in the navigation order
func (n *NavigationHandler) getNextView(current ViewType) ViewType {
	switch current {
	case ViewDashboard:
		return ViewToolBrowser
	case ViewToolBrowser:
		return ViewBundleExplorer
	case ViewBundleExplorer:
		return ViewMonitor
	case ViewMonitor:
		return ViewConfig
	case ViewConfig:
		return ViewHealth
	case ViewHealth:
		return ViewDashboard
	default:
		return ViewDashboard
	}
}

// getPreviousView returns the previous view in the navigation order
func (n *NavigationHandler) getPreviousView(current ViewType) ViewType {
	switch current {
	case ViewDashboard:
		return ViewHealth
	case ViewToolBrowser:
		return ViewDashboard
	case ViewBundleExplorer:
		return ViewToolBrowser
	case ViewMonitor:
		return ViewBundleExplorer
	case ViewConfig:
		return ViewMonitor
	case ViewHealth:
		return ViewConfig
	default:
		return ViewDashboard
	}
}

// GetViewName returns a human-readable name for the view
func GetViewName(view ViewType) string {
	switch view {
	case ViewDashboard:
		return "Dashboard"
	case ViewToolBrowser:
		return "Tool Browser"
	case ViewBundleExplorer:
		return "Bundle Explorer"
	case ViewMonitor:
		return "Install Manager"
	case ViewConfig:
		return "Configuration"
	case ViewHealth:
		return "Health Monitor"
	case ViewHelp:
		return "Help"
	default:
		return "Unknown"
	}
}