package collector

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
)

// TestFileCollectorEndToEnd exercises the full path: a live-appended log line is
// tailed, parsed, emitted on the channel, and reflected in the collector's
// health/count — i.e. exactly what the running dashboard depends on.
func TestFileCollectorEndToEnd(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "events.json")
	if err := os.WriteFile(p, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := NewFalco(p)
	out := make(chan event.Event, 8)
	go c.Run(ctx, out)

	time.Sleep(120 * time.Millisecond) // let the tailer attach + health poll

	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"time":"2026-06-15T12:00:00.0Z","priority":"Critical","rule":"Reverse shell","output":"nc -e launched"}` + "\n")
	f.Close()

	select {
	case ev := <-out:
		if ev.Source != "falco" {
			t.Errorf("source = %q, want falco", ev.Source)
		}
		if ev.Severity != event.Critical {
			t.Errorf("severity = %v, want Critical", ev.Severity)
		}
		if ev.Rule != "Reverse shell" {
			t.Errorf("rule = %q", ev.Rule)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no event received from live-appended log line")
	}

	if st := c.Status(); st.Health != Online || st.Count < 1 {
		t.Errorf("status = %+v, want Online with count>=1", st)
	}
}
