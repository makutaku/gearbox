package error

import (
	"fmt"
	"strings"
	"time"
	
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeConfiguration
	ErrorTypeInstallation
	ErrorTypeNetwork
	ErrorTypeSystem
	ErrorTypeValidation
	ErrorTypePermission
)

// AppError represents a structured application error
type AppError struct {
	Type        ErrorType
	Message     string
	Details     []string
	Context     map[string]interface{}
	Timestamp   time.Time
	Recoverable bool
	Cause       error
}

// Error implements the error interface
func (e AppError) Error() string {
	return e.Message
}

// ErrorManager manages application errors
type ErrorManager struct {
	errors []AppError
	max    int
	styles ErrorStyles
}

// ErrorStyles defines styling for different error types
type ErrorStyles struct {
	ErrorStyle     lipgloss.Style
	WarningStyle   lipgloss.Style
	InfoStyle      lipgloss.Style
	DetailStyle    lipgloss.Style
	TimestampStyle lipgloss.Style
}

// NewErrorManager creates a new error manager
func NewErrorManager(maxErrors int) *ErrorManager {
	return &ErrorManager{
		errors: make([]AppError, 0, maxErrors),
		max:    maxErrors,
		styles: ErrorStyles{
			ErrorStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")), // Red
			WarningStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")), // Yellow
			InfoStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")), // Blue
			DetailStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")), // Gray
			TimestampStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")), // Light Gray
		},
	}
}

// Handle processes an error and returns appropriate commands
func (em *ErrorManager) Handle(err error, context ...interface{}) tea.Cmd {
	appErr := em.categorizeError(err, context...)
	em.addError(appErr)
	
	if appErr.Recoverable {
		return tea.Printf("Warning: %s", appErr.Message)
	}
	
	// For non-recoverable errors, we might want to show an error view
	// but not quit immediately to allow user to see the error
	return nil
}

// categorizeError converts a generic error into a structured AppError
func (em *ErrorManager) categorizeError(err error, context ...interface{}) AppError {
	message := err.Error()
	errorType := em.inferErrorType(message)
	
	contextMap := make(map[string]interface{})
	for i, ctx := range context {
		contextMap[fmt.Sprintf("context_%d", i)] = ctx
	}
	
	return AppError{
		Type:        errorType,
		Message:     message,
		Details:     em.extractDetails(message),
		Context:     contextMap,
		Timestamp:   time.Now(),
		Recoverable: em.isRecoverable(errorType, message),
		Cause:       err,
	}
}

// inferErrorType attempts to categorize the error based on its message
func (em *ErrorManager) inferErrorType(message string) ErrorType {
	msgLower := strings.ToLower(message)
	
	switch {
	case strings.Contains(msgLower, "config"):
		return ErrorTypeConfiguration
	case strings.Contains(msgLower, "install") || strings.Contains(msgLower, "build"):
		return ErrorTypeInstallation
	case strings.Contains(msgLower, "network") || strings.Contains(msgLower, "connection"):
		return ErrorTypeNetwork
	case strings.Contains(msgLower, "system") || strings.Contains(msgLower, "os"):
		return ErrorTypeSystem
	case strings.Contains(msgLower, "validation") || strings.Contains(msgLower, "invalid"):
		return ErrorTypeValidation
	case strings.Contains(msgLower, "permission") || strings.Contains(msgLower, "access"):
		return ErrorTypePermission
	default:
		return ErrorTypeUnknown
	}
}

// extractDetails extracts additional details from error message
func (em *ErrorManager) extractDetails(message string) []string {
	var details []string
	
	// Split on common error delimiters
	if strings.Contains(message, ":") {
		parts := strings.SplitN(message, ":", 2)
		if len(parts) > 1 {
			details = append(details, strings.TrimSpace(parts[1]))
		}
	}
	
	return details
}

// isRecoverable determines if an error is recoverable
func (em *ErrorManager) isRecoverable(errorType ErrorType, message string) bool {
	switch errorType {
	case ErrorTypeConfiguration, ErrorTypeValidation:
		return true // User can fix configuration
	case ErrorTypeNetwork:
		return true // Network issues might be temporary
	case ErrorTypePermission:
		return false // Usually requires intervention
	case ErrorTypeSystem:
		return false // System errors are usually fatal
	default:
		return strings.Contains(strings.ToLower(message), "warning")
	}
}

// addError adds an error to the manager
func (em *ErrorManager) addError(appErr AppError) {
	em.errors = append(em.errors, appErr)
	
	// Keep only the most recent errors
	if len(em.errors) > em.max {
		em.errors = em.errors[len(em.errors)-em.max:]
	}
}

// GetRecentErrors returns the most recent errors
func (em *ErrorManager) GetRecentErrors(count int) []AppError {
	if count <= 0 || count > len(em.errors) {
		count = len(em.errors)
	}
	
	start := len(em.errors) - count
	return em.errors[start:]
}

// GetErrorsByType returns errors filtered by type
func (em *ErrorManager) GetErrorsByType(errorType ErrorType) []AppError {
	var filtered []AppError
	for _, err := range em.errors {
		if err.Type == errorType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// RenderError renders an error with appropriate styling
func (em *ErrorManager) RenderError(appErr AppError) string {
	var style lipgloss.Style
	
	switch appErr.Type {
	case ErrorTypeConfiguration, ErrorTypeValidation:
		style = em.styles.WarningStyle
	case ErrorTypeNetwork:
		style = em.styles.InfoStyle
	default:
		style = em.styles.ErrorStyle
	}
	
	timestamp := em.styles.TimestampStyle.Render(appErr.Timestamp.Format("15:04:05"))
	message := style.Render(appErr.Message)
	
	result := fmt.Sprintf("%s %s", timestamp, message)
	
	// Add details if available
	for _, detail := range appErr.Details {
		result += "\n  " + em.styles.DetailStyle.Render("└ "+detail)
	}
	
	return result
}

// Clear clears all errors
func (em *ErrorManager) Clear() {
	em.errors = make([]AppError, 0, em.max)
}

// HasErrors returns true if there are any errors
func (em *ErrorManager) HasErrors() bool {
	return len(em.errors) > 0
}

// HasCriticalErrors returns true if there are any non-recoverable errors
func (em *ErrorManager) HasCriticalErrors() bool {
	for _, err := range em.errors {
		if !err.Recoverable {
			return true
		}
	}
	return false
}