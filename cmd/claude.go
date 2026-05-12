package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/frknikiz/curspace/internal/config"
	"github.com/spf13/cobra"
)

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Manage Claude Code integration",
}

var claudeTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage named Claude API tokens",
}

var claudeTokenAddCmd = &cobra.Command{
	Use:   "add <name> [token]",
	Short: "Save or update a named Claude API token",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := ""
		if len(args) == 2 {
			value = args[1]
		} else {
			fmt.Print("Claude API token: ")
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil && len(input) == 0 {
				return fmt.Errorf("reading token: %w", err)
			}
			value = strings.TrimSpace(input)
		}

		if err := config.SetClaudeToken(args[0], value); err != nil {
			return err
		}
		fmt.Printf("%s Saved Claude token: %s\n", successStyle.Render("✓"), infoStyle.Render(args[0]))
		return nil
	},
}

var claudeTokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved Claude API token names",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.ClaudeTokens) == 0 {
			fmt.Println(wsEmptyStyle.Render("No Claude tokens saved."))
			return nil
		}
		for _, token := range cfg.ClaudeTokens {
			fmt.Printf("  %s %s\n", wsBulletStyle.Render("▸"), wsNameStyle.Render(token.Name))
		}
		return nil
	},
}

var claudeTokenRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a saved Claude API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveClaudeToken(args[0]); err != nil {
			return err
		}
		fmt.Printf("%s Removed Claude token: %s\n", successStyle.Render("✓"), infoStyle.Render(args[0]))
		return nil
	},
}

var claudeTokenPrintCmd = &cobra.Command{
	Use:    "print <name>",
	Short:  "Print a saved Claude API token",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := config.ClaudeTokenValue(args[0])
		if err != nil {
			return err
		}
		fmt.Print(value)
		return nil
	},
}

func init() {
	claudeTokenCmd.AddCommand(claudeTokenAddCmd)
	claudeTokenCmd.AddCommand(claudeTokenListCmd)
	claudeTokenCmd.AddCommand(claudeTokenRemoveCmd)
	claudeTokenCmd.AddCommand(claudeTokenPrintCmd)
	claudeCmd.AddCommand(claudeTokenCmd)
	rootCmd.AddCommand(claudeCmd)
}
