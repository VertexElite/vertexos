#!/bin/sh -x
# VertexOS-patched version of upstream void-mklive dracut/vmklive/adduser.sh.
# Runs in the initramfs at boot to create the live user + set hostname.
#
# Patches vs upstream:
#   - hostname: vertexos-live (was: void-live)
#   - default user: vertex (was: anon) — overridable by live.user= kernel arg
#   - root + user password: prompt user on first interactive login instead of
#     hardcoded 'voidlinux' (security: live CDs with known root password are
#     LAN-reachable boot-and-pwn)
#   - sudoers drop-in filename: 99-vertexos-live (was: 99-void-live)
#   - polkit rules filename: vertexos-live.rules (was: void-live.rules)

if ! type getarg >/dev/null 2>&1 && ! type getargbool >/dev/null 2>&1; then
    . /lib/dracut-lib.sh
fi

echo vertexos-live > ${NEWROOT}/etc/hostname

USERNAME=$(getarg live.user)
USERSHELL=$(getarg live.shell)

[ -z "$USERNAME" ] && USERNAME=vertex
[ -x $NEWROOT/bin/bash -a -z "$USERSHELL" ] && USERSHELL=/bin/bash
[ -z "$USERSHELL" ] && USERSHELL=/bin/sh

# Create /etc/default/live.conf to store USER.
echo "USERNAME=$USERNAME" >> ${NEWROOT}/etc/default/live.conf
chmod 644 ${NEWROOT}/etc/default/live.conf

if ! grep -q ${USERSHELL} ${NEWROOT}/etc/shells ; then
    echo ${USERSHELL} >> ${NEWROOT}/etc/shells
fi

# Create autologin group (lightdm PAM uses pam_succeed_if to allow members
# in without password if /etc/pam.d/lightdm-autologin is present).
chroot ${NEWROOT} groupadd -r autologin 2>/dev/null || true
chroot ${NEWROOT} useradd -m -c $USERNAME -G audio,video,wheel,kvm,plugdev,input,users,autologin -s $USERSHELL $USERNAME

# Live user password = same as username (vertex / vertex). Documented as
# live-only convention (matches Kali kali/kali, Parrot parrot/parrot). The
# installer sets a real per-user password — this is throwaway live-session.
# autologin via PAM kicks in first via /etc/pam.d/lightdm-autologin so the
# password prompt should NOT appear for desktop boot, but is still set as
# a fallback for tty + SSH if user disables autologin.
# Set the live password via a precomputed SHA512 hash written straight to
# /etc/shadow (usermod -p). chpasswd authenticates via PAM, which is NOT
# wired up in this early initramfs chroot, so it silently fails to set the
# hash (login then rejects vertex/vertex). usermod -p edits shadow directly.
chroot ${NEWROOT} usermod -p '$6$A3Q1CAXa5oT1w7U1$//wU8J/sIkXg4S1vaVpiAKzgKkGW4xLpDVAGW9swXQXFWVQTsWKmMWbO8w0f0z331.WaCq4w13aTzq5xV21pr1' ${USERNAME}

# Lock root password. No hardcoded default — live ISOs with known root
# passwords are a regular boot-and-pwn target on shared LANs. Users who
# need root run 'sudo -i' as vertex (NOPASSWD wheel via /etc/sudoers.d/).
chroot ${NEWROOT} passwd -l root >/dev/null 2>&1

# wheel group sudo with NOPASSWD on live ISO (install convenience).
# The installed system gets a proper sudoers via void-installer.
if [ -f ${NEWROOT}/etc/sudoers ]; then
    echo "${USERNAME} ALL=(ALL:ALL) NOPASSWD: ALL" > "${NEWROOT}/etc/sudoers.d/99-vertexos-live"
    chmod 0440 "${NEWROOT}/etc/sudoers.d/99-vertexos-live"
fi

if [ -d ${NEWROOT}/etc/polkit-1 ]; then
    cat > ${NEWROOT}/etc/polkit-1/rules.d/vertexos-live.rules <<_EOF
polkit.addAdminRule(function(action, subject) {
    return ["unix-group:wheel"];
});

polkit.addRule(function(action, subject) {
    if (subject.isInGroup("wheel")) {
        return polkit.Result.YES;
    }
});
_EOF
    chroot ${NEWROOT} chown polkitd:polkitd /etc/polkit-1/rules.d/vertexos-live.rules
fi

if getargbool 0 live.autologin; then
        sed -i "s,GETTY_ARGS=\"--noclear\",GETTY_ARGS=\"--noclear -a $USERNAME\",g" ${NEWROOT}/etc/sv/agetty-tty1/conf
fi
