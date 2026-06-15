# AIDE — filesystem integrity

Catches the *tampering / persistence* half of a supply-chain or agent compromise:
modified system binaries, planted global packages, new cron/sudoers/PAM/SSH-key
entries. `vtx-aide-check` normalizes AIDE's report into `added:/removed:/changed:`
lines so VertexSense renders integrity drift live.

## Files

- `aide.conf` → `/etc/aide.conf` — policy (binaries, global pkg dirs, `/etc`,
  cron, services, sudoers, PAM, `~/.ssh`; volatile paths excluded).
- `vtx-aide-check` → `/usr/local/bin/vtx-aide-check` — init-or-check adapter,
  output to `/var/log/aide/aide.log`.
- `sv/vtx-aide/run` → `/etc/sv/vtx-aide/run` — runit service (6h interval).

## Install

```bash
install -m0644 aide.conf      /etc/aide.conf
install -m0755 vtx-aide-check /usr/local/bin/vtx-aide-check
install -Dm0755 sv/vtx-aide/run /etc/sv/vtx-aide/run
vtx-aide-check                          # builds the baseline DB on first run
ln -s /etc/sv/vtx-aide /var/service/    # enable the periodic check (runit)
```

First run initializes `/var/lib/aide/aide.db` and logs a baseline marker; every
run after that appends only real changes. Re-baseline after legitimate updates:
`rm /var/lib/aide/aide.db && vtx-aide-check`.
