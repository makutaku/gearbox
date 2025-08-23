package views

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

func TestHealthView_AsyncHealthCheckExecution(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Set some mock data
	tools := []orchestrator.ToolConfig{
		{Name: "test-tool", Category: "test"},
	}
	installed := make(map[string]*manifest.InstallationRecord)
	installed["test-tool"] = &manifest.InstallationRecord{Version: "1.0"}
	
	hv.SetData(tools, installed)
	
	// Test individual async health checks
	memoryCmd := hv.RunMemoryCheckAsync()
	memoryMsg := memoryCmd()
	
	if memMsg, ok := memoryMsg.(MemoryCheckCompleteMsg); ok {
		// Should have a valid result
		if memMsg.Result.Message == "" {
			t.Error("Expected memory check result message")
		}
		
		// Apply the result
		hv.Update(memMsg)
		
		// Verify that memory check was updated
		if hv.systemChecks[2].Message == "Checking..." {
			t.Error("Memory check still shows 'Checking...' after async update")
		}
		
		t.Logf("Successfully updated memory check: %s", hv.systemChecks[2].Message)
	} else {
		t.Errorf("Expected MemoryCheckCompleteMsg, got %T", memoryMsg)
	}
}

func TestHealthView_RealHealthCheckExecution(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Test individual real health checks
	memoryResult := hv.checkMemory()
	diskResult := hv.checkDiskSpace()
	internetResult := hv.checkInternet()
	
	// Check that results contain real data, not "Checking..."
	if memoryResult.Message == "Checking..." || memoryResult.Message == "" {
		t.Errorf("Memory check should return real data, got '%s'", memoryResult.Message)
	}
	
	if diskResult.Message == "Checking..." || diskResult.Message == "" {
		t.Errorf("Disk check should return real data, got '%s'", diskResult.Message)
	}
	
	if internetResult.Message == "Checking..." || internetResult.Message == "" {
		t.Errorf("Internet check should return real data, got '%s'", internetResult.Message)
	}
	
	// Should have valid statuses
	if memoryResult.Status != HealthStatusPassing && memoryResult.Status != HealthStatusWarning && memoryResult.Status != HealthStatusFailing {
		t.Errorf("Memory check has invalid status: %v", memoryResult.Status)
	}
	
	t.Logf("Memory check: %s - %s", memoryResult.Message, memoryResult.Status.String())
	t.Logf("Disk check: %s - %s", diskResult.Message, diskResult.Status.String())
	t.Logf("Internet check: %s - %s", internetResult.Message, internetResult.Status.String())
}

func TestHealthView_ThreadSafeUpdates(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Create individual health check result messages
	memoryMsg := MemoryCheckCompleteMsg{
		Result: HealthCheckUpdate{
			Index: 2, 
			Status: HealthStatusPassing, 
			Message: "Mock memory result",
		},
	}
	
	diskMsg := DiskCheckCompleteMsg{
		Result: HealthCheckUpdate{
			Index: 3, 
			Status: HealthStatusWarning, 
			Message: "Mock disk result",
		},
	}
	
	// Apply results via individual messages (simulates thread-safe async updates)
	hv.Update(memoryMsg)
	hv.Update(diskMsg)
	
	// Verify updates were applied correctly
	if hv.systemChecks[2].Message != "Mock memory result" {
		t.Errorf("Memory result not applied, got '%s'", hv.systemChecks[2].Message)
	}
	
	if hv.systemChecks[2].Status != HealthStatusPassing {
		t.Errorf("Memory status not applied, got %v", hv.systemChecks[2].Status)
	}
	
	if hv.systemChecks[3].Message != "Mock disk result" {
		t.Errorf("Disk result not applied, got '%s'", hv.systemChecks[3].Message)
	}
	
	if hv.systemChecks[3].Status != HealthStatusWarning {
		t.Errorf("Disk status not applied, got %v", hv.systemChecks[3].Status)
	}
}

func TestHealthView_PerformanceUnderLoad(t *testing.T) {
	hv := NewHealthView()
	hv.SetSize(80, 24)
	
	// Test multiple rapid updates (simulating UI responsiveness)
	for i := 0; i < 100; i++ {
		start := time.Now()
		
		// Navigation should be fast
		hv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		hv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		
		elapsed := time.Since(start)
		if elapsed > time.Millisecond {
			t.Errorf("Navigation iteration %d took too long: %v", i, elapsed)
			break
		}
	}
	
	// Rendering should also be fast
	for i := 0; i < 50; i++ {
		start := time.Now()
		content := hv.Render()
		elapsed := time.Since(start)
		
		if elapsed > 5*time.Millisecond {
			t.Errorf("Render iteration %d took too long: %v", i, elapsed)
			break
		}
		
		if content == "" {
			t.Errorf("Render iteration %d returned empty content", i)
			break
		}
	}
}

// Add String method to HealthStatus for better test output
func (h HealthStatus) String() string {
	switch h {
	case HealthStatusPending:
		return "Pending"
	case HealthStatusPassing:
		return "Passing"
	case HealthStatusWarning:
		return "Warning"
	case HealthStatusFailing:
		return "Failing"
	default:
		return "Unknown"
	}
}