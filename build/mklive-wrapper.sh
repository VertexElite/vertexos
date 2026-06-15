#!/usr/bin/env bash
# VertexOS ISO build wrapper.
# Drives void-mklive inside the vertexos-builder container.
#
# Usage:
#   ./build/mklive-wrapper.sh                 # builds dated output ISO
#   ./build/mklive-wrapper.sh -o my.iso       # custom output name

set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
OUT_DIR="$ROOT/build/out"
mkdir -p "$OUT_DIR"

# Mirror selection — primary upstream + working fallback.
# repo-default has been refusing connections during dev; repo-de is the
# current working baseline. Update this list if upstream becomes reachable.
MIRROR="https://repo-de.voidlinux.org"
ARCH="x86_64-musl"

PKGS=$(grep -vE '^\s*(#|$)' "$ROOT/build/packages.txt" | tr '\n' ' ')
OUT_NAME="vertexos-alpha-${ARCH}-$(date +%Y%m%d).iso"

# Allow flag override of output name.
while getopts "o:" opt; do
    case "$opt" in
        o) OUT_NAME="$OPTARG" ;;
        *) echo "Usage: $0 [-o output.iso]" >&2; exit 2 ;;
    esac
done

echo "==> VertexOS ISO build"
echo "    Arch:   $ARCH"
echo "    Mirror: $MIRROR"
echo "    Out:    $OUT_DIR/$OUT_NAME"
echo "    Pkgs:   $(echo $PKGS | wc -w) packages"

docker run --rm --privileged \
    -v "$ROOT/build/void-mklive:/mklive" \
    -v "$OUT_DIR:/out" \
    -w /mklive \
    vertexos-builder:latest \
    sh -c "./mklive.sh -a $ARCH \
        -p \"$PKGS\" \
        -r $MIRROR/current \
        -r $MIRROR/current/musl \
        -o /out/$OUT_NAME"

echo "==> Done: $OUT_DIR/$OUT_NAME"
ls -lh "$OUT_DIR/$OUT_NAME"
