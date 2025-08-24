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
	// Core services (now using interfaces)
	Orchestrator OrchestratorService
	Manifest     ManifestService
	TaskManager  TaskService
	
	// Handlers (now using interfaces)
	Navigator NavigationService
	Router    MessageRoutingService
	
	// Lifecycle management
	Lifecycle *ViewLifecycleManager
	
	// Views (now using interfaces)
	Dashboard      DashboardService
	ToolBrowser    ToolBrowserService
	BundleExplorer BundleExplorerService
	InstallManager InstallManagerService
	ConfigView     ConfigService
	HealthView     HealthService
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
	// Create core services (concrete types)
	orch, err := orchestrator.NewOrchestratorBuilder(f.orchestratorOpts).Build()
	if err != nil {
		return nil, err
	}
	
	manifestMgr := manifest.NewManager()
	taskManager := tasks.NewTaskManager(orch, DefaultMaxParallel)
	
	// Create handlers (concrete types)
	navigationHandler := NewNavigationHandler()
	
	// Create views (concrete types)
	dashboard := views.NewDashboard()
	toolBrowser := views.NewToolBrowserNew()
	bundleExplorer := views.NewBundleExplorerNew()
	installManager := views.NewInstallManagerNew()
	configView := views.NewConfigView()
	healthView := views.NewHealthView()
	
	// Setup task provider for install manager
	taskProvider := NewTaskManagerProvider(taskManager)
	installManager.SetTaskProvider(taskProvider)
	
	// Create message router (concrete type)
	messageRouter := f.createMessageRouter(healthView)
	
	// Create lifecycle manager and register view services
	lifecycle := NewViewLifecycleManager()
	dashboardService := NewDashboardAdapter(dashboard)
	toolBrowserService := NewToolBrowserAdapter(toolBrowser)
	bundleExplorerService := NewBundleExplorerAdapter(bundleExplorer)
	installManagerService := NewInstallManagerAdapter(installManager)
	configService := NewConfigAdapter(configView)
	healthService := NewHealthAdapter(healthView)
	
	// Register views with lifecycle manager
	lifecycle.RegisterView(ViewDashboard, dashboardService)
	lifecycle.RegisterView(ViewToolBrowser, toolBrowserService)
	lifecycle.RegisterView(ViewBundleExplorer, bundleExplorerService)
	lifecycle.RegisterView(ViewMonitor, installManagerService)
	lifecycle.RegisterView(ViewConfig, configService)
	lifecycle.RegisterView(ViewHealth, healthService)
	
	// Wrap concrete types in adapters to implement interfaces
	return &Dependencies{
		Orchestrator:   NewOrchestratorAdapter(orch),
		Manifest:       NewManifestAdapter(manifestMgr),
		TaskManager:    NewTaskAdapter(taskManager),
		Navigator:      NewNavigationAdapter(navigationHandler),
		Router:         NewMessageRoutingAdapter(messageRouter),
		Lifecycle:      lifecycle,
		Dashboard:      dashboardService,
		ToolBrowser:    toolBrowserService,
		BundleExplorer: bundleExplorerService,
		InstallManager: installManagerService,
		ConfigView:     configService,
		HealthView:     healthService,
	}, nil
}

// createMessageRouter creates and configures the message router
func (f *DependencyFactory) createMessageRouter(healthView *views.HealthView) *MessageRouter {
	router := NewMessageRouter()
	
	// Register health view handler with the concrete view
	// We keep using the concrete view here since the handler needs the specific implementation
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

