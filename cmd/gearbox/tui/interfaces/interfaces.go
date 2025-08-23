package interfaces

import (
	tea "github.com/charmbracelet/bubbletea"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// ToolManager defines the interface for tool management operations
type ToolManager interface {
	InstallTools(tools []string, options orchestrator.InstallationOptions) tea.Cmd
	GetInstalledTools() (map[string]*manifest.InstallationRecord, error)
	GetAvailableTools() ([]orchestrator.ToolConfig, error)
}

// HealthChecker defines the interface for health checking operations
type HealthChecker interface {
	RunCheck(checkName string) tea.Cmd
	GetHealthStatus() map[string]HealthStatus
	GetLastCheckTime() map[string]interface{}
}

// HealthStatus represents the status of a health check
type HealthStatus int

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusPending
	HealthStatusPass
	HealthStatusWarning
	HealthStatusFail
)

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	GetConfig(key string) (string, error)
	SetConfig(key, value string) error
	GetAllConfig() (map[string]string, error)
	SaveConfig() error
}

// TaskProvider defines the interface for task management
type TaskProvider interface {
	StartTask(name string, cmd tea.Cmd) (taskID string)
	GetTaskStatus(taskID string) TaskStatus
	CancelTask(taskID string) error
	WatchUpdates() tea.Cmd
}

// TaskStatus represents the status of a task
type TaskStatus struct {
	ID       string
	Name     string
	Status   string
	Progress float64
	Output   []string
	Error    error
}

// DataProvider defines the interface for data loading operations
type DataProvider interface {
	LoadToolsAndBundles() ([]orchestrator.ToolConfig, []orchestrator.BundleConfig, error)
	LoadInstalledTools() (map[string]*manifest.InstallationRecord, error)
	LoadUnifiedStatus() (map[string]*manifest.InstallationRecord, error)
	ReloadManifest() (map[string]*manifest.InstallationRecord, error)
}

// ViewRenderer defines the interface for view rendering
type ViewRenderer interface {
	Render() string
	SetSize(width, height int)
	Update(msg tea.Msg) tea.Cmd
}

// NavigationProvider defines the interface for navigation handling
type NavigationProvider interface {
	HandleKeyPress(msg tea.KeyMsg, currentView ViewType) (newView ViewType, cmd tea.Cmd, handled bool)
	GetViewName(view ViewType) string
	GetNextView(current ViewType) ViewType
	GetPreviousView(current ViewType) ViewType
}

// MessageHandler defines the interface for message handling
type MessageHandler interface {
	HandleMessage(msg tea.Msg) tea.Cmd
}

// MessageRouter defines the interface for message routing
type MessageRouter interface {
	RegisterHandler(msgType interface{}, handler MessageHandler)
	Route(msg tea.Msg) tea.Cmd
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
	ViewNext
	ViewPrevious
)

// Logger defines the interface for logging operations
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// ErrorHandler defines the interface for error handling
type ErrorHandler interface {
	Handle(err error, context ...interface{}) tea.Cmd
	GetRecentErrors(count int) []AppError
	HasCriticalErrors() bool
	Clear()
}

// AppError represents a structured application error
type AppError struct {
	Type        ErrorType
	Message     string
	Details     []string
	Recoverable bool
}

// ErrorType represents different categories of errors
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeConfiguration
	ErrorTypeInstallation
	ErrorTypeNetwork
	ErrorTypeSystem
	ErrorTypeValidation
	ErrorTypePermission
)