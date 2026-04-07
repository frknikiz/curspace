package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/scanner"
)

var (
	promptTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 2).
				MarginBottom(1)

	promptBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	promptLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A49FA5")).
				MarginBottom(1)

	promptHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	promptHelpKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	promptHelpDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A49FA5"))

	promptAppStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

type PromptModel struct {
	textInput textinput.Model
	projects  []scanner.Project
	done      bool
	cancelled bool
}

func NewPromptModel(placeholder string) PromptModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Italic(true)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.Prompt = "❯ "
	return PromptModel{textInput: ti}
}

func (m PromptModel) Value() string {
	if v := m.textInput.Value(); v != "" {
		return v
	}
	return generateWorkspaceName(m.projects)
}

func (m PromptModel) Cancelled() bool {
	return m.cancelled
}

func (m PromptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m PromptModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	title := promptTitleStyle.Render(" CURSPACE  Workspace Name ")
	label := promptLabelStyle.Render("Choose a name for your new workspace:")
	input := m.textInput.View()

	box := promptBoxStyle.Render(fmt.Sprintf("%s\n\n%s", label, input))

	hint := promptHelpStyle.Render("Press enter without a name to use the auto-generated one.")

	help := promptHelpStyle.Render(
		promptHelpKeyStyle.Render("enter") + promptHelpDescStyle.Render(" confirm") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#3C3C3C")).Render(" │ ") +
			promptHelpKeyStyle.Render("esc") + promptHelpDescStyle.Render(" cancel"),
	)

	return promptAppStyle.Render(fmt.Sprintf("%s\n%s\n%s\n%s", title, box, hint, help))
}

func RunPrompt(projects []scanner.Project) (string, error) {
	placeholder := generateWorkspaceName(projects)
	model := NewPromptModel(placeholder)
	model.projects = projects
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("running prompt: %w", err)
	}

	final := result.(PromptModel)
	if final.Cancelled() {
		return "", nil
	}

	return final.Value(), nil
}
