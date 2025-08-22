package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"gearbox/cmd/gearbox/tui/styles"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// BundleExplorer represents the bundle explorer view
type BundleExplorer struct {
	width  int
	height int
	
	// Data
	bundles        []orchestrator.BundleConfig
	installedTools map[string]*manifest.InstallationRecord
	
	// UI state
	cursor          int      // Index in selectable bundles
	scrollOffset    int      // Line offset in the rendered content
	expandedBundles map[string]bool
	selectedCategory string
	
	// Categories
	categories []string
	
	// Cached rendering data
	renderedLines    []string // All rendered lines
	selectableLines  []int    // Line indices that correspond to selectable bundles
	lastHeight      int      // Track last render height for moveUp/Down
}

// NewBundleExplorer creates a new bundle explorer view
func NewBundleExplorer() *BundleExplorer {
	return &BundleExplorer{
		expandedBundles: make(map[string]bool),
		installedTools:  make(map[string]*manifest.InstallationRecord),
		categories:      []string{"foundation", "domain", "language", "workflow", "infrastructure"},
	}
}

// SetSize updates the size of the bundle explorer
func (be *BundleExplorer) SetSize(width, height int) {
	be.width = width
	be.height = height
}

// SetData updates the bundle explorer data
func (be *BundleExplorer) SetData(bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	be.bundles = bundles
	be.installedTools = installed
}

// Update handles bundle explorer updates
func (be *BundleExplorer) Update(msg tea.Msg) tea.Cmd {
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
	
	return nil
}

// Render returns the rendered bundle explorer view
func (be *BundleExplorer) Render() string {
	if be.width == 0 || be.height == 0 {
		return "Loading..."
	}
	
	// Render components first to know their exact heights
	title := styles.TitleStyle().Render("Bundle Explorer")
	categoryBar := be.renderCategoryBar()
	helpBar := be.renderHelpBar()
	
	// Calculate exact heights
	titleHeight := lipgloss.Height(title)
	categoryHeight := lipgloss.Height(categoryBar) 
	helpHeight := lipgloss.Height(helpBar)
	
	// Calculate space for bundle tree
	// We need to ensure the total doesn't exceed be.height
	fixedComponentsHeight := titleHeight + categoryHeight + helpHeight
	availableForTree := be.height - fixedComponentsHeight
	
	// Ensure we have at least some space for the tree
	if availableForTree < 5 {
		// Not enough space - we need to make room
		availableForTree = 5
	}
	
	// Render the tree with the available space
	bundleTree := be.renderBundleTree(availableForTree)
	
	// Now assemble all components
	components := []string{title, categoryBar, bundleTree, helpBar}
	result := lipgloss.JoinVertical(lipgloss.Left, components...)
	
	// If the result is too tall, we need to trim it
	resultHeight := lipgloss.Height(result)
	if resultHeight > be.height {
		// The tree must have rendered taller than expected
		// Reduce tree height and try again
		availableForTree = availableForTree - (resultHeight - be.height)
		if availableForTree < 1 {
			availableForTree = 1
		}
		bundleTree = be.renderBundleTree(availableForTree)
		components = []string{title, categoryBar, bundleTree, helpBar}
		result = lipgloss.JoinVertical(lipgloss.Left, components...)
	}
	
	return result
}

func (be *BundleExplorer) renderCategoryBar() string {
	var categories []string
	for _, cat := range be.categories {
		if cat == be.selectedCategory || be.selectedCategory == "" {
			categories = append(categories, styles.HighlightStyle().Render("["+cat+"]"))
		} else {
			categories = append(categories, styles.MutedStyle().Render(cat))
		}
	}
	
	return lipgloss.NewStyle().
		Width(be.width).
		Padding(0, 2).
		Render(strings.Join(categories, "  "))
}

func (be *BundleExplorer) renderBundleTree(requestedHeight int) string {
	// Store height for navigation  
	be.lastHeight = requestedHeight
	
	// Rebuild the rendered content and track selectable lines
	be.rebuildRenderedContent()
	
	// IMPORTANT: BoxStyle().Height(n) sets the CONTENT height to n
	// The actual rendered height will be n + borders(2) + padding(2) = n + 4
	// So if we want the total rendered height to be requestedHeight,
	// we need to set the content height to requestedHeight - 4
	contentHeight := requestedHeight - 4
	if contentHeight < 1 {
		contentHeight = 1
	}
	
	// If there's nothing to render
	if len(be.renderedLines) == 0 {
		return styles.BoxStyle().
			Width(be.width).
			Height(contentHeight).
			Render("No bundles available")
	}
	
	// Trim trailing empty lines to get actual content length
	actualContentLines := len(be.renderedLines)
	for actualContentLines > 0 && be.renderedLines[actualContentLines-1] == "" {
		actualContentLines--
	}
	
	// Check if we need scrolling
	needsScroll := actualContentLines > contentHeight
	
	// Find the current selection line
	selectedLine := -1
	if be.cursor >= 0 && be.cursor < len(be.selectableLines) {
		selectedLine = be.selectableLines[be.cursor]
	}
	
	// Calculate viewport for scrolling
	viewportHeight := contentHeight
	if needsScroll {
		// Reserve lines for indicators within the content area
		viewportHeight = contentHeight - 2
		if viewportHeight < 1 {
			viewportHeight = 1
		}
	}
	
	// Update scroll offset to keep selection visible
	if selectedLine >= 0 && needsScroll {
		if selectedLine < be.scrollOffset {
			be.scrollOffset = selectedLine
		} else if selectedLine >= be.scrollOffset + viewportHeight {
			be.scrollOffset = selectedLine - viewportHeight + 1
		}
		
		// Clamp scroll offset
		maxScroll := max(0, actualContentLines - viewportHeight)
		be.scrollOffset = max(0, min(be.scrollOffset, maxScroll))
	} else if !needsScroll {
		be.scrollOffset = 0
	}
	
	// Build the display content - always exactly contentHeight lines
	var displayLines []string
	
	if needsScroll {
		// Add top indicator
		if be.scrollOffset > 0 {
			displayLines = append(displayLines, styles.MutedStyle().Render("↑ More above"))
		} else {
			displayLines = append(displayLines, "")
		}
		
		// Add viewport content
		for i := 0; i < viewportHeight; i++ {
			lineIdx := be.scrollOffset + i
			if lineIdx < len(be.renderedLines) {
				displayLines = append(displayLines, be.renderedLines[lineIdx])
			} else {
				displayLines = append(displayLines, "")
			}
		}
		
		// Add bottom indicator
		if be.scrollOffset + viewportHeight < actualContentLines {
			displayLines = append(displayLines, styles.MutedStyle().Render("↓ More below"))
		} else {
			displayLines = append(displayLines, "")
		}
	} else {
		// No scrolling - just show all content with padding
		displayLines = append(displayLines, be.renderedLines...)
		// Pad to exact height
		for len(displayLines) < contentHeight {
			displayLines = append(displayLines, "")
		}
		// Truncate if too long
		if len(displayLines) > contentHeight {
			displayLines = displayLines[:contentHeight]
		}
	}
	
	// Join and render
	content := strings.Join(displayLines, "\n")
	
	return styles.BoxStyle().
		Width(be.width).
		Height(contentHeight).
		Render(content)
}

func (be *BundleExplorer) renderBundleLine(bundle orchestrator.BundleConfig, selected bool) string {
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
	
	// Apply style
	if selected {
		return styles.SelectedStyle().Render(line)
	}
	
	if installedCount == len(bundle.Tools) && len(bundle.Tools) > 0 {
		return styles.SuccessStyle().Render(line)
	}
	
	return line
}

func (be *BundleExplorer) renderBundleDetails(bundle orchestrator.BundleConfig) []string {
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
		// No indentation - button's border provides visual separation
		buttonText := "[i] Install Bundle"
		button := styles.ButtonStyle(false).Render(buttonText)
		details = append(details, button)
	}
	
	details = append(details, "") // Empty line after details
	
	return details
}

func (be *BundleExplorer) renderHelpBar() string {
	helps := []string{
		"[↑/↓] Navigate",
		"[Enter/Space] Expand",
		"[c] Cycle Category",
		"[i] Install Bundle",
		"[Tab] Switch View",
	}
	
	return styles.MutedStyle().Render(strings.Join(helps, "  "))
}

// Helper methods

// rebuildRenderedContent rebuilds the complete rendered output and tracks selectable lines
func (be *BundleExplorer) rebuildRenderedContent() {
	be.renderedLines = nil
	be.selectableLines = nil
	
	bundlesByCategory := be.groupBundlesByCategory()
	selectableBundles := be.getSelectableBundles()
	
	// Create a map for quick lookup of bundle indices
	bundleIndexMap := make(map[string]int)
	for i, bundle := range selectableBundles {
		bundleIndexMap[bundle.Name] = i
	}
	
	lineIndex := 0
	
	for _, category := range be.categories {
		if be.selectedCategory != "" && be.selectedCategory != category {
			continue
		}
		
		bundles := bundlesByCategory[category]
		if len(bundles) == 0 {
			continue
		}
		
		// Category header
		categoryTitle := styles.SubtitleStyle().Render(strings.Title(category) + " Tier")
		be.renderedLines = append(be.renderedLines, categoryTitle)
		lineIndex++
		
		// Bundles in this category
		for _, bundle := range bundles {
			// Check if this bundle is selectable and if it's the current selection
			bundleIdx, isSelectable := bundleIndexMap[bundle.Name]
			isCurrentBundle := isSelectable && bundleIdx == be.cursor
			
			// Track this line as selectable
			if isSelectable {
				be.selectableLines = append(be.selectableLines, lineIndex)
			}
			
			bundleLine := be.renderBundleLine(bundle, isCurrentBundle)
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

func (be *BundleExplorer) groupBundlesByCategory() map[string][]orchestrator.BundleConfig {
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

func (be *BundleExplorer) moveUp() {
	bundles := be.getSelectableBundles()
	if be.cursor > 0 && len(bundles) > 0 {
		be.cursor--
	}
}

func (be *BundleExplorer) moveDown() {
	bundles := be.getSelectableBundles()
	if be.cursor < len(bundles)-1 {
		be.cursor++
	}
}

// getSelectableBundles returns only the bundles that can be selected (filtered by category)
func (be *BundleExplorer) getSelectableBundles() []orchestrator.BundleConfig {
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

func (be *BundleExplorer) countTotalLines() int {
	count := 0
	bundlesByCategory := be.groupBundlesByCategory()
	
	for _, category := range be.categories {
		if be.selectedCategory != "" && be.selectedCategory != category {
			continue
		}
		
		bundles := bundlesByCategory[category]
		if len(bundles) == 0 {
			continue
		}
		
		count++ // Category header
		
		for _, bundle := range bundles {
			count++ // Bundle line
			if be.expandedBundles[bundle.Name] {
				// Count expanded details
				count += be.countBundleDetailLines(bundle)
			}
		}
		
		count++ // Empty line between categories
	}
	
	return count
}

func (be *BundleExplorer) countBundleDetailLines(bundle orchestrator.BundleConfig) int {
	count := 0
	
	if len(bundle.Tools) > 0 {
		count++ // "Tools:" header
		count += len(bundle.Tools)
	}
	
	if len(bundle.IncludesBundles) > 0 {
		count++ // "Includes bundles:" header
		count += len(bundle.IncludesBundles)
	}
	
	if len(bundle.SystemPackages) > 0 {
		count++ // "System packages:" header
		count += len(bundle.SystemPackages)
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
		count += 2 // Empty line + button
	}
	
	count++ // Empty line after details
	
	return count
}

func (be *BundleExplorer) toggleExpanded() {
	// Get the selected bundle
	bundles := be.getSelectableBundles()
	if be.cursor >= 0 && be.cursor < len(bundles) {
		bundle := bundles[be.cursor]
		be.expandedBundles[bundle.Name] = !be.expandedBundles[bundle.Name]
	}
}

func (be *BundleExplorer) cycleCategory() {
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
	be.cursor = 0
	be.scrollOffset = 0
}

// GetSelectedBundle returns the currently selected bundle
func (be *BundleExplorer) GetSelectedBundle() *orchestrator.BundleConfig {
	bundles := be.getSelectableBundles()
	if be.cursor >= 0 && be.cursor < len(bundles) {
		return &bundles[be.cursor]
	}
	return nil
}

// GetUninstalledTools returns the list of uninstalled tools from the selected bundle
func (be *BundleExplorer) GetUninstalledTools(bundle *orchestrator.BundleConfig) []string {
	var uninstalled []string
	
	for _, toolName := range bundle.Tools {
		if _, installed := be.installedTools[toolName]; !installed {
			uninstalled = append(uninstalled, toolName)
		}
	}
	
	return uninstalled
}