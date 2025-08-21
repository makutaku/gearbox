package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// loadConfig loads the configuration from a JSON file
func loadConfig(path string) (Config, error) {
	var config Config

	// Check if the config file exists and use it
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			file, err := os.Open(path)
			if err != nil {
				return config, fmt.Errorf("failed to open config file %s: %w", path, err)
			}
			defer file.Close()

			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&config); err != nil {
				return config, fmt.Errorf("failed to decode config from %s: %w", path, err)
			}
			return config, nil
		} else {
			// If a specific path was provided but doesn't exist, return error
			return config, fmt.Errorf("config file not found: %s", path)
		}
	}

	// Look for default config files in order of preference
	possiblePaths := []string{
		"config/tools.json",
		"../config/tools.json",
		"../../config/tools.json",
		"../../../config/tools.json",
		filepath.Join(os.Getenv("HOME"), ".gearbox", "tools.json"),
		"/etc/gearbox/tools.json",
	}

	for _, configPath := range possiblePaths {
		if _, err := os.Stat(configPath); err == nil {
			file, err := os.Open(configPath)
			if err != nil {
				continue
			}
			defer file.Close()

			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&config); err != nil {
				continue
			}
			return config, nil
		}
	}

	return config, fmt.Errorf("no valid configuration file found in any of the expected locations")
}

// loadConfigStreaming loads configuration from byte data
func loadConfigStreaming(data []byte) (Config, error) {
	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to decode config from data: %w", err)
	}

	return config, nil
}