package tui

import (
	"reflect"
	
	tea "github.com/charmbracelet/bubbletea"
	"gearbox/cmd/gearbox/tui/views"
)

// MessageHandler defines the interface for components that can handle messages
type MessageHandler interface {
	HandleMessage(msg tea.Msg) tea.Cmd
}

// MessageRouter routes messages to appropriate handlers
type MessageRouter struct {
	handlers map[reflect.Type][]MessageHandler
}

// NewMessageRouter creates a new message router
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[reflect.Type][]MessageHandler),
	}
}

// Register registers a handler for a specific message type
func (r *MessageRouter) Register(msgType reflect.Type, handler MessageHandler) {
	r.handlers[msgType] = append(r.handlers[msgType], handler)
}

// Route routes a message to all registered handlers and returns the first non-nil command
func (r *MessageRouter) Route(msg tea.Msg) tea.Cmd {
	msgType := reflect.TypeOf(msg)
	if handlers, exists := r.handlers[msgType]; exists {
		for _, handler := range handlers {
			if cmd := handler.HandleMessage(msg); cmd != nil {
				return cmd
			}
		}
	}
	return nil
}

// HealthViewMessageHandler wraps the health view to implement MessageHandler
type HealthViewMessageHandler struct {
	view *views.HealthView
}

// NewHealthViewMessageHandler creates a new health view message handler
func NewHealthViewMessageHandler(view *views.HealthView) *HealthViewMessageHandler {
	return &HealthViewMessageHandler{view: view}
}

// HandleMessage handles health check messages
func (h *HealthViewMessageHandler) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case views.MemoryCheckCompleteMsg,
		 views.DiskCheckCompleteMsg,
		 views.InternetCheckCompleteMsg,
		 views.BuildToolsCheckCompleteMsg,
		 views.GitCheckCompleteMsg,
		 views.PathCheckCompleteMsg,
		 views.RustToolchainCheckCompleteMsg,
		 views.GoToolchainCheckCompleteMsg,
		 views.ToolUpdatesCheckCompleteMsg,
		 views.NextHealthCheckMsg:
		return h.view.Update(msg)
	}
	return nil
}

// SetupMessageRouter configures the message router with all handlers
func (m *Model) setupMessageRouter() *MessageRouter {
	router := NewMessageRouter()
	
	// Register health view handler for all health check messages
	// Cast interface back to concrete type for handler
	if healthAdapter, ok := m.healthView.(*HealthAdapter); ok {
		healthHandler := NewHealthViewMessageHandler(healthAdapter.health)
		
		// Register all health check message types
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
	}
	
	return router
}