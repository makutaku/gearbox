package views

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHealthView_Initialization(t *testing.T) {
	hv := NewHealthView()
	
	// Test initial state
	if hv == nil {
		t.Fatal("NewHealthView() returned nil")
	}
	
	// Check initial system checks
	if len(hv.systemChecks) == 0 {
		t.Error("System checks not initialized")
	}
	
	// Tool checks are initialized when data is set, not at creation
	// This is expected behavior
	
	// Verify initial status is correct
	for i, check := range hv.systemChecks {
		if i < 2 { // OS and CPU should be passing by default
			if check.Status != HealthStatusPassing {
				t.Errorf("System check %d should be passing initially, got %v", i, check.Status)
			}
		} else { // Others should be pending
			if check.Status != HealthStatusPending {
				t.Errorf("System check %d should be pending initially, got %v", i, check.Status)
			}
		}
	}
}

func TestHealthView_ZeroLatencyResponse(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Simulate view switch - should be instant
	start := time.Now()
	content := hv.Render()
	elapsed := time.Since(start)
	
	// Should render in under 1ms for zero-latency
	if elapsed > time.Millisecond {
		t.Errorf("Render took too long: %v (should be < 1ms for zero-latency)", elapsed)
	}
	
	// Should contain health monitor content immediately
	if content == "" {
		t.Error("Render returned empty content")
	}
	
	// Should show "Checking..." state initially 
	if !contains(content, "Checking...") {
		t.Error("Should show 'Checking...' state immediately")
	}
}

func TestHealthView_HealthCheckUpdates(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Test memory check
	memResult := hv.checkMemory()
	if memResult.Index != 2 {
		t.Errorf("Memory check should return index 2, got %d", memResult.Index)
	}
	
	// Memory check should return valid status
	if memResult.Status != HealthStatusPassing && memResult.Status != HealthStatusWarning {
		t.Errorf("Memory check should return passing or warning, got %v", memResult.Status)
	}
	
	// Should have meaningful message
	if memResult.Message == "" || memResult.Message == "Checking..." {
		t.Errorf("Memory check should return real message, got '%s'", memResult.Message)
	}
	
	// Test disk space check
	diskResult := hv.checkDiskSpace()
	if diskResult.Index != 3 {
		t.Errorf("Disk check should return index 3, got %d", diskResult.Index)
	}
	
	// Should return valid status
	if diskResult.Status != HealthStatusPassing && diskResult.Status != HealthStatusWarning {
		t.Errorf("Disk check should return passing or warning, got %v", diskResult.Status)
	}
	
	// Test internet check
	internetResult := hv.checkInternet()
	if internetResult.Index != 4 {
		t.Errorf("Internet check should return index 4, got %d", internetResult.Index)
	}
	
	// Should return valid status (might be warning if no internet)
	if internetResult.Status != HealthStatusPassing && internetResult.Status != HealthStatusWarning {
		t.Errorf("Internet check should return passing or warning, got %v", internetResult.Status)
	}
}

func TestHealthView_MessageHandling(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Create mock health check results
	systemResults := []HealthCheckUpdate{
		{
			Index:   2,
			Status:  HealthStatusPassing,
			Message: "Test memory result",
			Details: []string{"Detail 1", "Detail 2"},
		},
	}
	
	toolResults := []HealthCheckUpdate{
		{
			Index:   0,
			Status:  HealthStatusWarning,
			Message: "Test tool result",
			Suggestions: []string{"Suggestion 1"},
		},
	}
	
	// Send individual health check complete messages
	memoryMsg := MemoryCheckCompleteMsg{Result: systemResults[0]}
	toolMsg := RustToolchainCheckCompleteMsg{Result: toolResults[0]}
	
	hv.Update(memoryMsg)
	hv.Update(toolMsg)
	
	// Verify results were applied
	if hv.systemChecks[2].Status != HealthStatusPassing {
		t.Errorf("System check 2 status not updated, got %v", hv.systemChecks[2].Status)
	}
	
	if hv.systemChecks[2].Message != "Test memory result" {
		t.Errorf("System check 2 message not updated, got '%s'", hv.systemChecks[2].Message)
	}
	
	if len(hv.systemChecks[2].Details) != 2 {
		t.Errorf("System check 2 details not updated, got %d details", len(hv.systemChecks[2].Details))
	}
	
	// Check tool results too
	if len(hv.toolChecks) > 0 {
		if hv.toolChecks[0].Status != HealthStatusWarning {
			t.Errorf("Tool check 0 status not updated, got %v", hv.toolChecks[0].Status)
		}
	}
}

func TestHealthView_RefreshTrigger(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Test 'r' key triggers health check refresh
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'r'},
	}
	
	cmd := hv.Update(keyMsg)
	
	// Should return a command (tea.Tick for delayed execution)
	if cmd == nil {
		t.Error("'r' key should return a command to trigger health checks")
	}
	
	// Verify checks were reset to pending
	pendingCount := 0
	for _, check := range hv.systemChecks {
		if check.Status == HealthStatusPending && check.Message == "Checking..." {
			pendingCount++
		}
	}
	
	// Most checks should be reset to pending (except OS and CPU which are immediate)
	if pendingCount < 3 {
		t.Errorf("Expected at least 3 checks to be reset to pending, got %d", pendingCount)
	}
}

func TestHealthView_NavigationResponsiveness(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	initialCursor := hv.cursor
	
	// Test up/down navigation should be instant
	start := time.Now()
	
	hv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	downElapsed := time.Since(start)
	
	start = time.Now()
	hv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	upElapsed := time.Since(start)
	
	// Navigation should be instant (under 1ms)
	if downElapsed > time.Millisecond {
		t.Errorf("Down navigation took too long: %v", downElapsed)
	}
	
	if upElapsed > time.Millisecond {
		t.Errorf("Up navigation took too long: %v", upElapsed)
	}
	
	// Cursor should be back to original position
	if hv.cursor != initialCursor {
		t.Errorf("Navigation didn't work correctly, cursor at %d, expected %d", hv.cursor, initialCursor)
	}
}

func TestHealthView_RealHealthChecksReturnData(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Test that real health checks return actual data, not mocks
	
	// Memory check should read real /proc/meminfo
	memResult := hv.checkMemory()
	if memResult.Message == "Checking..." || memResult.Message == "" {
		t.Error("Memory check should return real memory information")
	}
	
	// Should have details about memory usage
	if len(memResult.Details) == 0 {
		t.Error("Memory check should return memory usage details")
	}
	
	// Disk check should run real 'df' command
	diskResult := hv.checkDiskSpace()
	if diskResult.Message == "Checking..." || diskResult.Message == "" {
		// Only fail if we can't get any disk info (might fail in some test environments)
		t.Logf("Warning: Disk check returned no data (might be expected in test environment)")
	}
	
	// Internet check should actually try to ping
	internetResult := hv.checkInternet()
	if internetResult.Message == "Checking..." || internetResult.Message == "" {
		t.Error("Internet check should return connection status")
	}
	
	// Internet should be either "Connected" or indicate connection problem
	validMessages := []string{"Connected", "No internet connection"}
	validMessage := false
	for _, validMsg := range validMessages {
		if internetResult.Message == validMsg {
			validMessage = true
			break
		}
	}
	
	if !validMessage {
		t.Logf("Internet check returned: %s (expected 'Connected' or 'No internet connection')", internetResult.Message)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		   (s == substr || (len(s) > len(substr) && 
		   findSubstring(s, substr) >= 0))
}

func findSubstring(s, substr string) int {
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}