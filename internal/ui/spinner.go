package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SpinnerDoneMsg struct {
	Err error
}

type SpinnerModel struct {
	spinner  spinner.Model
	message  string
	done     bool
	err      error
	work     func() error
	quitting bool
}

func NewSpinnerModel(message string, work func() error) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	return SpinnerModel{
		spinner: s,
		message: message,
		work:    work,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runWork())
}

func (m SpinnerModel) runWork() tea.Cmd {
	return func() tea.Msg {
		err := m.work()
		return SpinnerDoneMsg{Err: err}
	}
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case SpinnerDoneMsg:
		m.done = true
		m.err = msg.Err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m SpinnerModel) View() string {
	if m.done {
		if m.err != nil {
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
			return errStyle.Render("✗ " + m.err.Error())
		}
		doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
		return doneStyle.Render("✓ Done!")
	}

	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A49FA5"))
	return fmt.Sprintf("%s %s", m.spinner.View(), msgStyle.Render(m.message))
}

func (m SpinnerModel) Quitting() bool {
	return m.quitting
}

func (m SpinnerModel) Error() error {
	return m.err
}

func RunWithSpinner(message string, work func() error) error {
	model := NewSpinnerModel(message, work)
	p := tea.NewProgram(model)
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("spinner: %w", err)
	}

	final := result.(SpinnerModel)
	if final.Quitting() {
		return fmt.Errorf("cancelled")
	}

	return final.Error()
}
