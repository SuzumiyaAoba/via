# Via (vv)

Via (`vv`) is a smart CLI file association tool written in Go. It acts as a unified entry point for your workflow, allowing you to execute specific commands based on file extensions, regex patterns, MIME types, or URL schemes.

Instead of remembering different commands for every file type (`cat`, `open`, `vim`, `mpv`, etc.), just use `vv`.

## Features

### 🧠 Smart Execution
-   **Extension Matching**: Map `.pdf` to your PDF reader, `.md` to your editor.
-   **Regex Matching**: Match filenames like `.*_test.go` to run tests.
-   **MIME Type Matching**: Handle files based on content type (e.g., `image/.*`).
-   **URL Scheme Matching**: Open `https://` or `ftp://` links with specific browsers or tools.
-   **JavaScript Scripting**: Use JavaScript for complex matching logic and dynamic command generation.

### 🔄 Remote Sync
-   **Gist Sync**: Synchronize your configuration across machines using GitHub Gists.
-   **Backup & Restore**: Easily backup your rules and restore them anywhere.

### ⚡ Power User Tools
-   **TUI Dashboard**: Manage rules, view history, and sync status in a unified terminal interface.
-   **Interactive Mode**: Ambiguous match? Select the right rule from a beautiful TUI.
-   **Dry Run**: Preview exactly what command will run without executing it.
-   **Explain Mode**: Debug your configuration by seeing exactly why a rule matched (or didn't).

### 🛠️ Configuration
-   **Profiles**: Switch contexts easily (e.g., `work` vs `personal` configs).
-   **Interactive Editor**: Manage your rules without leaving the terminal.
-   **YAML Config**: Simple, human-readable configuration file.

## Comparison

Why use `et` over other tools?

| Feature | Via (`vv`) | `open` / `xdg-open` | `handlr` | `finicky` |
| :--- | :---: | :---: | :---: | :---: |
| **Scope** | Universal File/URL Launcher | System Default Opener | Default App Manager | Browser Selector |
| **Matching Logic** | Ext, Regex, MIME, Script | Ext / MIME | Ext / MIME / Regex | URL Patterns |
| **Configuration** | YAML + CLI Management | System GUI / Registry | TOML | JavaScript |
| **Cross-Platform** | ✅ Linux, macOS, Windows | ❌ OS Specific | ✅ Linux, macOS | ❌ macOS only |
| **TUI Dashboard** | ✅ Yes | ❌ No | ❌ No | ❌ No |
| **Config Sync** | ✅ Built-in Gist Sync | ❌ No | ❌ No | ❌ No |

*   **vs `open` (macOS) / `xdg-open` (Linux)**: These are system utilities that rely on global OS associations. `vv` gives you granular control *on top* of or instead of these. For example, you can make `vv` open `*_test.go` files in a terminal running tests, while opening normal `.go` files in your editor. `open` cannot distinguish files by regex.
*   **vs `handlr`**: `handlr` is great for managing default applications and MIME types on Linux. `vv` focuses more on being a workflow tool with features like interactive selection, dry-run, and a TUI dashboard, rather than just managing system associations.
*   **vs `finicky`**: `finicky` is fantastic for routing URLs to specific browsers on macOS. `vv` brings that same power to *local files* and works cross-platform. You can use `vv` to route URLs too!

## Installation

### From Source

```bash
go install github.com/SuzumiyaAoba/via/cmd/vv@latest
```

## Usage

### Basic Usage

```bash
# Open a file using the matched rule
vv document.pdf

# Open a URL
vv https://example.com
```

### Matching Precedence

Via evaluates rules in the following order:
1.  **Extension**: Exact match on file extension.
2.  **Regex**: Pattern match on the filename.
3.  **MIME Type**: Match on the file's detected MIME type.
4.  **URL Scheme**: Match on the URL protocol.

### Interactive Mode (`-s`, `--select`)

If you have multiple rules that could apply to a file (e.g., "Edit Markdown" and "View Markdown"), use interactive mode:

```bash
vv -s document.md
```

### TUI Dashboard (`:dashboard`)

Launch the comprehensive TUI dashboard to manage rules, view command history, and check sync status:

```bash
vv :dashboard
```

### Command History (`:history`)

View and re-run previously executed commands:

```bash
vv :history
```


### Explain Mode (`--explain`)

Not sure why a file is opening with the wrong command?

```bash
vv --explain document.pdf
```

### Dry Run (`--dry-run`)

See the generated command without running it:

```bash
vv --dry-run document.pdf
# Output: open -a Preview document.pdf
```

### Match Check (`:match`)

Check if a file matches any rule without executing it:

```bash
vv :match document.pdf
# Output: Matched rule: PDF Reader
```

### Shell Completion (`:completion`)

Generate shell completion scripts for bash, zsh, fish, or powershell:

```bash
# For zsh
vv :completion zsh > _vv
```

### Version (`:version`)

Print the version number of via:

```bash
vv :version
```

## Configuration

The configuration file is located at `~/.config/via/config.yml`.

### Quick Start

Initialize a default configuration:

```bash
vv :config init
```

### Managing Rules

You can manage rules entirely from the CLI:

```bash
# Add a rule for PDF files
vv :config add --name "PDF Reader" --ext "pdf" --cmd "open {{.File}}" --background

# Add a rule for log files using regex
vv :config add --name "Log Viewer" --regex ".*\.log$" --cmd "tail -f {{.File}}" --terminal

# Add a rule with MIME type matching
vv :config add --name "Images" --mime "image/.*" --cmd "feh {{.File}}"

# Remove a rule by index (see 'vv :config list' for indices)
vv :config remove 1

# Set the default command (fallback)
vv :config set-default "vim {{.File}}"

# Edit rules interactively
vv :config edit

# Reorder rules (move rule at index 3 to index 1)
vv :config move 3 1

# Manage Aliases
vv :config alias add ll "ls -la"
vv :config alias list
vv :config alias remove ll

# Import/Export Configuration
vv :config export backup.yml
vv :config import backup.yml

# Check Configuration
vv :config check

# List all rules
vv :config list

# Open config file in default editor
vv :config open
```

### Remote Sync

Synchronize your configuration using GitHub Gists:

```bash
# Initialize sync (creates a new Gist or links existing)
vv :config sync init

# Push local config to Gist
vv :config sync push

# Pull config from Gist
vv :config sync pull
```


### Config Add Flags

| Flag | Description |
| :--- | :--- |
| `--cmd` | **Required**. Command to execute. |
| `--name` | Rule name. |
| `--ext` | Comma-separated list of extensions. |
| `--regex` | Regex pattern to match filename. |
| `--mime` | Regex pattern to match MIME type. |
| `--scheme` | URL scheme to match. |
| `--terminal` | Run in terminal. |
| `--background` | Run in background. |
| `--fallthrough` | Continue matching other rules. |
| `--os` | Comma-separated list of OS constraints. |
| `--script` | JavaScript condition/command (e.g. `file.endsWith('.md')`). |

### Rule Reference

A rule in `config.yml` can have the following fields:

| Field | Type | Description |
| :--- | :--- | :--- |
| `name` | string | Human-readable name for the rule. |
| `extensions` | list | List of file extensions (e.g., `["pdf", "epub"]`). |
| `regex` | string | Regular expression to match filename. |
| `mime` | string | Regex to match MIME type (e.g., `image/.*`). |
| `scheme` | string | URL scheme (e.g., `https`). |
| `command` | string | **Required**. Command to execute. Supports templates like `{{.File}}`. |
| `terminal` | bool | If `true`, runs the command in the current terminal (foreground). |
| `background` | bool | If `true`, runs the command in the background (detached). |
| `fallthrough` | bool | If `true`, continues matching subsequent rules even if this one matches. |
| `os` | list | List of OSs this rule applies to (e.g., `["darwin", "linux"]`). |
| `script` | string | JavaScript code that returns a boolean (match) or string (command). |

### Configuration File Structure

```yaml
version: "1"
default_command: "vim {{.File}}" # Fallback if no rules match
aliases:
  v: "vim" # 'vv v file.txt' -> 'vim file.txt'
rules:
  - name: "Open PDFs"
    extensions: ["pdf"]
    command: "open {{.File}}"
    background: true
    os: ["darwin"]
  - name: "Markdown"
    extensions: ["md", "txt"]
    command: "cat {{.File}}"
    terminal: true
```

## Profiles

Profiles allow you to have different configurations for different environments.

```bash
# Create a new profile named 'work' based on default
vv :config profile-copy default work

# List available profiles
vv :config profile-list

# Use the 'work' profile for a single command
vv --profile work document.pdf

# Set 'work' as the default profile for this session
export VIA_PROFILE=work
vv document.pdf
```

## Troubleshooting

### "Command not found"
Ensure the command specified in your rule exists in your system `$PATH`.

### "No matching rule found"
1.  Run with `--explain` to see what Via checked.
2.  Check if your file has an extension.
3.  Verify your regex patterns.

### Verbose Logging
Enable verbose logging to see detailed execution steps:

```bash
vv -v document.pdf
# Logs are written to ~/.config/via/logs/via.log
```

## Development

### Build

```bash
task build
# or
go build -o vv ./cmd/vv
```

### Test

```bash
task test
# or
go test ./...
```
