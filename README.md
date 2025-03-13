A TCP server-client app where clients send signed prime numbers, and the server collects unique prime numbers.

- Go 1.23.6

## How to execute server

```bash
go run cmd/server/main.go -max=200
```

## How to execute clients (from multiple terminals)

```bash
go run cmd/client/main.go  # Terminal 1
go run cmd/client/main.go  # Terminal 2
go run cmd/client/main.go  # Terminal 3
...
```

### How to test

#### Test whole system

```bash
go test -v ./...
```

#### Test individual tests

```bash
go test -v ./cmd

go test -v ./pkg/primes
go test -v ./pkg/auth
go test -v ./pkg/pool
```

### Compile binaries and execute (optional)

```bash
go build -o server cmd/server/main.go
go build -o client cmd/client/main.go

./server -max=20000

./client  # Terminal 1
./client  # Terminal 2
./client  # Terminal 3
...
```
