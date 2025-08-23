package sudo

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ValidatePassword tests if the provided password is valid for sudo
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test sudo access with the password
	cmd := exec.CommandContext(ctx, "sudo", "-S", "-p", "", "true")
	cmd.Stdin = strings.NewReader(password + "\n")
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if it's an authentication error
		if strings.Contains(stderr.String(), "Sorry, try again") ||
		   strings.Contains(stderr.String(), "incorrect password") ||
		   strings.Contains(stderr.String(), "authentication failure") {
			return fmt.Errorf("incorrect password")
		}
		return fmt.Errorf("sudo validation failed: %v", err)
	}

	return nil
}

// IsPasswordRequired checks if sudo password is required
func IsPasswordRequired() (bool, error) {
	// Test if we can run sudo without a password (passwordless sudo)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", "-n", "true")
	err := cmd.Run()
	
	// If sudo -n succeeds, no password is required
	if err == nil {
		return false, nil
	}

	// Check if it's specifically asking for a password
	if exitError, ok := err.(*exec.ExitError); ok {
		// Exit code 1 typically means password required
		if exitError.ExitCode() == 1 {
			return true, nil
		}
	}

	// Other error - might be sudo not available
	return false, fmt.Errorf("unable to test sudo access: %v", err)
}

// ExecuteWithPassword executes a command with sudo using the provided password
func ExecuteWithPassword(password string, command string, args ...string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Create the sudo command
	sudoArgs := append([]string{"-S", "-p", "", command}, args...)
	cmd := exec.Command("sudo", sudoArgs...)
	cmd.Stdin = strings.NewReader(password + "\n")

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Include stderr in error message for debugging
		if stderr.Len() > 0 {
			return fmt.Errorf("command failed: %v (stderr: %s)", err, stderr.String())
		}
		return fmt.Errorf("command failed: %v", err)
	}

	return nil
}

// SecureClear overwrites sensitive data in memory
func SecureClear(data []byte) {
	for i := range data {
		data[i] = 0
	}
}