#!/usr/bin/env bash
# VertexOS ISO build wrapper — full branding + hardening pipeline.
#
# Calls mkiso.sh -b xfce (NOT mklive.sh) so the resulting ISO is actually
# installable (mkiso bundles void-installer + lightdm autostart that mklive
# alone doesn't).
#
# Before invoking mkiso:
#   - applies VertexOS branding patches to the void-mklive submodule (in the
#     docker container only — submodule on host stays clean)
#   - patches installer.sh / mklive.sh strings via sed
#   - replaces dracut/vmklive/adduser.sh with our patched version that writes
#     vertexos-live (not void-live) to hostname + locks root password
#   - replaces data/issue with VertexOS branded banner
#
# mkiso invocation:
#   - -T "VertexOS"  bootloader title
#   - -I /overlay    rootfs overlay (branded /etc, audit rules, sysctl, etc.)
#   - -C "<args>"    kernel command line hardening (apparmor=1 lockdown=integrity)
#
# Our extra package list is appended via mkiso's pass-through (-- -p "...").
#
# Usage:
#   ./build/mklive-wrapper.sh                 # builds dated output ISO
#   ./build/mklive-wrapper.sh -o my.iso       # custom output name

set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
OUT_DIR="$ROOT/build/out"
OVERLAY="$ROOT/build/rootfs-overlay"
PATCHES="$ROOT/build/vertex-vmklive-patches"
mkdir -p "$OUT_DIR"

MIRROR="https://repo-de.voidlinux.org"   # repo-default has been refusing; revisit later
ARCH="x86_64-musl"
VARIANT="xfce"
BOOT_TITLE="VertexOS"

# Kernel command line hardening — applied to every boot entry (live + UEFI).
# Sources for each flag are in docs/threat-model.md.
KERNEL_CMDLINE="apparmor=1 security=apparmor lockdown=integrity \
slab_nomerge init_on_alloc=1 init_on_free=1 page_alloc.shuffle=1 \
randomize_kstack_offset=on vsyscall=none debugfs=off \
ipv6.disable=0 module.sig_enforce=0 \
rd.shell=0 rd.emergency=halt panic=10 oops=panic \
live.user=vertex live.autologin"
#   apparmor=1 security=apparmor: ensure AppArmor LSM active (CVE-2026-23268..23411 etc)
#   lockdown=integrity: block kernel image modification post-boot
#   slab_nomerge: each kmem_cache is unique (mitigates cross-cache exploits)
#   init_on_alloc/free: zero kernel memory on alloc + free (info leak mitigation)
#   page_alloc.shuffle: page allocator entropy
#   randomize_kstack_offset: per-syscall kernel stack offset randomization
#   vsyscall=none: kill vsyscall page (legacy attack surface)
#   debugfs=off: no debugfs (info leaks)
#   module.sig_enforce=0: live ISO needs unsigned mods for hardware probing; install-time will toggle on
#   rd.shell=0 rd.emergency=halt: block initramfs debug shell (Secure Boot bypass class)
#   panic=10 oops=panic: hang on panic for inspection; never silently recover
#   live.user=vertex live.autologin: dracut adduser pulls username from kernel cmdline

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
echo "    Title:   $BOOT_TITLE"
echo "    Extra:   $(echo $EXTRA_PKGS | wc -w) packages on top of variant base"
echo "    Overlay: $(find $OVERLAY -type f | wc -l) files in rootfs overlay"
echo "    Patches: $(ls $PATCHES | wc -l) vmklive patches"
echo "    Out:     $OUT_DIR/$OUT_NAME"
echo

docker run --rm --privileged \
    -v "$ROOT/build/void-mklive:/mklive" \
    -v "$OUT_DIR:/out" \
    -v "$OVERLAY:/overlay:ro" \
    -v "$PATCHES:/patches:ro" \
    -w /mklive \
    vertexos-builder:latest \
    sh -c "
        set -e

        echo '==> Applying VertexOS branding patches to void-mklive ...'

        # Overlay our patched adduser.sh (kills the void-live hostname overwrite + root pw)
        cp /patches/dracut-vmklive-adduser.sh dracut/vmklive/adduser.sh
        chmod +x dracut/vmklive/adduser.sh

        # Replace data/issue used by upstream live system pre-overlay
        cp /patches/data-issue data/issue

        # Patch installer.sh title strings
        sed -i 's|Void Linux installation -- https://www.voidlinux.org|VertexOS installation — https://github.com/VertexElite/vertexos|g' installer.sh
        sed -i 's|Void Linux installation menu|VertexOS installation menu|g' installer.sh
        sed -i 's|Welcome to the Void Linux installation|Welcome to the VertexOS installation|g' installer.sh
        sed -i 's|Void Linux has been installed successfully|VertexOS has been installed successfully|g' installer.sh
        sed -i 's|the Void Linux Handbook|the VertexOS docs (github.com/VertexElite/vertexos)|g' installer.sh
        sed -i 's|the Void IRC channel|GitHub issues|g' installer.sh

        # Patch ISO volume label
        sed -i 's|VOID_LIVE|VERTEXOS_LIVE|g' isolinux/isolinux.cfg.in mklive.sh

        # mklive.sh default BOOT_TITLE
        sed -i 's|BOOT_TITLE=\"Void Linux\"|BOOT_TITLE=\"VertexOS\"|g' mklive.sh

        echo '==> Running mkiso.sh with overlay + hardened cmdline ...'
        ./mkiso.sh -a $ARCH -b $VARIANT -d $DATE \\
            -r $MIRROR/current \\
            -r $MIRROR/current/musl \\
            -- -p \"$EXTRA_PKGS\" \\
               -I /overlay \\
               -T \"$BOOT_TITLE\" \\
               -C \"$KERNEL_CMDLINE\" \\
               -o /out/$OUT_NAME
    "

echo
if [ -f "$OUT_DIR/$OUT_NAME" ]; then
    echo "==> Done: $OUT_DIR/$OUT_NAME"
    ls -lh "$OUT_DIR/$OUT_NAME"
    cd "$OUT_DIR" && sha256sum "$OUT_NAME" | tee "${OUT_NAME}.sha256"
else
    echo "==> Build failed — no ISO produced at $OUT_DIR/$OUT_NAME" >&2
    ls -la "$OUT_DIR" >&2
    exit 1
fi
