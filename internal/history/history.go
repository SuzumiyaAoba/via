package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"` // The file or command executed
	RuleName  string    `json:"rule_name,omitempty"`
}

const MaxHistorySize = 100

var customHistoryPath string

func SetHistoryPath(path string) {
	customHistoryPath = path
}

// UserHomeDir is a variable to allow mocking in tests
var UserHomeDir = os.UserHomeDir

func GetHistoryPath() (string, error) {
	if customHistoryPath != "" {
		return customHistoryPath, nil
	}
	home, err := UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(home, ".config", "entry", "history.json"), nil
}

func LoadHistory() ([]HistoryEntry, error) {
	path, err := GetHistoryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []HistoryEntry{}, nil
		}
		return nil, err
	}

	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func AddEntry(command, ruleName string) error {
	entries, err := LoadHistory()
	if err != nil {
		// If error loading, start fresh
		entries = []HistoryEntry{}
	}

	newEntry := HistoryEntry{
		Timestamp: time.Now(),
		Command:   command,
		RuleName:  ruleName,
	}

	// Prepend new entry
	entries = append([]HistoryEntry{newEntry}, entries...)

	// Trim to max size
	if len(entries) > MaxHistorySize {
		entries = entries[:MaxHistorySize]
	}

	return saveHistory(entries)
}

func saveHistory(entries []HistoryEntry) error {
	path, err := GetHistoryPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func ClearHistory() error {
	_, err := GetHistoryPath()
	if err != nil {
		return err
	}
	
	// Just write empty array
	return saveHistory([]HistoryEntry{})
}
