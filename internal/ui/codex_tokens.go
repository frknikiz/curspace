package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/frknikiz/curspace/internal/config"
)

type codexTokenPick struct {
	editorPick editorPick
	cursor     int
}

func (m AppModel) updateCodexTokenPick(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	maxCursor := len(m.codexTokens)
	switch km.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.view = viewMain
		return m, loadWorkspacesCmd
	case "up", "k":
		if m.codexTokenPick.cursor > 0 {
			m.codexTokenPick.cursor--
		}
	case "down", "j":
		if m.codexTokenPick.cursor < maxCursor {
			m.codexTokenPick.cursor++
		}
	case "enter":
		tokenName := ""
		if m.codexTokenPick.cursor < len(m.codexTokens) {
			tokenName = m.codexTokens[m.codexTokenPick.cursor].Name
		}
		return m.beginModelPick(m.codexTokenPick.editorPick, tokenName, "codex")
	}

	return m, nil
}

func (m AppModel) launchCodex(pick editorPick, tokenName, model string) (tea.Model, tea.Cmd) {
	m.view = viewMain
	var err error
	if m.openCodexWithModel != nil {
		err = m.openCodexWithModel(pick.primaryPath, pick.extraPaths, tokenName, model)
	} else if m.openCodex != nil {
		err = m.openCodex(pick.primaryPath, pick.extraPaths, tokenName)
	} else {
		err = fmt.Errorf("codex launcher is not configured")
	}
	if err != nil {
		m.statusMsg = fmt.Sprintf("Codex: %v", err)
		m.statusErr = true
	} else if tokenName != "" {
		m.statusMsg = fmt.Sprintf("Opened '%s' in Codex with token '%s'", pick.label, tokenName)
		m.statusErr = false
	} else {
		m.statusMsg = fmt.Sprintf("Opened '%s' in Codex", pick.label)
		m.statusErr = false
	}
	return m, loadWorkspacesCmd
}

func (m AppModel) renderCodexTokenPick() string {
	var s []string
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("choose Codex token"))
	s = append(s, "")
	s = append(s, "  "+appDetailStyle.Render(fmt.Sprintf("Target: %s", m.codexTokenPick.editorPick.label)))
	s = append(s, "")

	for i, token := range m.codexTokens {
		isActive := i == m.codexTokenPick.cursor
		cur := "  "
		if isActive {
			cur = appCursorStyle.Render("▸ ")
		}
		name := appNameStyle.Render(token.Name)
		if isActive {
			name = appSelectedNameStyle.Render(token.Name)
		}
		s = append(s, fmt.Sprintf("  %s%s  %s", cur, name, appDetailStyle.Render("use selected OpenAI API key")))
	}

	noTokenIdx := len(m.codexTokens)
	isActive := noTokenIdx == m.codexTokenPick.cursor
	cur := "  "
	if isActive {
		cur = appCursorStyle.Render("▸ ")
	}
	name := appNameStyle.Render("current Codex login / environment")
	if isActive {
		name = appSelectedNameStyle.Render("current Codex login / environment")
	}
	s = append(s, fmt.Sprintf("  %s%s", cur, name))

	s = append(s, "")
	items := []struct{ key, desc string }{
		{"↑↓", "navigate"},
		{"↵", "open"},
		{"esc", "cancel"},
	}
	s = append(s, "  "+renderHelp(items))
	return appPadding.Render(strings.Join(s, "\n"))
}

func (m AppModel) updateCodexTokens(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.confirmingToken {
		switch km.String() {
		case "y", "Y":
			if len(m.codexTokens) > 0 && m.tokenCursor < len(m.codexTokens) {
				name := m.codexTokens[m.tokenCursor].Name
				if err := config.RemoveCodexToken(name); err != nil {
					m.statusMsg = err.Error()
					m.statusErr = true
				} else {
					_ = m.refreshConfig()
					if m.tokenCursor >= len(m.codexTokens) {
						m.tokenCursor = max(0, len(m.codexTokens)-1)
					}
					m.statusMsg = fmt.Sprintf("Removed Codex token '%s'", name)
					m.statusErr = false
				}
			}
		}
		m.confirmingToken = false
		return m, nil
	}

	switch km.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.view = viewSettings
		return m, nil
	case "up", "k":
		if m.tokenCursor > 0 {
			m.tokenCursor--
		}
	case "down", "j":
		if m.tokenCursor < len(m.codexTokens)-1 {
			m.tokenCursor++
		}
	case "a", "enter":
		m.pendingTokenName = ""
		ti := newStyledInput("work")
		ti.Focus()
		m.tokenNameInput = ti
		m.view = viewCodexTokenName
		return m, textinput.Blink
	case "d":
		if len(m.codexTokens) > 0 {
			m.confirmingToken = true
		}
	}

	return m, nil
}

func (m AppModel) updateCodexTokenName(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEsc:
			m.view = viewCodexTokens
			return m, nil
		case tea.KeyEnter:
			name := strings.TrimSpace(m.tokenNameInput.Value())
			if name == "" {
				m.statusMsg = "Token name cannot be empty"
				m.statusErr = true
				m.view = viewCodexTokens
				return m, nil
			}
			m.pendingTokenName = name
			ti := newStyledInput("sk-...")
			ti.EchoMode = textinput.EchoPassword
			ti.CharLimit = 4096
			ti.Width = 72
			ti.Focus()
			m.tokenValueInput = ti
			m.view = viewCodexTokenValue
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.tokenNameInput, cmd = m.tokenNameInput.Update(msg)
	return m, cmd
}

func (m AppModel) updateCodexTokenValue(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEsc:
			m.view = viewCodexTokens
			return m, nil
		case tea.KeyEnter:
			value := strings.TrimSpace(m.tokenValueInput.Value())
			if err := config.SetCodexToken(m.pendingTokenName, value); err != nil {
				m.statusMsg = err.Error()
				m.statusErr = true
			} else {
				_ = m.refreshConfig()
				m.tokenCursor = codexTokenIndexByName(m.codexTokens, m.pendingTokenName)
				m.statusMsg = fmt.Sprintf("Saved Codex token '%s'", m.pendingTokenName)
				m.statusErr = false
			}
			m.pendingTokenName = ""
			m.view = viewCodexTokens
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.tokenValueInput, cmd = m.tokenValueInput.Update(msg)
	return m, cmd
}

func codexTokenIndexByName(tokens []config.CodexToken, name string) int {
	for i, token := range tokens {
		if token.Name == name {
			return i
		}
	}
	return 0
}

func (m AppModel) renderCodexTokens() string {
	var s []string
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("Codex tokens"))
	s = append(s, "")
	s = append(s, "  "+appMutedHintStyle.Render("Token values are saved to ~/.curspace/config.json and are never shown here."))
	s = append(s, "")

	if len(m.codexTokens) == 0 {
		s = append(s, appEmptyStyle.Render("No Codex tokens saved. Press a to add one."))
	} else {
		for i, token := range m.codexTokens {
			isActive := i == m.tokenCursor
			cur := "  "
			if isActive {
				cur = appCursorStyle.Render("▸ ")
			}
			name := appNameStyle.Render(token.Name)
			if isActive {
				name = appSelectedNameStyle.Render(token.Name)
			}
			s = append(s, fmt.Sprintf("  %s%s  %s", cur, name, appDetailStyle.Render("saved")))
		}
	}

	if m.statusMsg != "" {
		s = append(s, "")
		style := appStatusOkStyle
		prefix := "✓"
		if m.statusErr {
			style = appStatusErrStyle
			prefix = "✗"
		}
		s = append(s, "  "+style.Render(prefix)+" "+style.Render(m.statusMsg))
	}

	if m.confirmingToken && len(m.codexTokens) > 0 && m.tokenCursor < len(m.codexTokens) {
		s = append(s, "")
		s = append(s, "  "+appConfirmStyle.Render(
			fmt.Sprintf("Remove token '%s'? (y/n)", m.codexTokens[m.tokenCursor].Name),
		))
	}

	s = append(s, "")
	items := []struct{ key, desc string }{
		{"a", "add"},
		{"d", "remove"},
		{"↑↓", "navigate"},
		{"esc", "back"},
	}
	s = append(s, "  "+renderHelp(items))
	return appPadding.Render(strings.Join(s, "\n"))
}

func (m AppModel) renderCodexTokenName() string {
	var s []string
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("new Codex token"))
	s = append(s, "")
	box := inputBoxStyle.Render(fmt.Sprintf("%s\n\n%s", inputLabelStyle.Render("Token name:"), m.tokenNameInput.View()))
	s = append(s, box)
	s = append(s, "")
	s = append(s, "  "+renderHelp([]struct{ key, desc string }{
		{"↵", "continue"},
		{"esc", "cancel"},
	}))
	return appPadding.Render(strings.Join(s, "\n"))
}

func (m AppModel) renderCodexTokenValue() string {
	var s []string
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("new Codex token"))
	s = append(s, "")
	label := inputLabelStyle.Render(fmt.Sprintf("Token value for %s:", m.pendingTokenName))
	box := inputBoxStyle.Render(fmt.Sprintf("%s\n\n%s", label, m.tokenValueInput.View()))
	s = append(s, box)
	s = append(s, "")
	s = append(s, "  "+appMutedHintStyle.Render("The value will be saved, then hidden from token lists."))
	s = append(s, "")
	s = append(s, "  "+renderHelp([]struct{ key, desc string }{
		{"↵", "save"},
		{"esc", "cancel"},
	}))
	return appPadding.Render(strings.Join(s, "\n"))
}
