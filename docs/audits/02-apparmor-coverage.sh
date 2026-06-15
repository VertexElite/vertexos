#!/bin/bash
# A-Z audit — 02: AppArmor coverage.
# Reports: AppArmor LSM enabled, profile count, enforce/complain ratio,
# what's confined vs unconfined among running processes.

set -uo pipefail

cat <<EOF
# VertexOS — AppArmor Coverage Audit

Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ) UTC

## LSM state

EOF

if [ -f /sys/kernel/security/lsm ]; then
    echo "Active LSMs: \`$(cat /sys/kernel/security/lsm)\`"
else
    echo "**FAIL: /sys/kernel/security/lsm not readable — LSM disabled or kernel missing CONFIG_SECURITY?**"
fi

echo
echo "## AppArmor module"
echo
if [ -d /sys/kernel/security/apparmor ]; then
    echo "- Module loaded: ✅"
    echo "- Enabled flag: \`$(cat /sys/module/apparmor/parameters/enabled 2>/dev/null || echo unknown)\`"
    echo "- Hash/policy version: \`$(cat /sys/kernel/security/apparmor/policy/profiles | head -1 2>/dev/null || echo unknown)\`"
else
    echo "**FAIL: /sys/kernel/security/apparmor missing — AppArmor not loaded.**"
fi

echo
echo "## Profile inventory"
echo
if command -v aa-status >/dev/null 2>&1; then
    aa-status 2>&1 | sed -n '1,30p'
elif [ -d /sys/kernel/security/apparmor/profiles ]; then
    LOADED=$(wc -l < /sys/kernel/security/apparmor/profiles)
    ENFORCE=$(grep -c "(enforce)" /sys/kernel/security/apparmor/profiles)
    COMPLAIN=$(grep -c "(complain)" /sys/kernel/security/apparmor/profiles)
    KILL=$(grep -c "(kill)" /sys/kernel/security/apparmor/profiles)
    cat <<P
- Loaded profiles: ${LOADED}
- Enforce: ${ENFORCE}
- Complain: ${COMPLAIN}
- Kill: ${KILL}
P
    echo
    echo "First 20 profiles:"
    echo '```'
    head -20 /sys/kernel/security/apparmor/profiles
    echo '```'
else
    echo "(no aa-status, no /sys profiles dir — apparmor-utils probably missing)"
fi

echo
echo "## Process confinement coverage"
echo
TOTAL=$(ps -e -o pid= | wc -l)
if [ -d /sys/kernel/security/apparmor ]; then
    CONFINED=0
    UNCONFINED=0
    for pid in $(ps -e -o pid=); do
        attr="/proc/$pid/attr/current"
        [ -r "$attr" ] || continue
        val=$(cat "$attr" 2>/dev/null)
        case "$val" in
            unconfined*) UNCONFINED=$((UNCONFINED+1));;
            "") ;;
            *) CONFINED=$((CONFINED+1));;
        esac
    done
    PCT=$((CONFINED * 100 / (CONFINED + UNCONFINED + 1)))
    cat <<P
- Total processes: ${TOTAL}
- Confined: ${CONFINED}
- Unconfined: ${UNCONFINED}
- Confinement coverage: ${PCT}%
P
    echo
    echo "Top 10 unconfined processes (security-relevant?):"
    echo '```'
    for pid in $(ps -e -o pid=); do
        attr="/proc/$pid/attr/current"
        [ -r "$attr" ] || continue
        if grep -q "unconfined" "$attr" 2>/dev/null; then
            ps -p "$pid" -o pid,user,comm --no-headers 2>/dev/null
        fi
    done | head -10
    echo '```'
fi

echo
echo "## Recommendation"
echo
if [ -d /sys/kernel/security/apparmor ]; then
    if [ "${ENFORCE:-0}" -lt 5 ]; then
        echo "⚠️ Fewer than 5 enforce profiles loaded. Ship base profiles for: browsers, file manager, media players, network tools."
    fi
    if [ "${COMPLAIN:-0}" -gt 0 ]; then
        echo "⚠️ Complain-mode profiles present. These log but don't enforce — review and promote to enforce."
    fi
else
    echo "❌ AppArmor not active. Phase 1 task: enable AppArmor LSM via kernel cmdline."
fi
