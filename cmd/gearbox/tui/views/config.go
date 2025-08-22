package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"gearbox/cmd/gearbox/tui/styles"
)

// ConfigView represents the configuration view
type ConfigView struct {
	width  int
	height int

	// Configuration items
	configs []ConfigItem
	cursor  int
	editing bool

	// Text input for editing
	textInput textinput.Model

	// Scroll state
	scrollOffset int
	
	// Status message
	statusMessage string
}

// ConfigItem represents a configuration item
type ConfigItem struct {
	Key         string
	Value       string
	Default     string
	Description string
	Type        string // "string", "number", "boolean", "choice"
	Choices     []string
	Editable    bool
}

// NewConfigView creates a new configuration view
func NewConfigView() *ConfigView {
	ti := textinput.New()
	ti.CharLimit = 100
	ti.Width = 50

	return &ConfigView{
		textInput: ti,
		configs: []ConfigItem{
			{
				Key:         "DEFAULT_BUILD_TYPE",
				Value:       "standard",
				Description: "Default build type for tool installations",
				Type:        "choice",
				Choices:     []string{"minimal", "standard", "maximum"},
				Editable:    true,
			},
			{
				Key:         "MAX_PARALLEL_JOBS",
				Value:       "4",
				Description: "Maximum number of parallel build jobs",
				Type:        "number",
				Editable:    true,
			},
			{
				Key:         "INSTALL_PREFIX",
				Value:       "/usr/local",
				Description: "Installation prefix for binaries",
				Type:        "string",
				Editable:    true,
			},
			{
				Key:         "USE_BUILD_CACHE",
				Value:       "true",
				Description: "Enable build cache for faster reinstallations",
				Type:        "boolean",
				Editable:    true,
			},
			{
				Key:         "CACHE_DIR",
				Value:       "~/tools/cache",
				Description: "Directory for build cache",
				Type:        "string",
				Editable:    true,
			},
			{
				Key:         "SKIP_COMMON_DEPS",
				Value:       "false",
				Description: "Skip installation of common dependencies",
				Type:        "boolean",
				Editable:    true,
			},
			{
				Key:         "RUN_TESTS",
				Value:       "false",
				Description: "Run test suites after building tools",
				Type:        "boolean",
				Editable:    true,
			},
			{
				Key:         "VERBOSE_OUTPUT",
				Value:       "false",
				Description: "Enable verbose output during installations",
				Type:        "boolean",
				Editable:    true,
			},
			{
				Key:         "AUTO_UPDATE_PATH",
				Value:       "true",
				Description: "Automatically update PATH after installations",
				Type:        "boolean",
				Editable:    true,
			},
			{
				Key:         "CLEANUP_BUILD_DIR",
				Value:       "true",
				Description: "Clean up build directories after installation",
				Type:        "boolean",
				Editable:    true,
			},
		},
	}
}

// SetSize updates the size of the config view
func (cv *ConfigView) SetSize(width, height int) {
	cv.width = width
	cv.height = height
	cv.textInput.Width = min(width-20, 50)
}

// Update handles config view updates
func (cv *ConfigView) Update(msg tea.Msg) tea.Cmd {
	if cv.editing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				cv.editing = false
				cv.textInput.Blur()
				cv.textInput.SetValue("")
				return nil
			case "enter":
				// Save the value and auto-save to file
				cv.configs[cv.cursor].Value = cv.textInput.Value()
				cv.editing = false
				cv.textInput.Blur()
				cv.textInput.SetValue("")
				// Auto-save the configuration
				return cv.saveConfig()
			default:
				var cmd tea.Cmd
				cv.textInput, cmd = cv.textInput.Update(msg)
				return cmd
			}
		}
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				cv.moveUp()
			case "down", "j":
				cv.moveDown()
			case "enter", " ":
				cv.startEditing()
			case "r":
				cv.resetToDefault()
				// Auto-save after reset
				return cv.saveConfig()
			case "R":
				// Reset all to defaults
				cv.resetAllToDefaults()
				// Auto-save after reset
				return cv.saveConfig()
			}
		case configSaveSuccessMsg:
			// Config saved successfully
			cv.statusMessage = "✓ Configuration saved to ~/.gearboxrc"
			// Clear message after a delay
			return tea.Tick(time.Second*3, func(time.Time) tea.Msg {
				return clearStatusMsg{}
			})
		case configSaveErrorMsg:
			// Handle save error
			cv.statusMessage = fmt.Sprintf("✗ Error saving config: %v", msg.err)
			// Clear message after a delay
			return tea.Tick(time.Second*5, func(time.Time) tea.Msg {
				return clearStatusMsg{}
			})
		case clearStatusMsg:
			cv.statusMessage = ""
			return nil
		}
	}

	return nil
}

// Render returns the rendered config view
func (cv *ConfigView) Render() string {
	if cv.width == 0 || cv.height == 0 {
		return "Loading..."
	}

	// Title
	title := styles.TitleStyle().Render("Configuration Settings")

	// Content
	contentHeight := cv.height - 4
	content := cv.renderContent(contentHeight)

	// Help bar
	helpBar := cv.renderHelpBar()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		helpBar,
	)
}

func (cv *ConfigView) renderContent(height int) string {
	contentHeight := height - 2 // Account for box borders
	
	// Build all content lines first
	var allLines []string
	
	// Add description
	desc := styles.MutedStyle().Render("Configure Gearbox behavior and preferences")
	allLines = append(allLines, desc, "")
	
	// Render all config items to calculate their actual height
	configStartLine := len(allLines)
	for i, config := range cv.configs {
		item := cv.renderConfigItem(config, i == cv.cursor)
		itemLines := strings.Split(item, "\n")
		allLines = append(allLines, itemLines...)
	}
	
	// Calculate which line the cursor is on
	cursorLine := configStartLine
	for i := 0; i < cv.cursor; i++ {
		item := cv.renderConfigItem(cv.configs[i], false)
		cursorLine += len(strings.Split(item, "\n"))
	}
	
	// Check if we need scroll indicators
	needsScrollUp := len(allLines) > contentHeight && cv.scrollOffset > 0
	needsScrollDown := len(allLines) > contentHeight && 
		(cv.scrollOffset + contentHeight) < len(allLines)
	
	// Adjust effective height for scroll indicators
	effectiveHeight := contentHeight
	if needsScrollUp {
		effectiveHeight--
	}
	if needsScrollDown {
		effectiveHeight--
	}
	
	// Adjust scroll to keep cursor visible
	if cursorLine < cv.scrollOffset {
		cv.scrollOffset = cursorLine
	} else if cursorLine >= cv.scrollOffset + effectiveHeight {
		// Find how many lines the current item takes
		currentItemLines := len(strings.Split(cv.renderConfigItem(cv.configs[cv.cursor], true), "\n"))
		cv.scrollOffset = cursorLine - effectiveHeight + currentItemLines
	}
	
	// Clamp scroll offset
	maxScroll := len(allLines) - effectiveHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	cv.scrollOffset = max(0, min(cv.scrollOffset, maxScroll))
	
	// Build display lines
	var displayLines []string
	
	// Add scroll indicators and visible content
	if needsScrollUp {
		displayLines = append(displayLines, styles.MutedStyle().Render("↑ More above"))
	}
	
	// Add visible lines
	start := cv.scrollOffset
	end := min(start + effectiveHeight, len(allLines))
	if start < len(allLines) {
		displayLines = append(displayLines, allLines[start:end]...)
	}
	
	if needsScrollDown {
		displayLines = append(displayLines, styles.MutedStyle().Render("↓ More below"))
	}
	
	content := strings.Join(displayLines, "\n")
	
	return styles.BoxStyle().
		Width(cv.width).
		Height(height).
		Render(content)
}

func (cv *ConfigView) renderConfigItem(item ConfigItem, selected bool) string {
	// Key and value
	key := fmt.Sprintf("%-25s", item.Key)
	
	var value string
	if cv.editing && selected {
		value = cv.textInput.View()
	} else {
		// Format value based on type
		switch item.Type {
		case "boolean":
			if item.Value == "true" {
				value = styles.SuccessStyle().Render("✓ " + item.Value)
			} else {
				value = styles.MutedStyle().Render("✗ " + item.Value)
			}
		case "choice":
			value = fmt.Sprintf("[%s]", item.Value)
		default:
			value = item.Value
		}
	}

	line := fmt.Sprintf("%s = %s", key, value)

	// Apply selection style only when not editing
	if selected && !cv.editing {
		line = styles.SelectedStyle().Render(line)
	}

	// Add description
	desc := "  " + styles.MutedStyle().Render(item.Description)

	// Add choices for choice type
	var choices string
	if item.Type == "choice" && len(item.Choices) > 0 {
		choices = "  " + styles.MutedStyle().Render("Options: "+strings.Join(item.Choices, ", "))
	}

	// Combine all parts
	result := line + "\n" + desc
	if choices != "" {
		result += "\n" + choices
	}

	return result
}

func (cv *ConfigView) renderHelpBar() string {
	helps := []string{
		"[↑/↓] Navigate",
		"[Enter/Space] Edit",
		"[r] Reset Field",
		"[R] Reset All to Defaults",
		"[Esc] Cancel Edit",
	}

	helpText := styles.MutedStyle().Render(strings.Join(helps, "  "))
	
	// Add status message if present
	if cv.statusMessage != "" {
		return helpText + "\n" + cv.statusMessage
	}
	
	return helpText
}

// Helper methods

func (cv *ConfigView) moveUp() {
	if cv.cursor > 0 {
		cv.cursor--
	}
}

func (cv *ConfigView) moveDown() {
	if cv.cursor < len(cv.configs)-1 {
		cv.cursor++
	}
}

func (cv *ConfigView) startEditing() {
	if !cv.configs[cv.cursor].Editable {
		return
	}

	cv.editing = true
	cv.textInput.SetValue(cv.configs[cv.cursor].Value)
	cv.textInput.Focus()
	cv.textInput.CursorEnd()
}

func (cv *ConfigView) resetToDefault() {
	// Reset current field to its default value
	// Use hardcoded defaults for now since configs don't have Default field populated yet
	switch cv.configs[cv.cursor].Key {
	case "DEFAULT_BUILD_TYPE":
		cv.configs[cv.cursor].Value = "standard"
	case "MAX_PARALLEL_JOBS":
		cv.configs[cv.cursor].Value = "4"
	case "INSTALL_PREFIX":
		cv.configs[cv.cursor].Value = "/usr/local"
	case "USE_BUILD_CACHE":
		cv.configs[cv.cursor].Value = "true"
	case "CACHE_DIR":
		cv.configs[cv.cursor].Value = "~/tools/cache"
	case "SKIP_COMMON_DEPS":
		cv.configs[cv.cursor].Value = "false"
	case "RUN_TESTS":
		cv.configs[cv.cursor].Value = "false"
	case "VERBOSE_OUTPUT":
		cv.configs[cv.cursor].Value = "false"
	case "AUTO_UPDATE_PATH":
		cv.configs[cv.cursor].Value = "true"
	case "CLEANUP_BUILD_DIR":
		cv.configs[cv.cursor].Value = "true"
	}
}

func (cv *ConfigView) resetAllToDefaults() {
	// Reset all fields to their default values
	for i := range cv.configs {
		// Temporarily save cursor position
		originalCursor := cv.cursor
		cv.cursor = i
		cv.resetToDefault()
		cv.cursor = originalCursor
	}
}

func (cv *ConfigView) saveConfig() tea.Cmd {
	return func() tea.Msg {
		// Build config file content
		var lines []string
		lines = append(lines, "# Gearbox Configuration File")
		lines = append(lines, "# Generated by Gearbox TUI")
		lines = append(lines, "")
		
		// Add each configuration setting
		for _, config := range cv.configs {
			// Add comment with description
			lines = append(lines, fmt.Sprintf("# %s", config.Description))
			// Add the key=value pair
			lines = append(lines, fmt.Sprintf("%s=%s", config.Key, config.Value))
			lines = append(lines, "")
		}
		
		// Write to ~/.gearboxrc
		configPath := filepath.Join(os.Getenv("HOME"), ".gearboxrc")
		content := strings.Join(lines, "\n")
		
		err := os.WriteFile(configPath, []byte(content), 0644)
		if err != nil {
			return configSaveErrorMsg{err}
		}
		
		// Return a success message
		return configSaveSuccessMsg{}
	}
}

// Message types for config operations
type configSaveSuccessMsg struct{}
type configSaveErrorMsg struct{ err error }
type clearStatusMsg struct{}