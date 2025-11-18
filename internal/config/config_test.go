package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	
	configContent := `
rules:
  - extensions: [txt]
    command: "echo text"
  - regex: ".*\\.log$"
    command: "echo log"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Load valid config",
			path:    configPath,
			wantErr: false,
		},
		{
			name:    "File not found",
			path:    filepath.Join(tmpDir, "nonexistent.yml"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			if !tt.wantErr {
				assert.Len(t, cfg.Rules, 2)

				// Verify regex rule
				foundRegex := false
				for _, rule := range cfg.Rules {
					if rule.Regex == ".*\\.log$" {
						foundRegex = true
						break
					}
				}
				assert.True(t, foundRegex, "LoadConfig() did not load regex rule")
			}
		})
	}
}
