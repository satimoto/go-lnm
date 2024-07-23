# go-lnm
Satimoto Lightning Node Manager using golang

## Development

### Run
```bash
go run ./cmd/lnm
```

## Build

### Run
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/main cmd/lnm/main.go
```