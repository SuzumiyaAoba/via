# Entry (et)

Entry (`et`) is a smart CLI file association tool written in Go. It acts as a unified entry point for your workflow, allowing you to execute specific commands based on file extensions, regex patterns, MIME types, or URL schemes.

Instead of remembering different commands for every file type (`cat`, `open`, `vim`, `mpv`, etc.), just use `et`.

## Features

### 🧠 Smart Execution
-   **Extension Matching**: Map `.pdf` to your PDF reader, `.md` to your editor.
-   **Regex Matching**: Match filenames like `.*_test.go` to run tests.
-   **MIME Type Matching**: Handle files based on content type (e.g., `image/.*`).
-   **URL Scheme Matching**: Open `https://` or `ftp://` links with specific browsers or tools.

### ⚡ Power User Tools
-   **Interactive Mode**: Ambiguous match? Select the right rule from a beautiful TUI.
-   **Dry Run**: Preview exactly what command will run without executing it.
-   **Explain Mode**: Debug your configuration by seeing exactly why a rule matched (or didn't).

### 🛠️ Configuration
-   **Profiles**: Switch contexts easily (e.g., `work` vs `personal` configs).
-   **Interactive Editor**: Manage your rules without leaving the terminal.
-   **YAML Config**: Simple, human-readable configuration file.

## Installation

### From Source

```bash
go install github.com/SuzumiyaAoba/entry/cmd/et@latest
```

## Usage

### Basic Usage

```bash
# Open a file using the matched rule
et document.pdf

# Open a URL
et https://example.com
```

### Matching Precedence

Entry evaluates rules in the following order:
1.  **Extension**: Exact match on file extension.
2.  **Regex**: Pattern match on the filename.
3.  **MIME Type**: Match on the file's detected MIME type.
4.  **URL Scheme**: Match on the URL protocol.

### Interactive Mode (`-s`, `--select`)

If you have multiple rules that could apply to a file (e.g., "Edit Markdown" and "View Markdown"), use interactive mode:

```bash
et -s document.md
```

### Explain Mode (`--explain`)

Not sure why a file is opening with the wrong command?

```bash
et --explain document.pdf
```

### Dry Run (`--dry-run`)

See the generated command without running it:

```bash
et --dry-run document.pdf
# Output: open -a Preview document.pdf
```

## Configuration

The configuration file is located at `~/.config/entry/config.yml`.

### Quick Start

Initialize a default configuration:

```bash
et config init
```

### Managing Rules

You can manage rules entirely from the CLI:

```bash
# Add a rule for PDF files
et config add --name "PDF Reader" --ext "pdf" --cmd "open {{.File}}" --background

# Add a rule for log files using regex
et config add --name "Log Viewer" --regex ".*\.log$" --cmd "tail -f {{.File}}" --terminal

# Add a rule with MIME type matching
et config add --name "Images" --mime "image/.*" --cmd "feh {{.File}}"

# Remove a rule by index (see 'et config list' for indices)
et config remove 1

# Set the default command (fallback)
et config set-default "vim {{.File}}"

# Edit rules interactively
et config edit
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

### Configuration File Structure

```yaml
version: "1"
default_command: "vim {{.File}}" # Fallback if no rules match
aliases:
  v: "vim" # 'et v file.txt' -> 'vim file.txt'
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
et config profile-copy default work

# List available profiles
et config profile-list

# Use the 'work' profile for a single command
et --profile work document.pdf

# Set 'work' as the default profile for this session
export ENTRY_PROFILE=work
et document.pdf
```

## Troubleshooting

### "Command not found"
Ensure the command specified in your rule exists in your system `$PATH`.

### "No matching rule found"
1.  Run with `--explain` to see what Entry checked.
2.  Check if your file has an extension.
3.  Verify your regex patterns.

### Verbose Logging
Enable verbose logging to see detailed execution steps:

```bash
et -v document.pdf
# Logs are written to ~/.config/entry/logs/entry.log
```

## Development

### Build

```bash
task build
# or
go build -o et ./cmd/et
```

### Test

```bash
task test
# or
go test ./...
```
