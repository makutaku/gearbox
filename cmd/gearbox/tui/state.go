package tui

import (
	"fmt"
	"time"

	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

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
)

// AppState holds the global application state
type AppState struct {
	// Core data
	Tools          []orchestrator.ToolConfig
	Bundles        []orchestrator.BundleConfig
	InstalledTools map[string]*manifest.InstallationRecord
	Config         map[string]string
	
	// UI state
	CurrentView    ViewType
	ViewStack      []ViewType
	SearchQuery    string
	SelectedTools  map[string]bool
	SelectedBundle string
	Initialized    bool // Track whether lazy initialization has occurred
	
	// Background tasks
	InstallQueue   []InstallTask
	ActiveTasks    []InstallTask
	CompletedTasks []InstallTask
	
	// Health status
	HealthStatus   *HealthCheckResult
	LastHealthCheck time.Time
	
	// UI preferences
	Theme          string
	CompactMode    bool
}

// InstallTask represents a tool installation task
type InstallTask struct {
	ID         string
	Tool       orchestrator.ToolConfig
	BuildType  string
	Status     TaskStatus
	Progress   float64
	Stage      string
	Output     []string
	StartTime  time.Time
	EndTime    time.Time
	Error      error
}

// TaskStatus represents the status of an installation task
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
)

// HealthCheckResult represents system health check results
type HealthCheckResult struct {
	Overall     HealthStatus
	Checks      []HealthCheck
	LastUpdated time.Time
}

// HealthStatus represents the overall health status
type HealthStatus int

const (
	HealthStatusOK HealthStatus = iota
	HealthStatusWarning
	HealthStatusError
)

// HealthCheck represents a single health check
type HealthCheck struct {
	Name        string
	Status      HealthStatus
	Message     string
	Details     string
	Suggestions []string
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		Tools:          []orchestrator.ToolConfig{},
		Bundles:        []orchestrator.BundleConfig{},
		InstalledTools: make(map[string]*manifest.InstallationRecord),
		Config:         make(map[string]string),
		SelectedTools:  make(map[string]bool),
		CurrentView:    ViewDashboard,
		ViewStack:      []ViewType{},
		InstallQueue:   []InstallTask{},
		ActiveTasks:    []InstallTask{},
		CompletedTasks: []InstallTask{},
		Theme:          "default",
	}
}

// Navigation methods

// PushView pushes a new view onto the navigation stack
func (s *AppState) PushView(view ViewType) {
	s.ViewStack = append(s.ViewStack, s.CurrentView)
	s.CurrentView = view
}

// PopView pops the current view and returns to the previous one
func (s *AppState) PopView() bool {
	if len(s.ViewStack) > 0 {
		s.CurrentView = s.ViewStack[len(s.ViewStack)-1]
		s.ViewStack = s.ViewStack[:len(s.ViewStack)-1]
		return true
	}
	return false
}

// Tool selection methods

// ToggleToolSelection toggles the selection state of a tool
func (s *AppState) ToggleToolSelection(toolName string) {
	if s.SelectedTools[toolName] {
		delete(s.SelectedTools, toolName)
	} else {
		s.SelectedTools[toolName] = true
	}
}

// ClearToolSelection clears all selected tools
func (s *AppState) ClearToolSelection() {
	s.SelectedTools = make(map[string]bool)
}

// GetSelectedTools returns a list of selected tool names
func (s *AppState) GetSelectedTools() []string {
	var selected []string
	for tool := range s.SelectedTools {
		selected = append(selected, tool)
	}
	return selected
}

// Task management methods

// AddToInstallQueue adds a tool to the installation queue
func (s *AppState) AddToInstallQueue(tool orchestrator.ToolConfig, buildType string) {
	task := InstallTask{
		ID:        generateTaskID(),
		Tool:      tool,
		BuildType: buildType,
		Status:    TaskStatusPending,
		StartTime: time.Now(),
	}
	s.InstallQueue = append(s.InstallQueue, task)
}

// StartNextTask moves the next pending task to active
func (s *AppState) StartNextTask() *InstallTask {
	if len(s.InstallQueue) == 0 {
		return nil
	}
	
	// Find the first pending task
	for i, task := range s.InstallQueue {
		if task.Status == TaskStatusPending {
			task.Status = TaskStatusRunning
			s.InstallQueue[i] = task
			s.ActiveTasks = append(s.ActiveTasks, task)
			return &task
		}
	}
	
	return nil
}

// UpdateTaskProgress updates the progress of an active task
func (s *AppState) UpdateTaskProgress(taskID string, progress float64, stage string, output string) {
	for i, task := range s.ActiveTasks {
		if task.ID == taskID {
			task.Progress = progress
			task.Stage = stage
			if output != "" {
				task.Output = append(task.Output, output)
			}
			s.ActiveTasks[i] = task
			
			// Also update in install queue
			for j, qTask := range s.InstallQueue {
				if qTask.ID == taskID {
					s.InstallQueue[j] = task
					break
				}
			}
			break
		}
	}
}

// CompleteTask marks a task as completed
func (s *AppState) CompleteTask(taskID string, err error) {
	for i, task := range s.ActiveTasks {
		if task.ID == taskID {
			task.EndTime = time.Now()
			if err != nil {
				task.Status = TaskStatusFailed
				task.Error = err
			} else {
				task.Status = TaskStatusCompleted
			}
			
			// Move to completed
			s.CompletedTasks = append(s.CompletedTasks, task)
			
			// Remove from active
			s.ActiveTasks = append(s.ActiveTasks[:i], s.ActiveTasks[i+1:]...)
			
			// Update in install queue
			for j, qTask := range s.InstallQueue {
				if qTask.ID == taskID {
					s.InstallQueue[j] = task
					break
				}
			}
			break
		}
	}
}

// Health check methods

// UpdateHealthStatus updates the health check results
func (s *AppState) UpdateHealthStatus(result *HealthCheckResult) {
	s.HealthStatus = result
	s.LastHealthCheck = time.Now()
}

// GetHealthSummary returns a summary of the health status
func (s *AppState) GetHealthSummary() string {
	if s.HealthStatus == nil {
		return "Unknown"
	}
	
	switch s.HealthStatus.Overall {
	case HealthStatusOK:
		return "All systems OK"
	case HealthStatusWarning:
		warningCount := 0
		for _, check := range s.HealthStatus.Checks {
			if check.Status == HealthStatusWarning {
				warningCount++
			}
		}
		return fmt.Sprintf("%d warnings", warningCount)
	case HealthStatusError:
		errorCount := 0
		for _, check := range s.HealthStatus.Checks {
			if check.Status == HealthStatusError {
				errorCount++
			}
		}
		return fmt.Sprintf("%d errors", errorCount)
	default:
		return "Unknown"
	}
}

// Helper functions

func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// String returns the string representation of ViewType
func (v ViewType) String() string {
	switch v {
	case ViewDashboard:
		return "Dashboard"
	case ViewToolBrowser:
		return "Tool Browser"
	case ViewBundleExplorer:
		return "Bundle Explorer"
	case ViewMonitor:
		return "Monitor"
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

// String returns the string representation of TaskStatus
func (t TaskStatus) String() string {
	switch t {
	case TaskStatusPending:
		return "Pending"
	case TaskStatusRunning:
		return "Running"
	case TaskStatusCompleted:
		return "Completed"
	case TaskStatusFailed:
		return "Failed"
	case TaskStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// String returns the string representation of HealthStatus
func (h HealthStatus) String() string {
	switch h {
	case HealthStatusOK:
		return "OK"
	case HealthStatusWarning:
		return "Warning"
	case HealthStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}