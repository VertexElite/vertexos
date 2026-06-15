package collector

import (
	"testing"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

func TestParseFalco(t *testing.T) {
	line := `{"time":"2026-06-15T12:00:00.000000000Z","priority":"Critical","rule":"Terminal shell in container","output":"shell in container"}`
	ev, ok := parseFalco(line)
	if !ok {
		t.Fatal("expected ok")
	}
	if ev.Severity != event.Critical {
		t.Errorf("sev = %v, want Critical", ev.Severity)
	}
	if ev.Rule != "Terminal shell in container" {
		t.Errorf("rule = %q", ev.Rule)
	}
	if ev.Time.IsZero() {
		t.Error("time not parsed")
	}
	if _, ok := parseFalco("not json"); ok {
		t.Error("non-json should be dropped")
	}
}

func TestParseSuricataAlertsOnly(t *testing.T) {
	alert := `{"timestamp":"2026-06-15T12:00:00.000000+0000","event_type":"alert","src_ip":"203.0.113.7","dest_ip":"10.0.0.5","proto":"TCP","alert":{"signature":"ET SCAN","category":"Attempted Information Leak","severity":1}}`
	ev, ok := parseSuricata(alert)
	if !ok {
		t.Fatal("alert should parse")
	}
	if ev.Severity != event.Error { // suricata sev 1 -> Error
		t.Errorf("sev = %v, want Error", ev.Severity)
	}
	if ev.Rule != "ET SCAN" {
		t.Errorf("rule = %q", ev.Rule)
	}
	flow := `{"timestamp":"2026-06-15T12:00:00.000000+0000","event_type":"flow","src_ip":"1.1.1.1"}`
	if _, ok := parseSuricata(flow); ok {
		t.Error("non-alert event_type should be dropped")
	}
}

func TestNftParser(t *testing.T) {
	p := nftParser("vtx-")
	line := `Jun 15 12:00:00 host kernel: vtx-drop IN=eth0 OUT= SRC=198.51.100.9 DST=10.0.0.5 PROTO=TCP SPT=51000 DPT=23`
	ev, ok := p(line)
	if !ok {
		t.Fatal("prefixed line should parse")
	}
	if ev.Rule != "nft-drop" {
		t.Errorf("rule = %q", ev.Rule)
	}
	if ev.Message != "drop 198.51.100.9 -> 10.0.0.5 TCP dpt 23" {
		t.Errorf("msg = %q", ev.Message)
	}
	if _, ok := p("Jun 15 12:00:00 host kernel: unrelated message"); ok {
		t.Error("line without prefix should be dropped")
	}
}

func TestParseAIDE(t *testing.T) {
	ev, ok := parseAIDE("removed: /usr/bin/sudo")
	if !ok || ev.Severity != event.Error || ev.Rule != "aide-removed" {
		t.Errorf("removed: ok=%v sev=%v rule=%q", ok, ev.Severity, ev.Rule)
	}
	ev, ok = parseAIDE("changed: /etc/ssh/sshd_config")
	if !ok || ev.Severity != event.Warning || ev.Rule != "aide-changed" {
		t.Errorf("changed: ok=%v sev=%v rule=%q", ok, ev.Severity, ev.Rule)
	}
	if _, ok := parseAIDE("AIDE found no differences"); ok {
		t.Error("non-change line should be dropped")
	}
}
