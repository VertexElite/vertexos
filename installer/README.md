# installer/

Post-install branding + hooks for the VertexOS live ISO.

The live ISO installs to disk via Void's bundled **`void-installer`** — run
`sudo void-installer` from the live desktop, or use the **Install VertexOS**
desktop launcher. (Calamares is not packaged for Void; a graphical guided
installer would require building it from source — tracked as a future item.)

The hardened overlay is copied onto the target by void-installer's rootfs copy,
so the installed system inherits the VertexOS hardening + desktop by default.

Planned contents:

- `hooks/` — post-install scripts to re-assert the hardened sysctl, enable the
  nftables / AppArmor / usbguard runit services, and install the picom / XFCE
  defaults into the target root.
