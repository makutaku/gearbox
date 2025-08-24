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

// SelectableItemType represents the type of item that can be selected
type SelectableItemType int

const (
	ItemTypeCategory SelectableItemType = iota
	ItemTypeBundle
)

// SelectableItem represents an item that can be navigated to
type SelectableItem struct {
	Type     SelectableItemType
	Name     string      // Category name or bundle name
	Category string      // Category this item belongs to (for bundles)
	LineIndex int        // Line index in rendered content
}

// BundleExplorerNew demonstrates the new layout system
type BundleExplorerNew struct {
	// Data
	bundles        []orchestrator.BundleConfig
	installedTools map[string]*manifest.InstallationRecord
	
	// UI state
	cursor          int
	selectedBundle  string // Track selected bundle by name instead of index
	expandedBundles map[string]bool
	expandedCategories map[string]bool // Track expanded state of categories
	selectedCategory string
	categories      []string // Dynamically discovered from bundle data
	
	// Navigation state
	selectableItems []SelectableItem // All items that can be navigated to
	selectedIndex   int              // Current position in selectableItems
	
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
		expandedCategories: make(map[string]bool),
		installedTools:  make(map[string]*manifest.InstallationRecord),
		categories:      []string{}, // Will be discovered from bundle data
		cursor:          0,
		selectedBundle:  "", // Will be set when data is loaded
		selectedIndex:   0,
		selectableItems: []SelectableItem{},
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
		case "a":
			// Expand all in current category (if on category)
			if be.selectedIndex >= 0 && be.selectedIndex < len(be.selectableItems) {
				item := be.selectableItems[be.selectedIndex]
				if item.Type == ItemTypeCategory {
					be.expandAllInCategory(item.Name, true)
				}
			}
		case "A":
			// Collapse all in current category (if on category)
			if be.selectedIndex >= 0 && be.selectedIndex < len(be.selectableItems) {
				item := be.selectableItems[be.selectedIndex]
				if item.Type == ItemTypeCategory {
					be.expandAllInCategory(item.Name, false)
				}
			}
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
	
	// Build header text based on current selection
	var headerText string
	if be.selectedIndex >= 0 && be.selectedIndex < len(be.selectableItems) {
		item := be.selectableItems[be.selectedIndex]
		if item.Type == ItemTypeCategory {
			// Show category info when category is selected
			headerText = fmt.Sprintf(
				"Bundle Explorer | Selected: %s Category | Filter: %s",
				strings.Title(item.Name),
				be.selectedCategory,
			)
		} else {
			// Show bundle info when bundle is selected
			headerText = fmt.Sprintf(
				"Bundle Explorer | Category: %s | %d bundles",
				be.selectedCategory,
				len(be.bundles),
			)
		}
	} else {
		// Default header
		headerText = fmt.Sprintf(
			"Bundle Explorer | Category: %s | %d bundles",
			be.selectedCategory,
			len(be.bundles),
		)
	}
	
	header := headerStyle.Render(headerText)
	
	// Footer (help) - update to show category actions when on category
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	var footerText string
	if be.selectedIndex >= 0 && be.selectedIndex < len(be.selectableItems) {
		item := be.selectableItems[be.selectedIndex]
		if item.Type == ItemTypeCategory {
			footerText = "[↑/↓] Navigate  [Enter/Space] Toggle  [a] Expand All  [A] Collapse All  [c] Filter  [Tab] Switch View"
		} else {
			footerText = "[↑/↓] Navigate  [Enter/Space] Toggle  [c] Cycle Category  [i] Install Bundle  [Tab] Switch View"
		}
	} else {
		footerText = "[↑/↓] Navigate  [Enter/Space] Toggle  [c] Cycle Category  [i] Install Bundle  [Tab] Switch View"
	}
	
	footer := footerStyle.Render(footerText)
	
	// Content (bundle list with cursor highlighting)
	be.updateContent()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		be.viewport.View(),
		footer,
	)
}

// updateContent shows appropriate content based on current bundle data state
func (be *BundleExplorerNew) updateContent() {
	if !be.ready {
		return
	}
	
	// If no bundles data loaded yet, show loading state
	if len(be.bundles) == 0 {
		be.setLoadingState()
		return
	}
	
	// Show bundles content
	be.updateViewportContentTUI()
}

// setLoadingState shows loading message without any processing
func (be *BundleExplorerNew) setLoadingState() {
	loadingContent := `Loading bundles...

Discovering available tool bundles.
This happens in the background.`
	be.viewport.SetContent(loadingContent)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (be *BundleExplorerNew) updateViewportContentTUI() {
	be.rebuildRenderedContent()
	
	// Find the line index of the currently selected item
	selectedLineIndex := -1
	if be.selectedIndex >= 0 && be.selectedIndex < len(be.selectableItems) {
		selectedLineIndex = be.selectableItems[be.selectedIndex].LineIndex
	}
	
	// Apply highlighting to the correct line
	var lines []string
	for i, line := range be.renderedLines {
		if i == selectedLineIndex {
			// Highlight the selected line
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
	
	// Calculate how many lines this bundle needs (header + expanded details if applicable)
	totalLinesNeeded := 1 // Always need at least the header line
	if be.selectedBundle != "" && be.expandedBundles[be.selectedBundle] {
		// Find the bundle to calculate its expanded content size
		for _, bundle := range be.bundles {
			if bundle.Name == be.selectedBundle {
				details := be.renderBundleDetails(bundle)
				totalLinesNeeded += len(details)
				break
			}
		}
	}
	
	// Ensure selected line is visible by scrolling viewport
	if lineIndex < top {
		// Line above viewport - scroll up
		// Try to include category header context when possible
		contextOffset := lineIndex
		
		// Look backwards for a category header (ends with " Tier")
		for i := lineIndex - 1; i >= 0 && i >= lineIndex - 3; i-- {
			if i < len(be.renderedLines) && strings.HasSuffix(be.renderedLines[i], " Tier") {
				contextOffset = i
				break
			}
		}
		
		be.viewport.SetYOffset(contextOffset)
	} else if lineIndex + totalLinesNeeded - 1 > bottom {
		// Line (plus expanded content) extends below viewport - scroll down
		// Ensure there's enough space to show the header and all expanded details
		newOffset := lineIndex - be.viewport.Height + totalLinesNeeded
		if newOffset < 0 {
			newOffset = 0
		}
		be.viewport.SetYOffset(newOffset)
	}
}

// ensureSelectionVisible syncs viewport to show the currently selected bundle
func (be *BundleExplorerNew) ensureSelectionVisible() {
	if be.selectedBundle == "" {
		return
	}
	
	// Find the line index of the currently selected bundle
	if lineIndex, exists := be.bundleLineMap[be.selectedBundle]; exists {
		be.syncViewportWithLine(lineIndex)
	}
}

// ensureSelectionVisibleByIndex syncs viewport to show the currently selected item by index
func (be *BundleExplorerNew) ensureSelectionVisibleByIndex() {
	if be.selectedIndex < 0 || be.selectedIndex >= len(be.selectableItems) {
		return
	}
	
	// Get the line index of the selected item
	item := be.selectableItems[be.selectedIndex]
	be.syncViewportWithLine(item.LineIndex)
}

// SetData updates data and refreshes content
func (be *BundleExplorerNew) SetData(bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	be.bundles = bundles
	be.installedTools = installed
	
	// Dynamically discover categories from the bundle data
	be.categories = be.discoverCategories()
	
	// Initialize categories as expanded by default
	if len(be.expandedCategories) == 0 {
		for _, cat := range be.categories {
			be.expandedCategories[cat] = true
		}
	}
	
	// If selectedCategory is not valid anymore, reset it
	validCategory := false
	for _, cat := range be.categories {
		if cat == be.selectedCategory {
			validCategory = true
			break
		}
	}
	if !validCategory {
		be.selectedCategory = "" // Show all categories
	}
	
	// Build selectable items to initialize navigation
	if be.ready {
		be.rebuildRenderedContent()
		
		// Initialize selectedIndex if not set
		if be.selectedIndex == 0 && len(be.selectableItems) > 0 {
			// Set to first item
			be.selectedIndex = 0
			item := be.selectableItems[0]
			if item.Type == ItemTypeBundle {
				be.selectedBundle = item.Name
			}
		}
		
		be.updateContent()
		// Ensure selected item is visible after data changes
		be.ensureSelectionVisibleByIndex()
	}
}

// updateViewportContent is deprecated - using TUI best practices instead

// Business logic methods - robust approach with line index mapping
func (be *BundleExplorerNew) rebuildRenderedContent() {
	be.renderedLines = nil
	be.bundleLineMap = make(map[string]int)
	be.selectableItems = nil // Reset selectable items
	
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
		
		// Add category as a selectable item
		be.selectableItems = append(be.selectableItems, SelectableItem{
			Type:      ItemTypeCategory,
			Name:      category,
			Category:  category,
			LineIndex: lineIndex,
		})
		
		// Category header with expansion indicator and stats
		var expandIcon string
		if be.expandedCategories[category] {
			expandIcon = "▼"
		} else {
			expandIcon = "▶"
		}
		
		// Calculate category stats
		installedInCategory := 0
		totalInCategory := len(bundles)
		for _, bundle := range bundles {
			allInstalled := true
			for _, toolName := range bundle.Tools {
				if _, ok := be.installedTools[toolName]; !ok {
					allInstalled = false
					break
				}
			}
			if allInstalled && len(bundle.Tools) > 0 {
				installedInCategory++
			}
		}
		
		subtitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))
		categoryTitle := fmt.Sprintf("%s %s Tier (%d/%d bundles)", 
			expandIcon, 
			strings.Title(category), 
			installedInCategory,
			totalInCategory,
		)
		be.renderedLines = append(be.renderedLines, subtitleStyle.Render(categoryTitle))
		lineIndex++
		
		// Only show bundles if category is expanded
		if be.expandedCategories[category] {
			// Bundles in this category
			for _, bundle := range bundles {
				// Add bundle as a selectable item
				be.selectableItems = append(be.selectableItems, SelectableItem{
					Type:      ItemTypeBundle,
					Name:      bundle.Name,
					Category:  category,
					LineIndex: lineIndex,
				})
				
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

// discoverCategories dynamically discovers all categories from bundle data
func (be *BundleExplorerNew) discoverCategories() []string {
	categorySet := make(map[string]bool)
	bundlesByCategory := be.groupBundlesByCategory()
	
	// Collect all categories that actually have bundles
	for category, bundles := range bundlesByCategory {
		if len(bundles) > 0 {
			categorySet[category] = true
		}
	}
	
	// Convert to sorted slice for consistent ordering
	var categories []string
	
	// Preferred order for common categories (if they exist)
	preferredOrder := []string{"foundation", "language", "domain", "workflow", "infrastructure", "legacy"}
	
	for _, category := range preferredOrder {
		if categorySet[category] {
			categories = append(categories, category)
			delete(categorySet, category) // Remove from set so we don't add it twice
		}
	}
	
	// Add any remaining categories alphabetically
	for category := range categorySet {
		categories = append(categories, category)
	}
	
	return categories
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
	if len(be.selectableItems) == 0 {
		return
	}
	
	// Move to previous item
	if be.selectedIndex > 0 {
		be.selectedIndex--
		
		// Update selectedBundle if we're on a bundle
		item := be.selectableItems[be.selectedIndex]
		if item.Type == ItemTypeBundle {
			be.selectedBundle = item.Name
		} else {
			be.selectedBundle = "" // Clear selection when on category
		}
		
		if be.ready {
			be.updateContent()
			// Ensure selected line stays visible after navigation
			be.ensureSelectionVisibleByIndex()
		}
	}
}

func (be *BundleExplorerNew) moveDown() {
	if len(be.selectableItems) == 0 {
		return
	}
	
	// Move to next item
	if be.selectedIndex < len(be.selectableItems)-1 {
		be.selectedIndex++
		
		// Update selectedBundle if we're on a bundle
		item := be.selectableItems[be.selectedIndex]
		if item.Type == ItemTypeBundle {
			be.selectedBundle = item.Name
		} else {
			be.selectedBundle = "" // Clear selection when on category
		}
		
		if be.ready {
			be.updateContent()
			// Ensure selected line stays visible after navigation
			be.ensureSelectionVisibleByIndex()
		}
	}
}

func (be *BundleExplorerNew) toggleExpanded() {
	if be.selectedIndex < 0 || be.selectedIndex >= len(be.selectableItems) {
		return
	}
	
	item := be.selectableItems[be.selectedIndex]
	
	switch item.Type {
	case ItemTypeCategory:
		// Toggle category expansion
		be.expandedCategories[item.Name] = !be.expandedCategories[item.Name]
	case ItemTypeBundle:
		// Toggle bundle expansion
		be.expandedBundles[item.Name] = !be.expandedBundles[item.Name]
	}
	
	// Need to rebuild content when expanding/collapsing
	if be.ready {
		be.updateContent()
		// Ensure selected line stays visible after expansion/collapse
		be.ensureSelectionVisibleByIndex()
	}
}

// expandAllInCategory expands or collapses all bundles in a category
func (be *BundleExplorerNew) expandAllInCategory(category string, expand bool) {
	bundlesByCategory := be.groupBundlesByCategory()
	if bundles, ok := bundlesByCategory[category]; ok {
		for _, bundle := range bundles {
			be.expandedBundles[bundle.Name] = expand
		}
	}
	
	// Rebuild content to show changes
	if be.ready {
		be.updateContent()
		be.ensureSelectionVisibleByIndex()
	}
}

func (be *BundleExplorerNew) cycleCategory() {
	if len(be.categories) <= 1 {
		return // No point in cycling if there's 0 or 1 categories
	}
	
	// Find current category index
	currentIndex := -1
	for i, cat := range be.categories {
		if cat == be.selectedCategory {
			currentIndex = i
			break
		}
	}
	
	// If selectedCategory is empty ("") or not found, start from -1 so we cycle to index 0
	if currentIndex == -1 {
		if be.selectedCategory == "" {
			currentIndex = -1 // Will cycle to 0
		} else {
			currentIndex = -1 // Reset to start if current category not found
		}
	}
	
	// Cycle to next category (with wraparound)
	nextIndex := (currentIndex + 1) % (len(be.categories) + 1)
	
	// Special case: index len(be.categories) means "show all" (empty filter)
	if nextIndex == len(be.categories) {
		be.selectedCategory = ""
	} else {
		be.selectedCategory = be.categories[nextIndex]
	}
	
	// Reset selection to first item when changing category filter
	be.selectedIndex = 0
	be.selectedBundle = ""
	
	// Rebuild content to show new category filter
	if be.ready {
		be.updateContent()
		
		// Set selection to first available item
		if len(be.selectableItems) > 0 {
			be.selectedIndex = 0
			item := be.selectableItems[0]
			if item.Type == ItemTypeBundle {
				be.selectedBundle = item.Name
			}
		}
		
		be.ensureSelectionVisibleByIndex()
	}
}

// GetSelectedBundle returns the currently selected bundle
func (be *BundleExplorerNew) GetSelectedBundle() *orchestrator.BundleConfig {
	// Check if current selection is a bundle
	if be.selectedIndex >= 0 && be.selectedIndex < len(be.selectableItems) {
		item := be.selectableItems[be.selectedIndex]
		if item.Type == ItemTypeBundle {
			// Find the bundle by name
			for _, bundle := range be.bundles {
				if bundle.Name == item.Name {
					return &bundle
				}
			}
		}
	}
	
	// Fallback to old method if needed
	if be.selectedBundle != "" {
		bundles := be.getSelectableBundles()
		for _, bundle := range bundles {
			if bundle.Name == be.selectedBundle {
				return &bundle
			}
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