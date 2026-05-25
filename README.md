# lscompute

## Install

```bash
sudo snap install lscompute
sudo snap connect lscompute:hardware-observe # TODO: auto connect
```

To build and install from source, refer to [here](#build-snap).

## Usage

Currently the only output format is JSON:
```bash
lscompute 
```

## Development

### Run tests

```bash
go test -count 1 -failfast ./...
```

### Compile and run

```bash
go run ./cmd/lscompute
```

### Build binaries

```bash
go build ./cmd/lscompute
```

For cross-compilation, set the `GOOS` and `GOARCH` environment variables. For example, to build for ARM64 architecture on Linux:
```bash
GOOS=linux GOARCH=arm64 go build -o lscompute-arm64 ./cmd/lscompute
```

### Build snap

```bash
snapcraft -v
```

Then install the snap and connect the required interfaces:
```bash
sudo snap install --dangerous *.snap
sudo snap connect lscompute:hardware-observe 
```

