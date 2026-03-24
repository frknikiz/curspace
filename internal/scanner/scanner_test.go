package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", path, err)
	}
}

func TestDetectGoProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "mygoproject")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "go.mod"), []byte("module test"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Go {
		t.Errorf("expected Go type, got %s", projects[0].Type)
	}
	if projects[0].Name != "mygoproject" {
		t.Errorf("expected name 'mygoproject', got %s", projects[0].Name)
	}
}

func TestDetectNodeProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "myapp")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "package.json"), []byte("{}"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Node {
		t.Errorf("expected Node type, got %s", projects[0].Type)
	}
}

func TestDetectJavaProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "javaapp")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "pom.xml"), []byte("<project/>"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Java {
		t.Errorf("expected Java type, got %s", projects[0].Type)
	}
}

func TestDetectRustProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "rustapp")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "Cargo.toml"), []byte("[package]"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Rust {
		t.Errorf("expected Rust type, got %s", projects[0].Type)
	}
}

func TestDetectPythonProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "pyapp")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "requirements.txt"), []byte("flask"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Python {
		t.Errorf("expected Python type, got %s", projects[0].Type)
	}
}

func TestDetectDotNetProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "dotnetapp")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "MyApp.csproj"), []byte("<Project/>"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != DotNet {
		t.Errorf("expected .NET type, got %s", projects[0].Type)
	}
}

func TestDetectPHPProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "phpapp")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "composer.json"), []byte("{}"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != PHP {
		t.Errorf("expected PHP type, got %s", projects[0].Type)
	}
}

func TestIgnoreDirectories(t *testing.T) {
	// Given
	dir := t.TempDir()
	nodeModules := filepath.Join(dir, "node_modules", "somelib")
	mustMkdirAll(t, nodeModules)
	mustWriteFile(t, filepath.Join(nodeModules, "package.json"), []byte("{}"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects (ignored dirs), got %d", len(projects))
	}
}

func TestDetectGitOnlyProject(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "project-mfm")
	mustMkdirAll(t, filepath.Join(projectDir, ".git"))
	mustWriteFile(t, filepath.Join(projectDir, "config.yaml"), []byte("key: val"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Git {
		t.Errorf("expected Git type, got %s", projects[0].Type)
	}
	if projects[0].Name != "project-mfm" {
		t.Errorf("expected name 'project-mfm', got %s", projects[0].Name)
	}
}

func TestGitProjectUpgradedByMarker(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "myservice")
	mustMkdirAll(t, filepath.Join(projectDir, ".git"))
	mustWriteFile(t, filepath.Join(projectDir, "go.mod"), []byte("module svc"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Type != Go {
		t.Errorf("expected Go type (upgraded from Git), got %s", projects[0].Type)
	}
}

func TestGitRepoAtRootNotDetected(t *testing.T) {
	// Given — .git at the scan root itself should not be listed
	dir := t.TempDir()
	mustMkdirAll(t, filepath.Join(dir, ".git"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects (root .git ignored), got %d", len(projects))
	}
}

func TestMultipleGitProjectsUnderRoot(t *testing.T) {
	// Given
	dir := t.TempDir()
	for _, name := range []string{"svc-a", "svc-b", "infra-config"} {
		mustMkdirAll(t, filepath.Join(dir, name, ".git"))
	}
	// svc-a also has go.mod
	mustWriteFile(t, filepath.Join(dir, "svc-a", "go.mod"), []byte("module a"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(projects))
	}

	types := make(map[string]ProjectType)
	for _, p := range projects {
		types[p.Name] = p.Type
	}
	if types["svc-a"] != Go {
		t.Errorf("svc-a: expected Go, got %s", types["svc-a"])
	}
	if types["svc-b"] != Git {
		t.Errorf("svc-b: expected Git, got %s", types["svc-b"])
	}
	if types["infra-config"] != Git {
		t.Errorf("infra-config: expected Git, got %s", types["infra-config"])
	}
}

func TestMaxDepth(t *testing.T) {
	// Given
	dir := t.TempDir()
	deepDir := filepath.Join(dir, "a", "b", "c", "d", "e")
	mustMkdirAll(t, deepDir)
	mustWriteFile(t, filepath.Join(deepDir, "go.mod"), []byte("module deep"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 2})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects (beyond max depth), got %d", len(projects))
	}
}

func TestMaxDepthAllowsShallow(t *testing.T) {
	// Given
	dir := t.TempDir()
	shallowDir := filepath.Join(dir, "a", "b")
	mustMkdirAll(t, shallowDir)
	mustWriteFile(t, filepath.Join(shallowDir, "go.mod"), []byte("module shallow"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 3})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project within depth, got %d", len(projects))
	}
}

func TestDeduplication(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "multi")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "go.mod"), []byte("module m"))
	mustWriteFile(t, filepath.Join(projectDir, "package.json"), []byte("{}"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 deduplicated project, got %d", len(projects))
	}
}

func TestMultipleRootsDeduplication(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "proj")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "go.mod"), []byte("module m"))

	// When
	projects, err := Scan(context.Background(), ScanOptions{
		Roots:    []string{dir, dir},
		MaxDepth: 10,
	})

	// Then
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project across duplicate roots, got %d", len(projects))
	}
}

func TestContextCancellation(t *testing.T) {
	// Given
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "proj")
	mustMkdirAll(t, projectDir)
	mustWriteFile(t, filepath.Join(projectDir, "go.mod"), []byte("module m"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When
	_, err := Scan(ctx, ScanOptions{Roots: []string{dir}, MaxDepth: 10})

	// Then
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
}
