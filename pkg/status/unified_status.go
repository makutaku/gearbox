package status

import (
	"os/exec"
	"strings"
	
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// ToolStatus represents the comprehensive status of a tool
type ToolStatus struct {
	Name                string
	Installed           bool
	Version             string
	Source              string // "gearbox", "system", "unknown"
	BinaryPaths         []string
	InManifest          bool
	ManifestVersion     string
	LiveDetection       bool
	NeedsSync          bool // True if manifest and live detection disagree
}

// UnifiedStatusService provides consistent tool status across CLI and TUI
type UnifiedStatusService struct {
	manifest    *manifest.Manager
	orchestrator *orchestrator.Orchestrator
	toolConfig  []orchestrator.ToolConfig
}

// NewUnifiedStatusService creates a new unified status service
func NewUnifiedStatusService() (*UnifiedStatusService, error) {
	// Initialize manifest manager
	manifestMgr := manifest.NewManager()
	
	// Initialize orchestrator for live detection
	opts := orchestrator.InstallationOptions{
		BuildType: "standard",
	}
	orch, err := orchestrator.NewOrchestratorBuilder(opts).Build()
	if err != nil {
		return nil, err
	}
	
	return &UnifiedStatusService{
		manifest:     manifestMgr,
		orchestrator: orch,
		toolConfig:   orch.GetConfig().Tools, // Get tool configuration from orchestrator
	}, nil
}

// GetToolStatus returns comprehensive status for a single tool
func (s *UnifiedStatusService) GetToolStatus(toolName string) (*ToolStatus, error) {
	// Get tool configuration
	toolConfig := s.findToolConfig(toolName)
	if toolConfig == nil {
		return &ToolStatus{
			Name:      toolName,
			Installed: false,
			Source:    "unknown",
		}, nil
	}
	
	status := &ToolStatus{
		Name: toolName,
	}
	
	// Check manifest
	manifestData, err := s.manifest.Load()
	if err == nil && manifestData != nil && manifestData.Installations != nil {
		if record, exists := manifestData.Installations[toolName]; exists {
			status.InManifest = true
			status.ManifestVersion = record.Version
			status.BinaryPaths = record.BinaryPaths
			status.Source = "gearbox"
		}
	}
	
	// Check live system detection
	status.LiveDetection = s.isToolInstalledLive(toolConfig)
	if status.LiveDetection {
		status.Version = s.getToolVersionLive(toolConfig)
		status.Installed = true
		
		// If not in manifest but live detected, it's a system install
		if !status.InManifest {
			status.Source = "system"
		}
	}
	
	// Determine if sync is needed
	status.NeedsSync = status.InManifest != status.LiveDetection
	
	// Final installed status (true if either manifest or live detection)
	status.Installed = status.InManifest || status.LiveDetection
	
	return status, nil
}

// GetAllToolsStatus returns status for all configured tools
func (s *UnifiedStatusService) GetAllToolsStatus() (map[string]*ToolStatus, error) {
	result := make(map[string]*ToolStatus)
	
	// Get all tools from config
	for _, tool := range s.toolConfig {
		status, err := s.GetToolStatus(tool.Name)
		if err != nil {
			continue // Skip tools with errors
		}
		result[tool.Name] = status
	}
	
	// Also check manifest for any tools not in config
	manifestData, err := s.manifest.Load()
	if err == nil {
		for toolName := range manifestData.Installations {
			if _, exists := result[toolName]; !exists {
				status, err := s.GetToolStatus(toolName)
				if err == nil {
					result[toolName] = status
				}
			}
		}
	}
	
	return result, nil
}

// SyncManifestWithSystem reconciles manifest with actual system state
func (s *UnifiedStatusService) SyncManifestWithSystem() error {
	allStatus, err := s.GetAllToolsStatus()
	if err != nil {
		return err
	}
	
	// Create a tracker for handling pre-existing tools
	tracker, err := manifest.NewTracker()
	if err != nil {
		return err
	}
	
	for _, status := range allStatus {
		if status.NeedsSync {
			if status.LiveDetection && !status.InManifest {
				// Tool exists on system but not in manifest - add it as pre-existing
				err := tracker.TrackPreExisting(status.Name, strings.Join(status.BinaryPaths, ","), status.Version)
				if err != nil {
					// Log error but continue with other tools
					continue
				}
			} else if !status.LiveDetection && status.InManifest {
				// Tool in manifest but not on system - mark as missing or remove
				// This could be handled based on user preference
				// For now, we'll leave it in the manifest as it might be temporarily unavailable
			}
		}
	}
	
	return nil
}

// GetInstalledCount returns count of installed tools (for TUI dashboard)
func (s *UnifiedStatusService) GetInstalledCount() (int, error) {
	allStatus, err := s.GetAllToolsStatus()
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, status := range allStatus {
		if status.Installed {
			count++
		}
	}
	
	return count, nil
}

// Private helper methods

func (s *UnifiedStatusService) findToolConfig(toolName string) *orchestrator.ToolConfig {
	for _, tool := range s.toolConfig {
		if tool.Name == toolName {
			return &tool
		}
	}
	return nil
}

func (s *UnifiedStatusService) isToolInstalledLive(tool *orchestrator.ToolConfig) bool {
	if tool.BinaryName == "" {
		return false
	}
	
	// Check if binary is in PATH
	_, err := exec.LookPath(tool.BinaryName)
	return err == nil
}

func (s *UnifiedStatusService) getToolVersionLive(tool *orchestrator.ToolConfig) string {
	if tool.TestCommand == "" {
		return "installed"
	}
	
	binaryName := tool.BinaryName
	if binaryName == "" {
		binaryName = tool.Name
	}
	
	parts := strings.Fields(tool.TestCommand)
	cmd := exec.Command(binaryName, parts...)
	output, err := cmd.Output()
	if err != nil {
		return "installed"
	}
	
	return s.extractVersionFromOutput(string(output))
}

func (s *UnifiedStatusService) extractVersionFromOutput(output string) string {
	// This would use the same extraction logic as in utils.go
	// For now, simplified version
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Simple version extraction - could be enhanced
			return line
		}
	}
	return "unknown"
}