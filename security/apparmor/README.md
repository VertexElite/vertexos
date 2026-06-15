# security/apparmor/

AppArmor profiles shipped by VertexOS, installed to `/etc/apparmor.d/`.

VertexOS runs AppArmor in **enforcing** mode by default. Profiles here tighten
over the upstream `apparmor-profiles` package for the browsers, mail clients,
and network tools we ship.

Empty for now — Phase 0 enables the AppArmor LSM but ships no custom profiles.
