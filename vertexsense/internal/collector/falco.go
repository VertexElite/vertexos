package collector

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// DefaultFalcoPath is Falco's JSON output file (json_output: true).
const DefaultFalcoPath = "/var/log/falco/events.json"

// NewFalco tails Falco's JSON event stream.
func NewFalco(path string) Collector {
	if path == "" {
		path = DefaultFalcoPath
	}
	return &fileCollector{name: "falco", source: path, parse: parseFalco}
}

type falcoLine struct {
	Time     string `json:"time"`
	Priority string `json:"priority"`
	Rule     string `json:"rule"`
	Output   string `json:"output"`
}

func parseFalco(line string) (event.Event, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "{") {
		return event.Event{}, false
	}
	var f falcoLine
	if json.Unmarshal([]byte(line), &f) != nil {
		return event.Event{}, false
	}
	if f.Rule == "" && f.Output == "" {
		return event.Event{}, false
	}
	ev := event.Event{
		Severity: event.ParseSeverity(f.Priority),
		Rule:     f.Rule,
		Message:  f.Output,
	}
	if t, err := time.Parse(time.RFC3339Nano, f.Time); err == nil {
		ev.Time = t
	}
	return ev, true
}
