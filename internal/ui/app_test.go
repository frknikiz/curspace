package ui

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frknikiz/curspace/internal/config"
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

func TestClaudeTokenPickPassesSelectedToken(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if err := config.Save(&config.Config{
		Roots:        []string{},
		MaxDepth:     10,
		ClaudeTokens: []config.ClaudeToken{{Name: "work", Value: "sk-ant-work"}},
	}); err != nil {
		t.Fatalf("Save config failed: %v", err)
	}

	var gotToken string
	m := NewAppModel(AppConfig{
		OpenClaude: func(primaryPath string, extraPaths []string, tokenName string) error {
			gotToken = tokenName
			return nil
		},
	})
	m.editorPick = editorPick{
		label:       "svc",
		primaryPath: "/projects/svc",
		cursor:      1,
	}

	model, _ := m.runEditorPick()
	picking := model.(AppModel)
	if picking.view != viewClaudeTokenPick {
		t.Fatalf("expected Claude token picker, got %v", picking.view)
	}

	model, _ = picking.updateClaudeTokenPick(tea.KeyMsg{Type: tea.KeyEnter})
	done := model.(AppModel)
	if gotToken != "work" {
		t.Fatalf("selected token: got %q, want work", gotToken)
	}
	if done.statusErr {
		t.Fatalf("expected successful status, got %q", done.statusMsg)
	}
}

func TestSettingsCanOpenClaudeTokenManager(t *testing.T) {
	m := NewAppModel(AppConfig{
		ClaudeTokens: []config.ClaudeToken{{Name: "work", Value: "sk-ant-work"}},
	})
	m.view = viewSettings
	m.settingsCursor = 2

	model, _ := m.updateSettings(tea.KeyMsg{Type: tea.KeyEnter})
	got := model.(AppModel)

	if got.view != viewClaudeTokens {
		t.Fatalf("expected Claude token manager, got %v", got.view)
	}
	if !strings.Contains(got.renderClaudeTokens(), "work") {
		t.Fatal("expected token manager to render saved token name")
	}
}

func TestClaudeTokenManagerSavesTokenFromInputs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	m := NewAppModel(AppConfig{})
	m.view = viewClaudeTokenName
	m.tokenNameInput = newStyledInput("work")
	m.tokenNameInput.SetValue("work")

	model, _ := m.updateClaudeTokenName(tea.KeyMsg{Type: tea.KeyEnter})
	valueModel := model.(AppModel)
	if valueModel.view != viewClaudeTokenValue {
		t.Fatalf("expected token value input, got %v", valueModel.view)
	}

	valueModel.tokenValueInput.SetValue("sk-ant-work")
	model, _ = valueModel.updateClaudeTokenValue(tea.KeyMsg{Type: tea.KeyEnter})
	done := model.(AppModel)

	if done.view != viewClaudeTokens {
		t.Fatalf("expected token manager after save, got %v", done.view)
	}
	if len(done.claudeTokens) != 1 || done.claudeTokens[0].Name != "work" {
		t.Fatalf("expected saved token in model, got %#v", done.claudeTokens)
	}

	value, err := config.ClaudeTokenValue("work")
	if err != nil {
		t.Fatalf("ClaudeTokenValue failed: %v", err)
	}
	if value != "sk-ant-work" {
		t.Fatalf("saved token value: got %q, want sk-ant-work", value)
	}
}
