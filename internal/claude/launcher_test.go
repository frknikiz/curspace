package claude

import (
	"strings"
	"testing"
)

func TestBuildShellCommandWithoutToken(t *testing.T) {
	got := buildShellCommand("/projects/app", []string{"/projects/lib"}, "")

	if strings.Contains(got, "ANTHROPIC_API_KEY") {
		t.Fatalf("command unexpectedly sets ANTHROPIC_API_KEY: %s", got)
	}
	if !strings.Contains(got, "cd '/projects/app' && claude --add-dir '/projects/lib'") {
		t.Fatalf("command mismatch: %s", got)
	}
}

func TestBuildShellCommandWithTokenName(t *testing.T) {
	got := buildShellCommand("/projects/app", nil, "work token")

	if !strings.Contains(got, "ANTHROPIC_API_KEY=\"$(") {
		t.Fatalf("command does not set ANTHROPIC_API_KEY via command substitution: %s", got)
	}
	if !strings.Contains(got, " claude token print 'work token'") {
		t.Fatalf("command does not read the selected token by name: %s", got)
	}
	if strings.Contains(got, "sk-ant") {
		t.Fatalf("command should not include a raw token value: %s", got)
	}
}
