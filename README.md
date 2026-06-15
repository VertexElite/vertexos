# VertexOS

Security-hardened Linux desktop. Void musl base, Windows 7 Aero UX, sensor-native.

**Status:** alpha — Phase 0 / ISO build pipeline working.

## What this is

A security distro built on the Void Linux musl base with a single opinionated stance:
the hardening, the sensors, and the boot path ship as the default. Not optional layers.

- **Base:** Void Linux musl (no glibc CVE class, rolling, independent)
- **Init:** runit
- **Kernel:** stock Void `linux` 6.18.x — already ships every KSPP flag Chimera Linux's `linux-lts` sets (`HARDENED_USERCOPY_DEFAULT_ON`, `SLAB_FREELIST_HARDENED`, `FORTIFY_SOURCE`, `STACKPROTECTOR_STRONG`, `MITIGATION_PAGE_TABLE_ISOLATION`, AppArmor built-in). Custom srcpkg with `ZERO_CALL_USED_REGS` planned for Phase 1.
- **DE:** XFCE4 + picom + Aero-Glass-XFCE4 theme
- **MAC:** AppArmor enforcing
- **Network:** nftables default-deny + XDP Sentinel hook
- **Sensors:** Falco / Suricata / Zeek / Wazuh / ntopng — aggregated into one Go TUI (`VertexSense`)

## Build

Requires Docker, QEMU+KVM, ~3GB disk.

```bash
# 0. Fetch submodules (void-mklive + Aero theme). REQUIRED — without this
#    build/void-mklive/ is empty and the Docker build silently produces a
#    broken image.
git submodule update --init --recursive

# 1. Build the builder container (one-time)
docker build -f build/Containerfile.vertex -t vertexos-builder:latest .

# 2. Build the ISO
./build/mklive-wrapper.sh

# 3. Boot test in QEMU
qemu-system-x86_64 -enable-kvm -m 2G -cdrom build/out/vertexos-alpha-*.iso
```

For headless verify: see `docs/build-and-verify.md`.

## Repo layout

```
build/        ISO build wrapper + void-mklive submodule
kernel/       Hardened sysctl drop-ins + future custom kernel config
packages/     VertexOS-specific srcpkgs (XBPS templates)
security/    AppArmor profiles, nftables, Falco rules, Suricata rules
ui/           XFCE defaults, picom config, Aero theme submodule, SDDM theme
installer/    Calamares branding + post-install hooks
vertexsense/  Go TUI sensor dashboard
docs/         Install guide, security model, threat model
```

## License

TBD — kernel/userland stay GPL/upstream. VertexOS-specific code targets MIT.

## Coordinated disclosure

Security issues: see SECURITY.md (TODO).
