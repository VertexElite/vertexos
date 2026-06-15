# installer/

Calamares branding + post-install hooks for the VertexOS live ISO.

Planned contents (Phase 2):

- `calamares/branding/vertexos/` — branding.desc, slideshow, logos.
- `calamares/settings.conf` — module sequence (partition, luks, users, …).
- `hooks/` — post-install scripts that drop the hardened sysctl, enable the
  nftables/AppArmor/usbguard runit services, and install the picom/XFCE
  defaults into the target root.

Empty for now — Phase 0 builds a live ISO only, no guided installer yet.
