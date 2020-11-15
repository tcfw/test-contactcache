# test-contactcache

# Requirements

- Go (>=1.15)
- Make (for building)

## Build

`make build` or simply `make` will compile the program and place the binary in ./build (or BUILD_DIR env var)

## Running testing

`make test` will run all tests

## Running

To start the server and serve HTTPS requests (e.g.):
`./build/contactcache start -l 127.0.0.1:8443 --tls-key ./key.pem --tls-cert cert.pem`

See `contactcache start --help` for more options

## Config

A YAML file can be placed in `/etc/contactcache/` `$HOME/.contactcache` or the current directory.

You can also pass configuration via env vars prefixed with CONTACTCACHE\_ (e.g. to set 'tls.key' you may set `CONTACTCACHE_TLS_KEY`)

- `tls`: sets TLS config (see below)
- `tls.key`: TLS private key
- `tls.cert`: TLS certificate
- `address`: Address to listen on
- `backend.address`: The backend server
- `cache.address` The caching endpoint
- `cache.password`: Redis password

## TLS Cert generation (self-signed)

`DO NOT USE FOR PRODUCTION` - Correctly signed certificates should be used for production

Generating temporarily TLS certificates:
`openssl req -new -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -x509 -nodes -days 365 -out cert.pem -keyout key.pem`
