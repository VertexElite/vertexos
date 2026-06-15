// Package collector turns each security engine's output into a stream of
// normalized event.Event values, and tracks per-sensor health for the dashboard.
//
// Most sensors write line-oriented files (Falco/Suricata JSON, nft kernel log,
// AIDE report), so they share fileCollector: tail the source, parse each line,
// emit events, and report Online/Offline based on whether the source exists.
package collector

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/VertexElite/vertexos/vertexsense/internal/event"
	"github.com/VertexElite/vertexos/vertexsense/internal/tail"
)

// Health is the live state of a sensor's data source.
type Health int

const (
	Offline Health = iota // source not present / not readable
	Online                // source present, being tailed
	Degraded              // source present but erroring
)

func (h Health) String() string {
	switch h {
	case Online:
		return "online"
	case Degraded:
		return "degraded"
	default:
		return "offline"
	}
}

// Status is a snapshot of a collector for the UI.
type Status struct {
	Name   string
	Source string
	Health Health
	Count  int
	Last   time.Time
	Note   string
}

// Collector streams normalized events and reports its own health.
type Collector interface {
	Name() string
	Source() string
	Run(ctx context.Context, out chan<- event.Event)
	Status() Status
}

// base provides thread-safe status bookkeeping shared by all collectors.
type base struct {
	mu     sync.Mutex
	health Health
	count  int
	last   time.Time
	note   string
}

func (b *base) record(ev event.Event) {
	b.mu.Lock()
	b.count++
	if ev.Time.After(b.last) {
		b.last = ev.Time
	}
	b.mu.Unlock()
}

func (b *base) setHealth(h Health, note string) {
	b.mu.Lock()
	b.health, b.note = h, note
	b.mu.Unlock()
}

func (b *base) snap(name, source string) Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	return Status{Name: name, Source: source, Health: b.health, Count: b.count, Last: b.last, Note: b.note}
}

// ParseFunc converts one raw line into an event. ok=false drops the line
// (e.g. a Suricata flow record that isn't an alert).
type ParseFunc func(line string) (event.Event, bool)

// fileCollector tails a single file source and parses each line.
type fileCollector struct {
	base
	name      string
	source    string
	fromStart bool
	parse     ParseFunc
}

func (c *fileCollector) Name() string   { return c.name }
func (c *fileCollector) Source() string { return c.source }
func (c *fileCollector) Status() Status { return c.snap(c.name, c.source) }

func (c *fileCollector) Run(ctx context.Context, out chan<- event.Event) {
	lines := make(chan string, 256)
	go tail.Follow(ctx, c.source, c.fromStart, 500*time.Millisecond, lines)
	go c.watchHealth(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case line := <-lines:
			ev, ok := c.parse(line)
			if !ok {
				continue
			}
			ev.Source = c.name
			if ev.Time.IsZero() {
				ev.Time = time.Now()
			}
			c.record(ev)
			select {
			case out <- ev:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *fileCollector) watchHealth(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	check := func() {
		if _, err := os.Stat(c.source); err == nil {
			c.setHealth(Online, "")
		} else {
			c.setHealth(Offline, "source not present")
		}
	}
	check()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			check()
		}
	}
}
