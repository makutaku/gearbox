package app

import (
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
	"gearbox/cmd/gearbox/tui/tasks"
)

// Model represents the main TUI application model
type Model struct {
	orchestrator *orchestrator.Orchestrator
	manifest     *manifest.Manager
	state        *AppState
	taskManager  *tasks.TaskManager
	
	// Views
	dashboard      *views.Dashboard
	toolBrowser    *views.ToolBrowserNew
	bundleExplorer *views.BundleExplorerNew
	installManager *views.InstallManagerNew
	configView     *views.ConfigView
	healthView     *views.HealthView
	
	// Navigation and messaging
	navigator NavigationHandler
	router    MessageRouter
	
	// UI state
	width    int
	height   int
	ready    bool
	err      error
}

// NavigationHandler defines the interface for navigation handling
type NavigationHandler interface {
	HandleKeyPress(msg KeyMsg, currentView ViewType) (ViewType, Cmd, bool)
}

// MessageRouter defines the interface for message routing
type MessageRouter interface {
	Route(msg Msg) Cmd
}

// AppState holds the global application state
type AppState struct {
	// Core data
	Tools          []orchestrator.ToolConfig
	Bundles        []orchestrator.BundleConfig
	InstalledTools map[string]*manifest.InstallationRecord
	Config         map[string]string
	
	// UI state
	CurrentView  ViewType
	Initialized  bool
	Ready        bool
	Error        error
}

// ViewType represents different views in the TUI
type ViewType int

const (
	ViewDashboard ViewType = iota
	ViewToolBrowser
	ViewBundleExplorer
	ViewMonitor
	ViewConfig
	ViewHealth
	ViewHelp
	
	// Navigation helpers (not actual views)
	ViewNext
	ViewPrevious
)

// Bubble Tea framework types (to avoid imports in this package)
type Msg interface{}
type Cmd func() Msg
type KeyMsg interface {
	String() string
	Type() interface{}
}

// GetOrchestrator returns the orchestrator instance
func (m *Model) GetOrchestrator() *orchestrator.Orchestrator {
	return m.orchestrator
}

// GetState returns the application state
func (m *Model) GetState() *AppState {
	return m.state
}

// GetHealthView returns the health view instance
func (m *Model) GetHealthView() *views.HealthView {
	return m.healthView
}

// SetSize updates the model dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// IsReady returns whether the model is ready
func (m *Model) IsReady() bool {
	return m.ready
}

// GetError returns the current error state
func (m *Model) GetError() error {
	return m.err
}