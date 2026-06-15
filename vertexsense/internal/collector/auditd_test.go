package collector

import (
	"testing"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

func TestParseAudit(t *testing.T) {
	// Agent credential access — should be Critical, rule "agent_credentials".
	line := `type=SYSCALL msg=audit(1718000000.123:456): arch=c000003e syscall=257 success=yes exe="/usr/bin/cat" comm="cat" key="vertex_agent_credentials"`
	ev, ok := parseAudit(line)
	if !ok {
		t.Fatal("expected match")
	}
	if ev.Severity != event.Critical {
		t.Errorf("sev = %v, want Critical", ev.Severity)
	}
	if ev.Rule != "agent_credentials" {
		t.Errorf("rule = %q", ev.Rule)
	}
	if ev.Time.IsZero() {
		t.Error("audit timestamp not parsed")
	}

	// Package-manager exec — Notice via pkg_ prefix.
	pkg := `type=SYSCALL msg=audit(1718000001.0:7): exe="/usr/bin/npm" comm="npm" key="vertex_pkg_npm"`
	if ev, ok := parseAudit(pkg); !ok || ev.Severity != event.Notice || ev.Rule != "pkg_npm" {
		t.Errorf("pkg: ok=%v sev=%v rule=%q", ok, ev.Severity, ev.Rule)
	}

	// Non-vertex key (or no key) — dropped.
	if _, ok := parseAudit(`type=SYSCALL msg=audit(1.0:1): key="something_else"`); ok {
		t.Error("non-vertex key should be dropped")
	}
	if _, ok := parseAudit(`type=DAEMON_START msg=audit(1.0:1): op=start`); ok {
		t.Error("keyless line should be dropped")
	}
}

func TestAuditField(t *testing.T) {
	line := `a=1 exe="/usr/bin/x y" comm="z" key="vertex_net"`
	if got := auditField(line, "exe"); got != "/usr/bin/x y" {
		t.Errorf("exe = %q", got)
	}
	if got := auditField(line, "a"); got != "1" {
		t.Errorf("a = %q", got)
	}
	if got := auditField(line, "missing"); got != "" {
		t.Errorf("missing = %q", got)
	}
}
