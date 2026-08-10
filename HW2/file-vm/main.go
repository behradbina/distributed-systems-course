package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type FileService struct {
	baseDir string
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

func (fs *FileService) GetFile(args *GetFileArgs, reply *GetFileReply) error {
	cleanName := filepath.Clean(args.Filename)
	fullPath := filepath.Join(fs.baseDir, cleanName)

	info, err := os.Stat(fullPath)
	if err != nil {
		reply.Success = false
		reply.Error = fmt.Sprintf("File not found: %v", err)
		return nil
	}
	if info.IsDir() {
		reply.Success = false
		reply.Error = "Requested path is a directory"
		return nil
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		reply.Success = false
		reply.Error = fmt.Sprintf("Failed to read file: %v", err)
		return nil
	}

	ext := filepath.Ext(fullPath)
	reply.ContentType = mime.TypeByExtension(ext)
	reply.Content = base64.StdEncoding.EncodeToString(data)
	reply.Size = info.Size()
	reply.Success = true
	return nil
}

func main() {
	port := flag.Int("port", 8083, "Port to listen on (e.g., 8083)")
	dir := flag.String("dir", "./files", "Directory to serve files from")
	flag.Parse()

	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", *dir)
	}
	FileService := &FileService{baseDir: *dir}
	if err := rpc.Register(FileService); err != nil {
		log.Fatalf("Failed to register RPC service: %v", err)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	log.Printf("File server listening on %s (JSON-RPC over TCP)", addr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("Shutting down file server...")
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		}
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go jsonrpc.ServeConn(conn)
	}
}
