package ui

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frknikiz/curspace/internal/discovery"
	"github.com/frknikiz/curspace/internal/scanner"
)

func TestStartSelectorRescanPreservesState(t *testing.T) {
	m := NewAppModel(AppConfig{
		Roots:    []string{"/projects"},
		MaxDepth: 4,
	})
	m.view = viewSelector
	m.projects = []scanner.Project{
		{Name: "svc-a", Path: "/projects/svc-a", Type: scanner.Go},
		{Name: "svc-b", Path: "/projects/svc-b", Type: scanner.Node},
	}
	m.initSelector()
	m.search = "svc-b"
	m.selected[1] = true

	cmd := m.startSelectorRescan()
	if cmd == nil {
		t.Fatal("expected rescan command")
	}
	if m.view != viewScanning {
		t.Fatalf("expected scanning view, got %v", m.view)
	}
	if !m.scan.forceRefresh {
		t.Fatal("expected rescan to force refresh")
	}
	if m.scan.returnView != viewSelector {
		t.Fatalf("expected return view selector, got %v", m.scan.returnView)
	}
	if m.scan.preserveSearch != "svc-b" {
		t.Fatalf("expected preserved search, got %q", m.scan.preserveSearch)
	}
	if !m.scan.preserveSelectedPaths["/projects/svc-b"] {
		t.Fatalf("expected selected path to be preserved: %#v", m.scan.preserveSelectedPaths)
	}
	if m.activeScanID == 0 {
		t.Fatal("expected active scan id to be set")
	}
}

func TestScanDoneRestoresSelectionAndFilter(t *testing.T) {
	now := time.Date(2026, time.March, 24, 12, 0, 0, 0, time.UTC)
	m := NewAppModel(AppConfig{})
	m.scan = scanIntent{
		returnView:            viewSelector,
		forceRefresh:          true,
		preserveSearch:        "svc-b",
		preserveSelectedPaths: map[string]bool{"/projects/svc-b": true},
	}
	m.activeScanID = 7

	model, _ := m.Update(scanDoneMsg{
		scanID: 7,
		result: discovery.Result{
			Projects: []scanner.Project{
				{Name: "svc-a", Path: "/projects/svc-a", Type: scanner.Go},
				{Name: "svc-b", Path: "/projects/svc-b", Type: scanner.Node},
			},
			Source:    discovery.SourceFresh,
			Timestamp: now,
		},
	})
	got := model.(AppModel)

	if got.view != viewSelector {
		t.Fatalf("expected selector view, got %v", got.view)
	}
	if got.search != "svc-b" {
		t.Fatalf("expected restored search, got %q", got.search)
	}
	if len(got.filtered) != 1 {
		t.Fatalf("expected one filtered project, got %d", len(got.filtered))
	}
	if !got.selected[1] {
		t.Fatalf("expected project 1 to remain selected: %#v", got.selected)
	}
	if got.lastScanSource != discovery.SourceFresh {
		t.Fatalf("expected fresh source, got %q", got.lastScanSource)
	}
	if !got.lastScanAt.Equal(now) {
		t.Fatalf("expected scan timestamp %v, got %v", now, got.lastScanAt)
	}
	if got.activeScanID != 0 {
		t.Fatalf("expected scan to be cleared, got active id %d", got.activeScanID)
	}
}

func TestCancelledScanResultIsIgnored(t *testing.T) {
	m := NewAppModel(AppConfig{})
	m.view = viewScanning
	m.scan = scanIntent{returnView: viewMain, forceRefresh: true}
	m.activeScanID = 3

	model, _ := m.updateScanning(tea.KeyMsg{Type: tea.KeyEsc})
	cancelled := model.(AppModel)

	if cancelled.view != viewMain {
		t.Fatalf("expected main view after cancel, got %v", cancelled.view)
	}
	if cancelled.activeScanID != 0 {
		t.Fatalf("expected active scan id cleared, got %d", cancelled.activeScanID)
	}

	model, _ = cancelled.Update(scanDoneMsg{
		scanID: 3,
		result: discovery.Result{
			Projects: []scanner.Project{{Name: "svc", Path: "/projects/svc", Type: scanner.Go}},
		},
	})
	ignored := model.(AppModel)

	if ignored.view != viewMain {
		t.Fatalf("expected stale scan result to keep main view, got %v", ignored.view)
	}
	if len(ignored.projects) != 0 {
		t.Fatalf("expected stale scan result to be ignored, got %#v", ignored.projects)
	}
}

func TestPathSuggestionsOnlyReturnDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	dirMatch := filepath.Join(tmpDir, "projects")
	fileMatch := filepath.Join(tmpDir, "project.txt")
	otherDir := filepath.Join(tmpDir, "workspace")

	if err := os.MkdirAll(dirMatch, 0o755); err != nil {
		t.Fatalf("mkdir projects: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	if err := os.WriteFile(fileMatch, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got := pathSuggestions(filepath.Join(tmpDir, "pro"))
	want := []string{dirMatch + string(filepath.Separator)}

	if !slices.Equal(got, want) {
		t.Fatalf("pathSuggestions mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestPathSuggestionsPreserveTildePaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	projectsDir := filepath.Join(tmpDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("mkdir projects: %v", err)
	}

	got := pathSuggestions("~/pro")
	want := []string{"~/projects" + string(filepath.Separator)}

	if !slices.Equal(got, want) {
		t.Fatalf("pathSuggestions mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestSyncRootSuggestionsUsesCurrentInput(t *testing.T) {
	tmpDir := t.TempDir()
	projectsDir := filepath.Join(tmpDir, "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatalf("mkdir projects: %v", err)
	}

	m := NewAppModel(AppConfig{})
	m.rootInput = newPathInput("~/projects")
	m.rootInput.SetValue(filepath.Join(tmpDir, "pro"))
	m.syncRootSuggestions()

	got := m.rootInput.AvailableSuggestions()
	want := []string{projectsDir + string(filepath.Separator)}

	if !slices.Equal(got, want) {
		t.Fatalf("available suggestions mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}
