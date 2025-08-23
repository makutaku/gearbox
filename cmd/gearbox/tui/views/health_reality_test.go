package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// TestHealthView_RealityCheck - This test should mirror exactly what happens in the real TUI
func TestHealthView_RealityCheck(t *testing.T) {
	// Create health view exactly like the real TUI does
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Set some mock data like the real TUI does
	tools := []orchestrator.ToolConfig{
		{Name: "test-tool", Category: "test"},
	}
	installed := make(map[string]*manifest.InstallationRecord)
	hv.SetData(tools, installed)
	
	// Render initial state - this should show "Checking..."
	initialContent := hv.Render()
	t.Logf("Initial render:\n%s", initialContent)
	
	// Count how many items are in "Checking..." state initially
	checkingCount := 0
	lines := splitLines(initialContent)
	for _, line := range lines {
		if containsString(line, "Checking...") {
			checkingCount++
			t.Logf("Found checking line: %s", line)
		}
	}
	t.Logf("Initial checking items: %d", checkingCount)
	
	// Now simulate what happens when user presses 'r' to refresh health checks
	// This should trigger individual async health checks
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'r'},
	}
	
	// Process the 'r' key - this should return individual async commands
	cmd := hv.Update(keyMsg)
	if cmd == nil {
		t.Fatal("Expected health check commands after 'r' key, got nil")
	}
	
	// Execute the individual async commands and apply their results
	// This simulates each health check completing asynchronously
	
	// Test memory check
	memoryCmd := hv.RunMemoryCheckAsync()
	memoryMsg := memoryCmd()
	if memMsg, ok := memoryMsg.(MemoryCheckCompleteMsg); ok {
		t.Logf("Memory check: '%s' (status: %v)", memMsg.Result.Message, memMsg.Result.Status)
		hv.Update(memMsg)
	}
	
	// Test disk check 
	diskCmd := hv.RunDiskCheckAsync()
	diskMsg := diskCmd()
	if dMsg, ok := diskMsg.(DiskCheckCompleteMsg); ok {
		t.Logf("Disk check: '%s' (status: %v)", dMsg.Result.Message, dMsg.Result.Status)
		hv.Update(dMsg)
	}
	
	// Test internet check
	internetCmd := hv.RunInternetCheckAsync()
	internetMsg := internetCmd()
	if iMsg, ok := internetMsg.(InternetCheckCompleteMsg); ok {
		t.Logf("Internet check: '%s' (status: %v)", iMsg.Result.Message, iMsg.Result.Status)
		hv.Update(iMsg)
	}
	
	// Test remaining checks
	buildCmd := hv.RunBuildToolsCheckAsync()
	buildMsg := buildCmd()
	if bMsg, ok := buildMsg.(BuildToolsCheckCompleteMsg); ok {
		t.Logf("Build tools check: '%s' (status: %v)", bMsg.Result.Message, bMsg.Result.Status)
		hv.Update(bMsg)
	}
	
	gitCmd := hv.RunGitCheckAsync()
	gitMsg := gitCmd()
	if gMsg, ok := gitMsg.(GitCheckCompleteMsg); ok {
		t.Logf("Git check: '%s' (status: %v)", gMsg.Result.Message, gMsg.Result.Status)
		hv.Update(gMsg)
	}
	
	pathCmd := hv.RunPathCheckAsync()
	pathMsg := pathCmd()
	if pMsg, ok := pathMsg.(PathCheckCompleteMsg); ok {
		t.Logf("PATH check: '%s' (status: %v)", pMsg.Result.Message, pMsg.Result.Status)
		hv.Update(pMsg)
	}
	
	// Render again - this should show the updated results
	updatedContent := hv.Render()
	t.Logf("Updated render:\n%s", updatedContent)
	
	// Count how many items are still in "Checking..." state
	stillCheckingCount := 0
	updatedLines := splitLines(updatedContent)
	for _, line := range updatedLines {
		if containsString(line, "Checking...") {
			stillCheckingCount++
			t.Logf("Still checking: %s", line)
		}
	}
	
	if stillCheckingCount > 0 {
		t.Logf("Items still checking: %d (some may be tool checks)", stillCheckingCount)
	}
	
	// Verify specific checks that the user reported as stuck
	memoryWorking := false
	diskWorking := false
	internetWorking := false
	
	for _, line := range updatedLines {
		if containsString(line, "Memory") && !containsString(line, "Checking...") {
			memoryWorking = true
			t.Logf("Memory check is working: %s", line)
		}
		if containsString(line, "Disk Space") && !containsString(line, "Checking...") {
			diskWorking = true
			t.Logf("Disk check is working: %s", line)
		}
		if containsString(line, "Internet Connection") && !containsString(line, "Checking...") {
			internetWorking = true
			t.Logf("Internet check is working: %s", line)
		}
	}
	
	if !memoryWorking {
		t.Error("Memory check is still showing 'Checking...' - this matches the user's report!")
	}
	if !diskWorking {
		t.Error("Disk Space check is still showing 'Checking...' - this matches the user's report!")
	}
	if !internetWorking {
		t.Error("Internet Connection check is still showing 'Checking...' - this matches the user's report!")
	}
}

// TestHealthView_DirectFunctionCalls - This tests the functions directly (should pass)
func TestHealthView_DirectFunctionCalls(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Call the functions directly - these should work
	memResult := hv.checkMemory()
	t.Logf("Direct memory check: Index=%d, Message='%s', Status=%v", memResult.Index, memResult.Message, memResult.Status)
	
	diskResult := hv.checkDiskSpace()
	t.Logf("Direct disk check: Index=%d, Message='%s', Status=%v", diskResult.Index, diskResult.Message, diskResult.Status)
	
	internetResult := hv.checkInternet()
	t.Logf("Direct internet check: Index=%d, Message='%s', Status=%v", internetResult.Index, internetResult.Message, internetResult.Status)
	
	// These should all return real data
	if memResult.Message == "Checking..." || memResult.Message == "" {
		t.Errorf("Direct memory check failed: %s", memResult.Message)
	}
	if diskResult.Message == "Checking..." || diskResult.Message == "" {
		t.Errorf("Direct disk check failed: %s", diskResult.Message) 
	}
	if internetResult.Message == "Checking..." || internetResult.Message == "" {
		t.Errorf("Direct internet check failed: %s", internetResult.Message)
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