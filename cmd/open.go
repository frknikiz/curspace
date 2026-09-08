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
	"github.com/frknikiz/curspace/internal/codex"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/cursor"
	"github.com/frknikiz/curspace/internal/discovery"
	"github.com/frknikiz/curspace/internal/litellm"
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
	editorCodex  = "codex"
)

func validateEditor(value string) error {
	switch value {
	case editorCursor, editorClaude, editorCodex:
		return nil
	default:
		return fmt.Errorf("invalid --editor value %q (allowed: cursor, claude, codex)", value)
	}
}

func launchEditor(editor string, folders []workspace.WorkspaceFolder, wsPath string) error {
	switch editor {
	case editorClaude, editorCodex:
		if len(folders) == 0 {
			return fmt.Errorf("no folders to open in %s", editor)
		}
		extras := make([]string, 0, len(folders)-1)
		for _, f := range folders[1:] {
			extras = append(extras, f.Path)
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if editor == editorCodex {
			tokenName, err := chooseCodexTokenName(cfg.CodexTokens)
			if err != nil {
				return err
			}
			model, err := chooseModelName("Codex", cfg.LiteLLMBaseURL, cfg.CodexModel, tokenName, cfg.CodexTokens, "OPENAI_API_KEY")
			if err != nil {
				return err
			}
			return codex.Open(folders[0].Path, extras, cfg.Terminal, tokenName, model, cfg.LiteLLMBaseURL)
		}
		tokenName, err := chooseClaudeTokenName(cfg.ClaudeTokens)
		if err != nil {
			return err
		}
		model, err := chooseModelName("Claude", cfg.LiteLLMBaseURL, cfg.ClaudeModel, tokenName, cfg.ClaudeTokens, "ANTHROPIC_AUTH_TOKEN")
		if err != nil {
			return err
		}
		return claude.Open(folders[0].Path, extras, cfg.Terminal, tokenName, model, cfg.LiteLLMBaseURL)
	default:
		return cursor.Open(wsPath)
	}
}

func chooseModelName(provider, baseURL, configured, tokenName string, tokens any, envName string) (string, error) {
	if strings.TrimSpace(baseURL) == "" {
		return strings.TrimSpace(configured), nil
	}
	var tokenValue string
	switch values := tokens.(type) {
	case []config.ClaudeToken:
		for _, token := range values {
			if token.Name == tokenName {
				tokenValue = token.Value
				break
			}
		}
	case []config.CodexToken:
		for _, token := range values {
			if token.Name == tokenName {
				tokenValue = token.Value
				break
			}
		}
	}
	if tokenValue == "" {
		tokenValue = os.Getenv(envName)
	}
	if tokenValue == "" && envName == "OPENAI_API_KEY" {
		tokenValue = os.Getenv("CODEX_API_KEY")
	}
	models, err := litellm.Fetch(context.Background(), baseURL, tokenValue)
	if err != nil {
		return "", fmt.Errorf("fetching %s models: %w", provider, err)
	}

	fmt.Println()
	fmt.Println(infoStyle.Render(provider + " model"))
	for i, model := range models {
		fmt.Printf("  %d) %s\n", i+1, model.ID)
	}
	defaultChoice := 1
	for i, model := range models {
		if model.ID == configured {
			defaultChoice = i + 1
			break
		}
	}
	fmt.Printf("Select model [%d]: ", defaultChoice)
	input, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
	if readErr != nil && len(input) == 0 {
		return "", fmt.Errorf("reading model selection: %w", readErr)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return models[defaultChoice-1].ID, nil
	}
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(models) {
		return "", fmt.Errorf("invalid model selection %q", input)
	}
	return models[choice-1].ID, nil
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

func chooseCodexTokenName(tokens []config.CodexToken) (string, error) {
	if len(tokens) == 0 {
		return "", nil
	}

	fmt.Println()
	fmt.Println(infoStyle.Render("Codex token"))
	for i, token := range tokens {
		fmt.Printf("  %d) %s\n", i+1, token.Name)
	}
	fmt.Println("  0) current Codex login / environment")
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
	Short: "Select projects and open as workspace in Cursor, Claude Code, or Codex CLI",
	Long:  "Scans for projects (or uses cache), presents a TUI selector, creates a workspace, and opens it in Cursor, Claude Code, or Codex CLI.",
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
	openCmd.Flags().StringVarP(&openEditor, "editor", "e", editorCursor, "Editor to launch: cursor, claude, or codex")
	rootCmd.AddCommand(openCmd)
}
