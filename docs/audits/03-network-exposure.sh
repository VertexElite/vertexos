#!/bin/bash
# A-Z audit — 03: Network exposure.
# Reports: nftables ruleset state, listening sockets, attack surface.

set -uo pipefail

cat <<EOF
# VertexOS — Network Exposure Audit

Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ) UTC

## Listening sockets (ss -tulnp)

\`\`\`
$(ss -tulnp 2>&1 | head -40)
\`\`\`

## Externally-reachable TCP listeners (0.0.0.0 / :: only)

\`\`\`
$(ss -tlnp 2>&1 | awk 'NR==1 || $5 ~ /^(0\.0\.0\.0|\[::\]):/ { print }')
\`\`\`

## nftables ruleset

\`\`\`
$(nft list ruleset 2>&1 | head -80)
\`\`\`

EOF

echo "## Counter summary"
echo
if nft list ruleset >/dev/null 2>&1; then
    DEFAULT_INPUT_POLICY=$(nft list ruleset | grep -E "chain input" -A1 | grep "policy" | awk '{print $NF}' | tr -d ';')
    DEFAULT_FWD_POLICY=$(nft list ruleset | grep -E "chain forward" -A1 | grep "policy" | awk '{print $NF}' | tr -d ';')
    cat <<P
- Default input policy: \`${DEFAULT_INPUT_POLICY:-unknown}\`  $([ "$DEFAULT_INPUT_POLICY" = "drop" ] && echo "✅" || echo "⚠️ should be 'drop' for security")
- Default forward policy: \`${DEFAULT_FWD_POLICY:-unknown}\`  $([ "$DEFAULT_FWD_POLICY" = "drop" ] && echo "✅" || echo "⚠️")
P
else
    echo "**FAIL: nftables not active.**"
fi

echo
echo "## Per-interface state"
echo
echo '```'
ip -brief addr show 2>&1
echo '```'

echo
echo "## DNS resolver chain"
echo
echo '```'
cat /etc/resolv.conf 2>&1
echo '```'
if pgrep -f stubby >/dev/null; then
    echo "✅ stubby (DoT) running"
else
    echo "⚠️ stubby NOT running — DNS over TLS not active"
fi

echo
echo "## Active routes"
echo
echo '```'
ip route 2>&1
echo '```'

echo
echo "## Established connections (snapshot)"
echo
echo '```'
ss -tnp state established 2>&1 | head -20
echo '```'

echo
echo "## Recommendation"
echo
LISTEN_COUNT=$(ss -tlnp 2>/dev/null | awk 'NR>1 && $5 ~ /^(0\.0\.0\.0|\[::\]):/' | wc -l)
echo "- Externally-reachable listeners: ${LISTEN_COUNT}"
if [ "$LISTEN_COUNT" -gt 1 ]; then
    echo "  ⚠️ Each external listener is attack surface. Review whether each is intentional for VertexOS desktop posture."
fi
if [ "${DEFAULT_INPUT_POLICY:-}" != "drop" ]; then
    echo "- 🚨 INPUT default policy is not drop. Ship /etc/nftables.conf with default-drop policy."
fi
