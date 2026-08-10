# Auth Server – VM2 (JSON-RPC)

This service provides user authentication using JSON-RPC over TCP.

## Requirements

- Go 1.16 or later
- A running network interface (not `localhost` only; use real IP addresses in production)

## Running the Server

1. Ensure `users.json` is present in the same directory (or provide a custom path).
2. Run the server:

```bash
go run main.go -port 8082 -users users.json
```

You can change the port and users file path via flags:

- `-port` : listening port (default 8082)
- `-users`: path to JSON user file (default "users.json")

The server binds to `0.0.0.0`, so it accepts connections from any network interface.

## RPC Procedure Exported

- **Service**: `AuthService`
- **Method**: `Login`
- **Arguments**:
  ```json
  {"Username": "string", "Password": "string"}
  ```
- **Reply**:
  ```json
  {"Success": bool, "Error": "string"}
  ```

## Notes

- passwords are hashed before storing and comparing.
- The server handles multiple concurrent clients using a new goroutine per connection.