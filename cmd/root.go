package cmd

import (
	"fmt"
	"os"

	"github.com/frknikiz/curspace/internal/claude"
	"github.com/frknikiz/curspace/internal/codex"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/cursor"
	"github.com/frknikiz/curspace/internal/ui"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "curspace",
	Short:   "Terminal-first project discovery and workspace launcher for Cursor, Claude Code, and Codex CLI",
	Long:    "Curspace discovers projects across your filesystem, lets you select them via TUI, and launches multi-folder workspaces in Cursor, Claude Code, or Codex CLI.",
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return ui.RunApp(ui.AppConfig{
			Roots:          cfg.Roots,
			MaxDepth:       cfg.MaxDepth,
			Terminal:       cfg.Terminal,
			DefaultEditor:  cfg.DefaultEditor,
			ClaudeTokens:   cfg.ClaudeTokens,
			CodexTokens:    cfg.CodexTokens,
			LiteLLMBaseURL: cfg.LiteLLMBaseURL,
			ClaudeModel:    cfg.ClaudeModel,
			CodexModel:     cfg.CodexModel,
			OpenCursor:     cursor.Open,
			OpenCodex: func(primaryPath string, extraPaths []string, tokenName string) error {
				current, err := config.Load()
				if err != nil {
					return err
				}
				return codex.Open(primaryPath, extraPaths, current.Terminal, tokenName)
			},
			OpenClaude: func(primaryPath string, extraPaths []string, tokenName string) error {
				current, err := config.Load()
				if err != nil {
					return err
				}
				return claude.Open(primaryPath, extraPaths, current.Terminal, tokenName)
			},
			OpenCodexWithModel: func(primaryPath string, extraPaths []string, tokenName, model string) error {
				current, err := config.Load()
				if err != nil {
					return err
				}
				return codex.Open(primaryPath, extraPaths, current.Terminal, tokenName, model, current.LiteLLMBaseURL)
			},
			OpenClaudeWithModel: func(primaryPath string, extraPaths []string, tokenName, model string) error {
				current, err := config.Load()
				if err != nil {
					return err
				}
				return claude.Open(primaryPath, extraPaths, current.Terminal, tokenName, model, current.LiteLLMBaseURL)
			},
		})
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
