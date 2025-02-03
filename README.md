# Backend code for event-loop

Visit CONTRIBUTING.md for setup-help.

## Regenerating localhost certificates

> Necessary for https for go backend servers

```sh
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
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
