# Grafana Alloy remote configuration Server (Go-ARCS)

Inspired by [opsplane-services implementation](https://github.com/opsplane-services/alloy-remote-config-server)

Build upon the [official Grafana Alloy Api remotecfg API spec](https://github.com/grafana/alloy-remote-config/tree/main)

## building

Build Client and Server with 'make' (see [makefile](./makefile) for details)

```sh
make build # builds client and server
make build-sever # or build-client for specific
make run # build and runs the server
make publish # builds and publishes client and server
```

## running

### client

Can be run as a docker container

```sh
docker exec go-arcs-client [scope] [action] [attributes]
```

## server

Can be run directly or included in compose.yaml (see example)

```sh
docker run -p 8080:8080 -v [configs]:/tmp go-arcs-server
```