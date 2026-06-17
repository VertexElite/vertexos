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

## Auto-wrap + FUSE context-seal

`vtx-autowrap on` installs per-user shims in `~/.local/bin` so coding agents
(`claude`/`gemini`/`cursor`/...) run under the sandbox automatically — no
`vtx-sandbox` prefix. `vtx-agent-shim` resolves the real binary and re-execs it
confined; `VTX_SANDBOXED` guards recursion, `VTX_NO_SANDBOX=1` opts out.

`vtx-sandbox` also mounts a **FUSE context-seal** (`vtx-ctx-seal`): a passthrough
mirror of `$HOME` with credential paths (`.ssh`, `.aws`, `.gnupg`, `.config/gh`,
`*.env`, browser stores, ...) redacted to ENOENT, mounted under `/tmp` and handed
to the agent as `$HOME`. Because VertexOS sets `user.max_user_namespaces=0`,
bind/bwrap sealing can't run unprivileged — FUSE (via setuid `fusermount3`) is
the one path that can. AppArmor stays the hard backstop. Skip with `VTX_NO_SEAL=1`.
