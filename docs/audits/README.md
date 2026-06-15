# A-Z Audit scripts

Self-contained probes for the running system. Each emits markdown to stdout;
some also emit a JSON sidecar for tooling.

## Run on a target VertexOS instance

```bash
sudo apt-get install -y apparmor-utils libcap-ng-utils  # not all distros ship aa-status/getcap by default
cd docs/audits
sudo ./run-all.sh > audit-$(date +%Y%m%d-%H%M).md
```

## Scripts

| Script | What it answers |
|---|---|
| `01-kernel-hardening.sh` | Which KSPP/Chimera CONFIG flags are actually compiled in? Which sysctl values from `99-vertex-harden.conf` actually applied? What CPU mitigations are active? |
| `02-apparmor-coverage.sh` | LSM loaded? Profile count? Enforce vs complain ratio? What fraction of processes are confined? |
| `03-network-exposure.sh` | What's listening externally? nftables default policy + ruleset? Is stubby (DoT) up? |
| `04-privesc-paths.sh` | SUID/SGID inventory, file capabilities, sudoers, world-writables, kernel-LPE prerequisite knobs. |

## Output

Each script's output should be **diffable across builds** — running the same
script on a v0.1 ISO vs a v0.2 ISO produces a delta showing exactly what
hardening landed (or regressed). Commit each baseline under
`docs/audits/history/` if you want a permanent record.

## Adding new dimensions

Drop a `0N-something.sh` in this dir following the format:

1. Emit pure markdown to stdout (no terminal colors).
2. Use `##` headers — `run-all.sh` will assemble them under a top-level doc.
3. End with a `## Recommendation` section that flags failures actionably.
