package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/litellm"
)

type modelPick struct {
	editorPick   editorPick
	provider     string
	tokenName    string
	models       []litellm.Model
	cursor       int
	search       string
	settingsKind int // 5 = Claude default, 6 = Codex default
}

func (m AppModel) beginModelPick(pick editorPick, tokenName, provider string) (tea.Model, tea.Cmd) {
	configured := m.codexModel
	if provider == "claude" {
		configured = m.claudeModel
	}
	if strings.TrimSpace(m.litellmBaseURL) == "" {
		if provider == "codex" {
			return m.launchCodex(pick, tokenName, configured)
		}
		return m.launchClaude(pick, tokenName, configured)
	}

	token := ""
	if provider == "codex" {
		for _, item := range m.codexTokens {
			if item.Name == tokenName {
				token = item.Value
				break
			}
		}
		if token == "" {
			token = os.Getenv("OPENAI_API_KEY")
		}
		if token == "" {
			token = os.Getenv("CODEX_API_KEY")
		}
	} else {
		for _, item := range m.claudeTokens {
			if item.Name == tokenName {
				token = item.Value
				break
			}
		}
		if token == "" {
			token = os.Getenv("ANTHROPIC_AUTH_TOKEN")
		}
		if token == "" {
			token = os.Getenv("ANTHROPIC_API_KEY")
		}
	}

	m.modelPick = modelPick{editorPick: pick, provider: provider, tokenName: tokenName}
	m.view = viewModelLoading
	return m, fetchModelsCmd(m.litellmBaseURL, token, provider)
}

func (m AppModel) beginSettingsModelPick(kind int) (tea.Model, tea.Cmd) {
	provider := "codex"
	token := ""
	if kind == 5 {
		provider = "claude"
		if len(m.claudeTokens) > 0 {
			token = m.claudeTokens[0].Value
		}
		if token == "" {
			token = os.Getenv("ANTHROPIC_AUTH_TOKEN")
		}
		if token == "" {
			token = os.Getenv("ANTHROPIC_API_KEY")
		}
	} else {
		if len(m.codexTokens) > 0 {
			token = m.codexTokens[0].Value
		}
		if token == "" {
			token = os.Getenv("OPENAI_API_KEY")
		}
		if token == "" {
			token = os.Getenv("CODEX_API_KEY")
		}
	}
	if strings.TrimSpace(m.litellmBaseURL) == "" {
		return m, nil
	}
	m.modelPick = modelPick{provider: provider, settingsKind: kind}
	m.view = viewModelLoading
	return m, fetchModelsCmd(m.litellmBaseURL, token, provider)
}

func fetchModelsCmd(baseURL, token, provider string) tea.Cmd {
	return func() tea.Msg {
		models, err := litellm.Fetch(context.Background(), baseURL, token)
		return modelFetchDoneMsg{provider: provider, models: models, err: err}
	}
}

func (m AppModel) finishModelFetch(msg modelFetchDoneMsg) (tea.Model, tea.Cmd) {
	if m.view != viewModelLoading || msg.provider != m.modelPick.provider {
		return m, nil
	}
	if msg.err != nil {
		m.view = viewMain
		m.statusMsg = fmt.Sprintf("%s models: %v", msg.provider, msg.err)
		m.statusErr = true
		return m, loadWorkspacesCmd
	}
	m.modelPick.models = msg.models
	m.modelPick.cursor = 0
	configured := m.codexModel
	if msg.provider == "claude" {
		configured = m.claudeModel
	}
	for i, model := range msg.models {
		if model.ID == configured {
			m.modelPick.cursor = i
			break
		}
	}
	if msg.provider == "codex" {
		m.view = viewCodexModelPick
	} else {
		m.view = viewClaudeModelPick
	}
	return m, nil
}

func (m AppModel) updateModelLoading(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && (km.String() == "esc" || km.String() == "ctrl+c") {
		if km.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		m.view = viewMain
		return m, loadWorkspacesCmd
	}
	return m, nil
}

func (m AppModel) updateModelPick(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.view = viewMain
		return m, loadWorkspacesCmd
	case "up", "k":
		if m.modelPick.cursor > 0 {
			m.modelPick.cursor--
		}
	case "down", "j":
		if m.modelPick.cursor < len(m.filteredModelIndexes())-1 {
			m.modelPick.cursor++
		}
	case "enter":
		filtered := m.filteredModelIndexes()
		if len(filtered) == 0 {
			return m, nil
		}
		selected := m.modelPick.models[filtered[m.modelPick.cursor]].ID
		if m.modelPick.settingsKind != 0 {
			m.statusErr = false
			if m.modelPick.settingsKind == 5 {
				m.claudeModel = selected
			} else {
				m.codexModel = selected
			}
			if cfg, err := config.Load(); err == nil {
				cfg.ClaudeModel, cfg.CodexModel = m.claudeModel, m.codexModel
				if err := config.Save(cfg); err != nil {
					m.statusMsg, m.statusErr = err.Error(), true
				}
			} else {
				m.statusMsg, m.statusErr = err.Error(), true
			}
			if !m.statusErr {
				m.statusMsg = "Default model saved"
			}
			m.view = viewSettings
			return m, nil
		}
		if m.modelPick.provider == "codex" {
			return m.launchCodex(m.modelPick.editorPick, m.modelPick.tokenName, selected)
		}
		return m.launchClaude(m.modelPick.editorPick, m.modelPick.tokenName, selected)
	default:
		if km.Type == tea.KeyRunes {
			m.modelPick.search += string(km.Runes)
			m.modelPick.cursor = 0
		} else if km.Type == tea.KeyBackspace && m.modelPick.search != "" {
			runes := []rune(m.modelPick.search)
			m.modelPick.search = string(runes[:len(runes)-1])
			m.modelPick.cursor = 0
		}
	}
	return m, nil
}

func (m AppModel) filteredModelIndexes() []int {
	query := strings.ToLower(strings.TrimSpace(m.modelPick.search))
	indexes := make([]int, 0, len(m.modelPick.models))
	for i, model := range m.modelPick.models {
		if query == "" || strings.Contains(strings.ToLower(model.ID), query) {
			indexes = append(indexes, i)
		}
	}
	return indexes
}

func (m AppModel) renderModelPick() string {
	title := "Claude model"
	if m.modelPick.provider == "codex" {
		title = "Codex model"
	}
	lines := []string{appTitleStyle.Render(" CURSPACE ") + "  " + appSubtitleStyle.Render("choose "+title), "", "  " + appDetailStyle.Render(fmt.Sprintf("Target: %s", m.modelPick.editorPick.label)), ""}
	filtered := m.filteredModelIndexes()
	lines = append(lines, "  "+appDetailStyle.Render("Search: ")+appNameStyle.Render(m.modelPick.search))
	lines = append(lines, "")
	for i, index := range filtered {
		model := m.modelPick.models[index]
		cursor := "  "
		name := appNameStyle.Render(model.ID)
		if i == m.modelPick.cursor {
			cursor = appCursorStyle.Render("▸ ")
			name = appSelectedNameStyle.Render(model.ID)
		}
		lines = append(lines, "  "+cursor+name)
	}
	lines = append(lines, "", "  "+renderHelp([]struct{ key, desc string }{{"type", "search"}, {"↑↓", "navigate"}, {"↵", "select"}, {"esc", "cancel"}}))
	return appPadding.Render(strings.Join(lines, "\n"))
}
