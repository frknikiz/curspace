package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	configDir  = ".curspace"
	configFile = "config.json"
)

type Config struct {
	Roots    []string `json:"roots"`
	MaxDepth int      `json:"max_depth"`
	// Terminal selects the macOS/Linux terminal app used to launch Claude Code.
	// Leave empty for auto-detect. Supported values: "iterm", "terminal" (macOS),
	// or any executable name on Linux (overrides $TERMINAL).
	Terminal string `json:"terminal,omitempty"`
	// DefaultEditor skips the editor picker when opening workspaces or projects.
	// Allowed: "" (always ask), "cursor", "claude".
	DefaultEditor string `json:"default_editor,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Roots:    []string{},
		MaxDepth: 10,
	}
}

func configDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

func configFilePath() (string, error) {
	dir, err := configDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFile), nil
}

func Load() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if saveErr := Save(cfg); saveErr != nil {
				return nil, fmt.Errorf("creating default config: %w", saveErr)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.MaxDepth == 0 {
		cfg.MaxDepth = 10
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	dir, err := configDirPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	path, err := configFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("saving config: %w", err)
	}

	return nil
}

func NormalizePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}

	return filepath.Clean(abs), nil
}

func AddRoot(path string) error {
	normalized, err := NormalizePath(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(normalized)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", normalized)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", normalized)
	}

	cfg, err := Load()
	if err != nil {
		return err
	}

	if slices.Contains(cfg.Roots, normalized) {
		return fmt.Errorf("root already exists: %s", normalized)
	}

	cfg.Roots = append(cfg.Roots, normalized)
	return Save(cfg)
}

func RemoveRoot(path string) error {
	normalized, err := NormalizePath(path)
	if err != nil {
		return err
	}

	cfg, err := Load()
	if err != nil {
		return err
	}

	if !slices.Contains(cfg.Roots, normalized) {
		return fmt.Errorf("root not found: %s", normalized)
	}

	cfg.Roots = slices.DeleteFunc(cfg.Roots, func(root string) bool {
		return root == normalized
	})
	return Save(cfg)
}
