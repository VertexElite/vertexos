# VertexOS live: boot straight to the desktop.
# For the live user 'vertex' on tty1 (no X yet), launch the session.
if [ -z "${DISPLAY:-}" ] && [ "$(tty 2>/dev/null)" = "/dev/tty1" ] && [ "$(id -un 2>/dev/null)" = "vertex" ]; then
    exec startx
fi
