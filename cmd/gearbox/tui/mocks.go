package tui

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
)

// MockOrchestrator provides a mock implementation for testing
type MockOrchestrator struct {
	tools   []orchestrator.ToolConfig
	bundles []orchestrator.BundleConfig
}

func (m *MockOrchestrator) InstallTool(name string, buildType string) error {
	// Simulate installation with random delay
	delay := time.Duration(rand.Intn(3000)) * time.Millisecond
	time.Sleep(delay)
	return nil
}

func (m *MockOrchestrator) InstallTools(toolNames []string) error {
	// Simulate installation with random delay for each tool
	for range toolNames {
		delay := time.Duration(rand.Intn(2000)) * time.Millisecond
		time.Sleep(delay)
	}
	return nil
}

func (m *MockOrchestrator) GetTools() []orchestrator.ToolConfig {
	return m.tools
}

func (m *MockOrchestrator) GetBundles() []orchestrator.BundleConfig {
	return m.bundles
}

// MockManifest provides a mock manifest implementation
type MockManifest struct {
	mu        sync.Mutex
	installed map[string]*manifest.InstallationRecord
}

func (m *MockManifest) Load() (*manifest.InstallationManifest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	return &manifest.InstallationManifest{
		Installations: m.installed,
	}, nil
}

func (m *MockManifest) Save(manifest *manifest.InstallationManifest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.installed = manifest.Installations
	return nil
}

func (m *MockManifest) AddInstallation(toolName string, record *manifest.InstallationRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.installed[toolName] = record
	return nil
}

func (m *MockManifest) GetAllTasks() []*views.TaskInfo {
	// Mock implementation - return empty slice
	return []*views.TaskInfo{}
}

func (m *MockManifest) StartTask(taskID string) error {
	// Mock implementation - delegate to task manager if available
	return nil
}

func (m *MockManifest) CancelTask(taskID string) error {
	// Mock implementation - delegate to task manager if available  
	return nil
}

// MockTaskManager provides a mock task manager for testing
type MockTaskManager struct {
	mu        sync.Mutex
	tasks     map[string]*MockTask
	nextID    int
	updateCh  chan tasks.TaskUpdateMsg
}

type MockTask struct {
	ID         string
	ToolName   string
	BuildType  string
	Status     views.TaskStatus
	Progress   float64
	Stage      string
	StartTime  string
	EndTime    string
	Duration   string
	Error      error
	Output     []string
}

func NewMockTaskManager() *MockTaskManager {
	return &MockTaskManager{
		tasks:    make(map[string]*MockTask),
		nextID:   1,
		updateCh: make(chan tasks.TaskUpdateMsg, 100),
	}
}

func (m *MockTaskManager) AddTask(tool orchestrator.ToolConfig, buildType string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	taskID := fmt.Sprintf("task-%d", m.nextID)
	m.nextID++
	
	task := &MockTask{
		ID:        taskID,
		ToolName:  tool.Name,
		BuildType: buildType,
		Status:    views.TaskStatusPending,
		Progress:  0.0,
		Stage:     "Queued",
		StartTime: time.Now().Format("15:04:05"),
	}
	
	m.tasks[taskID] = task
	return taskID
}

func (m *MockTaskManager) StartTask(taskID string) {
	go m.simulateTask(taskID)
}

func (m *MockTaskManager) CancelTask(taskID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if task, exists := m.tasks[taskID]; exists {
		task.Status = views.TaskStatusCancelled
		task.Stage = "Cancelled"
		task.EndTime = time.Now().Format("15:04:05")
	}
}

func (m *MockTaskManager) GetTask(taskID string) (*views.TaskInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	mockTask, exists := m.tasks[taskID]
	if !exists {
		return nil, false
	}
	
	taskInfo := &views.TaskInfo{
		ID:        mockTask.ID,
		ToolName:  mockTask.ToolName,
		BuildType: mockTask.BuildType,
		Status:    mockTask.Status,
		Progress:  mockTask.Progress,
		Stage:     mockTask.Stage,
		StartTime: mockTask.StartTime,
		EndTime:   mockTask.EndTime,
		Duration:  mockTask.Duration,
		Error:     mockTask.Error,
		Output:    mockTask.Output,
	}
	
	return taskInfo, true
}

func (m *MockTaskManager) GetAllTasks() []*views.TaskInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var allTasks []*views.TaskInfo
	for _, mockTask := range m.tasks {
		taskInfo := &views.TaskInfo{
			ID:        mockTask.ID,
			ToolName:  mockTask.ToolName,
			BuildType: mockTask.BuildType,
			Status:    mockTask.Status,
			Progress:  mockTask.Progress,
			Stage:     mockTask.Stage,
			StartTime: mockTask.StartTime,
			EndTime:   mockTask.EndTime,
			Duration:  mockTask.Duration,
			Error:     mockTask.Error,
			Output:    mockTask.Output,
		}
		allTasks = append(allTasks, taskInfo)
	}
	
	return allTasks
}

func (m *MockTaskManager) WatchUpdates() tea.Cmd {
	return func() tea.Msg {
		select {
		case update := <-m.updateCh:
			return update
		case <-time.After(100 * time.Millisecond):
			// Return empty message if no updates
			return nil
		}
	}
}

// simulateTask simulates a realistic installation process
func (m *MockTaskManager) simulateTask(taskID string) {
	m.mu.Lock()
	task, exists := m.tasks[taskID]
	if !exists {
		m.mu.Unlock()
		return
	}
	task.Status = views.TaskStatusRunning
	task.Stage = "Starting"
	m.mu.Unlock()
	
	// Send initial update
	m.updateCh <- tasks.TaskUpdateMsg{
		TaskID:   taskID,
		Progress: 0.0,
	}
	
	// Simulate installation stages
	stages := []struct {
		name     string
		duration time.Duration
		progress float64
	}{
		{"Checking dependencies", 500 * time.Millisecond, 0.1},
		{"Downloading source", 1000 * time.Millisecond, 0.3},
		{"Configuring build", 300 * time.Millisecond, 0.4},
		{"Compiling", 2000 * time.Millisecond, 0.8},
		{"Installing", 500 * time.Millisecond, 0.95},
		{"Verifying installation", 200 * time.Millisecond, 1.0},
	}
	
	for _, stage := range stages {
		// Check if cancelled
		m.mu.Lock()
		if task.Status == views.TaskStatusCancelled {
			m.mu.Unlock()
			return
		}
		
		task.Stage = stage.name
		task.Progress = stage.progress
		
		// Add some realistic output
		task.Output = append(task.Output, fmt.Sprintf("[%s] %s...", 
			time.Now().Format("15:04:05"), stage.name))
		if len(task.Output) > 10 {
			task.Output = task.Output[len(task.Output)-10:] // Keep last 10 lines
		}
		m.mu.Unlock()
		
		// Send progress update
		m.updateCh <- tasks.TaskUpdateMsg{
			TaskID:   taskID,
			Progress: stage.progress,
		}
		
		time.Sleep(stage.duration)
	}
	
	// Complete the task
	m.mu.Lock()
	task.Status = views.TaskStatusCompleted
	task.Stage = "Completed"
	task.Progress = 1.0
	task.EndTime = time.Now().Format("15:04:05")
	
	// Calculate duration
	startTime, _ := time.Parse("15:04:05", task.StartTime)
	endTime, _ := time.Parse("15:04:05", task.EndTime)
	duration := endTime.Sub(startTime)
	task.Duration = duration.Round(time.Second).String()
	
	task.Output = append(task.Output, fmt.Sprintf("[%s] ✅ Installation completed successfully!", 
		time.Now().Format("15:04:05")))
	m.mu.Unlock()
	
	// Send final update
	m.updateCh <- tasks.TaskUpdateMsg{
		TaskID:   taskID,
		Progress: 1.0,
	}
}

// MockTaskProvider adapts MockTaskManager to TaskProvider interface
type MockTaskProvider struct {
	taskManager *MockTaskManager
}

func NewMockTaskProvider(taskManager *MockTaskManager) *MockTaskProvider {
	return &MockTaskProvider{
		taskManager: taskManager,
	}
}

func (m *MockTaskProvider) GetTask(taskID string) (*views.TaskInfo, bool) {
	return m.taskManager.GetTask(taskID)
}

func (m *MockTaskProvider) GetAllTasks() []*views.TaskInfo {
	return m.taskManager.GetAllTasks()
}

func (m *MockTaskProvider) StartTask(taskID string) error {
	m.taskManager.StartTask(taskID)
	return nil
}

func (m *MockTaskProvider) CancelTask(taskID string) error {
	m.taskManager.CancelTask(taskID)
	return nil
}