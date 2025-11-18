package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type CommandData struct {
	File string
}

func Execute(commandTmpl string, file string, dryRun bool) error {
	tmpl, err := template.New("command").Parse(commandTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse command template: %w", err)
	}

	var cmdBuf bytes.Buffer
	data := CommandData{File: file}
	if err := tmpl.Execute(&cmdBuf, data); err != nil {
		return fmt.Errorf("failed to execute command template: %w", err)
	}

	cmdStr := cmdBuf.String()

	if dryRun {
		fmt.Println(cmdStr)
		return nil
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
