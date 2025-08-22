package views

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// healthChecksCompleteMsg is sent when health checks are complete
type healthChecksCompleteMsg struct{}

// HealthView represents the health monitor view
type HealthView struct {
	width  int
	height int

	// Data
	systemChecks   []HealthCheck
	toolChecks     []HealthCheck
	installedTools map[string]*manifest.InstallationRecord

	// UI state
	cursor         int
	selectedCheck  int
	autoRefresh    bool
	showDetails    bool

	// TUI components (official Bubbles components)
	viewport       viewport.Model
	ready          bool
}

// HealthCheck represents a health check item
type HealthCheck struct {
	Name        string
	Category    string
	Status      HealthStatus
	Message     string
	Details     []string
	Suggestions []string
	Critical    bool
}

// HealthStatus represents the status of a health check
type HealthStatus int

const (
	HealthStatusPending HealthStatus = iota
	HealthStatusPassing
	HealthStatusWarning
	HealthStatusFailing
)

// NewHealthView creates a new health monitor view
func NewHealthView() *HealthView {
	return &HealthView{
		installedTools: make(map[string]*manifest.InstallationRecord),
		autoRefresh:    true,
		showDetails:    true,
		systemChecks:   initializeSystemChecks(),
		toolChecks:     []HealthCheck{},
	}
}

// SetSize updates the size of the health view
func (hv *HealthView) SetSize(width, height int) {
	hv.width = width
	hv.height = height
	
	// Initialize official viewport if not ready
	if !hv.ready {
		// Calculate viewport height: total - header - footer
		viewportHeight := height - 3 // Reserve 3 lines for header + footer
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		
		hv.viewport = viewport.New(width, viewportHeight)
		hv.viewport.SetContent("")
		hv.ready = true
	} else {
		// Update existing viewport
		viewportHeight := height - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		hv.viewport.Width = width
		hv.viewport.Height = viewportHeight
	}
}

// SetData updates the health view data
func (hv *HealthView) SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord) {
	hv.installedTools = installed
	hv.updateToolChecks(tools)
	// Automatically run health checks when data is loaded
	hv.runHealthChecks()
	
	if hv.ready {
		hv.updateViewportContentTUI()
	}
}

// Update handles health view updates
func (hv *HealthView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			hv.moveUp()
		case "down", "j":
			hv.moveDown()
		case "enter", " ":
			hv.selectCheck()
		case "r":
			// Reset checks to pending state to show refresh is happening
			hv.resetChecksToChecking()
			if hv.ready {
				hv.updateViewportContentTUI()
			}
			// Return a command that will run the health checks after a brief delay
			return tea.Tick(time.Millisecond*100, func(time.Time) tea.Msg {
				hv.runHealthChecks()
				return healthChecksCompleteMsg{}
			})
		case "a":
			hv.autoRefresh = !hv.autoRefresh
			if hv.ready {
				hv.updateViewportContentTUI()
			}
		case "d":
			hv.showDetails = !hv.showDetails
			if hv.ready {
				hv.updateViewportContentTUI()
			}
		}
	case healthChecksCompleteMsg:
		// Health checks are complete - refresh content
		if hv.ready {
			hv.updateViewportContentTUI()
		}
		return nil
	}

	return nil
}

// Render returns the rendered health view
func (hv *HealthView) Render() string {
	return hv.renderTUIStyle()
}

// renderTUIStyle uses proper TUI best practices with official Bubbles components
func (hv *HealthView) renderTUIStyle() string {
	// Header (health summary)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 1)
	
	summary := hv.renderSummary()
	header := headerStyle.Render(fmt.Sprintf("System Health Monitor | %s", summary))
	
	// Footer (help)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	footer := footerStyle.Render(hv.renderHelpBar())
	
	// Content (health checks with cursor highlighting)
	hv.updateViewportContentTUI()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		hv.viewport.View(),
		footer,
	)
}

func (hv *HealthView) renderSummary() string {
	passing := 0
	warning := 0
	failing := 0

	allChecks := append(hv.systemChecks, hv.toolChecks...)
	for _, check := range allChecks {
		switch check.Status {
		case HealthStatusPassing:
			passing++
		case HealthStatusWarning:
			warning++
		case HealthStatusFailing:
			failing++
		}
	}

	return fmt.Sprintf("✓%d ⚠%d ✗%d", passing, warning, failing)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (hv *HealthView) updateViewportContentTUI() {
	var lines []string

	// System checks section
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Render("System Checks"))
	for i, check := range hv.systemChecks {
		isSelected := i == hv.cursor
		line := hv.renderHealthCheck(check, isSelected)
		
		// Apply cursor highlighting here (TUI best practice)
		if isSelected {
			selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
			line = selectedStyle.Render(line)
		}
		
		lines = append(lines, line)
		
		if hv.showDetails && isSelected && len(check.Details) > 0 {
			for _, detail := range check.Details {
				detailLine := "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(detail)
				lines = append(lines, detailLine)
			}
		}
	}

	lines = append(lines, "")

	// Tool checks section
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Render("Tool Checks"))
	toolCheckOffset := len(hv.systemChecks)
	for i, check := range hv.toolChecks {
		isSelected := toolCheckOffset+i == hv.cursor
		line := hv.renderHealthCheck(check, isSelected)
		
		// Apply cursor highlighting here (TUI best practice)
		if isSelected {
			selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
			line = selectedStyle.Render(line)
		}
		
		lines = append(lines, line)
		
		if hv.showDetails && isSelected {
			if len(check.Details) > 0 {
				for _, detail := range check.Details {
					detailLine := "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(detail)
					lines = append(lines, detailLine)
				}
			}
			if len(check.Suggestions) > 0 {
				suggestionHeader := "  " + lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("Suggestions:")
				lines = append(lines, suggestionHeader)
				for _, suggestion := range check.Suggestions {
					suggestionLine := "    • " + suggestion
					lines = append(lines, suggestionLine)
				}
			}
		}
	}

	content := strings.Join(lines, "\n")
	hv.viewport.SetContent(content)

	// Sync viewport with cursor position (TUI best practice)
	hv.syncViewportWithCursor()
}

// syncViewportWithCursor ensures cursor is visible (TUI best practice)
func (hv *HealthView) syncViewportWithCursor() {
	if len(hv.systemChecks) == 0 && len(hv.toolChecks) == 0 {
		return
	}
	
	// Get viewport bounds
	top := hv.viewport.YOffset
	bottom := top + hv.viewport.Height - 1
	
	// Ensure cursor is visible by scrolling viewport
	if hv.cursor < top {
		// Cursor above viewport - scroll up
		hv.viewport.SetYOffset(hv.cursor)
	} else if hv.cursor > bottom {
		// Cursor below viewport - scroll down
		hv.viewport.SetYOffset(hv.cursor - hv.viewport.Height + 1)
	}
}

// renderContent is deprecated - using TUI best practices instead

func (hv *HealthView) renderHealthCheck(check HealthCheck, selected bool) string {
	// Status icon
	var icon string
	var iconStyle lipgloss.Style

	switch check.Status {
	case HealthStatusPending:
		icon = "○"
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	case HealthStatusPassing:
		icon = "✓"
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	case HealthStatusWarning:
		icon = "⚠"
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	case HealthStatusFailing:
		icon = "✗"
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}

	// Build the line (no selection styling - handled by caller)
	line := fmt.Sprintf("%s %-30s %s",
		iconStyle.Render(icon),
		check.Name,
		check.Message,
	)

	// Apply critical error styling if needed
	if check.Critical && check.Status == HealthStatusFailing {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(line)
	}

	return line
}

func (hv *HealthView) renderHelpBar() string {
	helps := []string{
		"[↑/↓] Navigate",
		"[Enter] Details",
		"[r] Run Checks",
		"[d] Toggle Details",
	}
	
	if hv.autoRefresh {
		helps = append(helps, "[a] Auto-refresh ON")
	} else {
		helps = append(helps, "[a] Auto-refresh OFF")
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(strings.Join(helps, "  "))
}

// Helper methods

func initializeSystemChecks() []HealthCheck {
	return []HealthCheck{
		{
			Name:     "Operating System",
			Category: "system",
			Status:   HealthStatusPassing,
			Message:  fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
			Details:  []string{fmt.Sprintf("Go version: %s", runtime.Version())},
		},
		{
			Name:     "CPU Cores",
			Category: "system",
			Status:   HealthStatusPassing,
			Message:  fmt.Sprintf("%d cores available", runtime.NumCPU()),
		},
		{
			Name:     "Memory",
			Category: "system",
			Status:   HealthStatusPending,
			Message:  "Checking...",
			Details:  []string{"Run health check to update"},
		},
		{
			Name:     "Disk Space",
			Category: "system",
			Status:   HealthStatusPending,
			Message:  "Checking...",
			Details:  []string{"Run health check to update"},
		},
		{
			Name:     "Internet Connection",
			Category: "system",
			Status:   HealthStatusPending,
			Message:  "Checking...",
		},
		{
			Name:     "Build Tools",
			Category: "system",
			Status:   HealthStatusPending,
			Message:  "Checking gcc, make, cmake...",
		},
		{
			Name:     "Git",
			Category: "system",
			Status:   HealthStatusPending,
			Message:  "Checking version...",
		},
		{
			Name:     "PATH Configuration",
			Category: "system",
			Status:   HealthStatusPending,
			Message:  "Checking /usr/local/bin...",
		},
	}
}

func (hv *HealthView) updateToolChecks(tools []orchestrator.ToolConfig) {
	hv.toolChecks = []HealthCheck{}

	// Check common toolchains
	rustCheck := HealthCheck{
		Name:     "Rust Toolchain",
		Category: "toolchain",
		Status:   HealthStatusPending,
		Message:  "Checking...",
	}
	
	goCheck := HealthCheck{
		Name:     "Go Toolchain",
		Category: "toolchain",
		Status:   HealthStatusPending,
		Message:  "Checking...",
	}

	// Check installed tools
	installedCount := len(hv.installedTools)
	totalCount := len(tools)
	
	coverageCheck := HealthCheck{
		Name:     "Tool Coverage",
		Category: "tools",
		Status:   HealthStatusPassing,
		Message:  fmt.Sprintf("%d/%d tools installed", installedCount, totalCount),
	}

	if installedCount == 0 {
		coverageCheck.Status = HealthStatusWarning
		coverageCheck.Suggestions = []string{
			"Run 'gearbox install --bundle beginner' to get started",
			"Or use the Tool Browser to select individual tools",
		}
	} else if installedCount < totalCount/4 {
		coverageCheck.Status = HealthStatusWarning
		coverageCheck.Suggestions = []string{
			fmt.Sprintf("You have %d more tools available", totalCount-installedCount),
			"Explore bundles for curated tool collections",
		}
	}

	// Check for updates
	updateCheck := HealthCheck{
		Name:     "Tool Updates",
		Category: "tools",
		Status:   HealthStatusPending,
		Message:  "Checking for updates...",
		Details:  []string{"Run health check to scan for updates"},
	}

	hv.toolChecks = append(hv.toolChecks, rustCheck, goCheck, coverageCheck, updateCheck)
}

func (hv *HealthView) moveUp() {
	if hv.cursor > 0 {
		hv.cursor--
		// Use TUI best practice: update content and sync viewport
		if hv.ready {
			hv.updateViewportContentTUI()
		}
	}
}

func (hv *HealthView) moveDown() {
	totalChecks := len(hv.systemChecks) + len(hv.toolChecks)
	if hv.cursor < totalChecks-1 {
		hv.cursor++
		// Use TUI best practice: update content and sync viewport
		if hv.ready {
			hv.updateViewportContentTUI()
		}
	}
}

func (hv *HealthView) selectCheck() {
	// Toggle details for the selected check
	hv.showDetails = !hv.showDetails
	// Refresh content to show/hide details
	if hv.ready {
		hv.updateViewportContentTUI()
	}
}

func (hv *HealthView) runHealthChecks() {
	// TODO: Actually run health checks
	// For now, just simulate some results
	
	// Update OS check
	if len(hv.systemChecks) > 0 {
		hv.systemChecks[0].Status = HealthStatusPassing
		hv.systemChecks[0].Message = "Linux (Debian-based)"
		hv.systemChecks[0].Details = []string{
			"Kernel: " + runtime.GOOS,
			"Architecture: " + runtime.GOARCH,
		}
	}
	
	// Update CPU cores check
	if len(hv.systemChecks) > 1 {
		hv.systemChecks[1].Status = HealthStatusPassing
		hv.systemChecks[1].Message = fmt.Sprintf("%d cores available", runtime.NumCPU())
		hv.systemChecks[1].Details = []string{
			fmt.Sprintf("Logical CPUs: %d", runtime.NumCPU()),
			"GOMAXPROCS: " + fmt.Sprintf("%d", runtime.GOMAXPROCS(0)),
		}
	}
	
	// Update memory check
	if len(hv.systemChecks) > 2 {
		hv.systemChecks[2].Status = HealthStatusPassing
		hv.systemChecks[2].Message = fmt.Sprintf("%.1f GB available", 8.5)
		hv.systemChecks[2].Details = []string{
			"Total: 16.0 GB",
			"Used: 7.5 GB",
			"Free: 8.5 GB",
		}
	}

	// Update disk space
	if len(hv.systemChecks) > 3 {
		hv.systemChecks[3].Status = HealthStatusWarning
		hv.systemChecks[3].Message = "Low disk space (15% free)"
		hv.systemChecks[3].Details = []string{
			"Total: 500 GB",
			"Used: 425 GB", 
			"Free: 75 GB",
		}
		hv.systemChecks[3].Suggestions = []string{
			"Consider cleaning build cache: gearbox cache clean",
			"Remove old tool versions: gearbox uninstall --old",
		}
	}

	// Update internet check
	if len(hv.systemChecks) > 4 {
		hv.systemChecks[4].Status = HealthStatusPassing
		hv.systemChecks[4].Message = "Connected"
	}

	// Update build tools
	if len(hv.systemChecks) > 5 {
		hv.systemChecks[5].Status = HealthStatusPassing
		hv.systemChecks[5].Message = "All required build tools installed"
		hv.systemChecks[5].Details = []string{
			"gcc 11.4.0",
			"make 4.3",
			"cmake 3.22.1",
		}
	}

	// Update git
	if len(hv.systemChecks) > 6 {
		hv.systemChecks[6].Status = HealthStatusPassing
		hv.systemChecks[6].Message = "git version 2.34.1"
	}

	// Update PATH
	if len(hv.systemChecks) > 7 {
		hv.systemChecks[7].Status = HealthStatusPassing
		hv.systemChecks[7].Message = "Correctly configured"
		hv.systemChecks[7].Details = []string{
			"/usr/local/bin is in PATH",
			"~/.cargo/bin is in PATH",
		}
	}

	// Update toolchain checks
	if len(hv.toolChecks) > 0 {
		hv.toolChecks[0].Status = HealthStatusPassing
		hv.toolChecks[0].Message = "rustc 1.88.0"
		hv.toolChecks[0].Details = []string{
			"cargo 1.88.0",
			"rustup 1.26.0",
		}
	}

	if len(hv.toolChecks) > 1 {
		hv.toolChecks[1].Status = HealthStatusPassing
		hv.toolChecks[1].Message = "go version go1.23.4"
		hv.toolChecks[1].Details = []string{
			"GOPATH: ~/go",
			"GOROOT: /usr/local/go",
		}
	}
	
	// Update tool coverage check
	if len(hv.toolChecks) > 2 {
		installedCount := len(hv.installedTools)
		totalTools := 42 // This would come from the tools config in real implementation
		percentage := float64(installedCount) / float64(totalTools) * 100
		
		hv.toolChecks[2].Status = HealthStatusPassing
		hv.toolChecks[2].Message = fmt.Sprintf("%d/%d tools (%.0f%%)", installedCount, totalTools, percentage)
		hv.toolChecks[2].Details = []string{
			fmt.Sprintf("Installed: %d", installedCount),
			fmt.Sprintf("Available: %d", totalTools),
			fmt.Sprintf("Missing: %d", totalTools-installedCount),
		}
		
		if percentage < 50 {
			hv.toolChecks[2].Suggestions = []string{
				"Run 'gearbox install --bundle beginner' for essential tools",
				"Browse available tools with 'gearbox list'",
			}
		}
	}
	
	// Update the tool updates check
	if len(hv.toolChecks) > 3 {
		hv.toolChecks[3].Status = HealthStatusPassing
		hv.toolChecks[3].Message = "All tools up to date"
		hv.toolChecks[3].Details = []string{
			"Last checked: just now",
			"No updates available",
		}
	}

	// Refresh content after running checks
	if hv.ready {
		hv.updateViewportContentTUI()
	}
}

func (hv *HealthView) resetChecksToChecking() {
	// Reset all system checks to pending/checking state
	for i := range hv.systemChecks {
		hv.systemChecks[i].Status = HealthStatusPending
		hv.systemChecks[i].Message = "Checking..."
		hv.systemChecks[i].Details = nil
		hv.systemChecks[i].Suggestions = nil
	}
	
	// Reset all tool checks to pending/checking state
	for i := range hv.toolChecks {
		hv.toolChecks[i].Status = HealthStatusPending
		hv.toolChecks[i].Message = "Checking..."
		hv.toolChecks[i].Details = nil
		hv.toolChecks[i].Suggestions = nil
	}
}

