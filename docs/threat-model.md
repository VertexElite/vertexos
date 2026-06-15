# VertexOS — Threat Model

> Last updated: 2026-06-15 (draft, expanding as agent research lands)

## Design center

VertexOS is a security-hardened Linux desktop whose **primary defensive
target is supply-chain attacks** against developers and the AI coding
agents they run. Every other security distro (Kali, Parrot, BlackArch)
is offensive — tools to attack with. VertexOS is the inverse: the
*hardened workstation* a developer can install AI agents on and not lose
their account, their keys, or their code.

## Adversary model

The adversaries we design against, in priority order:

### Primary — Supply-chain attackers
- **AI agent supply chain:** poisoned MCP servers, ClawHub-style malicious
  skill marketplaces, `.claude/settings.json` / `.mcp.json` injection,
  CLAUDE.md / GEMINI.md auto-load context poisoning, OAuth token theft
  from `~/.config/claude/`, agent prompt injection via web fetches.
- **Package supply chain:** npm preinstall hooks, Phantom-Gyp `binding.gyp`
  pattern, pip `setup.py` execution, cargo `build.rs`, GitHub Actions
  workflow token theft, dependency confusion, typosquatting.
- **Hardware supply chain:** untrusted USB devices, BadUSB/Rubber-Ducky
  HID injection, JieLi-class SoC firmware compromise via paired audio.

### Secondary — Network-borne malware
- BPH-hosted C2 (AS202412 Omegatech, AS30823 aurologic and successors)
- IoT botnet probing (Masjesu, AISURU/KimWolf cluster)
- Ransomware drive-by
- Watering-hole attacks against developer tools

### Tertiary — Local LPE
- Kernel exploits (unprivileged user-namespace, eBPF verifier, io_uring,
  netfilter)
- SUID binary abuse
- PolicyKit / D-Bus LPE

### Out of scope (for v0)
- Nation-state physical access (cold-boot, evil-maid beyond LUKS)
- Side-channel attacks (Spectre/Meltdown class — kernel handles)
- Targeted 0-day exploits we have no detection for

## Default defenses (this is what makes VertexOS *VertexOS*)

| Layer | Default-on defense |
|---|---|
| Kernel | Stock Void `linux` 6.18 — every KSPP/Chimera hardening flag (HARDENED_USERCOPY, SLAB_FREELIST_HARDENED, FORTIFY_SOURCE, STACKPROTECTOR_STRONG, MITIGATION_PAGE_TABLE_ISOLATION). Future: `vertex-kernel-config` srcpkg adds ZERO_CALL_USED_REGS + RANDSTRUCT_FULL. |
| Sysctl | `99-vertex-harden.conf` shipped to /etc/sysctl.d/ — tightens kptr_restrict, yama.ptrace_scope, perf_event_paranoid above Void stock + full IPv4/IPv6 stack hardening + `kernel.unprivileged_userns_clone = 0`. |
| LSM | AppArmor in **enforce** mode for browsers, file managers, terminals, editors, package managers, AI agents. |
| Network | nftables with default-deny input policy, stubby (DNS over TLS) instead of plaintext DNS, XDP Sentinel CIDR blocklist (11,800+ malicious /22s including AS202412 Omegatech ranges). |
| Detection | Falco + Sigma rules `ve-2026-001` (Phantom-Gyp), `ve-2026-002` (GitHub dead-drop config plant), `ve-2026-003` (Fake `/v1/api` endpoint). Suricata for net-side IDS. |
| Package | `xbps` is the system pm; for npm/pip/cargo/gem we ship `vertex-pkg-monitor` watching for the patterns above, plus AppArmor profiles confining each package manager. |
| AI agent | Tamper detection on `~/.config/claude/`, `~/.codeium/`, `~/.cursor/`, `.claude/settings.json`, `.mcp.json`, `CLAUDE.md`, `GEMINI.md`, `.cursor/rules/`. OAuth token files set immutable except by explicit user action. |
| USB | USBGuard set to **deny by default**, user-confirm new devices. Mitigates BadUSB / hardware injection. |
| Disk | LUKS2 enforced at installer time (no skip option — Phase 1 work on Calamares fork / Tauri installer). |
| Boot | sbctl-based Secure Boot with user-enrolled keys (not Microsoft-signed). |

## Threats we EXPLICITLY don't defend against (and why)

| Threat | Why not |
|---|---|
| Root-equivalent attacker who already has shell | Out of scope — we're protecting against pre-root attack paths. |
| Physical attacker with persistent access | Buys you LUKS. Beyond that, evil-maid is a different distro's problem. |
| Sophisticated targeted exploits with no IoC | We focus on classes we can detect at scale. |
| Microsoft/Apple-style cloud account lock-in | We're a local-first distro by design. |

## Layered defense → defense-in-depth scoring

For each attack pattern, we want **at least three independent layers** to
fail before compromise succeeds. Example: a malicious npm package would
have to defeat:
1. AppArmor confinement on `npm` (no write outside cwd + ~/.npm)
2. `vertex-pkg-monitor` detecting the install event
3. Sigma rule firing on Phantom-Gyp / preinstall-hook patterns
4. nftables / stubby blocking outbound to known C2 ASNs
5. (If C2 is novel) Falco alerting on suspicious subprocess spawning

If any 3 of those 5 must miss for the attacker to win, we're meeting the
bar.

## Open research (will land as agents complete)

- 2025-2026 Linux desktop CVE landscape (Agent A) — fold into mitigations table
- 2025-2026 AI-agent / MCP / supply-chain CVE landscape (Agent B) — fold into supply-chain-defense.md
- Branding/hardening gap audit (Agent C) — fold into overlay updates

## Coordinated disclosure

If you find a hole in VertexOS defaults:
1. Email usersafety@anthropic.com (Jamie — VertexElite has established
   relationship for coordinated disclosure)
2. CC the maintainer at the email in `git log --author`
3. We'll triage in 7 days, patch in 30, public-disclose in 90 unless agreed otherwise.

See `SECURITY.md` (TODO) for full policy.

---

Related:
- `docs/kernel-config.md` — kernel hardening detail
- `docs/supply-chain-defense.md` — pattern-by-pattern detection (TODO, expanding)
- `docs/audits/` — runnable probes that verify each layer is alive
