#!/bin/bash
# Run all 4 A-Z audit scripts and concatenate into one report.
# Usage:  sudo ./run-all.sh > audit-$(date +%Y%m%d).md
set -uo pipefail
DIR=$(cd "$(dirname "$0")" && pwd)

for script in \
    01-kernel-hardening.sh \
    02-apparmor-coverage.sh \
    03-network-exposure.sh \
    04-privesc-paths.sh ; do
    [ -f "$DIR/$script" ] || { echo "MISSING: $script"; continue; }
    bash "$DIR/$script"
    echo
    echo "---"
    echo
done
