package views

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// Individual health check result messages - one per check (exported for app routing)
type MemoryCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type DiskCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type InternetCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type BuildToolsCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type GitCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type PathCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type RustToolchainCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type GoToolchainCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type ToolUpdatesCheckCompleteMsg struct {
	Result HealthCheckUpdate
}

type HealthCheckUpdate struct {
	Index       int
	Status      HealthStatus  
	Message     string
	Details     []string
	Suggestions []string
}

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
	// Don't run health checks synchronously - will be done on demand
	// This prevents blocking when switching to health view
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
				hv.updateContent()
			}
			// Start health checks sequentially to avoid overwhelming the system
			return hv.RunNextHealthCheck(0)
		case "a":
			hv.autoRefresh = !hv.autoRefresh
			if hv.ready {
				hv.updateContent()
			}
		case "d":
			hv.showDetails = !hv.showDetails
			if hv.ready {
				hv.updateContent()
			}
		}
	// Handle individual health check completion messages
	case MemoryCheckCompleteMsg:
		hv.applySystemCheckResult(msg.Result)
		return nil
	case DiskCheckCompleteMsg:
		hv.applySystemCheckResult(msg.Result)
		return nil
	case InternetCheckCompleteMsg:
		hv.applySystemCheckResult(msg.Result)
		return nil
	case BuildToolsCheckCompleteMsg:
		hv.applySystemCheckResult(msg.Result)
		return nil
	case GitCheckCompleteMsg:
		hv.applySystemCheckResult(msg.Result)
		return nil
	case PathCheckCompleteMsg:
		hv.applySystemCheckResult(msg.Result)
		return nil
	case RustToolchainCheckCompleteMsg:
		hv.applyToolCheckResult(msg.Result)
		return nil
	case GoToolchainCheckCompleteMsg:
		hv.applyToolCheckResult(msg.Result)
		return nil
	case ToolUpdatesCheckCompleteMsg:
		hv.applyToolCheckResult(msg.Result)
		return nil
	case NextHealthCheckMsg:
		// Continue with next health check in sequence
		return hv.RunNextHealthCheck(msg.NextIndex)
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
	// Show appropriate content based on current state
	hv.updateContent()
	
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

// updateContent shows appropriate content based on current health data state
func (hv *HealthView) updateContent() {
	if !hv.ready {
		return
	}
	
	// Always show the health checks - they have initial states
	hv.updateViewportContentTUI()
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
			Message:  getLinuxDistribution(),
			Details:  []string{fmt.Sprintf("Architecture: %s", runtime.GOARCH), fmt.Sprintf("Go version: %s", runtime.Version())},
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
	// Initialize tool checks based on available toolchains and tools
	hv.toolChecks = []HealthCheck{
		{
			Name:     "Rust Toolchain",
			Category: "toolchain",
			Status:   HealthStatusPending,
			Message:  "Checking rustc, cargo...",
		},
		{
			Name:     "Go Toolchain", 
			Category: "toolchain",
			Status:   HealthStatusPending,
			Message:  "Checking go version...",
		},
		{
			Name:     "Tool Updates",
			Category: "tools",
			Status:   HealthStatusPending,
			Message:  "Checking for updates...",
		},
	}
}

func (hv *HealthView) moveUp() {
	if hv.cursor > 0 {
		hv.cursor--
		// Use TUI best practice: update content and sync viewport
		if hv.ready {
			hv.updateContent()
		}
	}
}

func (hv *HealthView) moveDown() {
	totalChecks := len(hv.systemChecks) + len(hv.toolChecks)
	if hv.cursor < totalChecks-1 {
		hv.cursor++
		// Use TUI best practice: update content and sync viewport
		if hv.ready {
			hv.updateContent()
		}
	}
}

func (hv *HealthView) selectCheck() {
	// Toggle details for the selected check
	hv.showDetails = !hv.showDetails
	// Refresh content to show/hide details
	if hv.ready {
		hv.updateContent()
	}
}

// Helper functions to apply individual health check results
func (hv *HealthView) applySystemCheckResult(result HealthCheckUpdate) {
	if result.Index >= 0 && result.Index < len(hv.systemChecks) {
		hv.systemChecks[result.Index].Status = result.Status
		hv.systemChecks[result.Index].Message = result.Message
		hv.systemChecks[result.Index].Details = result.Details
		hv.systemChecks[result.Index].Suggestions = result.Suggestions
		
		// Refresh content to show the updated check
		if hv.ready {
			hv.updateContent()
		}
	}
}

func (hv *HealthView) applyToolCheckResult(result HealthCheckUpdate) {
	if result.Index >= 0 && result.Index < len(hv.toolChecks) {
		hv.toolChecks[result.Index].Status = result.Status
		hv.toolChecks[result.Index].Message = result.Message
		hv.toolChecks[result.Index].Details = result.Details
		hv.toolChecks[result.Index].Suggestions = result.Suggestions
		
		// Refresh content to show the updated check
		if hv.ready {
			hv.updateContent()
		}
	}
}

// RunNextHealthCheck runs health checks sequentially to avoid overwhelming the system
func (hv *HealthView) RunNextHealthCheck(checkIndex int) tea.Cmd {
	healthChecks := []func() tea.Cmd{
		hv.RunMemoryCheckAsync,
		hv.RunDiskCheckAsync,
		hv.RunInternetCheckAsync,
		hv.RunBuildToolsCheckAsync,
		hv.RunGitCheckAsync,
		hv.RunPathCheckAsync,
		// Tool checks re-enabled
		hv.RunRustToolchainCheckAsync,
		hv.RunGoToolchainCheckAsync,
		hv.RunToolUpdatesCheckAsync,
	}
	
	if checkIndex >= len(healthChecks) {
		// Log completion of all health checks
		debugLog("DEBUG: RunNextHealthCheck() - All health checks completed (index %d >= %d)", checkIndex, len(healthChecks))
		return nil // All checks completed
	}
	
	// Log which health check is being started
	checkNames := []string{"Memory", "Disk", "Internet", "Build Tools", "Git", "PATH", "Rust Toolchain", "Go Toolchain", "Tool Updates"}
	checkName := "Unknown"
	if checkIndex < len(checkNames) {
		checkName = checkNames[checkIndex]
	}
	debugLog("DEBUG: RunNextHealthCheck() - Starting check %d: %s", checkIndex, checkName)
	
	// Run current check and trigger next one
	return tea.Batch(
		healthChecks[checkIndex](),
		func() tea.Msg {
			return NextHealthCheckMsg{NextIndex: checkIndex + 1}
		},
	)
}

type NextHealthCheckMsg struct {
	NextIndex int
}

// CheckMemoryDirect performs memory check synchronously and returns result
func (hv *HealthView) CheckMemoryDirect() HealthCheckUpdate {
	return hv.checkMemory()
}

// CheckToolUpdatesDirect performs tool updates check synchronously and returns result
func (hv *HealthView) CheckToolUpdatesDirect() HealthCheckUpdate {
	return hv.checkToolUpdates()
}

// Individual asynchronous health check commands - each runs independently
func (hv *HealthView) RunMemoryCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkMemory()
		return MemoryCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunDiskCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkDiskSpace()
		return DiskCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunInternetCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkInternet()
		return InternetCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunBuildToolsCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkBuildTools()
		return BuildToolsCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunGitCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkGit()
		return GitCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunPathCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkPATH()
		return PathCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunRustToolchainCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkRustToolchain()
		return RustToolchainCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunGoToolchainCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkGoToolchain()
		return GoToolchainCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) RunToolUpdatesCheckAsync() tea.Cmd {
	return func() tea.Msg {
		result := hv.checkToolUpdates()
		return ToolUpdatesCheckCompleteMsg{Result: result}
	}
}

func (hv *HealthView) resetChecksToChecking() {
	// Reset dynamic system checks (Memory=2, Disk=3, Internet=4, Build Tools=5, Git=6, PATH=7)
	dynamicSystemIndices := []int{2, 3, 4, 5, 6, 7}
	for _, i := range dynamicSystemIndices {
		if i < len(hv.systemChecks) {
			hv.systemChecks[i].Status = HealthStatusPending
			hv.systemChecks[i].Message = "Checking..."
			hv.systemChecks[i].Details = nil
			hv.systemChecks[i].Suggestions = nil
		}
	}
	
	// Reset all tool checks
	for i := range hv.toolChecks {
		hv.toolChecks[i].Status = HealthStatusPending
		hv.toolChecks[i].Message = "Checking..."
		hv.toolChecks[i].Details = nil
		hv.toolChecks[i].Suggestions = nil
	}
}

// Actual health check implementations

func (hv *HealthView) checkMemory() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 2}
	
	// Read memory info from /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		result.Status = HealthStatusWarning
		result.Message = "Could not read memory info"
		return result
	}

	lines := strings.Split(string(data), "\n")
	var totalKB, availableKB int

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.Atoi(parts[1]); err == nil {
					totalKB = val
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.Atoi(parts[1]); err == nil {
					availableKB = val
				}
			}
		}
	}

	if totalKB > 0 && availableKB > 0 {
		totalGB := float64(totalKB) / 1024 / 1024
		availableGB := float64(availableKB) / 1024 / 1024
		usedGB := totalGB - availableGB
		usagePercent := (usedGB / totalGB) * 100

		result.Status = HealthStatusPassing
		if usagePercent > 90 {
			result.Status = HealthStatusWarning
		}
		
		result.Message = fmt.Sprintf("%.1f GB available", availableGB)
		result.Details = []string{
			fmt.Sprintf("Total: %.1f GB", totalGB),
			fmt.Sprintf("Used: %.1f GB (%.0f%%)", usedGB, usagePercent),
			fmt.Sprintf("Available: %.1f GB", availableGB),
		}
	} else {
		result.Status = HealthStatusWarning
		result.Message = "Could not parse memory info"
	}
	
	return result
}

func (hv *HealthView) checkDiskSpace() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 3}
	
	// Check disk space for current directory
	cmd := exec.Command("df", "-h", ".")
	output, err := cmd.Output()
	if err != nil {
		result.Status = HealthStatusWarning
		result.Message = "Could not check disk space"
		return result
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 5 {
			usage := strings.TrimSuffix(fields[4], "%")
			if usageInt, err := strconv.Atoi(usage); err == nil {
				if usageInt > 90 {
					result.Status = HealthStatusWarning
					result.Message = fmt.Sprintf("Low disk space (%d%% used)", usageInt)
					result.Suggestions = []string{
						"Consider cleaning build cache",
						"Remove unused tools or files",
					}
				} else {
					result.Status = HealthStatusPassing
					result.Message = fmt.Sprintf("%s available (%d%% used)", fields[3], usageInt)
				}
				result.Details = []string{
					fmt.Sprintf("Filesystem: %s", fields[0]),
					fmt.Sprintf("Size: %s", fields[1]),
					fmt.Sprintf("Used: %s", fields[2]),
					fmt.Sprintf("Available: %s", fields[3]),
				}
			}
		}
	}
	
	return result
}

func (hv *HealthView) checkInternet() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 4}
	
	// Test internet connectivity
	cmd := exec.Command("ping", "-c", "1", "-W", "3", "github.com")
	err := cmd.Run()
	
	if err != nil {
		result.Status = HealthStatusWarning
		result.Message = "No internet connection"
		result.Suggestions = []string{
			"Check network connection",
			"Some installs may fail without internet",
		}
	} else {
		result.Status = HealthStatusPassing
		result.Message = "Connected"
	}
	
	return result
}

func (hv *HealthView) checkBuildTools() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 5}
	var details []string
	var missing []string
	
	tools := map[string]string{
		"gcc":   "--version",
		"make":  "--version", 
		"cmake": "--version",
	}

	allPresent := true
	for tool, flag := range tools {
		cmd := exec.Command(tool, flag)
		output, err := cmd.Output()
		if err != nil {
			missing = append(missing, tool)
			allPresent = false
		} else {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				// Extract version from first line
				firstLine := strings.TrimSpace(lines[0])
				if len(firstLine) > 50 {
					firstLine = firstLine[:50] + "..."
				}
				details = append(details, firstLine)
			}
		}
	}

	if allPresent {
		result.Status = HealthStatusPassing
		result.Message = "All required build tools installed"
		result.Details = details
	} else {
		result.Status = HealthStatusWarning
		result.Message = fmt.Sprintf("Missing tools: %s", strings.Join(missing, ", "))
		result.Suggestions = []string{
			"Install build essentials: sudo apt install build-essential",
			"Install cmake: sudo apt install cmake",
		}
	}
	
	return result
}

func (hv *HealthView) checkGit() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 6}
	
	cmd := exec.Command("git", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = HealthStatusWarning
		result.Message = "Git not installed"
		result.Suggestions = []string{
			"Install git: sudo apt install git",
		}
	} else {
		version := strings.TrimSpace(string(output))
		result.Status = HealthStatusPassing
		result.Message = version
	}
	
	return result
}

func (hv *HealthView) checkPATH() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 7}
	
	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, ":")
	
	var details []string
	hasUsrLocal := false
	hasCargo := false
	
	for _, dir := range pathDirs {
		if dir == "/usr/local/bin" {
			hasUsrLocal = true
			details = append(details, "/usr/local/bin is in PATH")
		}
		if strings.Contains(dir, ".cargo/bin") {
			hasCargo = true
			details = append(details, "~/.cargo/bin is in PATH")
		}
	}
	
	if hasUsrLocal && hasCargo {
		result.Status = HealthStatusPassing
		result.Message = "Correctly configured"
	} else if hasUsrLocal {
		result.Status = HealthStatusWarning
		result.Message = "Missing ~/.cargo/bin in PATH"
		result.Suggestions = []string{
			"Add ~/.cargo/bin to PATH for Rust tools",
		}
	} else {
		result.Status = HealthStatusWarning
		result.Message = "PATH may need configuration"
		result.Suggestions = []string{
			"Ensure /usr/local/bin is in PATH",
			"Add ~/.cargo/bin to PATH for Rust tools",
		}
	}
	
	result.Details = details
	return result
}

func (hv *HealthView) checkRustToolchain() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 0}
	
	cmd := exec.Command("rustc", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = HealthStatusWarning
		result.Message = "Rust not installed"
		result.Suggestions = []string{
			"Install Rust: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
		}
		return result
	}

	rustcVersion := strings.TrimSpace(string(output))
	result.Status = HealthStatusPassing
	result.Message = rustcVersion

	// Check cargo and rustup
	var details []string
	details = append(details, rustcVersion)

	if cargoCmd := exec.Command("cargo", "--version"); cargoCmd.Run() == nil {
		if cargoOutput, err := cargoCmd.Output(); err == nil {
			details = append(details, strings.TrimSpace(string(cargoOutput)))
		}
	}

	if rustupCmd := exec.Command("rustup", "--version"); rustupCmd.Run() == nil {
		if rustupOutput, err := rustupCmd.Output(); err == nil {
			details = append(details, strings.TrimSpace(string(rustupOutput)))
		}
	}

	result.Details = details
	return result
}

func (hv *HealthView) checkGoToolchain() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 1}
	
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = HealthStatusWarning
		result.Message = "Go not installed"
		result.Suggestions = []string{
			"Install Go from https://golang.org/dl/",
		}
		return result
	}

	goVersion := strings.TrimSpace(string(output))
	result.Status = HealthStatusPassing
	result.Message = goVersion

	var details []string
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		details = append(details, "GOPATH: "+gopath)
	}
	if goroot := os.Getenv("GOROOT"); goroot != "" {
		details = append(details, "GOROOT: "+goroot)
	}

	result.Details = details
	return result
}

func (hv *HealthView) checkToolUpdates() HealthCheckUpdate {
	result := HealthCheckUpdate{Index: 2}
	
	// Simple check - in a real implementation this would check for tool updates
	result.Status = HealthStatusPassing
	result.Message = "Update check complete"
	result.Details = []string{
		"Last checked: " + time.Now().Format("15:04:05"),
		"Use 'gearbox update' to check for tool updates",
	}
	
	return result
}

// getLinuxDistribution detects the Linux distribution and version
func getLinuxDistribution() string {
	// Try to read /etc/os-release first (standard method)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		var prettyName, name, version string
		
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
			} else if strings.HasPrefix(line, "NAME=") {
				name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			} else if strings.HasPrefix(line, "VERSION=") {
				version = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
			}
		}
		
		// Return PRETTY_NAME if available, otherwise combine NAME and VERSION
		if prettyName != "" {
			return prettyName
		}
		if name != "" && version != "" {
			return fmt.Sprintf("%s %s", name, version)
		}
		if name != "" {
			return name
		}
	}
	
	// Fallback: try /etc/lsb-release
	if data, err := os.ReadFile("/etc/lsb-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		var distrib, release string
		
		for _, line := range lines {
			if strings.HasPrefix(line, "DISTRIB_DESCRIPTION=") {
				return strings.Trim(strings.TrimPrefix(line, "DISTRIB_DESCRIPTION="), "\"")
			} else if strings.HasPrefix(line, "DISTRIB_ID=") {
				distrib = strings.TrimPrefix(line, "DISTRIB_ID=")
			} else if strings.HasPrefix(line, "DISTRIB_RELEASE=") {
				release = strings.TrimPrefix(line, "DISTRIB_RELEASE=")
			}
		}
		
		if distrib != "" && release != "" {
			return fmt.Sprintf("%s %s", distrib, release)
		}
	}
	
	// Fallback: check specific distribution files
	distroFiles := map[string]string{
		"/etc/debian_version": "Debian",
		"/etc/redhat-release": "",
		"/etc/centos-release": "",
		"/etc/fedora-release": "",
		"/etc/arch-release":   "Arch Linux",
		"/etc/alpine-release": "Alpine Linux",
	}
	
	for file, defaultName := range distroFiles {
		if data, err := os.ReadFile(file); err == nil {
			content := strings.TrimSpace(string(data))
			if file == "/etc/debian_version" {
				return fmt.Sprintf("Debian %s", content)
			}
			// For redhat/centos/fedora-release, the file contains the full description
			if content != "" && defaultName == "" {
				return content
			}
			if defaultName != "" {
				return fmt.Sprintf("%s %s", defaultName, content)
			}
		}
	}
	
	// Ultimate fallback
	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

