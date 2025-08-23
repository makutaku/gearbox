//go:build debug
// +build debug

package views

import (
	"fmt"
	"os"
)

// debugLog writes debug messages to file only in debug builds
func debugLog(format string, args ...interface{}) {
	if debugFile, err := os.OpenFile("/tmp/gearbox-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(debugFile, format+"\n", args...)
		debugFile.Close()
	}
}

// debugEnabled returns true in debug builds
const debugEnabled = true