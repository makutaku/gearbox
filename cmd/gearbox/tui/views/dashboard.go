package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	
	"gearbox/cmd/gearbox/tui/styles"
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
}

// SetData updates the dashboard data
func (d *Dashboard) SetData(tools []orchestrator.ToolConfig, bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	d.tools = tools
	d.bundles = bundles
	d.installedTools = installed
}

// Render returns the rendered dashboard view
func (d *Dashboard) Render() string {
	if d.width == 0 || d.height == 0 {
		return "Loading..."
	}

	// Calculate available content height (minus navigation and status bars)
	contentHeight := d.height - 4

	// Render components
	systemStatus := d.renderSystemStatus()
	quickActions := d.renderQuickActions()
	recentActivity := d.renderRecentActivity()
	recommendations := d.renderRecommendations()

	// Layout the dashboard
	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		systemStatus,
		lipgloss.NewStyle().Width(2).Render(" "),
		quickActions,
	)

	// Combine all sections
	sections := lipgloss.JoinVertical(
		lipgloss.Left,
		topRow,
		lipgloss.NewStyle().Height(1).Render(""),
		recentActivity,
		lipgloss.NewStyle().Height(1).Render(""),
		recommendations,
	)

	// Center the content
	return lipgloss.NewStyle().
		Width(d.width).
		Height(contentHeight).
		Padding(1, 2).
		Render(sections)
}

func (d *Dashboard) renderSystemStatus() string {
	installedCount := len(d.installedTools)
	totalTools := len(d.tools)
	bundleCount := d.countActiveBundles()
	diskUsage := d.calculateDiskUsage()

	content := fmt.Sprintf(
		"● Tools Installed: %d/%d\n● Bundles Active: %d\n● Disk Usage: %s\n● Health: ✓ %s",
		installedCount, totalTools, bundleCount, diskUsage, d.healthStatus,
	)

	width := (d.width - 6) / 2
	return styles.BoxStyle().
		Width(width).
		Height(6).
		Render(styles.TitleStyle().Render("System Status") + "\n" + content)
}

func (d *Dashboard) renderQuickActions() string {
	actions := []string{
		"[i] Install Tools",
		"[b] Browse Bundles",
		"[c] Configuration",
		"[h] Health Check",
	}

	content := strings.Join(actions, "\n")
	
	width := (d.width - 6) / 2
	return styles.BoxStyle().
		Width(width).
		Height(6).
		Render(styles.TitleStyle().Render("Quick Actions") + "\n" + content)
}

func (d *Dashboard) renderRecentActivity() string {
	// Get recent installations
	activities := d.getRecentActivities()
	
	if len(activities) == 0 {
		activities = append(activities, "No recent activity")
	}

	content := strings.Join(activities, "\n")
	
	return styles.BoxStyle().
		Width(d.width - 4).
		Render(styles.TitleStyle().Render("Recent Activity") + "\n" + content)
}

func (d *Dashboard) renderRecommendations() string {
	recommendations := d.getRecommendations()
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "💡 Everything looks good! Explore new tools with [t]")
	}

	content := strings.Join(recommendations, "\n")
	
	return styles.BoxStyle().
		Width(d.width - 4).
		Render(styles.TitleStyle().Render("Recommendations") + "\n" + content)
}

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