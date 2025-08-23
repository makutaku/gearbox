package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// Options contains configuration for TUI behavior
type Options struct {
	DemoMode     bool   // Use mock data, simulate installations
	TestMode     bool   // For automated testing
	TestScenario string // Specific test scenario to run
}

// RunWithOptions starts the TUI with specific configuration
func RunWithOptions(opts Options) error {
	if opts.TestMode {
		return runTestScenario(opts.TestScenario)
	}
	
	model, err := NewModelWithOptions(opts)
	if err != nil {
		return err
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// DemoModel represents a TUI model for testing with mock data
type DemoModel struct {
	tools          []orchestrator.ToolConfig
	bundles        []orchestrator.BundleConfig
	installedTools map[string]*manifest.InstallationRecord
	taskManager    *MockTaskManager
	
	// Views
	dashboard      *views.Dashboard
	toolBrowser    *views.ToolBrowserNew
	bundleExplorer *views.BundleExplorerNew
	installManager *views.InstallManagerNew
	configView     *views.ConfigView
	healthView     *views.HealthView
	
	// Navigation
	navigator *NavigationHandler
	
	// UI state
	state   *AppState
	width   int
	height  int
	ready   bool
	err     error
}

// NewModelWithOptions creates a new TUI model with specific options
func NewModelWithOptions(opts Options) (TUIModel, error) {
	if opts.DemoMode {
		return NewDemoModel()
	}
	
	// Use regular model for normal operation
	return NewModel()
}

// NewDemoModel creates a TUI model with mock data for safe testing
func NewDemoModel() (TUIModel, error) {
	// Create mock orchestrator
	mockOrch := &MockOrchestrator{
		tools:   generateMockTools(),
		bundles: generateMockBundles(),
	}
	
	// Create mock manifest
	mockManifest := &MockManifest{
		installed: generateMockInstalledTools(),
	}
	
	// Create demo task manager
	mockTaskManager := NewMockTaskManager()
	
	// Create demo task provider
	taskProvider := NewMockTaskProvider(mockTaskManager)
	
	// Create app state
	state := NewAppState()
	
	// Create views
	dashboard := views.NewDashboard()
	toolBrowser := views.NewToolBrowserNew()
	bundleExplorer := views.NewBundleExplorerNew() 
	installManager := views.NewInstallManagerNew()
	installManager.SetTaskProvider(taskProvider)
	configView := views.NewConfigView()
	healthView := views.NewHealthView()
	
	// Create demo model with mock data
	model := &DemoModel{
		tools:          mockOrch.tools,
		bundles:        mockOrch.bundles,
		installedTools: mockManifest.installed,
		taskManager:    mockTaskManager,
		dashboard:      dashboard,
		toolBrowser:    toolBrowser,
		bundleExplorer: bundleExplorer,
		installManager: installManager,
		configView:     configView,
		healthView:     healthView,
		navigator:      NewNavigationHandler(),
		state:          state,
		ready:          true, // Start ready for testing
	}
	
	// Initialize views with data
	model.dashboard.SetData(model.tools, model.bundles, model.installedTools)
	model.toolBrowser.SetData(model.tools, model.installedTools)
	model.bundleExplorer.SetData(model.bundles, model.installedTools)
	model.healthView.SetData(model.tools, model.installedTools)
	
	return model, nil
}

// Implement tea.Model interface for DemoModel

func (m DemoModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
	)
}

func (m DemoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}
		// Update view sizes
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
	
	// Handle health check messages (same as main Model)
	default:
		// Delegate to current view for health check messages
		if m.ready && m.state.CurrentView == ViewHealth {
			// Let health view handle its own messages
			cmd := m.healthView.Update(msg)
			return m, cmd
		}
	}

	// Delegate to current view
	if m.ready {
		return m.updateCurrentView(msg)
	}

	return m, nil
}

func (m DemoModel) View() string {
	if m.err != nil {
		return m.errorView()
	}

	if !m.ready {
		return m.loadingView()
	}

	// Render current view
	return m.renderCurrentView()
}

// TUIModel interface implementation for DemoModel

// SetSize updates the model dimensions and all views
func (m *DemoModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	
	// Update view sizes
	viewHeight := max(5, height - 2) // Minimum 5 lines for views
	m.dashboard.SetSize(width, viewHeight)
	m.toolBrowser.SetSize(width, viewHeight)
	m.bundleExplorer.SetSize(width, viewHeight)
	m.installManager.SetSize(width, viewHeight)
	m.configView.SetSize(width, viewHeight)
	m.healthView.SetSize(width, viewHeight)
}

// GetSize returns the current model dimensions
func (m DemoModel) GetSize() (width, height int) {
	return m.width, m.height
}

// GetCurrentView returns the current active view
func (m DemoModel) GetCurrentView() ViewType {
	return m.state.CurrentView
}

// SetCurrentView sets the current active view
func (m *DemoModel) SetCurrentView(view ViewType) {
	m.state.CurrentView = view
}

// IsReady returns whether the model is ready for interaction
func (m DemoModel) IsReady() bool {
	return m.ready
}

// GetState returns the app state
func (m DemoModel) GetState() *AppState {
	return m.state
}

// GetError returns the current error if any
func (m DemoModel) GetError() error {
	return m.err
}

// SwitchToView switches to the specified view
func (m *DemoModel) SwitchToView(view ViewType) tea.Cmd {
	m.state.CurrentView = view
	return nil
}

// HandleNavigation handles navigation key presses
func (m DemoModel) HandleNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handleKeyPress(msg)
}

// View access methods for testing and inspection

// GetDashboard returns the dashboard view
func (m DemoModel) GetDashboard() DashboardService {
	return NewDashboardAdapter(m.dashboard)
}

// GetToolBrowser returns the tool browser view
func (m DemoModel) GetToolBrowser() ToolBrowserService {
	return NewToolBrowserAdapter(m.toolBrowser)
}

// GetBundleExplorer returns the bundle explorer view
func (m DemoModel) GetBundleExplorer() BundleExplorerService {
	return NewBundleExplorerAdapter(m.bundleExplorer)
}

// GetInstallManager returns the install manager view
func (m DemoModel) GetInstallManager() InstallManagerService {
	return NewInstallManagerAdapter(m.installManager)
}

// GetConfigView returns the config view
func (m DemoModel) GetConfigView() ConfigService {
	return NewConfigAdapter(m.configView)
}

// GetHealthView returns the health view
func (m DemoModel) GetHealthView() HealthService {
	return NewHealthAdapter(m.healthView)
}

// DemoModel helper methods

func (m DemoModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Track previous view to detect when we switch TO health view
	previousView := m.state.CurrentView
	
	// Handle navigation first using NavigationHandler
	newView, navCmd, handled := m.navigator.HandleKeyPress(msg, m.state.CurrentView)
	if handled {
		// Check for view change
		if newView != m.state.CurrentView {
			m.state.CurrentView = newView
			
			// Check if we switched to health view and trigger auto-refresh
			if autoHealthCmd := m.checkForHealthViewSwitch(previousView); autoHealthCmd != nil {
				// Combine the navigation command with health check trigger
				if navCmd != nil {
					return m, tea.Batch(navCmd, autoHealthCmd)
				}
				return m, autoHealthCmd
			}
			
			return m, navCmd
		}
		return m, navCmd
	}

	// If the key wasn't handled by navigation, delegate to the current view
	if m.ready {
		newModel, currentViewCmd := m.updateCurrentView(msg)
		
		// Check if we switched to health view and trigger auto-refresh
		if autoHealthCmd := m.checkForHealthViewSwitch(previousView); autoHealthCmd != nil {
			// Combine the current view command with health check trigger
			if currentViewCmd != nil {
				return newModel, tea.Batch(currentViewCmd, autoHealthCmd)
			}
			return newModel, autoHealthCmd
		}
		
		return newModel, currentViewCmd
	}

	return m, nil
}

// checkForHealthViewSwitch detects when we switch TO health view and returns health check command
func (m DemoModel) checkForHealthViewSwitch(previousView ViewType) tea.Cmd {
	// Only trigger if we just switched TO health view (not already on it)
	if previousView != ViewHealth && m.state.CurrentView == ViewHealth {
		// Trigger health checks
		return m.healthView.RunNextHealthCheck(0)
	}
	
	return nil
}


func (m DemoModel) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state.CurrentView {
	case ViewDashboard:
		// Dashboard view doesn't need updates for now
		return m, nil
	case ViewToolBrowser:
		// Check for install key before delegating
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "i" {
			// Add selected tools to task manager
			selectedTools := m.toolBrowser.GetSelectedTools()
			for _, toolName := range selectedTools {
				// Find the tool config
				for _, tool := range m.tools {
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
		// Delegate to bundle explorer
		cmd := m.bundleExplorer.Update(msg)
		return m, cmd
	case ViewMonitor:
		// Delegate to install manager
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

func (m DemoModel) renderCurrentView() string {
	var viewContent string
	
	switch m.state.CurrentView {
	case ViewDashboard:
		viewContent = m.dashboard.Render()
	case ViewToolBrowser:
		viewContent = m.toolBrowser.Render()
	case ViewBundleExplorer:
		viewContent = m.bundleExplorer.Render()
	case ViewMonitor:
		viewContent = m.installManager.Render()
	case ViewConfig:
		viewContent = m.configView.Render()
	case ViewHealth:
		viewContent = m.healthView.Render()
	case ViewHelp:
		viewContent = m.renderHelp()
	default:
		viewContent = m.dashboard.Render()
	}

	// Add navigation bar
	navBar := m.renderNavigationBar()
	statusBar := m.renderStatusBar()
	
	return navBar + "\n" + viewContent + "\n" + statusBar
}

func (m DemoModel) renderNavigationBar() string {
	return "[D]ashboard | [T]ools | [B]undles | [M]onitor | [C]onfig | [H]ealth | [?] Help | [q] Quit"
}

func (m DemoModel) renderStatusBar() string {
	return fmt.Sprintf("Tools: %d/%d installed | Demo Mode", len(m.installedTools), len(m.tools))
}

func (m DemoModel) renderHelp() string {
	return `
Gearbox TUI Demo Help

Navigation:
  Tab       - Next view
  Shift+Tab - Previous view
  ↑/↓       - Navigate lists
  Enter     - Select/Confirm
  /         - Search
  q         - Quit

Views:
  D - Dashboard         T - Tool Browser      B - Bundle Explorer
  M - Monitor           C - Configuration     H - Health Monitor
  ? - Help

This is DEMO MODE - no real installations will be performed.
All operations are simulated for testing purposes.
`
}

func (m DemoModel) loadingView() string {
	return "Loading Demo TUI..."
}

func (m DemoModel) errorView() string {
	return fmt.Sprintf("Demo Error: %v", m.err)
}

// runTestScenario runs automated test scenarios
func runTestScenario(scenario string) error {
	fmt.Printf("🧪 Running test scenario: %s\n", scenario)
	
	switch scenario {
	case "basic-nav":
		return testBasicNavigation()
	case "tool-install":
		return testToolInstallation()
	case "bundle-install":
		return testBundleInstallation()
	default:
		return fmt.Errorf("unknown test scenario: %s", scenario)
	}
}

// Test scenarios
func testBasicNavigation() error {
	fmt.Println("Testing basic navigation...")
	
	// Simulate TUI creation and navigation
	model, err := NewDemoModel()
	if err != nil {
		return err
	}
	
	// Simulate key sequences
	keySequence := []string{
		"tab", "tab", "tab", // Navigate between views
		"down", "down", "up", // Navigate within view
		"q", // Quit
	}
	
	for _, key := range keySequence {
		fmt.Printf("  Simulating key: %s\n", key)
		time.Sleep(100 * time.Millisecond) // Simulate user input timing
		
		// In a real test, we would send the key to the model and verify state
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		_, _ = model.Update(msg)
	}
	
	fmt.Println("✅ Basic navigation test completed")
	return nil
}

func testToolInstallation() error {
	fmt.Println("Testing tool installation flow...")
	
	model, err := NewDemoModel()
	if err != nil {
		return err
	}
	
	// Simulate tool installation flow
	steps := []struct {
		action string
		key    string
	}{
		{"Navigate to Tools view", "t"},
		{"Move to second tool", "down"},
		{"Select tool", " "},
		{"Install selected tool (starts automatically)", "i"},
		{"Now in Monitor view showing progress", ""},
	}
	
	for _, step := range steps {
		if step.key != "" {
			fmt.Printf("  %s (key: %s)\n", step.action, step.key)
			time.Sleep(100 * time.Millisecond)
			
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(step.key)}
			_, _ = model.Update(msg)
		} else {
			fmt.Printf("  %s\n", step.action)
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	fmt.Println("✅ Tool installation test completed")
	return nil
}

func testBundleInstallation() error {
	fmt.Println("Testing bundle installation flow...")
	
	model, err := NewDemoModel()
	if err != nil {
		return err
	}
	
	// Simulate bundle installation flow
	steps := []struct {
		action string
		key    string
	}{
		{"Navigate to Bundles view", "b"},
		{"Move to second bundle", "down"},
		{"Expand bundle details", "enter"},
		{"Install bundle", "i"},
		{"Switch to Monitor view", "tab"},
		{"Start installation", "s"},
	}
	
	for _, step := range steps {
		fmt.Printf("  %s (key: %s)\n", step.action, step.key)
		time.Sleep(100 * time.Millisecond)
		
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(step.key)}
		_, _ = model.Update(msg)
	}
	
	fmt.Println("✅ Bundle installation test completed")
	return nil
}

// generateMockTools creates realistic mock tool data for testing
func generateMockTools() []orchestrator.ToolConfig {
	return []orchestrator.ToolConfig{
		{
			Name:        "fd",
			Description: "A fast alternative to find",
			Category:    "Core",
			Language:    "rust",
			Repository:  "https://github.com/sharkdp/fd",
			BinaryName:  "fd",
		},
		{
			Name:        "ripgrep",
			Description: "Fast line-oriented search tool",
			Category:    "Core", 
			Language:    "rust",
			Repository:  "https://github.com/BurntSushi/ripgrep",
			BinaryName:  "rg",
		},
		{
			Name:        "fzf",
			Description: "Command-line fuzzy finder",
			Category:    "Core",
			Language:    "go",
			Repository:  "https://github.com/junegunn/fzf",
			BinaryName:  "fzf",
		},
		{
			Name:        "bat",
			Description: "Cat clone with syntax highlighting",
			Category:    "Text",
			Language:    "rust",
			Repository:  "https://github.com/sharkdp/bat",
			BinaryName:  "bat",
		},
		{
			Name:        "eza",
			Description: "Modern replacement for ls",
			Category:    "System",
			Language:    "rust",
			Repository:  "https://github.com/eza-community/eza",
			BinaryName:  "eza",
		},
	}
}

// generateMockBundles creates realistic mock bundle data for testing
func generateMockBundles() []orchestrator.BundleConfig {
	return []orchestrator.BundleConfig{
		{
			Name:        "beginner",
			Description: "Essential tools for new developers",
			Category:    "foundation",
			Tools:       []string{"fd", "ripgrep", "fzf"},
		},
		{
			Name:        "rust-dev",
			Description: "Rust development environment",
			Category:    "language",
			Tools:       []string{"fd", "ripgrep", "bat", "eza"},
		},
		{
			Name:        "core-tools",
			Description: "Core command-line utilities",
			Category:    "workflow",
			Tools:       []string{"fd", "ripgrep", "fzf", "bat"},
		},
	}
}

// generateMockInstalledTools creates mock installation records
func generateMockInstalledTools() map[string]*manifest.InstallationRecord {
	return map[string]*manifest.InstallationRecord{
		"fd": {
			Method:      manifest.MethodSourceBuild,
			Version:     "8.6.0",
			InstalledAt: time.Now().Add(-24 * time.Hour),
			BinaryPaths: []string{"/usr/local/bin/fd"},
			UserRequested: true,
		},
		"ripgrep": {
			Method:      manifest.MethodSourceBuild,
			Version:     "13.0.0",
			InstalledAt: time.Now().Add(-12 * time.Hour),
			BinaryPaths: []string{"/usr/local/bin/rg"},
			UserRequested: true,
		},
	}
}