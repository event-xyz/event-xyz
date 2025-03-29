# Backend code for event-loop

Visit [`CONTRIBUTING.md`](docs/CONTRIBUTING.md) for setup-help.

## Running in developement mode
To run in development mode, set `ELOOP_DEV=1` in `data/.env`

### Regenerating localhost certificates

> Necessary for https for go backend servers

```sh
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")

# pfx file for browsers
openssl pkcs12 -export -out localhost.pfx -inkey localhost.key -in localhost.crt
```

## Running back-end as a container

- have docker install [https://docs.docker.com/get-started/get-docker/](https://docs.docker.com/get-started/get-docker/)
- the container works with a shared volume at `${PWD}/data`, so ensure that the `config.json` file is present there
- the events.db file will be present in the same shared volume, so clean the db with `make clean_db` target to work with a fresh database before running the container.

```bash
make build-container-image

make run-container
# this runs the container in DEV mode to expose few
# routes for testing
```

While working with the front-end and you are making use of a docker container, the container uses an http connection, for now u could a proxy server [mitmproxy](https://mitmproxy.org/), also this uses `https://localhost:8080` by default, so you can change the port exposed on the host side to be something else

```bash
mitmproxy --mode reverse:http://localhost:<whichever port you are using> --set ssl_insecure=true
```

## Regenerating Test QR's

The test qr's work only with the `labels.csv`, generated with a specific env key at `event-loop-backend/.env`. Test QR's stored in `/test-data` of repository

`handlers/authentication.go`

```go
func GenerateQR(signedString string, i int) ([]byte, error) {
	// -- snip --
	err = qrcode.WriteFile(signedString, qrcode.Medium, 256, fmt.Sprintf("part-%d.png", i))
	if err != nil {
		return nil, err
	}
	// -- snip --
}
```


## Deployment

- Run `docker compose up` in the directory containing `compose.yml`, with the
application configuration(`config.json`) and environment variables(`.env`) in `./data`
- Move nginx.conf to `/etc/nginx/nginx.conf`
- Point domain to server and run certbot to generate SSL certificates
