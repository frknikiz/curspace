package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/scanner"
)

// ── Styles ────────────────────────────────────────────────────────

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginBottom(1)

	searchBarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	searchLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	searchTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	searchCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				Blink(true)

	activeItemStyle = lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(lipgloss.Color("#FAFAFA")).
			Bold(true)

	checkOn = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	checkOff = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	cursorIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	projectNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	projectNameActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Underline(true)

	projectPathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Italic(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			MarginTop(1)

	selectedCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true)

	totalCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	helpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5"))

	helpSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C3C3C"))

	noMatchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Italic(true).
			PaddingLeft(4).
			MarginTop(1)

	scrollHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			Align(lipgloss.Right)
)

// ── Model ─────────────────────────────────────────────────────────

type SelectorModel struct {
	items      []scanner.Project
	filtered   []int
	selected   map[int]bool
	cursor     int
	search     string
	quitting   bool
	done       bool
	width      int
	height     int
	scrollTop  int
	maxVisible int
}

func NewSelectorModel(projects []scanner.Project) SelectorModel {
	filtered := make([]int, len(projects))
	for i := range projects {
		filtered[i] = i
	}
	return SelectorModel{
		items:      projects,
		filtered:   filtered,
		selected:   make(map[int]bool),
		maxVisible: 15,
	}
}

func (m SelectorModel) SelectedProjects() []scanner.Project {
	var result []scanner.Project
	for idx := range m.selected {
		if m.selected[idx] {
			result = append(result, m.items[idx])
		}
	}
	return result
}

func (m SelectorModel) Cancelled() bool {
	return m.quitting
}

func (m SelectorModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.maxVisible = max(5, msg.Height-10)
		m.clampScroll()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.search != "" {
				m.search = ""
				m.applyFilter()
			} else {
				m.quitting = true
				return m, tea.Quit
			}

		case "enter":
			if len(m.selected) > 0 {
				m.done = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.search == "" || msg.String() == "up" {
				if m.cursor > 0 {
					m.cursor--
					m.ensureVisible()
				}
				if msg.String() == "k" && m.search == "" {
					return m, nil
				}
				if msg.String() == "up" {
					return m, nil
				}
			}
			m.search += msg.String()
			m.applyFilter()

		case "down", "j":
			if m.search == "" || msg.String() == "down" {
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
					m.ensureVisible()
				}
				if msg.String() == "j" && m.search == "" {
					return m, nil
				}
				if msg.String() == "down" {
					return m, nil
				}
			}
			m.search += msg.String()
			m.applyFilter()

		case " ":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				idx := m.filtered[m.cursor]
				if m.selected[idx] {
					delete(m.selected, idx)
				} else {
					m.selected[idx] = true
				}
			}

		case "tab":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				idx := m.filtered[m.cursor]
				if m.selected[idx] {
					delete(m.selected, idx)
				} else {
					m.selected[idx] = true
				}
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
					m.ensureVisible()
				}
			}

		case "ctrl+a":
			for _, idx := range m.filtered {
				m.selected[idx] = true
			}

		case "ctrl+d":
			m.selected = make(map[int]bool)

		case "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.ensureVisible()
			}

		case "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}

		case "backspace":
			if len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
				m.applyFilter()
			}

		default:
			if msg.Type == tea.KeyRunes {
				m.search += string(msg.Runes)
				m.applyFilter()
			}
		}
	}

	return m, nil
}

func (m *SelectorModel) applyFilter() {
	if m.search == "" {
		m.filtered = make([]int, len(m.items))
		for i := range m.items {
			m.filtered[i] = i
		}
	} else {
		m.filtered = nil
		lower := strings.ToLower(m.search)
		for i, p := range m.items {
			name := strings.ToLower(p.Name)
			path := strings.ToLower(p.Path)
			ptype := strings.ToLower(string(p.Type))
			if strings.Contains(name, lower) || strings.Contains(path, lower) || strings.Contains(ptype, lower) {
				m.filtered = append(m.filtered, i)
			}
		}
	}
	m.cursor = 0
	m.scrollTop = 0
}

func (m *SelectorModel) ensureVisible() {
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}
	if m.cursor >= m.scrollTop+m.maxVisible {
		m.scrollTop = m.cursor - m.maxVisible + 1
	}
}

func (m *SelectorModel) clampScroll() {
	maxScroll := max(len(m.filtered)-m.maxVisible, 0)
	if m.scrollTop > maxScroll {
		m.scrollTop = maxScroll
	}
	if m.scrollTop < 0 {
		m.scrollTop = 0
	}
}

func (m SelectorModel) View() string {
	if m.done || m.quitting {
		return ""
	}

	var sections []string

	// Title
	sections = append(sections, titleStyle.Render(" CURSPACE  Project Selector "))

	// Search bar
	searchContent := searchLabelStyle.Render("  ") + " "
	if m.search != "" {
		searchContent += searchTextStyle.Render(m.search) + searchCursorStyle.Render("█")
	} else {
		searchContent += helpDescStyle.Render("Type to filter projects...") + searchCursorStyle.Render("█")
	}
	sections = append(sections, searchBarStyle.Render(searchContent))

	// Status line
	selCount := selectedCountStyle.Render(fmt.Sprintf("%d selected", len(m.selected)))
	totCount := totalCountStyle.Render(fmt.Sprintf("%d/%d projects", len(m.filtered), len(m.items)))
	sections = append(sections, statusBarStyle.Render(fmt.Sprintf("%s  •  %s", selCount, totCount)))

	// Project list
	if len(m.filtered) == 0 {
		sections = append(sections, noMatchStyle.Render("No projects match your filter."))
	} else {
		var listLines []string

		// Scroll indicator top
		if m.scrollTop > 0 {
			listLines = append(listLines, scrollHintStyle.Render(fmt.Sprintf("  ▲ %d more above", m.scrollTop)))
		}

		end := min(m.scrollTop+m.maxVisible, len(m.filtered))

		for i := m.scrollTop; i < end; i++ {
			idx := m.filtered[i]
			p := m.items[idx]
			isActive := i == m.cursor
			isSelected := m.selected[idx]

			// Cursor
			cur := "  "
			if isActive {
				cur = cursorIndicator.Render("▸ ")
			}

			// Checkbox
			chk := checkOff.Render("○")
			if isSelected {
				chk = checkOn.Render("●")
			}

			// Project name
			name := projectNameStyle.Render(p.Name)
			if isActive {
				name = projectNameActiveStyle.Render(p.Name)
			}

			// Type tag
			tag := renderProjectTypeTag(p.Type)

			// Path (truncate if needed)
			pth := p.Path
			maxPathLen := 50
			if m.width > 0 {
				maxPathLen = max(20, m.width-len(p.Name)-len(string(p.Type))-20)
			}
			if len(pth) > maxPathLen {
				pth = "…" + pth[len(pth)-maxPathLen+1:]
			}
			pathStr := projectPathStyle.Render(pth)

			line := fmt.Sprintf("%s%s %s %s  %s", cur, chk, name, tag, pathStr)

			if isActive {
				line = activeItemStyle.Render(line)
			}

			listLines = append(listLines, line)
		}

		// Scroll indicator bottom
		remaining := len(m.filtered) - end
		if remaining > 0 {
			listLines = append(listLines, scrollHintStyle.Render(fmt.Sprintf("  ▼ %d more below", remaining)))
		}

		sections = append(sections, strings.Join(listLines, "\n"))
	}

	// Help bar
	helpItems := []struct{ key, desc string }{
		{"↑↓/ctrl+n/p", "navigate"},
		{"space/tab", "toggle"},
		{"enter", "confirm"},
		{"ctrl+a", "select all"},
		{"ctrl+d", "clear all"},
		{"esc", "clear/quit"},
	}
	var helpParts []string
	for _, h := range helpItems {
		helpParts = append(helpParts,
			helpKeyStyle.Render(h.key)+helpDescStyle.Render(" "+h.desc),
		)
	}
	sep := helpSepStyle.Render(" │ ")
	sections = append(sections, helpBarStyle.Render(strings.Join(helpParts, sep)))

	return appStyle.Render(strings.Join(sections, "\n"))
}

func RunSelector(projects []scanner.Project) ([]scanner.Project, error) {
	model := NewSelectorModel(projects)
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running selector: %w", err)
	}

	final := result.(SelectorModel)
	if final.Cancelled() {
		return nil, nil
	}

	return final.SelectedProjects(), nil
}
