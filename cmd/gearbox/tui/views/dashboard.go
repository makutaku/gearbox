package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// Dashboard represents the main dashboard view
type Dashboard struct {
	width  int
	height int
	
	// Data
	tools          []orchestrator.ToolConfig
	bundles        []orchestrator.BundleConfig
	installedTools map[string]*manifest.InstallationRecord
	healthStatus   string

	// TUI components (official Bubbles components)
	viewport       viewport.Model
	ready          bool
}

// NewDashboard creates a new dashboard view
func NewDashboard() *Dashboard {
	return &Dashboard{
		installedTools: make(map[string]*manifest.InstallationRecord),
		healthStatus:   "All systems OK",
	}
}

// SetSize updates the size of the dashboard
func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height
	
	// Initialize official viewport if not ready
	if !d.ready {
		// Calculate viewport height: total - header - footer
		viewportHeight := height - 3 // Reserve 3 lines for header + footer
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		
		d.viewport = viewport.New(width, viewportHeight)
		d.viewport.SetContent("")
		d.ready = true
	} else {
		// Update existing viewport
		viewportHeight := height - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		d.viewport.Width = width
		d.viewport.Height = viewportHeight
	}
}

// SetData updates the dashboard data
func (d *Dashboard) SetData(tools []orchestrator.ToolConfig, bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	d.tools = tools
	d.bundles = bundles
	d.installedTools = installed
	
	if d.ready {
		d.updateViewportContentTUI()
	}
}

// Update handles dashboard updates
func (d *Dashboard) Update(msg tea.Msg) tea.Cmd {
	// Dashboard doesn't handle specific keys since the quick actions
	// are handled by the main app navigation (i, b, c, h keys)
	// The dashboard is primarily a read-only overview
	return nil
}

// Render returns the rendered dashboard view
func (d *Dashboard) Render() string {
	return d.renderTUIStyle()
}

// renderTUIStyle uses proper TUI best practices with official Bubbles components
func (d *Dashboard) renderTUIStyle() string {
	// Header (dashboard title with summary)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 1)
	
	installedCount := len(d.installedTools)
	totalTools := len(d.tools)
	bundleCount := d.countActiveBundles()
	
	header := headerStyle.Render(fmt.Sprintf(
		"Dashboard | Tools: %d/%d | Bundles: %d | %s",
		installedCount, totalTools, bundleCount, d.healthStatus,
	))
	
	// Footer (help)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	footer := footerStyle.Render(
		"[T] Tool Browser  [B] Bundle Explorer  [M] Monitor  [C] Configuration  [H] Health Check",
	)
	
	// Content (dashboard sections)
	d.updateViewportContentTUI()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		d.viewport.View(),
		footer,
	)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (d *Dashboard) updateViewportContentTUI() {
	var sections []string
	
	// System Status Section
	installedCount := len(d.installedTools)
	totalTools := len(d.tools)
	bundleCount := d.countActiveBundles()
	diskUsage := d.calculateDiskUsage()
	
	systemSection := fmt.Sprintf(
		"📊 System Status\n"+
		"   • Tools Installed: %d/%d\n"+
		"   • Bundles Active: %d\n"+
		"   • Disk Usage: %s\n"+
		"   • Health: ✓ %s",
		installedCount, totalTools, bundleCount, diskUsage, d.healthStatus,
	)
	
	// Quick Actions Section
	quickActionsSection := "🚀 Quick Actions\n" +
		"   [T] Install Tools\n" +
		"   [B] Browse Bundles\n" +
		"   [M] Monitor Installations\n" +
		"   [C] Configuration\n" +
		"   [H] Health Check"
	
	// Recent Activity Section
	activities := d.getRecentActivities()
	if len(activities) == 0 {
		activities = append(activities, "   No recent activity")
	} else {
		// Prefix each activity with proper indentation
		for i, activity := range activities {
			activities[i] = "   " + activity
		}
	}
	
	recentSection := "📈 Recent Activity\n" + strings.Join(activities, "\n")
	
	// Recommendations Section
	recommendations := d.getRecommendations()
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "   💡 Everything looks good! Explore new tools with [T]")
	} else {
		// Prefix each recommendation with proper indentation
		for i, rec := range recommendations {
			recommendations[i] = "   " + rec
		}
	}
	
	recommendationsSection := "💡 Recommendations\n" + strings.Join(recommendations, "\n")
	
	// Combine all sections
	sections = append(sections, systemSection)
	sections = append(sections, "") // empty line
	sections = append(sections, quickActionsSection)
	sections = append(sections, "") // empty line
	sections = append(sections, recentSection)
	sections = append(sections, "") // empty line
	sections = append(sections, recommendationsSection)
	
	content := strings.Join(sections, "\n")
	d.viewport.SetContent(content)
}

// renderSystemStatus is deprecated - using TUI best practices instead
// renderQuickActions is deprecated - using TUI best practices instead
// renderRecentActivity is deprecated - using TUI best practices instead
// renderRecommendations is deprecated - using TUI best practices instead

// Helper methods

func (d *Dashboard) countActiveBundles() int {
	// Count bundles where all tools are installed
	activeCount := 0
	for _, bundle := range d.bundles {
		if d.isBundleActive(bundle) {
			activeCount++
		}
	}
	return activeCount
}

func (d *Dashboard) isBundleActive(bundle orchestrator.BundleConfig) bool {
	if len(bundle.Tools) == 0 {
		return false
	}
	
	for _, toolName := range bundle.Tools {
		if _, installed := d.installedTools[toolName]; !installed {
			return false
		}
	}
	return true
}

func (d *Dashboard) calculateDiskUsage() string {
	// Estimate disk usage based on installed tools
	totalMB := 0
	for toolName := range d.installedTools {
		// Rough estimates per tool type
		for _, tool := range d.tools {
			if tool.Name == toolName {
				switch tool.Language {
				case "rust":
					totalMB += 5
				case "go":
					totalMB += 10
				case "c", "cpp":
					totalMB += 15
				default:
					totalMB += 8
				}
				break
			}
		}
	}
	
	if totalMB < 1024 {
		return fmt.Sprintf("%d MB", totalMB)
	}
	return fmt.Sprintf("%.1f GB", float64(totalMB)/1024)
}

func (d *Dashboard) getRecentActivities() []string {
	activities := []string{}
	
	// Get recent installations
	for toolName, record := range d.installedTools {
		if record.InstalledAt.After(time.Now().Add(-24 * time.Hour)) {
			timeAgo := time.Since(record.InstalledAt)
			activities = append(activities, fmt.Sprintf("✓ Installed %s (%s ago)", toolName, formatDuration(timeAgo)))
		}
	}
	
	// Limit to 5 most recent
	if len(activities) > 5 {
		activities = activities[:5]
	}
	
	return activities
}

func (d *Dashboard) getRecommendations() []string {
	recommendations := []string{}
	
	// Check for starship without nerd-fonts
	hasStarship := false
	hasNerdFonts := false
	for toolName := range d.installedTools {
		if toolName == "starship" {
			hasStarship = true
		}
		if toolName == "nerd-fonts" {
			hasNerdFonts = true
		}
	}
	
	if hasStarship && !hasNerdFonts {
		recommendations = append(recommendations, "💡 Consider installing 'nerd-fonts' - pairs well with your installed starship")
	}
	
	if hasNerdFonts && !hasStarship {
		recommendations = append(recommendations, "💡 Consider installing 'starship' - a customizable prompt that works great with Nerd Fonts")
	}
	
	// Suggest beginner bundle if no tools installed
	if len(d.installedTools) == 0 {
		recommendations = append(recommendations, "💡 New to Gearbox? Try the 'beginner' bundle for essential tools")
	}
	
	return recommendations
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}

// calculateHalfBoxWidth is deprecated - using TUI best practices instead
// calculateFullBoxWidth is deprecated - using TUI best practices instead