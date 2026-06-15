package ui

import (
	"github.com/VertexElite/vertexos/vertexsense/internal/event"
	"github.com/charmbracelet/lipgloss"
)

// VertexOS Aero palette — cyan accent over deep slate, matching the brand.
var (
	cAccent = lipgloss.Color("#2dd4ff")
	cBlue   = lipgloss.Color("#5b8cff")
	cText   = lipgloss.Color("#e8ecf1")
	cMuted  = lipgloss.Color("#8a93a0")
	cFaint  = lipgloss.Color("#5b626d")
	cLine   = lipgloss.Color("#2a313b")
	cOnline = lipgloss.Color("#34d399")
	cOff    = lipgloss.Color("#5b626d")
)

var sevColor = map[event.Severity]lipgloss.Color{
	event.Info:     lipgloss.Color("#6b7280"),
	event.Notice:   lipgloss.Color("#2dd4ff"),
	event.Warning:  lipgloss.Color("#f59e0b"),
	event.Error:    lipgloss.Color("#fb7185"),
	event.Critical: lipgloss.Color("#ef4444"),
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(cAccent)
	subtleStyle = lipgloss.NewStyle().Foreground(cMuted)
	faintStyle  = lipgloss.NewStyle().Foreground(cFaint)
	textStyle   = lipgloss.NewStyle().Foreground(cText)
	blueStyle   = lipgloss.NewStyle().Foreground(cBlue)

	headerBar = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).BorderForeground(cLine)

	panelTitle = lipgloss.NewStyle().Bold(true).Foreground(cBlue)
	panelBox   = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).BorderForeground(cLine)

	footerStyle = lipgloss.NewStyle().Foreground(cFaint).Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).BorderTop(true).BorderForeground(cLine)
)

// sevBadge renders a fixed-width, coloured severity tag.
func sevBadge(s event.Severity) string {
	return lipgloss.NewStyle().Bold(true).Foreground(sevColor[s]).Render(s.Short())
}

func sourceTag(src string) string {
	return lipgloss.NewStyle().Foreground(cBlue).Render(padRight(src, 8))
}

func dot(online bool) string {
	if online {
		return lipgloss.NewStyle().Foreground(cOnline).Render("●")
	}
	return lipgloss.NewStyle().Foreground(cOff).Render("○")
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + spaces(n-len(s))
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}
