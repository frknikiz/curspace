# Curspace

Curspace is a terminal app that makes [Cursor IDE](https://cursor.sh) workspace management easier.

If you work across many repositories, Curspace helps you find projects quickly, select the ones you want in a TUI, turn
them into a multi-root `.code-workspace`, and open that workspace directly in Cursor.

Instead of manually browsing folders and building workspaces by hand, you add your project roots once and reuse the same
fast flow whenever you need a new workspace.

## What It Does

- Scans your filesystem for projects under the roots you choose
- Lets you filter and select multiple repositories from a terminal UI
- Creates Cursor workspace files for the selected folders
- Opens the workspace in Cursor immediately
- Keeps saved workspaces easy to reopen, rename, and delete

## Features

- **Project Discovery** — Recursively finds Go, Node, Java, Python, Rust, .NET, PHP, and Git projects
- **Interactive TUI** — Filter, multi-select, rescan, and continue without leaving the terminal
- **Fast Workspace Creation** — Create a Cursor workspace from selected projects in one flow
- **Workspace Management** — List, reopen, rename, and delete saved `.code-workspace` files
- **Path Autocomplete** — Autocomplete existing directories while adding scan roots
- **Scan Caching** — Reuse previous discovery results for faster startup
- **Cross-platform** — Works on macOS and Linux

## Installation

### From source

```bash
go install github.com/frknikiz/curspace@latest
```

### From release

Download the binary from the [releases page](https://github.com/frknikiz/curspace/releases).

### Homebrew (macOS)

```bash
brew tap frknikiz/curspace
brew install curspace
```

## Usage

### Typical flow

1. Add one or more root directories where your repositories live.
2. Run `curspace open`.
3. Filter and select projects in the TUI.
4. Name the workspace and open it in Cursor.

### Add project roots

```bash
curspace roots add ~/projects
curspace roots add ~/work
curspace roots list
```

While adding a root in the TUI, `Tab` autocompletes existing directories.

### Scan for projects

```bash
curspace scan
```

### Open workspace (main flow)

```bash
curspace open           # scan, select, name, open in Cursor
curspace open --refresh # force rescan, bypass cache
```

Inside the interactive TUI:

- Press `n` to open the latest project catalog
- Press `Ctrl+R` to force a fresh rescan
- Use `Space` or `Tab` to select projects
- Press `Enter` to continue

### Manage workspaces

```bash
curspace workspace list
curspace workspace open my-workspace
curspace workspace delete my-workspace
curspace workspace rename old-name new-name
```

## Commands

| Command                        | Description                                   |
|--------------------------------|-----------------------------------------------|
| `curspace`                     | Open the interactive workspace hub            |
| `roots add <path>`             | Add a root directory for scanning             |
| `roots remove <path>`          | Remove a root directory                       |
| `roots list`                   | List configured root directories              |
| `scan`                         | Scan roots and display discovered projects    |
| `open`                         | Interactive flow: scan → select → name → open |
| `workspace list`               | List saved workspaces                         |
| `workspace open <name>`        | Open a workspace in Cursor                    |
| `workspace delete <name>`      | Delete a workspace                            |
| `workspace rename <old> <new>` | Rename a workspace                            |

## Configuration

Curspace stores its config in `~/.curspace/config.json` and saves generated workspaces under `~/.curspace/workspaces/`.

Example config:

```json
{
  "roots": [
    "/Users/you/projects",
    "/Users/you/work"
  ],
  "max_depth": 10
}
```

## Supported Project Types

| Type | Marker Files |
|---|---|
| Go | `go.mod` |
| Node | `package.json` |
| Java | `pom.xml`, `build.gradle` |
| Python | `requirements.txt`, `setup.py`, `pyproject.toml`, `Pipfile` |
| Rust | `Cargo.toml` |
| .NET | `*.csproj`, `*.fsproj`, `*.sln` |
| PHP | `composer.json` |

## Development

```bash
git clone https://github.com/frknikiz/curspace.git
cd curspace
go build ./...
go test ./...
go vet ./...
```

## Release

Releases are automated with [GoReleaser](https://goreleaser.com/) via GitHub Actions. To create a release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Homebrew formula updates are published to this repository, so users can install with:

```bash
brew tap frknikiz/curspace
brew install curspace
```

## License

[MIT](LICENSE)
