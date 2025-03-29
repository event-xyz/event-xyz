# Eventloop

Checkout [`CONTRIBUTING.md`](docs/CONTRIBUTING.md).

Access the frontend repo for eventloop: [`eventloop-frontend`](https://github.com/homebrew-ec-foss/eventloop-frontend)

# How to run eventloop locally?

Run locally or within a docker container.
If running locally, set `ELOOP_DEV=1` environment variable to expose some additional routes for testing.

## Running within a container

- Ensure docker is installed: [https://docs.docker.com/get-started/get-docker/](https://docs.docker.com/get-started/get-docker/)
- The container operates on a shared volume at `${PWD}/data`, ensure that `config.json` and `.env` is present in the volume.
- The `events.db` SQLite database will be present in the same shared volume, so run the `make clean_db` target to work with a fresh database instance before running the container.

```sh
make build-container-image

# this runs the container in DEV mode to expose additional
# routes for testing
make run-container
```

## Running locally 

### Without HTTPS

```sh
# eventloop listens on port 8080
go run -v .
```

### With HTTPS

The eventloop back-end uses a HTTPS connection.

You can either generate your own localhost certificates or use a proxy server like [mitmproxy](https://mitmproxy.org/).

### 1. Using a proxy server (Recommended)

When running locally:

```sh
# eventloop listens on port 8000
# the proxy server proxies requests from port 8080 to 8000
ELOOP_LOCAL=1 go run -v .
```

- Run the proxy server

```sh
mitmproxy --mode reverse:http://localhost:8000@8080 --set ssl_insecure=true
```

When running within a docker container:
```sh
mitmproxy --mode reverse:http://localhost:8080@<port-number> --set ssl_insecure=true
```

You will have to update the frontend with the new `port-number`. 

### 2. Regenerating localhost certificates (Alternate approach)

- Generate certificates

```sh
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

# pfx file for browsers
openssl pkcs12 -export -out localhost.pfx -inkey localhost.key -in localhost.crt
```

- Change the `r.Run()` function in [`main.go`](main.go) to

```go
  if err := r.RunTLS(":8080", "localhost.crt", "localhost.key"); err != nil {
	log.Fatal(err)
  }
```

## QR Codes
QR codes for all participants are automatically generated when an event is created. QRs are stored in the `../test-data/qr-png/` directory. To regenerate them, just hit the `admin/create` endpoint with participant data again.

# Mail

eventloop has a mailer cli tool which will automatically send emails to all participating teams with their generated QR codes. Checkout [`README.md`](mail/README.md)

# Deployment

- Run `docker compose up` in the directory containing `compose.yml`, with the application configuration(`config.json`) and environment variables(`.env`) in the `./data` directory.
- Move nginx.conf to `/etc/nginx/nginx.conf`
- Point domain to server and run certbot to generate SSL certificates
