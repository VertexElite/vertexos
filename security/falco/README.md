# Falco rules — host runtime detection

`vertexos-rules.yaml` is the syscall-level detection layer, tuned for
supply-chain compromise and AI coding-agent abuse (not generic container noise).

## Install

```bash
install -m0644 vertexos-rules.yaml /etc/falco/rules.d/vertexos-rules.yaml
# /etc/falco/falco.yaml must enable JSON output to the path VertexSense tails:
#   json_output: true
#   file_output:
#     enabled: true
#     filename: /var/log/falco/events.json
falco -V /etc/falco/rules.d/vertexos-rules.yaml   # validate
sv restart falco                                   # runit
```

## What it catches

| Rule | Priority | Scenario |
|------|----------|----------|
| Supply-chain net tool in package install | WARNING | `curl`/`nc` run by an `npm`/`pip`/… postinstall |
| Package install reads credentials | CRITICAL | dependency reads `~/.ssh`, `~/.aws`, `.npmrc`, `.env`… |
| Coding agent reads credentials | CRITICAL | agent process touches secrets it shouldn't |
| Reverse shell pattern | CRITICAL | `bash -i` + `/dev/tcp`, `nc -e`, `socat exec:` |
| Pipe to shell execution | WARNING | `curl … \| sh` |
| Persistence write | WARNING | writes to cron/rc/systemd-user/sudoers/git-hooks |
| Setuid bit set | ERROR | new setuid-root backdoor |
| Execution from temp dir | WARNING | binary run from `/tmp`, `/dev/shm`, `/var/tmp` |
| Crypto miner indicator | CRITICAL | miner binary or `stratum+tcp` |
| Shadow file read | CRITICAL | `/etc/shadow` read by an unexpected process |

Rules are tagged (`supply_chain`, `agent`, `credential_access`, `persistence`,
MITRE technique IDs) so you can filter or route them. Tune with per-workload
exceptions as needed — alpha defaults err toward visibility.
