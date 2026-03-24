package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/cursor"
	"github.com/frknikiz/curspace/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	wsSuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	wsNameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	wsHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	wsEmptyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#A49FA5")).Italic(true)
	wsBulletStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage saved workspaces",
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved workspaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := workspace.List()
		if err != nil {
			return err
		}

		if len(names) == 0 {
			fmt.Println(wsEmptyStyle.Render("No workspaces found."))
			return nil
		}

		fmt.Println()
		fmt.Println(wsHeaderStyle.Render("  Saved Workspaces"))
		fmt.Println()
		for _, name := range names {
			fmt.Printf("  %s %s\n", wsBulletStyle.Render("▸"), wsNameStyle.Render(name))
		}
		fmt.Println()
		return nil
	},
}

var workspaceOpenCmd = &cobra.Command{
	Use:   "open <name>",
	Short: "Open a saved workspace in Cursor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := workspace.Open(args[0])
		if err != nil {
			return err
		}

		if err := cursor.Open(path); err != nil {
			return fmt.Errorf("opening in Cursor: %w", err)
		}

		fmt.Printf("%s Opened workspace %s in Cursor\n", wsSuccessStyle.Render("✓"), wsNameStyle.Render(args[0]))
		return nil
	},
}

var workspaceDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a saved workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := workspace.Delete(args[0]); err != nil {
			return err
		}

		fmt.Printf("%s Deleted workspace: %s\n", wsSuccessStyle.Render("✓"), wsNameStyle.Render(args[0]))
		return nil
	},
}

var workspaceRenameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a saved workspace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := workspace.Rename(args[0], args[1]); err != nil {
			return err
		}

		fmt.Printf("%s Renamed %s → %s\n", wsSuccessStyle.Render("✓"), wsNameStyle.Render(args[0]), wsNameStyle.Render(args[1]))
		return nil
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceOpenCmd)
	workspaceCmd.AddCommand(workspaceDeleteCmd)
	workspaceCmd.AddCommand(workspaceRenameCmd)
	rootCmd.AddCommand(workspaceCmd)
}
