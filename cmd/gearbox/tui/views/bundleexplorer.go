package views

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// BundleExplorerNew demonstrates the new layout system
type BundleExplorerNew struct {
	// Data
	bundles        []orchestrator.BundleConfig
	installedTools map[string]*manifest.InstallationRecord
	
	// UI state
	cursor          int
	selectedBundle  string // Track selected bundle by name instead of index
	expandedBundles map[string]bool
	selectedCategory string
	categories      []string
	
	// TUI components (official Bubbles components)
	viewport viewport.Model
	ready    bool
	width    int
	height   int
	
	// Cached content
	renderedLines   []string
	bundleLineMap   map[string]int // Maps bundle name to its line index for robust highlighting
}

// NewBundleExplorerNew creates a new bundle explorer with TUI best practices
func NewBundleExplorerNew() *BundleExplorerNew {
	return &BundleExplorerNew{
		expandedBundles: make(map[string]bool),
		installedTools:  make(map[string]*manifest.InstallationRecord),
		categories:      []string{"foundation", "domain", "language", "workflow", "infrastructure"},
		cursor:          0,
		selectedBundle:  "", // Will be set when data is loaded
	}
}

// SetSize updates the bundle explorer size (TUI best practices)
func (be *BundleExplorerNew) SetSize(width, height int) {
	be.width = width
	be.height = height
	
	// Initialize official viewport if not ready
	if !be.ready {
		// Calculate viewport height: total - header - footer
		viewportHeight := height - 3 // Reserve 3 lines for header + footer
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		
		be.viewport = viewport.New(width, viewportHeight)
		be.viewport.SetContent("")
		be.ready = true
	} else {
		// Update existing viewport
		viewportHeight := height - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		be.viewport.Width = width
		be.viewport.Height = viewportHeight
	}
}

// Update handles input - business logic only
func (be *BundleExplorerNew) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			be.moveUp()
		case "down", "j":
			be.moveDown()
		case "enter", " ":
			be.toggleExpanded()
		case "c":
			be.cycleCategory()
		}
	}
	
	// Don't rebuild viewport content every update - only when data actually changes
	return nil
}

// Render is now trivial - just ask layout to render
func (be *BundleExplorerNew) Render() string {
	return be.renderTUIStyle()
}

// renderTUIStyle uses proper TUI best practices with official Bubbles components
func (be *BundleExplorerNew) renderTUIStyle() string {
	// Header (category filter and bundle info)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 1)
	
	header := headerStyle.Render(fmt.Sprintf(
		"Bundle Explorer | Category: %s | %d bundles",
		be.selectedCategory,
		len(be.bundles),
	))
	
	// Footer (help)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	footer := footerStyle.Render(
		"[↑/↓] Navigate  [Enter/Space] Expand  [c] Cycle Category  [i] Install Bundle  [Tab] Switch View",
	)
	
	// Content (bundle list with cursor highlighting)
	be.updateViewportContentTUI()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		be.viewport.View(),
		footer,
	)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (be *BundleExplorerNew) updateViewportContentTUI() {
	be.rebuildRenderedContent()
	
	// Use robust line index mapping instead of fragile string matching
	selectedLineIndex := -1
	if be.selectedBundle != "" {
		if lineIdx, exists := be.bundleLineMap[be.selectedBundle]; exists {
			selectedLineIndex = lineIdx
		}
	}
	
	// Apply highlighting to the correct line
	var lines []string
	for i, line := range be.renderedLines {
		if i == selectedLineIndex {
			// Highlight the selected bundle line
			selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
			line = selectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	
	content := strings.Join(lines, "\n")
	be.viewport.SetContent(content)
	
	// DON'T interfere with manual scrolling - let viewport handle scrolling naturally
	// Only auto-scroll when selection changes, not on every content update
}

// syncViewportWithLine ensures selected line is visible (TUI best practice)
func (be *BundleExplorerNew) syncViewportWithLine(lineIndex int) {
	if lineIndex < 0 || len(be.renderedLines) == 0 {
		return
	}
	
	// Get viewport bounds
	top := be.viewport.YOffset
	bottom := top + be.viewport.Height - 1
	
	// Ensure selected line is visible by scrolling viewport
	if lineIndex < top {
		// Line above viewport - scroll up
		be.viewport.SetYOffset(lineIndex)
	} else if lineIndex > bottom {
		// Line below viewport - scroll down
		be.viewport.SetYOffset(lineIndex - be.viewport.Height + 1)
	}
}

// SetData updates data and refreshes content
func (be *BundleExplorerNew) SetData(bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	be.bundles = bundles
	be.installedTools = installed
	
	// Initialize selectedBundle to first available bundle if none selected
	if be.selectedBundle == "" && len(bundles) > 0 {
		selectableBundles := be.getSelectableBundles()
		if len(selectableBundles) > 0 {
			be.selectedBundle = selectableBundles[0].Name
		}
	}
	
	if be.ready {
		be.updateViewportContentTUI()
	}
}

// updateViewportContent is deprecated - using TUI best practices instead

// Business logic methods - robust approach with line index mapping
func (be *BundleExplorerNew) rebuildRenderedContent() {
	be.renderedLines = nil
	be.bundleLineMap = make(map[string]int)
	
	lineIndex := 0
	bundlesByCategory := be.groupBundlesByCategory()
	
	for _, category := range be.categories {
		if be.selectedCategory != "" && be.selectedCategory != category {
			continue
		}
		
		bundles := bundlesByCategory[category]
		if len(bundles) == 0 {
			continue
		}
		
		// Category header
		subtitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
		categoryTitle := subtitleStyle.Render(strings.Title(category) + " Tier")
		be.renderedLines = append(be.renderedLines, categoryTitle)
		lineIndex++
		
		// Bundles in this category
		for _, bundle := range bundles {
			// Record the exact line index for this bundle (BEFORE adding the line)
			be.bundleLineMap[bundle.Name] = lineIndex
			
			// Always render without highlighting - highlighting will be applied in updateViewportContentTUI
			bundleLine := be.renderBundleLine(bundle, false)
			be.renderedLines = append(be.renderedLines, bundleLine)
			lineIndex++
			
			// Show expanded details
			if be.expandedBundles[bundle.Name] {
				details := be.renderBundleDetails(bundle)
				be.renderedLines = append(be.renderedLines, details...)
				lineIndex += len(details)
			}
		}
		
		// Empty line between categories
		be.renderedLines = append(be.renderedLines, "")
		lineIndex++
	}
}

// getSelectedLineIndex function removed - no longer needed with name-based selection

func (be *BundleExplorerNew) groupBundlesByCategory() map[string][]orchestrator.BundleConfig {
	grouped := make(map[string][]orchestrator.BundleConfig)
	
	for _, bundle := range be.bundles {
		category := bundle.Category
		if category == "" {
			// Infer category from bundle name or tags
			if strings.HasSuffix(bundle.Name, "-dev") {
				if strings.Contains(bundle.Name, "python") || 
				   strings.Contains(bundle.Name, "nodejs") ||
				   strings.Contains(bundle.Name, "rust") ||
				   strings.Contains(bundle.Name, "go") {
					category = "language"
				} else {
					category = "domain"
				}
			} else if strings.Contains(bundle.Name, "beginner") ||
			          strings.Contains(bundle.Name, "intermediate") ||
			          strings.Contains(bundle.Name, "advanced") {
				category = "foundation"
			} else if strings.Contains(bundle.Name, "docker") ||
			          strings.Contains(bundle.Name, "cloud") ||
			          strings.Contains(bundle.Name, "database") {
				category = "infrastructure"
			} else if strings.Contains(bundle.Name, "debugging") ||
			          strings.Contains(bundle.Name, "deployment") ||
			          strings.Contains(bundle.Name, "review") {
				category = "workflow"
			}
		}
		
		if category != "" {
			grouped[category] = append(grouped[category], bundle)
		}
	}
	
	return grouped
}

func (be *BundleExplorerNew) getSelectableBundles() []orchestrator.BundleConfig {
	var selectableBundles []orchestrator.BundleConfig
	bundlesByCategory := be.groupBundlesByCategory()
	
	for _, category := range be.categories {
		if be.selectedCategory != "" && be.selectedCategory != category {
			continue
		}
		
		bundles := bundlesByCategory[category]
		selectableBundles = append(selectableBundles, bundles...)
	}
	
	return selectableBundles
}

func (be *BundleExplorerNew) renderBundleLine(bundle orchestrator.BundleConfig, selected bool) string {
	// Expansion indicator
	var expandIcon string
	if be.expandedBundles[bundle.Name] {
		expandIcon = "▼"
	} else {
		expandIcon = "▶"
	}
	
	// Installation status
	installedCount := 0
	for _, toolName := range bundle.Tools {
		if _, ok := be.installedTools[toolName]; ok {
			installedCount++
		}
	}
	
	status := fmt.Sprintf("(%d/%d tools)", installedCount, len(bundle.Tools))
	
	// Build line
	line := fmt.Sprintf("%s %s %s - %s",
		expandIcon,
		bundle.Name,
		status,
		bundle.Description,
	)
	
	// Apply style (will be handled by viewport if selected)
	if installedCount == len(bundle.Tools) && len(bundle.Tools) > 0 {
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		return successStyle.Render(line)
	}
	
	return line
}

func (be *BundleExplorerNew) renderBundleDetails(bundle orchestrator.BundleConfig) []string {
	var details []string
	indent := "    "
	
	// Tools section
	if len(bundle.Tools) > 0 {
		details = append(details, indent+"Tools:")
		for _, toolName := range bundle.Tools {
			var status string
			if _, installed := be.installedTools[toolName]; installed {
				status = "✓"
			} else {
				status = "○"
			}
			details = append(details, indent+"  "+status+" "+toolName)
		}
	}
	
	// Dependencies section
	if len(bundle.IncludesBundles) > 0 {
		details = append(details, indent+"Includes bundles:")
		for _, dep := range bundle.IncludesBundles {
			details = append(details, indent+"  • "+dep)
		}
	}
	
	// System packages section
	if len(bundle.SystemPackages) > 0 {
		details = append(details, indent+"System packages:")
		for _, pkg := range bundle.SystemPackages {
			details = append(details, indent+"  • "+pkg)
		}
	}
	
	// Action button
	allInstalled := true
	for _, toolName := range bundle.Tools {
		if _, ok := be.installedTools[toolName]; !ok {
			allInstalled = false
			break
		}
	}
	
	if !allInstalled {
		details = append(details, "")
		buttonStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(0, 1)
		buttonText := "[i] Install Bundle"
		button := buttonStyle.Render(buttonText)
		details = append(details, button)
	}
	
	details = append(details, "") // Empty line after details
	
	return details
}

func (be *BundleExplorerNew) moveUp() {
	selectableBundles := be.getSelectableBundles()
	if len(selectableBundles) == 0 {
		return
	}
	
	// Find current bundle index
	currentIndex := -1
	for i, bundle := range selectableBundles {
		if bundle.Name == be.selectedBundle {
			currentIndex = i
			break
		}
	}
	
	// Move to previous bundle
	if currentIndex > 0 {
		be.selectedBundle = selectableBundles[currentIndex-1].Name
		be.cursor = currentIndex - 1
		
		if be.ready {
			be.updateViewportContentTUI()
		}
	}
}

func (be *BundleExplorerNew) moveDown() {
	selectableBundles := be.getSelectableBundles()
	if len(selectableBundles) == 0 {
		return
	}
	
	// Find current bundle index
	currentIndex := -1
	for i, bundle := range selectableBundles {
		if bundle.Name == be.selectedBundle {
			currentIndex = i
			break
		}
	}
	
	// Move to next bundle
	if currentIndex >= 0 && currentIndex < len(selectableBundles)-1 {
		be.selectedBundle = selectableBundles[currentIndex+1].Name
		be.cursor = currentIndex + 1
		
		if be.ready {
			be.updateViewportContentTUI()
		}
	}
}

func (be *BundleExplorerNew) toggleExpanded() {
	if be.selectedBundle != "" {
		be.expandedBundles[be.selectedBundle] = !be.expandedBundles[be.selectedBundle]
		// Need to rebuild content when expanding/collapsing
		if be.ready {
			be.updateViewportContentTUI()
		}
	}
}

func (be *BundleExplorerNew) cycleCategory() {
	if be.selectedCategory == "" {
		be.selectedCategory = be.categories[0]
	} else {
		for i, cat := range be.categories {
			if cat == be.selectedCategory {
				if i < len(be.categories)-1 {
					be.selectedCategory = be.categories[i+1]
				} else {
					be.selectedCategory = ""
				}
				break
			}
		}
	}
	
	// Reset to first bundle in new category
	be.cursor = 0
	selectableBundles := be.getSelectableBundles()
	if len(selectableBundles) > 0 {
		be.selectedBundle = selectableBundles[0].Name
	} else {
		be.selectedBundle = ""
	}
	
	// Need to rebuild content when changing category
	if be.ready {
		be.updateViewportContentTUI()
	}
}

// GetSelectedBundle returns the currently selected bundle
func (be *BundleExplorerNew) GetSelectedBundle() *orchestrator.BundleConfig {
	if be.selectedBundle == "" {
		return nil
	}
	
	bundles := be.getSelectableBundles()
	for _, bundle := range bundles {
		if bundle.Name == be.selectedBundle {
			return &bundle
		}
	}
	return nil
}

// GetUninstalledTools returns the list of uninstalled tools from the selected bundle
func (be *BundleExplorerNew) GetUninstalledTools(bundle *orchestrator.BundleConfig) []string {
	var uninstalled []string
	
	for _, toolName := range bundle.Tools {
		if _, installed := be.installedTools[toolName]; !installed {
			uninstalled = append(uninstalled, toolName)
		}
	}
	
	return uninstalled
}

// CategoryBarWidget removed - using TUI best practices with header integration