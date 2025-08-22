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
	cursor          int
	scrollOffset    int
	expandedBundles map[string]bool
	selectedCategory string
	
	// Categories
	categories []string
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
	contentHeight := be.height - 6
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
	// Group bundles by category
	bundlesByCategory := be.groupBundlesByCategory()
	
	var lines []string
	bundleIndex := 0
	selectedLineIndex := -1
	currentLineIndex := 0
	
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
		lines = append(lines, categoryTitle)
		currentLineIndex++
		
		// Bundles in this category
		for _, bundle := range bundles {
			isSelected := bundleIndex == be.cursor
			if isSelected {
				selectedLineIndex = currentLineIndex
			}
			bundleLine := be.renderBundleLine(bundle, isSelected)
			lines = append(lines, bundleLine)
			currentLineIndex++
			bundleIndex++
			
			// Show expanded details
			if be.expandedBundles[bundle.Name] {
				details := be.renderBundleDetails(bundle)
				lines = append(lines, details...)
				currentLineIndex += len(details)
			}
		}
		
		lines = append(lines, "") // Empty line between categories
		currentLineIndex++
	}
	
	// Handle scrolling based on the selected line position
	visibleLines := height - 2
	if selectedLineIndex >= 0 {
		if selectedLineIndex >= be.scrollOffset+visibleLines {
			be.scrollOffset = selectedLineIndex - visibleLines + 1
		} else if selectedLineIndex < be.scrollOffset {
			be.scrollOffset = selectedLineIndex
		}
	}
	
	// Get visible portion
	start := be.scrollOffset
	end := min(start+visibleLines, len(lines))
	if start < len(lines) {
		lines = lines[start:end]
	}
	
	content := strings.Join(lines, "\n")
	
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
		details = append(details, indent+styles.ButtonStyle(false).Render("[i] Install Bundle"))
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
	if be.cursor > 0 && be.cursor <= len(bundles) {
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