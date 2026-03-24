package scanner

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

type ProjectType string

const (
	Go     ProjectType = "Go"
	Java   ProjectType = "Java"
	Python ProjectType = "Python"
	Node   ProjectType = "Node"
	Rust   ProjectType = "Rust"
	DotNet ProjectType = ".NET"
	PHP    ProjectType = "PHP"
	Git    ProjectType = "Git"
)

type Project struct {
	Name string      `json:"name"`
	Path string      `json:"path"`
	Type ProjectType `json:"type"`
}

type ScanOptions struct {
	MaxDepth int
	Roots    []string
}

var markerFiles = map[string]ProjectType{
	"go.mod":           Go,
	"package.json":     Node,
	"pom.xml":          Java,
	"build.gradle":     Java,
	"build.gradle.kts": Java,
	"Cargo.toml":       Rust,
	"requirements.txt": Python,
	"setup.py":         Python,
	"pyproject.toml":   Python,
	"Pipfile":          Python,
	"composer.json":    PHP,
}

var dotNetGlobs = []string{"*.csproj", "*.fsproj", "*.sln"}

var ignoreDirs = map[string]bool{
	"node_modules": true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".venv":        true,
	"venv":         true,
	"bin":          true,
	"obj":          true,
	".idea":        true,
	".vscode":      true,
}

func Scan(ctx context.Context, opts ScanOptions) ([]Project, error) {
	seen := make(map[string]bool)
	var projects []Project

	for _, root := range opts.Roots {
		discovered, err := scanRoot(ctx, root, opts.MaxDepth)
		if err != nil {
			return nil, fmt.Errorf("scanning root %s: %w", root, err)
		}
		for _, p := range discovered {
			if !seen[p.Path] {
				seen[p.Path] = true
				projects = append(projects, p)
			}
		}
	}

	return projects, nil
}

func scanRoot(ctx context.Context, root string, maxDepth int) ([]Project, error) {
	var projects []Project
	// path -> index in projects slice, so we can override Git type with specific type
	seenIdx := make(map[string]int)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("skipping unreadable path %s: %v", path, err)
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() {
			if d.Type()&fs.ModeSymlink != 0 {
				return fs.SkipDir
			}

			// .git directory: mark parent as project, then skip contents
			if d.Name() == ".git" {
				parentDir := filepath.Dir(path)
				if parentDir != root {
					if _, exists := seenIdx[parentDir]; !exists {
						seenIdx[parentDir] = len(projects)
						projects = append(projects, Project{
							Name: filepath.Base(parentDir),
							Path: parentDir,
							Type: Git,
						})
					}
				}
				return fs.SkipDir
			}

			if ignoreDirs[d.Name()] {
				return fs.SkipDir
			}

			rel, relErr := filepath.Rel(root, path)
			if relErr == nil && rel != "." {
				depth := len(strings.Split(rel, string(filepath.Separator)))
				if depth > maxDepth {
					return fs.SkipDir
				}
			}

			return nil
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		dir := filepath.Dir(path)
		name := d.Name()

		if projectType, ok := markerFiles[name]; ok {
			addOrUpgrade(&projects, seenIdx, dir, projectType)
			return nil
		}

		for _, glob := range dotNetGlobs {
			if matched, _ := filepath.Match(glob, name); matched {
				addOrUpgrade(&projects, seenIdx, dir, DotNet)
				return nil
			}
		}

		return nil
	})

	return projects, err
}

// addOrUpgrade adds a project or upgrades its type if it was previously detected
// only via .git (generic Git type → specific language type).
func addOrUpgrade(projects *[]Project, seenIdx map[string]int, dir string, projectType ProjectType) {
	if idx, exists := seenIdx[dir]; exists {
		// Override generic Git type with specific language type
		if (*projects)[idx].Type == Git {
			(*projects)[idx].Type = projectType
		}
		return
	}
	seenIdx[dir] = len(*projects)
	*projects = append(*projects, Project{
		Name: filepath.Base(dir),
		Path: dir,
		Type: projectType,
	})
}
