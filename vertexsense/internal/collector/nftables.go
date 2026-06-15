package collector

import (
	"regexp"
	"strings"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// nftables logs dropped packets to the kernel ring buffer via
// `log prefix "vtx-drop "` rules; syslog/socklog persists them here.
const (
	DefaultNftLog    = "/var/log/messages"
	DefaultNftPrefix = "vtx-"
)

var nftFields = regexp.MustCompile(`SRC=(\S+).*?DST=(\S+).*?PROTO=(\S+)(?:.*?DPT=(\S+))?`)

// NewNftables tails the kernel log and surfaces lines carrying the VertexOS
// nftables log prefix. prefix/path default to the VertexOS conventions.
func NewNftables(path, prefix string) Collector {
	if path == "" {
		path = DefaultNftLog
	}
	if prefix == "" {
		prefix = DefaultNftPrefix
	}
	return &fileCollector{name: "nftables", source: path, parse: nftParser(prefix)}
}

func nftParser(prefix string) ParseFunc {
	return func(line string) (event.Event, bool) {
		if !strings.Contains(line, prefix) {
			return event.Event{}, false
		}
		msg := strings.TrimSpace(line)
		if m := nftFields.FindStringSubmatch(line); m != nil {
			dpt := ""
			if m[4] != "" {
				dpt = " dpt " + m[4]
			}
			msg = "drop " + m[1] + " -> " + m[2] + " " + m[3] + dpt
		}
		sev := event.Notice
		low := strings.ToLower(line)
		if strings.Contains(low, "invalid") || strings.Contains(low, "scan") {
			sev = event.Warning
		}
		return event.Event{Severity: sev, Rule: "nft-drop", Message: msg}, true
	}
}
