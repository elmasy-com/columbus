# Columbus server

Columbus Project is an API first subdomain discovery service, blazingly fast subdomain enumeration service with advanced features. 

![Subdomain Lookup](https://columbus.elmasy.com/gif/lookup.gif)
*Columbus returned 638 subdomains of tesla.com in 0.231 sec.*

## Usage

By default Columbus returns only the subdomains in a JSON string array:
```bash
curl 'https://columbus.elmasy.com/lookup/github.com'
```

But we think of the bash lovers, so if you don't want to mess with JSON and a newline separated list is your wish, then include the `Accept: text/plain` header.
```bash
DOMAIN="github.com"

curl -s -H "Accept: text/plain" "https://columbus.elmasy.com/lookup/$DOMAIN" | \
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

**For more, check the [features](https://columbus.elmasy.com/tools) or the [API documentation](https://columbus.elmasy.com/swagger/index.html).**

## Entries

Currently, entries are got from [Certificate Transparency](https://certificate.transparency.dev/).

## Command Line

```
Usage of columbus-server:
  -check
    	Check for updates.
  -config string
    	Path to the config file.
  -version
    	Print version informations.
```

`-check`: Check the lates version on GitHub.
Prints `up-to-date` and returns `0` if no update required.
Prints the latest tag (eg.: `v0.9.1`) and returns `1` if new release available.
In case of error, prints the error message and returns `2`.

## Build

```bash
git clone https://github.com/elmasy-com/columbus-server
make build
```

## Install

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

Copy the config file to `/etc/columbus/server.conf`.

Set the permission to 0600.
```bash
chmod -R 0600 /etc/columbus
```

Set the owner of the config file:
```bash
chown -R columbus-server:columbus /etc/columbus
```

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
 
