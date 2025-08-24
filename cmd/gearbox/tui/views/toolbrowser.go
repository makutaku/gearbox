package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// ToolBrowserNew represents the new tool browser with layout system


type ToolBrowserNew struct {
	// Data
	tools          []orchestrator.ToolConfig
	filteredTools  []orchestrator.ToolConfig
	installedTools map[string]*manifest.InstallationRecord
	
	// UI state
	searchInput    textinput.Model
	cursor         int
	selectedTools  map[string]bool
	searchActive   bool
	showPreview    bool
	selectedCategory string
	categories     []string
	
	// Deprecated layout system removed - using TUI best practices
	
	// TUI components (official Bubbles components)
	viewport       viewport.Model
	ready          bool
	width          int
	height         int
}

// NewToolBrowserNew creates a new tool browser with layout system
func NewToolBrowserNew() *ToolBrowserNew {
	ti := textinput.New()
	ti.Placeholder = "Search tools..."
	ti.CharLimit = 50
	ti.Width = 40
	
	tb := &ToolBrowserNew{
		searchInput:    ti,
		selectedTools:  make(map[string]bool),
		installedTools: make(map[string]*manifest.InstallationRecord),
		showPreview:    true,
		categories:     []string{"All", "Core", "Development", "System", "Text", "Media", "UI"},
		selectedCategory: "All",
		filteredTools:  []orchestrator.ToolConfig{},
	}
	
	return tb
}

// SetSize updates the layout bounds
func (tb *ToolBrowserNew) SetSize(width, height int) {
	tb.width = width
	tb.height = height
	
	// Initialize official viewport if not ready
	if !tb.ready {
		// Calculate viewport height: total - header - footer
		viewportHeight := height - 3 // Reserve 3 lines for header + footer
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		
		tb.viewport = viewport.New(width, viewportHeight)
		tb.viewport.SetContent("")
		tb.ready = true
	} else {
		// Update existing viewport
		viewportHeight := height - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		tb.viewport.Width = width
		tb.viewport.Height = viewportHeight
	}
	
	// Update search input width
	tb.searchInput.Width = min(width-10, 50)
}

// syncViewportWithCursor ensures cursor is visible (TUI best practice)
func (tb *ToolBrowserNew) syncViewportWithCursor() {
	if len(tb.filteredTools) == 0 {
		return
	}
	
	// Get viewport bounds
	top := tb.viewport.YOffset
	bottom := top + tb.viewport.Height - 1
	
	// Ensure cursor is visible by scrolling viewport
	if tb.cursor < top {
		// Cursor above viewport - scroll up
		tb.viewport.SetYOffset(tb.cursor)
	} else if tb.cursor > bottom {
		// Cursor below viewport - scroll down
		tb.viewport.SetYOffset(tb.cursor - tb.viewport.Height + 1)
	}
}

// Update handles tool browser updates
func (tb *ToolBrowserNew) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if tb.searchActive {
			switch msg.String() {
			case "esc":
				tb.searchActive = false
				tb.searchInput.SetValue("")
				tb.searchInput.Blur()
				tb.applyFilters()
				return nil
			case "enter":
				tb.searchActive = false
				tb.searchInput.Blur()
				tb.applyFilters()
				return nil
			default:
				tb.searchInput, cmd = tb.searchInput.Update(msg)
				tb.applyFilters()
				return cmd
			}
		} else {
			switch msg.String() {
			case "/":
				tb.searchActive = true
				tb.searchInput.Focus()
				return textinput.Blink
			case "up", "k":
				tb.moveUp()
			case "down", "j":
				tb.moveDown()
			case " ":
				tb.toggleSelection()
			case "enter":
				// View tool details
			case "c":
				tb.cycleCategory()
			case "p":
				tb.togglePreview()
			case "a":
				tb.selectAll()
			case "A":
				tb.deselectAll()
			}
		}
	}
	
	// Don't rebuild viewport every update - only when needed
	return nil
}

// Render returns the rendered tool browser view
func (tb *ToolBrowserNew) Render() string {
	return tb.renderTUIStyle()
}

// renderTUIStyle uses proper TUI best practices with official Bubbles components
func (tb *ToolBrowserNew) renderTUIStyle() string {
	// Header (search bar and category info)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 1)
	
	searchDisplay := tb.searchInput.View()
	if !tb.searchActive {
		searchDisplay = tb.searchInput.Placeholder
	}
	
	header := headerStyle.Render(fmt.Sprintf(
		"🔍 %s | Category: %s | %d/%d tools",
		searchDisplay,
		tb.selectedCategory,
		len(tb.filteredTools),
		len(tb.tools),
	))
	
	// Footer (help)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	footer := footerStyle.Render(
		"[/] Search  [↑/↓] Navigate  [Space] Select  [Enter] Details  [c] Category  [i] Install",
	)
	
	// Content (list items with cursor highlighting)
	// Show appropriate content based on current state
	tb.updateContent()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tb.viewport.View(),
		footer,
	)
}

// updateContent shows appropriate content based on current data state
func (tb *ToolBrowserNew) updateContent() {
	if !tb.ready {
		return
	}
	
	// If no tools data loaded yet, show loading state
	if len(tb.tools) == 0 {
		tb.setLoadingState()
		return
	}
	
	// If tools data is loaded but filtered tools is empty, apply filters
	if len(tb.filteredTools) == 0 {
		tb.applyFilters()
	}
	
	// If still no filtered tools after applying filters, show empty state
	if len(tb.filteredTools) == 0 {
		tb.setEmptyState()
		return
	}
	
	// Show tools content
	tb.updateViewportContentTUI()
}

// setLoadingState shows loading message without any processing
func (tb *ToolBrowserNew) setLoadingState() {
	loadingContent := `Loading tools...

Discovering available development tools.
This happens in the background.`
	tb.viewport.SetContent(loadingContent)
}

// setEmptyState shows when no tools match current filters
func (tb *ToolBrowserNew) setEmptyState() {
	emptyContent := `No tools found.

Try changing the category filter or search terms.`
	tb.viewport.SetContent(emptyContent)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (tb *ToolBrowserNew) updateViewportContentTUI() {
	var lines []string
	
	for i, tool := range tb.filteredTools {
		line := tb.renderToolItem(tool, false) // Don't use selected parameter
		
		// Apply cursor highlighting here (TUI best practice)
		if i == tb.cursor {
			selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
			line = selectedStyle.Render(line)
		}
		
		lines = append(lines, line)
	}
	
	content := strings.Join(lines, "\n")
	tb.viewport.SetContent(content)
	
	// Sync viewport with cursor position (TUI best practice)
	tb.syncViewportWithCursor()
}

// SetData updates the tool browser data
// SetData updates the tool browser data
func (tb *ToolBrowserNew) SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord) {
	debugLog("ToolBrowserNew.SetData: received %d tools, %d installed", len(tools), len(installed))
	tb.tools = tools
	tb.installedTools = installed
	tb.applyFilters()
	debugLog("ToolBrowserNew.SetData: applyFilters completed, filteredTools=%d", len(tb.filteredTools))
	
	// CRITICAL FIX: Always update the display after setting new data
	// This ensures the viewport shows the latest tool status
	if tb.ready {
		tb.updateContent()
		debugLog("ToolBrowserNew.SetData: updateContent() called to refresh display")
	}
}

// LoadFullContent loads the complete viewport content (called when view becomes active)
func (tb *ToolBrowserNew) LoadFullContent() {
	// This method is called from the async task
	// It should ensure content is properly loaded and displayed
	debugLog("ToolBrowserNew.LoadFullContent: ready=%v, tools=%d, installed=%d", 
		tb.ready, len(tb.tools), len(tb.installedTools))
	if tb.ready {
		tb.updateContent()
		debugLog("ToolBrowserNew.LoadFullContent: updateContent() completed")
	} else {
		debugLog("ToolBrowserNew.LoadFullContent: skipped - not ready")
	}
}

// updateViewportContent is deprecated - using TUI best practices instead

func (tb *ToolBrowserNew) renderToolItem(tool orchestrator.ToolConfig, selected bool) string {
	// Status indicators
	var status string
	if _, installed := tb.installedTools[tool.Name]; installed {
		status = "✓"
	} else {
		status = "○"
	}
	
	// Selection indicator
	var selection string
	if tb.selectedTools[tool.Name] {
		selection = "▣"
	} else {
		selection = "▢"
	}
	
	// Build the item
	name := tool.Name
	if len(name) > 20 {
		name = name[:17] + "..."
	}
	
	category := fmt.Sprintf("[%s]", tool.Category)
	
	// Format with consistent spacing
	line := fmt.Sprintf("%s %s %-20s %s", selection, status, name, category)
	
	// Apply style for installed tools
	if _, installed := tb.installedTools[tool.Name]; installed {
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		return mutedStyle.Render(line)
	}
	
	return line
}

// togglePreview switches between full width and split view
func (tb *ToolBrowserNew) togglePreview() {
	tb.showPreview = !tb.showPreview
	// Preview functionality can be implemented in TUI style later
}

// Business logic methods
func (tb *ToolBrowserNew) applyFilters() {
	tb.filteredTools = []orchestrator.ToolConfig{}
	
	searchTerm := strings.ToLower(tb.searchInput.Value())
	
	for _, tool := range tb.tools {
		// Category filter
		if tb.selectedCategory != "All" && tool.Category != tb.selectedCategory {
			continue
		}
		
		// Search filter
		if searchTerm != "" {
			nameMatch := strings.Contains(strings.ToLower(tool.Name), searchTerm)
			descMatch := strings.Contains(strings.ToLower(tool.Description), searchTerm)
			langMatch := strings.Contains(strings.ToLower(tool.Language), searchTerm)
			
			if !nameMatch && !descMatch && !langMatch {
				continue
			}
		}
		
		tb.filteredTools = append(tb.filteredTools, tool)
	}
	
	// Reset cursor if needed
	if tb.cursor >= len(tb.filteredTools) {
		tb.cursor = max(0, len(tb.filteredTools)-1)
	}
	
	// Note: applyFilters() only updates the filtered data
	// Display refresh is handled by the caller (SetData, LoadFullContent, etc.)
}

func (tb *ToolBrowserNew) moveUp() {
	if tb.cursor > 0 {
		tb.cursor--
		// Use TUI best practice: update content and sync viewport
		if tb.ready {
			tb.updateViewportContentTUI()
		}
	}
}

func (tb *ToolBrowserNew) moveDown() {
	if tb.cursor < len(tb.filteredTools)-1 {
		tb.cursor++
		// Use TUI best practice: update content and sync viewport
		if tb.ready {
			tb.updateViewportContentTUI()
		}
	}
}

func (tb *ToolBrowserNew) toggleSelection() {
	if tb.cursor < len(tb.filteredTools) {
		tool := tb.filteredTools[tb.cursor]
		if tb.selectedTools[tool.Name] {
			delete(tb.selectedTools, tool.Name)
		} else {
			tb.selectedTools[tool.Name] = true
		}
		// Need to rebuild to show selection changes
		if tb.ready {
			tb.updateViewportContentTUI()
		}
	}
}

func (tb *ToolBrowserNew) cycleCategory() {
	currentIdx := 0
	for i, cat := range tb.categories {
		if cat == tb.selectedCategory {
			currentIdx = i
			break
		}
	}
	
	nextIdx := (currentIdx + 1) % len(tb.categories)
	tb.selectedCategory = tb.categories[nextIdx]
	tb.applyFilters()
}

func (tb *ToolBrowserNew) selectAll() {
	for _, tool := range tb.filteredTools {
		tb.selectedTools[tool.Name] = true
	}
	if tb.ready {
		tb.updateViewportContentTUI()
	}
}

func (tb *ToolBrowserNew) deselectAll() {
	tb.selectedTools = make(map[string]bool)
	if tb.ready {
		tb.updateViewportContentTUI()
	}
}

// GetSelectedTools returns the list of selected tool names
func (tb *ToolBrowserNew) GetSelectedTools() []string {
	var selected []string
	for name := range tb.selectedTools {
		selected = append(selected, name)
	}
	return selected
}

// ClearSelection clears all selected tools
func (tb *ToolBrowserNew) ClearSelection() {
	tb.selectedTools = make(map[string]bool)
	if tb.ready {
		tb.updateViewportContentTUI()
	}
}

