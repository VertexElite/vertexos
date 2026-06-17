# VertexOS live: start the desktop on the console (fallback if /etc/profile.d
# isn't sourced for this shell). Guarded to tty1 so SSH/other ttys are unaffected.
if [ -z "${DISPLAY:-}" ] && [ "$(tty 2>/dev/null)" = "/dev/tty1" ]; then
    exec startx
fi
