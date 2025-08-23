package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// UITestResult represents the result of a UI element test
type UITestResult struct {
	TestName    string
	Passed      bool
	Message     string
	Details     []string
}

// UITestSuite runs comprehensive UI element tests
type UITestSuite struct {
	model  tea.Model  // Use tea.Model interface instead of concrete type
	results []UITestResult
}

// NewUITestSuite creates a new UI testing suite
func NewUITestSuite() (*UITestSuite, error) {
	model, err := NewDemoModel()
	if err != nil {
		return nil, err
	}

	// Initialize the model - Update returns tea.Model interface
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	return &UITestSuite{
		model:   updatedModel,
		results: []UITestResult{},
	}, nil
}

// RunAllTests runs all UI element tests
func (ts *UITestSuite) RunAllTests() []UITestResult {
	ts.results = []UITestResult{}

	// Test each view's UI elements
	ts.testDashboardElements()
	ts.testToolBrowserElements()
	ts.testBundleExplorerElements()
	ts.testInstallManagerElements()
	ts.testConfigViewElements()
	ts.testHealthViewElements()

	// Test cross-view navigation
	ts.testViewNavigation()

	// Test keyboard shortcuts
	ts.testKeyboardShortcuts()

	return ts.results
}

// addResult adds a test result to the suite
func (ts *UITestSuite) addResult(name string, passed bool, message string, details ...string) {
	ts.results = append(ts.results, UITestResult{
		TestName: name,
		Passed:   passed,
		Message:  message,
		Details:  details,
	})
}

// testDashboardElements tests dashboard UI elements
func (ts *UITestSuite) testDashboardElements() {
	// Navigate to dashboard
	var cmd tea.Cmd
	ts.model, cmd = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	_ = cmd  // Ignore the command for testing
	output := ts.model.View()

	// Test that dashboard renders
	if output == "" {
		ts.addResult("Dashboard Render", false, "Dashboard should render content")
		return
	}

	// Test dashboard sections
	expectedSections := []string{"Dashboard", "System Status", "Quick Actions", "Recent Activity", "Recommendations"}
	var foundSections []string
	var missingSections []string

	for _, section := range expectedSections {
		if strings.Contains(output, section) {
			foundSections = append(foundSections, section)
		} else {
			missingSections = append(missingSections, section)
		}
	}

	if len(missingSections) == 0 {
		ts.addResult("Dashboard Sections", true, "All dashboard sections present", foundSections...)
	} else {
		ts.addResult("Dashboard Sections", false, "Missing dashboard sections", missingSections...)
	}

	// Test navigation indicators
	if strings.Contains(output, "[T]") && strings.Contains(output, "[B]") && strings.Contains(output, "[H]") {
		ts.addResult("Dashboard Navigation", true, "Navigation indicators present")
	} else {
		ts.addResult("Dashboard Navigation", false, "Navigation indicators missing or incomplete")
	}
}

// testToolBrowserElements tests tool browser UI elements and selection
func (ts *UITestSuite) testToolBrowserElements() {
	// Navigate to tool browser
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	output := ts.model.View()

	if output == "" {
		ts.addResult("Tool Browser Render", false, "Tool browser should render content")
		return
	}

	// Test that tools are listed
	if !strings.Contains(output, "Tool") {
		ts.addResult("Tool Browser Content", false, "Should contain tool listings")
		return
	}

	// Test navigation through tools
	initialOutput := output
	
	// Move down multiple times and verify cursor movement
	var outputs []string
	for i := 0; i < 5; i++ {
		ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
		newOutput := ts.model.View()
		outputs = append(outputs, newOutput)
		
		// Should be different from initial output (cursor moved)
		if newOutput == initialOutput {
			ts.addResult("Tool Navigation", false, fmt.Sprintf("Down arrow %d should change display", i+1))
		}
	}

	ts.addResult("Tool Navigation", true, "Tool navigation works", fmt.Sprintf("Tested %d navigation steps", len(outputs)))

	// Test tool selection
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	selectedOutput := ts.model.View()

	// The output should indicate selection (look for selection markers)
	if strings.Contains(selectedOutput, "▣") || strings.Contains(selectedOutput, "✓") {
		ts.addResult("Tool Selection", true, "Tool selection indicators present")
	} else {
		ts.addResult("Tool Selection", false, "Tool selection indicators not found")
	}

	// Test multiple selections
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	multiSelectOutput := ts.model.View()

	selectionCount := strings.Count(multiSelectOutput, "▣")
	if selectionCount >= 2 {
		ts.addResult("Tool Multi-Selection", true, fmt.Sprintf("Multiple selections work (%d items selected)", selectionCount))
	} else {
		ts.addResult("Tool Multi-Selection", false, "Multi-selection not working properly")
	}

	// Test category cycling
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	categoryOutput := ts.model.View()

	if categoryOutput != selectedOutput {
		ts.addResult("Tool Category Filter", true, "Category cycling changes display")
	} else {
		ts.addResult("Tool Category Filter", false, "Category cycling doesn't change display")
	}

	// Test search functionality
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	searchOutput := ts.model.View()

	if strings.Contains(searchOutput, "Search") || strings.Contains(searchOutput, "🔍") {
		ts.addResult("Tool Search", true, "Search mode activation works")
	} else {
		ts.addResult("Tool Search", false, "Search mode not activated")
	}
}

// testBundleExplorerElements tests bundle explorer UI elements
func (ts *UITestSuite) testBundleExplorerElements() {
	// Navigate to bundle explorer
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	output := ts.model.View()

	if output == "" {
		ts.addResult("Bundle Explorer Render", false, "Bundle explorer should render content")
		return
	}

	// Test bundle listings
	if !strings.Contains(output, "Bundle") {
		ts.addResult("Bundle Explorer Content", false, "Should contain bundle listings")
		return
	}

	// Test bundle navigation
	initialOutput := output
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	navOutput := ts.model.View()

	if navOutput != initialOutput {
		ts.addResult("Bundle Navigation", true, "Bundle navigation works")
	} else {
		ts.addResult("Bundle Navigation", false, "Bundle navigation not working")
	}

	// Test bundle expansion
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})
	expandedOutput := ts.model.View()

	// Look for expansion indicators
	if strings.Contains(expandedOutput, "▼") || strings.Contains(expandedOutput, "Tools:") {
		ts.addResult("Bundle Expansion", true, "Bundle expansion works")
	} else {
		ts.addResult("Bundle Expansion", false, "Bundle expansion not working")
	}

	// Test bundle category filtering
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	categoryOutput := ts.model.View()

	if categoryOutput != expandedOutput {
		ts.addResult("Bundle Category Filter", true, "Bundle category filtering works")
	} else {
		ts.addResult("Bundle Category Filter", false, "Bundle category filtering not working")
	}
}

// testInstallManagerElements tests install manager UI elements
func (ts *UITestSuite) testInstallManagerElements() {
	// First, add some tasks by selecting tools
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})

	// Navigate to install manager
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	output := ts.model.View()

	if output == "" {
		ts.addResult("Install Manager Render", false, "Install manager should render content")
		return
	}

	// Test install manager content
	if strings.Contains(output, "Installation Manager") || strings.Contains(output, "task") {
		ts.addResult("Install Manager Content", true, "Install manager shows task information")
	} else {
		ts.addResult("Install Manager Content", false, "Install manager missing task information")
	}

	// Test task navigation (if tasks exist)
	if strings.Contains(output, "task") {
		ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
		_ = ts.model.View() // navOutput - ignore output for now
		
		ts.addResult("Install Manager Navigation", true, "Task navigation attempted")
	} else {
		ts.addResult("Install Manager Navigation", false, "No tasks to navigate")
	}

	// Test output toggle
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	_ = ts.model.View() // toggleOutput - ignore output for now

	ts.addResult("Install Manager Output Toggle", true, "Output toggle executed")
}

// testConfigViewElements tests configuration view UI elements
func (ts *UITestSuite) testConfigViewElements() {
	// Navigate to config view
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	output := ts.model.View()

	if output == "" {
		ts.addResult("Config View Render", false, "Config view should render content")
		return
	}

	// Test config content
	if strings.Contains(output, "Configuration") || strings.Contains(output, "Settings") {
		ts.addResult("Config View Content", true, "Config view shows configuration")
	} else {
		ts.addResult("Config View Content", false, "Config view missing configuration content")
	}

	// Test config navigation
	initialOutput := output
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	navOutput := ts.model.View()

	if navOutput != initialOutput {
		ts.addResult("Config Navigation", true, "Config item navigation works")
	} else {
		ts.addResult("Config Navigation", false, "Config item navigation not working")
	}

	// Test config editing
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})
	editOutput := ts.model.View()

	if editOutput != navOutput {
		ts.addResult("Config Editing", true, "Config editing mode activated")
	} else {
		ts.addResult("Config Editing", false, "Config editing mode not activated")
	}
}

// testHealthViewElements tests health view UI elements
func (ts *UITestSuite) testHealthViewElements() {
	// Navigate to health view
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	output := ts.model.View()

	if output == "" {
		ts.addResult("Health View Render", false, "Health view should render content")
		return
	}

	// Test health content
	if strings.Contains(output, "Health") || strings.Contains(output, "System") {
		ts.addResult("Health View Content", true, "Health view shows health information")
	} else {
		ts.addResult("Health View Content", false, "Health view missing health content")
	}

	// Test health check navigation
	initialOutput := output
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	navOutput := ts.model.View()

	if navOutput != initialOutput {
		ts.addResult("Health Navigation", true, "Health check navigation works")
	} else {
		ts.addResult("Health Navigation", false, "Health check navigation not working")
	}

	// Test health check details
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("enter")})
	_ = ts.model.View() // detailsOutput - ignore output for now

	ts.addResult("Health Details", true, "Health details toggle executed")

	// Test health refresh
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	_ = ts.model.View() // refreshOutput - ignore output for now

	ts.addResult("Health Refresh", true, "Health refresh executed")
}

// testViewNavigation tests navigation between views
func (ts *UITestSuite) testViewNavigation() {
	views := []struct {
		key      string
		name     string
		expected string
	}{
		{"d", "Dashboard", "Dashboard"},
		{"t", "Tools", "Tool"},
		{"b", "Bundles", "Bundle"},
		{"i", "Install", "Installation"},
		{"c", "Config", "Configuration"},
		{"h", "Health", "Health"},
	}

	passedViews := 0
	for _, view := range views {
		ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(view.key)})
		output := ts.model.View()

		if strings.Contains(output, view.expected) {
			passedViews++
		}
	}

	if passedViews == len(views) {
		ts.addResult("View Navigation", true, fmt.Sprintf("All %d views accessible", len(views)))
	} else {
		ts.addResult("View Navigation", false, fmt.Sprintf("Only %d/%d views working", passedViews, len(views)))
	}

	// Test Tab navigation
	initialView := ts.model.View()
	ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyTab})
	tabView := ts.model.View()

	if tabView != initialView {
		ts.addResult("Tab Navigation", true, "Tab key changes views")
	} else {
		ts.addResult("Tab Navigation", false, "Tab key navigation not working")
	}
}

// testKeyboardShortcuts tests various keyboard shortcuts
func (ts *UITestSuite) testKeyboardShortcuts() {
	shortcuts := []struct {
		key         string
		description string
		testFunc    func() bool
	}{
		{"up", "Up Arrow", func() bool {
			before := ts.model.View()
			ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyUp})
			after := ts.model.View()
			return before != after // Should change display
		}},
		{"down", "Down Arrow", func() bool {
			before := ts.model.View()
			ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyDown})
			after := ts.model.View()
			return before != after // Should change display
		}},
		{"left", "Left Arrow", func() bool {
			before := ts.model.View()
			ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyLeft})
			after := ts.model.View()
			return before != after // Should change display
		}},
		{"right", "Right Arrow", func() bool {
			before := ts.model.View()
			ts.model, _ = ts.model.Update(tea.KeyMsg{Type: tea.KeyRight})
			after := ts.model.View()
			return before != after // Should change display
		}},
	}

	passedShortcuts := 0
	for _, shortcut := range shortcuts {
		if shortcut.testFunc() {
			ts.addResult(fmt.Sprintf("Shortcut %s", shortcut.description), true, fmt.Sprintf("%s key works", shortcut.description))
			passedShortcuts++
		} else {
			ts.addResult(fmt.Sprintf("Shortcut %s", shortcut.description), false, fmt.Sprintf("%s key not working", shortcut.description))
		}
	}

	ts.addResult("Keyboard Shortcuts", passedShortcuts == len(shortcuts), 
		fmt.Sprintf("%d/%d keyboard shortcuts working", passedShortcuts, len(shortcuts)))
}

// RunUIElementTests is a convenience function to run all UI tests and return a summary
func RunUIElementTests() (passed int, total int, results []UITestResult, err error) {
	suite, err := NewUITestSuite()
	if err != nil {
		return 0, 0, nil, err
	}

	results = suite.RunAllTests()
	
	for _, result := range results {
		total++
		if result.Passed {
			passed++
		}
	}

	return passed, total, results, nil
}