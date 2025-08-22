package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"gearbox/cmd/gearbox/tui/styles"
)

// We'll use TaskStatus from interfaces.go instead

// InstallManager represents the installation manager view
type InstallManager struct {
	width  int
	height int
	
	// Task provider
	taskProvider   TaskProvider
	
	// Task IDs from task manager
	taskIDs        []string
	
	// UI state
	cursor         int
	scrollOffset   int
	selectedTaskID string
	showOutput     bool
	
	// Progress bars
	progressBars   map[string]progress.Model
}

// NewInstallManager creates a new install manager view
func NewInstallManager() *InstallManager {
	return &InstallManager{
		taskIDs:      []string{},
		progressBars: make(map[string]progress.Model),
		showOutput:   true,
	}
}

// SetSize updates the size of the install manager
func (im *InstallManager) SetSize(width, height int) {
	im.width = width
	im.height = height
	
	// Update progress bar widths
	for id, prog := range im.progressBars {
		prog.Width = im.width / 2
		im.progressBars[id] = prog
	}
}

// SetTaskProvider sets the task provider
func (im *InstallManager) SetTaskProvider(provider TaskProvider) {
	im.taskProvider = provider
}

// AddTaskID adds a new task ID to track
func (im *InstallManager) AddTaskID(taskID string) {
	im.taskIDs = append(im.taskIDs, taskID)
	
	// Create progress bar for this task
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = im.width / 2
	im.progressBars[taskID] = prog
}

// HandleTaskUpdate handles task update messages
func (im *InstallManager) HandleTaskUpdate(taskID string, progress float64) {
	// Progress will be updated when we fetch the task in the Update method
	// This method is just a notification that an update occurred
}

// Update handles install manager updates
func (im *InstallManager) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			im.moveUp()
		case "down", "j":
			im.moveDown()
		case "enter":
			im.selectCurrentTask()
		case "o":
			im.showOutput = !im.showOutput
		case "s":
			// Start installations
			return im.startPendingTasks()
		case "c":
			// Cancel current task
			if im.selectedTaskID != "" && im.taskProvider != nil {
				im.taskProvider.CancelTask(im.selectedTaskID)
			}
		case "C":
			// Cancel all tasks
			if im.taskProvider != nil {
				for _, taskID := range im.taskIDs {
					im.taskProvider.CancelTask(taskID)
				}
			}
		}
	}
	
	// Update progress bars
	var cmds []tea.Cmd
	for id, prog := range im.progressBars {
		if task := im.findTask(id); task != nil && task.Status == TaskStatusRunning {
			// Progress is already set in the task, just update the model
			newModel, cmd := prog.Update(msg)
			if newProg, ok := newModel.(progress.Model); ok {
				im.progressBars[id] = newProg
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	
	return tea.Batch(cmds...)
}

// Render returns the rendered install manager view
func (im *InstallManager) Render() string {
	if im.width == 0 || im.height == 0 {
		return "Loading..."
	}
	
	// Calculate layout
	// Reserve space for status bar (2) and help bar (2)
	availableHeight := im.height - 4
	queueHeight := availableHeight / 3
	detailHeight := availableHeight - queueHeight
	
	// Render components
	queueView := im.renderQueue(queueHeight)
	detailView := im.renderDetails(detailHeight)
	statusBar := im.renderStatusBar()
	helpBar := im.renderHelpBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		queueView,
		detailView,
		statusBar,
		helpBar,
	)
}

func (im *InstallManager) renderQueue(height int) string {
	title := styles.TitleStyle().Render("Installation Queue")
	
	// Get all tasks from task provider
	var allTasks []*TaskInfo
	if im.taskProvider != nil {
		// Get tasks for our tracked IDs
		for _, id := range im.taskIDs {
			if task, ok := im.taskProvider.GetTask(id); ok {
				allTasks = append(allTasks, task)
			}
		}
	}
	
	if len(allTasks) == 0 {
		content := lipgloss.NewStyle().
			Height(height-3).
			Width(im.width-4).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No tasks in queue\n\nAdd tools from the Tool Browser [T]")
		
		return styles.BoxStyle().
			Width(im.width).
			Height(height).
			Render(title + "\n" + content)
	}
	
	// Calculate visible items
	visibleItems := height - 4
	if im.cursor >= im.scrollOffset+visibleItems {
		im.scrollOffset = im.cursor - visibleItems + 1
	} else if im.cursor < im.scrollOffset {
		im.scrollOffset = im.cursor
	}
	
	var items []string
	for i := im.scrollOffset; i < min(im.scrollOffset+visibleItems, len(allTasks)); i++ {
		task := allTasks[i]
		items = append(items, im.renderTaskItem(task, i == im.cursor))
	}
	
	content := strings.Join(items, "\n")
	
	return styles.BoxStyle().
		Width(im.width).
		Height(height).
		Render(title + "\n" + content)
}

func (im *InstallManager) renderTaskItem(task *TaskInfo, selected bool) string {
	// Status icon
	var statusIcon string
	statusStyle := lipgloss.NewStyle()
	
	switch task.Status {
	case TaskStatusPending:
		statusIcon = "⏳"
		statusStyle = statusStyle.Foreground(styles.CurrentTheme.Warning)
	case TaskStatusRunning:
		statusIcon = "🔄"
		statusStyle = statusStyle.Foreground(styles.CurrentTheme.Info)
	case TaskStatusCompleted:
		statusIcon = "✅"
		statusStyle = statusStyle.Foreground(styles.CurrentTheme.Success)
	case TaskStatusFailed:
		statusIcon = "❌"
		statusStyle = statusStyle.Foreground(styles.CurrentTheme.Error)
	case TaskStatusCancelled:
		statusIcon = "⏹️"
		statusStyle = statusStyle.Foreground(styles.CurrentTheme.Muted)
	}
	
	// Build the item
	name := fmt.Sprintf("%-20s", task.ToolName)
	buildType := fmt.Sprintf("[%s]", task.BuildType)
	
	// Progress or stage
	var progress string
	if task.Status == TaskStatusRunning {
		if task.Progress > 0 {
			progress = fmt.Sprintf("%.0f%% %s", task.Progress*100, task.Stage)
		} else {
			progress = task.Stage
		}
	} else if task.Status == TaskStatusCompleted {
		progress = fmt.Sprintf("Completed in %s", task.Duration)
	} else if task.Status == TaskStatusFailed {
		progress = "Failed"
	}
	
	item := fmt.Sprintf("%s %s %s %s",
		statusStyle.Render(statusIcon),
		name,
		buildType,
		progress,
	)
	
	// Apply selection style
	if selected {
		return styles.SelectedStyle().Render(item)
	}
	
	return item
}

func (im *InstallManager) renderDetails(height int) string {
	selectedTask := im.getSelectedTask()
	if selectedTask == nil {
		return styles.BoxStyle().
			Width(im.width).
			Height(height).
			Render("Select a task to view details")
	}
	
	title := styles.TitleStyle().Render(fmt.Sprintf("Task Details: %s", selectedTask.ToolName))
	
	// Task information
	info := fmt.Sprintf(
		"Build Type: %s\nStatus: %s\nStarted: %s",
		selectedTask.BuildType,
		im.getStatusText(selectedTask.Status),
		selectedTask.StartTime,
	)
	
	if selectedTask.Status == TaskStatusCompleted || selectedTask.Status == TaskStatusFailed {
		info += fmt.Sprintf("\nEnded: %s", selectedTask.EndTime)
		info += fmt.Sprintf("\nDuration: %s", selectedTask.Duration)
	}
	
	// Progress bar for running tasks
	var progressBar string
	if selectedTask.Status == TaskStatusRunning && selectedTask.Progress > 0 {
		if prog, ok := im.progressBars[selectedTask.ID]; ok {
			// ViewAs takes the progress value to render
			progressBar = "\n\n" + prog.ViewAs(selectedTask.Progress)
		}
	}
	
	// Output section
	var output string
	if im.showOutput && len(selectedTask.Output) > 0 {
		outputTitle := styles.SubtitleStyle().Render("Output:")
		
		// Get last N lines of output
		outputLines := selectedTask.Output
		maxLines := height - 10
		if len(outputLines) > maxLines {
			outputLines = outputLines[len(outputLines)-maxLines:]
		}
		
		outputContent := lipgloss.NewStyle().
			Foreground(styles.CurrentTheme.Muted).
			Render(strings.Join(outputLines, "\n"))
		
		output = "\n\n" + outputTitle + "\n" + outputContent
	}
	
	// Error message
	var errorMsg string
	if selectedTask.Error != nil {
		errorMsg = "\n\n" + styles.ErrorStyle().Render("Error: " + selectedTask.Error.Error())
	}
	
	content := info + progressBar + output + errorMsg
	
	return styles.BoxStyle().
		Width(im.width).
		Height(height).
		Render(title + "\n\n" + content)
}

func (im *InstallManager) renderStatusBar() string {
	pending := 0
	running := 0
	completed := 0
	failed := 0
	
	// Count tasks by status
	if im.taskProvider != nil {
		for _, taskID := range im.taskIDs {
			if task, ok := im.taskProvider.GetTask(taskID); ok {
				switch task.Status {
				case TaskStatusPending:
					pending++
				case TaskStatusRunning:
					running++
				case TaskStatusCompleted:
					completed++
				case TaskStatusFailed, TaskStatusCancelled:
					failed++
				}
			}
		}
	}
	
	status := fmt.Sprintf(
		"Queue: %d pending | %d running | %d completed | %d failed",
		pending, running, completed, failed,
	)
	
	return styles.StatusBarStyle().
		Width(im.width).
		Render(status)
}

func (im *InstallManager) renderHelpBar() string {
	helps := []string{
		"[s] Start",
		"[c] Cancel Current",
		"[C] Cancel All",
		"[o] Toggle Output",
		"[Enter] Details",
	}
	
	return styles.MutedStyle().Render(strings.Join(helps, "  "))
}

// Helper methods

func (im *InstallManager) getSelectedTask() *TaskInfo {
	if im.taskProvider == nil || im.cursor < 0 || im.cursor >= len(im.taskIDs) {
		return nil
	}
	
	taskID := im.taskIDs[im.cursor]
	task, _ := im.taskProvider.GetTask(taskID)
	return task
}

func (im *InstallManager) moveUp() {
	if im.cursor > 0 {
		im.cursor--
	}
}

func (im *InstallManager) moveDown() {
	if im.cursor < len(im.taskIDs)-1 {
		im.cursor++
	}
}

func (im *InstallManager) selectCurrentTask() {
	if task := im.getSelectedTask(); task != nil {
		im.selectedTaskID = task.ID
	}
}

func (im *InstallManager) startPendingTasks() tea.Cmd {
	return func() tea.Msg {
		// Start all pending tasks
		if im.taskProvider != nil {
			for _, taskID := range im.taskIDs {
				if task, ok := im.taskProvider.GetTask(taskID); ok {
					if task.Status == TaskStatusPending {
						im.taskProvider.StartTask(taskID)
					}
				}
			}
		}
		return nil
	}
}

func (im *InstallManager) findTask(taskID string) *TaskInfo {
	if im.taskProvider != nil {
		if task, ok := im.taskProvider.GetTask(taskID); ok {
			return task
		}
	}
	return nil
}

func (im *InstallManager) getStatusText(status TaskStatus) string {
	switch status {
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

func formatTaskDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}