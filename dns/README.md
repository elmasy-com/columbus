# columbus-dns

A DNS server to collect domains to Columbus Server.

The goal of `columbus-dns` is to make it easy/possible for users to easily contribute to the Columbus Project.

By setting the DNS servers to `columbus-dns` while enumerationg subdomains, hunting bugs, etc you can contribute to the Columbus Project.

## Design

```
                               _________________
  |---> dig exmaple.com A ---->| COLUMBUS-DNS  | -----> Insert -----> columbus.elmasy.com
  |                            -----------------
  ^                                   |
  |                                   V
Alice <------- 93.184.216.34 <---------

```

- Only domains with valid answer will be sent to the server (`NOERROR`)

# IMPORTANT!

**THIS SERVER IS NOT MEANT TO USED AS A DAILY DNS RESOLVER!** 


# Install

The most secure way to run `columbus-dns` is to run as a **non root** user on a non-privileged port (eg.: 1053) and *redirect* traffic deisgnated to port 53.

nftables example:
```
table inet filter {
	chain inbound_ipv4 {
    ...
		tcp dport 1053 accept
		udp dport 1053 accept
	}

	chain inbound_ipv6 {
    ...
	}

	chain input {
		type filter hook input priority filter; policy drop;
		ct state { established, related } accept
		iifname "lo" accept
		meta protocol vmap { ip : jump inbound_ipv4, ip6 : jump inbound_ipv6 }
		tcp dport 22 accept
		reject
	}

  ...
}

table ip nat {
	chain prerouting {
		type nat hook prerouting priority filter; policy accept;
		udp dport 53 redirect to :1053
		tcp dport 53 redirect to :1053
	}

	chain postrouting {
		type nat hook postrouting priority filter; policy accept;
	}
}
```

The checksum file is signed with key `10BC80B36072944B5678AF395D00FD9E9F2A3725`.

```bash
gpg --receive-key 10BC80B36072944B5678AF395D00FD9E9F2A3725
```

1. Create the user:
```bash
adduser --no-create-home --disabled-login "columbus-dns"
```

2. Install (and edit if needed) `columbus-dns.service`:
```bash
wget -q -O "columbus-dns.service" "https://raw.githubusercontent.com/elmasy-com/columbus-dns/main/columbus-dns.service"
```
```bash
mv columbus-dns.service /etc/systemd/system/
```

3. Install and edit `columbus-dns.conf`
```bash
mkdir /etc/columbus
```
```bash
wget -q -O "columbus-dns.conf" "https://raw.githubusercontent.com/elmasy-com/columbus-dns/main/columbus-dns.conf"
```
```bash
sudo mv columbus-dns.conf /etc/columbus/dns.conf
```
```bash
chown columbus-dns:columbus-dns /etc/columbus/dns.conf
```
```bash
sudo chmod 0640 /etc/columbus/dns.conf
```
```bash
nano /etc/columbus/dns.conf
```

4. Install the binary:

For Linux/AMD64:
```bash
wget -q 'https://github.com/elmasy-com/columbus-dns/releases/latest/download/columbus-dns-linux-amd64' -O columbus-dns-linux-amd64 && \
wget -q 'https://github.com/elmasy-com/columbus-dns/releases/latest/download/checksums' -O checksums && \
gpg --verify checksums && sha512sum --ignore-missing -c checksums && rm checksums && \
sudo install columbus-dns-linux-amd64 /usr/bin/columbus-dns
```

5. Start the service
```bash
sudo systemctl daemon-reload
```
```bash
sudo systemctl enable --now columbus-dns.service
```