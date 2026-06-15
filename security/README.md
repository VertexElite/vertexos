# VertexOS security stack

The hardening ships as the default — not optional layers. The opinionated stance:
**VertexOS should be the box you can run a coding agent (Claude Code, Codex, …) and
untrusted dependencies on, and *see* it the moment something turns hostile.**

## Threat model

Two adversaries drive every rule here:

1. **Supply-chain compromise** — a malicious dependency or its install/postinstall
   script (`npm`/`pip`/`cargo`/…) stealing credentials, fetching a second stage,
   mining, or installing persistence.
2. **AI coding-agent abuse** — an agent that is prompt-injected, jailbroken, or
   running attacker-supplied code, then reading secrets, opening a reverse shell,
   exfiltrating data, or persisting.

## Defense in depth

| Layer | Component | What it stops | Feeds |
|-------|-----------|---------------|-------|
| **Prevent** | `apparmor/` — `vertexos-dev-sandbox` + `vtx-sandbox` | Agent/deps reading SSH/AWS/GPG/tokens or writing persistence | kernel audit |
| **Prevent** | `nftables/` — default-deny inbound, logged drops | Inbound exposure; logs egress drops in appliance mode | `vtx-` kernel log |
| **Detect (host)** | `falco/` — `vertexos-rules.yaml` | Postinstall shells, credential reads, reverse shells, pipe-to-shell, persistence, miners, setuid, /etc/shadow reads | `/var/log/falco/events.json` |
| **Detect (net)** | `suricata/` — `vertexos.rules` | Exfil to file-drops, mining handshakes, DNS tunnelling, scripted fetch, raw-IP package pulls | `eve.json` |
| **Detect (integrity)** | `aide/` — policy + `vtx-aide-check` | Tampering/persistence on binaries, global pkg dirs, cron, sudoers, PAM, SSH keys | `/var/log/aide/aide.log` |
| **Aggregate** | `../vertexsense/` (Go TUI) | One live dashboard over every feed above | — |

Every detection layer writes to a path **VertexSense** tails, so the whole stack
surfaces in a single screen.

## Run an agent safely

```bash
vtx-sandbox claude            # agent runs with secrets + persistence out of reach
vertexsense                   # watch the sensors in real time
```

## Status

Alpha. Rules are conservative and threat-model-driven; expect per-workload
allow-list tuning once running on real traffic. CrowdSec bouncer + XDP Sentinel
land in Phase 3 (see `../kernel/` and `nftables/` TODOs).
