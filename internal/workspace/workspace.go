package workspace

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	curspaceDir  = ".curspace"
	workspaceDir = "workspaces"
	extension    = ".code-workspace"
)

type WorkspaceFolder struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

type WorkspaceFile struct {
	Folders  []WorkspaceFolder `json:"folders"`
	Settings map[string]any    `json:"settings"`
}

func workspaceDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, curspaceDir, workspaceDir), nil
}

func workspaceFilePath(name string) (string, error) {
	dir, err := workspaceDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+extension), nil
}

func defaultSettings() map[string]any {
	return map[string]any{
		"files.autoSave": "afterDelay",
	}
}

func Create(name string, folders []WorkspaceFolder) (string, error) {
	dir, err := workspaceDirPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating workspace directory: %w", err)
	}

	wsFile := WorkspaceFile{
		Folders:  folders,
		Settings: defaultSettings(),
	}

	data, err := json.MarshalIndent(wsFile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling workspace: %w", err)
	}

	path, err := workspaceFilePath(name)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("writing workspace file: %w", err)
	}

	return path, nil
}

func List() ([]string, error) {
	dir, err := workspaceDirPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if before, ok := strings.CutSuffix(name, extension); ok {
			names = append(names, before)
		}
	}

	return names, nil
}

func Open(name string) (string, error) {
	path, err := workspaceFilePath(name)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("workspace not found: %s", name)
	}

	return path, nil
}

func Delete(name string) error {
	path, err := workspaceFilePath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found: %s", name)
	}

	return os.Remove(path)
}

type WorkspaceInfo struct {
	Name        string
	FolderCount int
	ModTime     time.Time
}

func ListDetailed() ([]WorkspaceInfo, error) {
	dir, err := workspaceDirPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}

	var infos []WorkspaceInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), extension) {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), extension)

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		folderCount := 0
		if err == nil {
			var ws WorkspaceFile
			if json.Unmarshal(data, &ws) == nil {
				folderCount = len(ws.Folders)
			}
		}

		infos = append(infos, WorkspaceInfo{
			Name:        name,
			FolderCount: folderCount,
			ModTime:     fileInfo.ModTime(),
		})
	}

	slices.SortFunc(infos, func(a, b WorkspaceInfo) int {
		return cmp.Compare(b.ModTime.UnixNano(), a.ModTime.UnixNano())
	})

	return infos, nil
}

func Rename(oldName, newName string) error {
	oldPath, err := workspaceFilePath(oldName)
	if err != nil {
		return err
	}

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("workspace not found: %s", oldName)
	}

	newPath, err := workspaceFilePath(newName)
	if err != nil {
		return err
	}

	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("workspace already exists: %s", newName)
	}

	return os.Rename(oldPath, newPath)
}
