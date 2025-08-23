package views

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SudoPrompt represents a secure sudo password prompt
type SudoPrompt struct {
	textInput textinput.Model
	err       error
	cancelled bool
	submitted bool
	width     int
	height    int
}

// SudoPromptResult represents the result of the sudo prompt
type SudoPromptResult struct {
	Password  string
	Cancelled bool
	Error     error
}

// NewSudoPrompt creates a new sudo password prompt
func NewSudoPrompt() *SudoPrompt {
	ti := textinput.New()
	ti.Placeholder = "Enter sudo password..."
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &SudoPrompt{
		textInput: ti,
	}
}

// SetSize updates the prompt size
func (sp *SudoPrompt) SetSize(width, height int) {
	sp.width = width
	sp.height = height
	
	// Center the input field
	if width > 60 {
		sp.textInput.Width = 50
	} else {
		sp.textInput.Width = width - 10
	}
}

// Update handles prompt updates
func (sp *SudoPrompt) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			sp.submitted = true
			return nil
		case tea.KeyEsc:
			sp.cancelled = true
			return nil
		case tea.KeyCtrlC:
			sp.cancelled = true
			return nil
		}
	}

	var cmd tea.Cmd
	sp.textInput, cmd = sp.textInput.Update(msg)
	return cmd
}

// Render returns the prompt's view
func (sp *SudoPrompt) Render() string {
	if sp.width == 0 || sp.height == 0 {
		return "Loading sudo prompt..."
	}

	// Title
	title := "🔐 Sudo Authentication Required"
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	// Message
	message := "Installation requires administrative privileges.\nPlease enter your sudo password to continue:"
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		MarginBottom(2)

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(0, 1).
		MarginBottom(2)

	// Instructions
	instructions := "Press Enter to confirm • Esc to cancel"
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	// Error message
	var errorMsg string
	if sp.err != nil {
		errorMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render("❌ " + sp.err.Error())
	}

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(title),
		messageStyle.Render(message),
		inputStyle.Render(sp.textInput.View()),
		instructStyle.Render(instructions),
		errorMsg,
	)

	// Center everything in the available space
	return lipgloss.Place(
		sp.width,
		sp.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// IsSubmitted returns true if the user submitted the password
func (sp *SudoPrompt) IsSubmitted() bool {
	return sp.submitted
}

// IsCancelled returns true if the user cancelled the prompt
func (sp *SudoPrompt) IsCancelled() bool {
	return sp.cancelled
}

// GetPassword returns the entered password
func (sp *SudoPrompt) GetPassword() string {
	return sp.textInput.Value()
}

// SetError sets an error message
func (sp *SudoPrompt) SetError(err error) {
	sp.err = err
}

// Reset clears the prompt state
func (sp *SudoPrompt) Reset() {
	sp.textInput.SetValue("")
	sp.err = nil
	sp.cancelled = false
	sp.submitted = false
	sp.textInput.Focus()
}

// ClearPassword securely clears the password from memory
func (sp *SudoPrompt) ClearPassword() {
	// Overwrite the password in memory for security
	password := sp.textInput.Value()
	if len(password) > 0 {
		// Create a slice of zeros to overwrite the password
		zeros := make([]byte, len(password))
		copy([]byte(password), zeros)
	}
	sp.textInput.SetValue("")
}