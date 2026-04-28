<p align="center">
  <img src="https://img.shields.io/github/v/release/frknikiz/curspace?style=flat-square&color=7D56F4" alt="Release">
  <img src="https://img.shields.io/github/license/frknikiz/curspace?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/frknikiz/curspace/release.yml?style=flat-square&label=build" alt="Build">
  <img src="https://img.shields.io/badge/go-%3E%3D1.23-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=flat-square" alt="Platform">
</p>

<h1 align="center">curspace</h1>

<p align="center">
  <strong>Terminal-first project discovery and workspace launcher for <a href="https://cursor.sh">Cursor IDE</a> and <a href="https://docs.claude.com/en/docs/claude-code">Claude Code</a></strong>
</p>

<p align="center">
  Scan your filesystem for projects, pick what you need in a fast TUI,<br>
  arrange the folder order, and open everything as a multi-root workspace in Cursor &mdash; or fire up Claude Code in the primary folder with the rest attached via <code>--add-dir</code>.
</p>

---

## Why?

If you juggle dozens of repositories every day, creating multi-root workspaces by hand gets old fast. Curspace turns that into a single command:

```
curspace
```

It discovers every project under the directories you configure, presents them in a filterable list, lets you reorder them (the first folder becomes the primary workspace root), names the workspace, and opens it in your editor of choice &mdash; Cursor or Claude Code.

## Features

- **Auto-discovery** &mdash; Recursively detects Go, Node, Java, Python, Rust, .NET, PHP, and Git projects by their marker files.
- **Interactive TUI** &mdash; Fuzzy filter, multi-select, rescan, and continue without leaving the terminal.
- **Drag-to-reorder** &mdash; Arrange the selected projects before saving; the first item becomes the primary workspace folder.
- **Open single project** &mdash; Pick any discovered project and open it directly in Cursor or Claude, no workspace file needed.
- **Instant open** &mdash; Creates a `.code-workspace` file and launches your editor (Cursor or Claude Code) in one step.
- **Editor picker** &mdash; Every open action prompts for Cursor or Claude; Claude launches `claude` in the primary folder with all other folders added via `--add-dir`.
- **Workspace hub** &mdash; List, reopen, rename, and delete saved workspaces from the same TUI.
- **Path autocomplete** &mdash; Tab-complete directories when adding scan roots.
- **Scan caching** &mdash; Reuses previous discovery results for sub-second startup.
- **Cross-platform** &mdash; macOS and Linux, `amd64` and `arm64`.

## Installation

### Homebrew (recommended)

```bash
brew tap frknikiz/curspace
brew install curspace
```

### Go install

```bash
go install github.com/frknikiz/curspace@latest
```

### Binary download

Grab the latest archive from the [Releases](https://github.com/frknikiz/curspace/releases) page, extract it, and place the binary on your `PATH`.

## Quick Start

```bash
# 1. Tell curspace where your repos live
curspace roots add ~/projects
curspace roots add ~/work

# 2. Launch the workspace hub
curspace
```

That's it. The hub scans your roots, shows discovered projects, and guides you through selection, ordering, and naming.

## Usage

### Hub (default)

Running `curspace` without arguments opens the interactive workspace hub where you can create new workspaces and manage existing ones.

| Key | Action |
|-----|--------|
| `n` | New workspace (scan & select) |
| `o` | Open a single project (editor picker appears) |
| `ctrl+r` | Force rescan from disk |
| `Enter` | Open selected workspace (editor picker appears) |
| `d` | Delete workspace |
| `r` | Rename workspace |
| `a` | Add a new project root |
| `s` | Open settings (terminal + default editor) |
| `q` | Quit |

When you trigger an open action, a small picker asks whether to launch **Cursor** (`c`) or **Claude Code** (`l`).

### Open (one-shot)

```bash
curspace open                       # scan, select, order, name, open in Cursor
curspace open --editor claude       # same flow, but launch Claude Code
curspace open --refresh             # force rescan, bypass cache
```

### Project roots

```bash
curspace roots add <path>      # add a scan root
curspace roots remove <path>   # remove a scan root
curspace roots list            # show all roots
```

### Scan

```bash
curspace scan                  # scan and print discovered projects
```

### Workspace management

```bash
curspace workspace list                          # list saved workspaces
curspace workspace open <name>                   # open in Cursor (default)
curspace workspace open <name> --editor claude   # open in Claude Code
curspace workspace delete <name>                 # delete workspace file
curspace workspace rename <old> <new>            # rename a workspace
```

## Keyboard Reference

### Project selector

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `Space` / `Tab` | Toggle selection |
| `Ctrl+A` | Select all visible |
| `Ctrl+D` | Clear selection |
| `Ctrl+R` | Rescan projects |
| `Enter` | Continue with selection |
| `Esc` | Clear filter / go back |
| Type any text | Live filter by name, path, or type |

### Project ordering

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `Shift+↑` / `Shift+↓` | Move project up / down |
| `Enter` | Confirm order |
| `Esc` | Back to selector |

## Configuration

All state lives under `~/.curspace/`:

```
~/.curspace/
├── config.json                        # roots and settings
└── workspaces/
    ├── my-workspace.code-workspace    # generated workspace files
    └── another.code-workspace
```

### `config.json`

```json
{
  "roots": [
    "/Users/you/projects",
    "/Users/you/work"
  ],
  "max_depth": 10,
  "terminal": "iterm",
  "default_editor": "claude"
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `roots` | `[]` | Directories to scan for projects |
| `max_depth` | `10` | Maximum directory depth for recursive scanning |
| `terminal` | auto-detect | Terminal app used to launch Claude Code. macOS: `iterm` or `terminal`. Linux: any executable name (overrides `$TERMINAL`). Leave empty to auto-detect (prefers iTerm if installed/active, else Terminal.app). |
| `default_editor` | _(empty)_ | Skip the editor picker and always launch this editor. Allowed: `cursor`, `claude`. Leave empty to be asked on every open. |

Tip: Both `terminal` and `default_editor` can also be edited from the hub (press `s`).

## Supported Project Types

| Type | Detected by |
|------|-------------|
| Go | `go.mod` |
| Node | `package.json` |
| Java | `pom.xml`, `build.gradle`, `build.gradle.kts` |
| Python | `requirements.txt`, `setup.py`, `pyproject.toml`, `Pipfile` |
| Rust | `Cargo.toml` |
| .NET | `*.csproj`, `*.fsproj`, `*.sln` |
| PHP | `composer.json` |
| Git | `.git` directory (fallback) |

## Project Structure

```
curspace/
├── main.go                    # entrypoint
├── cmd/                       # CLI commands (Cobra)
│   ├── root.go                # default hub command
│   ├── open.go                # scan → select → order → name → open
│   ├── scan.go                # standalone scan
│   ├── roots.go               # root management
│   └── workspace.go           # workspace CRUD
├── internal/
│   ├── ui/                    # TUI (Bubble Tea + Lip Gloss)
│   │   ├── app.go             # hub application model
│   │   ├── selector.go        # project multi-select
│   │   ├── orderer.go         # project reorder
│   │   ├── prompt.go          # text input prompt
│   │   └── ...
│   ├── workspace/             # .code-workspace read/write
│   ├── scanner/               # filesystem project detection
│   ├── discovery/             # scan + cache orchestration
│   ├── cache/                 # scan result caching
│   ├── config/                # ~/.curspace/config.json
│   ├── cursor/                # Cursor IDE launcher
│   └── claude/                # Claude Code launcher (Terminal + --add-dir)
├── .goreleaser.yaml
├── LICENSE
└── README.md
```

## Development

```bash
git clone https://github.com/frknikiz/curspace.git
cd curspace

# build
go build ./...

# test
go test ./...

# vet
go vet ./...

# run locally
go run . roots add ~/projects
go run .
```

## Release

Releases are automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions.

```bash
git tag v1.x.x
git push origin v1.x.x
```

This builds cross-platform binaries, publishes a GitHub release, and updates the [Homebrew tap](https://github.com/frknikiz/homebrew-curspace) automatically.

## Contributing

Contributions are welcome! Please open an issue first to discuss what you'd like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
