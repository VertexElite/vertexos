package collector

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// DefaultSuricataPath is Suricata's EVE JSON log.
const DefaultSuricataPath = "/var/log/suricata/eve.json"

// NewSuricata tails Suricata's eve.json and surfaces only alert records.
func NewSuricata(path string) Collector {
	if path == "" {
		path = DefaultSuricataPath
	}
	return &fileCollector{name: "suricata", source: path, parse: parseSuricata}
}

type suricataLine struct {
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	SrcIP     string `json:"src_ip"`
	DestIP    string `json:"dest_ip"`
	Proto     string `json:"proto"`
	Alert     *struct {
		Signature string `json:"signature"`
		Category  string `json:"category"`
		Severity  int    `json:"severity"`
	} `json:"alert"`
}

func parseSuricata(line string) (event.Event, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "{") {
		return event.Event{}, false
	}
	var s suricataLine
	if json.Unmarshal([]byte(line), &s) != nil {
		return event.Event{}, false
	}
	if s.EventType != "alert" || s.Alert == nil {
		return event.Event{}, false // ignore flow/dns/http noise — alerts only
	}
	// Suricata severity: 1 = most severe.
	sev := event.Warning
	switch s.Alert.Severity {
	case 1:
		sev = event.Error
	case 2:
		sev = event.Warning
	case 3:
		sev = event.Notice
	}
	msg := strings.TrimSpace(s.SrcIP + " -> " + s.DestIP + " " + s.Proto)
	if s.Alert.Category != "" {
		msg += " (" + s.Alert.Category + ")"
	}
	ev := event.Event{Severity: sev, Rule: s.Alert.Signature, Message: msg}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999-0700"} {
		if t, err := time.Parse(layout, s.Timestamp); err == nil {
			ev.Time = t
			break
		}
	}
	return ev, true
}
