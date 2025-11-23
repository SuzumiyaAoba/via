package cli

import (
	"fmt"

	"github.com/SuzumiyaAoba/entry/internal/executor"
	"github.com/SuzumiyaAoba/entry/internal/history"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

func init() {
	historyCmd.AddCommand(historyClearCmd)
}

var historyCmd = &cobra.Command{
	Use:   ":history",
	Short: "View and execute command history",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistory(cmd)
	},
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear command history",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := history.ClearHistory(); err != nil {
			return fmt.Errorf("failed to clear history: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "History cleared")
		return nil
	},
}

func runHistory(cmd *cobra.Command) error {
	entries, err := history.LoadHistory()
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No history available")
		return nil
	}

	// Build options for selection
	type historyOption struct {
		Label string
		Entry history.HistoryEntry
	}

	var options []huh.Option[historyOption]
	for _, entry := range entries {
		label := fmt.Sprintf("%s  %s (%s)", 
			entry.Timestamp.Format("2006-01-02 15:04:05"), 
			entry.Command, 
			entry.RuleName)
		options = append(options, huh.NewOption(label, historyOption{Label: label, Entry: entry}))
	}

	var selected historyOption
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[historyOption]().
				Title("Select a command to re-run").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	// Re-run the command
	// We need to construct a new root command execution or just call the handler
	// Calling the handler is tricky because we need config and executor.
	// Simplest way is to print the command and let user run it, or use executor to run it as if it was passed.
	// But wait, the history stores the "file" or "command" passed to `et`.
	// So we should basically re-invoke the logic for that file.
	
	fmt.Fprintf(cmd.OutOrStdout(), "Re-running: %s\n", selected.Entry.Command)
	
	// We can't easily re-invoke the whole CLI flow from here without circular deps or refactoring.
	// But we can exec a new process, or just tell the user.
	// Better: return the command to the caller? No, RunE returns error.
	
	// Let's try to execute it using the current process's logic if possible.
	// We are in `cli` package. We can call `handleFileExecution` or `handleCommandExecution`.
	// We need `cfg` and `exec`.
	
	// Load config again
	// Note: This duplicates logic from root.go, but it's acceptable for now.
	// Ideally we'd refactor `runRoot` to be more reusable.
	
	// For now, let's just print it and maybe execute it if it's a simple file.
	// Actually, `et` is about opening files. So `selected.Entry.Command` is likely a filename.
	
	// Let's try to execute it.
	// We need to get the config and executor.
	// Since we are inside `cli`, we can access `cfgFile` var but we need to load it.
	
	// This part is a bit hacky, but let's do it.
	// We will just print it for now as "Re-running" and then exit? 
	// No, the user expects it to run.
	
	// Let's spawn a new `et` process? That's safe.
	// Or just use `executor` to run the command if we knew what rule it matched.
	// But we only stored the RuleName for display.
	
	// Let's just output the command to stdout? No.
	
	// Let's try to re-run the root command logic.
	// We can't call `runRoot` easily.
	
	// Let's use `os/exec` to call `et` again with the argument.
	// This is the most robust way to ensure all logic (rules, profiles etc) is applied.
	
	exec := executor.NewExecutor(cmd.OutOrStdout(), false)
	return exec.ExecuteCommand("et", []string{selected.Entry.Command})
}
