package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ViewLifecycleManager manages view lifecycle events and coordination
type ViewLifecycleManager struct {
	currentView    ViewType
	previousView   ViewType
	initializedViews map[ViewType]bool
	activeServices map[ViewType]ViewService
}

// NewViewLifecycleManager creates a new lifecycle manager
func NewViewLifecycleManager() *ViewLifecycleManager {
	return &ViewLifecycleManager{
		currentView:      ViewDashboard,
		previousView:     ViewDashboard,
		initializedViews: make(map[ViewType]bool),
		activeServices:   make(map[ViewType]ViewService),
	}
}

// RegisterView registers a view service with the lifecycle manager
func (vlm *ViewLifecycleManager) RegisterView(viewType ViewType, service ViewService) {
	vlm.activeServices[viewType] = service
}

// SwitchToView handles the complete lifecycle when switching between views
func (vlm *ViewLifecycleManager) SwitchToView(newView ViewType) tea.Cmd {
	if newView == vlm.currentView {
		// No change needed
		return nil
	}

	var commands []tea.Cmd

	// Get current and new view services
	currentService := vlm.activeServices[vlm.currentView]
	newService := vlm.activeServices[newView]

	// Deactivate current view
	if currentService != nil {
		// Check if current view can be deactivated
		canDeactivate, message := currentService.CanDeactivate()
		if !canDeactivate {
			// TODO: Show confirmation dialog with message
			// For now, proceed anyway
			debugLog("View deactivation warning: %s", message)
		}

		// Deactivate current view
		if cmd := currentService.OnDeactivate(); cmd != nil {
			commands = append(commands, cmd)
		}
	}

	// Initialize new view if needed
	if newService != nil && !vlm.initializedViews[newView] {
		if cmd := newService.OnInitialize(); cmd != nil {
			commands = append(commands, cmd)
		}
		vlm.initializedViews[newView] = true
	}

	// Activate new view
	if newService != nil {
		if cmd := newService.OnActivate(); cmd != nil {
			commands = append(commands, cmd)
		}
	}

	// Update state
	vlm.previousView = vlm.currentView
	vlm.currentView = newView

	debugLog("Lifecycle: Switched from %v to %v", vlm.previousView, vlm.currentView)

	// Return combined commands
	if len(commands) > 0 {
		return tea.Batch(commands...)
	}
	return nil
}

// RefreshCurrentView triggers a refresh of the current view
func (vlm *ViewLifecycleManager) RefreshCurrentView() tea.Cmd {
	if service := vlm.activeServices[vlm.currentView]; service != nil {
		return service.OnRefresh()
	}
	return nil
}

// DestroyAllViews cleans up all views (called on application exit)
func (vlm *ViewLifecycleManager) DestroyAllViews() tea.Cmd {
	var commands []tea.Cmd

	for viewType, service := range vlm.activeServices {
		if vlm.initializedViews[viewType] {
			if cmd := service.OnDestroy(); cmd != nil {
				commands = append(commands, cmd)
			}
		}
	}

	if len(commands) > 0 {
		return tea.Batch(commands...)
	}
	return nil
}

// GetCurrentView returns the current active view
func (vlm *ViewLifecycleManager) GetCurrentView() ViewType {
	return vlm.currentView
}

// GetPreviousView returns the previous view
func (vlm *ViewLifecycleManager) GetPreviousView() ViewType {
	return vlm.previousView
}

// IsInitialized returns whether a view has been initialized
func (vlm *ViewLifecycleManager) IsInitialized(viewType ViewType) bool {
	return vlm.initializedViews[viewType]
}

// SetFocusOnCurrentView sets focus to a specific element in the current view
func (vlm *ViewLifecycleManager) SetFocusOnCurrentView(elementID string) error {
	if service := vlm.activeServices[vlm.currentView]; service != nil {
		return service.SetFocus(elementID)
	}
	return nil
}

// GetFocusableElementsForCurrentView returns focusable elements for the current view
func (vlm *ViewLifecycleManager) GetFocusableElementsForCurrentView() []string {
	if service := vlm.activeServices[vlm.currentView]; service != nil {
		return service.GetFocusableElements()
	}
	return []string{}
}

// ViewTransitionEvent represents a view transition event
type ViewTransitionEvent struct {
	FromView ViewType
	ToView   ViewType
	Success  bool
	Message  string
}

// ViewLifecycleMsg represents lifecycle messages
type ViewLifecycleMsg struct {
	ViewType ViewType
	Event    string
	Data     interface{}
}

// RefreshViewMsg requests a view refresh
type RefreshViewMsg struct {
	ViewType ViewType
}

// FocusElementMsg requests focus on a specific element
type FocusElementMsg struct {
	ElementID string
}