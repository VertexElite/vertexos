# VertexOS — Supply-Chain Defense Architecture

> The defensive thesis of VertexOS: **every published finding from VertexElite
> Security Research is the default on this OS.** If you've read the ClawHavoc
> report, the Miasma Worm writeup, the Convergent Architecture v3.0 doc — the
> defenses in those reports ship enabled out of the box here.

## Pattern-by-pattern defense map

Each row is a documented supply-chain attack pattern and the VertexOS
defaults that detect or block it. Reference column links to the VES report.

### AI agent supply chain

| Attack pattern | VertexOS default defense | Reference |
|---|---|---|
| ClawHub-style malicious skill marketplace (341 poisoned packages) | `vertex-pkg-monitor` daemon watching `~/.claude/skills/`, `~/.codeium/`, `~/.cursor/`. Cryptographic verification of skill manifests against known-good registry. | VES-CA-2026-002 |
| `.claude/settings.json` PreToolUse/PostToolUse hook injection | AppArmor profile on `claude` binary denies writes to `**/.claude/settings.json` and `**/.mcp.json` except via explicit user CLI command. Inotify watcher in `vertex-pkg-monitor` alerts on any change. | VES-CA-2026-001 (Layer 3) |
| `.mcp.json` malicious MCP server registration | Allowlist of vetted MCP servers in `/etc/vertexos/mcp-allowlist.json`. Any unlisted registration triggers Falco alert + user-prompt before activation. | VES-CA-2026-001 (Layer 3) |
| CLAUDE.md / GEMINI.md auto-load context poisoning (Miasma Level 5 — invisible to tool-call detection) | `vertex-context-guard` reads every `CLAUDE.md`/`GEMINI.md` in repos and computes hash diff against last user-confirmed version; warns on diff before agent reads. | VES-CA-2026-001 unique finding #2; Miasma Level 5 |
| OAuth token theft from agent credential files | `chattr +i` (immutable) flag on `~/.config/claude/.credentials.json`, removable only with explicit user `vertex-cred unlock`. Audit log of every read access. | VES-ECC-2026-001 |
| Cross-agent contamination via `gemini hooks migrate --from-claude` | AppArmor profile on `gemini` denies reads from `**/.claude/` unless `--explicit-migrate` flag passed and user-confirmed. | VES-CA-2026-001 unique finding #1 |
| AgentJacking via poisoned MCP server (Tenet Security, June 2026) | All MCP server outputs piped through `vertex-mcp-sanitizer` that strips control sequences and detects prompt-injection patterns before passing to the LLM. | VES-CA-2026-001 (Layer 3 update — pending Agent B research) |

### Package supply chain

| Attack pattern | VertexOS default defense | Reference |
|---|---|---|
| npm preinstall hook | AppArmor profile on `npm` denies all child-process spawning during install except `node` + `python` + the system shell, with no network egress except to registry.npmjs.org pinned by stubby DoT. | Miasma Level 1 |
| Phantom-Gyp (`binding.gyp` + `type:none` + no C/C++ source) | Sigma rule `ve-2026-001` — Falco fires on `npm install` event with these markers. Hard-block in `vertex-pkg-monitor`. **Validated zero-FP against 5 production apps including `sharp`.** | VES Miasma report |
| pip `setup.py` arbitrary code | AppArmor profile on `pip` denies subprocess spawning during install + network egress except to `pypi.org`. Prefer `pip install --no-build-isolation --only-binary=:all:` by default in shipped pip wrapper. | Standard Python supply-chain pattern |
| cargo `build.rs` arbitrary code | AppArmor profile on `cargo` denies network during build except to `crates.io` + `index.crates.io`. Future: cargo wrapper that warns on first encounter of a `build.rs` per crate. | Standard Rust supply-chain pattern |
| gem `post_install_message` + native extension | AppArmor profile on `gem` denies subprocess spawning during install. | Standard Ruby pattern |
| GitHub Actions `workflow_run` / `pull_request_target` token theft | `vertex-pkg-monitor` includes a `gh actions audit` subcommand that flags risky workflow triggers in cloned repos. | (CISA + recent GitHub advisories) |
| GitHub dead-drop config plant (`liuende501`, `windy629`, `HerGomUli`, `affaan-m` accounts) | Sigma rule `ve-2026-002` — Falco fires on any file write matching `(\.claude\|\.gemini\|\.cursor)/(settings\|rules)\.[jy][saml]+` originating from a `git clone` subprocess. | VES Miasma report |
| Fake `/v1/api` endpoint (zero-FP by definition) | Sigma rule `ve-2026-003` — Suricata fires on outbound HTTP POST to any `/v1/api` path. Path doesn't exist in legitimate deployments. | VES Miasma report |
| Dependency confusion (private package name resolves to public) | `vertex-pkg-monitor` checks every `package.json` / `requirements.txt` / `Cargo.toml` against a private-namespace registry the user maintains. | Standard pattern |
| Typosquatting (`reqests` vs `requests`) | Levenshtein-distance check in `vertex-pkg-monitor` against top-1000-package allowlist. Warn-only by default. | Standard pattern |

### Network supply chain

| Attack pattern | VertexOS default defense | Reference |
|---|---|---|
| Connection to AS202412 Omegatech BPH | XDP Sentinel CIDR blocklist with 11,800+ entries including `91.92.240.0/22`, `158.94.208.0/22`, `178.16.52.0/22`. Updated daily via runit cron from VertexElite/sentinel-blocklist repo. | VES-CA-2026-002 |
| Connection to AS30823 aurologic (upstream carrier) | Same blocklist | VES-CA-2026-002 |
| Connection to AS205759 GHOSTYNETWORKS (sibling BPH) | Same blocklist | VES-CA-2026-001 (Layer 4) |
| Connection to `slot0.asident-tr.com` / `gitlab.pwnhub.fail` | Domain blocklist in stubby + Suricata | VES-CA-2026-002 IOC list |
| Connection to `path3.xtracloud.net` (Qualcomm XTRA GPS structural opacity) | NetworkManager pre-up hook checks for and warns about `com.qti.qcc`-equivalent calls if any Qualcomm-derived firmware is running | VES-CA-2026-001 (Layer 5) |
| Connection to `gb4w8c3ygj-default-sea.rum.aliyuncs.com` (Qwen CLI behavioral telemetry beacon) | Suricata rule + stubby DoT blocklist. Hard block. | VES-CA-2026-001 (Layer 5) |
| Beacon to `thebeautifulmarchoftime` / `firedalazer` (Miasma C2 keyword) | Suricata content-match rule on HTTP body | VES Miasma report IOCs |

### Hardware supply chain

| Attack pattern | VertexOS default defense | Reference |
|---|---|---|
| BadUSB / HID injection from unknown USB device | USBGuard policy: deny by default, user-confirm new devices. UI prompt on insertion. | Standard pattern |
| JieLi AC692x-class SoC firmware OTA over BLE | `bluez` blocked from accepting RCSP/OTA frames except to allowlisted devices in `/etc/vertexos/bt-allowlist.json` | VES-JL-2026-001 v2.0 |
| USB-attached storage with malicious autorun | `autofs` disabled by default. File manager (`thunar`) AppArmor profile denies execution of binaries from removable mounts. | Standard pattern |

## Implementation status (June 15, 2026)

| Component | Status |
|---|---|
| AppArmor profiles for package managers (npm/pip/cargo/gem) | **Not shipped** — Phase 1 priority |
| `vertex-pkg-monitor` daemon | **Not yet built** — Rust + Tauri pattern, Phase 1.5 |
| `vertex-context-guard` (CLAUDE.md hash watcher) | **Not yet built** — Phase 1.5 |
| `vertex-mcp-sanitizer` | **Not yet built** — Phase 1.5 |
| XDP Sentinel + Omegatech CIDR blocklist | Code exists in maintainer's OVH deployment; **port to ISO Phase 3** |
| Sigma rules ve-2026-001/002/003 | Exist in `VertexElite/miasma-detection` repo; need wiring into Falco default config Phase 2 |
| USBGuard deny-by-default | Package installed; default policy still permissive — needs overlay /etc/usbguard/usbguard-daemon.conf |
| OAuth credential immutability | Manual flag today; needs `vertex-cred` CLI wrapper Phase 1.5 |
| stubby (DoT) | Package installed; runit service enabled in overlay (this commit) |
| nftables default-deny | Ruleset shipped at /etc/nftables.conf; service enabled in overlay (this commit) |

## Non-defenses

Things we deliberately don't ship as on-by-default because the friction
outweighs the benefit for the developer-workstation use case:

- Full-disk-encryption mandatory ON THE LIVE ISO (live is throwaway). LUKS
  is enforced at install time only.
- Network jailing of every desktop app (would break browser, IDEs).
- Real-time scanning of every file open (we use scheduled AIDE + ClamAV
  scans instead — performance trade-off).

## Updating this document

This file is a living artifact. As VertexElite Security Research publishes
new findings, the corresponding row gets added here and the corresponding
default gets shipped in the next ISO build. Treat this as the spec for
"what VertexOS defends".

---

Related:
- `docs/threat-model.md` — overall design center
- `docs/audits/` — runnable verification probes
- `VertexElite/miasma-detection` — open-source detection toolkit (upstream of Sigma rules)
