package collector

import (
	"strings"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// DefaultAideLog is where the VertexOS AIDE check service writes its report.
const DefaultAideLog = "/var/log/aide/aide.log"

// NewAIDE tails the AIDE report and surfaces filesystem-integrity changes.
func NewAIDE(path string) Collector {
	if path == "" {
		path = DefaultAideLog
	}
	return &fileCollector{name: "aide", source: path, parse: parseAIDE}
}

func parseAIDE(line string) (event.Event, bool) {
	l := strings.TrimSpace(line)
	low := strings.ToLower(l)
	var rule string
	sev := event.Warning
	switch {
	case strings.HasPrefix(low, "added:"), strings.HasPrefix(low, "added file:"):
		rule = "aide-added"
	case strings.HasPrefix(low, "removed:"), strings.HasPrefix(low, "removed file:"):
		rule, sev = "aide-removed", event.Error // disappearing files = higher concern
	case strings.HasPrefix(low, "changed:"), strings.HasPrefix(low, "changed file:"):
		rule = "aide-changed"
	default:
		return event.Event{}, false
	}
	return event.Event{Severity: sev, Rule: rule, Message: l}, true
}
