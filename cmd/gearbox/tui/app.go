package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

type Model struct {
	orchestrator *orchestrator.Orchestrator
	manifest     *manifest.Manager
	state        *AppState
	taskManager  *tasks.TaskManager
	
	// Views
	dashboard      *views.Dashboard
	toolBrowser    *views.ToolBrowser
	bundleExplorer *views.BundleExplorer
	installManager *views.InstallManager
	configView     *views.ConfigView
	healthView     *views.HealthView
	
	// UI state
	width    int
	height   int
	ready    bool
	err      error
}

func NewModel() (*Model, error) {
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
	toolBrowser := views.NewToolBrowser()
	bundleExplorer := views.NewBundleExplorer()
	installManager := views.NewInstallManager()
	installManager.SetTaskProvider(taskProvider)
	configView := views.NewConfigView()
	healthView := views.NewHealthView()
	
	return &Model{
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
	}, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadInitialData(),
		tea.EnterAltScreen,
		m.watchTaskUpdates(),
	)
}

func (m Model) watchTaskUpdates() tea.Cmd {
	return m.taskManager.WatchUpdates()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.ready = true
		}
		// Update view sizes
		m.dashboard.SetSize(m.width, m.height-2) // Account for nav and status bars
		m.toolBrowser.SetSize(m.width, m.height-2)
		m.bundleExplorer.SetSize(m.width, m.height-2)
		m.installManager.SetSize(m.width, m.height-2)
		m.configView.SetSize(m.width, m.height-2)
		m.healthView.SetSize(m.width, m.height-2)
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
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
		
	case tasks.TaskUpdateMsg:
		// Handle task updates
		// Pass updates to install manager if it's the current view
		if m.state.CurrentView == ViewInstallManager {
			m.installManager.HandleTaskUpdate(msg.TaskID, msg.Progress)
		}
		// Continue watching for more updates
		return m, m.watchTaskUpdates()
	}

	// Delegate to current view
	if m.ready {
		return m.updateCurrentView(msg)
	}

	return m, nil
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
	// Global keybindings
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Help):
		m.state.CurrentView = ViewHelp
		return m, nil
	case key.Matches(msg, keys.Tab):
		// Cycle through views
		m.nextView()
		return m, nil
	}

	// View-specific keybindings
	switch msg.String() {
	case "d", "D":
		m.state.CurrentView = ViewDashboard
		return m, nil
	case "t", "T":
		m.state.CurrentView = ViewToolBrowser
		return m, nil
	case "b", "B":
		m.state.CurrentView = ViewBundleExplorer
		return m, nil
	case "i", "I":
		m.state.CurrentView = ViewInstallManager
		return m, nil
	case "c", "C":
		m.state.CurrentView = ViewConfig
		return m, nil
	case "h", "H":
		m.state.CurrentView = ViewHealth
		return m, nil
	}

	// If the key wasn't handled above, delegate to the current view
	// This allows arrow keys and other navigation to work
	if m.ready {
		return m.updateCurrentView(msg)
	}

	return m, nil
}

func (m *Model) nextView() {
	views := []ViewType{
		ViewDashboard,
		ViewToolBrowser,
		ViewBundleExplorer,
		ViewInstallManager,
		ViewConfig,
		ViewHealth,
		ViewHelp,
	}
	
	currentIndex := -1
	for i, v := range views {
		if v == m.state.CurrentView {
			currentIndex = i
			break
		}
	}
	
	if currentIndex >= 0 {
		nextIndex := (currentIndex + 1) % len(views)
		m.state.CurrentView = views[nextIndex]
	}
}

func (m Model) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		// Load tools configuration
		configPath := filepath.Join("config", "tools.json")
		configData, err := os.ReadFile(configPath)
		var tools []orchestrator.ToolConfig
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load tools config")
		} else {
			var toolConfig orchestrator.Config
			if err := json.Unmarshal(configData, &toolConfig); err != nil {
				log.Warn().Err(err).Msg("Failed to parse tools config")
			} else {
				tools = toolConfig.Tools
			}
		}
		
		// Load bundles configuration
		bundlesPath := filepath.Join("config", "bundles.json")
		bundleData, err := os.ReadFile(bundlesPath)
		var bundles []orchestrator.BundleConfig
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load bundles")
		} else {
			var bundleConfig orchestrator.BundleConfiguration
			if err := json.Unmarshal(bundleData, &bundleConfig); err != nil {
				log.Warn().Err(err).Msg("Failed to parse bundles")
			} else {
				bundles = bundleConfig.Bundles
			}
		}
		
		// Load installed tools from manifest
		manifestData, err := m.manifest.Load()
		installed := make(map[string]*manifest.InstallationRecord)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load manifest")
		} else if manifestData != nil && manifestData.Installations != nil {
			installed = manifestData.Installations
		}

		return dataLoadedMsg{
			tools:     tools,
			bundles:   bundles,
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
			for _, toolName := range selectedTools {
				// Find the tool config
				for _, tool := range m.state.Tools {
					if tool.Name == toolName {
						// Add to task manager
						taskID := m.taskManager.AddTask(tool, "standard")
						// Add task ID to install manager
						m.installManager.AddTaskID(taskID)
						break
					}
				}
			}
			if len(selectedTools) > 0 {
				// Clear selection
				m.toolBrowser.ClearSelection()
				// Switch to install manager view
				m.state.CurrentView = ViewInstallManager
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
							break
						}
					}
				}
				
				if len(uninstalledTools) > 0 {
					// Switch to install manager view
					m.state.CurrentView = ViewInstallManager
					return m, nil
				}
			}
		}
		// Delegate to bundle explorer
		cmd := m.bundleExplorer.Update(msg)
		return m, cmd
	case ViewInstallManager:
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

func (m Model) renderCurrentView() string {
	var content string
	
	switch m.state.CurrentView {
	case ViewDashboard:
		content = m.renderDashboard()
	case ViewToolBrowser:
		content = m.renderToolBrowser()
	case ViewBundleExplorer:
		content = m.renderBundleExplorer()
	case ViewInstallManager:
		content = m.renderInstallManager()
	case ViewConfig:
		content = m.renderConfig()
	case ViewHealth:
		content = m.renderHealth()
	case ViewHelp:
		content = m.renderHelp()
	default:
		content = m.renderDashboard()
	}

	// Add navigation bar
	navBar := m.renderNavigationBar()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		navBar,
		content,
		m.renderStatusBar(),
	)
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
		{"[I]nstall", ViewInstallManager},
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

func (m Model) renderInstallManager() string {
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
  Tab       - Switch between views
  ↑/↓       - Navigate lists
  Enter     - Select/Confirm
  Esc       - Go back
  /         - Search
  q         - Quit

Views:
  D - Dashboard         - Overview and quick actions
  T - Tool Browser      - Browse and install individual tools
  B - Bundle Explorer   - Explore curated tool collections
  I - Install Manager   - Manage installation queue and progress
  C - Configuration     - Configure Gearbox settings
  H - Health Monitor    - System and tool health checks
  ? - Help             - This help screen

Tool Browser:
  Space     - Select/Deselect tool
  i         - Install selected tools
  c         - Cycle categories
  p         - Toggle preview

Bundle Explorer:
  Enter     - Expand/Collapse bundle
  i         - Install bundle
  c         - Cycle categories

Install Manager:
  s         - Start installations
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
		return errors.Wrap(err, "failed to run TUI")
	}

	return nil
}