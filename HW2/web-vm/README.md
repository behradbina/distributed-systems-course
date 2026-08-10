# Web Server – VM1 (JSON-RPC)

This web server provides a login page, authenticates users via JSON-RPC to the Auth Server (VM2), and after successful login displays an image or file fetched via JSON-RPC from the File Server (VM3).

## Requirements

- Go 1.16+
- Auth server (VM2) running and reachable (default port 8082)
- File server (VM3) running and reachable (default port 8083)
- At least one sample file (e.g., `images/sample.png`) placed in VM3's files directory

## Running the Server

```bash
go run main.go -port 8081 -auth <VM2_IP>:8082 -file <VM3_IP>:8083 -image images/sample.png
```

### Flags

- `-port` : web server listening port (default 8081)
- `-auth` : Auth server address (default localhost:8082)
- `-file` : File server address (default localhost:8083)
- `-image`: filename to request from file server after login (default images/sample.png)

## Usage

1. Open a browser and navigate to `http://<VM1_IP>:8080`
2. Enter one of the test credentials (e.g., alice / alice123)
3. Upon successful login, the browser will display the requested file from VM3 (image or any file).
4. If login fails, an error message is shown.

## RPC Communication

- **Auth RPC method:** `AuthService.Login`
- **File RPC method:** `FileService.GetFile` (returns base64 content)