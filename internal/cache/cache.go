package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/frknikiz/curspace/internal/scanner"
)

const (
	cacheDir  = ".curspace"
	cacheFile = "cache.json"
)

type CacheData struct {
	Projects  []scanner.Project `json:"projects"`
	Timestamp time.Time         `json:"timestamp"`
	RootsHash string            `json:"roots_hash"`
}

func cacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, cacheDir, cacheFile), nil
}

func Load() (*CacheData, error) {
	path, err := cacheFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading cache: %w", err)
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("parsing cache: %w", err)
	}

	return &cache, nil
}

func Save(data *CacheData) error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0o644); err != nil {
		return fmt.Errorf("writing cache: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("saving cache: %w", err)
	}

	return nil
}

func IsValid(cached *CacheData, currentRoots []string) bool {
	if cached == nil {
		return false
	}
	return cached.RootsHash == ComputeRootsHash(currentRoots)
}

func ComputeRootsHash(roots []string) string {
	sorted := slices.Clone(roots)
	slices.Sort(sorted)

	joined := strings.Join(sorted, "\n")
	hash := sha256.Sum256([]byte(joined))
	return fmt.Sprintf("%x", hash)
}

func Clear() error {
	path, err := cacheFilePath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clearing cache: %w", err)
	}

	return nil
}
