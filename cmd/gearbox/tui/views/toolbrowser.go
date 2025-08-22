package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"gearbox/cmd/gearbox/tui/styles"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// ToolBrowser represents the tool browser view
type ToolBrowser struct {
	width  int
	height int
	
	// Data
	tools          []orchestrator.ToolConfig
	filteredTools  []orchestrator.ToolConfig
	installedTools map[string]*manifest.InstallationRecord
	
	// UI state
	searchInput    textinput.Model
	cursor         int
	scrollOffset   int
	selectedTools  map[string]bool
	searchActive   bool
	showPreview    bool
	
	// Categories
	categories     []string
	selectedCategory string
}

// NewToolBrowser creates a new tool browser view
func NewToolBrowser() *ToolBrowser {
	ti := textinput.New()
	ti.Placeholder = "Search tools..."
	ti.CharLimit = 50
	ti.Width = 40
	
	return &ToolBrowser{
		searchInput:   ti,
		selectedTools: make(map[string]bool),
		installedTools: make(map[string]*manifest.InstallationRecord),
		showPreview:   true,
		categories:    []string{"All", "Core", "Development", "System", "Text", "Media", "UI"},
		selectedCategory: "All",
		filteredTools: []orchestrator.ToolConfig{}, // Initialize empty slice
	}
}

// SetSize updates the size of the tool browser
func (tb *ToolBrowser) SetSize(width, height int) {
	tb.width = width
	tb.height = height
	tb.searchInput.Width = min(width-10, 50)
}

// SetData updates the tool browser data
func (tb *ToolBrowser) SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord) {
	tb.tools = tools
	tb.installedTools = installed
	// Initialize selectedCategory if not set
	if tb.selectedCategory == "" {
		tb.selectedCategory = "All"
	}
	tb.applyFilters()
}

// Update handles tool browser updates
func (tb *ToolBrowser) Update(msg tea.Msg) tea.Cmd {
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
				tb.showPreview = !tb.showPreview
			case "a":
				tb.selectAll()
			case "A":
				tb.deselectAll()
			}
		}
	}
	
	return nil
}

// Render returns the rendered tool browser view
func (tb *ToolBrowser) Render() string {
	if tb.width == 0 || tb.height == 0 {
		return "Loading..."
	}
	
	// Calculate layout
	searchBarHeight := 3
	helpBarHeight := 2 // Increased to account for proper spacing
	contentHeight := tb.height - searchBarHeight - helpBarHeight
	
	// Render components
	searchBar := tb.renderSearchBar()
	
	var content string
	if tb.showPreview {
		// Split view with tool list and preview
		listWidth := tb.width / 2
		previewWidth := tb.width - listWidth - 2
		
		toolList := tb.renderToolList(listWidth, contentHeight)
		preview := tb.renderPreview(previewWidth, contentHeight)
		
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			toolList,
			lipgloss.NewStyle().Width(2).Render("  "),
			preview,
		)
	} else {
		// Full width tool list
		content = tb.renderToolList(tb.width-4, contentHeight)
	}
	
	helpBar := tb.renderHelpBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		searchBar,
		content,
		helpBar,
	)
}

func (tb *ToolBrowser) renderSearchBar() string {
	categorySelector := fmt.Sprintf("Category: [%s]", tb.selectedCategory)
	
	searchBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.CurrentTheme.Border).
		Padding(0, 1).
		Width(tb.searchInput.Width + 4).
		Render(tb.searchInput.View())
	
	stats := fmt.Sprintf("Showing %d/%d tools", len(tb.filteredTools), len(tb.tools))
	
	leftContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		searchBox,
		lipgloss.NewStyle().Width(4).Render("  "),
		categorySelector,
	)
	
	gap := tb.width - lipgloss.Width(leftContent) - lipgloss.Width(stats) - 4
	if gap < 0 {
		gap = 0
	}
	
	return lipgloss.NewStyle().
		Width(tb.width).
		Padding(0, 2).
		Render(
			leftContent +
			lipgloss.NewStyle().Width(gap).Render(" ") +
			stats,
		)
}

func (tb *ToolBrowser) renderToolList(width, height int) string {
	if len(tb.filteredTools) == 0 {
		return styles.BoxStyle().
			Width(width).
			Height(height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("No tools found")
	}
	
	// Calculate visible items
	visibleItems := height - 2
	if tb.cursor >= tb.scrollOffset+visibleItems {
		tb.scrollOffset = tb.cursor - visibleItems + 1
	} else if tb.cursor < tb.scrollOffset {
		tb.scrollOffset = tb.cursor
	}
	
	var items []string
	for i := tb.scrollOffset; i < min(tb.scrollOffset+visibleItems, len(tb.filteredTools)); i++ {
		tool := tb.filteredTools[i]
		items = append(items, tb.renderToolItem(tool, i == tb.cursor, width-4))
	}
	
	content := strings.Join(items, "\n")
	
	// Add scroll indicators
	if tb.scrollOffset > 0 {
		content = "↑ More above\n" + content
	}
	if tb.scrollOffset+visibleItems < len(tb.filteredTools) {
		content = content + "\n↓ More below"
	}
	
	return styles.BoxStyle().
		Width(width).
		Height(height).
		Render(content)
}

func (tb *ToolBrowser) renderToolItem(tool orchestrator.ToolConfig, selected bool, width int) string {
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
	
	// Calculate spacing
	baseContent := fmt.Sprintf("%s %s %s", selection, status, name)
	padding := width - lipgloss.Width(baseContent) - lipgloss.Width(category) - 2
	if padding < 0 {
		padding = 0
	}
	
	item := fmt.Sprintf("%s %s %s%s %s", selection, status, name, strings.Repeat(" ", padding), category)
	
	// Apply style
	if selected {
		return styles.SelectedStyle().Render(item)
	}
	
	if _, installed := tb.installedTools[tool.Name]; installed {
		return styles.MutedStyle().Render(item)
	}
	
	return item
}

func (tb *ToolBrowser) renderPreview(width, height int) string {
	if tb.cursor >= len(tb.filteredTools) {
		return styles.BoxStyle().
			Width(width).
			Height(height).
			Render("Select a tool to preview")
	}
	
	tool := tb.filteredTools[tb.cursor]
	
	// Title
	title := styles.TitleStyle().Render(tool.Name)
	
	// Description
	desc := lipgloss.NewStyle().
		Width(width-4).
		Render(tool.Description)
	
	// Installation status
	var status string
	if _, installed := tb.installedTools[tool.Name]; installed {
		status = styles.SuccessStyle().Render("✓ Installed")
	} else {
		status = styles.WarningStyle().Render("○ Not installed")
	}
	
	// Build types
	buildTypes := "Build Types:\n"
	for bt, flags := range tool.BuildTypes {
		buildTypes += fmt.Sprintf("  • %s: %s\n", bt, flags)
	}
	
	// Dependencies
	deps := "Dependencies:\n"
	if len(tool.Dependencies) > 0 {
		for _, dep := range tool.Dependencies {
			deps += fmt.Sprintf("  • %s\n", dep)
		}
	} else {
		deps += "  • None\n"
	}
	
	// Language and repository
	info := fmt.Sprintf("Language: %s\nRepository: %s\nBinary: %s\n",
		tool.Language, tool.Repository, tool.BinaryName)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		desc,
		"",
		status,
		"",
		buildTypes,
		deps,
		info,
	)
	
	return styles.BoxStyle().
		Width(width).
		Height(height).
		Render(content)
}

func (tb *ToolBrowser) renderHelpBar() string {
	helps := []string{
		"[/] Search",
		"[Space] Select",
		"[Enter] Details",
		"[c] Category",
		"[p] Toggle Preview",
		"[i] Install Selected",
	}
	
	return styles.MutedStyle().Render(strings.Join(helps, "  "))
}

// Helper methods

func (tb *ToolBrowser) applyFilters() {
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
}

func (tb *ToolBrowser) moveUp() {
	if tb.cursor > 0 {
		tb.cursor--
	}
}

func (tb *ToolBrowser) moveDown() {
	if tb.cursor < len(tb.filteredTools)-1 {
		tb.cursor++
	}
}

func (tb *ToolBrowser) toggleSelection() {
	if tb.cursor < len(tb.filteredTools) {
		tool := tb.filteredTools[tb.cursor]
		if tb.selectedTools[tool.Name] {
			delete(tb.selectedTools, tool.Name)
		} else {
			tb.selectedTools[tool.Name] = true
		}
	}
}

func (tb *ToolBrowser) cycleCategory() {
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

func (tb *ToolBrowser) selectAll() {
	for _, tool := range tb.filteredTools {
		tb.selectedTools[tool.Name] = true
	}
}

func (tb *ToolBrowser) deselectAll() {
	tb.selectedTools = make(map[string]bool)
}

// GetSelectedTools returns the list of selected tool names
func (tb *ToolBrowser) GetSelectedTools() []string {
	var selected []string
	for name := range tb.selectedTools {
		selected = append(selected, name)
	}
	return selected
}

// ClearSelection clears all selected tools
func (tb *ToolBrowser) ClearSelection() {
	tb.selectedTools = make(map[string]bool)
}

