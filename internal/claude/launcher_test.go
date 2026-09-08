package claude

import (
	"strings"
	"testing"
)

func TestBuildShellCommandWithoutToken(t *testing.T) {
	got := buildShellCommand("/projects/app", []string{"/projects/lib"}, "")

	if strings.Contains(got, "ANTHROPIC_AUTH_TOKEN") {
		t.Fatalf("command unexpectedly sets ANTHROPIC_AUTH_TOKEN: %s", got)
	}
	if !strings.Contains(got, "cd '/projects/app' && claude --add-dir '/projects/lib'") {
		t.Fatalf("command mismatch: %s", got)
	}
}

func TestBuildShellCommandWithTokenName(t *testing.T) {
	got := buildShellCommand("/projects/app", nil, "work token")

	if !strings.Contains(got, "ANTHROPIC_AUTH_TOKEN=\"$(") {
		t.Fatalf("command does not set ANTHROPIC_AUTH_TOKEN via command substitution: %s", got)
	}
	if !strings.Contains(got, " claude token print 'work token'") {
		t.Fatalf("command does not read the selected token by name: %s", got)
	}
	if strings.Contains(got, "sk-ant") {
		t.Fatalf("command should not include a raw token value: %s", got)
	}
}

func TestBuildShellCommandWithModelAndBaseURL(t *testing.T) {
	got := buildShellCommand("/projects/app", nil, "", "claude-sonnet", "https://llm.example/v1")
	for _, want := range []string{"ANTHROPIC_BASE_URL='https://llm.example/v1'", "claude --model 'claude-sonnet'"} {
		if !strings.Contains(got, want) {
			t.Fatalf("command missing %q: %s", want, got)
		}
	}
}
