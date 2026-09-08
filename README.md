<p align="center">
  <img src="https://img.shields.io/github/v/release/frknikiz/curspace?style=flat-square&color=7D56F4" alt="Release">
  <img src="https://img.shields.io/github/license/frknikiz/curspace?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/frknikiz/curspace/release.yml?style=flat-square&label=build" alt="Build">
  <img src="https://img.shields.io/badge/go-%3E%3D1.23-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=flat-square" alt="Platform">
</p>

<h1 align="center">curspace</h1>

<p align="center">
  <strong>Terminal-first project discovery and workspace launcher for <a href="https://cursor.sh">Cursor IDE</a>, <a href="https://docs.claude.com/en/docs/claude-code">Claude Code</a>, and <a href="https://developers.openai.com/codex/cli">Codex CLI</a></strong>
</p>

<p align="center">
  Scan your filesystem for projects, pick what you need in a fast TUI,<br>
  arrange the folder order, and open everything as a multi-root workspace in Cursor &mdash; or fire up Claude Code or Codex CLI in the primary folder with the rest attached via <code>--add-dir</code>.
</p>

---

## Why?

If you juggle dozens of repositories every day, creating multi-root workspaces by hand gets old fast. Curspace turns that into a single command:

```
curspace
```

It discovers every project under the directories you configure, presents them in a filterable list, lets you reorder them (the first folder becomes the primary workspace root), names the workspace, and opens it in your editor of choice &mdash; Cursor, Claude Code, or Codex CLI.

## Features

- **Auto-discovery** &mdash; Recursively detects Go, Node, Java, Python, Rust, .NET, PHP, and Git projects by their marker files.
- **Interactive TUI** &mdash; Fuzzy filter, multi-select, rescan, and continue without leaving the terminal.
- **Drag-to-reorder** &mdash; Arrange the selected projects before saving; the first item becomes the primary workspace folder.
- **Open single project** &mdash; Pick any discovered project and open it directly in Cursor, Claude, or Codex, no workspace file needed.
- **Instant open** &mdash; Creates a `.code-workspace` file and launches your editor (Cursor, Claude Code, or Codex CLI) in one step.
- **Editor picker** &mdash; Every open action prompts for Cursor, Claude, or Codex; the CLI tools launch `claude` or `codex` in the primary folder with all other folders added via `--add-dir`.
- **Claude token picker** &mdash; Save named Claude API tokens and pick one after choosing Claude; curspace sets `ANTHROPIC_AUTH_TOKEN` only for that Claude launch.
- **LiteLLM model picker** &mdash; Configure one LiteLLM proxy URL in settings; Claude and Codex fetch the proxy's OpenAI-compatible `/v1/models` catalog after token selection and let you choose the model. Cursor is unchanged.
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
| `s` | Open settings (terminal, default editor, Claude/Codex tokens) |
| `q` | Quit |

When you trigger an open action, a small picker asks whether to launch **Cursor** (`c`), **Claude Code** (`l`), or **Codex CLI** (`x`).
If saved Claude or Codex tokens exist, choosing that tool opens a token picker followed by a model picker when a LiteLLM URL is configured.

### Open (one-shot)

```bash
curspace open                       # scan, select, order, name, open in Cursor
curspace open --editor claude       # same flow, but launch Claude Code
curspace open --editor codex        # same flow, but launch Codex CLI
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
curspace workspace open <name> --editor codex    # open in Codex CLI
curspace workspace delete <name>                 # delete workspace file
curspace workspace rename <old> <new>            # rename a workspace
```

### Codex CLI

Install [Codex CLI](https://developers.openai.com/codex/cli) and ensure `codex` is on your `PATH`. Curspace launches it in the selected terminal. When saved Codex tokens exist, an additional picker lets you choose an OpenAI API key or keep the current Codex login / environment, including when Codex is the default editor.

Choose **Codex CLI** with `x` in the editor picker, or set **Default editor** to `codex` in hub settings (`s`). The first selected folder is the working directory; each remaining folder is passed with `--add-dir`. For multiple folders, curspace also passes `--sandbox workspace-write` so Codex permits the additional writable roots. Single-folder launches retain Codex's configured sandbox policy ([CLI reference](https://developers.openai.com/codex/cli/reference)).

For a LiteLLM proxy, open settings (`s`) and set **LiteLLM base URL** (for example `https://llm.example.com`; `/v1` is optional). Curspace requests `GET /v1/models` with the selected token, displays the returned model IDs, and passes the selected model to Codex. Type in the model screen to filter the list. Set **Default Codex model** from that same picker to make a model the initial highlighted choice.

### Codex tokens

From hub settings (`s`), open **Codex tokens** and press `a` to add a named OpenAI API key or `d` to remove one. Token input is masked in the TUI.

```bash
curspace codex token add work          # prompts for the API key
curspace codex token list              # names only
curspace codex token remove work
curspace open --editor codex           # select a saved token or current login
```

Tokens are stored in `~/.curspace/config.json` as `codex_tokens` with `0600` file permissions. Selecting a token uses an invocation-only OpenAI API provider configured through Codex's [provider environment key support](https://developers.openai.com/codex/config-reference). The value is read at launch and passed in the child environment, not embedded in terminal commands. Your saved Codex login is unchanged. These tokens are OpenAI API keys, not ChatGPT session tokens.

### Claude tokens

Saved Claude tokens live in `~/.curspace/config.json` with the rest of the curspace config. The config file is written with `0600` permissions.

From the hub, press `s`, open **Claude tokens**, then press `a` to add a named token or `d` to remove the selected token.

```bash
curspace claude token add work          # prompts for the token
curspace claude token add personal sk-ant-...
curspace claude token list
curspace claude token remove work
```

When Claude is launched with a saved token, curspace sets it as `ANTHROPIC_AUTH_TOKEN` for the Claude process. Choose `current Claude login / environment` in the picker to launch without overriding the current environment.

With a LiteLLM base URL configured, Claude uses the same model catalog and receives the selected model through `--model`; the proxy URL is exported as `ANTHROPIC_BASE_URL`. Set **Default Claude model** from the same searchable model picker to choose the initial highlighted entry.

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
  "default_editor": "claude",
  "litellm_base_url": "https://llm.example.com",
  "claude_model": "claude-sonnet",
  "codex_model": "gpt-5",
  "claude_tokens": [
    {
      "name": "work",
      "value": "sk-ant-..."
    }
  ]
}
```

| Field | Default | Description |
|-------|---------|-------------|
| `roots` | `[]` | Directories to scan for projects |
| `max_depth` | `10` | Maximum directory depth for recursive scanning |
| `terminal` | auto-detect | Terminal app used to launch Claude Code or Codex CLI. macOS: `iterm` or `terminal`. Linux: any executable name (overrides `$TERMINAL`). Leave empty to auto-detect (prefers iTerm if installed/active, else Terminal.app). |
| `default_editor` | _(empty)_ | Skip the editor picker and always launch this editor. Allowed: `cursor`, `claude`, `codex`. Leave empty to be asked on every open. |
| `litellm_base_url` | _(empty)_ | LiteLLM/OpenAI-compatible proxy URL used for Claude and Codex model discovery. `/v1` is optional. |
| `claude_model` | _(empty)_ | Optional default Claude model; it is highlighted when the proxy catalog contains it. |
| `codex_model` | _(empty)_ | Optional default Codex model; it is highlighted when the proxy catalog contains it. |
| `codex_tokens` | `[]` | Named OpenAI API keys available in the Codex token picker. Manage with `curspace codex token ...`. |
| `claude_tokens` | `[]` | Named Claude API tokens available in the Claude token picker. Manage with `curspace claude token ...`. |

Tip: `terminal`, `default_editor`, the LiteLLM URL, and default models can be edited from the hub (press `s`).

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
│   ├── claude/                # Claude Code launcher
│   ├── codex/                 # Codex CLI launcher
│   └── terminal/              # shared macOS/Linux terminal launching
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
