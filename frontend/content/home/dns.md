---
title: "DNS Server"
weight: 4
---


[Columbus DNS](https://github.com/elmasy-com/columbus-dns) is a high performance DNS server to collect and update domains for Columbus Project.

The goal of this DNS server is to make it easy for everybody to contribute to the Columbus Project by setting the DNS server to `dns.columbus.elmasy.com` while enumerationg subdomains, hunting bugs, etc.

***IMPORTANT***: This server is not meant to used as a daily DNS resolver!

Server:
- DNS: `dns.columbus.elmasy.com`
- IPv4: `142.132.164.231`
- IPv6: `2a01:4f8:1c1e:eddd::1`
- Protocol: `TCP` and `UDP`


Example:

```bash
dig @142.132.164.231 elmasy.com
```

```bash
subfinder -r 142.132.164.231 -d elmasy.com
```

```bash
amass enum -tr 142.132.164.231 -d elmasy.com
```