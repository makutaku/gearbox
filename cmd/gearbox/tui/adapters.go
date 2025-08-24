package tui

import (
	"fmt"
	
	tea "github.com/charmbracelet/bubbletea"
	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// OrchestratorAdapter wraps orchestrator.Orchestrator to implement OrchestratorService
type OrchestratorAdapter struct {
	orchestrator *orchestrator.Orchestrator
}

// NewOrchestratorAdapter creates a new orchestrator adapter
func NewOrchestratorAdapter(orch *orchestrator.Orchestrator) *OrchestratorAdapter {
	return &OrchestratorAdapter{orchestrator: orch}
}

// InstallTools installs specified tools
func (o *OrchestratorAdapter) InstallTools(toolNames []string) error {
	return o.orchestrator.InstallTools(toolNames)
}

// ListTools lists available tools with optional category filter
func (o *OrchestratorAdapter) ListTools(category string, verbose bool) error {
	return o.orchestrator.ListTools(category, verbose)
}

// ListBundles lists available bundles
func (o *OrchestratorAdapter) ListBundles(verbose bool) error {
	return o.orchestrator.ListBundles(verbose)
}

// ShowStatus shows status of tools
func (o *OrchestratorAdapter) ShowStatus(toolNames []string, manifestOnly bool, unified bool) error {
	return o.orchestrator.ShowStatus(toolNames, manifestOnly, unified)
}

// GetConfig returns the orchestrator configuration
func (o *OrchestratorAdapter) GetConfig() *orchestrator.Config {
	return o.orchestrator.GetConfig()
}

// RunDoctor runs diagnostic checks
func (o *OrchestratorAdapter) RunDoctor(toolNames []string) error {
	return o.orchestrator.RunDoctor(toolNames)
}

// ManifestAdapter wraps manifest.Manager to implement ManifestService
type ManifestAdapter struct {
	manager *manifest.Manager
}

// NewManifestAdapter creates a new manifest adapter
func NewManifestAdapter(mgr *manifest.Manager) *ManifestAdapter {
	return &ManifestAdapter{manager: mgr}
}

// Load returns the current manifest data
func (m *ManifestAdapter) Load() (*manifest.InstallationManifest, error) {
	return m.manager.Load()
}

// Save persists the manifest data
func (m *ManifestAdapter) Save(manifest *manifest.InstallationManifest) error {
	return m.manager.Save(manifest)
}

// Manifest adapter implements the simplified ManifestService interface

// TaskAdapter wraps tasks.TaskManager to implement TaskService
type TaskAdapter struct {
	manager *tasks.TaskManager
}

// NewTaskAdapter creates a new task adapter
func NewTaskAdapter(mgr *tasks.TaskManager) *TaskAdapter {
	return &TaskAdapter{manager: mgr}
}

// AddTask adds a new task to the queue
func (t *TaskAdapter) AddTask(tool orchestrator.ToolConfig, buildType string) string {
	return t.manager.AddTask(tool, buildType)
}

// StartTask starts execution of a queued task
func (t *TaskAdapter) StartTask(taskID string) error {
	return t.manager.StartTask(taskID)
}

// CancelTask cancels a running task
func (t *TaskAdapter) CancelTask(taskID string) error {
	return t.manager.CancelTask(taskID)
}

// GetAllTasks returns all tasks
func (t *TaskAdapter) GetAllTasks() []*tasks.InstallTask {
	return t.manager.GetAllTasks()
}

// WatchUpdates returns a command to watch for task updates
func (t *TaskAdapter) WatchUpdates() tea.Cmd {
	return t.manager.WatchUpdates()
}

// NavigationAdapter wraps NavigationHandler to implement NavigationService
type NavigationAdapter struct {
	handler *NavigationHandler
}

// NewNavigationAdapter creates a new navigation adapter
func NewNavigationAdapter(handler *NavigationHandler) *NavigationAdapter {
	return &NavigationAdapter{handler: handler}
}

// HandleKeyPress processes keyboard navigation and returns new view and commands
func (n *NavigationAdapter) HandleKeyPress(msg tea.KeyMsg, currentView ViewType) (ViewType, tea.Cmd, bool) {
	return n.handler.HandleKeyPress(msg, currentView)
}

// GetNextView returns the next view in navigation order
func (n *NavigationAdapter) GetNextView(current ViewType) ViewType {
	return n.handler.getNextView(current)
}

// GetPreviousView returns the previous view in navigation order
func (n *NavigationAdapter) GetPreviousView(current ViewType) ViewType {
	return n.handler.getPreviousView(current)
}

// IsNavigationKey checks if a key is a navigation key
func (n *NavigationAdapter) IsNavigationKey(key string) bool {
	_, exists := n.handler.keyBindings[key]
	return exists
}

// GetKeyBindings returns all key bindings
func (n *NavigationAdapter) GetKeyBindings() map[string]ViewType {
	// Return a copy to prevent external modification
	bindings := make(map[string]ViewType)
	for k, v := range n.handler.keyBindings {
		bindings[k] = v
	}
	return bindings
}

// MessageRoutingAdapter wraps MessageRouter to implement MessageRoutingService
type MessageRoutingAdapter struct {
	router *MessageRouter
}

// NewMessageRoutingAdapter creates a new message routing adapter
func NewMessageRoutingAdapter(router *MessageRouter) *MessageRoutingAdapter {
	return &MessageRoutingAdapter{router: router}
}

// Route routes a message to appropriate handlers
func (m *MessageRoutingAdapter) Route(msg tea.Msg) tea.Cmd {
	return m.router.Route(msg)
}

// DashboardAdapter wraps views.Dashboard to implement DashboardService
type DashboardAdapter struct {
	dashboard *views.Dashboard
}

// NewDashboardAdapter creates a new dashboard adapter
func NewDashboardAdapter(dashboard *views.Dashboard) *DashboardAdapter {
	return &DashboardAdapter{dashboard: dashboard}
}

// SetSize updates the view size
func (d *DashboardAdapter) SetSize(width, height int) {
	d.dashboard.SetSize(width, height)
}

// Render renders the view content
func (d *DashboardAdapter) Render() string {
	return d.dashboard.Render()
}

// Update handles view updates
func (d *DashboardAdapter) Update(msg tea.Msg) tea.Cmd {
	return d.dashboard.Update(msg)
}

// GetType returns the view type
func (d *DashboardAdapter) GetType() ViewType {
	return ViewDashboard
}

// IsReady returns whether the view is ready for interaction
func (d *DashboardAdapter) IsReady() bool {
	// Dashboard is always ready once created
	return true
}

// Lifecycle methods implementation

// OnActivate is called when the view becomes active
func (d *DashboardAdapter) OnActivate() tea.Cmd {
	// Dashboard could refresh statistics when activated
	return nil
}

// OnDeactivate is called when the view becomes inactive
func (d *DashboardAdapter) OnDeactivate() tea.Cmd {
	// Dashboard doesn't need special deactivation handling
	return nil
}

// OnInitialize is called once when the view is first created
func (d *DashboardAdapter) OnInitialize() tea.Cmd {
	// Dashboard initialization is handled in SetData
	return nil
}

// OnDestroy is called when the view is being destroyed
func (d *DashboardAdapter) OnDestroy() tea.Cmd {
	// Dashboard doesn't have resources to clean up
	return nil
}

// OnRefresh is called when the view needs to refresh its data
func (d *DashboardAdapter) OnRefresh() tea.Cmd {
	// Dashboard refresh would trigger data reload
	return nil
}

// CanDeactivate returns whether the view can be deactivated
func (d *DashboardAdapter) CanDeactivate() (bool, string) {
	// Dashboard has no unsaved state
	return true, ""
}

// GetFocusableElements returns elements that can receive focus
func (d *DashboardAdapter) GetFocusableElements() []string {
	// Dashboard quick actions could be focusable
	return []string{"tools", "bundles", "monitor", "config", "health"}
}

// SetFocus sets focus to a specific element
func (d *DashboardAdapter) SetFocus(elementID string) error {
	// Dashboard focus handling would be implemented here
	return nil
}

// SetData updates dashboard data
func (d *DashboardAdapter) SetData(tools []orchestrator.ToolConfig, bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	d.dashboard.SetData(tools, bundles, installed)
}

// GetRecommendations returns smart recommendations for the user
func (d *DashboardAdapter) GetRecommendations() []string {
	// This would need to be implemented in the actual Dashboard view
	// For now, return empty slice
	return []string{}
}

// GetSystemStatus returns current system status
func (d *DashboardAdapter) GetSystemStatus() map[string]interface{} {
	// This would need to be implemented in the actual Dashboard view
	// For now, return empty map
	return make(map[string]interface{})
}

// ToolBrowserAdapter wraps views.ToolBrowserNew to implement ToolBrowserService
type ToolBrowserAdapter struct {
	browser *views.ToolBrowserNew
}

// NewToolBrowserAdapter creates a new tool browser adapter
func NewToolBrowserAdapter(browser *views.ToolBrowserNew) *ToolBrowserAdapter {
	return &ToolBrowserAdapter{browser: browser}
}

// SetSize updates the view size
func (t *ToolBrowserAdapter) SetSize(width, height int) {
	t.browser.SetSize(width, height)
}

// Render renders the view content
func (t *ToolBrowserAdapter) Render() string {
	return t.browser.Render()
}

// Update handles view updates
func (t *ToolBrowserAdapter) Update(msg tea.Msg) tea.Cmd {
	return t.browser.Update(msg)
}

// GetType returns the view type
func (t *ToolBrowserAdapter) GetType() ViewType {
	return ViewToolBrowser
}

// IsReady returns whether the view is ready for interaction
func (t *ToolBrowserAdapter) IsReady() bool {
	// Tool browser is ready once created
	return true
}

// Lifecycle methods implementation

// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
// OnActivate is called when the view becomes active
func (t *ToolBrowserAdapter) OnActivate() tea.Cmd {
	// Debug: Log when OnActivate is called
	debugLog("ToolBrowserAdapter.OnActivate() called - triggering async loading")
	
	// Tool browser triggers fresh unified status loading when activated
	return tea.Batch(
		// First refresh the current display
		func() tea.Msg {
			debugLog("ToolBrowserAdapter: Calling LoadFullContent()")
			t.browser.LoadFullContent()
			debugLog("ToolBrowserAdapter: Sending ToolBrowserContentLoadedMsg")
			return ToolBrowserContentLoadedMsg{}
		},
		// Then trigger a fresh unified status check
		func() tea.Msg {
			debugLog("ToolBrowserAdapter: Sending unified-status trigger")
			return struct{ trigger string }{"unified-status"}
		},
	)
}

// OnDeactivate is called when the view becomes inactive
func (t *ToolBrowserAdapter) OnDeactivate() tea.Cmd {
	// Tool browser might save selection state
	return nil
}

// OnInitialize is called once when the view is first created
func (t *ToolBrowserAdapter) OnInitialize() tea.Cmd {
	// Tool browser initialization
	return nil
}

// OnDestroy is called when the view is being destroyed
func (t *ToolBrowserAdapter) OnDestroy() tea.Cmd {
	// Clear any cached data
	return nil
}

// OnRefresh is called when the view needs to refresh its data
func (t *ToolBrowserAdapter) OnRefresh() tea.Cmd {
	// Refresh tool data
	t.browser.LoadFullContent()
	return nil
}

// CanDeactivate returns whether the view can be deactivated
func (t *ToolBrowserAdapter) CanDeactivate() (bool, string) {
	// Check if there are unsaved selections
	selectedTools := t.browser.GetSelectedTools()
	if len(selectedTools) > 0 {
		return true, fmt.Sprintf("You have %d tools selected. They will be cleared if you navigate away.", len(selectedTools))
	}
	return true, ""
}

// GetFocusableElements returns elements that can receive focus
func (t *ToolBrowserAdapter) GetFocusableElements() []string {
	// Tool list, search box, category filter
	return []string{"search", "category-filter", "tool-list", "install-button"}
}

// SetFocus sets focus to a specific element
func (t *ToolBrowserAdapter) SetFocus(elementID string) error {
	// Tool browser focus handling
	switch elementID {
	case "search":
		// Focus search input
		return nil
	case "category-filter":
		// Focus category dropdown
		return nil
	case "tool-list":
		// Focus tool list
		return nil
	case "install-button":
		// Focus install button
		return nil
	default:
		return fmt.Errorf("unknown focusable element: %s", elementID)
	}
}

// SetData updates tool browser data
func (t *ToolBrowserAdapter) SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord) {
	t.browser.SetData(tools, installed)
}

// GetSelectedTools returns currently selected tools
func (t *ToolBrowserAdapter) GetSelectedTools() []string {
	return t.browser.GetSelectedTools()
}

// ClearSelection clears all selections
func (t *ToolBrowserAdapter) ClearSelection() {
	t.browser.ClearSelection()
}

// ToggleSelection toggles selection for a specific tool
func (t *ToolBrowserAdapter) ToggleSelection(toolName string) {
	// This would need to be implemented in the actual ToolBrowserNew view
	// For now, this is a placeholder
}

// FilterByCategory filters tools by category
func (t *ToolBrowserAdapter) FilterByCategory(category string) {
	// This would need to be implemented in the actual ToolBrowserNew view
	// For now, this is a placeholder
}

// Search searches tools by name or description
func (t *ToolBrowserAdapter) Search(query string) {
	// This would need to be implemented in the actual ToolBrowserNew view
	// For now, this is a placeholder
}

// LoadFullContent loads full tool content asynchronously
func (t *ToolBrowserAdapter) LoadFullContent() {
	t.browser.LoadFullContent()
}

// BundleExplorerAdapter wraps views.BundleExplorerNew to implement BundleExplorerService
type BundleExplorerAdapter struct {
	explorer *views.BundleExplorerNew
}

// NewBundleExplorerAdapter creates a new bundle explorer adapter
func NewBundleExplorerAdapter(explorer *views.BundleExplorerNew) *BundleExplorerAdapter {
	return &BundleExplorerAdapter{explorer: explorer}
}

// SetSize updates the view size
func (b *BundleExplorerAdapter) SetSize(width, height int) {
	b.explorer.SetSize(width, height)
}

// Render renders the view content
func (b *BundleExplorerAdapter) Render() string {
	return b.explorer.Render()
}

// Update handles view updates
func (b *BundleExplorerAdapter) Update(msg tea.Msg) tea.Cmd {
	return b.explorer.Update(msg)
}

// GetType returns the view type
func (b *BundleExplorerAdapter) GetType() ViewType {
	return ViewBundleExplorer
}

// IsReady returns whether the view is ready for interaction
func (b *BundleExplorerAdapter) IsReady() bool {
	return true
}

// Lifecycle methods implementation (basic implementations for remaining adapters)

// OnActivate is called when the view becomes active
func (b *BundleExplorerAdapter) OnActivate() tea.Cmd { return nil }
func (b *BundleExplorerAdapter) OnDeactivate() tea.Cmd { return nil }
func (b *BundleExplorerAdapter) OnInitialize() tea.Cmd { return nil }
func (b *BundleExplorerAdapter) OnDestroy() tea.Cmd { return nil }
func (b *BundleExplorerAdapter) OnRefresh() tea.Cmd { return nil }
func (b *BundleExplorerAdapter) CanDeactivate() (bool, string) { return true, "" }
func (b *BundleExplorerAdapter) GetFocusableElements() []string { return []string{"bundle-list", "expand-button", "install-button"} }
func (b *BundleExplorerAdapter) SetFocus(elementID string) error { return nil }

// SetData updates bundle explorer data
func (b *BundleExplorerAdapter) SetData(bundles []orchestrator.BundleConfig, installed map[string]*manifest.InstallationRecord) {
	b.explorer.SetData(bundles, installed)
}

// GetSelectedBundle returns the currently selected bundle
func (b *BundleExplorerAdapter) GetSelectedBundle() *orchestrator.BundleConfig {
	return b.explorer.GetSelectedBundle()
}

// GetUninstalledTools returns tools in a bundle that are not installed
func (b *BundleExplorerAdapter) GetUninstalledTools(bundle *orchestrator.BundleConfig) []string {
	return b.explorer.GetUninstalledTools(bundle)
}

// ExpandBundle expands/collapses a bundle
func (b *BundleExplorerAdapter) ExpandBundle(bundleName string) {
	// This would need to be implemented in the actual BundleExplorerNew view
}

// FilterByCategory filters bundles by category
func (b *BundleExplorerAdapter) FilterByCategory(category string) {
	// This would need to be implemented in the actual BundleExplorerNew view
}

// InstallManagerAdapter wraps views.InstallManagerNew to implement InstallManagerService
type InstallManagerAdapter struct {
	manager *views.InstallManagerNew
}

// NewInstallManagerAdapter creates a new install manager adapter
func NewInstallManagerAdapter(manager *views.InstallManagerNew) *InstallManagerAdapter {
	return &InstallManagerAdapter{manager: manager}
}

// SetSize updates the view size
func (i *InstallManagerAdapter) SetSize(width, height int) {
	i.manager.SetSize(width, height)
}

// Render renders the view content
func (i *InstallManagerAdapter) Render() string {
	return i.manager.Render()
}

// Update handles view updates
func (i *InstallManagerAdapter) Update(msg tea.Msg) tea.Cmd {
	return i.manager.Update(msg)
}

// GetType returns the view type
func (i *InstallManagerAdapter) GetType() ViewType {
	return ViewMonitor
}

// IsReady returns whether the view is ready for interaction
func (i *InstallManagerAdapter) IsReady() bool {
	return true
}

// Lifecycle methods implementation

func (i *InstallManagerAdapter) OnActivate() tea.Cmd { return nil }
func (i *InstallManagerAdapter) OnDeactivate() tea.Cmd { return nil }
func (i *InstallManagerAdapter) OnInitialize() tea.Cmd { return nil }
func (i *InstallManagerAdapter) OnDestroy() tea.Cmd { return nil }
func (i *InstallManagerAdapter) OnRefresh() tea.Cmd { return nil }
func (i *InstallManagerAdapter) CanDeactivate() (bool, string) { return true, "" }
func (i *InstallManagerAdapter) GetFocusableElements() []string { return []string{"task-list", "start-button", "cancel-button"} }
func (i *InstallManagerAdapter) SetFocus(elementID string) error { return nil }

// AddTaskID adds a task ID to monitor
func (i *InstallManagerAdapter) AddTaskID(taskID string) {
	i.manager.AddTaskID(taskID)
}

// RemoveTaskID removes a task ID from monitoring
func (i *InstallManagerAdapter) RemoveTaskID(taskID string) {
	// This would need to be implemented in the actual InstallManagerNew view
}

// HandleTaskUpdate handles task progress updates
func (i *InstallManagerAdapter) HandleTaskUpdate(taskID string, progress float64) {
	i.manager.HandleTaskUpdate(taskID, progress)
}

// GetActiveTaskCount returns the number of active tasks
func (i *InstallManagerAdapter) GetActiveTaskCount() int {
	// This would need to be implemented in the actual InstallManagerNew view
	return 0
}

// ToggleOutputDisplay toggles output visibility
func (i *InstallManagerAdapter) ToggleOutputDisplay() {
	// This would need to be implemented in the actual InstallManagerNew view
}

// ClearCompletedTasks removes completed tasks from display
func (i *InstallManagerAdapter) ClearCompletedTasks() {
	// This would need to be implemented in the actual InstallManagerNew view
}

// ConfigAdapter wraps views.ConfigView to implement ConfigService
type ConfigAdapter struct {
	config *views.ConfigView
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter(config *views.ConfigView) *ConfigAdapter {
	return &ConfigAdapter{config: config}
}

// SetSize updates the view size
func (c *ConfigAdapter) SetSize(width, height int) {
	c.config.SetSize(width, height)
}

// Render renders the view content
func (c *ConfigAdapter) Render() string {
	return c.config.Render()
}

// Update handles view updates
func (c *ConfigAdapter) Update(msg tea.Msg) tea.Cmd {
	return c.config.Update(msg)
}

// GetType returns the view type
func (c *ConfigAdapter) GetType() ViewType {
	return ViewConfig
}

// IsReady returns whether the view is ready for interaction
func (c *ConfigAdapter) IsReady() bool {
	return true
}

// Lifecycle methods implementation

func (c *ConfigAdapter) OnActivate() tea.Cmd { return nil }
func (c *ConfigAdapter) OnDeactivate() tea.Cmd { return nil }
func (c *ConfigAdapter) OnInitialize() tea.Cmd { return nil }
func (c *ConfigAdapter) OnDestroy() tea.Cmd { return nil }
func (c *ConfigAdapter) OnRefresh() tea.Cmd { return nil }
func (c *ConfigAdapter) CanDeactivate() (bool, string) { 
	// Config view might have unsaved changes
	return true, "Configuration changes will be lost if not saved"
}
func (c *ConfigAdapter) GetFocusableElements() []string { return []string{"settings-list", "save-button", "reset-button"} }
func (c *ConfigAdapter) SetFocus(elementID string) error { return nil }

// GetSettings returns all configuration settings
func (c *ConfigAdapter) GetSettings() map[string]interface{} {
	// This would need to be implemented in the actual ConfigView
	return make(map[string]interface{})
}

// SetSetting updates a configuration setting
func (c *ConfigAdapter) SetSetting(key string, value interface{}) error {
	// This would need to be implemented in the actual ConfigView
	return nil
}

// ResetSetting resets a setting to default value
func (c *ConfigAdapter) ResetSetting(key string) error {
	// This would need to be implemented in the actual ConfigView
	return nil
}

// SaveSettings persists all settings
func (c *ConfigAdapter) SaveSettings() error {
	// This would need to be implemented in the actual ConfigView
	return nil
}

// LoadSettings loads settings from storage
func (c *ConfigAdapter) LoadSettings() error {
	// This would need to be implemented in the actual ConfigView
	return nil
}

// GetSettingType returns the type of a setting
func (c *ConfigAdapter) GetSettingType(key string) string {
	// This would need to be implemented in the actual ConfigView
	return "string"
}

// HealthAdapter wraps views.HealthView to implement HealthService
type HealthAdapter struct {
	health *views.HealthView
}

// NewHealthAdapter creates a new health adapter
func NewHealthAdapter(health *views.HealthView) *HealthAdapter {
	return &HealthAdapter{health: health}
}

// SetSize updates the view size
func (h *HealthAdapter) SetSize(width, height int) {
	h.health.SetSize(width, height)
}

// Render renders the view content
func (h *HealthAdapter) Render() string {
	return h.health.Render()
}

// Update handles view updates
func (h *HealthAdapter) Update(msg tea.Msg) tea.Cmd {
	return h.health.Update(msg)
}

// GetType returns the view type
func (h *HealthAdapter) GetType() ViewType {
	return ViewHealth
}

// IsReady returns whether the view is ready for interaction
func (h *HealthAdapter) IsReady() bool {
	return true
}

// Lifecycle methods implementation

// OnActivate is called when the view becomes active
func (h *HealthAdapter) OnActivate() tea.Cmd {
	// Health view automatically starts health checks when activated
	return h.health.RunNextHealthCheck(0)
}

// OnDeactivate is called when the view becomes inactive
func (h *HealthAdapter) OnDeactivate() tea.Cmd {
	// Stop any ongoing health checks
	return nil
}

// OnInitialize is called once when the view is first created
func (h *HealthAdapter) OnInitialize() tea.Cmd {
	// Initialize health checks
	return nil
}

// OnDestroy is called when the view is being destroyed
func (h *HealthAdapter) OnDestroy() tea.Cmd {
	// Clean up any running health checks
	return nil
}

// OnRefresh is called when the view needs to refresh its data
func (h *HealthAdapter) OnRefresh() tea.Cmd {
	// Re-run all health checks
	return h.health.RunNextHealthCheck(0)
}

// CanDeactivate returns whether the view can be deactivated
func (h *HealthAdapter) CanDeactivate() (bool, string) {
	// Health checks can be interrupted
	return true, ""
}

// GetFocusableElements returns elements that can receive focus
func (h *HealthAdapter) GetFocusableElements() []string {
	// Health check list, refresh button, details toggle
	return []string{"health-list", "refresh-button", "details-toggle", "auto-refresh-toggle"}
}

// SetFocus sets focus to a specific element
func (h *HealthAdapter) SetFocus(elementID string) error {
	// Health view focus handling
	switch elementID {
	case "health-list":
		// Focus health check list
		return nil
	case "refresh-button":
		// Focus refresh button
		return nil
	case "details-toggle":
		// Focus details toggle
		return nil
	case "auto-refresh-toggle":
		// Focus auto-refresh toggle
		return nil
	default:
		return fmt.Errorf("unknown focusable element: %s", elementID)
	}
}

// SetData updates health view data
func (h *HealthAdapter) SetData(tools []orchestrator.ToolConfig, installed map[string]*manifest.InstallationRecord) {
	h.health.SetData(tools, installed)
}

// RunNextHealthCheck runs the next health check in sequence
func (h *HealthAdapter) RunNextHealthCheck(checkIndex int) tea.Cmd {
	return h.health.RunNextHealthCheck(checkIndex)
}

// GetHealthStatus returns overall health status
func (h *HealthAdapter) GetHealthStatus() (passing int, warning int, failing int) {
	// This would need to be implemented in the actual HealthView
	// For now, return default values
	return 0, 0, 0
}

// ToggleAutoRefresh toggles automatic refresh
func (h *HealthAdapter) ToggleAutoRefresh() {
	// This would need to be implemented in the actual HealthView
}

// ToggleDetails toggles detail visibility
func (h *HealthAdapter) ToggleDetails() {
	// This would need to be implemented in the actual HealthView
}

// GetSystemChecks returns system health checks
func (h *HealthAdapter) GetSystemChecks() []views.HealthCheck {
	// This would need to be implemented in the actual HealthView
	return []views.HealthCheck{}
}

// GetToolChecks returns tool health checks
func (h *HealthAdapter) GetToolChecks() []views.HealthCheck {
	// This would need to be implemented in the actual HealthView
	return []views.HealthCheck{}
}