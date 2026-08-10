# Part 3 — HTTP Calculator Service
**Containerisation & VM Deployment**

A minimal Go-based HTTP calculator service, packaged into a Docker image,
executed inside a virtual machine, and accessed from the host system.

---

## File Structure

```
part3/
├── main.go       — HTTP service (standard library only)
├── Dockerfile    — multi-stage build → scratch-based image
└── README.md     — this file
```

---

## API Endpoints

### `GET /health`

Simple health check endpoint. Returns a plain OK response if the service is running.

```bash
curl http://localhost:8081/health
```

```
"OK"
```

---

### `GET /calculate?op=OP&a=A&b=B`

Performs an arithmetic operation. All three query parameters are required.

| Parameter | Description                                  |
|-----------|----------------------------------------------|
| `op`      | Operation: `add sub mul div mod pow max min` |
| `a`       | First operand (any valid decimal number)     |
| `b`       | Second operand (any valid decimal number)    |

```bash
curl "http://localhost:8081/calculate?op=add&a=5&b=7"
```

```json
{
  "operation": "add",
  "a": 5,
  "b": 7,
  "result": 12
}
```

More examples:

```bash
curl "http://localhost:8081/calculate?op=pow&a=2&b=10"   # → 1024
curl "http://localhost:8081/calculate?op=div&a=9&b=3"    # → 3
curl "http://localhost:8081/calculate?op=mod&a=10&b=3"   # → 1
curl "http://localhost:8081/calculate?op=max&a=4.5&b=9"  # → 9
```

---

## Error Responses

All errors return JSON with `"error"` and `"code"` fields.

| Condition              | Example error message                             |
|------------------------|---------------------------------------------------|
| Missing parameter(s)   | `Missing query parameters`                        |
| Unknown operation      | `Unknown operation: "xyz"; supported: "add, ..."` |
| Non-numeric `a` or `b` | `Invalid parameter 'b'`                           |
| Division / modulo by 0 | `Division by zero`                                |
| Unknown path           | `Endpoint "/foo" not found`                       |

---

## Accessing the Service from the Host OS (Windows)

When the container is running inside a VirtualBox VM, two things are needed before
the Windows host can reach it: a port-forwarding rule in VirtualBox, and the right
IP address to connect to.

### Step 1 — Configure VirtualBox Port Forwarding

VirtualBox NAT keeps the VM on a private network. A port-forwarding rule maps a
host port to a guest port so traffic from Windows is forwarded into the VM.

1. Open VirtualBox → select your VM → **Settings → Network → Adapter 1** (must be NAT)
2. Click **Advanced → Port Forwarding**
3. Add a new rule:

| Name       | Host Port | Guest IP        | Guest Port |
|------------|-----------|-----------------|------------|
| calculator | 8081      | *(leave blank)* |   8081     |

4. Click OK → OK. The rule takes effect immediately, even while the VM is running.

---

### Step 2 — Building the Docker Image

```bash
# From the part3/ directory:
docker build -t calculator .
```

---

### Step 3 — Running the Container

```bash
# Default port 8081
docker run --rm -p 8081:8081 calculator

# Custom port (e.g. 9090 on host → 8081 in container)
docker run --rm -p 9090:8081 calculator

# Change the internal port via environment variable
docker run --rm -p 9090:9090 -e PORT=9090 calculator

# Run detached (background)
docker run -d --name calc -p 8081:8081 calculator

# View logs
docker logs calc

# Stop
docker stop calc
```

---

### Step 4 — Call the Service from Windows

Open PowerShell, CMD, Git Bash, or WSL:

```bash
# Health check
curl http://127.0.0.1:8081/health

# Arithmetic
curl "http://127.0.0.1:8081/calculate?op=add&a=5&b=7"
curl "http://127.0.0.1:8081/calculate?op=pow&a=2&b=10"
curl "http://127.0.0.1:8081/calculate?op=div&a=9&b=3"
```

---

## Dependencies

Standard library only — no external packages, no `go.mod` required.

`net/http` · `encoding/json` · `math` · `fmt` · `log` · `strconv` · `strings`