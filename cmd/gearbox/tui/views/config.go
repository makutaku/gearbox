package views

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	// TUI components (official Bubbles components)
	viewport viewport.Model
	ready    bool
	
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

	cv := &ConfigView{
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
	
	// Load existing configuration from ~/.gearboxrc
	cv.loadExistingConfig()
	
	return cv
}

// loadExistingConfig loads configuration values from ~/.gearboxrc
func (cv *ConfigView) loadExistingConfig() {
	configPath := filepath.Join(os.Getenv("HOME"), ".gearboxrc")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return // No existing config, use defaults
	}
	
	// Read config file
	file, err := os.Open(configPath)
	if err != nil {
		return // Can't read config, use defaults
	}
	defer file.Close()
	
	// Parse config file
	configValues := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			   (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			
			configValues[key] = value
		}
	}
	
	// Update config items with loaded values
	for i := range cv.configs {
		if value, exists := configValues[cv.configs[i].Key]; exists {
			cv.configs[i].Value = value
		}
	}
}

// SetSize updates the size of the config view
func (cv *ConfigView) SetSize(width, height int) {
	cv.width = width
	cv.height = height
	cv.textInput.Width = min(width-20, 50)
	
	// Initialize official viewport if not ready
	if !cv.ready {
		// Calculate viewport height: total - header - footer
		viewportHeight := height - 3 // Reserve 3 lines for header + footer
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		
		cv.viewport = viewport.New(width, viewportHeight)
		cv.viewport.SetContent("")
		cv.ready = true
	} else {
		// Update existing viewport
		viewportHeight := height - 3
		if viewportHeight < 1 {
			viewportHeight = 1
		}
		cv.viewport.Width = width
		cv.viewport.Height = viewportHeight
	}
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
	return cv.renderTUIStyle()
}

// renderTUIStyle uses proper TUI best practices with official Bubbles components
func (cv *ConfigView) renderTUIStyle() string {
	// Header (configuration title)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Bold(true).
		Padding(0, 1)
	
	header := headerStyle.Render("Configuration Settings")
	
	// Footer (help)
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)
	
	footer := footerStyle.Render(cv.renderHelpBar())
	
	// Content (configuration items with cursor highlighting)
	cv.updateViewportContentTUI()
	
	// Compose: header + viewport + footer (TUI best practice pattern)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		cv.viewport.View(),
		footer,
	)
}

// updateViewportContentTUI rebuilds content for the official viewport
func (cv *ConfigView) updateViewportContentTUI() {
	var lines []string
	
	// Add description
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Configure Gearbox behavior and preferences")
	lines = append(lines, desc, "")
	
	// Render all config items
	for i, config := range cv.configs {
		item := cv.renderConfigItem(config, i == cv.cursor)
		itemLines := strings.Split(item, "\n")
		lines = append(lines, itemLines...)
	}
	
	content := strings.Join(lines, "\n")
	cv.viewport.SetContent(content)
	
	// Sync viewport with cursor position (TUI best practice)
	cv.syncViewportWithCursor()
}

// syncViewportWithCursor ensures cursor is visible (TUI best practice)
func (cv *ConfigView) syncViewportWithCursor() {
	if len(cv.configs) == 0 {
		return
	}
	
	// Calculate which line the cursor is on (approximate)
	cursorLine := 2 // Start after description + empty line
	for i := 0; i < cv.cursor; i++ {
		item := cv.renderConfigItem(cv.configs[i], false)
		cursorLine += len(strings.Split(item, "\n"))
	}
	
	// Get viewport bounds
	top := cv.viewport.YOffset
	bottom := top + cv.viewport.Height - 1
	
	// Ensure cursor is visible by scrolling viewport
	if cursorLine < top {
		// Cursor above viewport - scroll up
		cv.viewport.SetYOffset(cursorLine)
	} else if cursorLine > bottom {
		// Cursor below viewport - scroll down
		cv.viewport.SetYOffset(cursorLine - cv.viewport.Height + 1)
	}
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
				value = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ " + item.Value)
			} else {
				value = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("✗ " + item.Value)
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
		selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
		line = selectedStyle.Render(line)
	}

	// Add description
	desc := "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(item.Description)

	// Add choices for choice type
	var choices string
	if item.Type == "choice" && len(item.Choices) > 0 {
		choices = "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Options: "+strings.Join(item.Choices, ", "))
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

	helpText := strings.Join(helps, "  ")
	
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
		// Use TUI best practice: update content and sync viewport
		if cv.ready {
			cv.updateViewportContentTUI()
		}
	}
}

func (cv *ConfigView) moveDown() {
	if cv.cursor < len(cv.configs)-1 {
		cv.cursor++
		// Use TUI best practice: update content and sync viewport
		if cv.ready {
			cv.updateViewportContentTUI()
		}
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
