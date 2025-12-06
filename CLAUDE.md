# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Via (`vv`) is a CLI file association tool written in Go that executes commands based on file extensions, regex patterns, MIME types, URL schemes, or JavaScript scripting. It provides intelligent file handling with interactive selection, dry-run mode, TUI dashboard, remote sync, command history, and detailed matching explanations.

## Build & Development Commands

### Building
```bash
# Build the binary (outputs to ./vv)
task build
# or
go build -o vv ./cmd/vv

# Run directly without building
go run ./cmd/vv <args>
```

### Testing
```bash
# Run all tests
task test
# or
go test ./...

# Run tests for a specific package
go test ./internal/matcher
go test ./internal/config
go test ./internal/executor
go test ./internal/cli

# Run specific test by name
go test ./internal/matcher -run TestMatch

# Run tests with coverage
task coverage

# Generate HTML coverage report
task coverage-html
```

### Cleanup & Release
```bash
# Clean build artifacts
task clean

# Create snapshot release (requires goreleaser)
task release
```

## Architecture

### Command Flow

The application follows a layered architecture with clear separation of concerns:

1. **CLI Layer** (`internal/cli/`): Cobra-based command handling with manual flag parsing to support pass-through arguments
2. **Configuration Layer** (`internal/config/`): YAML-based config loading from `~/.config/via/config.yml` with profile support
3. **Matching Layer** (`internal/matcher/`): Rule matching against file extensions, regex, MIME types, URL schemes, OS, and JavaScript scripts
4. **Execution Layer** (`internal/executor/`): Command execution with template support, JavaScript scripting, and background/terminal options
5. **Sync Layer** (`internal/sync/`): GitHub Gist integration for remote config synchronization
6. **History Layer** (`internal/history/`): JSON-based command execution history tracking
7. **TUI Layer** (`internal/tui/`): Bubbletea-based terminal dashboard for rule management
8. **Logger Layer** (`internal/logger/`): Structured logging with file rotation

### Execution Modes

The root command in `internal/cli/root.go` dispatches to different execution modes:

- **Normal execution**: Single file argument → rule matching → execute matched rule's command
- **Interactive mode** (`--select`, `-s`): Shows interactive selector (using Charm Huh) with all matching rules + system default
- **Explain mode** (`--explain`): Displays detailed matching information with styled output (using Charm Lipgloss)
- **Command execution**: Multiple arguments or no file match → execute as shell command (with alias support)
- **Config subcommands**: `:config` commands for configuration management (add, remove, edit, move, alias, sync, import/export)
- **Dashboard mode** (`:dashboard`): Full-featured TUI for managing rules, viewing history, and sync status
- **History mode** (`:history`): View and re-run previous commands
- **Match mode** (`:match`): Test which rules would match a given file

### Rule Matching System

Rules in `internal/matcher/matcher.go` are matched in order with multiple criteria:

-   **OS filtering**: Rules can target specific operating systems (checked first)
-   **URL scheme matching**: Matches URL schemes (http, https, ftp, etc.) - short-circuits if specified
-   **Extension matching**: Case-insensitive file extension completion (works with both files and URLs)
-   **Regex matching**: Pattern matching against full filename
-   **MIME type matching**: Content-based matching for files (not URLs) using mimetype detection
-   **JavaScript script matching**: Execute JavaScript code that returns boolean (match) or string (custom command)
-   **Fallthrough support**: Rules with `fallthrough: true` allow multiple rules to execute sequentially

The matcher provides two functions:
-   `Match()`: Returns matched rules until first non-fallthrough rule (used for execution)
-   `MatchAll()`: Returns all matching rules regardless of fallthrough (used for interactive mode)

**Important matching behavior**: Multiple criteria within a rule are evaluated with OR logic - if ANY criterion matches (extension, regex, mime, script), the rule matches.

### Command Template System

Commands in `internal/executor/executor.go` support Go template syntax with these fields:
-   `{{.File}}`: Original file argument
-   `{{.Dir}}`: Directory containing the file
-   `{{.Base}}`: Base filename with extension
-   `{{.Name}}`: Filename without extension
-   `{{.Ext}}`: File extension (with dot)

Example: `vim {{.File}}` or `open "{{.Dir}}/{{.Name}}.pdf"`

### Configuration Structure

Config file (`config.yml`) structure:
```yaml
version: "1"
default_command: "vim {{.File}}"  # Or use "default" as shorter alias
aliases:
  v: "vim"
sync:
  gist_id: "abc123..."  # GitHub Gist ID for remote sync
  token: ""  # Optional: GitHub token (usually passed via env var)
rules:
  - name: "Open PDFs"
    extensions: ["pdf"]
    command: "open {{.File}}"
    background: true
    os: ["darwin"]
  - name: "Markdown Files"
    extensions: ["md", "txt"]
    regex: ".*\\.md$"
    mime: "text/.*"
    scheme: "https"
    terminal: true
    fallthrough: true
    command: "cat {{.File}}"
  - name: "JavaScript Scripted Rule"
    script: "file.endsWith('.log') && file.includes('error')"
    command: "tail -f {{.File}}"
```

**Profiles**: Profiles are stored in `~/.config/via/profiles/<profile-name>.yml` and can be selected with `--profile` flag or `VIA_PROFILE` env var.

### CLI Command Organization

The CLI layer in `internal/cli/` is organized into focused handler files:

-   `root.go`: Main command dispatcher with manual flag parsing to support pass-through arguments
-   `execute.go`: File and command execution dispatch logic (handles file execution, default commands, and alias expansion)
-   `interactive.go`: Interactive selection mode with Charm Huh forms
-   `explain.go`: Detailed rule matching visualization with Lipgloss styling
-   `cmd_config.go`: Main config subcommand with nested commands
-   `cmd_alias.go`: Alias management (add, list, remove)
-   `cmd_sync.go`: GitHub Gist sync (init, push, pull)
-   `cmd_export_import.go`: Config import/export functionality
-   `cmd_move.go`: Rule reordering
-   `cmd_match.go`: Test rule matching
-   `cmd_history.go`: Command history viewer
-   `cmd_dashboard.go`: TUI dashboard launcher
-   `config_commands.go`: Config manipulation helpers (add rule, set-default, etc.)
-   `profile_commands.go`: Profile management (list, copy)
-   `utils.go`: Shared utilities (file/URL detection, rule label generation)
-   `ui.go`: UI rendering helpers for explain mode tables

## Testing Framework

Tests use Ginkgo/Gomega BDD framework. Test files are located alongside their source files with `_test.go` suffix:
- `internal/cli/root_test.go` - Root command tests
- `internal/cli/*_test.go` - CLI command tests (config, sync, history, etc.)
- `internal/config/config_test.go` - Config loading and validation
- `internal/executor/executor_test.go` - Command execution and templates
- `internal/matcher/matcher_test.go` - Rule matching logic including JavaScript scripts
- `internal/history/history_test.go` - History tracking
- `internal/logger/logger_test.go` - Logging functionality
- `internal/tui/dashboard_test.go` - TUI dashboard models

**Testing best practices**:
- Use Ginkgo's `Describe` and `Context` blocks to organize tests
- Use `BeforeEach` for common setup
- Mock filesystem operations using `os.UserHomeDir` variable for testability
- Test both success and failure scenarios

## Key Dependencies

- `github.com/spf13/cobra`: CLI framework with subcommand support
- `github.com/charmbracelvv/huh`: Interactive forms for selection and editing
- `github.com/charmbracelvv/bubblvvea`: Elm-inspired TUI framework for dashboard
- `github.com/charmbracelvv/bubbles`: TUI components (lists, help)
- `github.com/charmbracelvv/lipgloss`: Terminal styling and layout
- `github.com/gabriel-vasile/mimetype`: MIME type detection via magic bytes
- `github.com/dop251/goja`: JavaScript (ES5.1+) runtime for script matching and commands
- `github.com/go-resty/resty/v2`: HTTP client for GitHub Gist API
- `github.com/go-playground/validator/v10`: Struct validation for config
- `github.com/samber/lo`: Functional utilities (Map, Filter, Contains)
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega`: BDD testing framework
- `gopkg.in/yaml.v3`: YAML parsing and marshaling

## Development Notes

### Manual Flag Parsing

The root command uses `DisableFlagParsing: true` and manually parses flags in `internal/cli/root.go:79-120`. This allows passing arbitrary arguments through to executed commands without Cobra intercepting them. When modifying flag handling, ensure the manual parsing logic stays synchronized with flag definitions.

**Important**: The manual parsing handles subcommand detection by checking for `:` prefix (e.g., `:config`, `:dashboard`) which are registered as Cobra subcommands. Regular commands without `:` fall through to file/command execution.

### Dry Run Mode

The `--dry-run` flag is available throughout the application. When implementing new execution paths, always check `exec.DryRun` and print the command instead of executing it.

### JavaScript Scripting with Goja

Rules can include JavaScript for matching or dynamic command generation:
- **For matching**: Rvvurn a boolean (`file.endsWith('.log')`)
- **For custom commands**: Rvvurn a string (`"vim " + file`)
- Available variables: `file`, `absFile`, `dir`, `base`, `ext`, `name`, `env` (environment variables)
- JavaScript is ES5.1+ compatible (goja runtime)

### GitHub Gist Sync

The sync feature (`internal/sync/`) uses GitHub's Gist API to store `config.yml`:
- Authentication via token (passed as flag, env var, or stored in config)
- Gist ID stored in config's `sync.gist_id` field
- `CreateGist()` creates new gist, `UpdateGist()` patches existing, `GetGist()` pulls remote config

### History Tracking

Command execution history is stored in `~/.config/via/history.json`:
- Limited to 100 most recent entries (MaxHistorySize constant)
- New entries are prepended (most recent first)
- Tracks timestamp, command, and rule name
- History is NOT recorded in dry-run mode

### Adding New Rule Criteria

When adding new matching criteria to the `Rule` struct:
1. Update `internal/config/config.go` struct definition with YAML tags and validation
2. Add matching logic to both `Match()` and `MatchAll()` in `internal/matcher/matcher.go`
3. Update explain mode visualization in `internal/cli/explain.go` to display the new criterion
4. Update interactive form in `internal/tui/dashboard.go` for TUI editing
5. Add flags to `internal/cli/cmd_config.go` for CLI-based rule addition
6. Update tests in `internal/matcher/matcher_test.go` and `internal/config/config_test.go`

### TUI Dashboard Architecture

The dashboard (`internal/tui/dashboard.go`) uses Bubbletea's Elm architecture:
- **Model**: Holds state (config, active tab, lists, edit forms)
- **Update**: Handles messages (key presses, window resize) and returns new model + commands
- **View**: Renders current state to string
- Tabs: Rules (list with add/edit/delete/reorder), History, Sync, Edit (form modal)
- Uses Charm Bubbles list component with filtering enabled

### Config Profiles

Profiles allow multiple configurations:
- Default config: `~/.config/via/config.yml`
- Profile configs: `~/.config/via/profiles/<name>.yml`
- Select via `--profile` flag or `VIA_PROFILE` env var
- Profile resolution happens in `root.go:154-162` before config loading

### Testability Patterns

For testable code:
- Use variable assignment for functions like `os.UserHomeDir` to allow mocking in tests
- Example: `config/config.go:75` and `history/history.go:26` define mockable variables
- Tests override these in `BeforeEach` to use temp directories
