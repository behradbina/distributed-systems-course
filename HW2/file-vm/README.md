# File Server – VM3 (JSON-RPC)

This server provides static files (images, documents) using JSON-RPC over TCP to the web server (VM1) after successful login. The server handles multiple concurrent clients using a new goroutine per connection.

## Requirements

- Go 1.16 or later
- A directory containing the files to be shared (default: `./files`)
- A running network interface (not `localhost` only; use real IP addresses in production)

## Running the Server

1. Create a `files/images` directory and place at least one image or file inside it.
   ```bash
   mkdir files
   cp /path/to/some/image.jpg files/images/
   ```

2. Run the server:
   ```bash
   go run main.go -port 8083 -dir ./files
   ```

You can change the port and file path via flags:

- `-port` : listening port (default `8083`)
- `-dir`  : path to the directory containing files to serve (default `./files`)

The server binds to `0.0.0.0`, so it accepts connections from any network interface.

## RPC Procedure Exported

- **Service**: `FileService`
- **Method**: `GetFile`
- **Arguments**:
  ```json
  {"Filename": "string"}
  ```
- **Reply**:
  ```json
  {"Success": bool, "Error": "string", "Content": "string", 
  "ContentType": "string", "Size": "int64"}
  ```