package cmd

import (
	"fmt"
	"os"

	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/cursor"
	"github.com/frknikiz/curspace/internal/ui"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "curspace",
	Short:   "Terminal-first project discovery and workspace launcher for Cursor IDE",
	Long:    "Curspace discovers projects across your filesystem, lets you select them via TUI, and launches multi-folder workspaces in Cursor.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return ui.RunApp(ui.AppConfig{
			Roots:      cfg.Roots,
			MaxDepth:   cfg.MaxDepth,
			OpenCursor: cursor.Open,
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
