// Package errors provides structured error types with context for gearbox.
package errors

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// ErrorType represents different categories of errors.
type ErrorType string

const (
	// ValidationError indicates invalid input or configuration
	ValidationError ErrorType = "validation"
	// ConfigurationError indicates configuration-related issues
	ConfigurationError ErrorType = "configuration"
	// InstallationError indicates tool installation failures
	InstallationError ErrorType = "installation"
	// SystemError indicates system-level issues
	SystemError ErrorType = "system"
	// NetworkError indicates network-related issues
	NetworkError ErrorType = "network"
	// FileError indicates file system issues
	FileError ErrorType = "file"
	// PermissionError indicates permission-related issues
	PermissionError ErrorType = "permission"
	// DependencyError indicates dependency-related issues
	DependencyError ErrorType = "dependency"
)

// GearboxError represents a structured error with context and user-friendly messages.
type GearboxError struct {
	Type        ErrorType
	Operation   string
	UserMessage string
	Details     string
	Cause       error
	StackTrace  string
	Context     map[string]interface{}
}

// Error implements the error interface.
func (e *GearboxError) Error() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s failed: %v", e.Operation, e.Cause)
	}
	return fmt.Sprintf("%s failed: %s", e.Operation, e.Details)
}

// Unwrap returns the underlying cause error.
func (e *GearboxError) Unwrap() error {
	return e.Cause
}

// String returns a detailed string representation for debugging.
func (e *GearboxError) String() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("Type: %s", e.Type))
	parts = append(parts, fmt.Sprintf("Operation: %s", e.Operation))
	
	if e.UserMessage != "" {
		parts = append(parts, fmt.Sprintf("Message: %s", e.UserMessage))
	}
	
	if e.Details != "" {
		parts = append(parts, fmt.Sprintf("Details: %s", e.Details))
	}
	
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Cause: %v", e.Cause))
	}
	
	if len(e.Context) > 0 {
		contextParts := make([]string, 0, len(e.Context))
		for k, v := range e.Context {
			contextParts = append(contextParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("Context: {%s}", strings.Join(contextParts, ", ")))
	}
	
	return strings.Join(parts, "; ")
}

// WithContext adds context key-value pairs to the error.
func (e *GearboxError) WithContext(key string, value interface{}) *GearboxError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetSuggestion returns a user-friendly suggestion based on the error type and context.
func (e *GearboxError) GetSuggestion() string {
	switch e.Type {
	case ValidationError:
		return "Please check your input parameters and try again."
	case ConfigurationError:
		return "Run 'gearbox config wizard' to fix configuration issues."
	case InstallationError:
		if tool, ok := e.Context["tool"].(string); ok {
			return fmt.Sprintf("Try running 'gearbox install %s --force' or check system requirements with 'gearbox doctor'.", tool)
		}
		return "Try running 'gearbox doctor' to check system requirements."
	case SystemError:
		return "Check system resources and try again. Run 'gearbox doctor system' for diagnostics."
	case NetworkError:
		return "Check your internet connection and try again."
	case FileError:
		return "Check file permissions and available disk space."
	case PermissionError:
		return "Check file/directory permissions or run with appropriate privileges."
	case DependencyError:
		return "Run 'gearbox install --skip-common-deps=false' to install missing dependencies."
	default:
		return "Run 'gearbox doctor' for system diagnostics or check the documentation."
	}
}

// New creates a new GearboxError with the specified type and operation.
func New(errorType ErrorType, operation string) *GearboxError {
	return &GearboxError{
		Type:       errorType,
		Operation:  operation,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, errorType ErrorType, operation string) *GearboxError {
	return &GearboxError{
		Type:       errorType,
		Operation:  operation,
		Cause:      err,
		StackTrace: getStackTrace(),
		Context:    make(map[string]interface{}),
	}
}

// WithMessage sets a user-friendly message for the error.
func (e *GearboxError) WithMessage(message string) *GearboxError {
	e.UserMessage = message
	return e
}

// WithDetails sets technical details for the error.
func (e *GearboxError) WithDetails(details string) *GearboxError {
	e.Details = details
	return e
}

// IsType checks if an error is of a specific GearboxError type.
func IsType(err error, errorType ErrorType) bool {
	var gearboxErr *GearboxError
	if errors.As(err, &gearboxErr) {
		return gearboxErr.Type == errorType
	}
	return false
}

// getStackTrace captures the current stack trace.
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	
	frames := runtime.CallersFrames(pcs[:n])
	var traces []string
	
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "gearbox") {
			if !more {
				break
			}
			continue
		}
		
		traces = append(traces, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		
		if !more {
			break
		}
	}
	
	return strings.Join(traces, "\n")
}

// Common error constructors for frequently used error types

// NewValidationError creates a validation error.
func NewValidationError(operation string, message string) *GearboxError {
	return New(ValidationError, operation).WithMessage(message)
}

// NewInstallationError creates an installation error.
func NewInstallationError(operation string, tool string, cause error) *GearboxError {
	return Wrap(cause, InstallationError, operation).
		WithContext("tool", tool).
		WithMessage(fmt.Sprintf("Failed to install %s", tool))
}

// NewConfigurationError creates a configuration error.
func NewConfigurationError(operation string, details string) *GearboxError {
	return New(ConfigurationError, operation).
		WithDetails(details).
		WithMessage("Configuration error detected")
}

// NewSystemError creates a system error.
func NewSystemError(operation string, cause error) *GearboxError {
	return Wrap(cause, SystemError, operation).
		WithMessage("System error occurred")
}

// NewNetworkError creates a network error.
func NewNetworkError(operation string, cause error) *GearboxError {
	return Wrap(cause, NetworkError, operation).
		WithMessage("Network error occurred")
}

// NewFileError creates a file system error.
func NewFileError(operation string, filePath string, cause error) *GearboxError {
	return Wrap(cause, FileError, operation).
		WithContext("file", filePath).
		WithMessage(fmt.Sprintf("File operation failed for %s", filePath))
}

// NewPermissionError creates a permission error.
func NewPermissionError(operation string, resource string) *GearboxError {
	return New(PermissionError, operation).
		WithContext("resource", resource).
		WithMessage(fmt.Sprintf("Permission denied for %s", resource))
}

// NewDependencyError creates a dependency error.
func NewDependencyError(operation string, dependency string, cause error) *GearboxError {
	return Wrap(cause, DependencyError, operation).
		WithContext("dependency", dependency).
		WithMessage(fmt.Sprintf("Missing or invalid dependency: %s", dependency))
}