// Package data provides utilities for loading mock football match data.
package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	configDir = ".golazo"
)

// GetConfigDir returns the path to the golazo config directory.
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir)
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}

	return configPath, nil
}

// GetMockDataPath returns the path to the mock data file.
func GetMockDataPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "matches.json"), nil
}

// LiveUpdate represents a single live update string.
type LiveUpdate struct {
	MatchID int
	Update  string
	Time    time.Time
}

// SaveLiveUpdate appends a live update to the storage.
func SaveLiveUpdate(matchID int, update string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	updatesFile := filepath.Join(configDir, fmt.Sprintf("updates_%d.json", matchID))

	var updates []LiveUpdate
	if data, err := os.ReadFile(updatesFile); err == nil {
		json.Unmarshal(data, &updates)
	}

	updates = append(updates, LiveUpdate{
		MatchID: matchID,
		Update:  update,
		Time:    time.Now(),
	})

	data, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("marshal updates: %w", err)
	}

	return os.WriteFile(updatesFile, data, 0644)
}

// GetLiveUpdates retrieves live updates for a match.
func GetLiveUpdates(matchID int) ([]string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	updatesFile := filepath.Join(configDir, fmt.Sprintf("updates_%d.json", matchID))
	data, err := os.ReadFile(updatesFile)
	if err != nil {
		return []string{}, nil // Return empty if file doesn't exist
	}

	var updates []LiveUpdate
	if err := json.Unmarshal(data, &updates); err != nil {
		return nil, fmt.Errorf("unmarshal updates: %w", err)
	}

	result := make([]string, 0, len(updates))
	for _, update := range updates {
		result = append(result, update.Update)
	}

	return result, nil
}
