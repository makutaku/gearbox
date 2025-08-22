package views

import (
	"fmt"
	"strings"

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
}

// ConfigItem represents a configuration item
type ConfigItem struct {
	Key         string
	Value       string
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
				// Save the value
				cv.configs[cv.cursor].Value = cv.textInput.Value()
				cv.editing = false
				cv.textInput.Blur()
				cv.textInput.SetValue("")
				// TODO: Actually save to config file
				return nil
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
			case "s":
				cv.saveConfig()
			}
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
	// Calculate visible items
	visibleItems := (height - 4) / 3 // Each config item takes ~3 lines
	if cv.cursor >= cv.scrollOffset+visibleItems {
		cv.scrollOffset = cv.cursor - visibleItems + 1
	} else if cv.cursor < cv.scrollOffset {
		cv.scrollOffset = cv.cursor
	}

	var lines []string

	// Add description
	desc := styles.MutedStyle().Render("Configure Gearbox behavior and preferences")
	lines = append(lines, desc, "")

	// Render visible config items
	for i := cv.scrollOffset; i < min(cv.scrollOffset+visibleItems, len(cv.configs)); i++ {
		item := cv.renderConfigItem(cv.configs[i], i == cv.cursor)
		lines = append(lines, item)
	}

	// Add scroll indicators
	if cv.scrollOffset > 0 {
		lines[2] = "↑ More above"
	}
	if cv.scrollOffset+visibleItems < len(cv.configs) {
		lines = append(lines, "↓ More below")
	}

	content := strings.Join(lines, "\n")

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

	// Apply selection style
	if selected {
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
		"[r] Reset to Default",
		"[s] Save All",
		"[Esc] Cancel Edit",
	}

	return styles.MutedStyle().Render(strings.Join(helps, "  "))
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
	// Reset current item to default value
	// TODO: Load defaults from config system
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

func (cv *ConfigView) saveConfig() {
	// TODO: Actually save configuration to file
	// For now, this is just a placeholder
}