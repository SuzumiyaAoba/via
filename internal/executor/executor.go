package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type CommandData struct {
	File string
	Dir  string
	Base string
	Name string
	Ext  string
}

var cmdBuf bytes.Buffer

func Execute(out io.Writer, commandTmpl string, file string, dryRun bool) error {
	cmdBuf.Reset()
	tmpl, err := template.New("command").Parse(commandTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse command template: %w", err)
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	dir := filepath.Dir(absFile)
	base := filepath.Base(absFile)
	ext := filepath.Ext(absFile)
	name := strings.TrimSuffix(base, ext)

	data := CommandData{
		File: file,
		Dir:  dir,
		Base: base,
		Name: name,
		Ext:  ext,
	}

	if err := tmpl.Execute(&cmdBuf, data); err != nil {
		return fmt.Errorf("failed to execute command template: %w", err)
	}

	cmdStr := cmdBuf.String()

	if dryRun {
		fmt.Fprintln(out, cmdStr)
		return nil
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func ExecuteCommand(out io.Writer, command string, args []string, dryRun bool) error {
	if dryRun {
		fmt.Fprintf(out, "%s %s\n", command, strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
