# AppArmor — the dev sandbox

`vertexos-dev-sandbox` is a MAC profile for running an AI coding agent (or any
untrusted dependency) with secrets and persistence out of reach. `vtx-sandbox`
is the launcher.

## What it allows / denies

- **Allows:** read-only system, toolchains (`rix`), read/write to the workspace
  (`~/projects`, `~/src`), build caches (`~/.npm`, `~/.cargo`, `~/go`, `~/.cache`),
  `/tmp`, and the network.
- **Denies:** reading `~/.ssh`, `~/.aws`, `~/.gnupg`, `~/.config/gh`, `.env`,
  browser cookies; writing `~/.bashrc`/autostart/cron/sudoers/`authorized_keys`;
  `mount`, `sys_module`, cross-process `ptrace`.

Tradeoff: git-over-SSH won't work *inside* the sandbox — that's the point. Push
from an unconfined shell or use a scoped token.

## Install & use

```bash
install -m0644 vertexos-dev-sandbox /etc/apparmor.d/vertexos-dev-sandbox
install -m0755 vtx-sandbox          /usr/local/bin/vtx-sandbox
apparmor_parser -r /etc/apparmor.d/vertexos-dev-sandbox

aa-complain vertexos-dev-sandbox    # start here: logs violations, blocks nothing
# … run your workflows, watch the audit log for denials, widen the profile …
aa-enforce  vertexos-dev-sandbox    # lock it down

vtx-sandbox claude                  # run the agent confined
```

Start in **complain** mode so you can see what a real workflow needs before you
enforce. Denials show up in the kernel audit log (and can be surfaced via Falco).
