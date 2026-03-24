package workspace

import (
	"encoding/json"
	"os"
	"testing"
)

func mustCreateWorkspace(t *testing.T, name string, folders []WorkspaceFolder) string {
	t.Helper()

	path, err := Create(name, folders)
	if err != nil {
		t.Fatalf("Create(%q) failed: %v", name, err)
	}

	return path
}

func TestCreateWritesValidJSON(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	folders := []WorkspaceFolder{
		{Name: "project-a", Path: "/projects/a"},
		{Name: "project-b", Path: "/projects/b"},
	}

	// When
	path, err := Create("test-workspace", folders)

	// Then
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading workspace file: %v", err)
	}

	var ws WorkspaceFile
	if err := json.Unmarshal(data, &ws); err != nil {
		t.Fatalf("invalid JSON in workspace file: %v", err)
	}

	if len(ws.Folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(ws.Folders))
	}

	if ws.Settings["files.autoSave"] != "afterDelay" {
		t.Error("expected default settings in workspace file")
	}
}

func TestListReturnsCreatedWorkspaces(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mustCreateWorkspace(t, "ws-alpha", []WorkspaceFolder{{Path: "/a"}})
	mustCreateWorkspace(t, "ws-beta", []WorkspaceFolder{{Path: "/b"}})

	// When
	names, err := List()

	// Then
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(names))
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["ws-alpha"] || !found["ws-beta"] {
		t.Errorf("expected ws-alpha and ws-beta, got %v", names)
	}
}

func TestListEmptyDirectory(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// When
	names, err := List()

	// Then
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(names))
	}
}

func TestDeleteRemovesFile(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mustCreateWorkspace(t, "to-delete", []WorkspaceFolder{{Path: "/a"}})

	// When
	err := Delete("to-delete")

	// Then
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	names, _ := List()
	for _, n := range names {
		if n == "to-delete" {
			t.Error("workspace was not deleted")
		}
	}
}

func TestDeleteNotFound(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// When
	err := Delete("nonexistent")

	// Then
	if err == nil {
		t.Error("expected error deleting nonexistent workspace")
	}
}

func TestRename(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mustCreateWorkspace(t, "old-name", []WorkspaceFolder{{Path: "/a"}})

	// When
	err := Rename("old-name", "new-name")

	// Then
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	names, _ := List()
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if found["old-name"] {
		t.Error("old workspace name still exists")
	}
	if !found["new-name"] {
		t.Error("new workspace name not found")
	}
}

func TestRenameConflict(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mustCreateWorkspace(t, "ws-a", []WorkspaceFolder{{Path: "/a"}})
	mustCreateWorkspace(t, "ws-b", []WorkspaceFolder{{Path: "/b"}})

	// When
	err := Rename("ws-a", "ws-b")

	// Then
	if err == nil {
		t.Error("expected error renaming to existing workspace")
	}
}

func TestOpen(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	mustCreateWorkspace(t, "my-ws", []WorkspaceFolder{{Path: "/a"}})

	// When
	path, err := Open("my-ws")

	// Then
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Error("workspace file does not exist at returned path")
	}
}

func TestOpenNotFound(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// When
	_, err := Open("nonexistent")

	// Then
	if err == nil {
		t.Error("expected error opening nonexistent workspace")
	}
}
