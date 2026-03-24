package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/discovery"
	"github.com/frknikiz/curspace/internal/scanner"
	"github.com/frknikiz/curspace/internal/ui"
	"github.com/spf13/cobra"
)

var (
	scanHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	scanCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	scanNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Bold(true)

	scanPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	scanTypeColors = map[scanner.ProjectType]string{
		scanner.Go:     "#00ADD8",
		scanner.Node:   "#68A063",
		scanner.Java:   "#F89820",
		scanner.Python: "#FFD43B",
		scanner.Rust:   "#DEA584",
		scanner.DotNet: "#512BD4",
		scanner.PHP:    "#777BB4",
		scanner.Git:    "#F05032",
	}
)

func scanTypeTag(pt scanner.ProjectType) string {
	color := "#888888"
	if c, ok := scanTypeColors[pt]; ok {
		color = c
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color(color)).
		Padding(0, 1).
		Bold(true).
		Render(string(pt))
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan root directories for projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Roots) == 0 {
			fmt.Println("No roots configured. Add one with: curspace roots add <path>")
			return nil
		}

		var projects []scanner.Project

		scanErr := ui.RunWithSpinner("Scanning for projects...", func() error {
			result, scanInnerErr := discovery.Discover(context.Background(), discovery.Options{
				Roots:        cfg.Roots,
				MaxDepth:     cfg.MaxDepth,
				ForceRefresh: true,
			})
			projects = result.Projects
			return scanInnerErr
		})
		if scanErr != nil {
			return fmt.Errorf("scanning: %w", scanErr)
		}

		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return nil
		}

		fmt.Println()
		fmt.Println(scanHeaderStyle.Render("  Discovered Projects"))
		fmt.Println()

		for _, p := range projects {
			fmt.Printf("  %s %s  %s\n",
				scanTypeTag(p.Type),
				scanNameStyle.Render(p.Name),
				scanPathStyle.Render(p.Path),
			)
		}

		fmt.Println()
		fmt.Println(scanCountStyle.Render(fmt.Sprintf("  Found %d project(s)", len(projects))))
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
