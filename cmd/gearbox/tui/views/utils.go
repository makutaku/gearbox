package views

// Common UI measurements and utilities for consistent layout

const (
	// Standard component heights
	TitleHeight      = 2  // Title with margin
	HelpBarHeight    = 2  // Help bar with spacing
	StatusBarHeight  = 2  // Status bar with spacing
	SearchBarHeight  = 3  // Search bar with border and spacing
	CategoryBarHeight = 2 // Category selector with spacing
	
	// Box component overhead
	BoxBorderHeight  = 2  // Top and bottom borders
	
	// Minimum heights
	MinContentHeight = 5  // Minimum height for content areas
)

// CalculateContentHeight returns the available height for content
// after accounting for standard UI components
func CalculateContentHeight(totalHeight int, components ...int) int {
	used := 0
	for _, h := range components {
		used += h
	}
	
	available := totalHeight - used
	if available < MinContentHeight {
		return MinContentHeight
	}
	return available
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}