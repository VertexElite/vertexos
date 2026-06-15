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

# Create new user. Groups include kvm and plugdev for hardware research.
chroot ${NEWROOT} useradd -m -c $USERNAME -G audio,video,wheel,kvm,plugdev,input,users -s $USERSHELL $USERNAME
chroot ${NEWROOT} passwd -d $USERNAME >/dev/null 2>&1

# Lock root + force-expire user password on first interactive login.
# No hardcoded 'voidlinux' default — live ISOs with known root passwords are
# a regular boot-and-pwn target on shared LAN segments.
chroot ${NEWROOT} passwd -l root >/dev/null 2>&1
chroot ${NEWROOT} passwd -e $USERNAME >/dev/null 2>&1

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
