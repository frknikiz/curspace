package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/frknikiz/curspace/internal/config"
	"github.com/spf13/cobra"
)

var codexCmd = &cobra.Command{
	Use:   "codex",
	Short: "Manage Codex CLI integration",
}

var codexTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage named Codex API tokens",
}

var codexTokenAddCmd = &cobra.Command{
	Use:   "add <name> [token]",
	Short: "Save or update a named Codex API token",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := ""
		if len(args) == 2 {
			value = args[1]
		} else {
			fmt.Print("Codex API token: ")
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil && len(input) == 0 {
				return fmt.Errorf("reading token: %w", err)
			}
			value = strings.TrimSpace(input)
		}

		if err := config.SetCodexToken(args[0], value); err != nil {
			return err
		}
		fmt.Printf("%s Saved Codex token: %s\n", successStyle.Render("✓"), infoStyle.Render(args[0]))
		return nil
	},
}

var codexTokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved Codex API token names",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if len(cfg.CodexTokens) == 0 {
			fmt.Println(wsEmptyStyle.Render("No Codex tokens saved."))
			return nil
		}
		for _, token := range cfg.CodexTokens {
			fmt.Printf("  %s %s\n", wsBulletStyle.Render("▸"), wsNameStyle.Render(token.Name))
		}
		return nil
	},
}

var codexTokenRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a saved Codex API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.RemoveCodexToken(args[0]); err != nil {
			return err
		}
		fmt.Printf("%s Removed Codex token: %s\n", successStyle.Render("✓"), infoStyle.Render(args[0]))
		return nil
	},
}

var codexTokenPrintCmd = &cobra.Command{
	Use:    "print <name>",
	Short:  "Print a saved Codex API token",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := config.CodexTokenValue(args[0])
		if err != nil {
			return err
		}
		fmt.Print(value)
		return nil
	},
}

func init() {
	codexTokenCmd.AddCommand(codexTokenAddCmd)
	codexTokenCmd.AddCommand(codexTokenListCmd)
	codexTokenCmd.AddCommand(codexTokenRemoveCmd)
	codexTokenCmd.AddCommand(codexTokenPrintCmd)
	codexCmd.AddCommand(codexTokenCmd)
	rootCmd.AddCommand(codexCmd)
}
