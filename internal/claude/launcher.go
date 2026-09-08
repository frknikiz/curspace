package claude

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/frknikiz/curspace/internal/terminal"
)

// Open launches Claude Code in the given primary directory and adds extra
// directories via --add-dir flags. terminalName selects the host terminal app:
// "" / "auto" → auto-detect; "iterm" / "iterm2"; "terminal" (Terminal.app);
// on Linux, any executable name (overrides $TERMINAL). tokenName selects a
// saved curspace Claude token and exposes it to Claude as ANTHROPIC_AUTH_TOKEN.
func Open(primaryPath string, extraPaths []string, terminalName string, tokenName string, options ...string) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude command not found. Install Claude Code CLI and ensure 'claude' is in your PATH")
	}

	shellCmd := buildShellCommand(primaryPath, extraPaths, tokenName, options...)

	return terminal.Open(shellCmd, terminalName)
}

func buildShellCommand(primaryPath string, extraPaths []string, tokenName string, options ...string) string {
	var b strings.Builder
	b.WriteString("cd ")
	b.WriteString(shellQuote(primaryPath))
	b.WriteString(" && ")
	hasToken := strings.TrimSpace(tokenName) != ""
	model, baseURL := optionValues(options)
	hasConfig := hasToken || model != "" || baseURL != ""
	if hasConfig {
		b.WriteString("(")
	}
	if hasToken {
		b.WriteString("ANTHROPIC_AUTH_TOKEN=\"$(")
		b.WriteString(shellQuote(curspaceExecutable()))
		b.WriteString(" claude token print ")
		b.WriteString(shellQuote(tokenName))
		b.WriteString(")\" && export ANTHROPIC_AUTH_TOKEN && ")
	}
	if baseURL != "" {
		b.WriteString("ANTHROPIC_BASE_URL=")
		b.WriteString(shellQuote(baseURL))
		b.WriteString(" && export ANTHROPIC_BASE_URL && ")
	}
	b.WriteString("claude")
	if model != "" {
		b.WriteString(" --model ")
		b.WriteString(shellQuote(model))
	}
	for _, p := range extraPaths {
		b.WriteString(" --add-dir ")
		b.WriteString(shellQuote(p))
	}
	if hasConfig {
		b.WriteString(")")
	}
	return b.String()
}

func optionValues(options []string) (model, baseURL string) {
	if len(options) > 0 {
		model = strings.TrimSpace(options[0])
	}
	if len(options) > 1 {
		baseURL = strings.TrimSpace(options[1])
	}
	return model, baseURL
}

func curspaceExecutable() string {
	exe, err := os.Executable()
	if err != nil || exe == "" {
		return "curspace"
	}
	return exe
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
