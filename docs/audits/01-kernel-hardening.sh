#!/bin/bash
# A-Z audit — 01: Kernel hardening verification.
# Compares running kernel CONFIG flags against KSPP recommended set + Chimera
# Linux's curated hardening profile. Dumps live sysctl values vs our
# 99-vertex-harden.conf intent. Emits human-readable markdown + a JSON sidecar.
#
# Run as root on the running system:
#   sudo ./01-kernel-hardening.sh > kernel-hardening.md

set -uo pipefail
OUT_JSON="${1:-/tmp/audit-kernel.json}"

CONFIG="/proc/config.gz"
[ -f "$CONFIG" ] || CONFIG="/boot/config-$(uname -r)"
[ -f "$CONFIG" ] || { echo "FATAL: no kernel config readable (no /proc/config.gz, no /boot/config-*)"; exit 1; }
CFG_CMD="zcat"; [[ "$CONFIG" == *.gz ]] || CFG_CMD="cat"

cfg() {
    local key="$1"
    $CFG_CMD "$CONFIG" 2>/dev/null | grep -E "^${key}=|^# ${key} " | head -1
}

# Render a markdown table row + capture JSON entry.
report() {
    local key="$1" target="$2" actual="$3" status="$4"
    printf "| %-44s | %-15s | %-15s | %s |\n" "$key" "$target" "$actual" "$status"
    JSON_ENTRIES+=("$(printf '{"key":"%s","target":"%s","actual":"%s","status":"%s"}' "$key" "$target" "$actual" "$status")")
}

JSON_ENTRIES=()

cat <<EOF
# VertexOS — Kernel Hardening Audit

Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ) UTC
Host: $(hostname)
Kernel: $(uname -srm)
Config source: ${CONFIG}

## CONFIG flags (KSPP + Chimera Linux baseline)

| Flag | Target | Actual | Status |
|------|--------|--------|--------|
EOF

# (flag, target_value)
KSPP_FLAGS=(
    "CONFIG_BPF_UNPRIV_DEFAULT_OFF y"
    "CONFIG_SLAB_FREELIST_HARDENED y"
    "CONFIG_SLAB_FREELIST_RANDOM y"
    "CONFIG_INIT_ON_ALLOC_DEFAULT_ON y"
    "CONFIG_INIT_ON_FREE_DEFAULT_ON y_or_off"
    "CONFIG_FORTIFY_SOURCE y"
    "CONFIG_HARDENED_USERCOPY y"
    "CONFIG_HARDENED_USERCOPY_DEFAULT_ON y"
    "CONFIG_LIST_HARDENED y"
    "CONFIG_STACKPROTECTOR_STRONG y"
    "CONFIG_MITIGATION_PAGE_TABLE_ISOLATION y"
    "CONFIG_MITIGATION_RETPOLINE y"
    "CONFIG_SECURITY_APPARMOR y"
    "CONFIG_SECCOMP y"
    "CONFIG_RANDOMIZE_BASE y"
    "CONFIG_RANDOMIZE_MEMORY y"
    "CONFIG_ZERO_CALL_USED_REGS y_optional"
    "CONFIG_RANDSTRUCT_FULL y_optional"
    "CONFIG_SECURITY_LOCKDOWN_LSM y_optional"
)

PASS=0; FAIL=0; OPTIONAL_MISS=0
for entry in "${KSPP_FLAGS[@]}"; do
    key="${entry%% *}"; target="${entry##* }"
    line=$(cfg "$key")
    if [[ "$line" == *"=y"* ]]; then
        actual=y
        status="✅ PASS"; PASS=$((PASS+1))
    elif [[ "$line" == *"is not set"* ]]; then
        actual=n
        if [[ "$target" == "y_optional" || "$target" == "y_or_off" ]]; then
            status="⚠️ OPT-MISS"; OPTIONAL_MISS=$((OPTIONAL_MISS+1))
        else
            status="❌ FAIL"; FAIL=$((FAIL+1))
        fi
    else
        actual="missing"
        status="❌ MISSING"; FAIL=$((FAIL+1))
    fi
    report "$key" "$target" "$actual" "$status"
done

cat <<EOF

**Summary:** ${PASS} pass · ${FAIL} fail · ${OPTIONAL_MISS} optional-miss

---

## Live sysctl values (vs our intent in kernel/sysctl/99-vertex-harden.conf)

| Sysctl | Our target | Running value | Status |
|--------|-----------|---------------|--------|
EOF

SYSCTL_TARGETS=(
    "kernel.kptr_restrict 2"
    "kernel.dmesg_restrict 1"
    "kernel.kexec_load_disabled 1"
    "kernel.unprivileged_bpf_disabled 1"
    "net.core.bpf_jit_harden 2"
    "kernel.yama.ptrace_scope 2"
    "kernel.perf_event_paranoid 3"
    "fs.protected_hardlinks 1"
    "fs.protected_symlinks 1"
    "fs.protected_fifos 2"
    "fs.protected_regular 2"
    "fs.suid_dumpable 0"
    "net.ipv4.conf.all.rp_filter 1"
    "net.ipv4.tcp_syncookies 1"
    "net.ipv4.conf.all.accept_redirects 0"
    "net.ipv4.conf.all.accept_source_route 0"
    "net.ipv4.icmp_echo_ignore_broadcasts 1"
    "net.ipv4.tcp_rfc1337 1"
    "kernel.randomize_va_space 2"
    "kernel.unprivileged_userns_clone 0"
)

SCT_PASS=0; SCT_FAIL=0
for entry in "${SYSCTL_TARGETS[@]}"; do
    key="${entry% *}"; target="${entry##* }"
    actual=$(sysctl -n "$key" 2>/dev/null)
    if [ -z "$actual" ]; then
        status="⚠️ ABSENT"
    elif [ "$actual" = "$target" ]; then
        status="✅ PASS"; SCT_PASS=$((SCT_PASS+1))
    else
        status="❌ DRIFT"; SCT_FAIL=$((SCT_FAIL+1))
    fi
    printf "| %-46s | %-10s | %-13s | %s |\n" "$key" "$target" "${actual:-N/A}" "$status"
done

cat <<EOF

**Summary:** ${SCT_PASS} match · ${SCT_FAIL} drift

---

## Mitigations active per /sys/devices/system/cpu/vulnerabilities

\`\`\`
$(for f in /sys/devices/system/cpu/vulnerabilities/*; do
    [ -f "$f" ] && printf "%-26s %s\n" "$(basename "$f"):" "$(cat "$f")"
done)
\`\`\`

## Boot command line (/proc/cmdline)

\`\`\`
$(cat /proc/cmdline)
\`\`\`

## ASLR state

- mmap_base randomization: $(cat /proc/sys/kernel/randomize_va_space)
- KASLR seed (kernel CONFIG_RANDOMIZE_BASE active): see CONFIG table above
- /proc/kallsyms zeroed for non-root (kptr_restrict): $(grep -c " 0000000000000000 " /proc/kallsyms || echo "0") of $(wc -l < /proc/kallsyms) lines

EOF

# JSON sidecar
{
    echo '{'
    echo '  "generated":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",'
    echo '  "kernel":"'$(uname -srm)'",'
    echo '  "config_flags":['
    printf '    %s' "${JSON_ENTRIES[@]}" | sed 's/}{/},\n    {/g'
    echo
    echo '  ]'
    echo '}'
} > "$OUT_JSON" 2>/dev/null || true
