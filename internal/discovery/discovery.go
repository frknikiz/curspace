package discovery

import (
	"context"
	"time"

	"github.com/frknikiz/curspace/internal/cache"
	"github.com/frknikiz/curspace/internal/scanner"
)

type Source string

const (
	SourceFresh Source = "fresh"
	SourceCache Source = "cache"
)

type Options struct {
	Roots        []string
	MaxDepth     int
	ForceRefresh bool
}

type Result struct {
	Projects  []scanner.Project
	Source    Source
	Timestamp time.Time
}

type dependencies struct {
	loadCache func() (*cache.CacheData, error)
	saveCache func(*cache.CacheData) error
	scan      func(context.Context, scanner.ScanOptions) ([]scanner.Project, error)
	now       func() time.Time
}

func Discover(ctx context.Context, opts Options) (Result, error) {
	return discover(ctx, opts, dependencies{
		loadCache: cache.Load,
		saveCache: cache.Save,
		scan:      scanner.Scan,
		now:       time.Now,
	})
}

func discover(ctx context.Context, opts Options, deps dependencies) (Result, error) {
	if !opts.ForceRefresh {
		cached, err := deps.loadCache()
		if err == nil && cache.IsValid(cached, opts.Roots) {
			return Result{
				Projects:  cached.Projects,
				Source:    SourceCache,
				Timestamp: cached.Timestamp,
			}, nil
		}
	}

	projects, err := deps.scan(ctx, scanner.ScanOptions{
		Roots:    opts.Roots,
		MaxDepth: opts.MaxDepth,
	})
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Projects:  projects,
		Source:    SourceFresh,
		Timestamp: deps.now(),
	}

	_ = deps.saveCache(&cache.CacheData{
		Projects:  projects,
		Timestamp: result.Timestamp,
		RootsHash: cache.ComputeRootsHash(opts.Roots),
	})

	return result, nil
}
