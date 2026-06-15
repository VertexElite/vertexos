# VertexSense

`VertexSense` — the VertexOS security telemetry TUI. It fuses the live output of
every sensor (Falco / Suricata / nftables / AIDE) into a single terminal
dashboard: one feed, one severity scale, per-sensor health.

```
◆ VertexSense  ·  VertexOS security telemetry                         19:11:43
CRIT 1  ERR 2  WARN 4  NOTE 2  INFO 0   ·   total 9
──────────────────────────────────────────────────────────────────────────────
 SENSORS            │ EVENT FEED
 ● falco     12     │ 19:11:38 CRIT falco    Terminal shell in container  …
   online · 2s ago  │ 19:11:33 ERR  suricata ET SCAN Potential SSH Scan  …
 ○ suricata  0      │ 19:11:28 NOTE nftables nft-drop  drop 198.51.100.9 …
   offline · —      │ …
──────────────────────────────────────────────────────────────────────────────
 ↑/↓ scroll · g/G top/bottom · p pause · c clear · 0-4 min-sev · s source · q quit
```

## Run

```bash
vertexsense              # tail the live sensor outputs (default paths)
vertexsense --demo       # synthetic feed — no live sensors required
vertexsense --snapshot   # render one frame to stdout and exit (no TTY)
```

Keys: `↑/↓` scroll · `g/G` top/bottom · `p` pause · `c` clear · `0–4` minimum
severity · `s` cycle source filter · `q` quit.

## Sources (override with flags)

| Sensor    | Default path                   | Format                    | Flag         |
|-----------|--------------------------------|---------------------------|--------------|
| Falco     | `/var/log/falco/events.json`   | JSON lines (`json_output`)| `--falco`    |
| Suricata  | `/var/log/suricata/eve.json`   | EVE JSON (alerts only)    | `--suricata` |
| nftables  | `/var/log/messages`            | kernel log, `vtx-` prefix | `--nftlog` / `--nftprefix` |
| AIDE      | `/var/log/aide/aide.log`       | added/changed/removed     | `--aide`     |

A sensor whose source file is absent shows **offline** — VertexSense never
crashes on a missing engine; it simply waits for the file to appear.

## Layout

```
cmd/vertexsense/      TUI entrypoint + flags
internal/event/       normalized Event + 5-level severity
internal/tail/        poll-based log follower (rotation/truncation safe)
internal/collector/   Falco · Suricata · nftables · AIDE · demo collectors
internal/ui/          bubbletea model + lipgloss Aero styling
```

Packaged as the `vertexsense` srcpkg under `packages/srcpkgs/`.

## Test

```bash
go test ./...        # event parsers, severity mapping, tailer rotation/creation
go vet ./...
```
