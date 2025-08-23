package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestApp_HealthViewAutomaticTrigger(t *testing.T) {
	// Create a new demo model (works without config files)
	tuiModel, err := NewDemoModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Cast to concrete type for testing (we know it's a DemoModel)
	model, ok := tuiModel.(*DemoModel)
	if !ok {
		t.Fatalf("Expected *DemoModel, got %T", tuiModel)
	}

	// Initialize the model
	model.ready = true
	model.state.Initialized = true

	// Simulate pressing 'H' key to switch to health view
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'H'},
	}

	// Process the key press
	newModel, cmd := model.handleKeyPress(keyMsg)
	if m, ok := newModel.(DemoModel); ok {
		*model = m
	} else {
		t.Fatalf("Expected DemoModel, got %T", newModel)
	}

	// Should switch to health view
	if tuiModel.GetCurrentView() != ViewHealth {
		t.Errorf("Expected current view to be ViewHealth, got %v", tuiModel.GetCurrentView())
	}

	// Should return a command to trigger health checks
	if cmd == nil {
		t.Error("Expected command to trigger health checks, got nil")
	}

	// Execute the command to see if it returns a BatchMsg (from RunNextHealthCheck)
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.BatchMsg); !ok {
			t.Errorf("Expected tea.BatchMsg from RunNextHealthCheck, got %T", msg)
		}
	}
}

func TestApp_HealthCheckManualRefresh(t *testing.T) {
	// Create a new demo model
	tuiModel, err := NewDemoModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Cast to concrete type for testing
	model, ok := tuiModel.(*DemoModel)
	if !ok {
		t.Fatalf("Expected *DemoModel, got %T", tuiModel)
	}

	// Initialize and set to health view
	model.ready = true
	model.state.Initialized = true
	tuiModel.SetCurrentView(ViewHealth)
	tuiModel.SetSize(80, 24)

	// Simulate pressing 'r' key for manual refresh
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'r'},
	}
	newModel, cmd := model.updateCurrentView(keyMsg)
	if m, ok := newModel.(DemoModel); ok {
		*model = m
	} else {
		t.Fatalf("Expected DemoModel, got %T", newModel)
	}

	// Should generate a command to start health checks
	if cmd == nil {
		t.Error("Expected command to trigger health refresh, got nil")
	}

	// Should be a BatchMsg from RunNextHealthCheck
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.BatchMsg); !ok {
			t.Errorf("Expected tea.BatchMsg from manual refresh, got %T", msg)
		}
	}
}

func TestApp_ZeroLatencyHealthViewSwitch(t *testing.T) {
	// Create a new model
	tuiModel, err := NewDemoModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Cast to concrete type for testing
	model, ok := tuiModel.(*DemoModel)
	if !ok {
		t.Fatalf("Expected *DemoModel, got %T", tuiModel)
	}

	// Initialize the model
	model.ready = true
	model.state.Initialized = true

	// Measure time to switch to health view
	start := time.Now()

	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'H'},
	}

	newModel, _ := model.handleKeyPress(keyMsg)
	if m, ok := newModel.(DemoModel); ok {
		*model = m
	} else {
		t.Fatalf("Expected DemoModel, got %T", newModel)
	}

	elapsed := time.Since(start)

	// Should switch view instantly (under 1ms)
	if elapsed > time.Millisecond {
		t.Errorf("Health view switch took too long: %v (should be < 1ms)", elapsed)
	}

	// Should switch to health view
	if tuiModel.GetCurrentView() != ViewHealth {
		t.Errorf("Expected current view to be ViewHealth, got %v", tuiModel.GetCurrentView())
	}
}

func TestApp_HealthViewRenderPerformance(t *testing.T) {
	// Create a new model
	tuiModel, err := NewDemoModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Cast to concrete type for testing
	model, ok := tuiModel.(*DemoModel)
	if !ok {
		t.Fatalf("Expected *DemoModel, got %T", tuiModel)
	}

	// Initialize and set to health view
	model.ready = true
	model.state.Initialized = true
	tuiModel.SetCurrentView(ViewHealth)
	tuiModel.SetSize(80, 24)

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
	tuiModel, err := NewDemoModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Cast to concrete type for testing
	model, ok := tuiModel.(*DemoModel)
	if !ok {
		t.Fatalf("Expected *DemoModel, got %T", tuiModel)
	}

	// Initialize
	model.ready = true
	model.state.Initialized = true

	// Test all navigation keys for responsiveness
	testKeys := []struct {
		key      rune
		expected ViewType
	}{
		{'D', ViewDashboard},
		{'T', ViewToolBrowser},
		{'B', ViewBundleExplorer},
		{'M', ViewMonitor},
		{'C', ViewConfig},
		{'H', ViewHealth},
	}

	for _, test := range testKeys {
		start := time.Now()

		keyMsg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{test.key},
		}

		newModel, _ := model.handleKeyPress(keyMsg)
		if m, ok := newModel.(DemoModel); ok {
			*model = m
		} else {
			t.Fatalf("Expected DemoModel, got %T", newModel)
		}

		elapsed := time.Since(start)

		// Navigation should be instant
		if elapsed > time.Millisecond {
			t.Errorf("Navigation to view '%c' took too long: %v", test.key, elapsed)
		}

		// Should switch to correct view
		if tuiModel.GetCurrentView() != test.expected {
			t.Errorf("Key '%c' should switch to view %v, got %v", test.key, test.expected, tuiModel.GetCurrentView())
		}
	}
}

func TestApp_HealthViewInitialState(t *testing.T) {
	// Create a new model
	tuiModel, err := NewDemoModel()
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Cast to concrete type for testing
	model, ok := tuiModel.(*DemoModel)
	if !ok {
		t.Fatalf("Expected *DemoModel, got %T", tuiModel)
	}

	// Initialize and set to health view
	model.ready = true
	model.state.Initialized = true
	tuiModel.SetCurrentView(ViewHealth)
	tuiModel.SetSize(80, 24)

	// Set up health view with data like the actual model does
	model.healthView.SetData(model.tools, model.installedTools)

	// Render the health view
	content := model.renderCurrentView()

	// Should show initial "Checking..." states or other pending messages
	checkingCount := 0
	lines := splitLines(content)
	for _, line := range lines {
		// Count any line with "Checking" (covers "Checking...", "Checking version...", etc.)
		if containsString(line, "Checking") {
			checkingCount++
		}
	}

	// Should have several items in checking state initially (3 exact "Checking..." + others with "Checking")
	if checkingCount < 3 {
		t.Errorf("Expected at least 3 items showing 'Checking', got %d. Content:\n%s", checkingCount, content)
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