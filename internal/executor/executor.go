package executor

import (
	"bytes"
	"fmt"
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

func Execute(commandTmpl string, file string, dryRun bool) error {
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
		File: file, // Keep original input as File? Or use absolute? Let's keep original for now, or maybe absolute is better.
		// Actually, for consistency, File should probably be what the user passed, but for safety absolute might be better.
		// Let's use absolute path for File as well to be safe.
		// Wait, if user passes relative path, they might expect relative.
		// Let's stick to what we had: File is input.
		// But for Dir/Base etc we need absolute or at least cleaned path.
		// Let's use absolute for derived values.
		Dir:  dir,
		Base: base,
		Name: name,
		Ext:  ext,
	}
	// Override File with absolute path? Or keep as is?
	// Previous implementation used `file` arg directly.
	// Let's keep `File` as the input argument for backward compatibility (if any),
	// but maybe we should expose `Abs` too?
	// For now, let's just populate the new fields.
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
