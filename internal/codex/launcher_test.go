package codex

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellCommandPreservesDirectories(t *testing.T) {
	primary := filepath.Join(t.TempDir(), "app's $HOME; project")
	if err := os.Mkdir(primary, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, extras := range [][]string{nil, {"/tmp/lib one", "/tmp/lib's $(echo unsafe); `echo unsafe`"}} {
		// A shell function captures the real working directory and arguments without launching Codex.
		script := "codex() { pwd -P; printf '%s\\n' \"$@\"; }; " + buildShellCommand(primary, extras, "")
		output, err := exec.Command("/bin/sh", "-c", script).CombinedOutput()
		if err != nil {
			t.Fatalf("shell failed: %v: %s", err, output)
		}
		realPrimary, err := filepath.EvalSymlinks(primary)
		if err != nil {
			t.Fatal(err)
		}
		want := realPrimary + "\n"
		if len(extras) == 0 {
			want += "\n"
		} else {
			want += "--sandbox\nworkspace-write\n"
		}
		for _, extra := range extras {
			want += "--add-dir\n" + extra + "\n"
		}
		if string(output) != want {
			t.Fatalf("got %q, want %q", output, want)
		}
	}
}

func TestOpenMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := Open("/tmp", nil, "", ""); err == nil || !strings.Contains(err.Error(), "codex command not found") {
		t.Fatalf("expected missing executable error, got %v", err)
	}
}

func TestTokenCommandScopesKeyAndStopsOnLookupFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			dir := t.TempDir()
			helper := filepath.Join(dir, "token helper")
			body := "#!/bin/sh\nprintf '%s' 'test-key'\n"
			if fail {
				body = "#!/bin/sh\nexit 1\n"
			}
			if err := os.WriteFile(helper, []byte(body), 0o700); err != nil {
				t.Fatal(err)
			}
			command := buildShellCommand(dir, []string{"/tmp/extra dir"}, "work's token")
			if strings.Contains(command, "test-key") {
				t.Fatal("raw key in command")
			}
			command = strings.ReplaceAll(command, shellQuote(curspaceExecutable()), shellQuote(helper))
			script := "CURSPACE_CODEX_API_KEY=parent; codex() { printf '%s\\n' \"$CURSPACE_CODEX_API_KEY\" \"$@\"; }; " + command + "; printf 'parent=%s' \"$CURSPACE_CODEX_API_KEY\""
			out, err := exec.Command("/bin/sh", "-c", script).CombinedOutput()
			if err != nil {
				t.Fatalf("shell failed: %v", err)
			}
			if fail {
				if string(out) != "parent=parent" {
					t.Fatalf("Codex ran after token failure: %s", out)
				}
			} else {
				for _, want := range []string{"test-key\n", "--sandbox\nworkspace-write\n", `model_provider="curspace_token"`, `env_key="CURSPACE_CODEX_API_KEY"`, "--add-dir\n/tmp/extra dir", "parent=parent"} {
					if !strings.Contains(string(out), want) {
						t.Fatalf("missing %q in %s", want, out)
					}
				}
			}
		})
	}
}

func TestBuildShellCommandWithLiteLLMModel(t *testing.T) {
	got := buildShellCommand("/projects/app", nil, "", "gpt-5", "https://llm.example/v1")
	for _, want := range []string{"codex -m 'gpt-5'", `model_provider="curspace_litellm"`, `base_url="https://llm.example/v1"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("command missing %q: %s", want, got)
		}
	}
}
