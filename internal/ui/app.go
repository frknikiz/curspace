package ui

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/frknikiz/curspace/internal/config"
	"github.com/frknikiz/curspace/internal/discovery"
	"github.com/frknikiz/curspace/internal/scanner"
	"github.com/frknikiz/curspace/internal/workspace"
)

// ═══════════════════════════════════════════════════════════════════
// Views
// ═══════════════════════════════════════════════════════════════════

type appView int

const (
	viewMain       appView = iota // workspace dashboard
	viewScanning                  // spinner while scanning
	viewSelector                  // project multi-select
	viewOrdering                  // drag-to-reorder selected projects
	viewNaming                    // workspace name input
	viewAddRoot                   // add root directory input
	viewEditorPick                // choose between Cursor and Claude
	viewSettings                  // terminal & default editor preferences
)

// settings option cycles
var (
	settingsTerminalOptions = []settingsOption{
		{value: "", label: "auto", hint: "auto-detect (iTerm if installed/active, else Terminal.app)"},
		{value: "iterm", label: "iterm", hint: "always launch iTerm"},
		{value: "terminal", label: "terminal", hint: "always launch Terminal.app"},
	}
	settingsEditorOptions = []settingsOption{
		{value: "", label: "ask", hint: "always show the editor picker"},
		{value: "cursor", label: "cursor", hint: "skip picker, always open in Cursor"},
		{value: "claude", label: "claude", hint: "skip picker, always open in Claude Code"},
	}
)

type settingsOption struct {
	value string
	label string
	hint  string
}

// ═══════════════════════════════════════════════════════════════════
// Styles
// ═══════════════════════════════════════════════════════════════════

var (
	appTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	appSubtitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Italic(true)

	appSectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	appCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	appSelectedNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true)

	appNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	appDetailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	appTimeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	appEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			PaddingLeft(4)

	appStatusOkStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Bold(true)

	appStatusErrStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true)

	appConfirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD43B")).
			Bold(true)

	appHelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	appHelpDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	appHelpSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3C3C3C"))

	appPadding = lipgloss.NewStyle().Padding(1, 2)

	appMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5"))

	appRootStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C8C8C8"))

	appMutedHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7A7A7A")).
				Italic(true)

	appFolderNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#C8C8C8"))

	appFolderPathStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Italic(true)

	appFolderTreeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4A4A4A"))

	rootSuggestionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8F8F8F"))

	rootSuggestionActiveStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FAFAFA")).
					Bold(true)

	// Selector styles
	selSearchBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	selSearchLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	selSearchTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	selCursorBlinkStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	selCheckOnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	selCheckOffStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))

	selProjectNameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	selProjectActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Bold(true).
				Underline(true)

	selPathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	selCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	selTotalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	selNoMatchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Italic(true).
			PaddingLeft(4)

	selScrollStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	// Input box
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5"))
)

// ═══════════════════════════════════════════════════════════════════
// Messages
// ═══════════════════════════════════════════════════════════════════

type workspacesLoadedMsg struct {
	workspaces []workspace.WorkspaceInfo
	err        error
}

type scanDoneMsg struct {
	scanID int
	result discovery.Result
	err    error
}

// ═══════════════════════════════════════════════════════════════════
// Model
// ═══════════════════════════════════════════════════════════════════

type AppConfig struct {
	Roots         []string
	MaxDepth      int
	Terminal      string
	DefaultEditor string
	OpenCursor    func(string) error
	OpenClaude    func(primaryPath string, extraPaths []string) error
}

type editorPick struct {
	label       string // workspace name or project name shown in status
	primaryPath string // first folder for Claude / workspace file for Cursor
	extraPaths  []string
	cursorPath  string // path passed to OpenCursor (workspace file or project dir)
	cursor      int    // 0 = Cursor, 1 = Claude
}

type scanIntent struct {
	returnView            appView
	forceRefresh          bool
	preserveSearch        string
	preserveSelectedPaths map[string]bool
	directOpen            bool
}

type AppModel struct {
	view          appView
	roots         []string
	maxDepth      int
	terminal      string
	defaultEditor string
	openCursor    func(string) error
	openClaude    func(primaryPath string, extraPaths []string) error

	// editor picker
	editorPick editorPick

	// settings
	settingsCursor   int    // 0 = terminal, 1 = default editor
	settingsTerminal string // pending edit value
	settingsEditor   string // pending edit value

	// terminal
	width  int
	height int

	// main view
	workspaces  []workspace.WorkspaceInfo
	wsCursor    int
	wsExpanded  map[int]bool
	confirming  bool // delete confirmation
	renaming    bool
	renameInput textinput.Model

	// scanning
	spinner      spinner.Model
	scan         scanIntent
	activeScanID int
	nextScanID   int

	// selector
	projects     []scanner.Project
	filtered     []int
	selected     map[int]bool
	selCursor    int
	search       string
	selScrollTop int

	// direct open (single project, no workspace creation)
	directOpen bool

	// ordering
	orderCursor    int
	orderScrollTop int

	// naming
	nameInput        textinput.Model
	selectedProjects []scanner.Project

	// add root
	rootInput textinput.Model

	// feedback
	statusMsg string
	statusErr bool

	lastScanSource   discovery.Source
	lastScanAt       time.Time
	lastProjectCount int

	quitting bool
}

func NewAppModel(cfg AppConfig) AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return AppModel{
		view:          viewMain,
		roots:         cfg.Roots,
		maxDepth:      cfg.MaxDepth,
		terminal:      cfg.Terminal,
		defaultEditor: cfg.DefaultEditor,
		openCursor:    cfg.OpenCursor,
		openClaude:    cfg.OpenClaude,
		spinner:       s,
		selected:      make(map[int]bool),
		wsExpanded:    make(map[int]bool),
	}
}

// ═══════════════════════════════════════════════════════════════════
// Init
// ═══════════════════════════════════════════════════════════════════

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(tea.WindowSize(), loadWorkspacesCmd)
}

func loadWorkspacesCmd() tea.Msg {
	ws, err := workspace.ListDetailed()
	return workspacesLoadedMsg{workspaces: ws, err: err}
}

func doScan(scanID int, roots []string, maxDepth int, forceRefresh bool) tea.Cmd {
	return func() tea.Msg {
		result, err := discovery.Discover(context.Background(), discovery.Options{
			Roots:        roots,
			MaxDepth:     maxDepth,
			ForceRefresh: forceRefresh,
		})
		if err != nil {
			return scanDoneMsg{scanID: scanID, err: err}
		}
		return scanDoneMsg{scanID: scanID, result: result}
	}
}

func (m *AppModel) refreshConfig() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	m.roots = cfg.Roots
	m.maxDepth = cfg.MaxDepth
	m.terminal = cfg.Terminal
	m.defaultEditor = cfg.DefaultEditor
	return nil
}

func (m *AppModel) startProjectSelection(forceRefresh bool) (tea.Model, tea.Cmd) {
	if err := m.refreshConfig(); err != nil {
		m.statusMsg = err.Error()
		m.statusErr = true
		return m, nil
	}
	if len(m.roots) == 0 {
		m.statusMsg = "No project roots configured. Press 'a' to add one."
		m.statusErr = true
		return m, nil
	}

	m.directOpen = false
	m.scan = scanIntent{
		returnView:   viewMain,
		forceRefresh: forceRefresh,
	}
	return m, m.startScan()
}

func (m *AppModel) startDirectOpenSelection(forceRefresh bool) (tea.Model, tea.Cmd) {
	if err := m.refreshConfig(); err != nil {
		m.statusMsg = err.Error()
		m.statusErr = true
		return m, nil
	}
	if len(m.roots) == 0 {
		m.statusMsg = "No project roots configured. Press 'a' to add one."
		m.statusErr = true
		return m, nil
	}

	m.directOpen = true
	m.scan = scanIntent{
		returnView:   viewMain,
		forceRefresh: forceRefresh,
		directOpen:   true,
	}
	return m, m.startScan()
}

func (m *AppModel) startSelectorRescan() tea.Cmd {
	m.scan = scanIntent{
		returnView:            viewSelector,
		forceRefresh:          true,
		preserveSearch:        m.search,
		preserveSelectedPaths: m.selectedProjectPaths(),
	}
	return m.startScan()
}

func (m *AppModel) startScan() tea.Cmd {
	m.nextScanID++
	m.activeScanID = m.nextScanID
	m.view = viewScanning
	return tea.Batch(
		m.spinner.Tick,
		doScan(m.activeScanID, m.roots, m.maxDepth, m.scan.forceRefresh),
	)
}

func (m *AppModel) clearScan() {
	m.scan = scanIntent{}
	m.activeScanID = 0
}

func (m *AppModel) selectedProjectPaths() map[string]bool {
	paths := make(map[string]bool, len(m.selected))
	for idx := range m.selected {
		if idx >= 0 && idx < len(m.projects) {
			paths[m.projects[idx].Path] = true
		}
	}
	return paths
}

func (m *AppModel) restoreSelectorState() {
	if len(m.scan.preserveSelectedPaths) > 0 {
		for i, project := range m.projects {
			if m.scan.preserveSelectedPaths[project.Path] {
				m.selected[i] = true
			}
		}
	}
	if m.scan.preserveSearch != "" {
		m.search = m.scan.preserveSearch
		m.applyProjectFilter()
	}
}

// ═══════════════════════════════════════════════════════════════════
// Update
// ═══════════════════════════════════════════════════════════════════

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case workspacesLoadedMsg:
		if msg.err == nil {
			m.workspaces = msg.workspaces
		}
		if m.wsCursor >= len(m.workspaces) {
			m.wsCursor = max(0, len(m.workspaces)-1)
		}
		return m, nil

	case scanDoneMsg:
		if msg.scanID != m.activeScanID {
			return m, nil
		}

		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Scan failed: %v", msg.err)
			m.statusErr = true
			m.view = m.scan.returnView
			m.clearScan()
			return m, nil
		}

		m.projects = msg.result.Projects
		m.lastScanSource = msg.result.Source
		m.lastScanAt = msg.result.Timestamp
		m.lastProjectCount = len(msg.result.Projects)
		m.directOpen = m.scan.directOpen
		m.initSelector()
		m.restoreSelectorState()
		m.view = viewSelector
		m.clearScan()
		return m, nil
	}

	switch m.view {
	case viewMain:
		return m.updateMain(msg)
	case viewScanning:
		return m.updateScanning(msg)
	case viewSelector:
		return m.updateSelector(msg)
	case viewOrdering:
		return m.updateOrdering(msg)
	case viewNaming:
		return m.updateNaming(msg)
	case viewAddRoot:
		return m.updateAddRoot(msg)
	case viewEditorPick:
		return m.updateEditorPick(msg)
	case viewSettings:
		return m.updateSettings(msg)
	}

	return m, nil
}

// ── Main view ─────────────────────────────────────────────────────

func (m AppModel) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.renaming {
		return m.updateRename(msg)
	}

	if m.confirming {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "y", "Y":
				if m.wsCursor < len(m.workspaces) {
					name := m.workspaces[m.wsCursor].Name
					if err := workspace.Delete(name); err != nil {
						m.statusMsg = err.Error()
						m.statusErr = true
					} else {
						m.statusMsg = fmt.Sprintf("Deleted '%s'", name)
						m.statusErr = false
					}
				}
				m.confirming = false
				return m, loadWorkspacesCmd
			default:
				m.confirming = false
			}
		}
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok {
		m.statusMsg = ""
		switch km.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.wsCursor > 0 {
				m.wsCursor--
			}

		case "down", "j":
			if m.wsCursor < len(m.workspaces)-1 {
				m.wsCursor++
			}

		case "n":
			return m.startProjectSelection(false)

		case "o":
			return m.startDirectOpenSelection(false)

		case "ctrl+r":
			return m.startProjectSelection(true)

		case "enter":
			if len(m.workspaces) > 0 && m.wsCursor < len(m.workspaces) {
				ws := m.workspaces[m.wsCursor]
				path, err := workspace.Open(ws.Name)
				if err != nil {
					m.statusMsg = err.Error()
					m.statusErr = true
					return m, nil
				}
				return m.beginEditorPick(editorPick{
					label:       ws.Name,
					primaryPath: primaryFolderPath(ws.Folders),
					extraPaths:  extraFolderPaths(ws.Folders),
					cursorPath:  path,
				})
			}

		case "d":
			if len(m.workspaces) > 0 {
				m.confirming = true
			}

		case "r":
			if len(m.workspaces) > 0 && m.wsCursor < len(m.workspaces) {
				m.renaming = true
				ti := newStyledInput("new-name")
				ti.SetValue(m.workspaces[m.wsCursor].Name)
				ti.Focus()
				m.renameInput = ti
				return m, textinput.Blink
			}

		case "tab":
			if len(m.workspaces) > 0 && m.wsCursor < len(m.workspaces) {
				if m.wsExpanded[m.wsCursor] {
					delete(m.wsExpanded, m.wsCursor)
				} else {
					m.wsExpanded[m.wsCursor] = true
				}
			}

		case "a":
			ti := newPathInput("~/projects")
			ti.Focus()
			m.rootInput = ti
			m.syncRootSuggestions()
			m.view = viewAddRoot
			return m, textinput.Blink

		case "s":
			if err := m.refreshConfig(); err != nil {
				m.statusMsg = err.Error()
				m.statusErr = true
				return m, nil
			}
			m.settingsTerminal = m.terminal
			m.settingsEditor = m.defaultEditor
			m.settingsCursor = 0
			m.view = viewSettings
			return m, nil
		}
	}

	return m, nil
}

// ── Rename sub-handler ────────────────────────────────────────────

func (m AppModel) updateRename(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyEsc:
			m.renaming = false
			return m, nil
		case tea.KeyEnter:
			oldName := m.workspaces[m.wsCursor].Name
			newName := m.renameInput.Value()
			if newName != "" && newName != oldName {
				if err := workspace.Rename(oldName, newName); err != nil {
					m.statusMsg = err.Error()
					m.statusErr = true
				} else {
					m.statusMsg = fmt.Sprintf("Renamed '%s' → '%s'", oldName, newName)
					m.statusErr = false
				}
			}
			m.renaming = false
			return m, loadWorkspacesCmd
		}
	}
	var cmd tea.Cmd
	m.renameInput, cmd = m.renameInput.Update(msg)
	return m, cmd
}

// ── Scanning view ─────────────────────────────────────────────────

func (m AppModel) updateScanning(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if km.String() == "ctrl+c" || km.String() == "esc" {
			m.view = m.scan.returnView
			m.clearScan()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// ── Selector view ─────────────────────────────────────────────────

func (m AppModel) updateSelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch km.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "ctrl+r":
		return m, m.startSelectorRescan()

	case "esc":
		if m.search != "" {
			m.search = ""
			m.applyProjectFilter()
		} else {
			m.directOpen = false
			m.view = viewMain
		}

	case "enter":
		if m.directOpen {
			if len(m.filtered) > 0 && m.selCursor < len(m.filtered) {
				idx := m.filtered[m.selCursor]
				p := m.projects[idx]
				m.directOpen = false
				return m.beginEditorPick(editorPick{
					label:       p.Name,
					primaryPath: p.Path,
					extraPaths:  nil,
					cursorPath:  p.Path,
				})
			}
		} else if len(m.selected) > 0 {
			m.selectedProjects = nil
			for idx := range m.selected {
				m.selectedProjects = append(m.selectedProjects, m.projects[idx])
			}
			slices.SortFunc(m.selectedProjects, func(a, b scanner.Project) int {
				return cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			})
			m.orderCursor = 0
			m.orderScrollTop = 0
			m.view = viewOrdering
			return m, nil
		}

	case "up":
		if m.selCursor > 0 {
			m.selCursor--
			m.ensureSelVisible()
		}

	case "down":
		if m.selCursor < len(m.filtered)-1 {
			m.selCursor++
			m.ensureSelVisible()
		}

	case " ":
		if len(m.filtered) > 0 && m.selCursor < len(m.filtered) {
			idx := m.filtered[m.selCursor]
			if m.selected[idx] {
				delete(m.selected, idx)
			} else {
				m.selected[idx] = true
			}
		}

	case "tab":
		if len(m.filtered) > 0 && m.selCursor < len(m.filtered) {
			idx := m.filtered[m.selCursor]
			if m.selected[idx] {
				delete(m.selected, idx)
			} else {
				m.selected[idx] = true
			}
			if m.selCursor < len(m.filtered)-1 {
				m.selCursor++
				m.ensureSelVisible()
			}
		}

	case "ctrl+a":
		for _, idx := range m.filtered {
			m.selected[idx] = true
		}

	case "ctrl+d":
		m.selected = make(map[int]bool)

	case "ctrl+n":
		if m.selCursor < len(m.filtered)-1 {
			m.selCursor++
			m.ensureSelVisible()
		}

	case "ctrl+p":
		if m.selCursor > 0 {
			m.selCursor--
			m.ensureSelVisible()
		}

	case "backspace":
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.applyProjectFilter()
		}

	default:
		if km.Type == tea.KeyRunes {
			m.search += string(km.Runes)
			m.applyProjectFilter()
		}
	}

	return m, nil
}

// ── Ordering view ─────────────────────────────────────────────────

func (m AppModel) updateOrdering(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch km.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		m.view = viewSelector
		return m, nil

	case "enter":
		autoName := generateWorkspaceName(m.selectedProjects)
		ti := newStyledInput(autoName)
		ti.Focus()
		m.nameInput = ti
		m.view = viewNaming
		return m, textinput.Blink

	case "up", "k":
		if m.orderCursor > 0 {
			m.orderCursor--
			m.ensureOrderVisible()
		}

	case "down", "j":
		if m.orderCursor < len(m.selectedProjects)-1 {
			m.orderCursor++
			m.ensureOrderVisible()
		}

	case "shift+up", "K":
		if m.orderCursor > 0 {
			m.selectedProjects[m.orderCursor], m.selectedProjects[m.orderCursor-1] =
				m.selectedProjects[m.orderCursor-1], m.selectedProjects[m.orderCursor]
			m.orderCursor--
			m.ensureOrderVisible()
		}

	case "shift+down", "J":
		if m.orderCursor < len(m.selectedProjects)-1 {
			m.selectedProjects[m.orderCursor], m.selectedProjects[m.orderCursor+1] =
				m.selectedProjects[m.orderCursor+1], m.selectedProjects[m.orderCursor]
			m.orderCursor++
			m.ensureOrderVisible()
		}
	}

	return m, nil
}

func (m *AppModel) ensureOrderVisible() {
	maxVis := max(5, m.height-14)
	if m.orderCursor < m.orderScrollTop {
		m.orderScrollTop = m.orderCursor
	}
	if m.orderCursor >= m.orderScrollTop+maxVis {
		m.orderScrollTop = m.orderCursor - maxVis + 1
	}
}

func (m AppModel) renderOrdering() string {
	var s []string

	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("arrange projects"))
	s = append(s, "")
	s = append(s, "  "+ordHintStyle.Render("The first project becomes the primary workspace folder."))
	s = append(s, "  "+ordHintStyle.Render("Use shift+↑/↓ to move items, ↵ to confirm."))
	s = append(s, "")

	maxVis := max(5, m.height-14)
	end := min(m.orderScrollTop+maxVis, len(m.selectedProjects))

	if m.orderScrollTop > 0 {
		s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▲ %d more", m.orderScrollTop)))
	}

	for i := m.orderScrollTop; i < end; i++ {
		p := m.selectedProjects[i]
		isActive := i == m.orderCursor

		cur := "  "
		if isActive {
			cur = appCursorStyle.Render("▸ ")
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

	remaining := len(m.selectedProjects) - end
	if remaining > 0 {
		s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▼ %d more", remaining)))
	}

	s = append(s, "")
	items := []struct{ key, desc string }{
		{"↑↓", "navigate"},
		{"⇧↑↓", "move"},
		{"↵", "confirm"},
		{"esc", "back"},
	}
	s = append(s, "  "+renderHelp(items))

	return appPadding.Render(strings.Join(s, "\n"))
}

// ── Naming view ───────────────────────────────────────────────────

func (m AppModel) updateNaming(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEsc:
			m.view = viewOrdering
			return m, nil
		case tea.KeyEnter:
			name := m.nameInput.Value()
			if name == "" {
				name = generateWorkspaceName(m.selectedProjects)
			}
			folders := make([]workspace.WorkspaceFolder, len(m.selectedProjects))
			for i, p := range m.selectedProjects {
				folders[i] = workspace.WorkspaceFolder{Name: p.Name, Path: p.Path}
			}
			wsPath, err := workspace.Create(name, folders)
			if err != nil {
				m.statusMsg = fmt.Sprintf("Create failed: %v", err)
				m.statusErr = true
				m.view = viewMain
				return m, loadWorkspacesCmd
			}

			model, pickCmd := m.beginEditorPick(editorPick{
				label:       name,
				primaryPath: primaryFolderPath(folders),
				extraPaths:  extraFolderPaths(folders),
				cursorPath:  wsPath,
			})
			return model, tea.Batch(loadWorkspacesCmd, pickCmd)
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

// ── Editor picker view ────────────────────────────────────────────

// beginEditorPick either shows the picker or, when the user has set a default
// editor in settings, runs that editor immediately and skips the view.
func (m AppModel) beginEditorPick(p editorPick) (tea.Model, tea.Cmd) {
	switch m.defaultEditor {
	case editorPickClaude:
		p.cursor = 1
		m.editorPick = p
		return m.runEditorPick()
	case editorPickCursor:
		p.cursor = 0
		m.editorPick = p
		return m.runEditorPick()
	}
	p.cursor = 0
	m.editorPick = p
	m.view = viewEditorPick
	return m, nil
}

const (
	editorPickCursor = "cursor"
	editorPickClaude = "claude"
)

func (m AppModel) updateEditorPick(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.editorPick.cursor > 0 {
			m.editorPick.cursor--
		}
	case "down", "j":
		if m.editorPick.cursor < 1 {
			m.editorPick.cursor++
		}
	case "c", "C":
		m.editorPick.cursor = 0
		return m.runEditorPick()
	case "l", "L":
		m.editorPick.cursor = 1
		return m.runEditorPick()
	case "enter":
		return m.runEditorPick()
	}

	return m, nil
}

func (m AppModel) runEditorPick() (tea.Model, tea.Cmd) {
	pick := m.editorPick
	m.view = viewMain

	if pick.cursor == 1 {
		if m.openClaude == nil {
			m.statusMsg = "Claude launcher is not configured"
			m.statusErr = true
			return m, loadWorkspacesCmd
		}
		if err := m.openClaude(pick.primaryPath, pick.extraPaths); err != nil {
			m.statusMsg = fmt.Sprintf("Claude: %v", err)
			m.statusErr = true
		} else {
			m.statusMsg = fmt.Sprintf("Opened '%s' in Claude", pick.label)
			m.statusErr = false
		}
		return m, loadWorkspacesCmd
	}

	if err := m.openCursor(pick.cursorPath); err != nil {
		m.statusMsg = fmt.Sprintf("Cursor: %v", err)
		m.statusErr = true
	} else {
		m.statusMsg = fmt.Sprintf("Opened '%s' in Cursor", pick.label)
		m.statusErr = false
	}
	return m, loadWorkspacesCmd
}

func (m AppModel) renderEditorPick() string {
	var s []string
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("choose editor"))
	s = append(s, "")
	s = append(s, "  "+appDetailStyle.Render(fmt.Sprintf("Target: %s", m.editorPick.label)))
	s = append(s, "")

	options := []struct {
		name string
		hint string
		key  string
	}{
		{"Cursor", "open as multi-root .code-workspace", "c"},
		{"Claude Code", "claude in primary folder + --add-dir for extras", "l"},
	}

	for i, opt := range options {
		isActive := i == m.editorPick.cursor
		cur := "  "
		if isActive {
			cur = appCursorStyle.Render("▸ ")
		}
		name := appNameStyle.Render(opt.name)
		if isActive {
			name = appSelectedNameStyle.Render(opt.name)
		}
		hotkey := appHelpKeyStyle.Render(fmt.Sprintf("[%s]", opt.key))
		hint := appDetailStyle.Render(opt.hint)
		s = append(s, fmt.Sprintf("  %s%s %s  %s", cur, hotkey, name, hint))
	}

	s = append(s, "")
	items := []struct{ key, desc string }{
		{"↑↓", "navigate"},
		{"↵", "open"},
		{"c", "Cursor"},
		{"l", "Claude"},
		{"esc", "back"},
	}
	s = append(s, "  "+renderHelp(items))
	return appPadding.Render(strings.Join(s, "\n"))
}

// ── Settings view ─────────────────────────────────────────────────

func (m AppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m, nil
	case "up", "k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "down", "j":
		if m.settingsCursor < 1 {
			m.settingsCursor++
		}
	case "left", "h":
		m.cycleSettingsValue(-1)
	case "right", "l", "tab", " ":
		m.cycleSettingsValue(1)
	case "enter":
		cfg, err := config.Load()
		if err != nil {
			m.statusMsg = err.Error()
			m.statusErr = true
			m.view = viewMain
			return m, nil
		}
		cfg.Terminal = m.settingsTerminal
		cfg.DefaultEditor = m.settingsEditor
		if err := config.Save(cfg); err != nil {
			m.statusMsg = err.Error()
			m.statusErr = true
		} else {
			m.terminal = m.settingsTerminal
			m.defaultEditor = m.settingsEditor
			m.statusMsg = "Settings saved"
			m.statusErr = false
		}
		m.view = viewMain
		return m, nil
	}

	return m, nil
}

func (m *AppModel) cycleSettingsValue(delta int) {
	if m.settingsCursor == 0 {
		m.settingsTerminal = cycleSettingsOption(settingsTerminalOptions, m.settingsTerminal, delta)
	} else {
		m.settingsEditor = cycleSettingsOption(settingsEditorOptions, m.settingsEditor, delta)
	}
}

func cycleSettingsOption(opts []settingsOption, current string, delta int) string {
	idx := 0
	for i, opt := range opts {
		if opt.value == current {
			idx = i
			break
		}
	}
	next := (idx + delta + len(opts)) % len(opts)
	return opts[next].value
}

func settingsLabelFor(opts []settingsOption, value string) string {
	for _, opt := range opts {
		if opt.value == value {
			return opt.label
		}
	}
	return opts[0].label
}

func settingsHintFor(opts []settingsOption, value string) string {
	for _, opt := range opts {
		if opt.value == value {
			return opt.hint
		}
	}
	return opts[0].hint
}

func (m AppModel) renderSettings() string {
	var s []string
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("settings"))
	s = append(s, "")
	s = append(s, "  "+appMutedHintStyle.Render("Persisted to ~/.curspace/config.json. Use ←/→ to change a value, ↵ to save."))
	s = append(s, "")

	rows := []struct {
		title   string
		options []settingsOption
		value   string
	}{
		{"Terminal", settingsTerminalOptions, m.settingsTerminal},
		{"Default editor", settingsEditorOptions, m.settingsEditor},
	}

	for i, row := range rows {
		isActive := i == m.settingsCursor
		cur := "  "
		if isActive {
			cur = appCursorStyle.Render("▸ ")
		}

		title := appNameStyle.Render(row.title)
		if isActive {
			title = appSelectedNameStyle.Render(row.title)
		}

		valueLabel := settingsLabelFor(row.options, row.value)
		hint := settingsHintFor(row.options, row.value)
		valueStyled := appHelpKeyStyle.Render("[ " + valueLabel + " ]")
		s = append(s, fmt.Sprintf("  %s%-18s  %s", cur, title, valueStyled))
		s = append(s, "      "+appDetailStyle.Render(hint))
		s = append(s, "")
	}

	items := []struct{ key, desc string }{
		{"↑↓", "row"},
		{"←→", "value"},
		{"↵", "save"},
		{"esc", "cancel"},
	}
	s = append(s, "  "+renderHelp(items))
	return appPadding.Render(strings.Join(s, "\n"))
}

func primaryFolderPath(folders []workspace.WorkspaceFolder) string {
	if len(folders) == 0 {
		return ""
	}
	return folders[0].Path
}

func extraFolderPaths(folders []workspace.WorkspaceFolder) []string {
	if len(folders) <= 1 {
		return nil
	}
	out := make([]string, 0, len(folders)-1)
	for _, f := range folders[1:] {
		out = append(out, f.Path)
	}
	return out
}

// ── Add root view ─────────────────────────────────────────────────

func (m AppModel) updateAddRoot(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit
		case tea.KeyEsc:
			m.view = viewMain
			return m, nil
		case tea.KeyEnter:
			path := m.rootInput.Value()
			if path != "" {
				if err := config.AddRoot(path); err != nil {
					m.statusMsg = err.Error()
					m.statusErr = true
				} else {
					normalized, _ := config.NormalizePath(path)
					m.roots = append(m.roots, normalized)
					m.statusMsg = fmt.Sprintf("Added root: %s. Press n to browse projects or ctrl+r to rescan.", normalized)
					m.statusErr = false
				}
				m.view = viewMain
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.rootInput, cmd = m.rootInput.Update(msg)
	m.syncRootSuggestions()
	return m, cmd
}

// ═══════════════════════════════════════════════════════════════════
// View
// ═══════════════════════════════════════════════════════════════════

func (m AppModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.view {
	case viewMain:
		return m.renderMain()
	case viewScanning:
		return m.renderScanning()
	case viewSelector:
		return m.renderSelector()
	case viewOrdering:
		return m.renderOrdering()
	case viewNaming:
		return m.renderNaming()
	case viewAddRoot:
		return m.renderAddRoot()
	case viewEditorPick:
		return m.renderEditorPick()
	case viewSettings:
		return m.renderSettings()
	}

	return ""
}

// ── Main view ─────────────────────────────────────────────────────

func (m AppModel) renderMain() string {
	var s []string

	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("workspace hub"))
	s = append(s, "")
	s = append(s, "  "+m.scanSummaryText())
	s = append(s, "  "+appMutedHintStyle.Render("n opens the latest catalog. ctrl+r forces a fresh rescan."))
	s = append(s, "")

	s = append(s, appSectionStyle.Render(fmt.Sprintf("  Workspaces (%d)", len(m.workspaces))))

	if len(m.workspaces) == 0 {
		s = append(s, appEmptyStyle.Render("No workspaces yet. Press n to create one, or o to open a single project."))
	} else {
		maxVisible := max(5, m.height-18)
		scrollTop := 0
		if m.wsCursor >= maxVisible {
			scrollTop = m.wsCursor - maxVisible + 1
		}
		end := min(scrollTop+maxVisible, len(m.workspaces))

		if scrollTop > 0 {
			s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▲ %d more", scrollTop)))
		}

		for i := scrollTop; i < end; i++ {
			ws := m.workspaces[i]
			isActive := i == m.wsCursor
			isExpanded := m.wsExpanded[i]

			cursor := "  "
			if isActive {
				cursor = appCursorStyle.Render("▸ ")
			}

			expandIcon := appFolderTreeStyle.Render("▶")
			if isExpanded {
				expandIcon = appFolderTreeStyle.Render("▼")
			}

			name := appNameStyle.Render(ws.Name)
			if isActive {
				name = appSelectedNameStyle.Render(ws.Name)
			}

			detail := appDetailStyle.Render(fmt.Sprintf("%d folders", ws.FolderCount))
			ago := appTimeStyle.Render(timeAgo(ws.ModTime))

			s = append(s, fmt.Sprintf("  %s%s %s  %s  %s", cursor, expandIcon, name, detail, ago))

			if isExpanded && len(ws.Folders) > 0 {
				for fi, folder := range ws.Folders {
					tree := appFolderTreeStyle.Render("├─")
					if fi == len(ws.Folders)-1 {
						tree = appFolderTreeStyle.Render("└─")
					}
					fName := folder.Name
					if fName == "" {
						fName = filepath.Base(folder.Path)
					}
					maxLen := max(20, m.width-len(fName)-20)
					fPath := appFolderPathStyle.Render(truncatePath(folder.Path, maxLen))
					s = append(s, fmt.Sprintf("       %s %s  %s", tree, appFolderNameStyle.Render(fName), fPath))
				}
			}
		}

		remaining := len(m.workspaces) - end
		if remaining > 0 {
			s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▼ %d more", remaining)))
		}
	}

	s = append(s, "")
	s = append(s, appSectionStyle.Render(fmt.Sprintf("  Project Roots (%d)", len(m.roots))))
	if len(m.roots) == 0 {
		s = append(s, appEmptyStyle.Render("No roots configured. Press a to add the first scan path."))
	} else {
		s = append(s, m.renderRootPreview(3)...)
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

	if m.confirming && m.wsCursor < len(m.workspaces) {
		s = append(s, "")
		s = append(s, "  "+appConfirmStyle.Render(
			fmt.Sprintf("Delete '%s'? (y/n)", m.workspaces[m.wsCursor].Name),
		))
	}

	if m.renaming {
		s = append(s, "")
		s = append(s, "  "+inputLabelStyle.Render("Rename to:")+"  "+m.renameInput.View())
	}

	s = append(s, "")
	s = append(s, m.renderMainHelp())

	return appPadding.Render(strings.Join(s, "\n"))
}

func (m AppModel) renderMainHelp() string {
	items := []struct{ key, desc string }{
		{"n", "new workspace"},
		{"o", "open project"},
		{"tab", "expand"},
		{"ctrl+r", "rescan"},
		{"↵", "open"},
		{"d", "delete"},
		{"r", "rename"},
		{"a", "add root"},
		{"s", "settings"},
		{"q", "quit"},
	}
	return "  " + renderHelp(items)
}

// ── Scanning view ─────────────────────────────────────────────────

func (m AppModel) renderScanning() string {
	var s []string
	subtitle := "prepare project catalog"
	if m.scan.forceRefresh {
		subtitle = "fresh project rescan"
	}
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render(subtitle))
	s = append(s, "")
	s = append(s, "")
	msgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A49FA5"))
	action := "Loading project catalog..."
	if m.scan.forceRefresh {
		action = "Rescanning projects from disk..."
	}
	s = append(s, "  "+m.spinner.View()+" "+msgStyle.Render(action))
	if len(m.roots) > 0 {
		s = append(s, "")
		s = append(s, "  "+appMetaStyle.Render(fmt.Sprintf("Scanning %d root(s)", len(m.roots))))
	}
	s = append(s, "")
	s = append(s, "  "+appHelpDescStyle.Render("esc cancel"))
	return appPadding.Render(strings.Join(s, "\n"))
}

// ── Selector view ─────────────────────────────────────────────────

func (m AppModel) renderSelector() string {
	var s []string

	subtitle := "select projects"
	if m.directOpen {
		subtitle = "open project"
	}
	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render(subtitle))
	s = append(s, "")
	s = append(s, "  "+m.scanSummaryText())
	s = append(s, "")

	searchContent := selSearchLabelStyle.Render("Filter") + "  "
	if m.search != "" {
		searchContent += selSearchTextStyle.Render(m.search) + selCursorBlinkStyle.Render("█")
	} else {
		searchContent += appHelpDescStyle.Render("Type a name, path, or type...") + selCursorBlinkStyle.Render("█")
	}
	s = append(s, selSearchBoxStyle.Render(searchContent))

	if m.directOpen {
		totCount := selTotalStyle.Render(fmt.Sprintf("%d/%d projects", len(m.filtered), len(m.projects)))
		s = append(s, "  "+totCount)
	} else {
		selCount := selCountStyle.Render(fmt.Sprintf("%d selected", len(m.selected)))
		totCount := selTotalStyle.Render(fmt.Sprintf("%d/%d projects", len(m.filtered), len(m.projects)))
		s = append(s, "  "+selCount+"  •  "+totCount)
	}
	s = append(s, "")

	if len(m.filtered) == 0 {
		switch {
		case len(m.projects) == 0:
			s = append(s, selNoMatchStyle.Render("No projects discovered yet. Press ctrl+r to rescan or esc to go back."))
		case m.search != "":
			s = append(s, selNoMatchStyle.Render(fmt.Sprintf("No projects match %q. Press esc to clear the filter.", m.search)))
		default:
			s = append(s, selNoMatchStyle.Render("No projects available in the current catalog."))
		}
	} else {
		maxVis := max(5, m.height-17)
		end := min(m.selScrollTop+maxVis, len(m.filtered))

		if m.selScrollTop > 0 {
			s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▲ %d more", m.selScrollTop)))
		}

		for i := m.selScrollTop; i < end; i++ {
			idx := m.filtered[i]
			p := m.projects[idx]
			isActive := i == m.selCursor

			cur := "  "
			if isActive {
				cur = appCursorStyle.Render("▸ ")
			}

			name := selProjectNameStyle.Render(p.Name)
			if isActive {
				name = selProjectActiveStyle.Render(p.Name)
			}

			tag := renderProjectTypeTag(p.Type)

			maxLen := 50
			if m.width > 0 {
				maxLen = max(20, m.width-len(p.Name)-len(string(p.Type))-28)
			}
			pathStr := selPathStyle.Render(truncatePath(p.Path, maxLen))

			if m.directOpen {
				s = append(s, fmt.Sprintf("  %s%s %s  %s", cur, name, tag, pathStr))
			} else {
				isSelected := m.selected[idx]
				chk := selCheckOffStyle.Render("○")
				if isSelected {
					chk = selCheckOnStyle.Render("●")
				}
				s = append(s, fmt.Sprintf("  %s%s %s %s  %s", cur, chk, name, tag, pathStr))
			}
		}

		remaining := len(m.filtered) - end
		if remaining > 0 {
			s = append(s, selScrollStyle.Render(fmt.Sprintf("    ▼ %d more", remaining)))
		}
	}

	s = append(s, "")
	var items []struct{ key, desc string }
	if m.directOpen {
		items = []struct{ key, desc string }{
			{"↑↓", "navigate"},
			{"↵", "choose editor"},
			{"ctrl+r", "rescan"},
			{"esc", "back"},
		}
	} else {
		items = []struct{ key, desc string }{
			{"↑↓", "navigate"},
			{"space/tab", "toggle"},
			{"↵", "continue"},
			{"ctrl+r", "rescan"},
			{"ctrl+a", "all"},
			{"ctrl+d", "clear"},
			{"esc", "back"},
		}
	}
	s = append(s, "  "+renderHelp(items))

	return appPadding.Render(strings.Join(s, "\n"))
}

// ── Naming view ───────────────────────────────────────────────────

func (m AppModel) renderNaming() string {
	var s []string

	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("workspace name"))
	s = append(s, "")

	label := inputLabelStyle.Render("Enter a name for your workspace:")
	input := m.nameInput.View()
	box := inputBoxStyle.Render(fmt.Sprintf("%s\n\n%s", label, input))
	s = append(s, box)

	s = append(s, "")
	selInfo := appDetailStyle.Render(fmt.Sprintf("%d project(s) selected", len(m.selectedProjects)))
	s = append(s, "  "+selInfo)
	s = append(s, "  "+appMutedHintStyle.Render("Press enter without a name to use the auto-generated one."))

	s = append(s, "")
	items := []struct{ key, desc string }{
		{"↵", "create & choose editor"},
		{"esc", "back"},
	}
	s = append(s, "  "+renderHelp(items))

	return appPadding.Render(strings.Join(s, "\n"))
}

// ── Add root view ─────────────────────────────────────────────────

func (m AppModel) renderAddRoot() string {
	var s []string

	s = append(s, appTitleStyle.Render(" CURSPACE ")+"  "+appSubtitleStyle.Render("add project root"))
	s = append(s, "")

	label := inputLabelStyle.Render("Enter a directory path to scan for projects:")
	input := m.rootInput.View()
	box := inputBoxStyle.Render(fmt.Sprintf("%s\n\n%s", label, input))
	s = append(s, box)

	if matches := m.rootInput.MatchedSuggestions(); len(matches) > 0 {
		s = append(s, "")
		s = append(s, "  "+appMetaStyle.Render("Autocomplete"))
		for i, suggestion := range matches[:min(len(matches), 4)] {
			style := rootSuggestionStyle
			prefix := "    "
			if i == m.rootInput.CurrentSuggestionIndex() {
				style = rootSuggestionActiveStyle
				prefix = "  ▸ "
			}
			s = append(s, prefix+style.Render(suggestion))
		}
		if extra := len(matches) - 4; extra > 0 {
			s = append(s, "    "+appMutedHintStyle.Render(fmt.Sprintf("+ %d more match(es)", extra)))
		}
	}

	s = append(s, "")
	s = append(s, "  "+appMutedHintStyle.Render("Examples: ~/projects, ~/work, /Volumes/code"))
	s = append(s, "")
	items := []struct{ key, desc string }{
		{"tab", "complete"},
		{"↑↓", "suggestions"},
		{"↵", "add"},
		{"esc", "cancel"},
	}
	s = append(s, "  "+renderHelp(items))

	return appPadding.Render(strings.Join(s, "\n"))
}

// ═══════════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════════

func (m AppModel) scanSummaryText() string {
	if len(m.roots) == 0 {
		return appStatusErrStyle.Render("No scan roots configured")
	}
	if m.lastScanAt.IsZero() {
		return appMetaStyle.Render(fmt.Sprintf("%d root(s) configured. No scan yet.", len(m.roots)))
	}

	sourceText := "Fresh scan"
	if m.lastScanSource == discovery.SourceCache {
		sourceText = "Cached catalog"
	}

	return strings.Join([]string{
		appStatusOkStyle.Render(fmt.Sprintf("%d projects", m.lastProjectCount)),
		appMetaStyle.Render(fmt.Sprintf("from %s", sourceText)),
		appTimeStyle.Render(timeAgo(m.lastScanAt)),
	}, "  •  ")
}

func (m AppModel) renderRootPreview(limit int) []string {
	var lines []string
	for i, root := range m.roots {
		if i >= limit {
			break
		}
		lines = append(lines, "    "+appRootStyle.Render(truncatePath(root, max(24, m.width-10))))
	}
	if extra := len(m.roots) - limit; extra > 0 {
		lines = append(lines, "    "+appMutedHintStyle.Render(fmt.Sprintf("+ %d more root(s)", extra)))
	}
	return lines
}

func (m *AppModel) initSelector() {
	m.filtered = make([]int, len(m.projects))
	for i := range m.projects {
		m.filtered[i] = i
	}
	m.selected = make(map[int]bool)
	m.selCursor = 0
	m.selScrollTop = 0
	m.search = ""
}

func (m *AppModel) applyProjectFilter() {
	if m.search == "" {
		m.filtered = make([]int, len(m.projects))
		for i := range m.projects {
			m.filtered[i] = i
		}
	} else {
		m.filtered = nil
		lower := strings.ToLower(m.search)
		for i, p := range m.projects {
			if strings.Contains(strings.ToLower(p.Name), lower) ||
				strings.Contains(strings.ToLower(p.Path), lower) ||
				strings.Contains(strings.ToLower(string(p.Type)), lower) {
				m.filtered = append(m.filtered, i)
			}
		}
	}
	m.selCursor = 0
	m.selScrollTop = 0
}

func (m *AppModel) ensureSelVisible() {
	maxVis := max(5, m.height-14)
	if m.selCursor < m.selScrollTop {
		m.selScrollTop = m.selCursor
	}
	if m.selCursor >= m.selScrollTop+maxVis {
		m.selScrollTop = m.selCursor - maxVis + 1
	}
}

func (m *AppModel) syncRootSuggestions() {
	m.rootInput.SetSuggestions(pathSuggestions(m.rootInput.Value()))
}

func generateWorkspaceName(projects []scanner.Project) string {
	if len(projects) == 0 {
		return fmt.Sprintf("workspace-%d", time.Now().Unix())
	}

	sanitize := func(name string) string {
		name = strings.ToLower(name)
		name = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return '-'
		}, name)
		name = strings.Trim(name, "-")
		return name
	}

	switch len(projects) {
	case 1:
		return sanitize(projects[0].Name)
	case 2:
		return sanitize(projects[0].Name) + "-" + sanitize(projects[1].Name)
	default:
		return fmt.Sprintf("%s-%s-and-%d-more",
			sanitize(projects[0].Name),
			sanitize(projects[1].Name),
			len(projects)-2,
		)
	}
}

func newStyledInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 100
	ti.Width = 40
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Italic(true)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	ti.Prompt = "❯ "
	return ti
}

func newPathInput(placeholder string) textinput.Model {
	ti := newStyledInput(placeholder)
	ti.CharLimit = 512
	ti.Width = 60
	ti.ShowSuggestions = true
	ti.CompletionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7A7A7A")).
		Italic(true)
	return ti
}

func pathSuggestions(input string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	absDir, displayDir, prefix, err := splitPathQuery(input)
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil
	}

	includeHidden := strings.HasPrefix(prefix, ".")
	prefixLower := strings.ToLower(prefix)
	suggestions := make([]string, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()
		if !includeHidden && strings.HasPrefix(name, ".") {
			continue
		}
		if prefix != "" && !strings.HasPrefix(strings.ToLower(name), prefixLower) {
			continue
		}
		if !entry.IsDir() {
			info, err := os.Stat(filepath.Join(absDir, name))
			if err != nil || !info.IsDir() {
				continue
			}
		}

		suggestion := joinPathSuggestion(displayDir, name) + string(filepath.Separator)
		suggestions = append(suggestions, suggestion)
	}

	slices.Sort(suggestions)
	return suggestions
}

func splitPathQuery(input string) (absDir, displayDir, prefix string, err error) {
	if input == "~" {
		absDir, err = expandUserPath(input)
		return absDir, "~", "", err
	}

	expanded, err := expandUserPath(input)
	if err != nil {
		return "", "", "", err
	}

	if strings.HasSuffix(input, string(filepath.Separator)) {
		absDir, err = filepath.Abs(expanded)
		displayDir = strings.TrimSuffix(input, string(filepath.Separator))
		if displayDir == "" {
			displayDir = string(filepath.Separator)
		}
		return absDir, displayDir, "", err
	}

	expandedDir := filepath.Dir(expanded)
	absDir, err = filepath.Abs(expandedDir)
	if err != nil {
		return "", "", "", err
	}

	displayDir = filepath.Dir(input)
	if displayDir == "." && !strings.Contains(input, string(filepath.Separator)) {
		displayDir = ""
	}

	return absDir, displayDir, filepath.Base(input), nil
}

func expandUserPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return home, nil
	}

	return filepath.Join(home, path[1:]), nil
}

func joinPathSuggestion(base, name string) string {
	if base == "" {
		return name
	}
	if base == "." {
		return "." + string(filepath.Separator) + name
	}
	if base == "~" {
		return "~" + string(filepath.Separator) + name
	}
	return filepath.Join(base, name)
}

func truncatePath(path string, maxLen int) string {
	if maxLen <= 0 || len(path) <= maxLen {
		return path
	}
	if maxLen <= 1 {
		return "…"
	}
	return "…" + path[len(path)-maxLen+1:]
}

func renderHelp(items []struct{ key, desc string }) string {
	sep := appHelpSepStyle.Render(" │ ")
	var parts []string
	for _, h := range items {
		parts = append(parts,
			appHelpKeyStyle.Render(h.key)+" "+appHelpDescStyle.Render(h.desc),
		)
	}
	return strings.Join(parts, sep)
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}

// ═══════════════════════════════════════════════════════════════════
// Public API
// ═══════════════════════════════════════════════════════════════════

func RunApp(cfg AppConfig) error {
	m := NewAppModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
