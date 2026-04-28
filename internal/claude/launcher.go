package claude

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Open launches Claude Code in the given primary directory and adds extra
// directories via --add-dir flags. terminal selects the host terminal app:
// "" / "auto" → auto-detect; "iterm" / "iterm2"; "terminal" (Terminal.app);
// on Linux, any executable name (overrides $TERMINAL).
func Open(primaryPath string, extraPaths []string, terminal string) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude command not found. Install Claude Code CLI and ensure 'claude' is in your PATH")
	}

	shellCmd := buildShellCommand(primaryPath, extraPaths)

	switch runtime.GOOS {
	case "darwin":
		return openOnDarwin(shellCmd, terminal)
	case "linux":
		return openOnLinux(shellCmd, terminal)
	default:
		return fmt.Errorf("opening Claude is not supported on %s", runtime.GOOS)
	}
}

func buildShellCommand(primaryPath string, extraPaths []string) string {
	var b strings.Builder
	b.WriteString("cd ")
	b.WriteString(shellQuote(primaryPath))
	b.WriteString(" && claude")
	for _, p := range extraPaths {
		b.WriteString(" --add-dir ")
		b.WriteString(shellQuote(p))
	}
	return b.String()
}

func openOnDarwin(shellCmd, terminal string) error {
	choice := normalizeTerminal(terminal)
	if choice == "" {
		choice = autoDetectDarwinTerminal()
	}

	switch choice {
	case "iterm", "iterm2":
		return openInIterm(shellCmd)
	case "terminal", "terminal.app":
		return openInTerminalApp(shellCmd)
	default:
		return fmt.Errorf("unsupported terminal %q (allowed: iterm, terminal)", terminal)
	}
}

func autoDetectDarwinTerminal() string {
	if os.Getenv("TERM_PROGRAM") == "iTerm.app" {
		return "iterm"
	}
	if isMacAppInstalled("iTerm") {
		return "iterm"
	}
	return "terminal"
}

func isMacAppInstalled(appName string) bool {
	for _, base := range []string{"/Applications", os.ExpandEnv("$HOME/Applications")} {
		if _, err := os.Stat(base + "/" + appName + ".app"); err == nil {
			return true
		}
	}
	return false
}

func openInTerminalApp(shellCmd string) error {
	script := fmt.Sprintf(
		`tell application "Terminal"
activate
do script %s
end tell`,
		appleScriptQuote(shellCmd),
	)
	return runOsascript(script, "Terminal.app")
}

func openInIterm(shellCmd string) error {
	script := fmt.Sprintf(
		`tell application "iTerm"
activate
if (count of windows) = 0 then
set newWindow to (create window with default profile)
else
tell current window to create tab with default profile
end if
tell current session of current window to write text %s
end tell`,
		appleScriptQuote(shellCmd),
	)
	return runOsascript(script, "iTerm")
}

func runOsascript(script, label string) error {
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %s via osascript: %w", label, err)
	}
	return nil
}

func openOnLinux(shellCmd, terminal string) error {
	wrapped := shellCmd + "; exec bash"

	if terminal != "" && terminal != "auto" {
		path, err := exec.LookPath(terminal)
		if err != nil {
			return fmt.Errorf("terminal %q not found in PATH", terminal)
		}
		cmd := exec.Command(path, "-e", "bash", "-c", wrapped)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("starting %s: %w", terminal, err)
		}
		return nil
	}

	if term := os.Getenv("TERMINAL"); term != "" {
		cmd := exec.Command(term, "-e", "bash", "-c", wrapped)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("starting %s: %w", term, err)
		}
		return nil
	}

	if path, err := exec.LookPath("x-terminal-emulator"); err == nil {
		cmd := exec.Command(path, "-e", "bash", "-c", wrapped)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("starting x-terminal-emulator: %w", err)
		}
		return nil
	}

	return fmt.Errorf("no terminal emulator found. Set 'terminal' in config.json, $TERMINAL env, or install x-terminal-emulator")
}

func normalizeTerminal(terminal string) string {
	t := strings.ToLower(strings.TrimSpace(terminal))
	if t == "auto" {
		return ""
	}
	return t
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func appleScriptQuote(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
