package tui

import (
	"fmt"
	"time"
	
	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
)

// TaskManagerProvider wraps the task manager to implement TaskProvider
type TaskManagerProvider struct {
	taskManager *tasks.TaskManager
}

// NewTaskManagerProvider creates a new task manager provider
func NewTaskManagerProvider(tm *tasks.TaskManager) *TaskManagerProvider {
	return &TaskManagerProvider{
		taskManager: tm,
	}
}

// GetTask returns task information
func (p *TaskManagerProvider) GetTask(taskID string) (*views.TaskInfo, bool) {
	task, exists := p.taskManager.GetTask(taskID)
	if !exists {
		return nil, false
	}
	
	info := &views.TaskInfo{
		ID:        task.ID,
		ToolName:  task.Tool.Name,
		BuildType: task.BuildType,
		Status:    views.TaskStatus(task.Status),
		Progress:  task.Progress,
		Stage:     task.Stage,
		Output:    task.Output,
		StartTime: task.StartTime.Format("15:04:05"),
		Error:     task.Error,
	}
	
	if !task.EndTime.IsZero() {
		info.EndTime = task.EndTime.Format("15:04:05")
		info.Duration = formatDuration(task.EndTime.Sub(task.StartTime))
	} else if task.Status == tasks.TaskStatusRunning {
		info.Duration = formatDuration(time.Since(task.StartTime))
	}
	
	return info, true
}

// GetAllTasks returns all tasks
func (p *TaskManagerProvider) GetAllTasks() []*views.TaskInfo {
	allTasks := p.taskManager.GetAllTasks()
	infos := make([]*views.TaskInfo, len(allTasks))
	
	for i, task := range allTasks {
		infos[i] = &views.TaskInfo{
			ID:        task.ID,
			ToolName:  task.Tool.Name,
			BuildType: task.BuildType,
			Status:    views.TaskStatus(task.Status),
			Progress:  task.Progress,
			Stage:     task.Stage,
			Output:    task.Output,
			StartTime: task.StartTime.Format("15:04:05"),
			Error:     task.Error,
		}
		
		if !task.EndTime.IsZero() {
			infos[i].EndTime = task.EndTime.Format("15:04:05")
			infos[i].Duration = formatDuration(task.EndTime.Sub(task.StartTime))
		} else if task.Status == tasks.TaskStatusRunning {
			infos[i].Duration = formatDuration(time.Since(task.StartTime))
		}
	}
	
	return infos
}

// StartTask starts a task
func (p *TaskManagerProvider) StartTask(taskID string) error {
	return p.taskManager.StartTask(taskID)
}

// CancelTask cancels a task
func (p *TaskManagerProvider) CancelTask(taskID string) error {
	return p.taskManager.CancelTask(taskID)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "< 1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}