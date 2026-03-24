package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/cursor"
	"github.com/frknikiz/curspace/internal/discovery"
	"github.com/frknikiz/curspace/internal/scanner"
	"github.com/frknikiz/curspace/internal/ui"
	"github.com/frknikiz/curspace/internal/workspace"
	"github.com/spf13/cobra"
)

var refreshFlag bool

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD43B"))
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Select projects and open as workspace in Cursor",
	Long:  "Scans for projects (or uses cache), presents a TUI selector, creates a workspace, and opens it in Cursor.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Roots) == 0 {
			fmt.Println(warnStyle.Render("No roots configured.") + " Add one with: " + infoStyle.Render("curspace roots add <path>"))
			return nil
		}

		var projects []scanner.Project

		if !refreshFlag {
			result, err := discovery.Discover(context.Background(), discovery.Options{
				Roots:    cfg.Roots,
				MaxDepth: cfg.MaxDepth,
			})
			if err == nil && result.Source == discovery.SourceCache {
				projects = result.Projects
				fmt.Printf("%s %s\n", infoStyle.Render("▸"), "Using cached scan results")
			}
		}

		if projects == nil {
			scanErr := ui.RunWithSpinner("Scanning for projects...", func() error {
				result, scanInnerErr := discovery.Discover(context.Background(), discovery.Options{
					Roots:        cfg.Roots,
					MaxDepth:     cfg.MaxDepth,
					ForceRefresh: refreshFlag,
				})
				projects = result.Projects
				return scanInnerErr
			})
			if scanErr != nil {
				return fmt.Errorf("scanning: %w", scanErr)
			}
		}

		if len(projects) == 0 {
			fmt.Println(warnStyle.Render("No projects found."))
			return nil
		}

		fmt.Printf("%s Found %d projects\n\n", successStyle.Render("✓"), len(projects))

		selected, err := ui.RunSelector(projects)
		if err != nil {
			return fmt.Errorf("project selection: %w", err)
		}
		if len(selected) == 0 {
			fmt.Println("No projects selected.")
			return nil
		}

		fmt.Printf("\n%s Selected %d project(s)\n\n", successStyle.Render("✓"), len(selected))

		wsName, err := ui.RunPrompt("my-workspace")
		if err != nil {
			return fmt.Errorf("workspace name input: %w", err)
		}
		if wsName == "" {
			fmt.Println("Cancelled.")
			return nil
		}

		folders := make([]workspace.WorkspaceFolder, len(selected))
		for i, p := range selected {
			folders[i] = workspace.WorkspaceFolder{
				Name: p.Name,
				Path: p.Path,
			}
		}

		wsPath, err := workspace.Create(wsName, folders)
		if err != nil {
			return fmt.Errorf("creating workspace: %w", err)
		}

		fmt.Printf("\n%s Created workspace: %s\n", successStyle.Render("✓"), infoStyle.Render(wsPath))

		if err := cursor.Open(wsPath); err != nil {
			return fmt.Errorf("opening in Cursor: %w", err)
		}

		fmt.Printf("%s Opened in Cursor!\n", successStyle.Render("✓"))
		return nil
	},
}

func init() {
	openCmd.Flags().BoolVar(&refreshFlag, "refresh", false, "Bypass cache and rescan projects")
	rootCmd.AddCommand(openCmd)
}
