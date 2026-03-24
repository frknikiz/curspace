package discovery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/frknikiz/curspace/internal/cache"
	"github.com/frknikiz/curspace/internal/scanner"
)

func TestDiscoverReturnsValidCache(t *testing.T) {
	roots := []string{"/projects"}
	cachedProjects := []scanner.Project{{Name: "cached", Path: "/projects/cached", Type: scanner.Go}}
	cachedAt := time.Date(2026, time.March, 24, 10, 0, 0, 0, time.UTC)

	scanCalled := false
	result, err := discover(context.Background(), Options{
		Roots:    roots,
		MaxDepth: 4,
	}, dependencies{
		loadCache: func() (*cache.CacheData, error) {
			return &cache.CacheData{
				Projects:  cachedProjects,
				Timestamp: cachedAt,
				RootsHash: cache.ComputeRootsHash(roots),
			}, nil
		},
		saveCache: func(*cache.CacheData) error { return nil },
		scan: func(context.Context, scanner.ScanOptions) ([]scanner.Project, error) {
			scanCalled = true
			return nil, nil
		},
		now: time.Now,
	})
	if err != nil {
		t.Fatalf("discover returned error: %v", err)
	}

	if scanCalled {
		t.Fatal("expected discover to use cache without scanning")
	}
	if result.Source != SourceCache {
		t.Fatalf("expected cache source, got %q", result.Source)
	}
	if !result.Timestamp.Equal(cachedAt) {
		t.Fatalf("expected cached timestamp %v, got %v", cachedAt, result.Timestamp)
	}
	if len(result.Projects) != 1 || result.Projects[0].Name != "cached" {
		t.Fatalf("unexpected projects returned: %#v", result.Projects)
	}
}

func TestDiscoverForceRefreshScansAndSaves(t *testing.T) {
	roots := []string{"/projects"}
	now := time.Date(2026, time.March, 24, 11, 30, 0, 0, time.UTC)
	freshProjects := []scanner.Project{{Name: "fresh", Path: "/projects/fresh", Type: scanner.Node}}

	scanCalled := false
	saveCalled := false
	result, err := discover(context.Background(), Options{
		Roots:        roots,
		MaxDepth:     5,
		ForceRefresh: true,
	}, dependencies{
		loadCache: func() (*cache.CacheData, error) {
			return &cache.CacheData{
				Projects:  []scanner.Project{{Name: "stale", Path: "/projects/stale", Type: scanner.Git}},
				Timestamp: now.Add(-time.Hour),
				RootsHash: cache.ComputeRootsHash(roots),
			}, nil
		},
		saveCache: func(data *cache.CacheData) error {
			saveCalled = true
			if data.RootsHash != cache.ComputeRootsHash(roots) {
				t.Fatalf("unexpected roots hash saved: %q", data.RootsHash)
			}
			if !data.Timestamp.Equal(now) {
				t.Fatalf("expected saved timestamp %v, got %v", now, data.Timestamp)
			}
			if len(data.Projects) != 1 || data.Projects[0].Name != "fresh" {
				t.Fatalf("unexpected projects saved: %#v", data.Projects)
			}
			return nil
		},
		scan: func(context.Context, scanner.ScanOptions) ([]scanner.Project, error) {
			scanCalled = true
			return freshProjects, nil
		},
		now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("discover returned error: %v", err)
	}

	if !scanCalled {
		t.Fatal("expected force refresh to call scanner")
	}
	if !saveCalled {
		t.Fatal("expected force refresh to save cache")
	}
	if result.Source != SourceFresh {
		t.Fatalf("expected fresh source, got %q", result.Source)
	}
	if !result.Timestamp.Equal(now) {
		t.Fatalf("expected timestamp %v, got %v", now, result.Timestamp)
	}
	if len(result.Projects) != 1 || result.Projects[0].Name != "fresh" {
		t.Fatalf("unexpected projects returned: %#v", result.Projects)
	}
}

func TestDiscoverScanErrorsBubbleUp(t *testing.T) {
	wantErr := errors.New("scan failed")

	_, err := discover(context.Background(), Options{
		Roots:    []string{"/projects"},
		MaxDepth: 3,
	}, dependencies{
		loadCache: func() (*cache.CacheData, error) { return nil, nil },
		saveCache: func(*cache.CacheData) error { return nil },
		scan: func(context.Context, scanner.ScanOptions) ([]scanner.Project, error) {
			return nil, wantErr
		},
		now: time.Now,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
