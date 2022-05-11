# Skillbox lesson 31

1. run mongodb at localhost:27017

2. test server:

```bash
go test ./test/... -v
```

3. run servers in the split terminal:

```bash
go run ./cmd/server/server.go -p 8000
go run ./cmd/server/server.go -p 9000
```

4. run proxy:

```bash
go run ./cmd/proxy/proxy.go
```
