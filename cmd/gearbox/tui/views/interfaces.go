package views

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