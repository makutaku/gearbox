package tui

import (
	"reflect"
	
	"gearbox/cmd/gearbox/tui/tasks"
	"gearbox/cmd/gearbox/tui/views"
	"gearbox/pkg/manifest"
	"gearbox/pkg/orchestrator"
)

// Dependencies holds all the dependencies for the TUI application
type Dependencies struct {
	// Core services
	Orchestrator *orchestrator.Orchestrator
	Manifest     *manifest.Manager
	TaskManager  *tasks.TaskManager
	
	// Handlers  
	Navigator *NavigationHandler
	Router    *MessageRouter
	
	// Views
	Dashboard      *views.Dashboard
	ToolBrowser    *views.ToolBrowserNew
	BundleExplorer *views.BundleExplorerNew
	InstallManager *views.InstallManagerNew
	ConfigView     *views.ConfigView
	HealthView     *views.HealthView
}

// DependencyFactory creates and manages dependencies
type DependencyFactory struct {
	orchestratorOpts orchestrator.InstallationOptions
}

// NewDependencyFactory creates a new dependency factory
func NewDependencyFactory(opts orchestrator.InstallationOptions) *DependencyFactory {
	return &DependencyFactory{
		orchestratorOpts: opts,
	}
}

// CreateDependencies creates all dependencies for the TUI application
func (f *DependencyFactory) CreateDependencies() (*Dependencies, error) {
	// Create core services
	orch, err := orchestrator.NewOrchestrator(f.orchestratorOpts)
	if err != nil {
		return nil, err
	}
	
	manifestMgr := manifest.NewManager()
	taskManager := tasks.NewTaskManager(orch, DefaultMaxParallel)
	
	// Create handlers
	navigator := NewNavigationHandler()
	
	// Create views
	dashboard := views.NewDashboard()
	toolBrowser := views.NewToolBrowserNew()
	bundleExplorer := views.NewBundleExplorerNew()
	installManager := views.NewInstallManagerNew()
	configView := views.NewConfigView()
	healthView := views.NewHealthView()
	
	// Setup task provider for install manager
	taskProvider := NewTaskManagerProvider(taskManager)
	installManager.SetTaskProvider(taskProvider)
	
	// Create and setup message router
	router := f.createMessageRouter(healthView)
	
	return &Dependencies{
		Orchestrator:   orch,
		Manifest:       manifestMgr,
		TaskManager:    taskManager,
		Navigator:      navigator,
		Router:         router,
		Dashboard:      dashboard,
		ToolBrowser:    toolBrowser,
		BundleExplorer: bundleExplorer,
		InstallManager: installManager,
		ConfigView:     configView,
		HealthView:     healthView,
	}, nil
}

// createMessageRouter creates and configures the message router
func (f *DependencyFactory) createMessageRouter(healthView *views.HealthView) *MessageRouter {
	router := NewMessageRouter()
	
	// Register health view handler
	healthHandler := NewHealthViewMessageHandler(healthView)
	
	// Setup router with handler (using the existing setupMessageRouter pattern)
	router.Register(reflect.TypeOf(views.MemoryCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.DiskCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.InternetCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.BuildToolsCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.GitCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.PathCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.RustToolchainCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.GoToolchainCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.ToolUpdatesCheckCompleteMsg{}), healthHandler)
	router.Register(reflect.TypeOf(views.NextHealthCheckMsg{}), healthHandler)
	
	return router
}

