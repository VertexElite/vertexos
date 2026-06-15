// Command vertexsense is the VertexOS security telemetry dashboard: it fuses the
// live output of Falco, Suricata, nftables and AIDE into a single terminal UI.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/VertexElite/vertexos/vertexsense/internal/collector"
	"github.com/VertexElite/vertexos/vertexsense/internal/event"
	"github.com/VertexElite/vertexos/vertexsense/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var (
		demo     = flag.Bool("demo", false, "emit a synthetic feed (no live sensors required)")
		snapshot = flag.Bool("snapshot", false, "render one frame to stdout and exit (no TTY needed)")
		w        = flag.Int("w", 120, "snapshot width")
		h        = flag.Int("h", 32, "snapshot height")

		falcoPath = flag.String("falco", collector.DefaultFalcoPath, "Falco JSON output path")
		suriPath  = flag.String("suricata", collector.DefaultSuricataPath, "Suricata eve.json path")
		nftPath   = flag.String("nftlog", collector.DefaultNftLog, "kernel log path for nftables drops")
		nftPrefix = flag.String("nftprefix", collector.DefaultNftPrefix, "nftables log prefix to match")
		aidePath  = flag.String("aide", collector.DefaultAideLog, "AIDE report log path")
	)
	flag.Parse()

	cols := []collector.Collector{
		collector.NewFalco(*falcoPath),
		collector.NewSuricata(*suriPath),
		collector.NewNftables(*nftPath, *nftPrefix),
		collector.NewAIDE(*aidePath),
	}
	if *demo {
		cols = append(cols, collector.NewDemo(0))
	}

	if *snapshot {
		fmt.Println(ui.Snapshot(cols, collector.DemoEvents(0), *w, *h))
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ch := make(chan event.Event, 512)
	for _, c := range cols {
		go c.Run(ctx, ch)
	}

	p := tea.NewProgram(ui.New(cols, ch, 2000), tea.WithAltScreen(), tea.WithContext(ctx))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "vertexsense:", err)
		os.Exit(1)
	}
}
