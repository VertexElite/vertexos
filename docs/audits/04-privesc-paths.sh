#!/bin/bash
# A-Z audit — 04: Privilege escalation surface.
# Maps SUID binaries, file capabilities, sudoers state, world-writable risks,
# kernel exploit prerequisites that survived our sysctl hardening.

set -uo pipefail

cat <<EOF
# VertexOS — Privilege Escalation Surface

Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ) UTC

## SUID/SGID binaries (system-wide)

\`\`\`
$(find / -xdev \( -perm -4000 -o -perm -2000 \) -type f -printf "%M %u %g %p\n" 2>/dev/null | sort | head -40)
\`\`\`

Total SUID/SGID files: $(find / -xdev \( -perm -4000 -o -perm -2000 \) -type f 2>/dev/null | wc -l)

> Each entry is a privilege-elevation target. Minimal Linux ships <10. Anything >20 = audit gap.

## File capabilities (cap-set binaries)

\`\`\`
$(getcap -r / 2>/dev/null | head -20)
\`\`\`

> These bypass SUID and give specific kernel privileges. \`cap_sys_admin\` and \`cap_net_admin\` are particularly sensitive.

## Sudoers state

\`\`\`
$(grep -v "^#\|^$" /etc/sudoers /etc/sudoers.d/* 2>/dev/null | head -20)
\`\`\`

## World-writable directories outside /tmp /var/tmp /dev/shm

\`\`\`
$(find / -xdev -type d -perm -0002 -not \( -path /tmp -o -path /var/tmp -o -path /dev/shm -o -path /run/lock -o -path "/var/cache/*" -o -path "/proc/*" -o -path "/sys/*" \) 2>/dev/null | head -20)
\`\`\`

## World-writable files (excluding sockets/devices)

\`\`\`
$(find / -xdev -type f -perm -0002 2>/dev/null | head -20)
\`\`\`

## Sensitive directory perms

EOF

for path in /etc/shadow /etc/sudoers /etc/ssh /root /var/log /etc/cron.d; do
    if [ -e "$path" ]; then
        perms=$(stat -c "%a %U:%G" "$path")
        echo "- \`$path\` → $perms"
    fi
done

cat <<EOF

## Kernel exploit prerequisites (status check)

| Prerequisite | State | Why it matters |
|---|---|---|
EOF

# Useful checks for typical LPE
KEXP=(
    "kernel.unprivileged_userns_clone unprivileged_userns_clone"
    "kernel.unprivileged_bpf_disabled unprivileged_bpf_disabled"
    "kernel.kexec_load_disabled kexec_load_disabled"
    "kernel.dmesg_restrict dmesg_restrict"
    "kernel.kptr_restrict kptr_restrict"
    "kernel.perf_event_paranoid perf_event_paranoid"
    "fs.suid_dumpable suid_dumpable"
)

for entry in "${KEXP[@]}"; do
    name="${entry%% *}"
    val=$(sysctl -n "$name" 2>/dev/null || echo "absent")
    desc=""
    case "$name" in
        kernel.unprivileged_userns_clone) desc="0 prevents user-namespace exploits";;
        kernel.unprivileged_bpf_disabled) desc="1 closes JIT-spray / verifier LPE class";;
        kernel.kexec_load_disabled) desc="1 prevents replacing kernel post-boot";;
        kernel.dmesg_restrict) desc="1 hides addresses leaked via printk";;
        kernel.kptr_restrict) desc="2 hides kernel pointers from non-root";;
        kernel.perf_event_paranoid) desc="3 disables perf_event_open for non-root";;
        fs.suid_dumpable) desc="0 prevents core-dump LPE";;
    esac
    printf "| \`%s\` | %s | %s |\n" "$name" "$val" "$desc"
done

cat <<EOF

## Group memberships that grant elevation

\`\`\`
$(for grp in wheel sudo adm root disk video audio docker kvm; do
    members=$(getent group "$grp" 2>/dev/null | cut -d: -f4)
    [ -n "$members" ] && echo "$grp: $members"
done)
\`\`\`

## Cron jobs (system + per-user)

\`\`\`
$(ls -la /etc/cron* 2>&1 | head -20)
\`\`\`

## Recommendation

EOF

SUID_COUNT=$(find / -xdev \( -perm -4000 -o -perm -2000 \) -type f 2>/dev/null | wc -l)
if [ "$SUID_COUNT" -gt 20 ]; then
    echo "- ⚠️ ${SUID_COUNT} SUID/SGID binaries is high. Each is an LPE target. Audit which can be removed (e.g. \`mount\`, \`fusermount\`, \`unix_chkpwd\`)."
fi

USERNS=$(sysctl -n kernel.unprivileged_userns_clone 2>/dev/null)
if [ "$USERNS" = "1" ]; then
    echo "- 🚨 \`kernel.unprivileged_userns_clone=1\` — open to user-namespace LPE class. Set to 0 unless rootless containers needed."
fi
