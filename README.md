````markdown
# Distributed Replicated Key-Value Store

A simple distributed key-value store implemented in **Go** to demonstrate the concepts of **Replication**, **Eventual Consistency**, **Strong Consistency**, **Versioning**, and **Conflict Resolution**.

This project was developed for the **Distributed Computing Fundamentals** course at the **University of Tehran**.

---

# Features

- Three independent replica servers
- HTTP-based communication between replicas
- Local storage on each replica
- Support for GET and PUT operations
- Eventual Consistency
- Simplified Strong Consistency using Majority Quorum
- Versioning for each key
- Conflict Resolution using Last-Write-Wins (LWW)
- Strong GET through Primary Replica
- Artificial network latency simulation
- Automatic execution of all required scenarios
- Performance measurements and experiment outputs

---

# Project Structure

```
distributed-kv-store/
│
├── client/
│   └── main.go
│
├── replica/
│   └── main.go
│
├── configs/
│   ├── replica1.json
│   ├── replica2.json
│   └── replica3.json
│
├── results/
│   ├── scenario1.txt
│   ├── scenario2.txt
│   ├── scenario3.txt
│   └── scenario4.txt
│
├── start.sh
├── go.mod
├── README.md
└── report.pdf
```

---

# System Architecture

The system consists of three independent replica servers.

```
                Client
                   |
        ---------------------
        |         |         |
     Replica1  Replica2  Replica3
        | \       | \      | \
        |  \______|__\_____|__\
          Replication Messages
```

Each replica:

- runs as an independent process
- stores data locally
- communicates with other replicas using HTTP
- performs replication over the network

---

# Technologies

- Go
- HTTP
- JSON
- Goroutines
- Mutex (sync.RWMutex)

---

# Requirements

- Go 1.22 or newer
- Linux/macOS (or Windows using Git Bash)

---

# Running the Project

## Step 1

Clone the repository.

```
git clone <repository-url>
cd distributed-kv-store
```

---

## Step 2

Make the launcher executable.

```
chmod +x start.sh
```

---

## Step 3

Run the project.

```
./start.sh
```

The launcher automatically:

- starts Replica 1
- starts Replica 2
- starts Replica 3
- waits for replicas to initialize
- starts the client
- executes the experiments

---

# Replica Configuration

Each replica has its own configuration file.

Example:

```
configs/replica1.json
```

```json
{
    "id": "replica1",
    "port": ":8081",
    "is_primary": true,
    "peers":
    {
        "replica2":"http://localhost:8082",
        "replica3":"http://localhost:8083"
    }
}
```

Configuration fields:

| Field | Description |
|--------|-------------|
| id | Replica identifier |
| port | Listening port |
| is_primary | Whether this node is the primary replica |
| peers | Other replicas in the cluster |

---

# Supported Operations

## PUT

```
PUT /put
```

Example JSON:

```json
{
    "key":"x",
    "value":"10",
    "consistency_mode":"eventual"
}
```

or

```json
{
    "key":"x",
    "value":"10",
    "consistency_mode":"strong"
}
```

---

## GET

```
GET /get?key=x
```

Returns

```json
{
    "key":"x",
    "value":"10",
    "version":2,
    "updated_by":"replica1"
}
```

---

## Strong GET

```
GET /strong_get?key=x
```

If the request is sent to a secondary replica, it is forwarded to the primary replica to guarantee a consistent read.

---

# Consistency Models

## Eventual Consistency

1. Write locally.
2. Return success immediately.
3. Replicate asynchronously.
4. Replicas eventually converge.

Advantages:

- Fast writes
- High availability

Disadvantages:

- Temporary stale reads

---

## Strong Consistency

1. Write locally.
2. Send replication requests.
3. Wait for majority acknowledgements.
4. Return success only if quorum is reached.

Advantages:

- No stale reads (using Strong GET)
- Immediate consistency

Disadvantages:

- Higher latency
- Writes may fail if quorum is unavailable

---

# Versioning

Each stored record contains

```json
{
    "key":"x",
    "value":"10",
    "version":3,
    "updated_by":"replica1",
    "timestamp":123456789
}
```

Each PUT operation increments the version number.

---

# Conflict Resolution

The project uses **Last-Write-Wins (LWW)**.

If two replicas receive different values with the same version:

1. Compare timestamps.
2. Keep the newest timestamp.
3. If timestamps are equal, choose the replica with the larger replica ID.

---

# Artificial Network Latency

The system supports configurable replication delays.

Possible values:

- 0 ms
- 500 ms
- 2000 ms

Latency is configured through

```
/set_latency
```

and is used for Scenario 4.

---

# Implemented Scenarios

## Scenario 1

Temporary inconsistency.

- Write to Replica 1
- Immediately read from Replica 2
- Observe stale read
- Wait
- Observe convergence

---

## Scenario 2

Replica failure.

- Stop one replica
- Execute PUT
- Compare Eventual and Strong consistency

---

## Scenario 3

Concurrent conflict.

- Two replicas write different values simultaneously
- Replication occurs
- Conflict resolved using LWW
- Verify convergence

---

## Scenario 4

Network latency.

Experiments performed with

- 0 ms
- 500 ms
- 2000 ms

Metrics collected:

- PUT latency
- Convergence time
- Number of stale reads
- Number of updated replicas

---

# Output Files

After execution, the client automatically generates

```
results/
```

containing

```
scenario1.txt
scenario2.txt
scenario3.txt
scenario4.txt
```

Each file contains the console output corresponding to its experiment.

---

# Metrics Collected

The project measures:

- PUT response time
- GET response time
- Replica convergence time
- Number of updated replicas
- Number of stale reads

---