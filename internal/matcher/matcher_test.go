package matcher

import (
	"testing"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	rules := []config.Rule{
		{
			Extensions: []string{"txt", "md"},
			Command:    "echo text",
		},
		{
			Regex:   ".*\\.log$",
			Command: "echo log",
		},
		{
			Extensions: []string{"go"},
			Command:    "echo go",
		},
	}

	tests := []struct {
		name     string
		filename string
		wantCmd  string
		wantErr  bool
	}{
		{
			name:     "Match extension txt",
			filename: "file.txt",
			wantCmd:  "echo text",
			wantErr:  false,
		},
		{
			name:     "Match extension md",
			filename: "README.md",
			wantCmd:  "echo text",
			wantErr:  false,
		},
		{
			name:     "Match regex log",
			filename: "app.log",
			wantCmd:  "echo log",
			wantErr:  false,
		},
		{
			name:     "Match extension go",
			filename: "main.go",
			wantCmd:  "echo go",
			wantErr:  false,
		},
		{
			name:     "No match",
			filename: "image.png",
			wantCmd:  "",
			wantErr:  false,
		},
		{
			name:     "Case insensitive extension",
			filename: "FILE.TXT",
			wantCmd:  "echo text",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(rules, tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantCmd != "" {
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantCmd, got.Command)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}
