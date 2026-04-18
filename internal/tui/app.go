package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GerardoFC8/claumeter/internal/config"
	"github.com/GerardoFC8/claumeter/internal/quota"
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
	tblTurns    table.Model // used only in detail mode

	search        searchState
	themeName     string             // active theme name; kept in sync with currentTheme
	detailMode    bool
	detailSession stats.SessionDetail

	plan        string       // active Claude plan name; "" means unset, no quota UI shown
	quotaStatus quota.Status // recomputed in rebuild()
}

func New(root string) Model {
	return newModelWithTheme(root, "dark", "")
}

// NewWithTheme creates a Model with the given named theme pre-applied.
// Valid names: "dark", "light", "high-contrast". Falls back to "dark".
func NewWithTheme(root, themeName string) Model {
	return newModelWithTheme(root, themeName, "")
}

// NewWithConfig creates a Model with theme and plan pre-applied.
// plan is one of "pro", "max-5x", "max-20x", or "" (no quota UI).
func NewWithConfig(root, themeName, plan string) Model {
	return newModelWithTheme(root, themeName, plan)
}

func newModelWithTheme(root, themeName, plan string) Model {
	applyTheme(themeByName(themeName))
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorAccent)
	return Model{
		root:      root,
		loading:   true,
		filter:    stats.FilterAll,
		spin:      sp,
		search:    newSearchState(),
		themeName: currentTheme.Name,
		plan:      plan,
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
		// Detail mode captures keys separately from normal and search modes.
		if m.detailMode {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				m.detailMode = false
				return m, nil
			case "t":
				m.cycleTheme()
				return m, nil
			default:
				var cmd tea.Cmd
				m.tblTurns, cmd = m.tblTurns.Update(msg)
				return m, cmd
			}
		}

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
		case "enter":
			// Drill into session detail when on the Sessions tab.
			if m.active == tabSessions && !m.loading {
				row := m.tblSess.SelectedRow()
				if len(row) > 0 {
					// Column 0 is the short (8-char) session ID.
					shortID := row[0]
					detail, ok := stats.BuildSessionDetail(m.allData, shortID)
					if ok {
						m.detailSession = detail
						m.detailMode = true
						m.tblTurns = newTurnsTable(detail, m.width)
						m.tblTurns.SetHeight(turnsTableHeight(m.height))
					}
				}
			}
			return m, nil
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
		case "t":
			m.cycleTheme()
			return m, nil
		case "Q":
			m.cyclePlan()
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
	m.quotaStatus = quota.Compute(m.allData, m.plan, time.Now())
	m.buildTablesWithQuery(m.search.query())
	m.resizeTables()
}

// rebuildFiltered rebuilds only the tables using the current search query,
// without re-running the full data pipeline (report stays intact).
func (m *Model) rebuildFiltered() {
	m.buildTablesWithQuery(m.search.query())
	m.resizeTables()
}

// cycleTheme advances to the next theme in allThemes(), persists the choice,
// and refreshes all style-dependent state. Safe to call even when loading.
func (m *Model) cycleTheme() {
	themes := allThemes()
	idx := 0
	for i, t := range themes {
		if t.Name == m.themeName {
			idx = i
			break
		}
	}
	next := themes[(idx+1)%len(themes)]
	applyTheme(next)
	m.themeName = next.Name

	// Recreate spinner and search state so their baked-in colors are refreshed.
	m.spin.Style = lipgloss.NewStyle().Foreground(colorAccent)
	m.search = newSearchState()

	// Persist the new theme; log to stderr on failure but keep it active.
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Defaults()
	}
	cfg.Theme = next.Name
	if err := config.Save(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "claumeter: could not save theme preference:", err)
	}
}

// cyclePlan rotates through: pro -> max-5x -> max-20x -> "" -> pro.
// Persists the choice via config.Save(); logs to stderr on failure.
func (m *Model) cyclePlan() {
	planRotation := []string{"pro", "max-5x", "max-20x", ""}
	idx := 0
	for i, p := range planRotation {
		if p == m.plan {
			idx = i
			break
		}
	}
	m.plan = planRotation[(idx+1)%len(planRotation)]
	m.quotaStatus = quota.Compute(m.allData, m.plan, time.Now())

	cfg, err := config.Load()
	if err != nil {
		cfg = config.Defaults()
	}
	cfg.Plan = m.plan
	if err := config.Save(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "claumeter: could not save plan preference:", err)
	}
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

	if m.detailMode {
		return m.renderDetailView()
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
	quotaTxt := m.renderQuotaBadge()

	left := lipgloss.JoinHorizontal(lipgloss.Left, title, "  ", tabsRow)

	// right side: quota (if configured) + filter badge
	var right string
	if quotaTxt != "" {
		right = quotaTxt + "  " + filterTxt
	} else {
		right = filterTxt
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	pad := m.width - leftWidth - rightWidth - 2
	if pad < 1 {
		pad = 1
	}

	row := left + strings.Repeat(" ", pad) + right
	return headerBarStyle.Width(m.width).Render(row)
}

// renderQuotaBadge returns a compact quota indicator for the header.
// Returns "" when quota is not configured (plan == "").
func (m Model) renderQuotaBadge() string {
	s := m.quotaStatus
	if !s.Configured {
		return ""
	}

	var pctStyle lipgloss.Style
	switch {
	case s.UsedPct >= 80:
		pctStyle = errorStyle
	case s.UsedPct >= 50:
		pctStyle = warnStyle
	default:
		pctStyle = goodStyle
	}

	tag := cardLabelStyle.Render("[" + s.Plan + "]")
	usage := fmt.Sprintf("%d/%d", s.UsedInWindow, s.Limit.MessagesPerWindow)
	pct := pctStyle.Render(fmt.Sprintf("(%.0f%%)", s.UsedPct))

	var resetPart string
	if s.ResetIn > 0 {
		resetPart = cardLabelStyle.Render(" · resets in " + formatQuotaDurationTUI(s.ResetIn))
	}

	return tag + "  " + cardLabelStyle.Render(usage) + " " + pct + resetPart
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
	if m.detailMode {
		keys = "esc=back  t=theme  q=quit"
	} else if m.search.active {
		keys = "enter=apply  esc=clear  ctrl+u=reset"
	} else {
		keys = "tab/h/l switch • 1-5 jump • f/F filter • / search • t=theme • Q=plan • j/k ↑↓ • g/G top/bot • q quit"
	}
	return footerStyle.Width(m.width).Render(keys)
}

// formatQuotaDurationTUI renders a duration compactly for the TUI header.
// Examples: "3h 12m", "45m", "1h".
func formatQuotaDurationTUI(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) - h*60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
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

func turnsTableHeight(windowHeight int) int {
	h := windowHeight - 6
	if h < 5 {
		h = 5
	}
	return h
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
	if m.detailMode {
		m.tblTurns.SetHeight(h)
	}
}

// renderDetailView renders the session drill-down screen.
func (m Model) renderDetailView() string {
	sd := m.detailSession
	dur := sd.LastSeen.Sub(sd.FirstSeen)
	shortID := shortSession(sd.SessionID)

	// Header bar.
	headerLine := fmt.Sprintf(
		"session %s  |  %s  |  %s  |  %s  |  %s tokens",
		accentStyle.Render(shortID),
		cardLabelStyle.Render(shortenPath(sd.Cwd)),
		cardLabelStyle.Render(formatDuration(dur)),
		goodStyle.Render(formatCost(sd.Totals.Cost)),
		cardValueStyle.Render(compactNumber(sd.Totals.GrandTotal())),
	)
	header := headerBarStyle.Width(m.width).Render(headerLine)

	// Body.
	body := sectionStyle.Render(m.tblTurns.View())

	// Footer.
	footer := m.renderFooter()

	bodyHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyHeight < 0 {
		bodyHeight = 0
	}
	bodyBox := lipgloss.NewStyle().Height(bodyHeight).Width(m.width).Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, header, bodyBox, footer)
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
