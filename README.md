# eventloop

Checkout [`CONTRIBUTING.md`](docs/CONTRIBUTING.md).
See the frontend repo for eventloop: [`eventloop-frontend`](https://github.com/homebrew-ec-foss/eventloop-frontend)
# How to run locally?
Run locally or inside a docker container. If running locally enable `ELOOP_DEV=1` to expose some additional routes for testing.
### Running as a container

- Have docker installed [https://docs.docker.com/get-started/get-docker/](https://docs.docker.com/get-started/get-docker/)
- the container works with a shared volume at `${PWD}/data`, ensure that the `config.json` is present there.
- the `events.db` file will be present in the same shared volume, so clean the db with `make clean_db` target to work with a fresh database before running the container.

```bash
make build-container-image

make run-container
# this runs the container in DEV mode to expose few
# routes for testing
```
### Running locally 
```
ELOOP_LOCAL=1 go run -v .
```

## HTTPS
eventloop back-end uses an HTTP connection, you can either generate your own localhost certificates or use a proxy server like [mitmproxy](https://mitmproxy.org/)

### Using a proxy server (Recommended)
When running locally:
```bash
mitmproxy --mode reverse:http://localhost:8000@8080 --set ssl_insecure=true
```
When running inside docker container:
```bash
mitmproxy --mode reverse:http://localhost:8080@<port-no> --set ssl_insecure=true
```
You will have to update the frontend with the new `port-no`. Alternatively you can also,
### Regenerating localhost certificates
```sh
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

# pfx file for browsers
openssl pkcs12 -export -out localhost.pfx -inkey localhost.key -in localhost.crt
```

Change the `r.Run()` function in [`main.go`](main.go) to
```go
  if err := r.RunTLS(":8080", "localhost.crt", "localhost.key"); err != nil {
	log.Fatal(err)
  }
```
# QR's
QR codes for all participants are automatically generated when an event is created. QR's are stored in `../test-data/qr-png/` directory. To regenerate them just hit the `admin/create` endpoint with participant data again.
# Mail
eventloop has a mailer cli tool which will automatically send emails to all participating teams with their generated QR codes. Checkout [`README.md`](mail/README.md)

# Deployment

- Run `docker compose up` in the directory containing `compose.yml`, with the
application configuration(`config.json`) and environment variables(`.env`) in `./data`
- Move nginx.conf to `/etc/nginx/nginx.conf`
- Point domain to server and run certbot to generate SSL certificates

