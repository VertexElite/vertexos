package collector

import (
	"context"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// demoCollector emits a synthetic, realistic event feed so the dashboard is
// demonstrable without live sensors (screenshots, CI, first-run preview). Events
// are tagged with real source names so the feed colouring/filtering looks true.
type demoCollector struct {
	base
	interval time.Duration
}

// NewDemo returns a synthetic feed collector. interval<=0 uses 1.2s.
func NewDemo(interval time.Duration) Collector {
	if interval <= 0 {
		interval = 1200 * time.Millisecond
	}
	return &demoCollector{interval: interval}
}

func (d *demoCollector) Name() string   { return "demo" }
func (d *demoCollector) Source() string { return "(synthetic)" }
func (d *demoCollector) Status() Status {
	s := d.snap("demo", "(synthetic)")
	s.Health = Online
	s.Note = "synthetic feed"
	return s
}

var demoSamples = []event.Event{
	{Source: "falco", Severity: event.Critical, Rule: "Terminal shell in container", Message: "shell spawned in container (user=root proc=bash parent=runc)"},
	{Source: "suricata", Severity: event.Error, Rule: "ET SCAN Potential SSH Scan", Message: "203.0.113.7 -> 10.0.0.5 TCP (Attempted Information Leak)"},
	{Source: "nftables", Severity: event.Notice, Rule: "nft-drop", Message: "drop 198.51.100.9 -> 10.0.0.5 TCP dpt 23"},
	{Source: "falco", Severity: event.Warning, Rule: "Write below etc", Message: "file below /etc opened for writing (file=/etc/cron.d/x)"},
	{Source: "aide", Severity: event.Error, Rule: "aide-removed", Message: "removed: /usr/bin/sudo"},
	{Source: "suricata", Severity: event.Warning, Rule: "ET POLICY curl User-Agent", Message: "10.0.0.5 -> 185.199.108.153 TCP (Potentially Bad Traffic)"},
	{Source: "nftables", Severity: event.Warning, Rule: "nft-drop", Message: "drop 192.0.2.44 -> 10.0.0.5 TCP dpt 3389 (scan)"},
	{Source: "aide", Severity: event.Warning, Rule: "aide-changed", Message: "changed: /etc/ssh/sshd_config"},
	{Source: "falco", Severity: event.Notice, Rule: "Read sensitive file", Message: "sensitive file opened for reading (file=/etc/shadow proc=cat)"},
}

// DemoEvents returns n synthetic events with staggered (ascending) timestamps,
// for snapshot rendering and tests. n<=0 returns one of each sample.
func DemoEvents(n int) []event.Event {
	if n <= 0 {
		n = len(demoSamples)
	}
	out := make([]event.Event, 0, n)
	base := time.Now().Add(-time.Duration(n) * 5 * time.Second)
	for i := 0; i < n; i++ {
		ev := demoSamples[i%len(demoSamples)]
		ev.Time = base.Add(time.Duration(i) * 5 * time.Second)
		out = append(out, ev)
	}
	return out
}

func (d *demoCollector) Run(ctx context.Context, out chan<- event.Event) {
	d.setHealth(Online, "synthetic feed")
	t := time.NewTicker(d.interval)
	defer t.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			ev := demoSamples[i%len(demoSamples)]
			ev.Time = time.Now()
			i++
			d.record(ev)
			select {
			case out <- ev:
			case <-ctx.Done():
				return
			}
		}
	}
}
