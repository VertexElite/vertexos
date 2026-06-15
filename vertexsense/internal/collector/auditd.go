package collector

import (
	"strconv"
	"strings"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// DefaultAuditLog is the Linux audit daemon's log.
const DefaultAuditLog = "/var/log/audit/audit.log"

// NewAuditd tails auditd and surfaces VertexOS-tagged events (the rules in
// /etc/audit/rules.d/99-vertex.rules tag everything with a `vertex_*` key:
// agent credential access, package-manager execs, module loads, identity /
// persistence / init changes). This is VertexOS's primary host sensor — Falco
// isn't packaged for Void, so auditd carries host-side detection.
func NewAuditd(path string) Collector {
	if path == "" {
		path = DefaultAuditLog
	}
	return &fileCollector{name: "auditd", source: path, parse: parseAudit}
}

// severity per audit key (or prefix). Package-manager execs are informational
// context; credential/identity/module events are high-signal.
var auditKeySeverity = map[string]event.Severity{
	"agent_credentials":    event.Critical,
	"identity":             event.Critical,
	"module_load":          event.Error,
	"module_unload":        event.Error,
	"agent_hook":           event.Warning,
	"agent_mcp_open":       event.Warning,
	"agent_settings_open":  event.Warning,
	"persistence":          event.Warning,
	"init":                 event.Warning,
	"net":                  event.Notice,
	"agent_ctx_open":       event.Notice,
	"home_writes":          event.Info,
}

func parseAudit(line string) (event.Event, bool) {
	key := auditField(line, "key")
	if !strings.HasPrefix(key, "vertex_") {
		return event.Event{}, false
	}
	short := strings.TrimPrefix(key, "vertex_")

	sev := event.Notice
	if strings.HasPrefix(short, "pkg_") {
		sev = event.Notice
	} else if s, ok := auditKeySeverity[short]; ok {
		sev = s
	}

	exe := auditField(line, "exe")
	comm := auditField(line, "comm")
	name := auditField(line, "name") // present on PATH records
	syscall := auditField(line, "syscall")

	var parts []string
	if exe != "" {
		parts = append(parts, exe)
	} else if comm != "" {
		parts = append(parts, comm)
	}
	if name != "" {
		parts = append(parts, "→ "+name)
	}
	if syscall != "" {
		parts = append(parts, "syscall="+syscall)
	}
	msg := strings.Join(parts, " ")
	if msg == "" {
		msg = strings.TrimSpace(line)
	}

	ev := event.Event{Severity: sev, Rule: short, Message: msg}
	if t, ok := auditTime(line); ok {
		ev.Time = t
	}
	return ev, true
}

// auditField extracts `name=value` or `name="value"` from an audit line.
func auditField(line, name string) string {
	i := strings.Index(line, name+"=")
	if i < 0 {
		return ""
	}
	v := line[i+len(name)+1:]
	if len(v) > 0 && v[0] == '"' {
		v = v[1:]
		if j := strings.IndexByte(v, '"'); j >= 0 {
			return v[:j]
		}
		return v
	}
	if j := strings.IndexByte(v, ' '); j >= 0 {
		return v[:j]
	}
	return v
}

// auditTime parses the epoch from `msg=audit(1718000000.123:456)`.
func auditTime(line string) (time.Time, bool) {
	i := strings.Index(line, "audit(")
	if i < 0 {
		return time.Time{}, false
	}
	rest := line[i+len("audit("):]
	end := strings.IndexByte(rest, ':')
	if end < 0 {
		return time.Time{}, false
	}
	secStr := rest[:end]
	if dot := strings.IndexByte(secStr, '.'); dot >= 0 {
		secStr = secStr[:dot]
	}
	sec, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(sec, 0), true
}
