# shareDir

`shareDir` is a P2P-based directory sharing tool.

## Installation

Ensure you have [Go](https://golang.org/doc/install) installed.

```bash
go build -o shareDir main.go
```

## Usage

### Tracker (Server)

```bash
./shareDir tracker start --port 8080
./shareDir tracker stop
```

### Peer (Client)

```bash
./shareDir peer start --id "peer-1" --addr "127.0.0.1:8080"
./shareDir peer stop
```
