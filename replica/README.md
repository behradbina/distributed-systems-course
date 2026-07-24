# Replica Server

## Description

Each replica is an independent HTTP server that stores key-value pairs locally and communicates with other replicas over the network. The replica supports both Eventual Consistency and Simplified Strong Consistency.

## Features

- Independent replica process
- Local in-memory key-value storage
- HTTP API for GET and PUT operations
- Replica-to-replica data replication
- Eventual Consistency
- Majority-based Strong Consistency
- Strong GET through the primary replica
- Version management for each key
- Conflict resolution using Last-Write-Wins (LWW)
- Configurable artificial network latency

## Configuration

Each replica is started using its own configuration file located in the `configs/` directory.

Example:

```bash
go run replica/main.go -config=configs/replica1.json
```

Available configuration files:

```
configs/
├── replica1.json
├── replica2.json
└── replica3.json
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/put` | Store or update a key-value pair |
| `/get` | Retrieve a value |
| `/replicate` | Receive replication messages from peers |
| `/strong_get` | Read from the primary replica |
| `/set_latency` | Configure artificial replication delay |

## Running

Start each replica in a separate terminal:

```bash
go run replica/main.go -config=configs/replica1.json
```

```bash
go run replica/main.go -config=configs/replica2.json
```

```bash
go run replica/main.go -config=configs/replica3.json
```

Alternatively, execute the provided launcher script:

```bash
./start.sh
```

which automatically starts all replicas before running the client.