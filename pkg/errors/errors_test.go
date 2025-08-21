package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestGearboxError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *GearboxError
		expected string
	}{
		{
			name: "with user message",
			err: &GearboxError{
				Type:        ValidationError,
				Operation:   "validate_input",
				UserMessage: "Invalid tool name provided",
			},
			expected: "Invalid tool name provided",
		},
		{
			name: "with cause error",
			err: &GearboxError{
				Type:      InstallationError,
				Operation: "install_tool",
				Cause:     errors.New("build failed"),
			},
			expected: "install_tool failed: build failed",
		},
		{
			name: "with details only",
			err: &GearboxError{
				Type:      SystemError,
				Operation: "check_system",
				Details:   "insufficient memory",
			},
			expected: "check_system failed: insufficient memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("GearboxError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGearboxError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	gearboxErr := &GearboxError{
		Type:      InstallationError,
		Operation: "install",
		Cause:     originalErr,
	}

	if unwrapped := gearboxErr.Unwrap(); unwrapped != originalErr {
		t.Errorf("GearboxError.Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestGearboxError_String(t *testing.T) {
	err := &GearboxError{
		Type:        ValidationError,
		Operation:   "validate_tool",
		UserMessage: "Invalid tool name",
		Details:     "Tool name contains invalid characters",
		Cause:       errors.New("regex mismatch"),
		Context: map[string]interface{}{
			"tool": "invalid-tool",
			"user": "testuser",
		},
	}

	result := err.String()

	// Check that all components are present in the string representation
	expectedParts := []string{
		"Type: validation",
		"Operation: validate_tool",
		"Message: Invalid tool name",
		"Details: Tool name contains invalid characters",
		"Cause: regex mismatch",
		"Context:",
		"tool=invalid-tool",
		"user=testuser",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("GearboxError.String() missing part: %s\nFull result: %s", part, result)
		}
	}
}

func TestGearboxError_WithContext(t *testing.T) {
	err := New(ValidationError, "test_operation")
	
	// Test adding context
	result := err.WithContext("tool", "fd").WithContext("version", "1.0.0")
	
	if result.Context["tool"] != "fd" {
		t.Errorf("WithContext() failed to set 'tool', got %v", result.Context["tool"])
	}
	
	if result.Context["version"] != "1.0.0" {
		t.Errorf("WithContext() failed to set 'version', got %v", result.Context["version"])
	}
	
	// Test that it returns the same error instance for chaining
	if result != err {
		t.Error("WithContext() should return the same error instance for method chaining")
	}
}

func TestGearboxError_GetSuggestion(t *testing.T) {
	tests := []struct {
		name        string
		err         *GearboxError
		expectedMsg string
	}{
		{
			name: "validation error",
			err:  New(ValidationError, "validate"),
			expectedMsg: "Please check your input parameters and try again.",
		},
		{
			name: "configuration error",
			err:  New(ConfigurationError, "load_config"),
			expectedMsg: "Run 'gearbox config wizard' to fix configuration issues.",
		},
		{
			name: "installation error with tool context",
			err:  New(InstallationError, "install").WithContext("tool", "ripgrep"),
			expectedMsg: "Try running 'gearbox install ripgrep --force' or check system requirements with 'gearbox doctor'.",
		},
		{
			name: "installation error without tool context",
			err:  New(InstallationError, "install"),
			expectedMsg: "Try running 'gearbox doctor' to check system requirements.",
		},
		{
			name: "system error",
			err:  New(SystemError, "check_memory"),
			expectedMsg: "Check system resources and try again. Run 'gearbox doctor system' for diagnostics.",
		},
		{
			name: "network error",
			err:  New(NetworkError, "download"),
			expectedMsg: "Check your internet connection and try again.",
		},
		{
			name: "file error",
			err:  New(FileError, "read_file"),
			expectedMsg: "Check file permissions and available disk space.",
		},
		{
			name: "permission error",
			err:  New(PermissionError, "write_file"),
			expectedMsg: "Check file/directory permissions or run with appropriate privileges.",
		},
		{
			name: "dependency error",
			err:  New(DependencyError, "check_deps"),
			expectedMsg: "Run 'gearbox install --skip-common-deps=false' to install missing dependencies.",
		},
		{
			name: "unknown error type",
			err:  &GearboxError{Type: ErrorType("unknown"), Operation: "test"},
			expectedMsg: "Run 'gearbox doctor' for system diagnostics or check the documentation.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.GetSuggestion(); got != tt.expectedMsg {
				t.Errorf("GearboxError.GetSuggestion() = %v, want %v", got, tt.expectedMsg)
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := New(ValidationError, "test_operation")
	
	if err.Type != ValidationError {
		t.Errorf("New() type = %v, want %v", err.Type, ValidationError)
	}
	
	if err.Operation != "test_operation" {
		t.Errorf("New() operation = %v, want %v", err.Operation, "test_operation")
	}
	
	if err.Context == nil {
		t.Error("New() should initialize Context map")
	}
	
	if err.StackTrace == "" {
		t.Error("New() should capture stack trace")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(originalErr, InstallationError, "install_tool")
	
	if err.Type != InstallationError {
		t.Errorf("Wrap() type = %v, want %v", err.Type, InstallationError)
	}
	
	if err.Operation != "install_tool" {
		t.Errorf("Wrap() operation = %v, want %v", err.Operation, "install_tool")
	}
	
	if err.Cause != originalErr {
		t.Errorf("Wrap() cause = %v, want %v", err.Cause, originalErr)
	}
	
	if err.Context == nil {
		t.Error("Wrap() should initialize Context map")
	}
}

func TestGearboxError_WithMessage(t *testing.T) {
	err := New(ValidationError, "test")
	message := "Test user message"
	
	result := err.WithMessage(message)
	
	if result.UserMessage != message {
		t.Errorf("WithMessage() = %v, want %v", result.UserMessage, message)
	}
	
	// Should return same instance for chaining
	if result != err {
		t.Error("WithMessage() should return the same error instance")
	}
}

func TestGearboxError_WithDetails(t *testing.T) {
	err := New(ValidationError, "test")
	details := "Test technical details"
	
	result := err.WithDetails(details)
	
	if result.Details != details {
		t.Errorf("WithDetails() = %v, want %v", result.Details, details)
	}
	
	// Should return same instance for chaining
	if result != err {
		t.Error("WithDetails() should return the same error instance")
	}
}

func TestIsType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errType  ErrorType
		expected bool
	}{
		{
			name:     "matching gearbox error type",
			err:      New(ValidationError, "test"),
			errType:  ValidationError,
			expected: true,
		},
		{
			name:     "non-matching gearbox error type",
			err:      New(ValidationError, "test"),
			errType:  InstallationError,
			expected: false,
		},
		{
			name:     "non-gearbox error",
			err:      errors.New("standard error"),
			errType:  ValidationError,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			errType:  ValidationError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsType(tt.err, tt.errType); got != tt.expected {
				t.Errorf("IsType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	operation := "validate_input"
	message := "Invalid input provided"
	
	err := NewValidationError(operation, message)
	
	if err.Type != ValidationError {
		t.Errorf("NewValidationError() type = %v, want %v", err.Type, ValidationError)
	}
	
	if err.Operation != operation {
		t.Errorf("NewValidationError() operation = %v, want %v", err.Operation, operation)
	}
	
	if err.UserMessage != message {
		t.Errorf("NewValidationError() message = %v, want %v", err.UserMessage, message)
	}
}

func TestNewInstallationError(t *testing.T) {
	operation := "install_tool"
	tool := "ripgrep"
	cause := errors.New("build failed")
	
	err := NewInstallationError(operation, tool, cause)
	
	if err.Type != InstallationError {
		t.Errorf("NewInstallationError() type = %v, want %v", err.Type, InstallationError)
	}
	
	if err.Operation != operation {
		t.Errorf("NewInstallationError() operation = %v, want %v", err.Operation, operation)
	}
	
	if err.Cause != cause {
		t.Errorf("NewInstallationError() cause = %v, want %v", err.Cause, cause)
	}
	
	if err.Context["tool"] != tool {
		t.Errorf("NewInstallationError() tool context = %v, want %v", err.Context["tool"], tool)
	}
	
	expectedMessage := "Failed to install " + tool
	if err.UserMessage != expectedMessage {
		t.Errorf("NewInstallationError() message = %v, want %v", err.UserMessage, expectedMessage)
	}
}

func TestNewConfigurationError(t *testing.T) {
	operation := "load_config"
	details := "Invalid JSON format"
	
	err := NewConfigurationError(operation, details)
	
	if err.Type != ConfigurationError {
		t.Errorf("NewConfigurationError() type = %v, want %v", err.Type, ConfigurationError)
	}
	
	if err.Operation != operation {
		t.Errorf("NewConfigurationError() operation = %v, want %v", err.Operation, operation)
	}
	
	if err.Details != details {
		t.Errorf("NewConfigurationError() details = %v, want %v", err.Details, details)
	}
	
	if err.UserMessage != "Configuration error detected" {
		t.Errorf("NewConfigurationError() message = %v, want %v", err.UserMessage, "Configuration error detected")
	}
}

func TestNewSystemError(t *testing.T) {
	operation := "check_memory"
	cause := errors.New("insufficient memory")
	
	err := NewSystemError(operation, cause)
	
	if err.Type != SystemError {
		t.Errorf("NewSystemError() type = %v, want %v", err.Type, SystemError)
	}
	
	if err.Operation != operation {
		t.Errorf("NewSystemError() operation = %v, want %v", err.Operation, operation)
	}
	
	if err.Cause != cause {
		t.Errorf("NewSystemError() cause = %v, want %v", err.Cause, cause)
	}
	
	if err.UserMessage != "System error occurred" {
		t.Errorf("NewSystemError() message = %v, want %v", err.UserMessage, "System error occurred")
	}
}

func TestNewNetworkError(t *testing.T) {
	operation := "download_file"
	cause := errors.New("connection timeout")
	
	err := NewNetworkError(operation, cause)
	
	if err.Type != NetworkError {
		t.Errorf("NewNetworkError() type = %v, want %v", err.Type, NetworkError)
	}
	
	if err.Cause != cause {
		t.Errorf("NewNetworkError() cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewFileError(t *testing.T) {
	operation := "read_file"
	filePath := "/path/to/file"
	cause := errors.New("file not found")
	
	err := NewFileError(operation, filePath, cause)
	
	if err.Type != FileError {
		t.Errorf("NewFileError() type = %v, want %v", err.Type, FileError)
	}
	
	if err.Context["file"] != filePath {
		t.Errorf("NewFileError() file context = %v, want %v", err.Context["file"], filePath)
	}
	
	expectedMessage := "File operation failed for " + filePath
	if err.UserMessage != expectedMessage {
		t.Errorf("NewFileError() message = %v, want %v", err.UserMessage, expectedMessage)
	}
}

func TestNewPermissionError(t *testing.T) {
	operation := "write_file"
	resource := "/etc/config"
	
	err := NewPermissionError(operation, resource)
	
	if err.Type != PermissionError {
		t.Errorf("NewPermissionError() type = %v, want %v", err.Type, PermissionError)
	}
	
	if err.Context["resource"] != resource {
		t.Errorf("NewPermissionError() resource context = %v, want %v", err.Context["resource"], resource)
	}
	
	expectedMessage := "Permission denied for " + resource
	if err.UserMessage != expectedMessage {
		t.Errorf("NewPermissionError() message = %v, want %v", err.UserMessage, expectedMessage)
	}
}

func TestNewDependencyError(t *testing.T) {
	operation := "check_dependencies"
	dependency := "rust"
	cause := errors.New("not found in PATH")
	
	err := NewDependencyError(operation, dependency, cause)
	
	if err.Type != DependencyError {
		t.Errorf("NewDependencyError() type = %v, want %v", err.Type, DependencyError)
	}
	
	if err.Context["dependency"] != dependency {
		t.Errorf("NewDependencyError() dependency context = %v, want %v", err.Context["dependency"], dependency)
	}
	
	expectedMessage := "Missing or invalid dependency: " + dependency
	if err.UserMessage != expectedMessage {
		t.Errorf("NewDependencyError() message = %v, want %v", err.UserMessage, expectedMessage)
	}
}

func TestErrorChaining(t *testing.T) {
	// Test method chaining works properly
	err := New(ValidationError, "test").
		WithMessage("Test message").
		WithDetails("Test details").
		WithContext("key1", "value1").
		WithContext("key2", "value2")
	
	if err.UserMessage != "Test message" {
		t.Error("Method chaining failed for WithMessage")
	}
	
	if err.Details != "Test details" {
		t.Error("Method chaining failed for WithDetails")
	}
	
	if err.Context["key1"] != "value1" || err.Context["key2"] != "value2" {
		t.Error("Method chaining failed for WithContext")
	}
}

func TestStackTraceCapture(t *testing.T) {
	err := New(ValidationError, "test")
	
	if err.StackTrace == "" {
		t.Error("Stack trace should be captured")
	}
	
	// Stack trace should contain gearbox-related files
	if !strings.Contains(err.StackTrace, "gearbox") {
		t.Error("Stack trace should contain gearbox-related files")
	}
}

// Benchmark tests for performance
func BenchmarkNewGearboxError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(ValidationError, "benchmark_test")
	}
}

func BenchmarkErrorWithContext(b *testing.B) {
	err := New(ValidationError, "benchmark_test")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = err.WithContext("iteration", i)
	}
}

func BenchmarkGetSuggestion(b *testing.B) {
	err := New(InstallationError, "benchmark_test").WithContext("tool", "ripgrep")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = err.GetSuggestion()
	}
}