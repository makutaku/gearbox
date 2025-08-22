package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a color theme for the TUI
type Theme struct {
	Name        string
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Success     lipgloss.Color
	Error       lipgloss.Color
	Warning     lipgloss.Color
	Info        lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	Border      lipgloss.Color
	Highlight   lipgloss.Color
	Muted       lipgloss.Color
}

// Default themes
var (
	DefaultTheme = Theme{
		Name:        "default",
		Primary:     lipgloss.Color("62"),    // Teal
		Secondary:   lipgloss.Color("99"),    // Purple
		Success:     lipgloss.Color("42"),    // Green
		Error:       lipgloss.Color("9"),     // Red
		Warning:     lipgloss.Color("214"),   // Orange
		Info:        lipgloss.Color("86"),    // Cyan
		Background:  lipgloss.Color("235"),   // Dark gray
		Foreground:  lipgloss.Color("252"),   // Light gray
		Border:      lipgloss.Color("240"),   // Medium gray
		Highlight:   lipgloss.Color("226"),   // Yellow
		Muted:       lipgloss.Color("246"),   // Gray
	}

	DarkTheme = Theme{
		Name:        "dark",
		Primary:     lipgloss.Color("39"),    // Blue
		Secondary:   lipgloss.Color("170"),   // Purple
		Success:     lipgloss.Color("70"),    // Green
		Error:       lipgloss.Color("196"),   // Red
		Warning:     lipgloss.Color("208"),   // Orange
		Info:        lipgloss.Color("45"),    // Cyan
		Background:  lipgloss.Color("232"),   // Very dark
		Foreground:  lipgloss.Color("255"),   // White
		Border:      lipgloss.Color("237"),   // Dark gray
		Highlight:   lipgloss.Color("220"),   // Yellow
		Muted:       lipgloss.Color("243"),   // Medium gray
	}

	LightTheme = Theme{
		Name:        "light",
		Primary:     lipgloss.Color("25"),    // Blue
		Secondary:   lipgloss.Color("91"),    // Purple
		Success:     lipgloss.Color("28"),    // Green
		Error:       lipgloss.Color("124"),   // Red
		Warning:     lipgloss.Color("166"),   // Orange
		Info:        lipgloss.Color("31"),    // Cyan
		Background:  lipgloss.Color("253"),   // Light gray
		Foreground:  lipgloss.Color("235"),   // Dark gray
		Border:      lipgloss.Color("250"),   // Light gray
		Highlight:   lipgloss.Color("226"),   // Yellow
		Muted:       lipgloss.Color("245"),   // Gray
	}
)

// CurrentTheme holds the active theme
var CurrentTheme = DefaultTheme

// SetTheme sets the current theme
func SetTheme(theme Theme) {
	CurrentTheme = theme
}

// GetTheme returns a theme by name
func GetTheme(name string) Theme {
	switch name {
	case "dark":
		return DarkTheme
	case "light":
		return LightTheme
	default:
		return DefaultTheme
	}
}

// Style builders using the current theme

// TitleStyle returns a style for titles
func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(CurrentTheme.Primary).
		MarginBottom(1)
}

// SubtitleStyle returns a style for subtitles
func SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Secondary).
		MarginBottom(1)
}

// BoxStyle returns a style for bordered boxes
func BoxStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Border).
		Padding(1)
}

// SelectedStyle returns a style for selected items
func SelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(CurrentTheme.Primary).
		Foreground(CurrentTheme.Background).
		Bold(true)
}

// ErrorStyle returns a style for error messages
func ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Error).
		Bold(true)
}

// SuccessStyle returns a style for success messages
func SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Success).
		Bold(true)
}

// WarningStyle returns a style for warning messages
func WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Warning).
		Bold(true)
}

// InfoStyle returns a style for info messages
func InfoStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Info)
}

// MutedStyle returns a style for muted text
func MutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Muted)
}

// HighlightStyle returns a style for highlighted text
func HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Highlight).
		Bold(true)
}

// NavBarStyle returns a style for the navigation bar
func NavBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(CurrentTheme.Primary).
		Foreground(CurrentTheme.Background).
		Padding(0, 2)
}

// StatusBarStyle returns a style for the status bar
func StatusBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(CurrentTheme.Background).
		Foreground(CurrentTheme.Muted).
		Padding(0, 2)
}

// ButtonStyle returns a style for buttons
func ButtonStyle(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().
			Background(CurrentTheme.Primary).
			Foreground(CurrentTheme.Background).
			Padding(0, 2).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Background(CurrentTheme.Background).
		Foreground(CurrentTheme.Foreground).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Border).
		Padding(0, 2)
}

// ProgressBarStyle returns styles for progress bars
func ProgressBarStyle() (full lipgloss.Style, empty lipgloss.Style) {
	full = lipgloss.NewStyle().
		Foreground(CurrentTheme.Success).
		Background(CurrentTheme.Success)
	
	empty = lipgloss.NewStyle().
		Foreground(CurrentTheme.Border).
		Background(CurrentTheme.Background)
	
	return full, empty
}

// ListItemStyle returns a style for list items
func ListItemStyle(selected bool, installed bool) lipgloss.Style {
	style := lipgloss.NewStyle()
	
	if selected {
		style = style.
			Background(CurrentTheme.Primary).
			Foreground(CurrentTheme.Background).
			Bold(true)
	} else {
		style = style.Foreground(CurrentTheme.Foreground)
	}
	
	if installed {
		style = style.Italic(true)
	}
	
	return style
}

// TabStyle returns a style for tabs
func TabStyle(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().
			Background(CurrentTheme.Primary).
			Foreground(CurrentTheme.Background).
			Padding(0, 2).
			Bold(true)
	}
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Muted).
		Padding(0, 2)
}