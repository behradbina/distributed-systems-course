# Part 1 — IPC via Named Pipes

Two independent Go processes that communicate through Linux named FIFOs (named pipes) to do calculations of user inputs.

---

## Architecture

```
┌─────────────────┐     /pipes/request.fifo    ┌──────────────────┐
│  interface.go   │ ─────────────────────────► │   worker.go      │
│  (user-facing)  │                            │  (calc engine)   │
│                 │ ◄───────────────────────── │                  │
└─────────────────┘    /pipes/response.fifo    └──────────────────┘
```

- **Worker** creates both FIFOs on startup and listens continuously.
- **Interface** connects to the existing FIFOs and provides an interactive REPL.
- Each request/response pair is a single `\n`-terminated line over the respective FIFO.
- Responses are JSON; requests are plain text (`OP A B`).

---

## How to Run

### Step 1 — Start the Worker (first terminal)

```bash
go run worker.go
```

The worker will block until the Interface connects.

### Step 2 — Start the Interface (second terminal)

```bash
go run interface.go
```

> **Order matters**: start Worker before Interface, because the Interface checks for the FIFO files at startup.

---

## Supported Operations

| Operation | Syntax      | Example       |
|-----------|-------------|---------------|
| ADD       | `ADD A B`   | `ADD 5 7`     |
| SUB       | `SUB A B`   | `SUB 10 3`    |
| MUL       | `MUL A B`   | `MUL 4 6`     |
| DIV       | `DIV A B`   | `DIV 8 2`     |
| MOD       | `MOD A B`   | `MOD 10 3`    |
| POW       | `POW A B`   | `POW 2 10`    |
| MAX       | `MAX A B`   | `MAX 3 9`     |
| MIN       | `MIN A B`   | `MIN 3 9`     |

Type `EXIT` or `QUIT` to shut down gracefully.

---

## Sample Interaction

```
ADD 5 7
5 ADD 7 = 12

DIV 10 0
Error: division_by_zero

MUL 3 0.5
3 MUL 0.5 = 1.5

POW 2 10
2 POW 10 = 1024

HELLO 1 2
Error: Unknown operation: HELLO

ADD abc 5
Error: Invalid operand A: ABC is not a number

MIN 2 1 0
Error: Invalid argument count: expected 3 tokens (OP A B), got 4

EXIT
Interface exited.
```

---

## Message Protocol

### Request (Interface → Worker)

Plain text, newline-terminated:

```
OP A B\n
```

Special command: `EXIT\n` — signals the Worker to shut down cleanly.

### Response (Worker → Interface)

JSON, newline-terminated:

**Success:**
```json
{"status":"OK","result":12,"operation":"ADD","a":5,"b":7}
```

**Error:**
```json
{"status":"ERR","error":"Division by zero","operation":"DIV","a":8,"b":0}
```

---

## Error Handling

| Error condition                  | Response error field            |
|----------------------------------|---------------------------------|
| Unknown operation                | `Unknown operation: <op>`       |
| Wrong number of arguments        | `Invalid argument count: ...`   |
| Non-numeric operand A            | `Invalid operand A: ...`        |
| Non-numeric operand B            | `Invalid operand B: ...`        |
| Division by zero                 | `Division by zero`              |
| Modulo by zero                   | `Modulo by zero`                |
| Pipe broken / Worker exited      | Logged; Interface exits cleanly |
| Worker not running at startup    | Interface prints error and exits|

---

## Dependencies

Only Go standard library packages are used:
- `os`, `bufio`, `fmt`, `log`, `strings`, `strconv`, `encoding/json`, `syscall`, `math`

No external dependencies — no `go.mod` required for `go run`.