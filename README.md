# shareDir

`shareDir` is a P2P-based directory sharing tool that includes a tracker for peer discovery and a peer client for file sharing.

## Features

- **Daemonized Architecture**: Both Tracker and Peer can run as daemons with background support.
- **Behavior Abstraction**: A unified `daemon` package manages process lifecycles (PID management, signals, logging) and abstracts server/client behaviors.
- **IP Tracking**: Integrated `iptracker` component for maintaining peer status and IP mapping.
- **Extensible Protocol**: Pluggable protocol system for file transport.

## Installation

Ensure you have [Go](https://golang.org/doc/install) installed.

```bash
go build -o shareDir main.go
```

## Usage

### Tracker (Server)

The tracker maintains a list of active peers and their IP addresses.

**Start the tracker:**
```bash
./shareDir tracker start --port 8080
```

**Start in background:**
```bash
./shareDir tracker start --port 8080 -d
```

**Stop the tracker:**
```bash
./shareDir tracker stop
```

### Peer (Client)

The peer registers itself with the tracker and communicates with other peers.

**Start a peer:**
```bash
./shareDir peer start --id "peer-1" --addr "127.0.0.1:8080"
```

**Start in background:**
```bash
./shareDir peer start --id "peer-1" --addr "127.0.0.1:8080" -d
```

**Stop the peer:**
```bash
./shareDir peer stop
```

## Project Structure

- `cmd/`: Command-line interface using Cobra.
- `daemon/`: Core daemon coordinator and lifecycle management.
- `iptracker/`: Peer discovery and IP tracking logic.
- `procotal/`: (Protocol) File transport protocols and meta-data handling.
- `main.go`: Application entry point.

## Logs and PIDs

When running in daemon mode (`-d`), logs and PID files are stored in:
- PID: `/tmp/share_{name}.pid`
- Logs: `/tmp/share_{name}.log`
