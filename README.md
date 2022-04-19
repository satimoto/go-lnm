# go-lsp
Satimoto Lightning Service Provider using golang

## Development

### Run
```bash
go run ./cmd/lsp
```

## Build

### Run
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '-s -w' -o bin/main cmd/lsp/main.go
```