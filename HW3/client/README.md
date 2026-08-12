# Client

## Description

The client is responsible for interacting with the distributed key-value store and executing the required experiments. It communicates with the replica servers through HTTP requests and automatically runs all scenarios required by the assignment.

## Responsibilities

- Send PUT requests
- Send GET requests
- Perform Strong GET operations
- Configure artificial network latency
- Execute all experimental scenarios
- Measure execution metrics such as:
  - PUT latency
  - Convergence time
  - Number of stale reads
  - Number of updated replicas
- Save the output of each scenario in the `results/` directory.

## Running

Ensure that all replica servers are already running, then execute:

```bash
go run client/main.go
```

The client will automatically execute all scenarios and generate:

```
results/
├── scenario1.txt
├── scenario2.txt
├── scenario3.txt
└── scenario4.txt
```