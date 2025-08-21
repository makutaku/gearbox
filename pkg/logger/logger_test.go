package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.Level != InfoLevel {
		t.Errorf("DefaultConfig() Level = %v, want %v", config.Level, InfoLevel)
	}
	
	if !config.Pretty {
		t.Error("DefaultConfig() Pretty should be true")
	}
	
	if config.TimeFormat != time.RFC3339 {
		t.Errorf("DefaultConfig() TimeFormat = %v, want %v", config.TimeFormat, time.RFC3339)
	}
	
	if config.Output != os.Stdout {
		t.Error("DefaultConfig() Output should be os.Stdout")
	}
}

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      DebugLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	if logger == nil {
		t.Fatal("New() should return a logger instance")
	}
	
	// Test that the logger was configured correctly by logging a message
	logger.Info("test message")
	
	// Check that output was written
	if buf.Len() == 0 {
		t.Error("Logger should have written output")
	}
	
	// Parse JSON output (since Pretty=false)
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON, got: %s", buf.String())
	}
	
	// Check log entry contents
	if logEntry["level"] != "info" {
		t.Errorf("Log entry level = %v, want %v", logEntry["level"], "info")
	}
	
	if logEntry["message"] != "test message" {
		t.Errorf("Log entry message = %v, want %v", logEntry["message"], "test message")
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	
	if logger == nil {
		t.Fatal("NewDefault() should return a logger instance")
	}
}

func TestNewQuiet(t *testing.T) {
	var buf bytes.Buffer
	
	// Temporarily replace global zerolog level for testing
	oldLevel := zerolog.GlobalLevel()
	defer zerolog.SetGlobalLevel(oldLevel)
	
	logger := NewQuiet()
	logger.logger = logger.logger.Output(&buf)
	
	// Info message should not appear (quiet mode = error level only)
	logger.Info("info message")
	if buf.Len() > 0 {
		t.Error("Quiet logger should not log info messages")
	}
	
	// Error message should appear
	buf.Reset()
	logger.Error("error message")
	if buf.Len() == 0 {
		t.Error("Quiet logger should log error messages")
	}
}

func TestNewVerbose(t *testing.T) {
	var buf bytes.Buffer
	
	// Create verbose logger with custom output
	config := DefaultConfig()
	config.Level = DebugLevel
	config.Pretty = false
	config.Output = &buf
	
	logger := New(config)
	
	// Debug message should appear in verbose mode
	logger.Debug("debug message")
	if buf.Len() == 0 {
		t.Error("Verbose logger should log debug messages")
	}
	
	// Parse and verify debug message
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
	}
	
	if logEntry["level"] != "debug" {
		t.Errorf("Log entry level = %v, want %v", logEntry["level"], "debug")
	}
}

func TestLoggingMethods(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      DebugLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	tests := []struct {
		name     string
		logFunc  func()
		level    string
		message  string
	}{
		{
			name:    "Debug",
			logFunc: func() { logger.Debug("debug test") },
			level:   "debug",
			message: "debug test",
		},
		{
			name:    "Info",
			logFunc: func() { logger.Info("info test") },
			level:   "info",
			message: "info test",
		},
		{
			name:    "Warn",
			logFunc: func() { logger.Warn("warn test") },
			level:   "warn",
			message: "warn test",
		},
		{
			name:    "Error",
			logFunc: func() { logger.Error("error test") },
			level:   "error",
			message: "error test",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			
			if buf.Len() == 0 {
				t.Errorf("%s should produce output", tt.name)
				return
			}
			
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Errorf("Logger output should be valid JSON: %v", err)
				return
			}
			
			if logEntry["level"] != tt.level {
				t.Errorf("%s level = %v, want %v", tt.name, logEntry["level"], tt.level)
			}
			
			if logEntry["message"] != tt.message {
				t.Errorf("%s message = %v, want %v", tt.name, logEntry["message"], tt.message)
			}
		})
	}
}

func TestFormattedLoggingMethods(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      DebugLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	tests := []struct {
		name     string
		logFunc  func()
		level    string
		expected string
	}{
		{
			name:     "Debugf",
			logFunc:  func() { logger.Debugf("debug %s %d", "test", 123) },
			level:    "debug",
			expected: "debug test 123",
		},
		{
			name:     "Infof",
			logFunc:  func() { logger.Infof("info %s %d", "test", 456) },
			level:    "info",
			expected: "info test 456",
		},
		{
			name:     "Warnf",
			logFunc:  func() { logger.Warnf("warn %s %d", "test", 789) },
			level:    "warn",
			expected: "warn test 789",
		},
		{
			name:     "Errorf",
			logFunc:  func() { logger.Errorf("error %s %d", "test", 999) },
			level:    "error",
			expected: "error test 999",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Errorf("Logger output should be valid JSON: %v", err)
				return
			}
			
			if logEntry["level"] != tt.level {
				t.Errorf("%s level = %v, want %v", tt.name, logEntry["level"], tt.level)
			}
			
			if logEntry["message"] != tt.expected {
				t.Errorf("%s message = %v, want %v", tt.name, logEntry["message"], tt.expected)
			}
		})
	}
}

func TestErrorWithErr(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      ErrorLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	testErr := &customError{msg: "test error"}
	logger.ErrorWithErr(testErr, "operation failed")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["level"] != "error" {
		t.Errorf("ErrorWithErr level = %v, want %v", logEntry["level"], "error")
	}
	
	if logEntry["message"] != "operation failed" {
		t.Errorf("ErrorWithErr message = %v, want %v", logEntry["message"], "operation failed")
	}
	
	if logEntry["error"] != "test error" {
		t.Errorf("ErrorWithErr error = %v, want %v", logEntry["error"], "test error")
	}
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	fieldLogger := logger.WithField("key", "value")
	fieldLogger.Info("test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["key"] != "value" {
		t.Errorf("WithField key = %v, want %v", logEntry["key"], "value")
	}
	
	if logEntry["message"] != "test message" {
		t.Errorf("WithField message = %v, want %v", logEntry["message"], "test message")
	}
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	
	fieldsLogger := logger.WithFields(fields)
	fieldsLogger.Info("test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["key1"] != "value1" {
		t.Errorf("WithFields key1 = %v, want %v", logEntry["key1"], "value1")
	}
	
	if logEntry["key2"] != float64(42) { // JSON unmarshals numbers as float64
		t.Errorf("WithFields key2 = %v, want %v", logEntry["key2"], 42)
	}
	
	if logEntry["key3"] != true {
		t.Errorf("WithFields key3 = %v, want %v", logEntry["key3"], true)
	}
}

func TestWithError(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	testErr := &customError{msg: "test error"}
	errorLogger := logger.WithError(testErr)
	errorLogger.Info("test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["error"] != "test error" {
		t.Errorf("WithError error = %v, want %v", logEntry["error"], "test error")
	}
}

func TestProgress(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	logger.Progress("installing tools", 3, 10)
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["operation"] != "installing tools" {
		t.Errorf("Progress operation = %v, want %v", logEntry["operation"], "installing tools")
	}
	
	if logEntry["current"] != float64(3) {
		t.Errorf("Progress current = %v, want %v", logEntry["current"], 3)
	}
	
	if logEntry["total"] != float64(10) {
		t.Errorf("Progress total = %v, want %v", logEntry["total"], 10)
	}
	
	if logEntry["percentage"] != 30.0 {
		t.Errorf("Progress percentage = %v, want %v", logEntry["percentage"], 30.0)
	}
	
	expectedMessage := "Progress: installing tools (3/10, 30.0%)"
	if logEntry["message"] != expectedMessage {
		t.Errorf("Progress message = %v, want %v", logEntry["message"], expectedMessage)
	}
}

func TestOperation(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	opLogger := logger.Operation("install_tool")
	opLogger.Info("tool installed")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["operation"] != "install_tool" {
		t.Errorf("Operation = %v, want %v", logEntry["operation"], "install_tool")
	}
}

func TestTool(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	toolLogger := logger.Tool("ripgrep")
	toolLogger.Info("tool status")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["tool"] != "ripgrep" {
		t.Errorf("Tool = %v, want %v", logEntry["tool"], "ripgrep")
	}
}

func TestDuration(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	
	duration := 5 * time.Second
	logger.Duration("install_ripgrep", duration)
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["operation"] != "install_ripgrep" {
		t.Errorf("Duration operation = %v, want %v", logEntry["operation"], "install_ripgrep")
	}
	
	if logEntry["duration"] != float64(5000) { // Duration in milliseconds
		t.Errorf("Duration = %v, want %v", logEntry["duration"], 5000)
	}
	
	expectedMessage := "Operation completed: install_ripgrep (took 5s)"
	if logEntry["message"] != expectedMessage {
		t.Errorf("Duration message = %v, want %v", logEntry["message"], expectedMessage)
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test global logger initialization
	globalLogger := GetGlobalLogger()
	if globalLogger == nil {
		t.Error("Global logger should be initialized")
	}
	
	// Test setting global logger
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	newLogger := New(config)
	SetGlobalLogger(newLogger)
	
	retrievedLogger := GetGlobalLogger()
	if retrievedLogger != newLogger {
		t.Error("SetGlobalLogger/GetGlobalLogger should work correctly")
	}
}

func TestGlobalLoggingFunctions(t *testing.T) {
	// Set up a test logger for global functions
	var buf bytes.Buffer
	config := Config{
		Level:      DebugLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	testLogger := New(config)
	SetGlobalLogger(testLogger)
	
	tests := []struct {
		name     string
		logFunc  func()
		level    string
		message  string
	}{
		{
			name:    "Global Debug",
			logFunc: func() { Debug("global debug") },
			level:   "debug",
			message: "global debug",
		},
		{
			name:    "Global Info",
			logFunc: func() { Info("global info") },
			level:   "info",
			message: "global info",
		},
		{
			name:    "Global Warn",
			logFunc: func() { Warn("global warn") },
			level:   "warn",
			message: "global warn",
		},
		{
			name:    "Global Error",
			logFunc: func() { Error("global error") },
			level:   "error",
			message: "global error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			
			if buf.Len() == 0 {
				t.Errorf("%s should produce output", tt.name)
				return
			}
			
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Errorf("Logger output should be valid JSON: %v", err)
				return
			}
			
			if logEntry["level"] != tt.level {
				t.Errorf("%s level = %v, want %v", tt.name, logEntry["level"], tt.level)
			}
			
			if logEntry["message"] != tt.message {
				t.Errorf("%s message = %v, want %v", tt.name, logEntry["message"], tt.message)
			}
		})
	}
}

func TestGlobalFormattedLoggingFunctions(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      DebugLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	testLogger := New(config)
	SetGlobalLogger(testLogger)
	
	tests := []struct {
		name     string
		logFunc  func()
		level    string
		expected string
	}{
		{
			name:     "Global Debugf",
			logFunc:  func() { Debugf("global debug %s", "formatted") },
			level:    "debug",
			expected: "global debug formatted",
		},
		{
			name:     "Global Infof",
			logFunc:  func() { Infof("global info %d", 42) },
			level:    "info",
			expected: "global info 42",
		},
		{
			name:     "Global Warnf",
			logFunc:  func() { Warnf("global warn %v", true) },
			level:    "warn",
			expected: "global warn true",
		},
		{
			name:     "Global Errorf",
			logFunc:  func() { Errorf("global error %s", "test") },
			level:    "error",
			expected: "global error test",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Errorf("Logger output should be valid JSON: %v", err)
				return
			}
			
			if logEntry["level"] != tt.level {
				t.Errorf("%s level = %v, want %v", tt.name, logEntry["level"], tt.level)
			}
			
			if logEntry["message"] != tt.expected {
				t.Errorf("%s message = %v, want %v", tt.name, logEntry["message"], tt.expected)
			}
		})
	}
}

func TestGlobalErrorWithErr(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      ErrorLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	testLogger := New(config)
	SetGlobalLogger(testLogger)
	
	testErr := &customError{msg: "global test error"}
	ErrorWithErr(testErr, "global operation failed")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("Logger output should be valid JSON: %v", err)
		return
	}
	
	if logEntry["level"] != "error" {
		t.Errorf("Global ErrorWithErr level = %v, want %v", logEntry["level"], "error")
	}
	
	if logEntry["error"] != "global test error" {
		t.Errorf("Global ErrorWithErr error = %v, want %v", logEntry["error"], "global test error")
	}
}

func TestPrettyFormatting(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     true,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	logger.Info("test message")
	
	output := buf.String()
	
	// Pretty output should contain emoji indicators
	if !strings.Contains(output, "📋 INFO") {
		t.Error("Pretty formatting should include emoji level indicators")
	}
	
	if !strings.Contains(output, "test message") {
		t.Error("Pretty formatting should include the message")
	}
	
	// Should not be JSON format when pretty=true
	var logEntry map[string]interface{}
	if json.Unmarshal([]byte(output), &logEntry) == nil {
		t.Error("Pretty formatted output should not be JSON")
	}
}

// Benchmark tests
func BenchmarkLoggerInfo(b *testing.B) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message")
		buf.Reset()
	}
}

func BenchmarkLoggerInfof(b *testing.B) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logger.Infof("benchmark test message %d", i)
		buf.Reset()
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	var buf bytes.Buffer
	config := Config{
		Level:      InfoLevel,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     &buf,
	}
	
	logger := New(config)
	fields := map[string]interface{}{
		"operation": "benchmark",
		"iteration": 0,
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		fields["iteration"] = i
		fieldsLogger := logger.WithFields(fields)
		fieldsLogger.Info("benchmark test")
		buf.Reset()
	}
}