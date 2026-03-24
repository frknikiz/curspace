package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func mustMkdirConfigDir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", path, err)
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{
		Roots:    []string{"/projects/a", "/projects/b"},
		MaxDepth: 5,
	}

	// When
	err := Save(cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Then
	if loaded.MaxDepth != cfg.MaxDepth {
		t.Errorf("MaxDepth: got %d, want %d", loaded.MaxDepth, cfg.MaxDepth)
	}
	if len(loaded.Roots) != len(cfg.Roots) {
		t.Fatalf("Roots length: got %d, want %d", len(loaded.Roots), len(cfg.Roots))
	}
	for i, r := range loaded.Roots {
		if r != cfg.Roots[i] {
			t.Errorf("Roots[%d]: got %q, want %q", i, r, cfg.Roots[i])
		}
	}
}

func TestLoadCreatesDefaultConfig(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// When
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Then
	if cfg.MaxDepth != 10 {
		t.Errorf("default MaxDepth: got %d, want 10", cfg.MaxDepth)
	}
	if len(cfg.Roots) != 0 {
		t.Errorf("default Roots: got %d, want 0", len(cfg.Roots))
	}

	configPath := filepath.Join(tmpDir, configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"relative path", "."},
		{"path with trailing slash", "/tmp/foo/"},
		{"tilde path", "~/projects"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			input := tt.input

			// When
			result, err := NormalizePath(input)

			// Then
			if err != nil {
				t.Fatalf("NormalizePath(%q) error: %v", input, err)
			}
			if !filepath.IsAbs(result) {
				t.Errorf("NormalizePath(%q) = %q, want absolute path", input, result)
			}
			if result != filepath.Clean(result) {
				t.Errorf("NormalizePath(%q) = %q, not clean", input, result)
			}
		})
	}
}

func TestAddRootDeduplication(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	rootDir := filepath.Join(tmpDir, "projects")
	mustMkdirConfigDir(t, rootDir)

	// When
	err := AddRoot(rootDir)
	if err != nil {
		t.Fatalf("first AddRoot failed: %v", err)
	}

	err = AddRoot(rootDir)

	// Then
	if err == nil {
		t.Error("expected error on duplicate AddRoot, got nil")
	}

	cfg, _ := Load()
	count := 0
	for _, r := range cfg.Roots {
		if r == rootDir {
			count++
		}
	}
	if count != 1 {
		t.Errorf("root appeared %d times, want 1", count)
	}
}

func TestRemoveRoot(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	rootDir := filepath.Join(tmpDir, "projects")
	mustMkdirConfigDir(t, rootDir)
	if err := AddRoot(rootDir); err != nil {
		t.Fatalf("AddRoot failed: %v", err)
	}

	// When
	err := RemoveRoot(rootDir)

	// Then
	if err != nil {
		t.Fatalf("RemoveRoot failed: %v", err)
	}

	cfg, _ := Load()
	for _, r := range cfg.Roots {
		if r == rootDir {
			t.Error("root was not removed")
		}
	}
}

func TestRemoveRootNotFound(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// When
	err := RemoveRoot("/nonexistent/path")

	// Then
	if err == nil {
		t.Error("expected error for non-existent root, got nil")
	}
}

func TestSaveAtomicity(t *testing.T) {
	// Given
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cfg := &Config{Roots: []string{"/a"}, MaxDepth: 3}
	if err := Save(cfg); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}

	// When
	cfg.Roots = append(cfg.Roots, "/b")
	err := Save(cfg)

	// Then
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, configDir, configFile))
	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(loaded.Roots) != 2 {
		t.Errorf("expected 2 roots after save, got %d", len(loaded.Roots))
	}
}
