package tui

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"

	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
	"gearbox/pkg/status"
)

type Model struct {
	orchestrator *orchestrator.Orchestrator
	manifest     *manifest.Manager
	state        *AppState
	taskManager  *tasks.TaskManager
	
	// Views
	dashboard      *views.Dashboard
	toolBrowser    *views.ToolBrowserNew
	bundleExplorer *views.BundleExplorerNew
	installManager *views.InstallManagerNew
	configView     *views.ConfigView
	healthView     *views.HealthView
	
	// UI state
	width    int
	height   int
	ready    bool
	err      error
}

func NewModel() (*Model, error) {
	// Set up file-based logging for TUI to avoid interfering with the interface
	logFile, err := os.OpenFile("/tmp/gearbox-tui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
	
	// Log session start
	if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "\n=== TUI SESSION STARTED ===\n")
		debugFile.Close()
	}
	
	// Initialize orchestrator with default options
	opts := orchestrator.InstallationOptions{
		BuildType: "standard",
	}
	orch, err := orchestrator.NewOrchestrator(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create orchestrator")
	}

	// Initialize manifest manager
	manifestMgr := manifest.NewManager()

	// Create app state
	state := NewAppState()
	
	// Create task manager
	taskManager := tasks.NewTaskManager(orch, 2)
	
	// Create task provider
	taskProvider := NewTaskManagerProvider(taskManager)
	
	// Create views
	dashboard := views.NewDashboard()
	toolBrowser := views.NewToolBrowserNew()
	bundleExplorer := views.NewBundleExplorerNew()
	installManager := views.NewInstallManagerNew()
	installManager.SetTaskProvider(taskProvider)
	configView := views.NewConfigView()
	healthView := views.NewHealthView()
	
	model := &Model{
		orchestrator:   orch,
		manifest:       manifestMgr,
		state:          state,
		taskManager:    taskManager,
		dashboard:      dashboard,
		toolBrowser:    toolBrowser,
		bundleExplorer: bundleExplorer,
		installManager: installManager,
		configView:     configView,
		healthView:     healthView,
		ready:          true, // Start ready since we're not blocking on data
		width:          80,   // Sensible default width
		height:         24,   // Sensible default height
	}

	// Initialize views with default sizes so they work immediately
	viewHeight := max(5, model.height - 2)
	model.dashboard.SetSize(model.width, viewHeight)
	model.toolBrowser.SetSize(model.width, viewHeight)
	model.bundleExplorer.SetSize(model.width, viewHeight)
	model.installManager.SetSize(model.width, viewHeight)
	model.configView.SetSize(model.width, viewHeight)
	model.healthView.SetSize(model.width, viewHeight)

	return model, nil
}

func (m Model) Init() tea.Cmd {
	// ABSOLUTE MINIMUM - only screen setup, zero other operations
	// Everything else is triggered lazily on first user interaction
	return tea.EnterAltScreen
}

// Removed loadDataAfterStartup - using lazy initialization on first key press instead

func (m Model) watchTaskUpdates() tea.Cmd {
	return m.taskManager.WatchUpdates()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Track previous view to detect when we switch TO health view
	previousView := m.state.CurrentView
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// TUI is already ready from initialization, just update sizes
		// Update view sizes
		// Calculate available height for content (excluding nav and status bars)
		// Nav bar and status bar each take 1 line, so views get height - 2
		viewHeight := max(5, m.height - 2) // Minimum 5 lines for views
		m.dashboard.SetSize(m.width, viewHeight)
		m.toolBrowser.SetSize(m.width, viewHeight)
		m.bundleExplorer.SetSize(m.width, viewHeight)
		m.installManager.SetSize(m.width, viewHeight)
		m.configView.SetSize(m.width, viewHeight)
		m.healthView.SetSize(m.width, viewHeight)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case dataLoadedMsg:
		m.state.Tools = msg.tools
		m.state.Bundles = msg.bundles
		m.state.InstalledTools = msg.installed
		// Update dashboard data
		m.dashboard.SetData(msg.tools, msg.bundles, msg.installed)
		// Update tool browser data
		m.toolBrowser.SetData(msg.tools, msg.installed)
		// Update bundle explorer data
		m.bundleExplorer.SetData(msg.bundles, msg.installed)
		// Update health view data
		m.healthView.SetData(msg.tools, msg.installed)
		// Start background unified status loading
		return m, m.loadUnifiedStatusBackground()

	case manifestReloadedMsg:
		// Update only the installed tools data
		m.state.InstalledTools = msg.installed
		// Update all views with the new installed tools data
		m.dashboard.SetData(m.state.Tools, m.state.Bundles, msg.installed)
		m.toolBrowser.SetData(m.state.Tools, msg.installed)
		m.bundleExplorer.SetData(m.state.Bundles, msg.installed)
		m.healthView.SetData(m.state.Tools, msg.installed)
		return m, nil

	case unifiedStatusLoadedMsg:
		// Update with unified status data (background loading complete)
		m.state.InstalledTools = msg.installed
		// Update all views with the comprehensive status data
		m.dashboard.SetData(m.state.Tools, m.state.Bundles, msg.installed)
		m.toolBrowser.SetData(m.state.Tools, msg.installed)
		m.bundleExplorer.SetData(m.state.Bundles, msg.installed)
		m.healthView.SetData(m.state.Tools, msg.installed)
		return m, nil

	case toolBrowserContentLoadedMsg:
		// Tool browser content loaded in background - trigger re-render
		// This ensures the UI updates to show the loaded tools
		return m, nil

	case healthCheckTriggerMsg:
		// Trigger health checks in the health view
		if m.state.CurrentView == ViewHealth {
			// Directly start sequential health checks
			cmd := m.healthView.RunNextHealthCheck(0)
			return m, cmd
		}
		return m, nil

	// Removed startupDataLoadMsg handler - using lazy initialization instead

	// Handle health check completion messages explicitly to ensure they get routed
	// even if there are any timing or view switching issues
	case views.MemoryCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.DiskCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.InternetCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.BuildToolsCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.GitCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.PathCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.RustToolchainCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.GoToolchainCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	case views.ToolUpdatesCheckCompleteMsg:
		cmd := m.healthView.Update(msg)
		return m, cmd
	
	// Handle sequential health check continuation message
	case views.NextHealthCheckMsg:
		if m.state.CurrentView == ViewHealth {
			cmd := m.healthView.Update(msg)
			return m, cmd
		}
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
		
	case tasks.TaskUpdateMsg:
		// Handle task updates
		// Pass updates to install manager if it's the current view
		if m.state.CurrentView == ViewMonitor {
			m.installManager.HandleTaskUpdate(msg.TaskID, msg.Progress)
		}
		
		// If a task completed successfully, reload manifest data to show newly installed tools
		if msg.Status == tasks.TaskStatusCompleted {
			// Reload manifest data to refresh installed tools
			return m, tea.Batch(m.watchTaskUpdates(), m.reloadManifestData())
		}
		
		// Continue watching for more updates
		return m, m.watchTaskUpdates()
	}

	// Delegate to current view
	if m.ready {
		newModel, cmd := m.updateCurrentView(msg)
		
		// Check if we switched to health view and trigger auto-refresh
		if autoHealthCmd := m.checkForHealthViewSwitch(previousView); autoHealthCmd != nil {
			// Combine the current command with health check trigger
			if cmd != nil {
				return newModel, tea.Batch(cmd, autoHealthCmd)
			}
			return newModel, autoHealthCmd
		}
		
		return newModel, cmd
	}

	return m, nil
}

// checkForHealthViewSwitch detects when we switch TO health view and returns health check command
func (m Model) checkForHealthViewSwitch(previousView ViewType) tea.Cmd {
	// Only trigger if we just switched TO health view (not already on it)
	if previousView != ViewHealth && m.state.CurrentView == ViewHealth {
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: checkForHealthViewSwitch() - Auto-detected switch from %v to Health view, triggering health checks\n", previousView)
			debugFile.Close()
		}
		return m.healthView.RunNextHealthCheck(0)
	}
	
	if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		if previousView == ViewHealth && m.state.CurrentView == ViewHealth {
			fmt.Fprintf(debugFile, "DEBUG: checkForHealthViewSwitch() - Already on health view, no action needed\n")
		} else if m.state.CurrentView != ViewHealth {
			fmt.Fprintf(debugFile, "DEBUG: checkForHealthViewSwitch() - Current view is %v, no health checks needed\n", m.state.CurrentView)
		}
		debugFile.Close()
	}
	
	return nil
}

func (m Model) View() string {
	if m.err != nil {
		return m.errorView()
	}

	if !m.ready {
		return m.loadingView()
	}

	// Render current view
	return m.renderCurrentView()
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Comprehensive debug logging - filter out terminal escape sequences
	keyStr := msg.String()
	shouldLog := true
	
	// Filter out common terminal escape sequences that interfere with logging
	if msg.Type == tea.KeyRunes {
		// Skip escape sequences like [D, [C, ?6c, etc.
		if strings.HasPrefix(keyStr, "[") || 
		   strings.Contains(keyStr, "?") ||
		   len(keyStr) == 1 && (keyStr[0] < 32 || keyStr[0] > 126) { // non-printable chars
			shouldLog = false
		}
	}
	
	if shouldLog {
		if logFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(logFile, "DEBUG: Key pressed: '%s' (Type: %v, Alt: %v) from view: %v\n", keyStr, msg.Type, msg.Alt, m.state.CurrentView)
			logFile.Close()
		}
	}

	// Lazy initialization on first key press - ensures instant responsiveness
	var initCmd tea.Cmd
	if !m.state.Initialized {
		m.state.Initialized = true
		initCmd = tea.Batch(
			m.watchTaskUpdates(),
			m.loadInitialData(),
		)
	}

	// Helper to combine commands
	combineCmd := func(cmd tea.Cmd) tea.Cmd {
		if initCmd != nil {
			return tea.Batch(initCmd, cmd)
		}
		return cmd
	}

	// Global keybindings
	switch {
	case key.Matches(msg, keys.Quit):
		// Log session end before quitting
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "=== TUI SESSION ENDED (quit key pressed) ===\n\n")
			debugFile.Close()
		}
		return m, tea.Quit // Don't combine with initCmd on quit
	case key.Matches(msg, keys.Help):
		m.state.CurrentView = ViewHelp
		return m, combineCmd(nil)
	case key.Matches(msg, keys.Tab):
		// Cycle through views forward
		navCmd := m.nextView()
		return m, combineCmd(navCmd)
	case msg.String() == "shift+tab":
		// Cycle through views backward
		navCmd := m.previousView()
		return m, combineCmd(navCmd)
	case key.Matches(msg, keys.Right):
		// Navigate to next view with right arrow
		navCmd := m.nextView()
		return m, combineCmd(navCmd)
	case key.Matches(msg, keys.Left):
		// Navigate to previous view with left arrow
		navCmd := m.previousView()
		return m, combineCmd(navCmd)
	}

	// View-specific keybindings  
	switch keyStr {
	case "D":
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: Switching to Dashboard view\n")
			debugFile.Close()
		}
		m.state.CurrentView = ViewDashboard
		return m, combineCmd(nil)
	case "T":
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: Switching to Tool Browser view\n")
			debugFile.Close()
		}
		m.state.CurrentView = ViewToolBrowser
		// Load full content asynchronously when switching to tool browser
		// Return immediately without blocking - content loads in background
		return m, combineCmd(m.loadToolBrowserContentAsync())
	case "B":
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: Switching to Bundle Explorer view\n")
			debugFile.Close()
		}
		m.state.CurrentView = ViewBundleExplorer
		return m, combineCmd(nil)
	case "M":
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: Switching to Monitor view\n")
			debugFile.Close()
		}
		m.state.CurrentView = ViewMonitor
		return m, combineCmd(nil)
	case "C":
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: Switching to Config view\n")
			debugFile.Close()
		}
		m.state.CurrentView = ViewConfig
		return m, combineCmd(nil)
	case "H":
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: h/H pressed, switching to health view (auto-refresh will trigger)\n")
			debugFile.Close()
		}
		m.state.CurrentView = ViewHealth
		// Health checks will be triggered automatically by checkForHealthViewSwitch
		return m, combineCmd(nil)
	}

	// Always handle key presses - don't wait for ready state
	// This ensures immediate responsiveness even during initialization
	newModel, currentViewCmd := m.updateCurrentView(msg)
	return newModel, combineCmd(currentViewCmd)
}

// Define the main navigation views (excluding Help)
var mainViews = []ViewType{
	ViewDashboard,
	ViewToolBrowser,
	ViewBundleExplorer,
	ViewMonitor,
	ViewConfig,
	ViewHealth,
}

func (m *Model) nextView() tea.Cmd {
	// Don't navigate away from Help view with arrows
	if m.state.CurrentView == ViewHelp {
		return nil
	}
	
	previousView := m.state.CurrentView
	
	currentIndex := -1
	for i, v := range mainViews {
		if v == m.state.CurrentView {
			currentIndex = i
			break
		}
	}
	
	if currentIndex >= 0 {
		nextIndex := (currentIndex + 1) % len(mainViews)
		newView := mainViews[nextIndex]
		
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: nextView() - navigating from %v to %v\n", m.state.CurrentView, newView)
			debugFile.Close()
		}
		
		m.state.CurrentView = newView
		
		// Check if we switched to health view and return health check command
		if previousView != ViewHealth && m.state.CurrentView == ViewHealth {
			if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				fmt.Fprintf(debugFile, "DEBUG: Arrow navigation to health view, triggering health checks\n")
				debugFile.Close()
			}
			return m.healthView.RunNextHealthCheck(0)
		}
	}
	
	return nil
}

func (m *Model) previousView() tea.Cmd {
	// Don't navigate away from Help view with arrows
	if m.state.CurrentView == ViewHelp {
		return nil
	}
	
	previousView := m.state.CurrentView
	
	currentIndex := -1
	for i, v := range mainViews {
		if v == m.state.CurrentView {
			currentIndex = i
			break
		}
	}
	
	if currentIndex >= 0 {
		prevIndex := currentIndex - 1
		if prevIndex < 0 {
			prevIndex = len(mainViews) - 1
		}
		newView := mainViews[prevIndex]
		
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "DEBUG: previousView() - navigating from %v to %v\n", m.state.CurrentView, newView)
			debugFile.Close()
		}
		
		m.state.CurrentView = newView
		
		// Check if we switched to health view and return health check command
		if previousView != ViewHealth && m.state.CurrentView == ViewHealth {
			if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				fmt.Fprintf(debugFile, "DEBUG: Arrow navigation to health view, triggering health checks\n")
				debugFile.Close()
			}
			return m.healthView.RunNextHealthCheck(0)
		}
	}
	
	return nil
}

func (m Model) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		// Load tools configuration
		configPath := filepath.Join("config", "tools.json")
		configData, err := os.ReadFile(configPath)
		var tools []orchestrator.ToolConfig
		if err != nil {
			zlog.Warn().Err(err).Msg("Failed to load tools config")
		} else {
			var toolConfig orchestrator.Config
			if err := json.Unmarshal(configData, &toolConfig); err != nil {
				zlog.Warn().Err(err).Msg("Failed to parse tools config")
			} else {
				tools = toolConfig.Tools
			}
		}
		
		// Load bundles configuration
		bundlesPath := filepath.Join("config", "bundles.json")
		bundleData, err := os.ReadFile(bundlesPath)
		var bundles []orchestrator.BundleConfig
		if err != nil {
			zlog.Warn().Err(err).Msg("Failed to load bundles")
		} else {
			var bundleConfig orchestrator.BundleConfiguration
			if err := json.Unmarshal(bundleData, &bundleConfig); err != nil {
				zlog.Warn().Err(err).Msg("Failed to parse bundles")
			} else {
				bundles = bundleConfig.Bundles
			}
		}
		
		// Load installed tools - use fast manifest-only loading for initial display
		// The unified status will be loaded in the background after initial render
		installed := make(map[string]*manifest.InstallationRecord)
		manifestData, err := m.manifest.Load()
		if err == nil && manifestData != nil && manifestData.Installations != nil {
			installed = manifestData.Installations
		}

		return dataLoadedMsg{
			tools:     tools,
			bundles:   bundles,
			installed: installed,
		}
	}
}

// loadToolBrowserContentAsync loads tool browser content asynchronously  
func (m Model) loadToolBrowserContentAsync() tea.Cmd {
	return func() tea.Msg {
		// This runs in a separate goroutine - completely non-blocking
		// Give the UI thread a chance to render first
		m.toolBrowser.LoadFullContent()
		return toolBrowserContentLoadedMsg{}
	}
}

// loadUnifiedStatusBackground loads unified status in background without blocking UI
func (m Model) loadUnifiedStatusBackground() tea.Cmd {
	return func() tea.Msg {
		// Load unified status service (this is the expensive operation)
		unifiedStatus, err := status.NewUnifiedStatusService()
		if err != nil {
			zlog.Warn().Err(err).Msg("Background unified status loading failed")
			return nil // No update needed
		}
		
		// Get all tool status from unified service
		allStatus, err := unifiedStatus.GetAllToolsStatus()
		if err != nil {
			zlog.Warn().Err(err).Msg("Background unified status check failed")
			return nil // No update needed
		}
		
		// Convert unified status to manifest records for TUI compatibility
		installed := make(map[string]*manifest.InstallationRecord)
		for toolName, toolStatus := range allStatus {
			if toolStatus.Installed {
				record := &manifest.InstallationRecord{
					Version:     toolStatus.Version,
					BinaryPaths: toolStatus.BinaryPaths,
				}
				if toolStatus.Source == "gearbox" {
					record.Method = "gearbox"
				} else {
					record.Method = "system" // For pre-existing tools
				}
				installed[toolName] = record
			}
		}
		
		return unifiedStatusLoadedMsg{
			installed: installed,
		}
	}
}

// runHealthChecksAsync is deprecated - now using individual async checks

// reloadManifestData reloads the installation data using unified status
func (m Model) reloadManifestData() tea.Cmd {
	return func() tea.Msg {
		// Load installed tools using unified status service
		installed := make(map[string]*manifest.InstallationRecord)
		
		unifiedStatus, err := status.NewUnifiedStatusService()
		if err != nil {
			zlog.Warn().Err(err).Msg("Failed to create unified status service during reload")
			// Fallback to manifest-only loading
			manifestData, err := m.manifest.Load()
			if err == nil && manifestData != nil && manifestData.Installations != nil {
				installed = manifestData.Installations
			}
		} else {
			// Get all tool status from unified service
			allStatus, err := unifiedStatus.GetAllToolsStatus()
			if err != nil {
				zlog.Warn().Err(err).Msg("Failed to get unified status during reload")
				manifestData, err := m.manifest.Load()
				if err == nil && manifestData != nil && manifestData.Installations != nil {
					installed = manifestData.Installations
				}
			} else {
				// Convert unified status to manifest records for TUI compatibility
				for toolName, toolStatus := range allStatus {
					if toolStatus.Installed {
						record := &manifest.InstallationRecord{
							Version:     toolStatus.Version,
							BinaryPaths: toolStatus.BinaryPaths,
						}
						if toolStatus.Source == "gearbox" {
							record.Method = "gearbox"
						} else {
							record.Method = "system"
						}
						installed[toolName] = record
					}
				}
			}
		}

		return manifestReloadedMsg{
			installed: installed,
		}
	}
}

func (m Model) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state.CurrentView {
	case ViewDashboard:
		// Dashboard view doesn't need updates for now
		return m, nil
	case ViewToolBrowser:
		// Check for install key before delegating
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "i" {
			// Add selected tools to task manager
			selectedTools := m.toolBrowser.GetSelectedTools()
			var addedTaskIDs []string
			for _, toolName := range selectedTools {
				// Find the tool config
				for _, tool := range m.state.Tools {
					if tool.Name == toolName {
						// Add to task manager
						taskID := m.taskManager.AddTask(tool, "standard")
						// Add task ID to install manager
						m.installManager.AddTaskID(taskID)
						// Start the task immediately
						m.taskManager.StartTask(taskID)
						addedTaskIDs = append(addedTaskIDs, taskID)
						break
					}
				}
			}
			if len(selectedTools) > 0 {
				// Clear selection
				m.toolBrowser.ClearSelection()
				// Switch to monitor view to show progress
				m.state.CurrentView = ViewMonitor
				return m, nil
			}
		}
		// Delegate to tool browser
		cmd := m.toolBrowser.Update(msg)
		return m, cmd
	case ViewBundleExplorer:
		// Check for install key before delegating
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "i" {
			// Get selected bundle
			selectedBundle := m.bundleExplorer.GetSelectedBundle()
			if selectedBundle != nil {
				// Get uninstalled tools from the bundle
				uninstalledTools := m.bundleExplorer.GetUninstalledTools(selectedBundle)
				
				// Add each uninstalled tool to task manager
				for _, toolName := range uninstalledTools {
					// Find the tool config
					for _, tool := range m.state.Tools {
						if tool.Name == toolName {
							// Add to task manager
							taskID := m.taskManager.AddTask(tool, "standard")
							// Add task ID to install manager
							m.installManager.AddTaskID(taskID)
							// Start the task immediately
							m.taskManager.StartTask(taskID)
							break
						}
					}
				}
				
				if len(uninstalledTools) > 0 {
					// Switch to monitor view
					m.state.CurrentView = ViewMonitor
					return m, nil
				}
			}
		}
		// Delegate to bundle explorer
		cmd := m.bundleExplorer.Update(msg)
		return m, cmd
	case ViewMonitor:
		// Delegate to installation monitor
		cmd := m.installManager.Update(msg)
		return m, cmd
	case ViewConfig:
		// Delegate to config view
		cmd := m.configView.Update(msg)
		return m, cmd
	case ViewHealth:
		// Delegate to health view
		cmd := m.healthView.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m Model) renderCurrentView() string {
	var viewContent string
	
	switch m.state.CurrentView {
	case ViewDashboard:
		viewContent = m.renderDashboard()
	case ViewToolBrowser:
		viewContent = m.renderToolBrowser()
	case ViewBundleExplorer:
		viewContent = m.renderBundleExplorer()
	case ViewMonitor:
		viewContent = m.renderMonitor()
	case ViewConfig:
		viewContent = m.renderConfig()
	case ViewHealth:
		viewContent = m.renderHealth()
	case ViewHelp:
		viewContent = m.renderHelp()
	default:
		viewContent = m.renderDashboard()
	}

	// Add navigation bar
	navBar := m.renderNavigationBar()
	statusBar := m.renderStatusBar()
	
	// Layout system should render exactly to its bounds - no additional constraint
	content := viewContent
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		navBar,
		content,
		statusBar,
	)
}

// constrainHeight ensures content doesn't exceed the given height
func (m Model) constrainHeight(content string, maxHeight int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxHeight {
		return content
	}
	
	// Truncate to fit
	truncated := strings.Join(lines[:maxHeight], "\n")
	return truncated
}

// constrainHeightPreserveScrolling constrains total height but allows internal scrolling
func (m Model) constrainHeightPreserveScrolling(content string, maxHeight int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxHeight {
		return content
	}
	
	// Simply truncate to maxHeight - the layout system should handle this properly
	// This preserves the viewport's internal scrolling capability
	return strings.Join(lines[:maxHeight], "\n")
}

// smartConstrainHeight constrains content while preserving the help bar at the bottom
func (m Model) smartConstrainHeight(content string, maxHeight int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxHeight {
		return content
	}
	
	// Find the help bar (last non-empty line that contains navigation keys)
	helpBarIndex := -1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && (strings.Contains(line, "[↑/↓]") || strings.Contains(line, "Navigate")) {
			helpBarIndex = i
			break
		}
	}
	
	if helpBarIndex == -1 {
		// No help bar found, use regular constraint
		return m.constrainHeight(content, maxHeight)
	}
	
	// Calculate how much content we can keep before help bar
	helpBarLines := lines[helpBarIndex:]
	availableForContent := maxHeight - len(helpBarLines)
	
	if availableForContent <= 0 {
		// Not enough space, just show help bar
		if len(helpBarLines) <= maxHeight {
			return strings.Join(helpBarLines, "\n")
		} else {
			return strings.Join(helpBarLines[:maxHeight], "\n")
		}
	}
	
	// Keep content + help bar
	contentLines := lines[:availableForContent]
	result := append(contentLines, helpBarLines...)
	return strings.Join(result, "\n")
}

func (m Model) renderNavigationBar() string {
	baseStyle := lipgloss.NewStyle().
		Padding(0, 2)
	
	activeStyle := baseStyle.Copy().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true)
	
	inactiveStyle := baseStyle.Copy().
		Foreground(lipgloss.Color("246"))

	tabs := []struct {
		label string
		view  ViewType
	}{
		{"[D]ashboard", ViewDashboard},
		{"[T]ools", ViewToolBrowser},
		{"[B]undles", ViewBundleExplorer},
		{"[M]onitor", ViewMonitor},
		{"[C]onfig", ViewConfig},
		{"[H]ealth", ViewHealth},
	}

	var renderedTabs []string
	for _, tab := range tabs {
		if tab.view == m.state.CurrentView {
			renderedTabs = append(renderedTabs, activeStyle.Render(tab.label))
		} else {
			renderedTabs = append(renderedTabs, inactiveStyle.Render(tab.label))
		}
	}

	navBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	return lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Width(m.width).
		Render(navBar)
}

func (m Model) renderStatusBar() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("246")).
		Width(m.width).
		Padding(0, 2)

	left := fmt.Sprintf("Tools: %d/%d installed", len(m.state.InstalledTools), len(m.state.Tools))
	right := "[?] Help  [q] Quit"
	
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 0 {
		gap = 0
	}
	
	return style.Render(left + lipgloss.NewStyle().Width(gap).Render(" ") + right)
}

func (m Model) renderDashboard() string {
	return m.dashboard.Render()
}

func (m Model) renderToolBrowser() string {
	return m.toolBrowser.Render()
}

func (m Model) renderBundleExplorer() string {
	return m.bundleExplorer.Render()
}

func (m Model) renderMonitor() string {
	return m.installManager.Render()
}

func (m Model) renderConfig() string {
	return m.configView.Render()
}

func (m Model) renderHealth() string {
	return m.healthView.Render()
}

func (m Model) renderHelp() string {
	helpText := `
Gearbox TUI Help

Navigation:
  ←/→       - Switch between views
  Tab       - Next view
  Shift+Tab - Previous view
  ↑/↓       - Navigate lists
  Enter     - Select/Confirm
  Esc       - Go back
  /         - Search
  q         - Quit

Views:
  D - Dashboard         - Overview and quick actions
  T - Tool Browser      - Browse and install individual tools
  B - Bundle Explorer   - Explore curated tool collections
  M - Monitor           - Monitor installation tasks and progress
  C - Configuration     - Configure Gearbox settings
  H - Health Monitor    - System and tool health checks
  ? - Help             - This help screen

Tool Browser:
  Space     - Select/Deselect tool
  i         - Install selected tools (starts immediately)
  c         - Cycle categories
  p         - Toggle preview

Bundle Explorer:
  Enter     - Expand/Collapse bundle
  i         - Install bundle (starts immediately)
  c         - Cycle categories

Monitor:
  s         - Start queued installations
  c         - Cancel current task
  o         - Toggle output display

Configuration:
  Enter     - Edit setting
  r         - Reset to default
  s         - Save all settings

Health Monitor:
  r         - Run health checks
  d         - Toggle details
  a         - Toggle auto-refresh
`
	
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height - 4).
		Padding(2).
		Render(helpText)
}

func (m Model) loadingView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Loading Gearbox TUI...")
}

func (m Model) errorView() string {
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(lipgloss.Color("9")).
		Render(fmt.Sprintf("Error: %v", m.err))
}

// Messages
type dataLoadedMsg struct {
	tools     []orchestrator.ToolConfig
	bundles   []orchestrator.BundleConfig
	installed map[string]*manifest.InstallationRecord
}

type manifestReloadedMsg struct {
	installed map[string]*manifest.InstallationRecord
}

type unifiedStatusLoadedMsg struct {
	installed map[string]*manifest.InstallationRecord
}

type toolBrowserContentLoadedMsg struct{}

type healthCheckTriggerMsg struct{}

type errMsg struct {
	err error
}

// Key bindings
type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Enter key.Binding
	Space key.Binding
	Tab   key.Binding
	Quit  key.Binding
	Help  key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// Run starts the TUI application
func Run() error {
	model, err := NewModel()
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// Log session end on error
		if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			fmt.Fprintf(debugFile, "=== TUI SESSION ENDED (error: %v) ===\n\n", err)
			debugFile.Close()
		}
		return errors.Wrap(err, "failed to run TUI")
	}

	// Log normal session end
	if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, "=== TUI SESSION ENDED (normal exit) ===\n\n")
		debugFile.Close()
	}

	return nil
}