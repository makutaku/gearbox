package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApp_HealthViewAutomaticTrigger(t *testing.T) {
	// Create a new model
	model, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize the model
	model.ready = true
	model.state.Initialized = true

	// Simulate pressing 'h' key to switch to health view
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'h'},
	}

	// Process the key press
	newModel, cmd := model.handleKeyPress(keyMsg)
	if m, ok := newModel.(Model); ok {
		*model = m
	} else {
		t.Fatalf("Expected Model, got %T", newModel)
	}

	// Should switch to health view
	if model.state.CurrentView != ViewHealth {
		t.Errorf("Expected current view to be ViewHealth, got %v", model.state.CurrentView)
	}

	// Should return a command to trigger health checks
	if cmd == nil {
		t.Error("Expected command to trigger health checks, got nil")
	}

	// Execute the command to see if it returns healthCheckTriggerMsg
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(healthCheckTriggerMsg); !ok {
			t.Errorf("Expected healthCheckTriggerMsg, got %T", msg)
		}
	}
}

func TestApp_HealthCheckTriggerHandling(t *testing.T) {
	// Create a new model
	model, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize and set to health view
	model.ready = true
	model.state.Initialized = true
	model.state.CurrentView = ViewHealth

	// Send healthCheckTriggerMsg
	triggerMsg := healthCheckTriggerMsg{}
	newModel, cmd := model.Update(triggerMsg)
	if m, ok := newModel.(Model); ok {
		*model = m
	} else {
		t.Fatalf("Expected Model, got %T", newModel)
	}

	// Should generate a command to trigger 'r' key in health view
	if cmd == nil {
		t.Error("Expected command to trigger health refresh, got nil")
	}
}

func TestApp_ZeroLatencyHealthViewSwitch(t *testing.T) {
	// Create a new model
	model, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize the model
	model.ready = true
	model.state.Initialized = true

	// Measure time to switch to health view
	start := time.Now()

	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'h'},
	}

	newModel, _ := model.handleKeyPress(keyMsg)
	if m, ok := newModel.(Model); ok {
		*model = m
	} else {
		t.Fatalf("Expected Model, got %T", newModel)
	}

	elapsed := time.Since(start)

	// Should switch view instantly (under 1ms)
	if elapsed > time.Millisecond {
		t.Errorf("Health view switch took too long: %v (should be < 1ms)", elapsed)
	}

	// Should switch to health view
	if model.state.CurrentView != ViewHealth {
		t.Errorf("Expected current view to be ViewHealth, got %v", model.state.CurrentView)
	}
}

func TestApp_HealthViewRenderPerformance(t *testing.T) {
	// Create a new model
	model, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize and set to health view
	model.ready = true
	model.state.Initialized = true
	model.state.CurrentView = ViewHealth
	model.width = 80
	model.height = 24

	// Measure render time
	start := time.Now()
	content := model.renderCurrentView()
	elapsed := time.Since(start)

	// Should render quickly (under 5ms)
	if elapsed > 5*time.Millisecond {
		t.Errorf("Health view render took too long: %v (should be < 5ms)", elapsed)
	}

	// Should have content
	if content == "" {
		t.Error("Health view rendered empty content")
	}

	// Should contain health monitor elements
	if !containsString(content, "Health") {
		t.Error("Health view should contain 'Health' in rendered content")
	}
}

func TestApp_NavigationResponsiveness(t *testing.T) {
	// Create a new model
	model, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize
	model.ready = true
	model.state.Initialized = true

	// Test all navigation keys for responsiveness
	testKeys := []struct {
		key      rune
		expected ViewType
	}{
		{'d', ViewDashboard},
		{'t', ViewToolBrowser},
		{'b', ViewBundleExplorer},
		{'m', ViewMonitor},
		{'c', ViewConfig},
		{'h', ViewHealth},
	}

	for _, test := range testKeys {
		start := time.Now()

		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{test.key},
		}

		newModel, _ := model.handleKeyPress(keyMsg)
		if m, ok := newModel.(Model); ok {
			*model = m
		} else {
			t.Fatalf("Expected Model, got %T", newModel)
		}

		elapsed := time.Since(start)

		// Navigation should be instant
		if elapsed > time.Millisecond {
			t.Errorf("Navigation to view '%c' took too long: %v", test.key, elapsed)
		}

		// Should switch to correct view
		if model.state.CurrentView != test.expected {
			t.Errorf("Key '%c' should switch to view %v, got %v", test.key, test.expected, model.state.CurrentView)
		}
	}
}

func TestApp_HealthViewInitialState(t *testing.T) {
	// Create a new model
	model, err := NewModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Initialize and set to health view
	model.ready = true
	model.state.Initialized = true
	model.state.CurrentView = ViewHealth

	// Render the health view
	content := model.renderCurrentView()

	// Should show initial "Checking..." states
	checkingCount := 0
	lines := splitLines(content)
	for _, line := range lines {
		if containsString(line, "Checking...") {
			checkingCount++
		}
	}

	// Should have several items in checking state initially
	if checkingCount < 3 {
		t.Errorf("Expected at least 3 items showing 'Checking...', got %d", checkingCount)
	}

	// Should have system health monitor title
	if !containsString(content, "Health Monitor") && !containsString(content, "System Health") {
		t.Error("Health view should contain health monitor title")
	}
}

// Helper functions
func containsString(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	
	var lines []string
	var current string
	
	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		lines = append(lines, current)
	}
	
	return lines
}