package views

import tea "github.com/charmbracelet/bubbletea"

// View represents a common interface for all TUI views
type View interface {
	// SetSize updates the view size
	SetSize(width, height int)
	
	// Update handles view updates
	Update(msg tea.Msg) tea.Cmd
	
	// Render returns only the view's content (no chrome)
	Render() string
}

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
)

// TaskProvider provides access to task information
type TaskProvider interface {
	GetTask(taskID string) (*TaskInfo, bool)
	GetAllTasks() []*TaskInfo
	StartTask(taskID string) error
	CancelTask(taskID string) error
}

// TaskInfo represents task information for views
type TaskInfo struct {
	ID         string
	ToolName   string
	BuildType  string
	Status     TaskStatus
	Progress   float64
	Stage      string
	Output     []string
	StartTime  string
	EndTime    string
	Duration   string
	Error      error
}