package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// OrchestratorService defines the interface for tool installation orchestration
type OrchestratorService interface {
	// InstallTools installs specified tools
	InstallTools(toolNames []string) error
	
	// ListTools lists available tools with optional category filter
	ListTools(category string, verbose bool) error
	
	// ListBundles lists available bundles
	ListBundles(verbose bool) error
	
	// ShowStatus shows status of tools
	ShowStatus(toolNames []string, manifestOnly bool, unified bool) error
	
	// GetConfig returns the orchestrator configuration
	GetConfig() *orchestrator.Config
	
	// RunDoctor runs diagnostic checks
	RunDoctor(toolNames []string) error
}

// ManifestService defines the interface for installation tracking and state management
type ManifestService interface {
	// Load returns the current manifest data
	Load() (*manifest.InstallationManifest, error)
	
	// Save persists the manifest data
	Save(m *manifest.InstallationManifest) error
}

// TaskService defines the interface for task management
type TaskService interface {
	// AddTask adds a new task to the queue
	AddTask(tool orchestrator.ToolConfig, buildType string) string
	
	// StartTask starts execution of a queued task
	StartTask(taskID string) error
	
	// CancelTask cancels a running task
	CancelTask(taskID string) error
	
	// GetAllTasks returns all tasks
	GetAllTasks() []*tasks.InstallTask
	
	// WatchUpdates returns a command to watch for task updates
	WatchUpdates() tea.Cmd
}

// NavigationService defines the interface for navigation handling
type NavigationService interface {
	// HandleKeyPress processes keyboard navigation and returns new view and commands
	HandleKeyPress(msg tea.KeyMsg, currentView ViewType) (ViewType, tea.Cmd, bool)
	
	// GetNextView returns the next view in navigation order
	GetNextView(current ViewType) ViewType
	
	// GetPreviousView returns the previous view in navigation order  
	GetPreviousView(current ViewType) ViewType
	
	// IsNavigationKey checks if a key is a navigation key
	IsNavigationKey(key string) bool
	
	// GetKeyBindings returns all key bindings
	GetKeyBindings() map[string]ViewType
}

// MessageRoutingService defines the interface for message routing
type MessageRoutingService interface {
	// Route routes a message to appropriate handlers
	Route(msg tea.Msg) tea.Cmd
}

// Note: MessageHandler is already defined in router.go

// ViewService defines the interface for view management with lifecycle methods
type ViewService interface {
	// SetSize updates the view size
	SetSize(width, height int)
	
	// Render renders the view content
	Render() string
	
	// Update handles view updates
	Update(msg tea.Msg) tea.Cmd
	
	// GetType returns the view type
	GetType() ViewType
	
	// IsReady returns whether the view is ready for interaction
	IsReady() bool
	
	// Lifecycle methods
	
	// OnActivate is called when the view becomes active (user navigates to it)
	OnActivate() tea.Cmd
	
	// OnDeactivate is called when the view becomes inactive (user navigates away)
	OnDeactivate() tea.Cmd
	
	// OnInitialize is called once when the view is first created
	OnInitialize() tea.Cmd
	
	// OnDestroy is called when the view is being destroyed/cleaned up
	OnDestroy() tea.Cmd
	
	// OnRefresh is called when the view needs to refresh its data
	OnRefresh() tea.Cmd
	
	// CanDeactivate returns whether the view can be deactivated (for unsaved changes, etc.)
	CanDeactivate() (bool, string)
	
	// GetFocusableElements returns elements that can receive focus
	GetFocusableElements() []string
	
	// SetFocus sets focus to a specific element
	SetFocus(elementID string) error
}

// DashboardService defines the interface for dashboard-specific functionality
type DashboardService interface {
	ViewService
	
	// SetData updates dashboard data
	SetData(tools []orchestrator.ToolConfig, bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord)
	
	// GetRecommendations returns smart recommendations for the user
	GetRecommendations() []string
	
	// GetSystemStatus returns current system status
	GetSystemStatus() map[string]interface{}
}

// ToolBrowserService defines the interface for tool browser functionality
type ToolBrowserService interface {
	ViewService
	
	// SetData updates tool browser data
	SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord)
	
	// GetSelectedTools returns currently selected tools
	GetSelectedTools() []string
	
	// ClearSelection clears all selections
	ClearSelection()
	
	// ToggleSelection toggles selection for a specific tool
	ToggleSelection(toolName string)
	
	// FilterByCategory filters tools by category
	FilterByCategory(category string)
	
	// Search searches tools by name or description
	Search(query string)
	
	// LoadFullContent loads full tool content asynchronously
	LoadFullContent()
}

// BundleExplorerService defines the interface for bundle explorer functionality
type BundleExplorerService interface {
	ViewService
	
	// SetData updates bundle explorer data
	SetData(bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord)
	
	// GetSelectedBundle returns the currently selected bundle
	GetSelectedBundle() *orchestrator.BundleConfig
	
	// GetUninstalledTools returns tools in a bundle that are not installed
	GetUninstalledTools(bundle *orchestrator.BundleConfig) []string
	
	// ExpandBundle expands/collapses a bundle
	ExpandBundle(bundleName string)
	
	// FilterByCategory filters bundles by category
	FilterByCategory(category string)
}

// InstallManagerService defines the interface for installation management
type InstallManagerService interface {
	ViewService
	
	// AddTaskID adds a task ID to monitor
	AddTaskID(taskID string)
	
	// RemoveTaskID removes a task ID from monitoring
	RemoveTaskID(taskID string)
	
	// HandleTaskUpdate handles task progress updates
	HandleTaskUpdate(taskID string, progress float64)
	
	// GetActiveTaskCount returns the number of active tasks
	GetActiveTaskCount() int
	
	// ToggleOutputDisplay toggles output visibility
	ToggleOutputDisplay()
	
	// ClearCompletedTasks removes completed tasks from display
	ClearCompletedTasks()
}

// ConfigService defines the interface for configuration management
type ConfigService interface {
	ViewService
	
	// GetSettings returns all configuration settings
	GetSettings() map[string]interface{}
	
	// SetSetting updates a configuration setting
	SetSetting(key string, value interface{}) error
	
	// ResetSetting resets a setting to default value
	ResetSetting(key string) error
	
	// SaveSettings persists all settings
	SaveSettings() error
	
	// LoadSettings loads settings from storage
	LoadSettings() error
	
	// GetSettingType returns the type of a setting
	GetSettingType(key string) string
}

// HealthService defines the interface for health monitoring
type HealthService interface {
	ViewService
	
	// SetData updates health view data
	SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord)
	
	// RunNextHealthCheck runs the next health check in sequence
	RunNextHealthCheck(checkIndex int) tea.Cmd
	
	// GetHealthStatus returns overall health status
	GetHealthStatus() (passing int, warning int, failing int)
	
	// ToggleAutoRefresh toggles automatic refresh
	ToggleAutoRefresh()
	
	// ToggleDetails toggles detail visibility
	ToggleDetails()
	
	// GetSystemChecks returns system health checks
	GetSystemChecks() []views.HealthCheck
	
	// GetToolChecks returns tool health checks  
	GetToolChecks() []views.HealthCheck
}

// Note: HealthCheck and HealthStatus are already defined in state.go