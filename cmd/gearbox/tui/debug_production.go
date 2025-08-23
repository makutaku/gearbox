//go:build !debug
// +build !debug

package tui

// debugLog is a no-op in production builds
func debugLog(format string, args ...interface{}) {
	// No-op in production builds
}

// debugEnabled returns false in production builds
const debugEnabled = false