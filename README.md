# download-public-ssl-certificate

this generator allows you to easily embedd SSL certificates for pinning into
your go binaries

## Usage

```go
//go:generate download-public-ssl-certificate -pkg ga -o certs.go www.google-analytics.com
```

or from bash:

```bash
download-public-ssl-certificate -pkg ga -o certs.go www.google-analytics.com
```
