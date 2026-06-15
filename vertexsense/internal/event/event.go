// Package event is the common security-event model that every VertexSense
// collector emits and the TUI renders. Keeping one normalized type means Falco,
// Suricata, nftables and AIDE all flow through a single feed and severity scale.
package event

import (
	"strings"
	"time"
)

// Severity is a normalized 5-level scale all sensors are mapped onto.
type Severity int

const (
	Info Severity = iota
	Notice
	Warning
	Error
	Critical
)

func (s Severity) String() string {
	switch s {
	case Info:
		return "INFO"
	case Notice:
		return "NOTICE"
	case Warning:
		return "WARN"
	case Error:
		return "ERROR"
	case Critical:
		return "CRIT"
	default:
		return "?"
	}
}

// Short is a fixed-width tag for column alignment in the TUI.
func (s Severity) Short() string {
	switch s {
	case Info:
		return "INFO"
	case Notice:
		return "NOTE"
	case Warning:
		return "WARN"
	case Error:
		return "ERR "
	case Critical:
		return "CRIT"
	default:
		return "??? "
	}
}

// ParseSeverity maps the free-text priority words used by Falco
// (emergency/alert/critical/error/warning/notice/informational/debug) and
// similar engines onto the normalized scale. Unknown -> Notice.
func ParseSeverity(s string) Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "emergency", "alert", "critical", "crit", "fatal":
		return Critical
	case "error", "err":
		return Error
	case "warning", "warn":
		return Warning
	case "notice", "note":
		return Notice
	case "informational", "info", "debug", "":
		return Info
	default:
		return Notice
	}
}

// Event is one normalized security observation.
type Event struct {
	Time     time.Time
	Source   string // collector name: falco, suricata, nftables, aide, demo
	Severity Severity
	Rule     string // rule / signature / short label
	Message  string // human-readable detail
}
