# VertexOS Kernel Configuration

## TL;DR — Why we use stock `linux`, not `linux-hardened`

Void Linux ships no `linux-hardened` package. Chimera Linux (musl, hardening-focused) does not either — their hardening lives in the `linux-lts` build config.

We verified (June 15, 2026) that **every KSPP / Chimera hardening flag is already enabled in Void's stock `linux` 6.18.35** built with their default config.

## Flag verification (Void stock 6.18.35)

```
CONFIG_BPF_UNPRIV_DEFAULT_OFF=y              ✅
CONFIG_SLAB_FREELIST_HARDENED=y              ✅
CONFIG_SLAB_FREELIST_RANDOM=y                ✅
CONFIG_INIT_ON_ALLOC_DEFAULT_ON=y            ✅
CONFIG_FORTIFY_SOURCE=y                      ✅
CONFIG_HARDENED_USERCOPY=y                   ✅
CONFIG_HARDENED_USERCOPY_DEFAULT_ON=y        ✅
CONFIG_LIST_HARDENED=y                       ✅
CONFIG_STACKPROTECTOR_STRONG=y               ✅
CONFIG_MITIGATION_PAGE_TABLE_ISOLATION=y     ✅ (PTI)
CONFIG_MITIGATION_RETPOLINE=y                ✅
CONFIG_SECURITY_APPARMOR=y                   ✅ (built-in)
CONFIG_SECCOMP=y                             ✅

# Delta vs anthraxx/linux-hardened
CONFIG_ZERO_CALL_USED_REGS                   ❌ not set
CONFIG_RANDSTRUCT_FULL                       ❌ NONE selected
CONFIG_INIT_ON_FREE_DEFAULT_ON               ❌ not set (perf trade-off, Chimera also disables)
```

## Phase 1 plan: `vertex-kernel-config` srcpkg

Custom XBPS srcpkg that re-builds `linux6.18` (or current LTS) with these additional flags:

```
CONFIG_ZERO_CALL_USED_REGS=y
CONFIG_RANDSTRUCT_FULL=y       # adds GCC plugin build dep
CONFIG_INIT_ON_FREE_DEFAULT_ON=y  # measurable perf hit; pending bench
```

Plus `CONFIG_SECURITY_LOCKDOWN_LSM=y` + `CONFIG_SECURITY_LOCKDOWN_LSM_EARLY=y` for full lockdown mode by default.

Until that srcpkg lands, stock `linux` is the supported kernel.

## Runtime hardening layer

The sysctl drop-in `kernel/sysctl/99-vertex-harden.conf` tightens over Void stock:

| Sysctl | Void stock | VertexOS |
|---|---|---|
| `kernel.kptr_restrict` | 1 | **2** |
| `kernel.yama.ptrace_scope` | 1 | **2** |
| `kernel.perf_event_paranoid` | 2 | **3** |
| `net.ipv4.conf.all.rp_filter` | unset | **1** |
| `net.ipv4.tcp_syncookies` | unset | **1** |
| `kernel.unprivileged_userns_clone` | varies | **0** |

Plus the full ICMP/source-route/redirect blockade and IPv6 RA disable.
