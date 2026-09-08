package cmd

import (
	"github.com/frknikiz/curspace/internal/config"
	"os"
	"strings"
	"testing"
)

func TestValidateEditor(t *testing.T) {
	for _, editor := range []string{"cursor", "claude", "codex"} {
		if err := validateEditor(editor); err != nil {
			t.Errorf("%s: %v", editor, err)
		}
	}
	if err := validateEditor("unknown"); err == nil {
		t.Fatal("accepted unknown editor")
	}
}

func TestLaunchCodexWithoutFolders(t *testing.T) {
	if err := launchEditor(editorCodex, nil, ""); err == nil || !strings.Contains(err.Error(), "no folders") {
		t.Fatalf("expected empty workspace error, got %v", err)
	}
}

func TestChooseCodexTokenName(t *testing.T) {
	tokens := []config.CodexToken{{Name: "work", Value: "secret"}, {Name: "personal", Value: "secret-2"}}
	for _, tc := range []struct {
		input, want string
		wantErr     bool
	}{
		{"\n", "work", false}, {"2\n", "personal", false}, {"0\n", "", false},
		{"3\n", "", true}, {"bad\n", "", true}, {"", "", true},
	} {
		f, err := os.CreateTemp(t.TempDir(), "input")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
		if _, err = f.WriteString(tc.input); err != nil {
			t.Fatal(err)
		}
		if _, err = f.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		original := os.Stdin
		os.Stdin = f
		got, err := chooseCodexTokenName(tokens)
		os.Stdin = original
		if got != tc.want || (err != nil) != tc.wantErr {
			t.Fatalf("input %q: got %q, %v", tc.input, got, err)
		}
	}
}
