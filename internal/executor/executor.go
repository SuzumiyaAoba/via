package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/SuzumiyaAoba/entry/internal/history"
	"github.com/dop251/goja"
)

type CommandData struct {
	File string
	Dir  string
	Base string
	Name string
	Ext  string
}

type ExecutionOptions struct {
	Background bool
	Terminal   bool
}

type Executor struct {
	Out    io.Writer
	DryRun bool
}

func NewExecutor(out io.Writer, dryRun bool) *Executor {
	return &Executor{
		Out:    out,
		DryRun: dryRun,
	}
}

func (e *Executor) Execute(commandTmpl string, file string, opts ExecutionOptions) error {
	var cmdBuf bytes.Buffer
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

	if e.DryRun {
		bg := ""
		if opts.Background {
			bg = " (background)"
		}
		fmt.Fprintf(e.Out, "%s%s\n", cmdStr, bg)
		return nil
	}

	cmd := exec.Command("sh", "-c", cmdStr)
	
	if opts.Background {
		// Detach process
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		// TODO: Set SysProcAttr for full detachment if needed
		
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start background command: %w", err)
		}
		
		if err := cmd.Process.Release(); err != nil {
			return fmt.Errorf("failed to release process: %w", err)
		}
		return nil
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = e.Out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	// Record history
	if !e.DryRun {
		_ = history.AddEntry(file, "")
	}

	return nil
}

func (e *Executor) ExecuteCommand(command string, args []string) error {
	if e.DryRun {
		fmt.Fprintf(e.Out, "%s %s\n", command, strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = e.Out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func (e *Executor) OpenSystem(path string) error {
	var cmdName string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmdName = "open"
		args = []string{path}
	case "windows":
		cmdName = "cmd"
		args = []string{"/c", "start", "", path}
	default: // linux, freebsd, openbsd, netbsd
		cmdName = "xdg-open"
		args = []string{path}
	}

	if e.DryRun {
		fmt.Fprintf(e.Out, "%s %s\n", cmdName, strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command(cmdName, args...)
	cmd.Stdin = nil // Detach stdin for openers?
	cmd.Stdout = e.Out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open system default: %w", err)
	}

	return nil
}

// ExecuteScript executes a JavaScript snippet to determine the command or match status.
// It returns:
// - command (string): The command to execute (if any)
// - match (bool): Whether the rule matched
// - error: Any error during execution
func (e *Executor) ExecuteScript(script string, file string) (string, bool, error) {
	vm := goja.New()

	// Setup environment
	absFile, _ := filepath.Abs(file)
	dir := filepath.Dir(absFile)
	base := filepath.Base(absFile)
	ext := filepath.Ext(absFile)
	name := strings.TrimSuffix(base, ext)

	// Expose variables
	vm.Set("file", file)
	vm.Set("absFile", absFile)
	vm.Set("dir", dir)
	vm.Set("base", base)
	vm.Set("ext", ext)
	vm.Set("name", name)
	
	// Expose env vars
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}
	vm.Set("env", env)

	// Run script
	val, err := vm.RunString(script)
	if err != nil {
		return "", false, fmt.Errorf("script execution failed: %w", err)
	}

	// Handle return value
	if val == nil || val.Export() == nil {
		return "", false, nil
	}

	// Check type
	export := val.Export()
	switch v := export.(type) {
	case string:
		// Returned a command string -> Match and execute this command
		return v, true, nil
	case bool:
		// Returned boolean -> Match if true, use default command (which might be empty if not provided in rule)
		// But wait, if rule has script, it might NOT have command.
		// If it returns true, we expect the caller to handle "what command to run".
		// If the rule has both Script and Command, Script returning true means use Command.
		return "", v, nil
	default:
		// Treat other truthy values as true? No, be strict.
		return "", false, nil
	}
}
