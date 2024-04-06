# Columbus

Columbus Project is an API first subdomain discovery service, blazingly fast subdomain enumeration service with advanced features. 

![Subdomain Lookup](https://columbus.elmasy.com/count_teslacom.gif)
*Columbus returned 763 subdomains of tesla.com in less than a second.*

## Usage

By default Columbus returns only the subdomains in a JSON string array:
```bash
curl 'https://columbus.elmasy.com/api/lookup/github.com'
```

But we think of the bash lovers, so if you don't want to mess with JSON and a newline separated list is your wish, then include the `Accept: text/plain` header.
```bash
DOMAIN="github.com"

curl -s -H "Accept: text/plain" "https://columbus.elmasy.com/api/lookup/$DOMAIN" | \
while read SUB
do
        if [[ "$SUB" == "" ]]
        then
                HOST="$DOMAIN"
        else
                HOST="${SUB}.${DOMAIN}"
        fi
        echo "$HOST"
done
```

**For more, check the [website](https://columbus.elmasy.com/) and the [API documentation](https://columbus.elmasy.com/swagger/).**

## Data source

- [Certificate Transparency](https://certificate.transparency.dev/)
- [DNS server](#dns)

See more statistics on the [website](https://columbus.elmasy.com/stat).

## Server

### Usage

```
Usage of columbus-server:
  -check
    	Check for updates.
  -config string
    	Path to the config file.
  -version
    	Print version informations.
```

### Build

```bash
git clone https://github.com/elmasy-com/columbus
make server-build
```

### Install

Create a new user:

```bash
adduser --system --no-create-home --disabled-login columbus-server
```

Create a new group:

```bash
addgroup --system columbus
```

Add the new user to the new group:

```bash
usermod -aG columbus columbus-server
```

Copy the binary to `/usr/bin/columbus-server`.

Make it executable:
```bash
chmod +x /usr/bin/columbus-server
```

Create a directory:
```bash
mkdir /etc/columbus
```

Set the permission to 0600.
```bash
chmod -R 0640 /etc/columbus
```

Set the owner of the config file:
```bash
chown -R columbus-server:columbus /etc/columbus
```

Copy the config file to `/etc/columbus/server.conf` and configure.


Install the service file (eg.: `/etc/systemd/system/columbus-server.service`).
```bash
cp columbus-server.service /etc/systemd/system/
```

Reload systemd:
```bash
systemctl daemon-reload
```

Start columbus:
```
systemctl start columbus-server
```

If you want to columbus start automatically:
```
systemctl enable columbus-server
```
 
## Scanner

This program is used to parse the certificates from a CT log and insert into the Columbus database. 

### Build

```bash
make scanner-build
```

### Install

> The `columbus-scanner.sha` is signed with [this key](https://keys.openpgp.org/vks/v1/by-fingerprint/10BC80B36072944B5678AF395D00FD9E9F2A3725).

Download the key:
```bash
gpg --recv-key 10BC80B36072944B5678AF395D00FD9E9F2A3725
```

1. Place the binary somewhere
2. Update and place the config file somewhere
3. Update and install `columbus-scanner.service` somewhere


## DNS

A DNS server to collect domains to Columbus Server.

The goal of `dns` is to make it easy/possible for users to easily contribute to the Columbus Project.

By setting the DNS servers to `dns` while enumerationg subdomains, hunting bugs, etc you can contribute to the Columbus Project.

### Design

```
                               _________________
  |---> dig exmaple.com A ---->| COLUMBUS-DNS  | -----> Insert -----> Columbus DB
  |                            -----------------
  ^                                   |
  |                                   V
Alice <------- 93.184.216.34 <---------

```

- Only domains with valid answer will be sent to the server (`NOERROR`)

#### IMPORTANT!

**THIS SERVER IS NOT MEANT TO USED AS A DAILY DNS RESOLVER!** 

### Install

The checksum file is signed with key `10BC80B36072944B5678AF395D00FD9E9F2A3725`.

```bash
gpg --receive-key 10BC80B36072944B5678AF395D00FD9E9F2A3725
```

### Install

Create a new user:

```bash
adduser --system --no-create-home --disabled-login columbus-dns
```

Create a new group (if not exists):

```bash
addgroup --system columbus
```

Add the new user to the new group:

```bash
usermod -aG columbus columbus-dns
```

Copy the binary to `/usr/bin/columbus-dns`.

Make it executable:
```bash
chmod +x /usr/bin/columbus-dns
```

Create a directory (if not exists):
```bash
mkdir /etc/columbus
```

Set the permission to 0600.
```bash
chmod -R 0640 /etc/columbus
```

Set the owner of the config file:
```bash
chown -R columbus-dns:columbus /etc/columbus
```

Copy the config file to `/etc/columbus/dns.conf` and configure.


Install the service file (eg.: `/etc/systemd/system/columbus-dns.service`).
```bash
cp columbus-dns.service /etc/systemd/system/
```

Reload systemd:
```bash
systemctl daemon-reload
```

Start Columbus DNS:
```
systemctl start columbus-dns
```

If you want to Columbus DNS start automatically:
```
systemctl enable columbus-dns
```

## Frontend

HTML + [tailwindcss](https://tailwindcss.com/) + [DaisyUI](https://daisyui.com/). The style is inspired by [Introduction theme](https://github.com/victoriadrake/hugo-theme-introduction).

## VHS

Create gifs with [VHS](https://github.com/charmbracelet/vhs).


## Author

[System administrator service and Cybersecurity for small and medium-sized businesses in and around Győr.](https://www.gorbe.io/)
