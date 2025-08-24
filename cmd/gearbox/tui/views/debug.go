//go:build debug
// +build debug

package views

import (
	"fmt"
	"os"
	"path/filepath"
)

// debugLog writes debug messages to secure user directory only in debug builds
func debugLog(format string, args ...interface{}) {
	// Get user cache directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // Fail silently in debug mode
	}
	
	// Create secure debug directory in user's cache
	debugDir := filepath.Join(homeDir, ".cache", "gearbox")
	if err := os.MkdirAll(debugDir, 0700); err != nil {
		return // Fail silently
	}
	
	// Open debug file with restrictive permissions (user-only read/write)
	debugPath := filepath.Join(debugDir, "tui-debug.log")
	if debugFile, err := os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); err == nil {
		fmt.Fprintf(debugFile, format+"
", args...)
		debugFile.Close()
	}
}

// debugEnabled returns true in debug builds
const debugEnabled = true