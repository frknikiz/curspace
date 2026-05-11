package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/claude"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/cursor"
	"github.com/frknikiz/curspace/internal/discovery"
	"github.com/frknikiz/curspace/internal/scanner"
	"github.com/frknikiz/curspace/internal/ui"
	"github.com/frknikiz/curspace/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	refreshFlag bool
	openEditor  string
)

const (
	editorCursor = "cursor"
	editorClaude = "claude"
)

func validateEditor(value string) error {
	switch value {
	case editorCursor, editorClaude:
		return nil
	default:
		return fmt.Errorf("invalid --editor value %q (allowed: cursor, claude)", value)
	}
}

func launchEditor(editor string, folders []workspace.WorkspaceFolder, wsPath string) error {
	switch editor {
	case editorClaude:
		if len(folders) == 0 {
			return fmt.Errorf("no folders to open in Claude")
		}
		extras := make([]string, 0, len(folders)-1)
		for _, f := range folders[1:] {
			extras = append(extras, f.Path)
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		tokenName, err := chooseClaudeTokenName(cfg.ClaudeTokens)
		if err != nil {
			return err
		}
		return claude.Open(folders[0].Path, extras, cfg.Terminal, tokenName)
	default:
		return cursor.Open(wsPath)
	}
}

func chooseClaudeTokenName(tokens []config.ClaudeToken) (string, error) {
	if len(tokens) == 0 {
		return "", nil
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Claude token"))
	for i, token := range tokens {
		fmt.Printf("  %d) %s\n", i+1, token.Name)
	}
	fmt.Println("  0) current Claude login / environment")
	fmt.Print("Select token [1]: ")

	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(input) == 0 {
		return "", fmt.Errorf("reading token selection: %w", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return tokens[0].Name, nil
	}

	choice, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid token selection %q", input)
	}
	if choice == 0 {
		return "", nil
	}
	if choice < 1 || choice > len(tokens) {
		return "", fmt.Errorf("token selection out of range: %d", choice)
	}
	return tokens[choice-1].Name, nil
}

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD43B"))
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Select projects and open as workspace in Cursor or Claude",
	Long:  "Scans for projects (or uses cache), presents a TUI selector, creates a workspace, and opens it in Cursor or Claude Code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateEditor(openEditor); err != nil {
			return err
		}

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

		ordered, err := ui.RunOrderer(selected)
		if err != nil {
			return fmt.Errorf("project ordering: %w", err)
		}
		if ordered == nil {
			fmt.Println("Cancelled.")
			return nil
		}

		fmt.Printf("%s Project order confirmed\n\n", successStyle.Render("✓"))

		wsName, err := ui.RunPrompt(ordered)
		if err != nil {
			return fmt.Errorf("workspace name input: %w", err)
		}
		if wsName == "" {
			fmt.Println("Cancelled.")
			return nil
		}

		folders := make([]workspace.WorkspaceFolder, len(ordered))
		for i, p := range ordered {
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

		if err := launchEditor(openEditor, folders, wsPath); err != nil {
			return fmt.Errorf("opening in %s: %w", openEditor, err)
		}

		fmt.Printf("%s Opened in %s!\n", successStyle.Render("✓"), openEditor)
		return nil
	},
}

func init() {
	openCmd.Flags().BoolVar(&refreshFlag, "refresh", false, "Bypass cache and rescan projects")
	openCmd.Flags().StringVarP(&openEditor, "editor", "e", editorCursor, "Editor to launch: cursor or claude")
	rootCmd.AddCommand(openCmd)
}
