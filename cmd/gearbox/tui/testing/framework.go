package testing

import (
	"context"
	"io"
	"testing"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// TestConfig provides configuration for TUI tests
type TestConfig struct {
	Width          int
	Height         int
	Timeout        time.Duration
	InitialMsg     tea.Msg
	ExpectedOutput string
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() TestConfig {
	return TestConfig{
		Width:   80,
		Height:  24,
		Timeout: 5 * time.Second,
	}
}

// TestFramework provides utilities for TUI testing
type TestFramework struct {
	t      *testing.T
	config TestConfig
}

// NewTestFramework creates a new test framework
func NewTestFramework(t *testing.T, config ...TestConfig) *TestFramework {
	cfg := DefaultTestConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	
	return &TestFramework{
		t:      t,
		config: cfg,
	}
}

// TestModel tests a Bubble Tea model with the given configuration
func (tf *TestFramework) TestModel(model tea.Model) *TestSession {
	return &TestSession{
		framework: tf,
		model:     model,
		tm:        teatest.NewTestModel(tf.t, model, teatest.WithInitialTermSize(tf.config.Width, tf.config.Height)),
	}
}

// TestSession represents a test session for a specific model
type TestSession struct {
	framework *TestFramework
	model     tea.Model
	tm        *teatest.TestModel
}

// SendKey sends a key message to the model
func (ts *TestSession) SendKey(keyType tea.KeyType) *TestSession {
	ts.tm.Send(tea.KeyMsg{Type: keyType})
	return ts
}

// SendKeyString sends a key string message to the model
func (ts *TestSession) SendKeyString(key string) *TestSession {
	ts.tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return ts
}

// SendMessage sends a custom message to the model
func (ts *TestSession) SendMessage(msg tea.Msg) *TestSession {
	ts.tm.Send(msg)
	return ts
}

// Resize sends a window resize message
func (ts *TestSession) Resize(width, height int) *TestSession {
	ts.tm.Send(tea.WindowSizeMsg{Width: width, Height: height})
	return ts
}

// WaitForFinish waits for the test to complete
func (ts *TestSession) WaitForFinish() *TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), ts.framework.config.Timeout)
	defer cancel()
	
	ts.tm.WaitFinished(ts.framework.t, teatest.WithFinalTimeout(ts.framework.config.Timeout))
	
	// Read the output from io.Reader and convert to string
	outputReader := ts.tm.FinalOutput(ts.framework.t)
	outputBytes, err := io.ReadAll(outputReader)
	if err != nil {
		ts.framework.t.Errorf("Failed to read final output: %v", err)
	}
	outputString := string(outputBytes)

	return &TestResult{
		session: ts,
		model:   ts.tm.FinalModel(ts.framework.t),
		output:  outputString,
		context: ctx,
	}
}

// TestResult contains the results of a test session
type TestResult struct {
	session *TestSession
	model   tea.Model
	output  string
	context context.Context
}

// AssertOutput asserts that the output contains the expected string
func (tr *TestResult) AssertOutput(expected string) *TestResult {
	if !contains(tr.output, expected) {
		tr.session.framework.t.Errorf("Expected output to contain %q, but got:\n%s", expected, tr.output)
	}
	return tr
}

// AssertOutputEquals asserts that the output equals the expected string exactly
func (tr *TestResult) AssertOutputEquals(expected string) *TestResult {
	if tr.output != expected {
		tr.session.framework.t.Errorf("Expected output to equal %q, but got:\n%s", expected, tr.output)
	}
	return tr
}

// AssertNoError asserts that the model has no error
func (tr *TestResult) AssertNoError() *TestResult {
	if errorModel, ok := tr.model.(interface{ GetError() error }); ok {
		if err := errorModel.GetError(); err != nil {
			tr.session.framework.t.Errorf("Expected no error, but got: %v", err)
		}
	}
	return tr
}

// GetModel returns the final model
func (tr *TestResult) GetModel() tea.Model {
	return tr.model
}

// GetOutput returns the final output
func (tr *TestResult) GetOutput() string {
	return tr.output
}

// TestSuite provides utilities for running multiple related tests
type TestSuite struct {
	t       *testing.T
	name    string
	config  TestConfig
	results []TestResult
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T, name string, config ...TestConfig) *TestSuite {
	cfg := DefaultTestConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	
	return &TestSuite{
		t:      t,
		name:   name,
		config: cfg,
	}
}

// Run runs a test within the suite
func (ts *TestSuite) Run(testName string, testFunc func(t *testing.T, framework *TestFramework)) *TestSuite {
	ts.t.Run(testName, func(t *testing.T) {
		framework := NewTestFramework(t, ts.config)
		testFunc(t, framework)
	})
	return ts
}

// BenchmarkView benchmarks view rendering performance
func (tf *TestFramework) BenchmarkView(b *testing.B, viewName string, model tea.Model) {
	b.Run(viewName, func(b *testing.B) {
		// Setup
		tm := teatest.NewTestModel(tf.t, model, teatest.WithInitialTermSize(tf.config.Width, tf.config.Height))
		defer tm.WaitFinished(tf.t, teatest.WithFinalTimeout(tf.config.Timeout))
		
		// Benchmark rendering
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.View()
		}
	})
}

// BenchmarkUpdate benchmarks model update performance
func (tf *TestFramework) BenchmarkUpdate(b *testing.B, updateName string, model tea.Model, msg tea.Msg) {
	b.Run(updateName, func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = model.Update(msg)
		}
	})
}

// StressTest performs a stress test with rapid key presses
func (tf *TestFramework) StressTest(model tea.Model, keyCount int, keyDelay time.Duration) *TestResult {
	session := tf.TestModel(model)
	
	// Send rapid key presses
	for i := 0; i < keyCount; i++ {
		session.SendKey(tea.KeyTab)
		if keyDelay > 0 {
			time.Sleep(keyDelay)
		}
	}
	
	return session.WaitForFinish()
}

// MemoryLeakTest checks for memory leaks during repeated operations
func (tf *TestFramework) MemoryLeakTest(model tea.Model, operations int) MemoryTestResult {
	// TODO: Implement memory leak detection
	// This would involve:
	// 1. Recording initial memory usage
	// 2. Performing operations
	// 3. Forcing GC
	// 4. Measuring final memory usage
	// 5. Comparing against baseline
	
	return MemoryTestResult{
		InitialMemory: 0,
		FinalMemory:   0,
		Operations:    operations,
		LeakDetected:  false,
	}
}

// MemoryTestResult contains the results of a memory leak test
type MemoryTestResult struct {
	InitialMemory uint64
	FinalMemory   uint64
	Operations    int
	LeakDetected  bool
}

// ValidationTestSuite runs a comprehensive suite of validation tests
func (tf *TestFramework) ValidationTestSuite(model tea.Model) {
	tf.t.Run("Navigation", func(t *testing.T) {
		// Test navigation between views
		session := tf.TestModel(model)
		result := session.
			SendKeyString("d").  // Dashboard
			SendKeyString("t").  // Tool Browser
			SendKeyString("b").  // Bundle Explorer
			SendKeyString("h").  // Health
			WaitForFinish()
		
		result.AssertNoError()
	})
	
	tf.t.Run("Resize", func(t *testing.T) {
		// Test window resizing
		session := tf.TestModel(model)
		result := session.
			Resize(120, 30).
			Resize(60, 15).
			Resize(80, 24).
			WaitForFinish()
		
		result.AssertNoError()
	})
	
	tf.t.Run("QuickNavigation", func(t *testing.T) {
		// Test rapid navigation
		session := tf.TestModel(model)
		result := session.
			SendKey(tea.KeyTab).
			SendKey(tea.KeyTab).
			SendKey(tea.KeyTab).
			SendKeyString("q").  // Quit
			WaitForFinish()
		
		result.AssertNoError()
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

// Helper function to find substring
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}