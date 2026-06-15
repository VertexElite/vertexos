# vertexsense/

`VertexSense` — the Go TUI that aggregates the security sensors
(Falco / Suricata / Zeek / Wazuh / ntopng) into a single dashboard.

Planned (Phase 2-3):

- `cmd/vertexsense/` — TUI entrypoint (bubbletea).
- `internal/` — per-sensor collectors that tail/stream each engine's output.
- packaged as the `vertexsense` srcpkg under `packages/`.

Empty for now — no Go source committed yet.
