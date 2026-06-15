// Package ui is the VertexSense bubbletea dashboard: one screen that fuses every
// sensor's feed, with per-sensor health, severity filtering, pause and scroll.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/collector"
	"github.com/VertexElite/vertexos/vertexsense/internal/event"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const sidebarWidth = 30

// evMsg delivers one collector event into the bubbletea loop.
type evMsg event.Event

// tickMsg drives the clock + relative "last seen" timers.
type tickMsg time.Time

// Model is the dashboard state.
type Model struct {
	width, height int
	ready         bool

	collectors []collector.Collector
	evCh       <-chan event.Event

	events []event.Event
	cap    int
	counts map[event.Severity]int

	vp       viewport.Model
	paused   bool
	minSev   event.Severity
	srcFltr  string // "" = all
	start    time.Time
	now      time.Time
}

// New builds a dashboard model. evCh may be nil (e.g. snapshot rendering).
func New(cols []collector.Collector, evCh <-chan event.Event, capacity int) Model {
	if capacity <= 0 {
		capacity = 2000
	}
	return Model{
		collectors: cols,
		evCh:       evCh,
		cap:        capacity,
		counts:     map[event.Severity]int{},
		start:      time.Now(),
		now:        time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd()}
	if m.evCh != nil {
		cmds = append(cmds, waitForEvent(m.evCh))
	}
	return tea.Batch(cmds...)
}

func waitForEvent(ch <-chan event.Event) tea.Cmd {
	return func() tea.Msg { return evMsg(<-ch) }
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		m.ready = true
		m.refresh()
		return m, nil

	case tickMsg:
		m.now = time.Time(msg)
		return m, tickCmd()

	case evMsg:
		m.ingest(event.Event(msg))
		return m, waitForEvent(m.evCh)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) ingest(ev event.Event) {
	m.events = append(m.events, ev)
	if len(m.events) > m.cap {
		m.events = m.events[len(m.events)-m.cap:]
	}
	m.counts[ev.Severity]++
	if !m.paused {
		m.refresh()
	}
}

func (m Model) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	case "p":
		m.paused = !m.paused
		if !m.paused {
			m.refresh()
		}
		return m, nil
	case "c":
		m.events = nil
		m.counts = map[event.Severity]int{}
		m.refresh()
		return m, nil
	case "s":
		m.srcFltr = m.cycleSource()
		m.refresh()
		return m, nil
	case "0", "1", "2", "3", "4":
		m.minSev = event.Severity(int(k.String()[0] - '0'))
		m.refresh()
		return m, nil
	case "g", "home":
		m.vp.GotoTop()
		return m, nil
	case "G", "end":
		m.vp.GotoBottom()
		return m, nil
	}
	// Everything else (↑/↓, k/j, pgup/pgdn) drives the feed viewport.
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(k)
	return m, cmd
}

// cycleSource rotates the source filter through all → each collector → all.
func (m Model) cycleSource() string {
	order := append([]string{""}, names(m.collectors)...)
	for i, n := range order {
		if n == m.srcFltr {
			return order[(i+1)%len(order)]
		}
	}
	return ""
}

func (m *Model) layout() {
	bodyH := m.height - 3 /*header*/ - 2 /*footer*/
	if bodyH < 3 {
		bodyH = 3
	}
	feedW := m.width - sidebarWidth - 2
	if feedW < 20 {
		feedW = 20
	}
	if m.vp.Width == 0 {
		m.vp = viewport.New(feedW, bodyH-1)
	} else {
		m.vp.Width, m.vp.Height = feedW, bodyH-1
	}
}

// refresh re-renders the event feed into the viewport.
func (m *Model) refresh() {
	if !m.ready {
		return
	}
	atBottom := m.vp.AtBottom()
	var b strings.Builder
	for _, ev := range m.events {
		if ev.Severity < m.minSev {
			continue
		}
		if m.srcFltr != "" && ev.Source != m.srcFltr {
			continue
		}
		b.WriteString(m.formatEvent(ev))
		b.WriteByte('\n')
	}
	m.vp.SetContent(strings.TrimRight(b.String(), "\n"))
	if atBottom && !m.paused {
		m.vp.GotoBottom()
	}
}

func (m Model) formatEvent(ev event.Event) string {
	ts := faintStyle.Render(ev.Time.Format("15:04:05"))
	head := fmt.Sprintf("%s %s %s ", ts, sevBadge(ev.Severity), sourceTag(ev.Source))
	rule := ""
	if ev.Rule != "" {
		rule = blueStyle.Render(ev.Rule) + "  "
	}
	avail := m.vp.Width - lipgloss.Width(head) - lipgloss.Width(rule)
	msg := textStyle.Render(truncate(ev.Message, avail))
	return head + rule + msg
}

func (m Model) View() string {
	if !m.ready {
		return "starting VertexSense…"
	}
	header := m.renderHeader()
	footer := m.renderFooter()
	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if bodyH < 1 {
		bodyH = 1
	}
	sidebar := panelBox.Width(sidebarWidth).Height(bodyH).Render(m.renderSidebar())
	feed := lipgloss.NewStyle().Width(m.width - sidebarWidth - 2).Height(bodyH).Render(
		panelTitle.Render("EVENT FEED") + "\n" + m.vp.View())
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, feed)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) renderHeader() string {
	left := titleStyle.Render("◆ VertexSense") + subtleStyle.Render("  ·  VertexOS security telemetry")
	clock := subtleStyle.Render(m.now.Format("15:04:05"))
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(clock) - 2
	line1 := left + spaces(max(1, gap)) + clock

	var badges []string
	for s := event.Critical; s >= event.Info; s-- {
		badges = append(badges, fmt.Sprintf("%s %d", sevBadge(s), m.counts[s]))
	}
	state := faintStyle.Render(fmt.Sprintf("total %d", len(m.events)))
	if m.paused {
		state = lipgloss.NewStyle().Foreground(cAccent).Render("⏸ PAUSED")
	}
	line2 := strings.Join(badges, faintStyle.Render("  ")) + faintStyle.Render("   ·   ") + state
	return headerBar.Width(m.width - 2).Render(line1 + "\n" + line2)
}

func (m Model) renderSidebar() string {
	var b strings.Builder
	b.WriteString(panelTitle.Render("SENSORS") + "\n\n")
	for _, c := range m.collectors {
		st := c.Status()
		online := st.Health == collector.Online
		name := textStyle.Render(padRight(st.Name, 9))
		cnt := faintStyle.Render(fmt.Sprintf("%4d", st.Count))
		b.WriteString(fmt.Sprintf("%s %s %s\n", dot(online), name, cnt))
		last := "—"
		if !st.Last.IsZero() {
			last = relTime(m.now, st.Last)
		}
		b.WriteString(faintStyle.Render("    "+st.Health.String()+" · "+last) + "\n\n")
	}
	up := relDuration(m.now.Sub(m.start))
	b.WriteString(faintStyle.Render("uptime "+up) + "\n")
	if m.minSev > event.Info || m.srcFltr != "" {
		f := "filter ≥" + m.minSev.String()
		if m.srcFltr != "" {
			f += " src=" + m.srcFltr
		}
		b.WriteString(lipgloss.NewStyle().Foreground(cAccent).Render(f) + "\n")
	}
	return b.String()
}

func (m Model) renderFooter() string {
	keys := "↑/↓ scroll · g/G top/bottom · p pause · c clear · 0-4 min-sev · s source · q quit"
	return footerStyle.Width(m.width - 2).Render(keys)
}

// Snapshot renders one static frame at the given size from a fixed event slice.
// Used by `--snapshot` and tests so the layout can render without a TTY.
func Snapshot(cols []collector.Collector, events []event.Event, w, h int) string {
	m := New(cols, nil, len(events)+1)
	m.width, m.height = w, h
	m.layout()
	m.ready = true
	m.now = time.Now()
	for _, ev := range events {
		m.events = append(m.events, ev)
		m.counts[ev.Severity]++
	}
	m.refresh()
	m.vp.GotoBottom()
	return m.View()
}

func names(cs []collector.Collector) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.Name()
	}
	return out
}

func relTime(now, t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return relDuration(now.Sub(t)) + " ago"
}

func relDuration(d time.Duration) string {
	switch {
	case d < time.Second:
		return "now"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
