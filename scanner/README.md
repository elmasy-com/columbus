# columbus-scanner

This program is used to parse the certificates from a CT log and insert into the Columbus database. 

## Build

- `go 1.19` required!

```bash
make build
```

## Install

> The `columbus-scanner.sha` is signed with [this key](https://keys.openpgp.org/vks/v1/by-fingerprint/10BC80B36072944B5678AF395D00FD9E9F2A3725).

Download the key:
```bash
gpg --recv-key 10BC80B36072944B5678AF395D00FD9E9F2A3725
```

Download and verify:
```bash 
wget -q 'https://github.com/elmasy-com/columbus-scanner/releases/latest/download/columbus-scanner' -O columbus-scanner && \
wget -q 'https://github.com/elmasy-com/columbus-scanner/releases/latest/download/columbus-scanner.sha' -O columbus-scanner.sha && \
wget -q 'https://raw.githubusercontent.com/elmasy-com/columbus-scanner/main/scanner.conf.example' -O scanner.conf && \
gpg --verify columbus-scanner.sha && sha512sum -c columbus-scanner.sha && rm columbus-scanner.sha && chmod +x columbus-scanner
```

1. Place the binary somewhere
2. Update and place the config file somewhere
3. Update and install `columbus-scanner.service` somewhere
