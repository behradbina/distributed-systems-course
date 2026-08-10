# Pub/Sub – Memory Monitoring and Alerting

This component adds **Publish/Subscribe** functionality to the distributed system:

- The **publisher** (web server) checks its memory usage after every memory consumption.
- When memory exceeds a threshold (default 300 MB), it sends an event to the subscriber.
- The **subscriber** receives the event and prints an alert.

## Files

- `publisher.go` – Web server with memory consumption endpoint and event publishing. It uses `runtime.ReadMemStats` (specifically `Sys`) to measure memory usage.
- `subscriber.go` – Simple HTTP server that receives events and prints alerts.

## How to Run

### 1. Start the Subscriber

```bash
go run subscriber.go -port 8084
```

It listens for POST requests at `/event`. Default port `8084`.

### 2. Start the Publisher (Web Server)

```bash
go run publisher.go \
  -port 8081 \
  -auth <VM2_IP>:8082 \
  -file <VM3_IP>:8083 \
  -image sample.jpg \
  -subscriber http://<SUBSCRIBER_IP>:8084/event \
  -threshold 300
```

- `-subscriber` tells the publisher where to send events.
- `-threshold` sets the memory alert level (in MB).

### 3. Test the System

#### Increase memory to trigger alert

```bash
curl "http://<PUBLISHER_IP>:8080/consume-memory?mb=100"
```

Repeat several times until total allocated memory exceeds the threshold (300 MB).  
The subscriber terminal will then print an alert similar to:

```
[ALERT] High memory usage detected!
  Service:    web-server
  Event type: HIGH_MEMORY_USAGE
  Memory:     345 MB (threshold: 300 MB)
  Timestamp:  2026-05-24T15:04:05Z
```

## Configuration

- **Publisher flags**:
  - `-port` : web server port (default 8080)
  - `-auth` : auth server address
  - `-file` : file server address
  - `-image` : file to fetch from VM3 after login
  - `-subscriber` : URL of the subscriber's `/event` endpoint
  - `-threshold` : memory threshold in MB (default 300)

- **Subscriber flags**:
  - `-port` : listening port (default 8084)