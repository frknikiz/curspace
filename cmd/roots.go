package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/spf13/cobra"
)

var (
	rootsSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	rootsPathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	rootsHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	rootsEmptyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#A49FA5")).Italic(true)
	rootsBulletStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
)

var rootsCmd = &cobra.Command{
	Use:   "roots",
	Short: "Manage project root directories",
}

var rootsAddCmd = &cobra.Command{
	Use:   "add <path>",
	Short: "Add a root directory for project scanning",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.AddRoot(args[0]); err != nil {
			return err
		}

		normalized, _ := config.NormalizePath(args[0])
		fmt.Printf("%s Added root: %s\n", rootsSuccessStyle.Render("✓"), rootsPathStyle.Render(normalized))
		return nil
	},
}

var rootsRemoveCmd = &cobra.Command{
	Use:   "remove <path>",
	Short: "Remove a root directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		normalized, err := config.NormalizePath(args[0])
		if err != nil {
			return err
		}

		if err := config.RemoveRoot(args[0]); err != nil {
			return err
		}

		fmt.Printf("%s Removed root: %s\n", rootsSuccessStyle.Render("✓"), rootsPathStyle.Render(normalized))
		return nil
	},
}

var rootsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all root directories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Roots) == 0 {
			fmt.Println(rootsEmptyStyle.Render("No roots configured. Add one with: curspace roots add <path>"))
			return nil
		}

		fmt.Println()
		fmt.Println(rootsHeaderStyle.Render("  Project Roots"))
		fmt.Println()
		for _, root := range cfg.Roots {
			fmt.Printf("  %s %s\n", rootsBulletStyle.Render("▸"), rootsPathStyle.Render(root))
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootsCmd.AddCommand(rootsAddCmd)
	rootsCmd.AddCommand(rootsRemoveCmd)
	rootsCmd.AddCommand(rootsListCmd)
	rootCmd.AddCommand(rootsCmd)
}
