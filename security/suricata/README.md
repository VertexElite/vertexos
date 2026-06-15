# Suricata rules — network detection

`vertexos.rules` covers the network side of supply-chain/agent compromise:
exfiltration to anonymous file-drops, crypto-mining handshakes, DNS tunnelling,
scripted/CLI fetches, and package pulls from raw IPs. SIDs use the local
`9.10x.xxx` range.

## Install

```bash
install -m0644 vertexos.rules /etc/suricata/rules/vertexos.rules
```

Add the rule file and make sure EVE JSON is on (VertexSense tails `eve.json`):

```yaml
# /etc/suricata/suricata.yaml
rule-files:
  - vertexos.rules
outputs:
  - eve-log:
      enabled: yes
      filename: /var/log/suricata/eve.json
      types:
        - alert
```

```bash
suricata -T -c /etc/suricata/suricata.yaml   # validate config + rules
sv restart suricata
```

Set `HOME_NET` to your LAN in `suricata.yaml`. The commented C2-port rule
(`sid:9100009`) is intentionally off — enable it only where those ports are
never used legitimately.
