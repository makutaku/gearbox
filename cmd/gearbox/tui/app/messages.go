package app

import (
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// Application-specific messages

// QuitRequestedMsg indicates that quit was requested
type QuitRequestedMsg struct{}

// DataLoadedMsg indicates that initial data has been loaded
type DataLoadedMsg struct {
	Tools    []orchestrator.ToolConfig
	Bundles  []orchestrator.BundleConfig
	Installed map[string]*manifest.InstallationRecord
}

// ManifestReloadedMsg indicates that manifest data has been reloaded
type ManifestReloadedMsg struct {
	Installed map[string]*manifest.InstallationRecord
}

// UnifiedStatusLoadedMsg indicates that unified status data has been loaded
type UnifiedStatusLoadedMsg struct {
	Installed map[string]*manifest.InstallationRecord
}

// ToolBrowserContentLoadedMsg indicates that tool browser content has been loaded
type ToolBrowserContentLoadedMsg struct{}

// HealthCheckTriggerMsg triggers health checks
type HealthCheckTriggerMsg struct{}

// ErrorMsg wraps application errors
type ErrorMsg struct {
	Err error
}

// Error returns the wrapped error
func (e ErrorMsg) Error() string {
	return e.Err.Error()
}

// StartupDataLoadMsg triggers startup data loading
type StartupDataLoadMsg struct{}

// NewQuitRequestedMsg creates a new quit requested message
func NewQuitRequestedMsg() QuitRequestedMsg {
	return QuitRequestedMsg{}
}

// NewDataLoadedMsg creates a new data loaded message
func NewDataLoadedMsg(tools []orchestrator.ToolConfig, bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) DataLoadedMsg {
	return DataLoadedMsg{
		Tools:     tools,
		Bundles:   bundles,
		Installed: installed,
	}
}

// NewManifestReloadedMsg creates a new manifest reloaded message
func NewManifestReloadedMsg(installed map[string]*manifest.InstallationRecord) ManifestReloadedMsg {
	return ManifestReloadedMsg{
		Installed: installed,
	}
}

// NewUnifiedStatusLoadedMsg creates a new unified status loaded message
func NewUnifiedStatusLoadedMsg(installed map[string]*manifest.InstallationRecord) UnifiedStatusLoadedMsg {
	return UnifiedStatusLoadedMsg{
		Installed: installed,
	}
}

// NewToolBrowserContentLoadedMsg creates a new tool browser content loaded message
func NewToolBrowserContentLoadedMsg() ToolBrowserContentLoadedMsg {
	return ToolBrowserContentLoadedMsg{}
}

// NewHealthCheckTriggerMsg creates a new health check trigger message
func NewHealthCheckTriggerMsg() HealthCheckTriggerMsg {
	return HealthCheckTriggerMsg{}
}

// NewErrorMsg creates a new error message
func NewErrorMsg(err error) ErrorMsg {
	return ErrorMsg{Err: err}
}

// NewStartupDataLoadMsg creates a new startup data load message
func NewStartupDataLoadMsg() StartupDataLoadMsg {
	return StartupDataLoadMsg{}
}