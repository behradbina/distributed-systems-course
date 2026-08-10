# Part 2: Concurrency, Scheduling, and Context Switching Effects

This section benchmarks Go's goroutine-based concurrency across two workload types (CPU-bound and IO-bound), varying the number of goroutines and available OS threads (`GOMAXPROCS`), and visualizes the results.

---

## Overview

The benchmark runs **640 total tasks** across different goroutine counts (`1, 2, 4, 8, 16, 32, 64`) and `GOMAXPROCS` values (`1, 2, NumCPU`), for both workload types. Results are saved to CSV and plotted as PNG charts.

### Workloads

| Workload   | Description                                                                 |
|------------|-----------------------------------------------------------------------------|
| CPU-bound  | Heavy math loop: `sqrt` + `log` computed 100,000 times per task             |
| IO-bound   | Short math loop followed by a **3 ms `time.Sleep`** to simulate I/O waiting |

### Metrics Collected

- **Total time** (ms) — wall-clock time to complete all tasks
- **Throughput** (tasks/ms) — tasks completed per millisecond
- **Average / Min / Max / Std deviation latency** (ms) — per-task timing statistics

---

## Project Structure

```
.
├── main.go          # Benchmark entry point
├── results/
│   ├── results.csv  # Raw benchmark data
│   ├── CPU-bound_throughput.png
│   ├── CPU-bound_total_time.png
│   ├── IO-bound_throughput.png
│   └── ...          # One chart per metric × workload
└── trace.out        # Go runtime execution trace
```

---

## Dependencies

Install the external plotting library:

```bash
go get gonum.org/v1/plot/...
```

Only standard library packages are used otherwise:
`os`, `fmt`, `math`, `sort`, `sync`, `time`, `strconv`, `runtime`, `runtime/trace`, `encoding/csv`, `path/filepath`

---

## Running the Benchmark

```bash
go run main.go
```

This will:
1. Run all benchmark combinations (3 GOMAXPROCS × 2 workloads × 7 goroutine counts = 42 runs)
2. Save results to `results/results.csv`
3. Generate PNG plots under `results/`
4. Write a runtime trace to `trace.out`

> **Note:** The `results/` directory must exist before running:
> ```bash
> mkdir -p results
> ```

---

## Generating & Viewing the Runtime Trace

The `main` function activates Go's built-in execution tracer, which writes to `trace.out`.

### View the trace

```bash
go tool trace trace.out
```

Expected output:
```
2026/05/12 14:30:39 Preparing trace for viewer...
2026/05/12 14:30:39 Splitting trace for viewer...
2026/05/12 14:30:39 Opening browser. Trace viewer is listening on http://127.0.0.1:<PORT>
```

Open the printed URL in your browser to explore goroutine states, OS thread activity, scheduler events, and GC pauses interactively.

---

### Plotting

Charts are generated with [`gonum/plot`](https://pkg.go.dev/gonum.org/v1/plot). For each metric, one chart is produced per workload type, with separate lines for each `GOMAXPROCS` value.