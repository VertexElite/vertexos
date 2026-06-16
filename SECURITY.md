# Security Policy

VertexOS is a security-hardened Void Linux (musl) distribution whose **primary
defensive target is supply-chain and AI-agent-abuse attacks** against developers
and the AI coding agents they run. We take security reports seriously and
appreciate responsible, coordinated disclosure.

## Supported versions

VertexOS is pre-1.0 (alpha). Security fixes land on `main` and ship in the next
ISO build; there are no long-term-support branches yet.

| Version            | Supported                  |
|--------------------|----------------------------|
| `main` (latest)    | ✅                          |
| `v0.1.x-alpha`     | ✅ (current release line)   |
| older pre-releases | ❌                          |

## Reporting a vulnerability

**Please report privately — do not open a public issue for a security bug.**

Preferred channel: **GitHub Security Advisories** — open the repository's
[Security tab](https://github.com/VertexElite/vertexos/security/advisories) →
**Report a vulnerability**. This creates a private advisory visible only to you
and the maintainers.

Please include:

- the affected component (overlay path, `vertexsense`, `cve-watch`, kernel
  config, an AppArmor / nftables / Suricata / AIDE rule, etc.) and the version
  or ISO SHA-256 it reproduces on,
- a clear description and impact — what an attacker gains,
- reproduction steps or a proof-of-concept, and
- any suggested mitigation.

Hardening bugs specific to this project are explicitly **in scope and welcome**:
a defense that ships *disarmed*, a blacklist or sysctl that never applies, an
AppArmor profile that fails to enforce, a module-blacklist that breaks boot.

## Disclosure process & timeline

We follow coordinated disclosure:

1. **Acknowledge** your report within **72 hours**.
2. **Triage and confirm** severity within **7 days**.
3. **Fix** developed on a private branch; you are credited unless you ask
   otherwise.
4. **Coordinated public disclosure** once a fixed ISO/commit is available, or at
   **90 days** from the initial report — whichever comes first. We will agree an
   embargo window with you.

Critical, actively-exploited issues are fast-tracked. Our daily `cve-watch`
radar already tracks weaponized CVEs against the shipped stack
(`docs/cve-watch/`), so an upstream-component CVE may already be classified in
[`docs/cve-mitigation-matrix.md`](docs/cve-mitigation-matrix.md).

## Scope

**In scope:** the VertexOS overlay, `vertexsense`, `cve-watch`, audit scripts,
the AppArmor / nftables / Suricata / AIDE rule-sets, the kernel config, and the
build pipeline in this repository.

**Out of scope:** vulnerabilities in upstream Void Linux packages or the Linux
kernel itself — please report those to their respective upstreams. A heads-up is
still welcome so we can add a mitigation to the matrix.

## Safe harbor

Good-faith security research conducted under this policy — testing your own
installs, avoiding privacy violations and service disruption, and giving us
reasonable time to remediate before any public disclosure — is welcome and will
not be pursued as a violation.
