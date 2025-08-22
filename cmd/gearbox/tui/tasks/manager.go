package tasks

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	
	"gearbox/pkg/orchestrator"
)

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
)

// InstallTask represents an installation task
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
	CancelChan chan bool
	
	mu sync.RWMutex
}

// TaskManager manages background installation tasks
type TaskManager struct {
	orchestrator *orchestrator.Orchestrator
	tasks        map[string]*InstallTask
	activeTasks  int
	maxParallel  int
	
	mu          sync.RWMutex
	updateChan  chan TaskUpdateMsg
}

// TaskUpdateMsg is sent when a task status changes
type TaskUpdateMsg struct {
	TaskID   string
	Status   TaskStatus
	Progress float64
	Stage    string
	Output   string
	Error    error
}

// NewTaskManager creates a new task manager
func NewTaskManager(orch *orchestrator.Orchestrator, maxParallel int) *TaskManager {
	if maxParallel <= 0 {
		maxParallel = 2
	}
	
	return &TaskManager{
		orchestrator: orch,
		tasks:        make(map[string]*InstallTask),
		maxParallel:  maxParallel,
		updateChan:   make(chan TaskUpdateMsg, 100),
	}
}

// AddTask adds a new installation task
func (tm *TaskManager) AddTask(tool orchestrator.ToolConfig, buildType string) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	task := &InstallTask{
		ID:         fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Tool:       tool,
		BuildType:  buildType,
		Status:     TaskStatusPending,
		StartTime:  time.Now(),
		Output:     []string{},
		CancelChan: make(chan bool, 1),
	}
	
	tm.tasks[task.ID] = task
	return task.ID
}

// StartTask starts a pending task if possible
func (tm *TaskManager) StartTask(taskID string) error {
	tm.mu.Lock()
	task, exists := tm.tasks[taskID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	if task.Status != TaskStatusPending {
		tm.mu.Unlock()
		return fmt.Errorf("task is not pending: %s", taskID)
	}
	
	if tm.activeTasks >= tm.maxParallel {
		tm.mu.Unlock()
		return fmt.Errorf("maximum parallel tasks reached")
	}
	
	task.Status = TaskStatusRunning
	tm.activeTasks++
	tm.mu.Unlock()
	
	// Start installation in background
	go tm.runInstallation(task)
	
	return nil
}

// StartNextPending starts the next pending task if any
func (tm *TaskManager) StartNextPending() {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	if tm.activeTasks >= tm.maxParallel {
		return
	}
	
	for _, task := range tm.tasks {
		if task.Status == TaskStatusPending {
			go func(taskID string) {
				if err := tm.StartTask(taskID); err != nil {
					log.Warn().Err(err).Str("task", taskID).Msg("Failed to start task")
				}
			}(task.ID)
			return
		}
	}
}

// runInstallation runs the actual installation
func (tm *TaskManager) runInstallation(task *InstallTask) {
	defer func() {
		tm.mu.Lock()
		tm.activeTasks--
		tm.mu.Unlock()
		
		// Try to start next pending task
		tm.StartNextPending()
	}()
	
	// Send initial update
	tm.sendUpdate(TaskUpdateMsg{
		TaskID: task.ID,
		Status: TaskStatusRunning,
		Stage:  "Preparing installation...",
	})
	
	// Create pipes for capturing output
	reader, writer := io.Pipe()
	defer reader.Close()
	
	// Start output reader
	outputDone := make(chan bool)
	go tm.readOutput(task, reader, outputDone)
	
	// For now, we'll simulate the installation
	// In a real implementation, we would integrate with the actual installation scripts
	
	// Simulate installation with progress updates
	// In real implementation, this would call the actual installation scripts
	err := tm.simulateInstallation(task, writer)
	
	writer.Close()
	<-outputDone
	
	// Update task status
	task.mu.Lock()
	task.EndTime = time.Now()
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err
		tm.sendUpdate(TaskUpdateMsg{
			TaskID: task.ID,
			Status: TaskStatusFailed,
			Error:  err,
		})
	} else {
		task.Status = TaskStatusCompleted
		task.Progress = 1.0
		tm.sendUpdate(TaskUpdateMsg{
			TaskID:   task.ID,
			Status:   TaskStatusCompleted,
			Progress: 1.0,
		})
	}
	task.mu.Unlock()
}

// simulateInstallation simulates an installation with progress updates
func (tm *TaskManager) simulateInstallation(task *InstallTask, output io.Writer) error {
	stages := []struct {
		name     string
		duration time.Duration
		progress float64
	}{
		{"Checking dependencies", 2 * time.Second, 0.1},
		{"Downloading source", 3 * time.Second, 0.3},
		{"Configuring build", 2 * time.Second, 0.4},
		{"Building from source", 5 * time.Second, 0.7},
		{"Running tests", 2 * time.Second, 0.8},
		{"Installing binaries", 1 * time.Second, 0.9},
		{"Updating PATH", 1 * time.Second, 1.0},
	}
	
	for _, stage := range stages {
		// Check for cancellation
		select {
		case <-task.CancelChan:
			return fmt.Errorf("installation cancelled")
		default:
		}
		
		// Update stage
		task.mu.Lock()
		task.Stage = stage.name
		task.Progress = stage.progress - 0.05
		task.mu.Unlock()
		
		tm.sendUpdate(TaskUpdateMsg{
			TaskID:   task.ID,
			Stage:    stage.name,
			Progress: stage.progress - 0.05,
		})
		
		// Simulate work with output
		fmt.Fprintf(output, "==> %s\n", stage.name)
		
		// Simulate progress
		steps := 10
		stepDuration := stage.duration / time.Duration(steps)
		progressStep := (stage.progress - task.Progress) / float64(steps)
		
		for i := 0; i < steps; i++ {
			select {
			case <-task.CancelChan:
				return fmt.Errorf("installation cancelled")
			case <-time.After(stepDuration):
				// Update progress
				task.mu.Lock()
				task.Progress += progressStep
				task.mu.Unlock()
				
				tm.sendUpdate(TaskUpdateMsg{
					TaskID:   task.ID,
					Progress: task.Progress,
					Output:   fmt.Sprintf("    %s: step %d/%d", stage.name, i+1, steps),
				})
				
				fmt.Fprintf(output, "    %s: step %d/%d\n", stage.name, i+1, steps)
			}
		}
		
		fmt.Fprintf(output, "    ✓ %s completed\n", stage.name)
	}
	
	fmt.Fprintf(output, "\n✅ Installation completed successfully!\n")
	fmt.Fprintf(output, "Tool '%s' is now available in your PATH.\n", task.Tool.Name)
	
	return nil
}

// readOutput reads output from the installation process
func (tm *TaskManager) readOutput(task *InstallTask, reader io.Reader, done chan bool) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		
		task.mu.Lock()
		task.Output = append(task.Output, line)
		// Keep only last 100 lines
		if len(task.Output) > 100 {
			task.Output = task.Output[len(task.Output)-100:]
		}
		task.mu.Unlock()
		
		// Parse progress from output if possible
		progress := tm.parseProgress(line)
		if progress >= 0 {
			task.mu.Lock()
			task.Progress = progress
			task.mu.Unlock()
		}
		
		// Send update
		tm.sendUpdate(TaskUpdateMsg{
			TaskID: task.ID,
			Output: line,
		})
	}
	
	close(done)
}

// parseProgress attempts to extract progress from output line
func (tm *TaskManager) parseProgress(line string) float64 {
	// Look for patterns like "50%" or "[50/100]"
	if strings.Contains(line, "%") {
		// Simple percentage parsing
		parts := strings.Fields(line)
		for _, part := range parts {
			if strings.HasSuffix(part, "%") {
				var pct float64
				if _, err := fmt.Sscanf(part, "%f%%", &pct); err == nil {
					return pct / 100.0
				}
			}
		}
	}
	
	return -1
}

// CancelTask cancels a running task
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.mu.RLock()
	task, exists := tm.tasks[taskID]
	tm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	if task.Status != TaskStatusRunning {
		return fmt.Errorf("task is not running")
	}
	
	select {
	case task.CancelChan <- true:
		return nil
	default:
		return fmt.Errorf("failed to send cancel signal")
	}
}

// GetTask returns a copy of the task
func (tm *TaskManager) GetTask(taskID string) (*InstallTask, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, false
	}
	
	// Return a copy to avoid race conditions
	taskCopy := *task
	return &taskCopy, true
}

// GetAllTasks returns copies of all tasks
func (tm *TaskManager) GetAllTasks() []*InstallTask {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	tasks := make([]*InstallTask, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}
	
	return tasks
}

// sendUpdate sends a task update message
func (tm *TaskManager) sendUpdate(update TaskUpdateMsg) {
	select {
	case tm.updateChan <- update:
	default:
		// Channel full, drop update
		log.Warn().Str("task", update.TaskID).Msg("Update channel full, dropping update")
	}
}

// WatchUpdates returns a command that watches for task updates
func (tm *TaskManager) WatchUpdates() tea.Cmd {
	return func() tea.Msg {
		return <-tm.updateChan
	}
}

// String returns the string representation of TaskStatus
func (s TaskStatus) String() string {
	switch s {
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