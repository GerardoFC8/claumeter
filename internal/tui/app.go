package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/stats"
	"github.com/GerardoFC8/claumeter/internal/usage"
)

type tab int

const (
	tabOverview tab = iota
	tabActivity
	tabSessions
	tabProjects
	tabTools
	tabCount
)

var tabLabels = []string{"Overview", "Activity", "Sessions", "Projects", "Tools"}

type loadedMsg struct {
	data usage.Data
	err  error
}

func loadCmd(root string) tea.Cmd {
	return func() tea.Msg {
		data, err := usage.ParseAll(root, nil)
		if err != nil {
			return loadedMsg{err: err}
		}
		return loadedMsg{data: data}
	}
}

type Model struct {
	root    string
	loading bool
	err     error

	allData usage.Data
	filter  stats.FilterPreset
	report  stats.Report

	active tab
	width  int
	height int

	spin        spinner.Model
	tblActivity table.Model
	tblSess     table.Model
	tblProj     table.Model

	search searchState
}

func New(root string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)
	return Model{
		root:    root,
		loading: true,
		filter:  stats.FilterAll,
		spin:    sp,
		search:  newSearchState(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick, loadCmd(m.root))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTables()
		return m, nil

	case tea.KeyMsg:
		// Search mode captures most keys; only ctrl+c quits unconditionally.
		if m.search.active {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				// Leave search mode, keep filter applied.
				m.search.active = false
				m.search.input.Blur()
				m.rebuildFiltered()
				return m, nil
			case "esc":
				// Leave search mode, clear filter.
				m.search.active = false
				m.search.input.SetValue("")
				m.search.input.Blur()
				m.rebuildFiltered()
				return m, nil
			case "ctrl+u":
				// Clear input, stay in search mode.
				m.search.input.SetValue("")
				m.rebuildFiltered()
				var cmd tea.Cmd
				m.search.input, cmd = m.search.input.Update(msg)
				return m, cmd
			default:
				var cmd tea.Cmd
				m.search.input, cmd = m.search.input.Update(msg)
				m.rebuildFiltered()
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "/":
			if !m.loading {
				m.search.active = true
				m.search.input.Focus()
				return m, textinput.Blink
			}
			return m, nil
		case "tab", "right", "l":
			m.active = (m.active + 1) % tabCount
			return m, nil
		case "shift+tab", "left", "h":
			m.active = (m.active - 1 + tabCount) % tabCount
			return m, nil
		case "1":
			m.active = tabOverview
			return m, nil
		case "2":
			m.active = tabActivity
			return m, nil
		case "3":
			m.active = tabSessions
			return m, nil
		case "4":
			m.active = tabProjects
			return m, nil
		case "5":
			m.active = tabTools
			return m, nil
		case "f":
			if !m.loading {
				m.filter = m.filter.Next()
				m.rebuild()
			}
			return m, nil
		case "F":
			if !m.loading {
				m.filter = m.filter.Prev()
				m.rebuild()
			}
			return m, nil
		}

	case loadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.allData = msg.data
			m.rebuild()
		}
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spin, cmd = m.spin.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	var cmd tea.Cmd
	switch m.active {
	case tabActivity:
		m.tblActivity, cmd = m.tblActivity.Update(msg)
	case tabSessions:
		m.tblSess, cmd = m.tblSess.Update(msg)
	case tabProjects:
		m.tblProj, cmd = m.tblProj.Update(msg)
	}
	return m, cmd
}

func (m *Model) rebuild() {
	filtered := m.filter.Apply(m.allData)
	m.report = stats.Build(filtered)
	m.buildTablesWithQuery(m.search.query())
	m.resizeTables()
}

// rebuildFiltered rebuilds only the tables using the current search query,
// without re-running the full data pipeline (report stays intact).
func (m *Model) rebuildFiltered() {
	m.buildTablesWithQuery(m.search.query())
	m.resizeTables()
}

func (m Model) View() string {
	if m.loading {
		return sectionStyle.Render(
			fmt.Sprintf("%s Loading Claude Code usage from %s…", m.spin.View(), m.root),
		)
	}
	if m.err != nil {
		return sectionStyle.Render(warnStyle.Render("Error: ") + m.err.Error())
	}

	header := m.renderHeader()
	body := m.renderBody()
	footer := m.renderFooter()
	searchBar := m.renderSearchBar()

	extraLines := lipgloss.Height(searchBar)
	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer) - extraLines
	if bodyHeight < 0 {
		bodyHeight = 0
	}
	bodyBox := lipgloss.NewStyle().Height(bodyHeight).Width(m.width).Render(body)

	if m.search.active || m.search.query() != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, bodyBox, searchBar, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, bodyBox, footer)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("claude-tui")

	tabs := make([]string, 0, int(tabCount))
	for i, label := range tabLabels {
		style := tabStyle
		if tab(i) == m.active {
			style = tabActiveStyle
		}
		tabs = append(tabs, style.Render(fmt.Sprintf("%d. %s", i+1, label)))
	}
	tabsRow := strings.Join(tabs, "")

	filterTxt := m.renderFilterBadge()

	left := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", tabsRow)
	leftWidth := lipgloss.Width(left)
	filterWidth := lipgloss.Width(filterTxt)
	pad := m.width - leftWidth - filterWidth - 2
	if pad < 1 {
		pad = 1
	}

	row := left + strings.Repeat(" ", pad) + filterTxt
	return headerBarStyle.Width(m.width).Render(row)
}

func (m Model) renderFilterBadge() string {
	label := m.filter.Label()
	if m.filter == stats.FilterAll {
		return cardLabelStyle.Render("filter: ") + accentStyle.Render(label)
	}
	from, to := m.filter.Range(time.Now())
	rng := fmt.Sprintf(" · %s → %s",
		from.Format("2006-01-02"),
		to.AddDate(0, 0, -1).Format("2006-01-02"),
	)
	return cardLabelStyle.Render("filter: ") + accentStyle.Render(label) + cardLabelStyle.Render(rng)
}

func (m Model) renderFooter() string {
	var keys string
	if m.search.active {
		keys = "enter=apply  esc=clear  ctrl+u=reset"
	} else {
		keys = "tab/h/l switch • 1-5 jump • f/F filter • / search • j/k ↑↓ • g/G top/bot • q quit"
	}
	return footerStyle.Width(m.width).Render(keys)
}

func (m Model) renderSearchBar() string {
	if !m.search.active && m.search.query() == "" {
		return ""
	}
	return m.search.renderBar(m.width)
}

func (m Model) renderBody() string {
	switch m.active {
	case tabOverview:
		body := renderOverview(m.report, m.width)
		if m.search.query() != "" {
			note := lipgloss.NewStyle().Foreground(colorMuted).Render(
				"search applies to Activity / Sessions / Projects / Tools",
			)
			return lipgloss.JoinVertical(lipgloss.Left, body, note)
		}
		return body
	case tabActivity:
		return m.renderActivityBody()
	case tabSessions:
		return sectionStyle.Render(m.tblSess.View())
	case tabProjects:
		return sectionStyle.Render(m.tblProj.View())
	case tabTools:
		return renderTools(filterToolStats(m.report.Tools, m.search.query()), m.width, m.height)
	}
	return ""
}

func (m Model) renderActivityBody() string {
	return sectionStyle.Render(m.tblActivity.View())
}

func (m *Model) resizeTables() {
	if m.width == 0 || m.height == 0 {
		return
	}
	h := m.height - 6
	if h < 5 {
		h = 5
	}
	// Activity table shrinks to fit content (filtered or unfiltered).
	// bubbles/table's SetHeight includes the 2-line header — add +2 so the
	// viewport actually shows all data rows.
	const headerLines = 2
	dataRows := len(m.tblActivity.Rows())
	activityH := dataRows + headerLines
	if activityH < headerLines+2 {
		activityH = headerLines + 2
	}
	if activityH > h {
		activityH = h
	}
	m.tblActivity.SetHeight(activityH)
	m.tblSess.SetHeight(h)
	m.tblProj.SetHeight(h)
}

func focusedTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorMuted).
		BorderBottom(true).
		Bold(true).
		Foreground(colorAccent)
	s.Selected = s.Selected.
		Foreground(colorFg).
		Background(colorSelected).
		Bold(true)
	return s
}
