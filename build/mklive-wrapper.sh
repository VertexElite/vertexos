#!/usr/bin/env bash
# VertexOS ISO build wrapper.
# Calls mkiso.sh -b xfce (NOT mklive.sh) so the resulting ISO is actually
# installable — mkiso bundles void-installer and lightdm autostart that
# mklive.sh by itself doesn't.
#
# Our security stack (apparmor, nftables, audit, clamav, etc.) is appended
# via -- -p "..." which passes through to mklive.
#
# Usage:
#   ./build/mklive-wrapper.sh                 # builds dated output ISO
#   ./build/mklive-wrapper.sh -o my.iso       # custom output name

set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
OUT_DIR="$ROOT/build/out"
mkdir -p "$OUT_DIR"

MIRROR="https://repo-de.voidlinux.org"   # repo-default has been refusing; revisit later
ARCH="x86_64-musl"
VARIANT="xfce"

# Our additive package list — security stack + desktop polish on top of the
# xfce variant. mkiso provides xfce4, lightdm, NetworkManager, dbus, polkit,
# udisks2, firefox, gvfs-* already.
EXTRA_PKGS=$(grep -vE '^\s*(#|$)' "$ROOT/build/packages.txt" | tr '\n' ' ')
DATE=$(date +%Y%m%d)
OUT_NAME="vertexos-alpha-${ARCH}-${DATE}.iso"

while getopts "o:" opt; do
    case "$opt" in
        o) OUT_NAME="$OPTARG" ;;
        *) echo "Usage: $0 [-o output.iso]" >&2; exit 2 ;;
    esac
done

echo "==> VertexOS ISO build"
echo "    Arch:    $ARCH"
echo "    Variant: $VARIANT  (XFCE + lightdm autologin + NetworkManager)"
echo "    Mirror:  $MIRROR"
echo "    Extra:   $(echo $EXTRA_PKGS | wc -w) packages on top of variant base"
echo "    Out:     $OUT_DIR/$OUT_NAME"
echo

# mkiso.sh produces a dated filename: void-live-<arch>-<date>-<variant>.iso
# We rename it to our convention at the end.
INNER_NAME="void-live-${ARCH}-${DATE}-${VARIANT}.iso"

docker run --rm --privileged \
    -v "$ROOT/build/void-mklive:/mklive" \
    -v "$OUT_DIR:/out" \
    -w /mklive \
    vertexos-builder:latest \
    sh -c "./mkiso.sh -a $ARCH -b $VARIANT -d $DATE \
        -r $MIRROR/current \
        -r $MIRROR/current/musl \
        -- -p \"$EXTRA_PKGS\" -o /out/$OUT_NAME"

echo
if [ -f "$OUT_DIR/$OUT_NAME" ]; then
    echo "==> Done: $OUT_DIR/$OUT_NAME"
    ls -lh "$OUT_DIR/$OUT_NAME"
elif [ -f "$OUT_DIR/$INNER_NAME" ]; then
    mv "$OUT_DIR/$INNER_NAME" "$OUT_DIR/$OUT_NAME"
    echo "==> Done (renamed from $INNER_NAME): $OUT_DIR/$OUT_NAME"
    ls -lh "$OUT_DIR/$OUT_NAME"
else
    echo "==> Build failed — no ISO produced." >&2
    exit 1
fi
