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
	
	// Title
	title := styles.TitleStyle().Render("Bundle Explorer")
	
	// Category filter
	categoryBar := be.renderCategoryBar()
	
	// Bundle tree
	// Title = 2 (with margin), Category bar = 2, Help bar = 1, spacing = 1, total = 6
	// But let's be more conservative to ensure content fits
	contentHeight := be.height - 7
	bundleTree := be.renderBundleTree(contentHeight)
	
	// Help bar
	helpBar := be.renderHelpBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		categoryBar,
		bundleTree,
		helpBar,
	)
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

func (be *BundleExplorer) renderBundleTree(height int) string {
	// Store height for navigation
	be.lastHeight = height
	
	// Rebuild the rendered content and track selectable lines
	be.rebuildRenderedContent()
	
	// Calculate viewport
	// BoxStyle adds: 2 lines for borders + 2 lines for padding = 4 total
	contentHeight := height - 4 // Account for box borders and padding
	
	// If there's nothing to render
	if len(be.renderedLines) == 0 {
		return styles.BoxStyle().
			Width(be.width).
			Height(height).
			Render("No bundles available")
	}
	
	// Calculate total content and viewport
	totalLines := len(be.renderedLines)
	
	// Use a fixed viewport that doesn't change
	// Always reserve space for potential scroll indicators
	viewportHeight := contentHeight - 2
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	
	// Determine if scrolling is needed
	needsScroll := totalLines > viewportHeight
	
	// Find the line of the current selection
	selectedLine := -1
	if be.cursor >= 0 && be.cursor < len(be.selectableLines) {
		selectedLine = be.selectableLines[be.cursor]
	}
	
	// Adjust scroll to keep selection visible with some padding
	if selectedLine >= 0 && needsScroll {
		// Keep 1 line of context above/below when possible
		scrollPadding := 1
		
		// If selected line would be above viewport
		if selectedLine < be.scrollOffset + scrollPadding {
			be.scrollOffset = max(0, selectedLine - scrollPadding)
		} else if selectedLine >= be.scrollOffset + viewportHeight - scrollPadding {
			// If selected line would be below viewport
			be.scrollOffset = min(totalLines - viewportHeight, selectedLine - viewportHeight + scrollPadding + 1)
		}
	}
	
	// Ensure scroll offset is valid
	if needsScroll {
		maxScroll := max(0, totalLines - viewportHeight)
		be.scrollOffset = max(0, min(be.scrollOffset, maxScroll))
	} else {
		be.scrollOffset = 0
	}
	
	// Build display with fixed layout
	var displayLines []string
	
	if needsScroll {
		// Top indicator or spacing
		if be.scrollOffset > 0 {
			displayLines = append(displayLines, styles.MutedStyle().Render("↑ More above"))
		} else {
			displayLines = append(displayLines, "") // Empty line for consistent spacing
		}
		
		// Content lines
		startLine := be.scrollOffset
		endLine := min(be.scrollOffset + viewportHeight, totalLines)
		displayLines = append(displayLines, be.renderedLines[startLine:endLine]...)
		
		// Pad if needed to maintain consistent height
		for len(displayLines) < viewportHeight + 1 {
			displayLines = append(displayLines, "")
		}
		
		// Bottom indicator or spacing
		if be.scrollOffset + viewportHeight < totalLines {
			displayLines = append(displayLines, styles.MutedStyle().Render("↓ More below"))
		} else {
			displayLines = append(displayLines, "") // Empty line for consistent spacing
		}
	} else {
		// No scrolling needed, show all content
		displayLines = be.renderedLines
		
		// Pad to fill the content area if needed
		for len(displayLines) < contentHeight {
			displayLines = append(displayLines, "")
		}
	}
	
	// Join and render
	content := strings.Join(displayLines, "\n")
	
	return styles.BoxStyle().
		Width(be.width).
		Height(height).
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