package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/scanner"
)

var (
	ordTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginBottom(1)

	ordNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			Width(3).
			Align(lipgloss.Right)

	ordActiveNumberStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true).
				Width(3).
				Align(lipgloss.Right)

	ordNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	ordActiveNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Underline(true)

	ordPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	ordCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	ordHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5"))

	ordMovingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	ordAppStyle = lipgloss.NewStyle().Padding(1, 2)
)

type OrdererModel struct {
	items     []scanner.Project
	cursor    int
	done      bool
	cancelled bool
	width     int
	height    int
	scrollTop int
}

func NewOrdererModel(projects []scanner.Project) OrdererModel {
	return OrdererModel{
		items: projects,
	}
}

func (m OrdererModel) Items() []scanner.Project {
	return m.items
}

func (m OrdererModel) Cancelled() bool {
	return m.cancelled
}

func (m OrdererModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m OrdererModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit

		case "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case "ctrl+up", "ctrl+k":
			if m.cursor > 0 {
				m.items[m.cursor], m.items[m.cursor-1] = m.items[m.cursor-1], m.items[m.cursor]
				m.cursor--
				m.ensureVisible()
			}

		case "ctrl+down", "ctrl+j":
			if m.cursor < len(m.items)-1 {
				m.items[m.cursor], m.items[m.cursor+1] = m.items[m.cursor+1], m.items[m.cursor]
				m.cursor++
				m.ensureVisible()
			}
		}
	}

	return m, nil
}

func (m *OrdererModel) ensureVisible() {
	maxVis := m.maxVisible()
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}
	if m.cursor >= m.scrollTop+maxVis {
		m.scrollTop = m.cursor - maxVis + 1
	}
}

func (m OrdererModel) maxVisible() int {
	return max(5, m.height-12)
}

func (m OrdererModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var s []string

	s = append(s, ordTitleStyle.Render(" CURSPACE  Arrange Projects "))
	s = append(s, "")
	s = append(s, ordHintStyle.Render("  The first project becomes the primary workspace folder."))
	s = append(s, ordHintStyle.Render("  Use ctrl+↑/↓ to move items, ↵ to confirm."))
	s = append(s, "")

	if len(m.items) == 0 {
		s = append(s, selNoMatchStyle.Render("No projects to order."))
	} else {
		maxVis := m.maxVisible()
		end := min(m.scrollTop+maxVis, len(m.items))

		if m.scrollTop > 0 {
			s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▲ %d more", m.scrollTop)))
		}

		for i := m.scrollTop; i < end; i++ {
			p := m.items[i]
			isActive := i == m.cursor

			cur := "  "
			if isActive {
				cur = ordCursorStyle.Render("▸ ")
			}

			numStyle := ordNumberStyle
			if isActive {
				numStyle = ordActiveNumberStyle
			}
			num := numStyle.Render(fmt.Sprintf("%d.", i+1))

			name := ordNameStyle.Render(p.Name)
			if isActive {
				name = ordActiveNameStyle.Render(p.Name)
			}

			tag := renderProjectTypeTag(p.Type)

			maxLen := 50
			if m.width > 0 {
				maxLen = max(20, m.width-len(p.Name)-len(string(p.Type))-28)
			}
			pathStr := ordPathStyle.Render(truncatePath(p.Path, maxLen))

			s = append(s, fmt.Sprintf("  %s%s %s %s  %s", cur, num, name, tag, pathStr))
		}

		remaining := len(m.items) - end
		if remaining > 0 {
			s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▼ %d more", remaining)))
		}
	}

	s = append(s, "")
	items := []struct{ key, desc string }{
		{"↑↓", "navigate"},
		{"ctrl+↑↓", "move"},
		{"↵", "confirm"},
		{"esc", "cancel"},
	}
	s = append(s, "  "+renderHelp(items))

	return ordAppStyle.Render(strings.Join(s, "\n"))
}

func RunOrderer(projects []scanner.Project) ([]scanner.Project, error) {
	model := NewOrdererModel(projects)
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running orderer: %w", err)
	}

	final := result.(OrdererModel)
	if final.Cancelled() {
		return nil, nil
	}

	return final.Items(), nil
}
