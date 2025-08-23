package tui

import "time"

// UI Layout Constants
const (
	// Terminal dimensions
	DefaultWidth  = 80
	DefaultHeight = 24
	MinWidth      = 40
	MinHeight     = 10
	
	// View layout
	HeaderHeight      = 1
	FooterHeight      = 1
	MinViewportHeight = 5
	
	// Content limits
	MaxTaskHistory    = 100
	MaxErrorHistory   = 50
	MaxLogLines       = 1000
)

// Performance Constants
const (
	// Task execution
	DefaultMaxParallel    = 2
	MinParallel          = 1
	MaxParallel          = 16
	
	// Timing
	RefreshInterval      = 100 * time.Millisecond
	HealthCheckTimeout   = 30 * time.Second
	TaskTimeout          = 300 * time.Second
	
	// Cache
	ContentCacheTTL      = 1 * time.Minute
	StatusCacheTTL       = 5 * time.Minute
)

// UI Style Constants
const (
	// Colors (will be used with lipgloss)
	PrimaryColor   = "#7C3AED"  // Purple
	SecondaryColor = "#10B981"  // Green
	ErrorColor     = "#EF4444"  // Red
	WarningColor   = "#F59E0B"  // Yellow
	InfoColor      = "#3B82F6"  // Blue
	MutedColor     = "#6B7280"  // Gray
	
	// Progress indicators
	ProgressCompleteChar = "█"
	ProgressIncompleteChar = "░"
	ProgressWidth = 20
)