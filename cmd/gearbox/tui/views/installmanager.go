package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstallManagerNew represents the new installation manager with TUI best practices
type InstallManagerNew struct {
	// Task provider
	taskProvider   TaskProvider
	
	// Task IDs from task manager
	taskIDs        []string
	
	// UI state
	cursor         int
	selectedTaskID string
	showOutput     bool
	
	// Progress bars
	progressBars   map[string]progress.Model
	
	// TUI components (official Bubbles components)
	viewport       viewport.Model
	ready          bool
	width          int
	height         int
}

// NewInstallManagerNew creates a new install manager with TUI best practices
func NewInstallManagerNew() *InstallManagerNew {
	return &InstallManagerNew{
		taskIDs:      []string{},
		progressBars: make(map[string]progress.Model),
		showOutput:   true,
	}
}


// SetSize updates the install manager size (TUI best practices)
func (im *InstallManagerNew) SetSize(width, height int) {
	im.width = width
	im.height = height
	
	// Initialize official viewport if not ready
	if !im.ready {
		// Calculate viewport height: total - header - footer
		viewportHeight := height - 3 // Reserve 3 lines for header + footer
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		
		im.viewport = viewport.New(width, viewportHeight)
		im.viewport.SetContent("")
		im.ready = true
	} else {
		// Update existing viewport
		viewportHeight := height - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		im.viewport.Width = width
		im.viewport.Height = viewportHeight
	}
	
	// Update progress bar widths
	for id, prog := range im.progressBars {
		prog.Width = width / 2
		im.progressBars[id] = prog
	}
}

// SetTaskProvider sets the task provider
func (im *InstallManagerNew) SetTaskProvider(provider TaskProvider) {
	im.taskProvider = provider
}

// AddTaskID adds a new task ID to track
func (im *InstallManagerNew) AddTaskID(taskID string) {
	im.taskIDs = append(im.taskIDs, taskID)
	
	// Create progress bar for this task
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50 // Will be updated in SetSize
	im.progressBars[taskID] = prog
	
	if im.ready {
		im.updateViewportContentTUI()
	}
}

// HandleTaskUpdate handles task update messages
func (im *InstallManagerNew) HandleTaskUpdate(taskID string, progress float64) {
	// Update will refresh the content automatically
	if im.ready {
		im.updateViewportContentTUI()
	}
}

// Update handles install manager updates
func (im *InstallManagerNew) Update(msg tea.Msg) tea.Cmd {
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
			newModel, cmd := prog.Update(msg)
			if newProg, ok := newModel.(progress.Model); ok {
				im.progressBars[id] = newProg
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	
	if im.ready {
		im.updateViewportContentTUI()
	}
	
	return tea.Batch(cmds...)
}

// Render returns the rendered install manager view
func (im *InstallManagerNew) Render() string {
	return im.renderTUIStyle()
}

// renderTUIStyle uses proper TUI best practices with official Bubbles components
func (im *InstallManagerNew) renderTUIStyle() string {
	// Header (installation queue info)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 1)
	
	header := headerStyle.Render(fmt.Sprintf(
		"Installation Manager | %d tasks in queue",
		len(im.taskIDs),
	))
	
	// Footer (help)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	footer := footerStyle.Render(
		"[↑/↓] Navigate  [s] Start Tasks  [c] Cancel Current  [o] Toggle Output  [Enter] Details",
	)
	
	// Content (task list with cursor highlighting)
	im.updateViewportContentTUI()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		im.viewport.View(),
		footer,
	)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (im *InstallManagerNew) updateViewportContentTUI() {
	var lines []string
	
	// Get all tasks from task provider
	var allTasks []*TaskInfo
	if im.taskProvider != nil {
		for _, id := range im.taskIDs {
			if task, ok := im.taskProvider.GetTask(id); ok {
				allTasks = append(allTasks, task)
			}
		}
	}
	
	if len(allTasks) == 0 {
		lines = []string{
			"",
			"No tasks in queue",
			"",
			"Add tools from the Tool Browser [T]",
			"",
		}
	} else {
		for i, task := range allTasks {
			line := im.renderTaskItem(task, i == im.cursor)
			lines = append(lines, line)
			
			// Add task details if this task is selected and showOutput is enabled
			if i == im.cursor && im.showOutput {
				details := im.renderTaskDetails(task)
				lines = append(lines, details...)
			}
		}
	}
	
	content := strings.Join(lines, "\n")
	im.viewport.SetContent(content)
	
	// Sync viewport with cursor position (TUI best practice)
	im.syncViewportWithCursor()
}

// syncViewportWithCursor ensures cursor is visible (TUI best practice)
func (im *InstallManagerNew) syncViewportWithCursor() {
	if len(im.taskIDs) == 0 {
		return
	}
	
	// Get viewport bounds
	top := im.viewport.YOffset
	bottom := top + im.viewport.Height - 1
	
	// Ensure cursor is visible by scrolling viewport
	if im.cursor < top {
		// Cursor above viewport - scroll up
		im.viewport.SetYOffset(im.cursor)
	} else if im.cursor > bottom {
		// Cursor below viewport - scroll down
		im.viewport.SetYOffset(im.cursor - im.viewport.Height + 1)
	}
}

// updateQueueContent is deprecated - using TUI best practices instead

func (im *InstallManagerNew) renderTaskItem(task *TaskInfo, selected bool) string {
	// Status icon
	var statusIcon string
	var statusColor lipgloss.Color
	
	switch task.Status {
	case TaskStatusPending:
		statusIcon = "⏳"
		statusColor = lipgloss.Color("11") // Yellow
	case TaskStatusRunning:
		statusIcon = "🔄"
		statusColor = lipgloss.Color("12") // Blue
	case TaskStatusCompleted:
		statusIcon = "✅"
		statusColor = lipgloss.Color("10") // Green
	case TaskStatusFailed:
		statusIcon = "❌"
		statusColor = lipgloss.Color("9")  // Red
	case TaskStatusCancelled:
		statusIcon = "⏹️"
		statusColor = lipgloss.Color("8")  // Gray
	}
	
	statusStyle := lipgloss.NewStyle().Foreground(statusColor)
	
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
	
	return item
}

// renderTaskDetails returns detailed information for a task (used in TUI style)
func (im *InstallManagerNew) renderTaskDetails(task *TaskInfo) []string {
	var details []string
	indent := "    "
	
	// Task information
	details = append(details, indent+"Details:")
	details = append(details, fmt.Sprintf("%sBuild Type: %s", indent, task.BuildType))
	details = append(details, fmt.Sprintf("%sStatus: %s", indent, im.getStatusText(task.Status)))
	details = append(details, fmt.Sprintf("%sStarted: %s", indent, task.StartTime))
	
	if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
		details = append(details, fmt.Sprintf("%sEnded: %s", indent, task.EndTime))
		details = append(details, fmt.Sprintf("%sDuration: %s", indent, task.Duration))
	}
	
	// Progress for running tasks
	if task.Status == TaskStatusRunning {
		if task.Progress > 0 {
			progressText := fmt.Sprintf("%sProgress: %.0f%% - %s", indent, task.Progress*100, task.Stage)
			details = append(details, progressText)
		} else {
			details = append(details, fmt.Sprintf("%sStage: %s", indent, task.Stage))
		}
	}
	
	// Error message
	if task.Error != nil {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		errorMsg := errorStyle.Render(fmt.Sprintf("%sError: %s", indent, task.Error.Error()))
		details = append(details, errorMsg)
	}
	
	// Recent output
	if len(task.Output) > 0 {
		details = append(details, indent+"Recent Output:")
		// Show last few lines of output
		outputLines := task.Output
		maxLines := 3
		if len(outputLines) > maxLines {
			outputLines = outputLines[len(outputLines)-maxLines:]
		}
		
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		for _, line := range outputLines {
			details = append(details, mutedStyle.Render(indent+"  "+line))
		}
	}
	
	details = append(details, "") // Empty line after details
	
	return details
}

func (im *InstallManagerNew) getStatusText(status TaskStatus) string {
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

// Helper methods
func (im *InstallManagerNew) moveUp() {
	if im.cursor > 0 {
		im.cursor--
		// Use TUI best practice: update content and sync viewport
		if im.ready {
			im.updateViewportContentTUI()
		}
	}
}

func (im *InstallManagerNew) moveDown() {
	if im.cursor < len(im.taskIDs)-1 {
		im.cursor++
		// Use TUI best practice: update content and sync viewport
		if im.ready {
			im.updateViewportContentTUI()
		}
	}
}

func (im *InstallManagerNew) selectCurrentTask() {
	if task := im.getSelectedTask(); task != nil {
		im.selectedTaskID = task.ID
	}
}

func (im *InstallManagerNew) getSelectedTask() *TaskInfo {
	if im.taskProvider == nil || im.cursor < 0 || im.cursor >= len(im.taskIDs) {
		return nil
	}
	
	taskID := im.taskIDs[im.cursor]
	task, _ := im.taskProvider.GetTask(taskID)
	return task
}

func (im *InstallManagerNew) findTask(taskID string) *TaskInfo {
	if im.taskProvider != nil {
		if task, ok := im.taskProvider.GetTask(taskID); ok {
			return task
		}
	}
	return nil
}

func (im *InstallManagerNew) startPendingTasks() tea.Cmd {
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

// All deprecated widgets and layout system code removed - using TUI best practices
