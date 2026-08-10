package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/rpc/jsonrpc"
)

type AuthClient struct {
	addr string
}

type LoginArgs struct {
	Username string
	Password string
}

type LoginReply struct {
	Success bool
	Error   string
}

func (c *AuthClient) Login(username, password string) (bool, string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return false, "", fmt.Errorf("Failed to connect to auth server: %v", err)
	}
	defer conn.Close()

	client := jsonrpc.NewClient(conn)
	args := LoginArgs{Username: username, Password: password}
	var reply LoginReply
	err = client.Call("AuthService.Login", args, &reply)
	if err != nil {
		return false, "", fmt.Errorf("RPC call failed: %v", err)
	}
	return reply.Success, reply.Error, nil
}

type FileClient struct {
	addr string
}

type GetFileArgs struct {
	Filename string
}

type GetFileReply struct {
	Success     bool
	Error       string
	Content     string
	ContentType string
	Size        int64
}

func (c *FileClient) GetFile(filename string) ([]byte, string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to connect to file server: %v", err)
	}
	defer conn.Close()

	client := jsonrpc.NewClient(conn)
	args := GetFileArgs{Filename: filename}
	var reply GetFileReply
	err = client.Call("FileService.GetFile", args, &reply)
	if err != nil {
		return nil, "", fmt.Errorf("RPC call failed: %v", err)
	}
	if !reply.Success {
		return nil, "", fmt.Errorf("File server error: %s", reply.Error)
	}
	data, err := base64.StdEncoding.DecodeString(reply.Content)
	if err != nil {
		return nil, "", fmt.Errorf("File to decode base64: %v", err)
	}
	return data, reply.ContentType, nil
}

var (
	authAddr string
	fileAddr string
	filePath string
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Template parse error: %v", err), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		authClient := &AuthClient{addr: authAddr}
		success, message, err := authClient.Login(username, password)
		if err != nil {
			http.Error(w, fmt.Sprintf("Authentication error: %v", err), http.StatusInternalServerError)
			return
		}
		if !success {
			http.Error(w, fmt.Sprintf("Login failed: %s", message), http.StatusUnauthorized)
			return
		}

		fileClient := &FileClient{addr: fileAddr}
		data, contentType, err := fileClient.GetFile(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch file: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		if _, err := w.Write(data); err != nil {
			log.Printf("Error writing file response: %v", err)
		}
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func main() {
	port := flag.Int("port", 8081, "Web server listenning port (e.g., 8081)")
	authAddrFlag := flag.String("aserver", "localhost:8082", "Auth server addres (host:port)")
	fileAddrFlag := flag.String("fserver", "localhost:8083", "File server addres (host:port)")
	filePathFlag := flag.String("file", "images/sample.png", "File/image to display from file server after login")
	flag.Parse()

	authAddr = *authAddrFlag
	fileAddr = *fileAddrFlag
	filePath = *filePathFlag

	http.HandleFunc("/", loginHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Web server listening on %s", addr)
	log.Printf("Auth server RPC at %s", authAddr)
	log.Printf("File server RPC at %s", fileAddr)
	log.Printf("Will display file: %s after login", filePath)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
