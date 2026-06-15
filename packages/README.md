# packages/

VertexOS-specific XBPS source packages (`srcpkgs`).

Each subdirectory is a `void-packages`-style template (`template` file +
optional patches) for software VertexOS ships that isn't in the Void repos, or
that we rebuild with different flags.

Planned (Phase 1+):

- `vertex-kernel-config/` — stock `linux` rebuilt with `ZERO_CALL_USED_REGS`,
  `RANDSTRUCT_FULL`, lockdown LSM. See `docs/kernel-config.md`.
- `vertexsense/` — the Go TUI sensor dashboard (`ui`/`vertexsense/` source).
- `vertexos-branding/` — wallpapers, plymouth/SDDM theme, os-release.

Empty for now — Phase 0 ships no custom srcpkgs.
