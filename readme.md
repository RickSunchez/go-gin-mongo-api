# Skillbox lesson 31

1. run mongodb at localhost:27017

1. test server:

```bash
go test ./test/... -v
```

1. run servers in the split terminal:

```bash
go run ./cmd/server/server.go -p 8000
go run ./cmd/server/server.go -p 9000
```

1. run proxy:

```bash
go run ./cmd/proxy/proxy.go
```
