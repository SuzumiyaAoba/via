package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	// Since Execute runs a command, we can test it by running a simple echo command
	// and checking if it runs without error. Capturing stdout is harder without refactoring
	// Execute to take an io.Writer, but for now we just check for errors.

	tests := []struct {
		name        string
		commandTmpl string
		file        string
		wantErr     bool
	}{
		{
			name:        "Simple echo",
			commandTmpl: "echo {{.File}}",
			file:        "test.txt",
			wantErr:     false,
		},
		{
			name:        "Invalid template",
			commandTmpl: "echo {{.File",
			file:        "test.txt",
			wantErr:     true,
		},
		{
			name:        "Command failure",
			commandTmpl: "false",
			file:        "test.txt",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Execute(tt.commandTmpl, tt.file)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
